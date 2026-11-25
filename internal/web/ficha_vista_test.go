package web

import (
	"strings"
	"testing"
)

// Spec: UC-04 RN-09, UML4. A régua do processamento acende um quarto por etapa concluída (ou
// concluída com atenção) e para a agulha no meio da etapa que falhou.
func TestReguaDoProcessamento(t *testing.T) {
	e := func(estados ...string) []etapaVista {
		var l []etapaVista
		for i, s := range estados {
			l = append(l, etapaVista{N: i + 1, Estado: s})
		}
		return l
	}
	casos := []struct {
		nome    string
		etapas  []etapaVista
		mostrar bool
		agulha  string
		acesas  int
	}{
		{"sem processamento", e("pendente", "pendente", "pendente", "pendente"), false, "", 0},
		{"tudo concluído", e("concluida", "concluida", "concluida", "concluida"), true, "4", 101},
		{"última com atenção conta como feita", e("concluida", "concluida", "concluida", "atencao"), true, "4", 101},
		{"duas etapas", e("concluida", "concluida", "pendente", "pendente"), true, "2", 51},
		{"falha na terceira", e("concluida", "concluida", "falhou", "pendente"), true, "falha", 63},
		{"falha na primeira", e("falhou", "pendente", "pendente", "pendente"), true, "falha", 13},
		{"sem etapas", nil, false, "", 0},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			r := regua(c.etapas)
			if len(r.Marcas) != 101 {
				t.Fatalf("marcações = %d, quer 101", len(r.Marcas))
			}
			if r.Mostrar != c.mostrar || r.Agulha != c.agulha {
				t.Errorf("mostrar=%v agulha=%q, quer %v e %q", r.Mostrar, r.Agulha, c.mostrar, c.agulha)
			}
			acesas := 0
			for _, m := range r.Marcas {
				if !m.Apagada {
					acesas++
				}
			}
			if acesas != c.acesas {
				t.Errorf("acesas = %d, quer %d", acesas, c.acesas)
			}
			if r.Marcas[0].Classe != "is-major" || r.Marcas[25].Classe != "is-major" || r.Marcas[5].Classe != "is-mid" || r.Marcas[1].Classe != "is-minor" {
				t.Errorf("classes de altura erradas: %q %q %q %q", r.Marcas[0].Classe, r.Marcas[25].Classe, r.Marcas[5].Classe, r.Marcas[1].Classe)
			}
		})
	}
}

// Spec: README (Nomes longos). O nome do cliente se divide em primeiro nome e resto, e a fonte
// reduz acima de 20 e de 40 caracteres.
func TestNomeFicha(t *testing.T) {
	casos := []struct {
		nome, primeiro, resto, classe string
	}{
		{"Marina Costa", "Marina", "Costa", ""},
		{"Cliente Volume 125", "Cliente", "Volume 125", ""},
		{"Ana", "Ana", "", ""},
		{"", "", "", ""},
		{"  Ana   Maria  ", "Ana", "Maria", ""},
		{strings.Repeat("a", 20), strings.Repeat("a", 20), "", ""},
		{strings.Repeat("a", 21), strings.Repeat("a", 21), "", " pf-ficha__nome--lg"},
		{strings.Repeat("ã", 40), strings.Repeat("ã", 40), "", " pf-ficha__nome--lg"},
		{strings.Repeat("ã", 41), strings.Repeat("ã", 41), "", " pf-ficha__nome--xl"},
	}
	for _, c := range casos {
		got := nomeFicha(c.nome)
		if got.Primeiro != c.primeiro || got.Resto != c.resto || got.Classe != c.classe {
			t.Errorf("nomeFicha(%q) = %+v, quer {%q %q %q}", c.nome, got, c.primeiro, c.resto, c.classe)
		}
	}
}

func TestPrimeiroNome(t *testing.T) {
	for nome, quer := range map[string]string{"Marina Costa": "Marina", "Ana": "Ana", "": "", "  Davi  Silva": "Davi"} {
		if got := primeiroNome(nome); got != quer {
			t.Errorf("primeiroNome(%q) = %q, quer %q", nome, got, quer)
		}
	}
}

// Spec: RN-05. O engajamento aparece em 10 marcas: score arredondado, entre 0 e 10.
func TestMarcasScore(t *testing.T) {
	casos := []struct {
		score  float64
		acesas int
	}{{0, 0}, {0.04, 0}, {0.05, 1}, {0.4, 4}, {0.75, 8}, {1, 10}, {1.4, 10}, {-0.3, 0}}
	for _, c := range casos {
		m := marcasScore(c.score)
		if len(m) != 10 {
			t.Fatalf("marcasScore(%v) tem %d marcas, quer 10", c.score, len(m))
		}
		n := 0
		for i, on := range m {
			if on {
				n++
				if i >= c.acesas {
					t.Errorf("marcasScore(%v): marca %d acesa fora da frente", c.score, i)
				}
			}
		}
		if n != c.acesas {
			t.Errorf("marcasScore(%v) acende %d, quer %d", c.score, n, c.acesas)
		}
	}
}
