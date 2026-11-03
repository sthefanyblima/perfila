package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/domain/service"
	"perfila/internal/shared/errs"
)

// GerarPerfilUseCase — UC-02 («include» Coletar, D-07). Spec: 3.5, RN-05.
type GerarPerfilUseCase struct {
	pre          Precondicoes
	coletar      *ColetarDadosUseCase
	perfilamento *service.PerfilamentoService
	perfis       port.PerfilRepository
}

// NovoGerarPerfilUseCase injeta as dependências (C-08).
func NovoGerarPerfilUseCase(pre Precondicoes, coletar *ColetarDadosUseCase, perfilamento *service.PerfilamentoService, perfis port.PerfilRepository) *GerarPerfilUseCase {
	return &GerarPerfilUseCase{pre: pre, coletar: coletar, perfilamento: perfilamento, perfis: perfis}
}

// Executar gera (ou atualiza) o único perfil do cliente e devolve o perfil salvo. Spec: UC-02.
func (uc *GerarPerfilUseCase) Executar(ctx context.Context, clienteID uuid.UUID) (entity.PerfilCliente, error) {
	if _, err := uc.pre.Verificar(ctx, clienteID); err != nil {
		return entity.PerfilCliente{}, err
	}
	eventos, err := uc.coletar.Executar(ctx, clienteID)
	if err != nil {
		return entity.PerfilCliente{}, err
	}
	novo := uc.perfilamento.GerarPerfil(clienteID, eventos)

	existente, err := uc.perfis.BuscarPorCliente(ctx, clienteID)
	switch {
	case err == nil:
		existente.AtualizarPerfil(novo)
		return uc.perfis.Salvar(ctx, existente)
	case errors.Is(err, errs.ErrPerfilNaoEncontrado):
		return uc.perfis.Salvar(ctx, novo)
	default:
		return entity.PerfilCliente{}, err
	}
}
