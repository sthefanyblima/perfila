package usecase

import (
	"context"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
)

// CadastrarClienteUseCase — UC-09 [ADIÇÃO, D-22]. Ponto de entrada do cliente no sistema.
// Como UC-08, não exige consentimento: o cliente nasce com consentimento NAO_CONCEDIDO
// (sem registro) e nada além de nome e e-mail é tratado até ele conceder. Spec: 3.5.
type CadastrarClienteUseCase struct {
	clientes port.ClienteRepository
	clock    port.Clock
}

// NovoCadastrarClienteUseCase injeta as dependências (C-08).
func NovoCadastrarClienteUseCase(clientes port.ClienteRepository, clock port.Clock) *CadastrarClienteUseCase {
	return &CadastrarClienteUseCase{clientes: clientes, clock: clock}
}

// Executar valida nome e e-mail, cria o cliente ativo e o persiste. Spec: UC-09.
func (uc *CadastrarClienteUseCase) Executar(ctx context.Context, nome, email string) (entity.Cliente, error) {
	c, err := entity.NovoCliente(uuid.New(), nome, email, uc.clock.Agora())
	if err != nil {
		return entity.Cliente{}, err
	}
	if err := uc.clientes.Salvar(ctx, c); err != nil {
		return entity.Cliente{}, err
	}
	return c, nil
}
