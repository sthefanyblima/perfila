package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/shared/errs"
)

// RecomendacaoService escolhe um produto para o cliente. Spec: UC-04 RN-07 (regra provisória, D-05).
type RecomendacaoService struct {
	produtos port.ProdutoRepository
	clock    port.Clock
}

// NovoRecomendacaoService injeta catálogo e relógio (C-07, C-08).
func NovoRecomendacaoService(produtos port.ProdutoRepository, clock port.Clock) *RecomendacaoService {
	return &RecomendacaoService{produtos: produtos, clock: clock}
}

// Recomendar devolve uma recomendação no estado GERADA. Candidatos são os produtos ativos
// da CategoriaPreferida; vence o primeiro por Nome, com desempate por ID.
// Sem categoria ou sem candidatos: ErrSemProdutoElegivel. Spec: RN-07.
func (s *RecomendacaoService) Recomendar(ctx context.Context, c entity.Cliente, p entity.PerfilCliente, seg entity.Segmento) (entity.RecomendacaoProduto, error) {
	if p.CategoriaPreferida == "" {
		return entity.RecomendacaoProduto{}, fmt.Errorf("%w: perfil sem categoria preferida", errs.ErrSemProdutoElegivel)
	}
	lista, err := s.produtos.ListarAtivosPorCategoria(ctx, p.CategoriaPreferida)
	if err != nil {
		return entity.RecomendacaoProduto{}, err
	}
	candidatos := make([]entity.Produto, 0, len(lista))
	for _, prod := range lista {
		if prod.Ativo && prod.Categoria == p.CategoriaPreferida {
			candidatos = append(candidatos, prod)
		}
	}
	if len(candidatos) == 0 {
		return entity.RecomendacaoProduto{}, fmt.Errorf("%w: categoria %s sem produtos ativos", errs.ErrSemProdutoElegivel, p.CategoriaPreferida)
	}
	sort.Slice(candidatos, func(i, j int) bool {
		if candidatos[i].Nome != candidatos[j].Nome {
			return candidatos[i].Nome < candidatos[j].Nome
		}
		return candidatos[i].ID.String() < candidatos[j].ID.String()
	})
	escolhido := candidatos[0]
	motivo := fmt.Sprintf("segmento:%s;categoria:%s", seg.Nome, p.CategoriaPreferida)
	return entity.NovaRecomendacao(uuid.New(), c.ID, escolhido.ID, motivo, s.clock.Agora()), nil
}
