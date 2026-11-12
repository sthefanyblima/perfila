package steps

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"golang.org/x/crypto/bcrypt"

	"perfila/internal/domain/entity"
	"perfila/internal/web"
)

// Estado dos cenários @web: o painel via httptest, sem navegador (Spec 3.7.1, bdd-rules).
// A "página" é o corpo depois de seguir um redirecionamento, como faria o navegador.
type estadoWeb struct {
	handler  http.Handler
	usuario  string
	senha    string
	cookies  map[string]string
	csrf     string
	pagina   string
	log      bytes.Buffer
	cadastro string // caminho da ficha do último cliente cadastrado
}

var (
	reCSRF      = regexp.MustCompile(`name="_csrf" value="([^"]+)"`)
	reEtapa     = regexp.MustCompile(`data-teste="etapa" data-estado="([a-z]+)"`)
	reFicha     = regexp.MustCompile(`^/painel/clientes/[0-9a-f-]{36}$`)
	reSintetico = regexp.MustCompile(`c-\d{3}`)
)

func registrarPassosWeb(sc *godog.ScenarioContext, m *Mundo) {
	sc.Step(`^que o painel está no ar com o operador "([^"]*)" e a senha "([^"]*)"$`, m.painelNoAr)
	sc.Step(`^a página "([^"]*)" é aberta sem sessão$`, func(c string) error { return m.webGET(c, false) })
	sc.Step(`^a página "([^"]*)" é aberta$`, func(c string) error { return m.webGET(c, true) })
	sc.Step(`^a resposta leva a "([^"]*)"$`, m.levaA)
	sc.Step(`^a resposta leva à ficha do cliente cadastrado$`, m.levaAFicha)
	sc.Step(`^o operador entra com a senha "([^"]*)" vindo de "([^"]*)"$`, m.entrar)
	sc.Step(`^o operador entra com a senha "([^"]*)"$`, func(s string) error { return m.entrar(s, "") })
	sc.Step(`^que o operador errou a senha (\d+) vezes$`, func(n int) error {
		for i := 0; i < n; i++ {
			if err := m.entrar("senha-errada", ""); err != nil {
				return err
			}
		}
		return nil
	})
	sc.Step(`^que o operador entrou no painel$`, m.entrouNoPainel)
	sc.Step(`^que se passaram (\d+) horas$`, func(h int) error { m.clock.Avancar(time.Duration(h) * time.Hour); return nil })
	sc.Step(`^a página (?:de destino )?mostra "([^"]*)"$`, m.paginaMostra)
	sc.Step(`^a página (?:de destino )?mostra o selo "([^"]*)"$`, func(e string) error {
		return m.paginaTemMarcador(`data-teste="selo" data-estado="`+e+`"`, true)
	})
	sc.Step(`^a página mostra o aviso "([^"]*)"$`, func(a string) error { return m.paginaTemMarcador(`data-teste="aviso-`+a+`"`, true) })
	sc.Step(`^a página pede a confirmação "([^"]*)"$`, func(a string) error {
		return m.paginaTemMarcador(`data-teste="confirmar-`+a+`"`, true)
	})
	sc.Step(`^a página de destino (não )?oferece "([^"]*)"$`, func(nao, a string) error {
		return m.paginaTemMarcador(`data-teste="acao-`+a+`"`, nao == "")
	})
	sc.Step(`^a página não mostra os painéis "([^"]*)", "([^"]*)", "([^"]*)" e "([^"]*)"$`, func(a, b, c, d string) error {
		for _, p := range []string{a, b, c, d} {
			if err := m.paginaTemMarcador(`data-teste="painel-`+p+`"`, false); err != nil {
				return err
			}
		}
		return nil
	})
	sc.Step(`^a lista mostra (\d+) clientes?$`, func(n int) error {
		if got := strings.Count(m.web.pagina, `data-teste="linha-cliente"`); got != n {
			return fmt.Errorf("linhas da lista: esperado %d, obtido %d", n, got)
		}
		return nil
	})
	sc.Step(`^a página mostra as etapas "([^"]*)", "([^"]*)", "([^"]*)" e "([^"]*)"$`, m.etapasNaPagina)
	sc.Step(`^o operador envia "([^"]*)" sem o token CSRF$`, func(req string) error {
		_, caminho, _ := strings.Cut(req, " ")
		return m.webPOST(caminho, url.Values{}, false)
	})
	sc.Step(`^a resposta tem os cabeçalhos de segurança do painel$`, m.cabecalhosSeguranca)
	sc.Step(`^o operador cadastra o cliente "([^"]*)" com o e-mail "([^"]*)"$`, m.cadastrarPeloPainel)
	sc.Step(`^que o operador cadastrou o cliente "([^"]*)" com o e-mail "([^"]*)"$`, m.cadastrarPeloPainel)
	sc.Step(`^o operador confirma "(conceder|revogar)" o consentimento do cliente "([^"]*)"$`, m.confirmaConsentimento)
	sc.Step(`^o operador processa o cliente "([^"]*)" pelo painel$`, func(c string) error {
		return m.webPOST("/painel/clientes/"+c+"/processamento", url.Values{}, true)
	})
	sc.Step(`^o operador registra para o cliente "([^"]*)" uma "([^"]*)" de valor "([^"]*)" na categoria "([^"]*)"$`, m.registrarPeloPainel)
	sc.Step(`^o operador confirma a expiração das recomendações vencidas$`, func() error {
		return m.webPOST("/painel/operacao/expirar", url.Values{}, true)
	})
	sc.Step(`^o log do painel não contém "([^"]*)", "([^"]*)" nem "([^"]*)"$`, m.logSemDados)
}

func (m *Mundo) painelNoAr(usuario, senha string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.MinCost)
	if err != nil {
		return err
	}
	chave := make([]byte, 32)
	_, _ = rand.Read(chave)
	m.web.usuario, m.web.senha, m.web.cookies = usuario, senha, map[string]string{}
	m.segmentos.Adicionar(entity.Segmento{ID: id("seg:padrao"), Nome: "padrao", Prioridade: 99, Ativo: true})
	c := m.uc()
	m.web.handler, err = web.NovoRouter(web.Dependencias{
		Usuario: usuario, SenhaHash: hash, ChaveSessao: chave, Finalidade: "segmentacao", VersaoTermo: "v1",
		Cadastrar: c.cadastrar, Consentimento: c.consentimento, Coletar: c.coletar, Processar: c.processar,
		ListarClientes: c.listarClientes, ConsultarCliente: c.consultarCliente, ConsultarResultado: c.consultarResultado,
		ListarSegmentos: c.listarSegmentos, ListarCatalogo: c.listarCatalogo, Expirar: c.expirar,
		MetaEventos: m.metaEventos, Clock: m.clock, Log: slog.New(slog.NewTextHandler(&m.web.log, nil)),
	})
	return err
}

// requisitar envia a requisição com os cookies guardados e, num redirecionamento, segue o
// destino com GET para preencher a "página".
func (m *Mundo) requisitar(req *http.Request, comSessao bool) {
	if comSessao {
		for k, v := range m.web.cookies {
			req.AddCookie(&http.Cookie{Name: k, Value: v})
		}
	}
	rec := httptest.NewRecorder()
	m.web.handler.ServeHTTP(rec, req)
	m.guardarCookies(rec)
	m.api.resposta = rec
	m.web.pagina = rec.Body.String()
	if loc := rec.Header().Get("Location"); rec.Code == http.StatusSeeOther && loc != "" {
		seguinte := httptest.NewRequest(http.MethodGet, loc, nil)
		for k, v := range m.web.cookies {
			seguinte.AddCookie(&http.Cookie{Name: k, Value: v})
		}
		destino := httptest.NewRecorder()
		m.web.handler.ServeHTTP(destino, seguinte)
		m.guardarCookies(destino)
		m.web.pagina = destino.Body.String()
	}
	if s := reCSRF.FindStringSubmatch(m.web.pagina); s != nil {
		m.web.csrf = s[1]
	}
}

func (m *Mundo) guardarCookies(rec *httptest.ResponseRecorder) {
	for _, c := range rec.Result().Cookies() {
		if c.Value == "" || (!c.Expires.IsZero() && c.Expires.Before(time.Unix(2, 0))) {
			delete(m.web.cookies, c.Name)
			continue
		}
		m.web.cookies[c.Name] = c.Value
	}
}

func caminhoSintetico(c string) string {
	return reSintetico.ReplaceAllStringFunc(c, func(n string) string { return id(n).String() })
}

func (m *Mundo) webGET(caminho string, comSessao bool) error {
	m.requisitar(httptest.NewRequest(http.MethodGet, caminhoSintetico(caminho), nil), comSessao)
	return nil
}

func (m *Mundo) webPOST(caminho string, form url.Values, comCSRF bool) error {
	if comCSRF {
		form.Set("_csrf", m.web.csrf)
	}
	req := httptest.NewRequest(http.MethodPost, caminhoSintetico(caminho), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	m.requisitar(req, true)
	return nil
}

func (m *Mundo) entrar(senha, volta string) error {
	alvo := "/painel/entrar"
	if volta != "" {
		alvo += "?volta=" + url.QueryEscape(caminhoSintetico(volta))
	}
	if err := m.webGET(alvo, true); err != nil {
		return err
	}
	return m.webPOST("/painel/entrar", url.Values{"usuario": {m.web.usuario}, "senha": {senha},
		"volta": {caminhoSintetico(volta)}}, true)
}

func (m *Mundo) entrouNoPainel() error {
	if err := m.entrar(m.web.senha, ""); err != nil {
		return err
	}
	if m.api.resposta.Code != http.StatusSeeOther {
		return fmt.Errorf("entrada recusada: status %d", m.api.resposta.Code)
	}
	return nil
}

func (m *Mundo) levaA(caminho string) error {
	loc := m.api.resposta.Header().Get("Location")
	u, err := url.Parse(loc)
	if m.api.resposta.Code != http.StatusSeeOther || err != nil || u.Path != caminhoSintetico(caminho) {
		return fmt.Errorf("esperava redirecionamento para %s, obtido %d %q", caminho, m.api.resposta.Code, loc)
	}
	return nil
}

func (m *Mundo) levaAFicha() error {
	u, err := url.Parse(m.api.resposta.Header().Get("Location"))
	if m.api.resposta.Code != http.StatusSeeOther || err != nil || !reFicha.MatchString(u.Path) {
		return fmt.Errorf("esperava a ficha do cliente, obtido %d %v", m.api.resposta.Code, u)
	}
	return nil
}

func (m *Mundo) paginaMostra(texto string) error {
	if !strings.Contains(m.web.pagina, texto) {
		return fmt.Errorf("a página não mostra %q", texto)
	}
	return nil
}

func (m *Mundo) paginaTemMarcador(marcador string, quer bool) error {
	if strings.Contains(m.web.pagina, marcador) != quer {
		if quer {
			return fmt.Errorf("a página não tem %s", marcador)
		}
		return fmt.Errorf("a página não deveria ter %s", marcador)
	}
	return nil
}

func (m *Mundo) etapasNaPagina(a, b, c, d string) error {
	var obtido []string
	for _, s := range reEtapa.FindAllStringSubmatch(m.web.pagina, -1) {
		obtido = append(obtido, s[1])
	}
	return iguais("etapas", strings.Join(obtido, ","), strings.Join([]string{a, b, c, d}, ","))
}

func (m *Mundo) cabecalhosSeguranca() error {
	h := m.api.resposta.Header()
	for k, quer := range map[string]string{
		"Content-Security-Policy": "default-src 'self'", "X-Content-Type-Options": "nosniff",
		"Referrer-Policy": "same-origin", "X-Frame-Options": "DENY", "Cache-Control": "no-store",
	} {
		if !strings.Contains(h.Get(k), quer) {
			return fmt.Errorf("cabeçalho %s: esperado %q, obtido %q", k, quer, h.Get(k))
		}
	}
	return nil
}

func (m *Mundo) cadastrarPeloPainel(nome, email string) error {
	return m.webPOST("/painel/clientes", url.Values{"nome": {nome}, "email": {email}}, true)
}

func (m *Mundo) confirmaConsentimento(acao, cliente string) error {
	base := "/painel/clientes/" + cliente + "/consentimento/" + acao
	if err := m.webGET(base, true); err != nil {
		return err
	}
	return m.webPOST(base, url.Values{"finalidade": {"segmentacao"}, "versaoTermo": {"v1"}}, true)
}

func (m *Mundo) registrarPeloPainel(cliente, tipo, valor, categoria string) error {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return err
	}
	return m.webPOST("/painel/clientes/"+cliente+"/eventos", url.Values{"tipo": {tipo}, "valor": {valor},
		"categoria": {categoria}, "origem": {"loja on-line"}, "quando": {m.clock.Agora().In(loc).Format("2006-01-02T15:04")}}, true)
}

func (m *Mundo) logSemDados(a, b, c string) error {
	log := m.web.log.String()
	if log == "" {
		return fmt.Errorf("o painel não registrou nada")
	}
	for _, s := range []string{a, b, c} {
		if strings.Contains(log, s) || strings.Contains(log, url.QueryEscape(s)) {
			return fmt.Errorf("o log do painel contém %q", s)
		}
	}
	return nil
}
