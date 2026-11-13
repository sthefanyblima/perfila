package entity

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"perfila/internal/shared/errs"
)

func segmentosBase() []Segmento {
	return []Segmento{
		{ID: uuid.New(), Nome: "premium", Criterio: "ticketMedio >= 500", Prioridade: 1, Ativo: true},
		{ID: uuid.New(), Nome: "engajado", Criterio: "scoreEngajamento >= 0.7", Prioridade: 2, Ativo: true},
		{ID: uuid.New(), Nome: "antigo", Criterio: "frequenciaCompras > 9", Prioridade: 2, Ativo: false},
		{ID: uuid.New(), Nome: "padrao", Prioridade: 99, Ativo: true},
	}
}

func campoDoErro(t *testing.T, err error) string {
	t.Helper()
	var ec *errs.ErroCampo
	if !errors.As(err, &ec) || !errors.Is(err, errs.ErrDadosInvalidos) {
		t.Fatalf("esperava ErroCampo com ErrDadosInvalidos, obtido %v", err)
	}
	return ec.Campo
}

// Spec: RN-14 Q-33 — regras do conjunto de segmentos (teste de tabela).
func TestValidarSegmentos(t *testing.T) {
	if err := ValidarSegmentos(segmentosBase()); err != nil {
		t.Fatalf("configuração base deveria ser válida: %v", err)
	}
	casos := []struct {
		nome  string
		mudar func(l []Segmento) []Segmento
		campo string
	}{
		{"padrao desativado", func(l []Segmento) []Segmento { l[3].Ativo = false; return l }, "ativo"},
		{"padrao com critério", func(l []Segmento) []Segmento { l[3].Criterio = "ticketMedio > 0"; return l }, "criterio"},
		{"segundo critério vazio", func(l []Segmento) []Segmento { l[1].Criterio = " "; return l }, "criterio"},
		{"prioridade repetida entre ativos", func(l []Segmento) []Segmento { l[1].Prioridade = 1; return l }, "prioridade"},
		{"inativo ativado com prioridade ocupada", func(l []Segmento) []Segmento { l[2].Ativo = true; return l }, "prioridade"},
		{"acima do padrao", func(l []Segmento) []Segmento { l[1].Prioridade = 99; return l }, "prioridade"},
		{"nome repetido sem diferenciar maiúsculas", func(l []Segmento) []Segmento { l[1].Nome = "Premium"; return l }, "nome"},
		{"nome com ponto e vírgula", func(l []Segmento) []Segmento { l[1].Nome = "a;b"; return l }, "nome"},
		{"nome com dois-pontos", func(l []Segmento) []Segmento { l[1].Nome = "a:b"; return l }, "nome"},
		{"nome vazio", func(l []Segmento) []Segmento { l[1].Nome = "  "; return l }, "nome"},
		{"nome longo", func(l []Segmento) []Segmento { l[1].Nome = strings.Repeat("a", 41); return l }, "nome"},
		{"descrição longa", func(l []Segmento) []Segmento { l[1].Descricao = strings.Repeat("a", 201); return l }, "descricao"},
		{"prioridade zero", func(l []Segmento) []Segmento { l[1].Prioridade = 0; return l }, "prioridade"},
		{"prioridade 1000", func(l []Segmento) []Segmento { l[3].Prioridade = 1000; return l }, "prioridade"},
		{"critério fora da gramática", func(l []Segmento) []Segmento { l[1].Criterio = "idade > 18"; return l }, "criterio"},
		{"sem padrao", func(l []Segmento) []Segmento { return l[:3] }, "criterio"},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			if got := campoDoErro(t, ValidarSegmentos(c.mudar(segmentosBase()))); got != c.campo {
				t.Errorf("campo %q, esperava %q", got, c.campo)
			}
		})
	}
	// Limites aceitos: nome com 40 caracteres, prioridade 1 e 999, inativo repetindo prioridade.
	l := segmentosBase()
	l[0].Nome, l[3].Prioridade = strings.Repeat("á", 40), 999
	if err := ValidarSegmentos(l); err != nil {
		t.Errorf("limites deveriam ser aceitos: %v", err)
	}
}

// Spec: RN-14 — a mensagem cita a condição errada.
func TestValidarCriterioMensagem(t *testing.T) {
	casos := map[string]string{
		"idade > 18":                            "idade > 18",
		"scoreEngajamento >= 0.5 && ticket > 1": "ticket > 1",
		"scoreEngajamento >= 0.5 &&":            "&&",
		"ticketMedio >> 1":                      "ticketMedio >> 1",
	}
	for criterio, cita := range casos {
		err := ValidarCriterio(criterio)
		var ec *errs.ErroCampo
		if !errors.As(err, &ec) || !strings.Contains(ec.Mensagem, cita) {
			t.Errorf("%q: esperava mensagem citando %q, obtido %v", criterio, cita, err)
		}
	}
	if err := ValidarCriterio(""); err != nil {
		t.Errorf("critério vazio é válido: %v", err)
	}
}

// Spec: RN-14 Q-43 — produto.
func TestProdutoValidar(t *testing.T) {
	base := Produto{ID: uuid.New(), Nome: "Atlas", Categoria: "livros", Preco: 90, Ativo: true}
	if err := base.Validar(); err != nil {
		t.Fatalf("produto base deveria ser válido: %v", err)
	}
	casos := []struct {
		nome  string
		mudar func(p *Produto)
		campo string
	}{
		{"nome vazio", func(p *Produto) { p.Nome = " " }, "nome"},
		{"nome longo", func(p *Produto) { p.Nome = strings.Repeat("a", 101) }, "nome"},
		{"categoria vazia", func(p *Produto) { p.Categoria = "" }, "categoria"},
		{"categoria longa", func(p *Produto) { p.Categoria = strings.Repeat("a", 41) }, "categoria"},
		{"preço negativo", func(p *Produto) { p.Preco = -1 }, "preco"},
		{"preço acima do teto", func(p *Produto) { p.Preco = 1_000_000.01 }, "preco"},
	}
	for _, c := range casos {
		p := base
		c.mudar(&p)
		if got := campoDoErro(t, p.Validar()); got != c.campo {
			t.Errorf("%s: campo %q, esperava %q", c.nome, got, c.campo)
		}
	}
	for _, preco := range []float64{0, 1_000_000} {
		p := base
		if err := p.AtualizarPreco(preco); err != nil || p.Preco != preco {
			t.Errorf("preço %v deveria ser aceito: %v", preco, err)
		}
	}
	p := base
	if err := p.AtualizarPreco(1_000_001); err == nil || p.Preco != 90 {
		t.Errorf("preço acima do teto deveria ser recusado sem mudar o produto")
	}
	if got := NormalizarCategoria("  Livros "); got != "livros" {
		t.Errorf("NormalizarCategoria = %q", got)
	}
}
