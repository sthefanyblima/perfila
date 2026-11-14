package usecase

import (
	"context"
	"errors"
	"testing"

	"perfila/internal/shared/errs"
)

// Spec: RN-17 — a confirmação é o nome do cadastro, sem diferenciar maiúsculas, acentos e espaços
// nas pontas; vazia nunca confere.
func TestConfirmacaoConfere(t *testing.T) {
	casos := []struct {
		confirmacao, nome string
		quer              bool
	}{
		{"Ana Sintética", "Ana Sintética", true},
		{"  ana sintetica ", "Ana Sintética", true},
		{"ANA SINTÉTICA", "Ana Sintética", true},
		{"João Conceição", "joao conceicao", true},
		{"Ana", "Ana Sintética", false},
		{"Ana  Sintética", "Ana Sintética", false}, // espaço no meio conta
		{"", "Ana Sintética", false},
		{"   ", "", false},
	}
	for _, c := range casos {
		if got := ConfirmacaoConfere(c.confirmacao, c.nome); got != c.quer {
			t.Errorf("ConfirmacaoConfere(%q, %q) = %v, esperado %v", c.confirmacao, c.nome, got, c.quer)
		}
	}
}

// Spec: RN-18, Q-51 — prazo menor que um dia é configuração inválida e não apaga nada.
func TestRetencaoPrazoInvalido(t *testing.T) {
	for _, prazo := range []int{0, -1} {
		uc := NovoAplicarRetencaoUseCase(nil, nil, nil, nil, nil, prazo)
		if _, err := uc.Executar(context.Background()); !errors.Is(err, errs.ErrDadosInvalidos) {
			t.Errorf("prazo %d: esperado dados_invalidos, obtido %v", prazo, err)
		}
	}
}
