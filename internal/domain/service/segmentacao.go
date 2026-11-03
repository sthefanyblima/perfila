package service

import (
	"context"
	"fmt"
	"sort"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/shared/errs"
)

// SegmentacaoService escolhe o segmento de um perfil. Spec: UC-03 RN-06.
type SegmentacaoService struct {
	segmentos port.SegmentoRepository
}

// NovoSegmentacaoService injeta o repositório de segmentos (C-08).
func NovoSegmentacaoService(segmentos port.SegmentoRepository) *SegmentacaoService {
	return &SegmentacaoService{segmentos: segmentos}
}

// Segmentar avalia os segmentos ativos em ordem crescente de Prioridade; o primeiro
// cujo critério casa vence. Sem segmento aplicável: ErrDadosInvalidos (OPEN_QUESTIONS Q-02).
// Spec: RN-06.
func (s *SegmentacaoService) Segmentar(ctx context.Context, p entity.PerfilCliente) (entity.Segmento, error) {
	lista, err := s.segmentos.ListarAtivos(ctx)
	if err != nil {
		return entity.Segmento{}, err
	}
	sort.SliceStable(lista, func(i, j int) bool { return lista[i].Prioridade < lista[j].Prioridade })
	for _, seg := range lista {
		if !seg.Ativo {
			continue
		}
		casa, err := seg.ClassificarPerfil(p)
		if err != nil {
			return entity.Segmento{}, fmt.Errorf("segmento %s: %w", seg.Nome, err)
		}
		if casa {
			return seg, nil
		}
	}
	return entity.Segmento{}, fmt.Errorf("%w: nenhum segmento ativo casa com o perfil", errs.ErrDadosInvalidos)
}
