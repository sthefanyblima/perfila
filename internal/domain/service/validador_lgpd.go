package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/shared/errs"
)

// ValidadorLGPD responde se o cliente tem consentimento válido para uma finalidade.
// Spec: RN-02, C-03, D-02, Q-48.
type ValidadorLGPD struct {
	consentimentos port.ConsentimentoRepository
}

// NovoValidadorLGPD injeta o repositório de consentimentos (C-08).
func NovoValidadorLGPD(consentimentos port.ConsentimentoRepository) *ValidadorLGPD {
	return &ValidadorLGPD{consentimentos: consentimentos}
}

// ConsentimentoValido confere a finalidade segmentacao, a que libera UC-01 a UC-04. Spec: RN-01.
func (v *ValidadorLGPD) ConsentimentoValido(ctx context.Context, clienteID uuid.UUID) (bool, error) {
	return v.ConsentimentoValidoPara(ctx, clienteID, entity.FinalidadeSegmentacao)
}

// ConsentimentoValidoPara [ADIÇÃO] v1.9: ausência de registro da finalidade equivale a
// NAO_CONCEDIDO (falso, sem erro). Spec: RN-02, Q-48.
func (v *ValidadorLGPD) ConsentimentoValidoPara(ctx context.Context, clienteID uuid.UUID, f entity.Finalidade) (bool, error) {
	c, err := v.consentimentos.BuscarPorCliente(ctx, clienteID, f)
	if errors.Is(err, errs.ErrConsentimentoNaoEncontrado) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return c.EstaValido(), nil
}
