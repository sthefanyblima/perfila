// Package usecase orquestra o domínio: aplica pré-condições, chama serviços e persiste.
//
// Spec: 3.5, RN-01, C-03, D-09. Fonte: UML7 (Casos de Uso) e UML4 (Sequência).
package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/domain/service"
	"perfila/internal/shared/errs"
)

// Precondicoes aplica RN-01, comum a UC-01..UC-04.
type Precondicoes struct {
	clientes  port.ClienteRepository
	validador *service.ValidadorLGPD
}

// NovasPrecondicoes injeta as dependências (C-08).
func NovasPrecondicoes(clientes port.ClienteRepository, validador *service.ValidadorLGPD) Precondicoes {
	return Precondicoes{clientes: clientes, validador: validador}
}

// Verificar checa, nesta ordem: cliente existe, não foi eliminado (v1.9), está ativo e tem
// consentimento válido. Spec: RN-01, RN-17.
func (p Precondicoes) Verificar(ctx context.Context, clienteID uuid.UUID) (entity.Cliente, error) {
	c, err := p.clientes.BuscarPorID(ctx, clienteID)
	if err != nil {
		return entity.Cliente{}, err
	}
	if c.EstaEliminado() {
		return entity.Cliente{}, fmt.Errorf("%w: %s", errs.ErrClienteEliminado, clienteID)
	}
	if !c.Ativo {
		return entity.Cliente{}, fmt.Errorf("%w: %s", errs.ErrClienteInativo, clienteID)
	}
	if err := p.consentimento(ctx, clienteID); err != nil {
		return entity.Cliente{}, err
	}
	return c, nil
}

// VerificarLeitura é RN-01 sem a exigência de cliente ativo: cliente inativo com consentimento
// válido pode ter o histórico consultado, só para leitura (Q-31). Nada de comportamento é lido
// sem consentimento (C-03).
func (p Precondicoes) VerificarLeitura(ctx context.Context, clienteID uuid.UUID) (entity.Cliente, error) {
	c, err := p.clientes.BuscarPorID(ctx, clienteID)
	if err != nil {
		return entity.Cliente{}, err
	}
	if err := p.consentimento(ctx, clienteID); err != nil {
		return entity.Cliente{}, err
	}
	return c, nil
}

func (p Precondicoes) consentimento(ctx context.Context, clienteID uuid.UUID) error {
	ok, err := p.validador.ConsentimentoValido(ctx, clienteID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%w: %s", errs.ErrConsentimentoInvalido, clienteID)
	}
	return nil
}
