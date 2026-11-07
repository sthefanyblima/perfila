// Package external implementa os gateways de serviços externos do UML8.
// No MVP: consentimento local (sobre o repositório) e integração e-commerce/CRM em stub.
//
// Spec: T-20, 3.6, D-12.
package external

import (
	"context"
	"time"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/domain/service"
)

var (
	_ port.ConsentimentoGateway = (*ConsentimentoLocal)(nil)
	_ port.IntegracaoGateway    = IntegracaoStub{}
)

// ConsentimentoLocal responde ao ConsentimentoGateway usando o ValidadorLGPD local.
type ConsentimentoLocal struct{ validador *service.ValidadorLGPD }

func NovoConsentimentoLocal(v *service.ValidadorLGPD) *ConsentimentoLocal {
	return &ConsentimentoLocal{validador: v}
}

func (g *ConsentimentoLocal) Validar(ctx context.Context, clienteID uuid.UUID) (bool, error) {
	return g.validador.ConsentimentoValido(ctx, clienteID)
}

// IntegracaoStub não busca nada: os eventos entram por ColetarDadosUseCase.Registrar (D-12, Q-03).
type IntegracaoStub struct{}

func (IntegracaoStub) BuscarEventos(context.Context, uuid.UUID, time.Time) ([]entity.EventoComportamental, error) {
	return []entity.EventoComportamental{}, nil
}
