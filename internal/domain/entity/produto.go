package entity

import (
	"fmt"

	"github.com/google/uuid"

	"perfila/internal/shared/errs"
)

// Produto do catálogo. Spec: 3.2, D-05. Fonte: UML2.
type Produto struct {
	ID        uuid.UUID
	Nome      string
	Categoria string
	Preco     float64 // D-14: float64 no MVP
	Ativo     bool
}

// AtualizarPreco exige preço maior ou igual a zero. Spec: 3.2.
func (p *Produto) AtualizarPreco(preco float64) error {
	if preco < 0 {
		return fmt.Errorf("%w: preço negativo", errs.ErrDadosInvalidos)
	}
	p.Preco = preco
	return nil
}
