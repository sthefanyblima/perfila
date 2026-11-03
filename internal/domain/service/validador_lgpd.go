package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"perfila/internal/domain/port"
	"perfila/internal/shared/errs"
)

// ValidadorLGPD responde se o cliente tem consentimento válido. Spec: RN-02, C-03, D-02.
type ValidadorLGPD struct {
	consentimentos port.ConsentimentoRepository
}

// NovoValidadorLGPD injeta o repositório de consentimentos (C-08).
func NovoValidadorLGPD(consentimentos port.ConsentimentoRepository) *ValidadorLGPD {
	return &ValidadorLGPD{consentimentos: consentimentos}
}

// ConsentimentoValido: ausência de registro equivale a NAO_CONCEDIDO (falso, sem erro).
func (v *ValidadorLGPD) ConsentimentoValido(ctx context.Context, clienteID uuid.UUID) (bool, error) {
	c, err := v.consentimentos.BuscarPorCliente(ctx, clienteID)
	if errors.Is(err, errs.ErrConsentimentoNaoEncontrado) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return c.EstaValido(), nil
}
