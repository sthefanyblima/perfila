package usecase

import (
	"context"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
)

// ColetarDadosUseCase — UC-01. Spec: 3.5, RN-01, RN-04, D-12.
type ColetarDadosUseCase struct {
	pre     Precondicoes
	eventos port.EventoRepository
	clock   port.Clock
}

// NovoColetarDadosUseCase injeta as dependências (C-08).
func NovoColetarDadosUseCase(pre Precondicoes, eventos port.EventoRepository, clock port.Clock) *ColetarDadosUseCase {
	return &ColetarDadosUseCase{pre: pre, eventos: eventos, clock: clock}
}

// Executar devolve os eventos do cliente (sequência, passos 3 e 4). Spec: UC-01 RN-01.
func (uc *ColetarDadosUseCase) Executar(ctx context.Context, clienteID uuid.UUID) ([]entity.EventoComportamental, error) {
	if _, err := uc.pre.Verificar(ctx, clienteID); err != nil {
		return nil, err
	}
	return uc.eventos.BuscarPorCliente(ctx, clienteID)
}

// Registrar [ADIÇÃO, D-12] valida e grava um evento. Spec: UC-01 RN-01 RN-04.
func (uc *ColetarDadosUseCase) Registrar(ctx context.Context, e entity.EventoComportamental) error {
	if _, err := uc.pre.Verificar(ctx, e.ClienteID); err != nil {
		return err
	}
	if err := e.Validar(uc.clock.Agora()); err != nil {
		return err
	}
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	e.Timestamp = e.Timestamp.UTC()
	return uc.eventos.Salvar(ctx, e)
}
