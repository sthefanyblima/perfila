package web

import (
	"context"
	"crypto/subtle"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"perfila/internal/application/usecase"
	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/shared/errs"
)

// Portal do titular: segundo adaptador HTML do pacote, sob /titular, com roteador, sessão, chave e
// templates próprios. Nenhuma rota recebe ID de cliente: o cliente é sempre o da sessão.
// Spec: v1.10, 3.7.2, RN-19, RN-20, Q-55 a Q-58.
const (
	cookieTitular       = "perfila_titular"
	cookieTitularEntrar = "perfila_titular_entrar" // CSRF de Entrar e do link (antes de existir sessão)
	caminhoPortal       = "/titular"
	duracaoSessaoPortal = 30 * time.Minute
	maxPedidosLink      = 10
	janelaPedidosLink   = 15 * time.Minute
)

// DependenciasPortal reúne o que o portal precisa; montada só na raiz de composição (C-08).
type DependenciasPortal struct {
	ChaveSessao []byte // PORTAL_CHAVE_SESSAO (32 bytes, diferente da do painel)
	Finalidade  string // termo vigente: PAINEL_FINALIDADE
	VersaoTermo string // termo vigente: PAINEL_VERSAO_TERMO

	Solicitar     *usecase.SolicitarAcessoTitularUseCase
	Autenticar    *usecase.AutenticarTitularUseCase
	MeusDados     *usecase.ConsultarMeusDadosUseCase
	Consentimento *usecase.GerenciarConsentimentoUseCase
	ExportarDados *usecase.ExportarDadosTitularUseCase
	Eliminacao    *usecase.PedirEliminacaoUseCase

	Clock port.Clock
	Log   *slog.Logger
}

type portal struct {
	d       DependenciasPortal
	telas   telas
	pedidos *limitador
}

// NovoPortal monta as rotas de /titular (Spec 3.7.2) com sessão, CSRF e cabeçalhos de segurança.
func NovoPortal(d DependenciasPortal) (http.Handler, error) {
	if d.Log == nil {
		d.Log = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	t, err := carregarTelasDe("portal_", "portal_layout.html", "/titular/static/")
	if err != nil {
		return nil, err
	}
	p := &portal{d: d, telas: t, pedidos: novoLimitadorCom(d.Clock, maxPedidosLink, janelaPedidosLink)}
	sessao := func(f func(http.ResponseWriter, *http.Request, usecase.MeusDados)) http.Handler {
		return p.exigirSessao(f)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /titular/static/", estaticosEm("/titular/static/"))
	mux.HandleFunc("GET /titular/{$}", p.inicio)
	mux.HandleFunc("GET /titular/entrar", p.telaEntrar)
	mux.HandleFunc("POST /titular/entrar", p.pedirLink)
	mux.HandleFunc("GET /titular/acesso", p.telaAcesso)
	mux.HandleFunc("POST /titular/acesso", p.entrar)
	mux.Handle("POST /titular/sair", sessao(p.sair))
	mux.Handle("GET /titular/meus-dados", sessao(p.meusDados))
	mux.Handle("GET /titular/revogar", sessao(p.confirmar("revogar")))
	mux.Handle("POST /titular/revogar", sessao(p.revogar))
	mux.Handle("GET /titular/conceder", sessao(p.confirmar("conceder")))
	mux.Handle("POST /titular/conceder", sessao(p.conceder))
	mux.Handle("GET /titular/dados", sessao(p.baixarDados))
	mux.Handle("GET /titular/eliminacao", sessao(p.confirmar("eliminacao")))
	mux.Handle("POST /titular/eliminacao", sessao(p.pedirEliminacao))
	mux.HandleFunc("/titular/", func(w http.ResponseWriter, r *http.Request) {
		p.renderErro(w, r, http.StatusNotFound, "Página não encontrada", "Confira o endereço ou volte para o início do portal.")
	})
	return cabecalhosSeguranca(p.registro(mux)), nil
}

// registro: método, caminho e status, nunca a query string, que leva o token do link (C-09, Q-56).
func (p *portal) registro(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		p.d.Log.LogAttrs(r.Context(), slog.LevelInfo, "portal_http",
			slog.String("metodo", r.Method), slog.String("caminho", r.URL.Path), slog.Int("status", rw.status))
	})
}

// acao registra a ação do titular só com o ID do cliente da sessão (C-09).
func (p *portal) acao(r *http.Request, nome string, clienteID uuid.UUID) {
	attrs := []slog.Attr{slog.String("acao", nome)}
	if clienteID != uuid.Nil {
		attrs = append(attrs, slog.String("clienteId", clienteID.String()))
	}
	p.d.Log.LogAttrs(r.Context(), slog.LevelInfo, "portal_acao", attrs...)
}

// ---- sessão própria (Q-55): clienteId|expiraEm|csrf assinado com PORTAL_CHAVE_SESSAO ----

type chaveTitular struct{}

func titularDe(r *http.Request) sessao {
	s, _ := r.Context().Value(chaveTitular{}).(sessao)
	return s
}

func (p *portal) gravarSessao(w http.ResponseWriter, s sessao) {
	gravarCookieEm(w, caminhoPortal, cookieTitular, assinar(p.d.ChaveSessao, s), s.Expira)
}

// exigirSessao leva a Entrar sem sessão válida, confere o CSRF em todo POST, renova a validade e
// carrega Meus dados a cada requisição: cliente eliminado ou inexistente encerra a sessão.
func (p *portal) exigirSessao(next func(http.ResponseWriter, *http.Request, usecase.MeusDados)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		agora := p.d.Clock.Agora()
		err := errSessaoAusente
		var s sessao
		if c, e := r.Cookie(cookieTitular); e == nil {
			s, err = validar(p.d.ChaveSessao, c.Value, agora)
		}
		id, errID := uuid.Parse(s.Operador)
		if err != nil || errID != nil {
			destino := "/titular/entrar"
			if errors.Is(err, errSessaoVencida) {
				destino += "?expirou=1"
			}
			p.redirecionar(w, r, destino)
			return
		}
		if r.Method == http.MethodPost {
			recebido := r.Header.Get("X-CSRF-Token")
			if recebido == "" {
				recebido = r.PostFormValue("_csrf")
			}
			if subtle.ConstantTimeCompare([]byte(recebido), []byte(s.CSRF)) != 1 {
				p.renderErro(w, r, http.StatusForbidden, "Envio recusado", "O formulário expirou. Recarregue a página e tente de novo.")
				return
			}
		}
		r = r.WithContext(context.WithValue(r.Context(), chaveTitular{}, s))
		dados, err := p.d.MeusDados.Executar(r.Context(), id)
		if errors.Is(err, errs.ErrClienteEliminado) || errors.Is(err, errs.ErrClienteNaoEncontrado) {
			gravarCookieEm(w, caminhoPortal, cookieTitular, "", time.Unix(1, 0))
			p.redirecionar(w, r, "/titular/entrar")
			return
		}
		if err != nil {
			p.falha(w, r, err)
			return
		}
		s.Expira = agora.Add(duracaoSessaoPortal)
		p.gravarSessao(w, s)
		next(w, r, dados)
	})
}

func (p *portal) redirecionar(w http.ResponseWriter, r *http.Request, destino string) {
	if r.Header.Get("HX-Request") != "" {
		w.Header().Set("HX-Redirect", destino)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	http.Redirect(w, r, destino, http.StatusSeeOther)
}

// tokenEntrar grava o CSRF dos formulários sem sessão (Entrar e o link).
func (p *portal) tokenEntrar(w http.ResponseWriter) string {
	t := tokenAleatorio()
	gravarCookieEm(w, caminhoPortal, cookieTitularEntrar, t, p.d.Clock.Agora().Add(time.Hour))
	return t
}

func csrfEntrarConfere(r *http.Request) bool {
	c, err := r.Cookie(cookieTitularEntrar)
	return err == nil && subtle.ConstantTimeCompare([]byte(c.Value), []byte(r.PostFormValue("_csrf"))) == 1
}

// ---- telas ----

type dadosEntrarTitular struct {
	Token, Email, ErroEmail string
	Enviado, Expirou        bool
	Erro                    string
}

type dadosAcesso struct {
	Token, Link string
	Invalido    bool
}

// dadosTitular alimenta Meus dados e as confirmações.
type dadosTitular struct {
	usecase.MeusDados
	Confirmar   string // revogar, conceder ou eliminacao
	Finalidade  string
	VersaoTermo string
	Erro        *alerta
}

func (d dadosTitular) Linha() []marco {
	return linhaConsentimento(dadosFicha{Cliente: d.Cliente, Consentimento: d.Consentimento, Movimentos: d.Historico}, textoOrigemTitular)
}

func (d dadosTitular) Concedido() bool { return d.Consentimento.Status == entity.StatusConcedido }

// PodeConceder: uma ação de consentimento por tela, e nunca com pedido de eliminação aberto (RN-20).
func (d dadosTitular) PodeConceder() bool { return !d.Concedido() && d.PedidoAberto == nil }

func (d dadosTitular) PodePedirEliminacao() bool { return !d.Concedido() && d.PedidoAberto == nil }

// FinalidadeEmPalavras traduz a finalidade para o titular (03-textos.md).
func (d dadosTitular) FinalidadeEmPalavras() string { return finalidadeEmPalavras(d.Finalidade) }

func (d dadosTitular) FinalidadeAtual() string {
	return finalidadeEmPalavras(d.Consentimento.Finalidade)
}

func finalidadeEmPalavras(f string) string {
	if f == "segmentacao" {
		return "segmentação: entender seu comportamento de compra para recomendar produtos"
	}
	return f
}

// textoOrigemTitular: a origem dos movimentos contada ao titular, sem o login da equipe.
func textoOrigemTitular(m entity.MovimentoConsentimento) string {
	switch m.Origem {
	case entity.OrigemPortal:
		return "por você, no portal"
	case entity.OrigemPainel:
		return "pela nossa equipe"
	case entity.OrigemAPI:
		return "por um sistema integrado"
	case entity.OrigemMigracao:
		return "reconstruído das datas"
	}
	return "pelo sistema"
}

// GET /titular — Meus dados com sessão; senão, Entrar.
func (p *portal) inicio(w http.ResponseWriter, r *http.Request) {
	destino := "/titular/entrar"
	if c, err := r.Cookie(cookieTitular); err == nil {
		if _, err := validar(p.d.ChaveSessao, c.Value, p.d.Clock.Agora()); err == nil {
			destino = "/titular/meus-dados"
		}
	}
	http.Redirect(w, r, destino, http.StatusSeeOther)
}

// GET /titular/entrar — formulário do e-mail; ?enviado=1 mostra a resposta de sempre (RN-19).
func (p *portal) telaEntrar(w http.ResponseWriter, r *http.Request) {
	d := dadosEntrarTitular{Enviado: r.URL.Query().Get("enviado") == "1", Expirou: r.URL.Query().Get("expirou") == "1"}
	d.Token = p.tokenEntrar(w)
	p.render(w, r, http.StatusOK, "entrar", pagina{Titulo: "Acesse seus dados", Dados: d})
}

// POST /titular/entrar — pede o link; a resposta é a mesma com ou sem cadastro (RN-19, Q-56).
func (p *portal) pedirLink(w http.ResponseWriter, r *http.Request) {
	d := dadosEntrarTitular{Email: strings.TrimSpace(r.PostFormValue("email"))}
	if !csrfEntrarConfere(r) {
		d.Token, d.Erro = p.tokenEntrar(w), "O formulário expirou. Tente de novo."
		p.render(w, r, http.StatusForbidden, "entrar", pagina{Titulo: "Acesse seus dados", Dados: d})
		return
	}
	ip := ipDe(r)
	if p.pedidos.bloqueado(ip) {
		d.Token, d.Erro = p.tokenEntrar(w), "Muitas tentativas. Aguarde 15 minutos."
		p.render(w, r, http.StatusTooManyRequests, "entrar", pagina{Titulo: "Acesse seus dados", Dados: d})
		return
	}
	p.pedidos.falhou(ip) // conta todo pedido, conhecido ou não
	err := p.d.Solicitar.Executar(r.Context(), d.Email)
	if errors.Is(err, errs.ErrDadosInvalidos) {
		d.Token, d.ErroEmail = p.tokenEntrar(w), "Informe o e-mail do seu cadastro."
		p.render(w, r, http.StatusUnprocessableEntity, "entrar", pagina{Titulo: "Acesse seus dados", Dados: d})
		return
	}
	if err != nil {
		p.falha(w, r, err)
		return
	}
	p.acao(r, "link_pedido", uuid.Nil)
	http.Redirect(w, r, "/titular/entrar?enviado=1", http.StatusSeeOther)
}

// GET /titular/acesso?t= — só mostra o botão: abrir o link não o consome (Q-56).
func (p *portal) telaAcesso(w http.ResponseWriter, r *http.Request) {
	d := dadosAcesso{Token: p.tokenEntrar(w), Link: r.URL.Query().Get("t")}
	p.render(w, r, http.StatusOK, "acesso", pagina{Titulo: "Entrar no portal", Dados: d})
}

// POST /titular/acesso — consome o link e grava a sessão do titular (RN-19).
func (p *portal) entrar(w http.ResponseWriter, r *http.Request) {
	if !csrfEntrarConfere(r) {
		p.renderErro(w, r, http.StatusForbidden, "Envio recusado", "O formulário expirou. Abra o link de novo.")
		return
	}
	id, err := p.d.Autenticar.Executar(r.Context(), r.PostFormValue("t"))
	if errors.Is(err, errs.ErrAcessoInvalido) {
		p.render(w, r, http.StatusUnauthorized, "acesso", pagina{Titulo: "Link inválido", Dados: dadosAcesso{Invalido: true}})
		return
	}
	if err != nil {
		p.falha(w, r, err)
		return
	}
	s := sessao{Operador: id.String(), Expira: p.d.Clock.Agora().Add(duracaoSessaoPortal), CSRF: tokenAleatorio()}
	p.gravarSessao(w, s)
	gravarCookieEm(w, caminhoPortal, cookieTitularEntrar, "", time.Unix(1, 0))
	p.acao(r, "entrada", id)
	http.Redirect(w, r, "/titular/meus-dados", http.StatusSeeOther)
}

// POST /titular/sair — apaga o cookie.
func (p *portal) sair(w http.ResponseWriter, r *http.Request, d usecase.MeusDados) {
	p.acao(r, "saida", d.Cliente.ID)
	gravarCookieEm(w, caminhoPortal, cookieTitular, "", time.Unix(1, 0))
	http.Redirect(w, r, "/titular/entrar", http.StatusSeeOther)
}

func (p *portal) dados(d usecase.MeusDados) dadosTitular {
	return dadosTitular{MeusDados: d, Finalidade: p.d.Finalidade, VersaoTermo: p.d.VersaoTermo}
}

// GET /titular/meus-dados — cadastro, autorização, histórico e direitos.
func (p *portal) meusDados(w http.ResponseWriter, r *http.Request, d usecase.MeusDados) {
	p.render(w, r, http.StatusOK, "meus_dados", pagina{Titulo: "Meus dados", Dados: p.dados(d)})
}

// GET /titular/{revogar,conceder,eliminacao} — confirmação em página própria (sem JavaScript).
func (p *portal) confirmar(acao string) func(http.ResponseWriter, *http.Request, usecase.MeusDados) {
	return func(w http.ResponseWriter, r *http.Request, d usecase.MeusDados) {
		v := p.dados(d)
		ok := map[string]bool{"revogar": v.Concedido(), "conceder": v.PodeConceder(), "eliminacao": v.PodePedirEliminacao()}[acao]
		if !ok { // a ação não cabe no estado atual: Meus dados mostra o que dá para fazer
			http.Redirect(w, r, "/titular/meus-dados", http.StatusSeeOther)
			return
		}
		v.Confirmar = acao
		p.render(w, r, http.StatusOK, "confirmar", pagina{Titulo: "Confirmar", Dados: v})
	}
}

func (p *portal) comOrigem(r *http.Request) context.Context {
	return usecase.ComOrigem(r.Context(), entity.OrigemPortal, "")
}

// POST /titular/revogar — UC-08 Revogar com a origem portal; efeito imediato (RN-03, RN-10).
func (p *portal) revogar(w http.ResponseWriter, r *http.Request, d usecase.MeusDados) {
	if err := p.d.Consentimento.Revogar(p.comOrigem(r), d.Cliente.ID); err != nil {
		p.recusar(w, r, d, "revogar", err)
		return
	}
	p.acao(r, "revogacao", d.Cliente.ID)
	http.Redirect(w, r, "/titular/meus-dados?aviso=autorizacao-retirada", http.StatusSeeOther)
}

// POST /titular/conceder — UC-08 Conceder com o termo vigente; nunca com pedido aberto (RN-20).
func (p *portal) conceder(w http.ResponseWriter, r *http.Request, d usecase.MeusDados) {
	if d.PedidoAberto != nil {
		http.Redirect(w, r, "/titular/meus-dados", http.StatusSeeOther)
		return
	}
	if err := p.d.Consentimento.Conceder(p.comOrigem(r), d.Cliente.ID, p.d.Finalidade, p.d.VersaoTermo); err != nil {
		p.recusar(w, r, d, "conceder", err)
		return
	}
	p.acao(r, "concessao", d.Cliente.ID)
	http.Redirect(w, r, "/titular/meus-dados?aviso=autorizacao-registrada", http.StatusSeeOther)
}

// GET /titular/dados — o mesmo arquivo da v1.9, só do titular da sessão (Q-49).
func (p *portal) baixarDados(w http.ResponseWriter, r *http.Request, d usecase.MeusDados) {
	dados, err := p.d.ExportarDados.Executar(r.Context(), d.Cliente.ID)
	if err != nil {
		p.falha(w, r, err)
		return
	}
	if err := escreverExportacao(w, dados); err != nil {
		p.falha(w, r, err)
		return
	}
	p.acao(r, "exportacao", d.Cliente.ID)
}

// POST /titular/eliminacao — PedirEliminacao (RN-20); a operação atende no painel.
func (p *portal) pedirEliminacao(w http.ResponseWriter, r *http.Request, d usecase.MeusDados) {
	if _, err := p.d.Eliminacao.Executar(r.Context(), d.Cliente.ID); err != nil {
		p.recusar(w, r, d, "eliminacao", err)
		return
	}
	p.acao(r, "eliminacao_pedida", d.Cliente.ID)
	http.Redirect(w, r, "/titular/meus-dados?aviso=eliminacao-pedida", http.StatusSeeOther)
}

// recusar volta a Meus dados com a mensagem do código de erro (03-textos.md).
func (p *portal) recusar(w http.ResponseWriter, r *http.Request, d usecase.MeusDados, acao string, err error) {
	if statusDe(err) == http.StatusInternalServerError {
		p.falha(w, r, err)
		return
	}
	titulo, texto := mensagemTitular(err, acao)
	v := p.dados(d)
	v.Erro = &alerta{Tom: "critical", Titulo: titulo, Texto: texto, Agora: true}
	p.render(w, r, statusDe(err), "meus_dados", pagina{Titulo: "Meus dados", Dados: v})
}

// mensagemTitular fala com o titular, sem jargão (03-textos.md, Portal do titular).
func mensagemTitular(err error, acao string) (string, string) {
	if errs.Codigo(err) == errs.ErrTransicaoInvalida.Codigo {
		switch acao {
		case "eliminacao":
			return "Autorização concedida", "Para pedir a eliminação, retire a autorização antes."
		case "revogar":
			return "Nada para retirar", "Não há autorização concedida para retirar."
		}
		return "Autorização já concedida", "Sua autorização já está registrada."
	}
	return mensagem(err, acao)
}

func (p *portal) falha(w http.ResponseWriter, r *http.Request, err error) {
	p.d.Log.ErrorContext(r.Context(), "portal_erro_interno", "caminho", r.URL.Path, "erro", err.Error())
	p.renderErro(w, r, http.StatusInternalServerError, "Algo falhou do nosso lado", "Nada foi alterado. Tente de novo em instantes.")
}

func (p *portal) renderErro(w http.ResponseWriter, r *http.Request, status int, titulo, texto string) {
	p.render(w, r, status, "erro", pagina{Titulo: titulo, Dados: dadosErro{titulo, texto}})
}

// render completa a moldura do portal: CSRF da sessão do titular e o toast do PRG.
func (p *portal) render(w http.ResponseWriter, r *http.Request, status int, tela string, pg pagina) {
	pg.CSRF = titularDe(r).CSRF
	if pg.Aviso == "" {
		pg.Aviso = avisoTitular(r.URL.Query().Get("aviso"))
	}
	renderTela(w, r, p.d.Log, p.telas[tela], tela, status, pg)
}

func avisoTitular(codigo string) string {
	switch codigo {
	case "autorizacao-retirada":
		return "Autorização retirada"
	case "autorizacao-registrada":
		return "Autorização registrada"
	case "eliminacao-pedida":
		return "Pedido de eliminação registrado"
	}
	return ""
}
