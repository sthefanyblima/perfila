package web

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"perfila/internal/domain/port"
)

// Sessão sem estado no servidor (Q-22): o cookie carrega operador, validade e token CSRF,
// assinados com HMAC-SHA256. Qualquer réplica valida a mesma sessão. Spec: 3.7.1.
const (
	cookieSessao   = "perfila_sessao"
	cookieEntrar   = "perfila_entrar" // CSRF do formulário Entrar (antes de existir sessão)
	duracaoSessao  = 8 * time.Hour
	maxTentativas  = 5
	janelaBloqueio = 5 * time.Minute
)

var (
	errSessaoAusente = errors.New("sessão ausente ou adulterada")
	errSessaoVencida = errors.New("sessão vencida")
)

type sessao struct {
	Operador string
	Expira   time.Time
	CSRF     string
}

func assinar(chave []byte, s sessao) string {
	corpo := base64.RawURLEncoding.EncodeToString([]byte(s.Operador)) + "|" +
		strconv.FormatInt(s.Expira.Unix(), 10) + "|" + s.CSRF
	return corpo + "." + base64.RawURLEncoding.EncodeToString(mac(chave, corpo))
}

func mac(chave []byte, corpo string) []byte {
	h := hmac.New(sha256.New, chave)
	h.Write([]byte(corpo))
	return h.Sum(nil)
}

// validar confere assinatura e validade; uma sessão vencida devolve errSessaoVencida.
func validar(chave []byte, valor string, agora time.Time) (sessao, error) {
	corpo, assinatura, ok := strings.Cut(valor, ".")
	if !ok {
		return sessao{}, errSessaoAusente
	}
	recebida, err := base64.RawURLEncoding.DecodeString(assinatura)
	if err != nil || !hmac.Equal(recebida, mac(chave, corpo)) {
		return sessao{}, errSessaoAusente
	}
	partes := strings.Split(corpo, "|")
	if len(partes) != 3 {
		return sessao{}, errSessaoAusente
	}
	op, err1 := base64.RawURLEncoding.DecodeString(partes[0])
	exp, err2 := strconv.ParseInt(partes[1], 10, 64)
	if err1 != nil || err2 != nil || partes[2] == "" {
		return sessao{}, errSessaoAusente
	}
	s := sessao{Operador: string(op), Expira: time.Unix(exp, 0).UTC(), CSRF: partes[2]}
	if !agora.Before(s.Expira) {
		return s, errSessaoVencida
	}
	return s, nil
}

func tokenAleatorio() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func gravarCookie(w http.ResponseWriter, nome, valor string, expira time.Time) {
	http.SetCookie(w, &http.Cookie{Name: nome, Value: valor, Path: "/painel", Expires: expira,
		HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode})
}

type chaveContexto struct{}

func comSessao(ctx context.Context, s sessao) context.Context {
	return context.WithValue(ctx, chaveContexto{}, s)
}

func sessaoDe(r *http.Request) sessao {
	s, _ := r.Context().Value(chaveContexto{}).(sessao)
	return s
}

func operadorDe(r *http.Request) string { return sessaoDe(r).Operador }

// exigirSessao leva a Entrar sem sessão válida (guardando o destino), renova a validade a cada
// uso e confere o CSRF em todo POST. Em requisição htmx, a ida para Entrar é HX-Redirect.
func (p *painel) exigirSessao(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		agora := p.d.Clock.Agora()
		var s sessao
		err := errSessaoAusente
		if c, e := r.Cookie(cookieSessao); e == nil {
			s, err = validar(p.d.ChaveSessao, c.Value, agora)
		}
		if err != nil {
			destino := "/painel/entrar?volta=" + url.QueryEscape(voltaPara(r))
			if errors.Is(err, errSessaoVencida) {
				destino += "&expirou=1"
			}
			if r.Header.Get("HX-Request") != "" {
				w.Header().Set("HX-Redirect", destino)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, destino, http.StatusSeeOther)
			return
		}
		if r.Method == http.MethodPost {
			recebido := r.Header.Get("X-CSRF-Token")
			if recebido == "" {
				recebido = r.PostFormValue("_csrf")
			}
			if subtle.ConstantTimeCompare([]byte(recebido), []byte(s.CSRF)) != 1 {
				p.renderErro(w, r, http.StatusForbidden, "Envio recusado", "O formulário expirou ou não veio do painel. Recarregue a página e tente de novo.")
				return
			}
		}
		s.Expira = agora.Add(duracaoSessao)
		gravarCookie(w, cookieSessao, assinar(p.d.ChaveSessao, s), s.Expira)
		next(w, r.WithContext(comSessao(r.Context(), s)))
	})
}

// voltaPara devolve o caminho a retomar depois de entrar; POST volta para a página da ficha.
func voltaPara(r *http.Request) string {
	if r.Method != http.MethodGet {
		if ref, err := url.Parse(r.Header.Get("Referer")); err == nil && destinoSeguro(ref.RequestURI()) {
			return ref.RequestURI()
		}
		return "/painel/clientes"
	}
	return r.URL.RequestURI()
}

// destinoSeguro aceita só caminhos do próprio painel (evita redirecionamento aberto).
func destinoSeguro(d string) bool {
	return strings.HasPrefix(d, "/painel/") && !strings.HasPrefix(d, "//") && !strings.Contains(d, `\`)
}

// limitador conta tentativas de entrada erradas por IP numa janela de 5 minutos, em memória e
// por réplica (Q-22). O IP fica só em memória: nunca vai para o log (C-09).
type limitador struct {
	mu     sync.Mutex
	clock  port.Clock
	falhas map[string][]time.Time
}

func novoLimitador(c port.Clock) *limitador {
	return &limitador{clock: c, falhas: map[string][]time.Time{}}
}

func (l *limitador) recentes(ip string) []time.Time {
	limite := l.clock.Agora().Add(-janelaBloqueio)
	out := l.falhas[ip][:0]
	for _, t := range l.falhas[ip] {
		if t.After(limite) {
			out = append(out, t)
		}
	}
	if len(out) == 0 {
		delete(l.falhas, ip)
		return nil
	}
	l.falhas[ip] = out
	return out
}

func (l *limitador) bloqueado(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.recentes(ip)) >= maxTentativas
}

func (l *limitador) falhou(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.falhas[ip] = append(l.recentes(ip), l.clock.Agora())
}

// ipDe usa o X-Real-IP do Nginx (UML8: as réplicas só são alcançadas por ele) ou o endereço da conexão.
func ipDe(r *http.Request) string {
	if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
