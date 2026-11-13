// Package web é o painel do operador: um adaptador de apresentação HTML, irmão de internal/api.
// Handlers só traduzem HTTP e formulários em chamadas de casos de uso; nenhuma regra de negócio
// mora aqui. Importa application, domain/entity e shared/errs; nunca infra (C-01).
//
// Spec: 3.7.1, Q-20 a Q-29. Fonte: UML7 (componente Web), docs/design/.
package web

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"perfila/internal/application/usecase"
	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/shared/errs"
)

// Dependencias reúne o que o painel precisa; montada só na raiz de composição (C-08).
type Dependencias struct {
	ChaveSessao []byte // PAINEL_CHAVE_SESSAO (32 bytes)
	Finalidade  string // PAINEL_FINALIDADE (Q-25)
	VersaoTermo string // PAINEL_VERSAO_TERMO (Q-25)

	Cadastrar          *usecase.CadastrarClienteUseCase
	Consentimento      *usecase.GerenciarConsentimentoUseCase
	Coletar            *usecase.ColetarDadosUseCase
	Processar          *usecase.ProcessarClienteUseCase
	ListarClientes     *usecase.ListarClientesUseCase
	ConsultarCliente   *usecase.ConsultarClienteUseCase
	ConsultarResultado *usecase.ConsultarResultadoUseCase
	ListarSegmentos    *usecase.ListarSegmentosUseCase
	ListarCatalogo     *usecase.ListarCatalogoUseCase
	Expirar            *usecase.ExpirarRecomendacoesUseCase
	Autenticar         *usecase.AutenticarOperadorUseCase  // v1.5 (Q-36, Q-38)
	Operadores         *usecase.GerenciarOperadoresUseCase // v1.5 (Q-35, Q-40)
	RegistrarAuditoria *usecase.RegistrarAuditoriaUseCase  // v1.5 (RN-13)
	ListarAuditoria    *usecase.ListarAuditoriaUseCase     // v1.5 (RN-13)
	MetaEventos        int                                 // ENGAJAMENTO_META_EVENTOS, para explicar o score

	Clock port.Clock
	Log   *slog.Logger
	// Saude confere o banco (nil = modo memória); Migracoes descreve o esquema aplicado.
	Saude     func(ctx context.Context) error
	Migracoes string
}

type painel struct {
	d          Dependencias
	telas      telas
	tentativas *limitador
}

// NovoRouter monta as rotas de /painel (Spec 3.7.1) com sessão, CSRF e cabeçalhos de segurança.
func NovoRouter(d Dependencias) (http.Handler, error) {
	if d.Log == nil {
		d.Log = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	t, err := carregarTelas()
	if err != nil {
		return nil, err
	}
	p := &painel{d: d, telas: t, tentativas: novoLimitador(d.Clock)}
	sessao := func(perm entity.Permissao, f http.HandlerFunc) http.Handler { return p.exigirSessao(perm, f) }

	mux := http.NewServeMux()
	mux.Handle("GET /painel/static/", estaticos())
	mux.HandleFunc("GET /painel/entrar", p.telaEntrar)
	mux.HandleFunc("POST /painel/entrar", p.entrar)
	mux.Handle("POST /painel/sair", sessao("", p.sair))
	mux.Handle("GET /painel/{$}", sessao("", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, inicio(operadorAtual(r)), http.StatusSeeOther)
	}))
	for _, rotas := range []func(*http.ServeMux, func(entity.Permissao, http.HandlerFunc) http.Handler){
		p.rotasClientes, p.rotasFicha, p.rotasConfiguracao, p.rotasUsuarios, p.rotasAuditoria,
	} {
		rotas(mux, sessao)
	}
	mux.HandleFunc("/painel/", p.naoEncontrada)
	return p.cabecalhos(p.registro(mux)), nil
}

// acao registra uma ação do operador: só operador, ação e ID do cliente (C-09, Q-24).
func (p *painel) acao(r *http.Request, nome, clienteID string) {
	attrs := []slog.Attr{slog.String("operador", operadorDe(r)), slog.String("acao", nome)}
	if clienteID != "" {
		attrs = append(attrs, slog.String("clienteId", clienteID))
	}
	p.d.Log.LogAttrs(r.Context(), slog.LevelInfo, "painel_acao", attrs...)
}

// auditar grava a ação na trilha (RN-13): err nil = SUCESSO; sem permissão ou credenciais =
// NEGADO; outro erro = FALHA (o código vai só para o log). Devolve o erro da gravação: numa
// leitura, quem chama não mostra a tela sem registro; numa ação, só registra no log.
func (p *painel) auditar(r *http.Request, acao entity.AcaoAuditoria, tipo entity.TipoAlvo, alvo uuid.UUID, err error) error {
	resultado := entity.ResultadoSucesso
	switch {
	case errors.Is(err, errs.ErrSemPermissao), errors.Is(err, errs.ErrCredenciaisInvalidas):
		resultado = entity.ResultadoNegado
	case err != nil:
		resultado = entity.ResultadoFalha
	}
	if e := p.d.RegistrarAuditoria.Executar(r.Context(), operadorDe(r), acao, tipo, alvo, resultado); e != nil {
		p.d.Log.ErrorContext(r.Context(), "auditoria_falhou", "acao", string(acao), "erro", e.Error())
		return e
	}
	return nil
}

// registro loga método, caminho (só UUIDs) e status; nunca a query string, que pode ter o termo
// de busca (C-09, Q-24).
func (p *painel) registro(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		p.d.Log.LogAttrs(r.Context(), slog.LevelInfo, "painel_http",
			slog.String("metodo", r.Method), slog.String("caminho", r.URL.Path), slog.Int("status", rw.status))
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// cabecalhos aplica os cabeçalhos de segurança da Spec 3.7.1 a toda resposta do painel.
func (p *painel) cabecalhos(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self'; script-src 'self'; "+
			"frame-ancestors 'none'; form-action 'self'; base-uri 'none'")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "same-origin")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
