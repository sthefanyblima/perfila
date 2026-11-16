package entity

import (
	"errors"
	"reflect"
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
			if !reflect.DeepEqual(r, esperado) {
				t.Fatalf("registro\n obtido   %+v\n esperado %+v", r, esperado)
			}
		})
	}
	if len(AcoesAuditoria()) != 26 {
		t.Fatalf("RN-13 lista 26 ações (17 da v1.5 e 9 da v1.6), há %d", len(AcoesAuditoria()))
	}
}

// Spec: RN-13 Q-44 — campos alterados: lista fechada, sem repetição, na ordem da lista.
func TestRegistroAuditoriaCampos(t *testing.T) {
	r, err := NovoRegistroAuditoria(uuid.New(), time.Now().UTC(), "caio", AcaoEditarSegmento, AlvoSegmento, uuid.New(),
		ResultadoSucesso, "prioridade", "nome", "prioridade")
	if err != nil || len(r.Campos) != 2 || r.Campos[0] != "nome" || r.Campos[1] != "prioridade" {
		t.Fatalf("campos %v, erro %v", r.Campos, err)
	}
	if r, _ := NovoRegistroAuditoria(uuid.New(), time.Now().UTC(), "caio", AcaoSair, AlvoNenhum, uuid.Nil, ResultadoSucesso); r.Campos != nil {
		t.Errorf("sem campos, esperava nil: %v", r.Campos)
	}
	if _, err := NovoRegistroAuditoria(uuid.New(), time.Now().UTC(), "caio", AcaoEditarProduto, AlvoProduto, uuid.New(),
		ResultadoSucesso, "email"); !errors.Is(err, errs.ErrDadosInvalidos) {
		t.Errorf("campo fora da lista deveria ser recusado: %v", err)
	}
}
