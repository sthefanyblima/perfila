package web

import (
	"testing"

	"perfila/internal/domain/entity"
)

// Spec: RN-15 — a barra de um grupo oculto tem sempre o mesmo tamanho e não mostra o número.
func TestBarraOcultaNaoRevelaONumero(t *testing.T) {
	oculta := entity.Contagem{Oculta: true}
	a := barrasHorizontais([]dadoBarra{{Rotulo: "a", Valor: oculta}, {Rotulo: "b", Valor: entity.Contagem{Valor: 50}}})
	b := barrasHorizontais([]dadoBarra{{Rotulo: "a", Valor: oculta}, {Rotulo: "b", Valor: entity.Contagem{Valor: 50}}})
	if a.Barras[0].L != b.Barras[0].L || !a.Barras[0].Oculta || a.Barras[0].Valor != "menos de 5" {
		t.Fatalf("barra oculta inesperada: %+v", a.Barras[0])
	}
	if a.Barras[1].L != maximoBarra {
		t.Fatalf("a maior barra visível deveria ocupar o máximo: %v", a.Barras[1].L)
	}
	c := colunas([]grupoColunas{{Rotulo: "jan/26", Series: []dadoBarra{{Valor: oculta}}}}, nil)
	if c.Barras[0].Valor != "<5" || c.Barras[0].A <= 0 {
		t.Fatalf("coluna oculta inesperada: %+v", c.Barras[0])
	}
}

// Spec: 3.7.1 — célula que começa como fórmula ganha um apóstrofo (categoria é texto livre).
func TestCelulaCSVSemFormula(t *testing.T) {
	for entrada, quer := range map[string]string{"=1+1": "'=1+1", "+x": "'+x", "-5": "'-5", "@a": "'@a", "livros": "livros", "": ""} {
		if obtido := seguro(entrada); obtido != quer {
			t.Errorf("seguro(%q) = %q, quer %q", entrada, obtido, quer)
		}
	}
}
