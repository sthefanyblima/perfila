package main

import "fmt"

// Clientes de volume (Q-46): 60 clientes sintéticos (c-101 a c-160, @exemplo.test) somados aos 14
// com história, para os relatórios terem grupos acima de 5 (RN-15). A história é determinística,
// derivada só do número do cliente, e passa pelos mesmos casos de uso dos outros exemplos.

const clientesVolume = 60

// categoriasVolume tem as categorias do catálogo de demonstração e uma rara ("papelaria", sem
// produto), que fica com menos de 5 compradores e aparece em "Outras categorias".
func categoriaVolume(i int) string {
	if i%25 == 7 {
		return "papelaria"
	}
	return []string{"livros", "eletronicos", "moda", "games", "beleza", "esportes", "casa"}[(i*3)%7]
}

func (e exemplos) volume() []exemplo {
	out := make([]exemplo, 0, clientesVolume)
	for i := 1; i <= clientesVolume; i++ {
		inicio := float64(8 + (i*41)%340) // cadastro espalhado pelos últimos ~12 meses
		passos := []passo{cad(inicio)}
		if i%10 == 0 { // 6 clientes nunca concedem
			out = append(out, exemplo{fmt.Sprintf("Cliente Volume %03d", 100+i), fmt.Sprintf("c-%03d@exemplo.test", 100+i), passos})
			continue
		}
		passos = append(passos, conc(inicio-1, "v1"))
		cat := categoriaVolume(i)
		compras := 1 + i%4
		intervalo := (inicio - 3) / float64(compras+1)
		for k := 1; k <= compras; k++ {
			passos = append(passos, comp(inicio-1-intervalo*float64(k), cat, float64(60+(i*53+k*17)%900)))
		}
		if n := i % 5; n > 0 {
			passos = append(passos, nav(min(1+float64(i%20), inicio-2), n))
		}
		passos = append(passos, proc(0.5))
		if i%10 == 3 { // 6 clientes revogam depois de processados
			passos = append(passos, rev(0.4))
		}
		out = append(out, exemplo{fmt.Sprintf("Cliente Volume %03d", 100+i), fmt.Sprintf("c-%03d@exemplo.test", 100+i), passos})
	}
	return out
}
