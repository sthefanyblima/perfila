package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/shared/errs"
)

// GerenciarConsentimentoUseCase — UC-08. Não exige consentimento válido, pois é quem o gerencia.
// Spec: 3.5, 3.4.1.
type GerenciarConsentimentoUseCase struct {
	clientes       port.ClienteRepository
	consentimentos port.ConsentimentoRepository
	clock          port.Clock
}

// NovoGerenciarConsentimentoUseCase injeta as dependências (C-08).
func NovoGerenciarConsentimentoUseCase(clientes port.ClienteRepository, consentimentos port.ConsentimentoRepository, clock port.Clock) *GerenciarConsentimentoUseCase {
	return &GerenciarConsentimentoUseCase{clientes: clientes, consentimentos: consentimentos, clock: clock}
}

// Conceder cria o registro se não existir, aplica a transição e persiste. Spec: UC-08.
func (uc *GerenciarConsentimentoUseCase) Conceder(ctx context.Context, clienteID uuid.UUID, finalidade, versaoTermo string) error {
	c, err := uc.carregar(ctx, clienteID)
	if err != nil {
		return err
	}
	if err := c.Conceder(finalidade, versaoTermo, uc.clock.Agora()); err != nil {
		return err
	}
	return uc.consentimentos.Salvar(ctx, c)
}

// Revogar aplica CONCEDIDO → REVOGADO e persiste. Spec: UC-08 RN-03.
func (uc *GerenciarConsentimentoUseCase) Revogar(ctx context.Context, clienteID uuid.UUID) error {
	c, err := uc.carregar(ctx, clienteID)
	if err != nil {
		return err
	}
	if err := c.Revogar(uc.clock.Agora()); err != nil {
		return err
	}
	return uc.consentimentos.Salvar(ctx, c)
}

// Consultar [ADIÇÃO]: sem registro, devolve um consentimento sintético em NAO_CONCEDIDO.
func (uc *GerenciarConsentimentoUseCase) Consultar(ctx context.Context, clienteID uuid.UUID) (entity.ConsentimentoLGPD, error) {
	return uc.carregar(ctx, clienteID)
}

func (uc *GerenciarConsentimentoUseCase) carregar(ctx context.Context, clienteID uuid.UUID) (entity.ConsentimentoLGPD, error) {
	if _, err := uc.clientes.BuscarPorID(ctx, clienteID); err != nil {
		return entity.ConsentimentoLGPD{}, err
	}
	c, err := uc.consentimentos.BuscarPorCliente(ctx, clienteID)
	if errors.Is(err, errs.ErrConsentimentoNaoEncontrado) {
		return entity.NovoConsentimento(uuid.New(), clienteID, uc.clock.Agora()), nil
	}
	return c, err
}
