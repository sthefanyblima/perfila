package entity

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"

	"perfila/internal/shared/errs"
)

var agoraTeste = time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)

// Spec: RN-06 — tabela de critérios.
func TestAvaliarCriterio(t *testing.T) {
	p := PerfilCliente{FrequenciaCompras: 3, TicketMedio: 500, ScoreEngajamento: 0.7}
	casos := []struct {
		criterio string
		casa     bool
		erro     bool
	}{
		{"", true, false},
		{"   ", true, false},
		{"ticketMedio >= 500", true, false},
		{"ticketMedio > 500", false, false},
		{"ticketMedio>=500&&frequenciaCompras==3", true, false},
		{"scoreEngajamento < 0.7", false, false},
		{"scoreEngajamento <= 0.7 && ticketMedio < 1", false, false},
		{"frequenciaCompras > 1 && idade > 1", false, true}, // valida todas as condições
		{"idade > 18", false, true},
		{"ticketMedio >> 1", false, true},
		{"ticketMedio > abc", false, true},
		{"ticketMedio >", false, true},
		{"ticketMedio > 1 &&", false, true},
		{"ticketMedio > -1", false, true},
		{"ticketMedio > 1,5", false, true},
		{"ticketMedio || 1", false, true},
	}
	for _, c := range casos {
		casa, err := AvaliarCriterio(c.criterio, p)
		if c.erro {
			if !errors.Is(err, errs.ErrDadosInvalidos) {
				t.Errorf("%q: esperava ErrDadosInvalidos, obtido %v", c.criterio, err)
			}
			continue
		}
		if err != nil || casa != c.casa {
			t.Errorf("%q: esperado (%v, nil), obtido (%v, %v)", c.criterio, c.casa, casa, err)
		}
	}
}

// Spec: 3.4.1 — todas as combinações estado × evento.
func TestConsentimentoTransicoes(t *testing.T) {
	estados := []StatusConsentimento{StatusNaoConcedido, StatusConcedido, StatusRevogado}
	for _, e := range estados {
		c := ConsentimentoLGPD{Status: e}
		errC := c.Conceder("segmentacao", "v1", agoraTeste)
		querOK := e != StatusConcedido
		if (errC == nil) != querOK {
			t.Errorf("Conceder a partir de %s: erro=%v", e, errC)
		}
		c = ConsentimentoLGPD{Status: e}
		errR := c.Revogar(agoraTeste)
		if (errR == nil) != (e == StatusConcedido) {
			t.Errorf("Revogar a partir de %s: erro=%v", e, errR)
		}
		if errR != nil && !errors.Is(errR, errs.ErrTransicaoInvalida) {
			t.Errorf("Revogar inválido deveria ser ErrTransicaoInvalida: %v", errR)
		}
	}
}

// Spec: 3.4.2 — todas as combinações estado × evento.
func TestRecomendacaoTransicoes(t *testing.T) {
	nova := func() RecomendacaoProduto {
		return NovaRecomendacao(uuid.New(), uuid.New(), uuid.New(), "m", agoraTeste)
	}
	r := nova()
	if err := r.Expirar(agoraTeste); !errors.Is(err, errs.ErrTransicaoInvalida) {
		t.Fatalf("GERADA→Expirar deveria falhar: %v", err)
	}
	if err := r.Disponibilizar(agoraTeste, 0); !errors.Is(err, errs.ErrDadosInvalidos) {
		t.Fatalf("validade zero deveria falhar: %v", err)
	}
	if err := r.Disponibilizar(agoraTeste, 30); err != nil || r.Status != StatusAtiva {
		t.Fatalf("GERADA→ATIVA: %v %s", err, r.Status)
	}
	if want := agoraTeste.AddDate(0, 0, 30); !r.Validade.Equal(want) {
		t.Fatalf("validade %v, esperado %v", r.Validade, want)
	}
	if err := r.Disponibilizar(agoraTeste, 30); !errors.Is(err, errs.ErrTransicaoInvalida) {
		t.Fatalf("ATIVA→Disponibilizar deveria falhar: %v", err)
	}
	if err := r.Expirar(agoraTeste); err != nil || r.Status != StatusExpirada {
		t.Fatalf("ATIVA→EXPIRADA: %v", err)
	}
	if err := r.Expirar(agoraTeste); !errors.Is(err, errs.ErrTransicaoInvalida) {
		t.Fatalf("EXPIRADA é final: %v", err)
	}
}

// Spec: RN-04.
func TestEventoValidar(t *testing.T) {
	base := EventoComportamental{ClienteID: uuid.New(), Tipo: TipoCompra, Origem: "web", Timestamp: agoraTeste, Valor: 1}
	if err := base.Validar(agoraTeste); err != nil {
		t.Fatalf("evento válido rejeitado: %v", err)
	}
	mut := []func(e *EventoComportamental){
		func(e *EventoComportamental) { e.ClienteID = uuid.Nil },
		func(e *EventoComportamental) { e.Tipo = "x" },
		func(e *EventoComportamental) { e.Origem = " " },
		func(e *EventoComportamental) { e.Timestamp = time.Time{} },
		func(e *EventoComportamental) { e.Timestamp = agoraTeste.Add(ToleranciaRelogio + time.Second) }, // Q-16
		func(e *EventoComportamental) { e.Valor = 0 },
		func(e *EventoComportamental) { e.Tipo = TipoNavegacao; e.Valor = -1 },
		func(e *EventoComportamental) { e.Valor = ValorMaximoEvento + 0.01 }, // Q-15
		func(e *EventoComportamental) { e.Valor = math.Inf(1) },
		func(e *EventoComportamental) { e.Valor = math.NaN() },
	}
	// Limites aceitos: tolerância de relógio e valor máximo (Q-15, Q-16).
	for i, f := range []func(e *EventoComportamental){
		func(e *EventoComportamental) { e.Timestamp = agoraTeste.Add(ToleranciaRelogio) },
		func(e *EventoComportamental) { e.Valor = ValorMaximoEvento },
	} {
		e := base
		f(&e)
		if err := e.Validar(agoraTeste); err != nil {
			t.Errorf("limite %d rejeitado: %v", i, err)
		}
	}
	for i, f := range mut {
		e := base
		f(&e)
		if err := e.Validar(agoraTeste); !errors.Is(err, errs.ErrDadosInvalidos) {
			t.Errorf("mutação %d: esperava ErrDadosInvalidos, obtido %v", i, err)
		}
	}
}

// Spec: RN-05 — janela (agora-30d, agora] e teto 1.
func TestCalcularScore(t *testing.T) {
	ev := func(d time.Duration) EventoComportamental { return EventoComportamental{Timestamp: agoraTeste.Add(-d)} }
	dia := 24 * time.Hour
	casos := []struct {
		eventos []EventoComportamental
		meta    int
		quer    float64
	}{
		{nil, 10, 0},
		{[]EventoComportamental{ev(29 * dia)}, 10, 0.1},
		{[]EventoComportamental{ev(30 * dia)}, 10, 0},
		{[]EventoComportamental{ev(-time.Second)}, 10, 0}, // futuro não conta
		{[]EventoComportamental{ev(0), ev(dia), ev(dia)}, 2, 1},
		{[]EventoComportamental{ev(0)}, 0, 0},
	}
	for i, c := range casos {
		var p PerfilCliente
		p.CalcularScore(c.eventos, agoraTeste, c.meta)
		if p.ScoreEngajamento != c.quer {
			t.Errorf("caso %d: esperado %v, obtido %v", i, c.quer, p.ScoreEngajamento)
		}
	}
}

func TestClienteEProduto(t *testing.T) {
	c := Cliente{Ativo: true}
	if err := c.AtualizarDados("", "a@b"); !errors.Is(err, errs.ErrDadosInvalidos) {
		t.Error("nome vazio deveria falhar")
	}
	if err := c.AtualizarDados(" Ana ", "a@b"); err != nil || c.Nome != "Ana" {
		t.Errorf("atualização válida: %v %q", err, c.Nome)
	}
	c.Desativar()
	if c.Ativo {
		t.Error("Desativar")
	}
	c.Ativar()
	if !c.Ativo {
		t.Error("Ativar")
	}
	p := Produto{}
	if err := p.AtualizarPreco(-1); !errors.Is(err, errs.ErrDadosInvalidos) {
		t.Error("preço negativo deveria falhar")
	}
	if err := p.AtualizarPreco(0); err != nil {
		t.Error("preço zero é válido")
	}
}

// Spec: UC-09, D-22.
func TestNovoCliente(t *testing.T) {
	c, err := NovoCliente(uuid.New(), " Ana ", "a@exemplo.test", agoraTeste)
	if err != nil || !c.Ativo || c.Nome != "Ana" || !c.DataCadastro.Equal(agoraTeste) {
		t.Fatalf("cliente válido: %+v %v", c, err)
	}
	if _, err := NovoCliente(uuid.New(), "", "a@exemplo.test", agoraTeste); !errors.Is(err, errs.ErrDadosInvalidos) {
		t.Fatalf("nome vazio deveria falhar: %v", err)
	}
}
