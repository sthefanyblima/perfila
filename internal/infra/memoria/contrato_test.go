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
		return contrato.Repositorios{
			Clientes:          memoria.NovoClienteRepository(),
			Eventos:           memoria.NovoEventoRepository(),
			Perfis:            memoria.NovoPerfilRepository(),
			Segmentos:         segmentos,
			Recomendacoes:     memoria.NovoRecomendacaoRepository(),
			Consentimentos:    memoria.NovoConsentimentoRepository(),
			Produtos:          produtos,
			Publicos:          memoria.NovoPublicoRepository(),
			AdicionarSegmento: func(_ *testing.T, s entity.Segmento) { segmentos.Adicionar(s) },
			AdicionarProduto:  func(_ *testing.T, p entity.Produto) { produtos.Adicionar(p) },
		}
	})
}
