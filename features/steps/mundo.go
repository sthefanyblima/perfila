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
	// portal do titular (v1.10)
	tokens     *memoria.TokenAcessoRepository
	pedidos    *memoria.PedidoEliminacaoRepository
	caixaSaida *memoria.CaixaSaidaSimulada

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
	direitos               estadoDireitos
	portal                 estadoPortal

	api estadoAPI
}

// Padrões da Spec 3.8.
const (
	metaPadrao     = 30
	validadePadrao = 30
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
		tokens:         memoria.NovoTokenAcessoRepository(),
		pedidos:        memoria.NovoPedidoEliminacaoRepository(),
		caixaSaida:     memoria.NovaCaixaSaidaSimulada(),
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

	// direitos do titular e retenção (v1.9)
	atualizarCadastro *usecase.AtualizarCadastroUseCase
	exportar          *usecase.ExportarDadosTitularUseCase
	eliminar          *usecase.EliminarDadosTitularUseCase
	retencao          *usecase.AplicarRetencaoUseCase

	// portal do titular (v1.10)
	solicitarAcesso *usecase.SolicitarAcessoTitularUseCase
	autenticar      *usecase.AutenticarTitularUseCase
	meusDados       *usecase.ConsultarMeusDadosUseCase
	pedirEliminacao *usecase.PedirEliminacaoUseCase
	listarPedidos   *usecase.ListarPedidosEliminacaoUseCase
	listarCaixa     *usecase.ListarCaixaSaidaUseCase
}

func (m *Mundo) uc() casos {
	pre := usecase.NovasPrecondicoes(m.clientes, service.NovoValidadorLGPD(m.consentimentos))
	coletar := usecase.NovoColetarDadosUseCase(pre, m.eventos, m.clock)
	c := casos{
		cadastrar:     usecase.NovoCadastrarClienteUseCase(m.clientes, m.clock),
		consentimento: usecase.NovoGerenciarConsentimentoUseCase(m.clientes, m.consentimentos, m.clock),
		coletar:       coletar,
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
	c.atualizarCadastro = usecase.NovoAtualizarCadastroUseCase(m.clientes)
	c.exportar = usecase.NovoExportarDadosTitularUseCase(m.clientes, c.consentimento, m.eventos, m.perfis, m.segmentos,
		m.recomendacoes, m.produtos, m.clock)
	c.eliminar = usecase.NovoEliminarDadosTitularUseCase(m.clientes, m.consentimentos, m.eventos, m.perfis, m.recomendacoes,
		m.tokens, m.pedidos, m.clock)
	c.retencao = usecase.NovoAplicarRetencaoUseCase(m.consentimentos, m.eventos, m.perfis, m.recomendacoes, m.clock, m.direitos.prazoDias)
	c.solicitarAcesso = usecase.NovoSolicitarAcessoTitularUseCase(m.clientes, m.tokens,
		external.NovoNotificadorSimulado(m.caixaSaida, m.clock), m.clock, m.portal.urlBase)
	c.autenticar = usecase.NovoAutenticarTitularUseCase(m.clientes, m.tokens, m.clock)
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
