package memoria_test

import (
	"testing"

	"perfila/internal/domain/entity"
	"perfila/internal/infra/memoria"
	"perfila/internal/infra/repository/contrato"
)

// Spec: T-19 — os fakes cumprem o mesmo contrato do PostgreSQL.
func TestContratoMemoria(t *testing.T) {
	contrato.Executar(t, func(t *testing.T) contrato.Repositorios {
		segmentos := memoria.NovoSegmentoRepository()
		produtos := memoria.NovoProdutoRepository()
		clientes, eventos, perfis := memoria.NovoClienteRepository(), memoria.NovoEventoRepository(), memoria.NovoPerfilRepository()
		recomendacoes, consentimentos := memoria.NovoRecomendacaoRepository(), memoria.NovoConsentimentoRepository()
		return contrato.Repositorios{
			Clientes:          clientes,
			Eventos:           eventos,
			Perfis:            perfis,
			Segmentos:         segmentos,
			Recomendacoes:     recomendacoes,
			Consentimentos:    consentimentos,
			Produtos:          produtos,
			Relatorios:        memoria.NovoRelatorioRepository(clientes, eventos, perfis, recomendacoes, consentimentos),
			AdicionarSegmento: func(_ *testing.T, s entity.Segmento) { segmentos.Adicionar(s) },
			AdicionarProduto:  func(_ *testing.T, p entity.Produto) { produtos.Adicionar(p) },
		}
	})
}
