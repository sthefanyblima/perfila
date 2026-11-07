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
		metaEventos:    metaPadrao,
		validadeDias:   validadePadrao,
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
	expirar       *usecase.ExpirarRecomendacoesUseCase
}

func (m *Mundo) uc() casos {
	pre := usecase.NovasPrecondicoes(m.clientes, service.NovoValidadorLGPD(m.consentimentos))
	coletar := usecase.NovoColetarDadosUseCase(pre, m.eventos, m.clock)
	return casos{
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
