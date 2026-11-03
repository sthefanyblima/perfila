package entity

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"perfila/internal/shared/errs"
)

// ConsentimentoLGPD registra a autorização do cliente para o tratamento dos seus dados.
// Spec: 3.2, 3.4.1, D-02, D-03. Fonte: UML2 e UML5.
type ConsentimentoLGPD struct {
	ID                uuid.UUID
	ClienteID         uuid.UUID
	Status            StatusConsentimento // [ADIÇÃO] substitui Concedido bool (D-02)
	Finalidade        string
	VersaoTermo       string
	DataConsentimento time.Time
	DataRevogacao     *time.Time // [ADIÇÃO]
	DataAtualizacao   time.Time  // [ADIÇÃO]
}

// NovoConsentimento cria um registro no estado inicial NAO_CONCEDIDO. Spec: 3.4.1, D-02.
func NovoConsentimento(id, clienteID uuid.UUID, agora time.Time) ConsentimentoLGPD {
	return ConsentimentoLGPD{
		ID:              id,
		ClienteID:       clienteID,
		Status:          StatusNaoConcedido,
		DataAtualizacao: agora.UTC(),
	}
}

// Conceder aplica as transições NAO_CONCEDIDO→CONCEDIDO e REVOGADO→CONCEDIDO.
// A transição é checada antes da guarda. Spec: 3.4.1.
func (c *ConsentimentoLGPD) Conceder(finalidade, versaoTermo string, agora time.Time) error {
	if c.Status != StatusNaoConcedido && c.Status != StatusRevogado {
		return fmt.Errorf("%w: conceder a partir de %s", errs.ErrTransicaoInvalida, c.Status)
	}
	finalidade, versaoTermo = strings.TrimSpace(finalidade), strings.TrimSpace(versaoTermo)
	if finalidade == "" || versaoTermo == "" {
		return fmt.Errorf("%w: finalidade e versão do termo são obrigatórias", errs.ErrDadosInvalidos)
	}
	agora = agora.UTC()
	c.Status = StatusConcedido
	c.Finalidade = finalidade
	c.VersaoTermo = versaoTermo
	c.DataConsentimento = agora
	c.DataAtualizacao = agora
	return nil
}

// Revogar aplica CONCEDIDO→REVOGADO; o efeito é imediato (RN-03). Spec: 3.4.1.
func (c *ConsentimentoLGPD) Revogar(agora time.Time) error {
	if c.Status != StatusConcedido {
		return fmt.Errorf("%w: revogar a partir de %s", errs.ErrTransicaoInvalida, c.Status)
	}
	agora = agora.UTC()
	c.Status = StatusRevogado
	c.DataRevogacao = &agora
	c.DataAtualizacao = agora
	return nil
}

// EstaValido é consulta, não transição: válido se e somente se CONCEDIDO. Spec: RN-02, D-03.
func (c ConsentimentoLGPD) EstaValido() bool { return c.Status == StatusConcedido }
