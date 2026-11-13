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

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/shared/errs"
)

// Sessão sem estado no servidor (Q-22): o cookie carrega o ID e o login do operador, o carimbo
// da senha, a validade e o token CSRF, assinados com HMAC-SHA256. Qualquer réplica valida a
// mesma sessão e confere o operador no banco a cada requisição (Q-38). Spec: 3.7.1.
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
	OperadorID uuid.UUID
	Operador   string // login: logs e auditoria
	Carimbo    int64  // SenhaAtualizadaEm em microssegundos
	Expira     time.Time
	CSRF       string
}

func novaSessao(o entity.Operador, expira time.Time) sessao {
	return sessao{OperadorID: o.ID, Operador: o.Login, Carimbo: o.SenhaAtualizadaEm.UnixMicro(), Expira: expira, CSRF: tokenAleatorio()}
}

func assinar(chave []byte, s sessao) string {
	corpo := s.OperadorID.String() + "|" + base64.RawURLEncoding.EncodeToString([]byte(s.Operador)) + "|" +
		strconv.FormatInt(s.Carimbo, 10) + "|" + strconv.FormatInt(s.Expira.Unix(), 10) + "|" + s.CSRF
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
	if len(partes) != 5 {
		return sessao{}, errSessaoAusente
	}
	id, err0 := uuid.Parse(partes[0])
	op, err1 := base64.RawURLEncoding.DecodeString(partes[1])
	carimbo, err2 := strconv.ParseInt(partes[2], 10, 64)
	exp, err3 := strconv.ParseInt(partes[3], 10, 64)
	if err0 != nil || err1 != nil || err2 != nil || err3 != nil || partes[4] == "" {
		return sessao{}, errSessaoAusente
	}
	s := sessao{OperadorID: id, Operador: string(op), Carimbo: carimbo, Expira: time.Unix(exp, 0).UTC(), CSRF: partes[4]}
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

// contextoSessao é o que a requisição autenticada carrega: a sessão e o operador conferido.
type contextoSessao struct {
	s  sessao
	op entity.Operador
}

func comSessao(ctx context.Context, s sessao, op entity.Operador) context.Context {
	return context.WithValue(ctx, chaveContexto{}, contextoSessao{s: s, op: op})
}

func sessaoDe(r *http.Request) sessao {
	c, _ := r.Context().Value(chaveContexto{}).(contextoSessao)
	return c.s
}

// operadorAtual é o operador conferido no banco nesta requisição (papel atual, Q-38).
func operadorAtual(r *http.Request) entity.Operador {
	c, _ := r.Context().Value(chaveContexto{}).(contextoSessao)
	return c.op
}

func operadorDe(r *http.Request) string { return sessaoDe(r).Operador }

// redirecionar usa HX-Redirect em requisição htmx (a página inteira muda) e 303 nas demais.
func redirecionar(w http.ResponseWriter, r *http.Request, destino string, statusHtmx int) {
	if r.Header.Get("HX-Request") != "" {
		w.Header().Set("HX-Redirect", destino)
		w.WriteHeader(statusHtmx)
		return
	}
	http.Redirect(w, r, destino, http.StatusSeeOther)
}

// inicio é a primeira tela do papel: Clientes para quem pode, senão Segmentos (Spec 3.7.1).
func inicio(o entity.Operador) string {
	if o.Papel.Pode(entity.PermClientes) {
		return "/painel/clientes"
	}
	return "/painel/segmentos"
}

// exigirSessao leva a Entrar sem sessão válida (guardando o destino), confere o operador no banco
// (inativo ou com outra senha encerra a sessão, Q-38), confere o CSRF em todo POST, leva a Minha
// senha quem tem senha provisória, aplica a permissão da rota (RN-11) e renova a validade.
// perm vazia = rota sem permissão (Minha senha, Sair).
func (p *painel) exigirSessao(perm entity.Permissao, next http.HandlerFunc) http.Handler {
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
			redirecionar(w, r, destino, http.StatusUnauthorized)
			return
		}
		op, err := p.d.Autenticar.Conferir(r.Context(), s.OperadorID, time.UnixMicro(s.Carimbo))
		if errors.Is(err, errs.ErrCredenciaisInvalidas) {
			r = r.WithContext(comSessao(r.Context(), s, entity.Operador{}))
			p.auditar(r, entity.AcaoSessaoEncerrada, entity.AlvoNenhum, uuid.Nil, nil)
			gravarCookie(w, cookieSessao, "", time.Unix(1, 0))
			redirecionar(w, r, "/painel/entrar?encerrada=1", http.StatusUnauthorized)
			return
		}
		if err != nil {
			p.falha(w, r, err, "")
			return
		}
		r = r.WithContext(comSessao(r.Context(), s, op))
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
		if op.TrocarSenha && r.URL.Path != "/painel/minha-senha" && r.URL.Path != "/painel/sair" {
			redirecionar(w, r, "/painel/minha-senha", http.StatusOK)
			return
		}
		if perm != "" && !op.Papel.Pode(perm) {
			p.semPermissao(w, r)
			return
		}
		next(w, r)
	})
}

// semPermissao responde 403 com a tela de erro e registra o acesso negado (RN-13), com o cliente
// como alvo quando a rota tem um.
func (p *painel) semPermissao(w http.ResponseWriter, r *http.Request) {
	tipo, alvo := entity.AlvoNenhum, uuid.Nil
	if id, err := uuid.Parse(r.PathValue("id")); err == nil && strings.HasPrefix(r.URL.Path, "/painel/clientes/") {
		tipo, alvo = entity.AlvoCliente, id
	}
	p.auditar(r, entity.AcaoAcessoNegado, tipo, alvo, errs.ErrSemPermissao)
	p.renderErro(w, r, http.StatusForbidden, "Sem permissão", "Seu perfil de acesso não permite abrir esta página. Se precisar dela, fale com quem administra o painel.")
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
