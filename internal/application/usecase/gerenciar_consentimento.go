package usecase

import (
	"context"
	"errors"
	"fmt"

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

// Conceder cria o registro se não existir, aplica a transição e grava registro e movimento
// CONCESSAO juntos. Spec: UC-08, RN-10.
func (uc *GerenciarConsentimentoUseCase) Conceder(ctx context.Context, clienteID uuid.UUID, finalidade, versaoTermo string) error {
	c, err := uc.carregarParaAlterar(ctx, clienteID)
	if err != nil {
		return err
	}
	if err := c.Conceder(finalidade, versaoTermo, uc.clock.Agora()); err != nil {
		return err
	}
	return uc.registrar(ctx, c, entity.MovimentoConcessao)
}

// Revogar aplica CONCEDIDO → REVOGADO e grava registro e movimento REVOGACAO juntos.
// Spec: UC-08, RN-03, RN-10.
func (uc *GerenciarConsentimentoUseCase) Revogar(ctx context.Context, clienteID uuid.UUID) error {
	c, err := uc.carregarParaAlterar(ctx, clienteID)
	if err != nil {
		return err
	}
	if err := c.Revogar(uc.clock.Agora()); err != nil {
		return err
	}
	return uc.registrar(ctx, c, entity.MovimentoRevogacao)
}

// Historico [ADIÇÃO] devolve os movimentos do cliente em ordem cronológica. Spec: UC-08, RN-10.
func (uc *GerenciarConsentimentoUseCase) Historico(ctx context.Context, clienteID uuid.UUID) ([]entity.MovimentoConsentimento, error) {
	if _, err := uc.clientes.BuscarPorID(ctx, clienteID); err != nil {
		return nil, err
	}
	return uc.consentimentos.ListarMovimentos(ctx, clienteID)
}

// registrar monta o movimento com a origem do contexto; se ela for inválida, nada é gravado.
func (uc *GerenciarConsentimentoUseCase) registrar(ctx context.Context, c entity.ConsentimentoLGPD, tipo entity.TipoMovimento) error {
	origem, operador := origemDe(ctx)
	m, err := entity.NovoMovimento(uuid.New(), c, tipo, origem, operador)
	if err != nil {
		return err
	}
	return uc.consentimentos.Registrar(ctx, c, m)
}

// Consultar [ADIÇÃO]: sem registro, devolve um consentimento sintético em NAO_CONCEDIDO.
func (uc *GerenciarConsentimentoUseCase) Consultar(ctx context.Context, clienteID uuid.UUID) (entity.ConsentimentoLGPD, error) {
	return uc.carregar(ctx, clienteID)
}

// carregarParaAlterar é carregar recusando o cliente eliminado: depois da eliminação, o
// consentimento só serve de prova (v1.9, RN-17).
func (uc *GerenciarConsentimentoUseCase) carregarParaAlterar(ctx context.Context, clienteID uuid.UUID) (entity.ConsentimentoLGPD, error) {
	cli, err := uc.clientes.BuscarPorID(ctx, clienteID)
	if err != nil {
		return entity.ConsentimentoLGPD{}, err
	}
	if cli.EstaEliminado() {
		return entity.ConsentimentoLGPD{}, fmt.Errorf("%w: %s", errs.ErrClienteEliminado, clienteID)
	}
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
