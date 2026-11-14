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

// Conceder cria o registro da finalidade se não existir, aplica a transição e grava registro e
// movimento CONCESSAO juntos. Finalidade fora do enumerado: ErrDadosInvalidos. Spec: UC-08, RN-10, Q-48.
func (uc *GerenciarConsentimentoUseCase) Conceder(ctx context.Context, clienteID uuid.UUID, finalidade entity.Finalidade, versaoTermo string) error {
	if !finalidade.Valida() {
		return fmt.Errorf("%w: finalidade %q", errs.ErrDadosInvalidos, finalidade)
	}
	c, err := uc.carregar(ctx, clienteID, finalidade)
	if err != nil {
		return err
	}
	if err := c.Conceder(finalidade, versaoTermo, uc.clock.Agora()); err != nil {
		return err
	}
	return uc.registrar(ctx, c, entity.MovimentoConcessao)
}

// Revogar aplica CONCEDIDO → REVOGADO na finalidade e grava registro e movimento REVOGACAO juntos;
// a outra finalidade não muda. Spec: UC-08, RN-03, RN-10, Q-48.
func (uc *GerenciarConsentimentoUseCase) Revogar(ctx context.Context, clienteID uuid.UUID, finalidade entity.Finalidade) error {
	if !finalidade.Valida() {
		return fmt.Errorf("%w: finalidade %q", errs.ErrDadosInvalidos, finalidade)
	}
	c, err := uc.carregar(ctx, clienteID, finalidade)
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

// Consultar [ADIÇÃO]: sem registro da finalidade, devolve um consentimento sintético em
// NAO_CONCEDIDO daquela finalidade. Spec: UC-08, Q-48.
func (uc *GerenciarConsentimentoUseCase) Consultar(ctx context.Context, clienteID uuid.UUID, finalidade entity.Finalidade) (entity.ConsentimentoLGPD, error) {
	if !finalidade.Valida() {
		return entity.ConsentimentoLGPD{}, fmt.Errorf("%w: finalidade %q", errs.ErrDadosInvalidos, finalidade)
	}
	return uc.carregar(ctx, clienteID, finalidade)
}

// ConsultarFinalidades [ADIÇÃO] v1.9: um registro por finalidade, na ordem de entity.Finalidades(),
// sintético onde não houver. Spec: UC-08, Q-48.
func (uc *GerenciarConsentimentoUseCase) ConsultarFinalidades(ctx context.Context, clienteID uuid.UUID) ([]entity.ConsentimentoLGPD, error) {
	out := make([]entity.ConsentimentoLGPD, 0, len(entity.Finalidades()))
	for _, f := range entity.Finalidades() {
		c, err := uc.carregar(ctx, clienteID, f)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (uc *GerenciarConsentimentoUseCase) carregar(ctx context.Context, clienteID uuid.UUID, finalidade entity.Finalidade) (entity.ConsentimentoLGPD, error) {
	if _, err := uc.clientes.BuscarPorID(ctx, clienteID); err != nil {
		return entity.ConsentimentoLGPD{}, err
	}
	c, err := uc.consentimentos.BuscarPorCliente(ctx, clienteID, finalidade)
	if errors.Is(err, errs.ErrConsentimentoNaoEncontrado) {
		return entity.NovoConsentimento(uuid.New(), clienteID, finalidade, uc.clock.Agora()), nil
	}
	return c, err
}
