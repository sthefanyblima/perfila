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
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"perfila/internal/application/usecase"
	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/shared/errs"
)

// Dependencias reúne o que o painel precisa; montada só na raiz de composição (C-08).
type Dependencias struct {
	ChaveSessao []byte // PAINEL_CHAVE_SESSAO (32 bytes)
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
	SalvarSegmento     *usecase.SalvarSegmentoUseCase      // v1.6 (RN-14, Q-33)
	PreverSegmentacao  *usecase.PreverSegmentacaoUseCase   // v1.6 (Q-42)
	SalvarProduto      *usecase.SalvarProdutoUseCase       // v1.6 (Q-43)
	GerarRelatorio     *usecase.GerarRelatorioUseCase      // v1.7 (RN-15)
	Importar           *usecase.ImportarEventosUseCase     // v1.8 (RN-16, Q-47)
	// Direitos do titular e retenção (v1.9, Q-48 a Q-53).
	AtualizarCadastro *usecase.AtualizarCadastroUseCase
	ExportarDados     *usecase.ExportarDadosTitularUseCase
	EliminarDados     *usecase.EliminarDadosTitularUseCase
	AplicarRetencao   *usecase.AplicarRetencaoUseCase
	// Portal do titular (v1.10, Q-57, Q-58): pedidos e caixa de saída simulada. CaixaSaida nil =
	// portal desligado, e a tela some da navegação.
	MeusDados   *usecase.ConsultarMeusDadosUseCase
	Pedidos     *usecase.ListarPedidosEliminacaoUseCase
	CaixaSaida  *usecase.ListarCaixaSaidaUseCase
	MetaEventos int // ENGAJAMENTO_META_EVENTOS, para explicar o score

	// Públicos para campanha (v1.11, UC-06 parcial, RN-21, RN-22).
	GerarPublico         *usecase.GerarPublicoUseCase
	ExportarPublico      *usecase.ExportarPublicoUseCase
	Relatorio            *usecase.RelatorioAgregadoSegmentosUseCase
	PublicosArquivoHoras int // PUBLICOS_ARQUIVO_HORAS, para o texto da confirmação (Q-64)

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
		p.rotasClientes, p.rotasFicha, p.rotasDireitos, p.rotasConfiguracao, p.rotasUsuarios, p.rotasAuditoria, p.rotasRelatorios, p.rotasImportacao, p.rotasPedidos, p.rotasPublicos,
	} {
		rotas(mux, sessao)
	}
	// Rota desconhecida: com sessão, a página de erro leva o menu; sem sessão, vai para Entrar.
	mux.Handle("/painel/", sessao("", p.naoEncontrada))
	return p.cabecalhos(p.registro(mux)), nil
}

// acao registra uma ação do operador: só operador, ação, ID do cliente e quantidades
// (C-09, Q-24, RN-16).
func (p *painel) acao(r *http.Request, nome, clienteID string, extras ...slog.Attr) {
	attrs := []slog.Attr{slog.String("operador", operadorDe(r)), slog.String("acao", nome)}
	if clienteID != "" {
		attrs = append(attrs, slog.String("clienteId", clienteID))
	}
	attrs = append(attrs, extras...)
	p.d.Log.LogAttrs(r.Context(), slog.LevelInfo, "painel_acao", attrs...)
}

// auditar grava a ação na trilha (RN-13): err nil = SUCESSO; sem permissão ou credenciais =
// NEGADO; outro erro = FALHA (o código vai só para o log). Devolve o erro da gravação: numa
// leitura, quem chama não mostra a tela sem registro; numa ação, só registra no log. campos são
// os nomes dos campos alterados de segmento ou produto (v1.6, Q-44).
func (p *painel) auditar(r *http.Request, acao entity.AcaoAuditoria, tipo entity.TipoAlvo, alvo uuid.UUID, err error, campos ...string) error {
	resultado := entity.ResultadoSucesso
	switch {
	case errors.Is(err, errs.ErrSemPermissao), errors.Is(err, errs.ErrCredenciaisInvalidas):
		resultado = entity.ResultadoNegado
	case err != nil:
		resultado = entity.ResultadoFalha
	}
	if err != nil {
		campos = nil // nada foi alterado
	}
	if e := p.d.RegistrarAuditoria.Executar(r.Context(), operadorDe(r), acao, tipo, alvo, resultado, campos...); e != nil {
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
func (p *painel) cabecalhos(next http.Handler) http.Handler { return cabecalhosSeguranca(next) }

// formularioUTF8 recusa formulário (urlencoded) com bytes que não são UTF-8: o PostgreSQL rejeitaria
// o texto e a pessoa veria um erro 500 em vez de uma resposta clara. Envio multipart (importação)
// tem validação própria. Spec: 3.7.1.
func formularioUTF8(r *http.Request) bool {
	if r.Method != http.MethodPost || !strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		return true
	}
	if err := r.ParseForm(); err != nil {
		return true // o erro de leitura segue o caminho normal de cada tela
	}
	for k, vs := range r.PostForm {
		if !utf8.ValidString(k) {
			return false
		}
		for _, v := range vs {
			if !utf8.ValidString(v) {
				return false
			}
		}
	}
	return true
}

// cabecalhosSeguranca vale para o painel e para o portal do titular (3.7.1, 3.7.2).
func cabecalhosSeguranca(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self'; script-src 'self'; "+
			"frame-ancestors 'none'; form-action 'self'; base-uri 'none'")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "same-origin")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Cache-Control", "no-store")
		// Atrás de um proxy que termina o TLS (hospedagem gratuita da demonstração), o HSTS vem daqui;
		// em HTTP puro (desenvolvimento local) o cabeçalho não é enviado.
		if r.Header.Get("X-Forwarded-Proto") == "https" {
			h.Set("Strict-Transport-Security", "max-age=31536000")
		}
		if !formularioUTF8(r) {
			http.Error(w, "Texto em codificação inválida: envie o formulário em UTF-8.", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}
