package entity

import "testing"

// Spec: RN-15, Q-46 — grupos de 1 a 4 clientes nunca aparecem como número.
func TestProtegerContagem(t *testing.T) {
	casos := []struct {
		valor, clientes int
		quer            Contagem
	}{
		{0, 0, Contagem{}},
		{1, 1, Contagem{Oculta: true}},
		{12, 4, Contagem{Oculta: true}},
		{5, 5, Contagem{Valor: 5}},
		{30, 6, Contagem{Valor: 30}},
	}
	for _, c := range casos {
		if obtido := ProtegerContagem(c.valor, c.clientes); obtido != c.quer {
			t.Errorf("ProtegerContagem(%d, %d) = %+v, quer %+v", c.valor, c.clientes, obtido, c.quer)
		}
	}
}

// Spec: RN-15 — média de grupo pequeno oculta; grupo vazio sem média.
func TestProtegerMedia(t *testing.T) {
	casos := []struct {
		soma     float64
		clientes int
		quer     Media
	}{
		{0, 0, Media{Vazia: true}},
		{400, 4, Media{Oculta: true}},
		{500, 5, Media{Valor: 100}},
		{1.5, 6, Media{Valor: 0.25}},
	}
	for _, c := range casos {
		if obtido := ProtegerMedia(c.soma, c.clientes); obtido != c.quer {
			t.Errorf("ProtegerMedia(%v, %d) = %+v, quer %+v", c.soma, c.clientes, obtido, c.quer)
		}
	}
}

// Spec: RN-15 — a taxa some se a parte ou o total estiver oculto, ou se o total for zero.
func TestProtegerPercentual(t *testing.T) {
	oculta := Contagem{Oculta: true}
	casos := []struct {
		parte, total Contagem
		quer         Percentual
	}{
		{Contagem{Valor: 5}, Contagem{Valor: 10}, Percentual{Valor: 50}},
		{Contagem{Valor: 0}, Contagem{Valor: 10}, Percentual{Valor: 0}},
		{oculta, Contagem{Valor: 9}, Percentual{Oculta: true}},
		{Contagem{Valor: 0}, oculta, Percentual{Oculta: true}},
		{Contagem{}, Contagem{}, Percentual{Oculta: true}},
	}
	for _, c := range casos {
		if obtido := ProtegerPercentual(c.parte, c.total); obtido != c.quer {
			t.Errorf("ProtegerPercentual(%+v, %+v) = %+v, quer %+v", c.parte, c.total, obtido, c.quer)
		}
	}
}
