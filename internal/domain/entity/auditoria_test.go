package entity

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"perfila/internal/shared/errs"
)

// Spec: 3.2 RN-13 Q-39 — enumerados e par tipo/ID do alvo.
func TestNovoRegistroAuditoria(t *testing.T) {
	data := time.Date(2026, 3, 10, 9, 0, 0, 0, time.FixedZone("BRT", -3*3600))
	alvo := uuid.New()
	casos := []struct {
		nome      string
		acao      AcaoAuditoria
		tipo      TipoAlvo
		id        uuid.UUID
		resultado ResultadoAuditoria
		erro      error
	}{
		{"ficha", AcaoLerFicha, AlvoCliente, alvo, ResultadoSucesso, nil},
		{"entrar sem alvo", AcaoEntrar, AlvoNenhum, uuid.Nil, ResultadoNegado, nil},
		{"operador", AcaoCriarOperador, AlvoOperador, alvo, ResultadoFalha, nil},
		{"segmento", AcaoAcessoNegado, AlvoSegmento, alvo, ResultadoNegado, nil},
		{"produto", AcaoAcessoNegado, AlvoProduto, alvo, ResultadoNegado, nil},
		{"ação desconhecida", AcaoAuditoria("listar_clientes"), AlvoNenhum, uuid.Nil, ResultadoSucesso, errs.ErrDadosInvalidos},
		{"tipo desconhecido", AcaoLerFicha, TipoAlvo("email"), alvo, ResultadoSucesso, errs.ErrDadosInvalidos},
		{"tipo sem ID", AcaoLerFicha, AlvoCliente, uuid.Nil, ResultadoSucesso, errs.ErrDadosInvalidos},
		{"ID sem tipo", AcaoEntrar, AlvoNenhum, alvo, ResultadoSucesso, errs.ErrDadosInvalidos},
		{"resultado desconhecido", AcaoSair, AlvoNenhum, uuid.Nil, ResultadoAuditoria("OK"), errs.ErrDadosInvalidos},
	}
	for _, tc := range casos {
		t.Run(tc.nome, func(t *testing.T) {
			id := uuid.New()
			r, err := NovoRegistroAuditoria(id, data, " Admin ", tc.acao, tc.tipo, tc.id, tc.resultado)
			if !errors.Is(err, tc.erro) {
				t.Fatalf("erro %v, esperado %v", err, tc.erro)
			}
			if tc.erro != nil {
				return
			}
			esperado := RegistroAuditoria{ID: id, Data: data.UTC(), Operador: "admin", Acao: tc.acao,
				AlvoTipo: tc.tipo, AlvoID: tc.id, Resultado: tc.resultado}
			if r != esperado {
				t.Fatalf("registro\n obtido   %+v\n esperado %+v", r, esperado)
			}
		})
	}
	if len(AcoesAuditoria()) != 17 {
		t.Fatalf("RN-13 lista 17 ações, há %d", len(AcoesAuditoria()))
	}
}
