package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"perfila/internal/shared/errs"
)

// RecomendacaoProduto liga um cliente a um produto sugerido. Spec: 3.2, 3.4.2, RN-08, D-08.
// Fonte: UML2 e UML6.
type RecomendacaoProduto struct {
	ID            uuid.UUID
	ClienteID     uuid.UUID
	ProdutoID     uuid.UUID
	Motivo        string
	Status        StatusRecomendacao // [ADIÇÃO] (D-08)
	DataGeracao   time.Time
	DataAtivacao  *time.Time // [ADIÇÃO]
	Validade      *time.Time // [ADIÇÃO]
	DataExpiracao *time.Time // [ADIÇÃO]
}

// NovaRecomendacao equivale a Gerar(): (início) → GERADA. Spec: 3.4.2.
func NovaRecomendacao(id, clienteID, produtoID uuid.UUID, motivo string, agora time.Time) RecomendacaoProduto {
	return RecomendacaoProduto{
		ID:          id,
		ClienteID:   clienteID,
		ProdutoID:   produtoID,
		Motivo:      motivo,
		Status:      StatusGerada,
		DataGeracao: agora.UTC(),
	}
}

// Disponibilizar aplica GERADA → ATIVA, com Validade = agora + validadeDias. Spec: 3.4.2, RN-08.
func (r *RecomendacaoProduto) Disponibilizar(agora time.Time, validadeDias int) error {
	if r.Status != StatusGerada {
		return fmt.Errorf("%w: disponibilizar a partir de %s", errs.ErrTransicaoInvalida, r.Status)
	}
	if validadeDias <= 0 {
		return fmt.Errorf("%w: validade deve ser positiva", errs.ErrDadosInvalidos)
	}
	agora = agora.UTC()
	validade := agora.AddDate(0, 0, validadeDias)
	r.Status = StatusAtiva
	r.DataAtivacao = &agora
	r.Validade = &validade
	return nil
}

// Expirar aplica ATIVA → EXPIRADA. Spec: 3.4.2.
func (r *RecomendacaoProduto) Expirar(agora time.Time) error {
	if r.Status != StatusAtiva {
		return fmt.Errorf("%w: expirar a partir de %s", errs.ErrTransicaoInvalida, r.Status)
	}
	agora = agora.UTC()
	r.Status = StatusExpirada
	r.DataExpiracao = &agora
	return nil
}
