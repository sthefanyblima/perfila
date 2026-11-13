package entity

import (
	"github.com/google/uuid"
)

// Produto do catálogo. Spec: 3.2, D-05. Fonte: UML2.
type Produto struct {
	ID        uuid.UUID
	Nome      string
	Categoria string
	Preco     float64 // D-14: float64 no MVP
	Ativo     bool
}

// AtualizarPreco exige preço de 0 a 1.000.000 (v1.6, Q-43). Spec: 3.2, RN-14.
func (p *Produto) AtualizarPreco(preco float64) error {
	if err := validarPreco(preco); err != nil {
		return err
	}
	p.Preco = preco
	return nil
}
