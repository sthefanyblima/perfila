package entity

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"perfila/internal/shared/errs"
)

// Cliente é o titular dos dados. Spec: 3.2. Fonte: UML2.
type Cliente struct {
	ID           uuid.UUID
	Nome         string
	Email        string
	DataCadastro time.Time
	Ativo        bool
	// DataEliminacao [ADIÇÃO v1.9] é preenchida quando o cadastro é anonimizado. Spec: RN-17, Q-50.
	DataEliminacao *time.Time
}

// Valores fixos do cadastro anonimizado pela eliminação. Spec: RN-17, Q-50.
const (
	NomeEliminado  = "Titular eliminado"
	EmailEliminado = "eliminado@anonimizado.invalid"
)

// NovoCliente cria um cliente ativo, com DataCadastro = agora. Nome e e-mail são
// obrigatórios (mesma regra de AtualizarDados). Spec: UC-09, D-22.
func NovoCliente(id uuid.UUID, nome, email string, agora time.Time) (Cliente, error) {
	c := Cliente{ID: id, DataCadastro: agora.UTC(), Ativo: true}
	if err := c.AtualizarDados(nome, email); err != nil {
		return Cliente{}, err
	}
	return c, nil
}

// AtualizarDados troca nome e e-mail; nenhum dos dois pode ser vazio. Cadastro eliminado não
// muda mais. Spec: 3.2, Q-48, RN-17.
func (c *Cliente) AtualizarDados(nome, email string) error {
	if c.EstaEliminado() {
		return fmt.Errorf("%w: cliente %s", errs.ErrClienteEliminado, c.ID)
	}
	nome, email = strings.TrimSpace(nome), strings.TrimSpace(email)
	if nome == "" || email == "" {
		return fmt.Errorf("%w: nome e e-mail são obrigatórios", errs.ErrDadosInvalidos)
	}
	c.Nome, c.Email = nome, email
	return nil
}

// Ativar marca o cliente como ativo. Spec: 3.2.
func (c *Cliente) Ativar() { c.Ativo = true }

// Desativar marca o cliente como inativo; UC-01 a UC-04 passam a falhar. Spec: RN-01.
func (c *Cliente) Desativar() { c.Ativo = false }

// Anonimizar troca nome e e-mail pelos valores fixos, desativa e marca DataEliminacao = agora.
// É irreversível: num cadastro já eliminado, ErrClienteEliminado. Spec: RN-17, Q-50.
func (c *Cliente) Anonimizar(agora time.Time) error {
	if c.EstaEliminado() {
		return fmt.Errorf("%w: cliente %s", errs.ErrClienteEliminado, c.ID)
	}
	quando := agora.UTC()
	c.Nome, c.Email, c.Ativo, c.DataEliminacao = NomeEliminado, EmailEliminado, false, &quando
	return nil
}

// EstaEliminado diz se o cadastro foi anonimizado pela eliminação. Spec: RN-17.
func (c Cliente) EstaEliminado() bool { return c.DataEliminacao != nil }
