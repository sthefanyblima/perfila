package usecase

import (
	"context"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/domain/service"
)

// SegmentarClienteUseCase — UC-03. Spec: 3.5, RN-06, D-06, D-07.
type SegmentarClienteUseCase struct {
	pre         Precondicoes
	perfis      port.PerfilRepository
	segmentacao *service.SegmentacaoService
}

// NovoSegmentarClienteUseCase injeta as dependências (C-08).
func NovoSegmentarClienteUseCase(pre Precondicoes, perfis port.PerfilRepository, segmentacao *service.SegmentacaoService) *SegmentarClienteUseCase {
	return &SegmentarClienteUseCase{pre: pre, perfis: perfis, segmentacao: segmentacao}
}

// Executar carrega o perfil, escolhe o segmento e persiste PerfilCliente.SegmentoID. Spec: UC-03.
func (uc *SegmentarClienteUseCase) Executar(ctx context.Context, clienteID uuid.UUID) (entity.Segmento, error) {
	if _, err := uc.pre.Verificar(ctx, clienteID); err != nil {
		return entity.Segmento{}, err
	}
	perfil, err := uc.perfis.BuscarPorCliente(ctx, clienteID)
	if err != nil {
		return entity.Segmento{}, err
	}
	seg, err := uc.segmentacao.Segmentar(ctx, perfil)
	if err != nil {
		return entity.Segmento{}, err
	}
	id := seg.ID
	perfil.SegmentoID = &id
	if _, err := uc.perfis.Salvar(ctx, perfil); err != nil {
		return entity.Segmento{}, err
	}
	return seg, nil
}
