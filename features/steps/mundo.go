// Package steps traduz os passos Gherkin em chamadas e asserções (B-12: passos finos,
// sem regra de negócio). Cada cenário recebe um Mundo novo (B-13: independência).
package steps

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"perfila/internal/application/usecase"
	"perfila/internal/domain/entity"
	"perfila/internal/domain/service"
	"perfila/internal/infra/external"
	"perfila/internal/infra/memoria"
	"perfila/internal/infra/relogio"
	"perfila/internal/infra/senha"
)

// Mundo guarda o estado de um cenário: relógio fixo, fakes em memória e último resultado.
type Mundo struct {
	ctx   context.Context
	clock *relogio.Fixo

	clientes       *memoria.ClienteRepository
	eventos        *memoria.EventoRepository
	perfis         *memoria.PerfilRepository
	segmentos      *memoria.SegmentoRepository
	recomendacoes  *memoria.RecomendacaoRepository
	consentimentos *memoria.ConsentimentoRepository
	produtos       *memoria.ProdutoRepository
	operadores     *memoria.OperadorRepository  // v1.5
	auditoria      *memoria.AuditoriaRepository // v1.5
	hash           senha.Bcrypt
	// portal do titular (v1.10)
	tokens     *memoria.TokenAcessoRepository
	pedidos    *memoria.PedidoEliminacaoRepository
	caixaSaida *memoria.CaixaSaidaSimulada
	publicos   *memoria.PublicoRepository // v1.11, RN-21

	metaEventos  int
	validadeDias int

	// últimos resultados
	err               error
	clienteCadastrado entity.Cliente
	consentimento     entity.ConsentimentoLGPD
	eventosColeta     []entity.EventoComportamental
	perfil            entity.PerfilCliente
	perfilIDOrigem    uuid.UUID
	segmento          entity.Segmento
	recomendacao      entity.RecomendacaoProduto
	temRecomendacao   bool
	// recomendação gerada antes da ação do cenário (Q-18)
	recomendacaoAnteriorID uuid.UUID
	processamento          usecase.ResultadoProcessamento
	leitura                estadoLeitura
	web                    estadoWeb
	ops                    estadoOperadores
	cfg                    estadoConfig // v1.6
	rel                    estadoRelatorio
	importacao             estadoImportacao
	direitos               estadoDireitos
	portal                 estadoPortal

	api estadoAPI

	pub estadoPublicos // v1.11, RN-21
}

// Padrões da Spec 3.8.
const (
	metaPadrao     = 30
	validadePadrao = 30
	horasArquivo   = 24 // PUBLICOS_ARQUIVO_HORAS (Q-64)
)

func novoMundo() *Mundo {
	return &Mundo{
		ctx:            context.Background(),
		clock:          relogio.NovoFixo(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
		clientes:       memoria.NovoClienteRepository(),
		eventos:        memoria.NovoEventoRepository(),
		perfis:         memoria.NovoPerfilRepository(),
		segmentos:      memoria.NovoSegmentoRepository(),
		recomendacoes:  memoria.NovoRecomendacaoRepository(),
		consentimentos: memoria.NovoConsentimentoRepository(),
		produtos:       memoria.NovoProdutoRepository(),
		operadores:     memoria.NovoOperadorRepository(),
		auditoria:      memoria.NovoAuditoriaRepository(),
		hash:           senha.NovoBcrypt(4), // custo mínimo do bcrypt: cenários rápidos
		ops:            estadoOperadores{sessoes: map[string]time.Time{}},
		tokens:         memoria.NovoTokenAcessoRepository(),
		pedidos:        memoria.NovoPedidoEliminacaoRepository(),
		caixaSaida:     memoria.NovaCaixaSaidaSimulada(),
		publicos:       memoria.NovoPublicoRepository(),
		metaEventos:    metaPadrao,
		validadeDias:   validadePadrao,
		direitos:       estadoDireitos{prazoDias: prazoRetencaoPadrao},
	}
}

// casos de uso montados com a configuração atual do cenário (raiz de composição de teste).
type casos struct {
	cadastrar     *usecase.CadastrarClienteUseCase
	consentimento *usecase.GerenciarConsentimentoUseCase
	coletar       *usecase.ColetarDadosUseCase
	gerarPerfil   *usecase.GerarPerfilUseCase
	segmentar     *usecase.SegmentarClienteUseCase
	recomendar    *usecase.GerarRecomendacaoUseCase
	processar     *usecase.ProcessarClienteUseCase

	listarClientes     *usecase.ListarClientesUseCase
	consultarCliente   *usecase.ConsultarClienteUseCase
	consultarResultado *usecase.ConsultarResultadoUseCase
	listarSegmentos    *usecase.ListarSegmentosUseCase
	listarCatalogo     *usecase.ListarCatalogoUseCase
	expirar            *usecase.ExpirarRecomendacoesUseCase

	autenticar         *usecase.AutenticarOperadorUseCase
	operadores         *usecase.GerenciarOperadoresUseCase
	registrarAuditoria *usecase.RegistrarAuditoriaUseCase
	listarAuditoria    *usecase.ListarAuditoriaUseCase
	expurgarAuditoria  *usecase.ExpurgarAuditoriaUseCase

	salvarSegmento *usecase.SalvarSegmentoUseCase // v1.6
	prever         *usecase.PreverSegmentacaoUseCase
	salvarProduto  *usecase.SalvarProdutoUseCase
	gerarRelatorio *usecase.GerarRelatorioUseCase
	importar       *usecase.ImportarEventosUseCase // RN-16
	// direitos do titular e retenção (v1.9)
	atualizarCadastro *usecase.AtualizarCadastroUseCase
	exportar          *usecase.ExportarDadosTitularUseCase
	eliminar          *usecase.EliminarDadosTitularUseCase
	retencao          *usecase.AplicarRetencaoUseCase

	// portal do titular (v1.10)
	solicitarAcesso   *usecase.SolicitarAcessoTitularUseCase
	autenticarTitular *usecase.AutenticarTitularUseCase
	meusDados         *usecase.ConsultarMeusDadosUseCase
	pedirEliminacao   *usecase.PedirEliminacaoUseCase
	listarPedidos     *usecase.ListarPedidosEliminacaoUseCase
	listarCaixa       *usecase.ListarCaixaSaidaUseCase
	gerarPublico      *usecase.GerarPublicoUseCase
	exportarPublico   *usecase.ExportarPublicoUseCase
	relatorio         *usecase.RelatorioAgregadoSegmentosUseCase
}

func (m *Mundo) uc() casos {
	validador := service.NovoValidadorLGPD(m.consentimentos)
	pre := usecase.NovasPrecondicoes(m.clientes, validador)
	coletar := usecase.NovoColetarDadosUseCase(pre, m.eventos, m.clock)
	gerar := usecase.NovoGerarPublicoUseCase(m.clientes, m.perfis, m.segmentos, m.publicos, validador)
	c := casos{
		cadastrar: usecase.NovoCadastrarClienteUseCase(m.clientes, m.clock),
		consentimento: usecase.NovoGerenciarConsentimentoUseCase(m.clientes, m.consentimentos, m.clock,
			usecase.NovoPropagarRevogacaoUseCase(m.publicos, m.clock, m.pub.logger())),
		gerarPublico: gerar,
		exportarPublico: usecase.NovoExportarPublicoUseCase(gerar, m.publicos, external.DestinoArquivoCSV{}, m.clock,
			horasArquivo, m.pub.logger()),
		relatorio: usecase.NovoRelatorioAgregadoSegmentosUseCase(m.segmentos, m.perfis, m.clientes, validador),
		coletar:   coletar,
		gerarPerfil: usecase.NovoGerarPerfilUseCase(pre, coletar,
			service.NovoPerfilamentoService(m.clock, m.metaEventos), m.perfis),
		segmentar: usecase.NovoSegmentarClienteUseCase(pre, m.perfis, service.NovoSegmentacaoService(m.segmentos)),
		recomendar: usecase.NovoGerarRecomendacaoUseCase(pre, m.perfis, m.segmentos,
			service.NovoRecomendacaoService(m.produtos, m.clock), m.recomendacoes, m.clock, m.validadeDias, nil),
		expirar: usecase.NovoExpirarRecomendacoesUseCase(m.recomendacoes, m.clock),
	}
	c.processar = usecase.NovoProcessarClienteUseCase(c.coletar, c.gerarPerfil, c.segmentar, c.recomendar)
	c.listarClientes = usecase.NovoListarClientesUseCase(m.clientes, m.consentimentos, m.perfis, m.segmentos)
	c.consultarCliente = usecase.NovoConsultarClienteUseCase(m.clientes, c.consentimento)
	c.consultarResultado = usecase.NovoConsultarResultadoUseCase(pre, m.eventos, m.perfis, m.segmentos, m.recomendacoes, m.produtos)
	c.listarSegmentos = usecase.NovoListarSegmentosUseCase(m.segmentos)
	c.listarCatalogo = usecase.NovoListarCatalogoUseCase(m.produtos)
	c.autenticar = usecase.NovoAutenticarOperadorUseCase(m.operadores, m.hash)
	c.operadores = usecase.NovoGerenciarOperadoresUseCase(m.operadores, m.hash, m.clock)
	c.registrarAuditoria = usecase.NovoRegistrarAuditoriaUseCase(m.auditoria, m.clock)
	c.listarAuditoria = usecase.NovoListarAuditoriaUseCase(m.auditoria)
	c.expurgarAuditoria = usecase.NovoExpurgarAuditoriaUseCase(m.auditoria, m.clock, retencaoPadrao)
	c.salvarSegmento = usecase.NovoSalvarSegmentoUseCase(m.segmentos)
	c.prever = usecase.NovoPreverSegmentacaoUseCase(m.segmentos, m.consentimentos, m.clientes, m.perfis)
	c.salvarProduto = usecase.NovoSalvarProdutoUseCase(m.produtos)
	c.gerarRelatorio = usecase.NovoGerarRelatorioUseCase(memoria.NovoRelatorioRepository(m.clientes, m.eventos, m.perfis,
		m.recomendacoes, m.consentimentos), m.segmentos, m.clock, fusoRelatorio)
	c.importar = usecase.NovoImportarEventosUseCase(coletar, pre, m.eventos, m.clock)
	c.atualizarCadastro = usecase.NovoAtualizarCadastroUseCase(m.clientes)
	c.exportar = usecase.NovoExportarDadosTitularUseCase(m.clientes, c.consentimento, m.eventos, m.perfis, m.segmentos,
		m.recomendacoes, m.produtos, m.clock)
	c.eliminar = usecase.NovoEliminarDadosTitularUseCase(m.clientes, m.consentimentos, m.eventos, m.perfis, m.recomendacoes,
		m.tokens, m.pedidos, m.clock)
	c.retencao = usecase.NovoAplicarRetencaoUseCase(m.consentimentos, m.eventos, m.perfis, m.recomendacoes, m.clock, m.direitos.prazoDias)
	c.solicitarAcesso = usecase.NovoSolicitarAcessoTitularUseCase(m.clientes, m.tokens,
		external.NovoNotificadorSimulado(m.caixaSaida, m.clock), m.clock, m.portal.urlBase)
	c.autenticarTitular = usecase.NovoAutenticarTitularUseCase(m.clientes, m.tokens, m.clock)
	c.meusDados = usecase.NovoConsultarMeusDadosUseCase(m.clientes, c.consentimento, m.pedidos)
	c.pedirEliminacao = usecase.NovoPedirEliminacaoUseCase(m.clientes, c.consentimento, m.pedidos, m.clock)
	c.listarPedidos = usecase.NovoListarPedidosEliminacaoUseCase(m.pedidos, m.clientes, c.consentimento)
	c.listarCaixa = usecase.NovoListarCaixaSaidaUseCase(m.caixaSaida, m.clock)
	return c
}

// id converte um identificador sintético ("c-001") em UUID determinístico (B-08).
func id(nome string) uuid.UUID { return uuid.NewSHA1(uuid.NameSpaceOID, []byte("perfila:"+nome)) }

func instante(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("data fora do ISO 8601 UTC (B-09): %q", s)
	}
	return t.UTC(), nil
}

func quase(a, b float64) bool { return math.Abs(a-b) < 1e-9 }
