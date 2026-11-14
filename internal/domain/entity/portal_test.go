package entity

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"perfila/internal/shared/errs"
)

var agoraPortal = time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)

// Spec: RN-19, Q-56 — formato do hash, validade de 30 minutos e uso único.
func TestNovoTokenAcesso(t *testing.T) {
	hash := strings.Repeat("ab", 32)
	casos := []struct {
		nome      string
		clienteID uuid.UUID
		hash      string
		erro      error
	}{
		{"válido", uuid.New(), hash, nil},
		{"sem cliente", uuid.Nil, hash, errs.ErrDadosInvalidos},
		{"hash curto", uuid.New(), "abc", errs.ErrDadosInvalidos},
		{"hash com maiúsculas", uuid.New(), strings.Repeat("AB", 32), errs.ErrDadosInvalidos},
		{"hash não hexadecimal", uuid.New(), strings.Repeat("zz", 32), errs.ErrDadosInvalidos},
	}
	for _, c := range casos {
		tok, err := NovoTokenAcesso(uuid.New(), c.clienteID, c.hash, agoraPortal)
		if !errors.Is(err, c.erro) && !(c.erro == nil && err == nil) {
			t.Errorf("%s: erro %v, esperado %v", c.nome, err, c.erro)
		}
		if c.erro == nil && !tok.ExpiraEm.Equal(agoraPortal.Add(30*time.Minute)) {
			t.Errorf("%s: expira em %v", c.nome, tok.ExpiraEm)
		}
	}
}

// Spec: RN-19 — vale antes de ExpiraEm e uma vez só (limites 29, 30 e 31 minutos).
func TestTokenAcessoUsar(t *testing.T) {
	for _, c := range []struct {
		minutos int
		vale    bool
	}{{0, true}, {29, true}, {30, false}, {31, false}} {
		tok, _ := NovoTokenAcesso(uuid.New(), uuid.New(), strings.Repeat("0f", 32), agoraPortal)
		quando := agoraPortal.Add(time.Duration(c.minutos) * time.Minute)
		if tok.EstaValido(quando) != c.vale {
			t.Errorf("%d min: válido = %v", c.minutos, !c.vale)
		}
		err := tok.Usar(quando)
		if c.vale && (err != nil || tok.UsadoEm == nil || !tok.UsadoEm.Equal(quando)) {
			t.Errorf("%d min: primeiro uso falhou: %v", c.minutos, err)
		}
		if !c.vale && !errors.Is(err, errs.ErrAcessoInvalido) {
			t.Errorf("%d min: esperado acesso_invalido, obtido %v", c.minutos, err)
		}
		if c.vale && !errors.Is(tok.Usar(quando), errs.ErrAcessoInvalido) {
			t.Errorf("%d min: o segundo uso deveria falhar", c.minutos)
		}
	}
}

// Spec: RN-20 — nasce ABERTO; ABERTO → ATENDIDO; atender de novo é transição inválida.
func TestPedidoEliminacao(t *testing.T) {
	if _, err := NovoPedidoEliminacao(uuid.New(), uuid.Nil, agoraPortal); !errors.Is(err, errs.ErrDadosInvalidos) {
		t.Fatalf("pedido sem cliente: %v", err)
	}
	p, err := NovoPedidoEliminacao(uuid.New(), uuid.New(), agoraPortal)
	if err != nil || p.Status != PedidoAberto || !p.DataPedido.Equal(agoraPortal) || p.DataAtendimento != nil {
		t.Fatalf("pedido novo: %+v, %v", p, err)
	}
	depois := agoraPortal.Add(time.Hour)
	if err := p.Atender(depois); err != nil || p.Status != PedidoAtendido || !p.DataAtendimento.Equal(depois) {
		t.Fatalf("atender: %+v, %v", p, err)
	}
	if err := p.Atender(depois); !errors.Is(err, errs.ErrTransicaoInvalida) {
		t.Fatalf("atender de novo: %v", err)
	}
}
