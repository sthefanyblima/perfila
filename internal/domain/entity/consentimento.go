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
	Finalidade        Finalidade          // v1.9: enumerada e parte da chave (Q-48)
	VersaoTermo       string
	DataConsentimento time.Time
	DataRevogacao     *time.Time // [ADIÇÃO]
	DataAtualizacao   time.Time  // [ADIÇÃO]
}

// NovoConsentimento cria o registro de uma finalidade no estado inicial NAO_CONCEDIDO.
// Spec: 3.4.1, D-02, Q-48.
func NovoConsentimento(id, clienteID uuid.UUID, finalidade Finalidade, agora time.Time) ConsentimentoLGPD {
	return ConsentimentoLGPD{
		ID:              id,
		ClienteID:       clienteID,
		Finalidade:      finalidade,
		Status:          StatusNaoConcedido,
		DataAtualizacao: agora.UTC(),
	}
}

// Conceder aplica as transições NAO_CONCEDIDO→CONCEDIDO e REVOGADO→CONCEDIDO.
// A transição é checada antes da guarda; a guarda exige a finalidade do registro (dentro do
// enumerado) e a versão do termo. Spec: 3.4.1, Q-48.
func (c *ConsentimentoLGPD) Conceder(finalidade Finalidade, versaoTermo string, agora time.Time) error {
	if c.Status != StatusNaoConcedido && c.Status != StatusRevogado {
		return fmt.Errorf("%w: conceder a partir de %s", errs.ErrTransicaoInvalida, c.Status)
	}
	versaoTermo = strings.TrimSpace(versaoTermo)
	if !finalidade.Valida() || finalidade != c.Finalidade {
		return fmt.Errorf("%w: finalidade %q não é a deste registro", errs.ErrDadosInvalidos, finalidade)
	}
	if versaoTermo == "" {
		return fmt.Errorf("%w: a versão do termo é obrigatória", errs.ErrDadosInvalidos)
	}
	agora = agora.UTC()
	c.Status = StatusConcedido
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
