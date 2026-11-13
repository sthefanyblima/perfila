package entity

// Valores protegidos dos relatórios agregados [ADIÇÃO, v1.7]. Não são persistidos: o caso de uso
// GerarRelatorio os monta a partir das contagens cruas do RelatorioRepository, e nenhum número
// cru sai dele. Spec: 3.2, RN-15, Q-46.

// MinimoGrupo é o menor número de clientes distintos por trás de um valor exibido. Spec: RN-15.
const MinimoGrupo = 5

// Contagem é um número de relatório; Oculta = 1 a 4 clientes por trás ("menos de 5").
type Contagem struct {
	Valor  int
	Oculta bool
}

// Media é uma média de grupo; Oculta = grupo de 1 a 4 clientes; Vazia = grupo sem clientes.
type Media struct {
	Valor  float64
	Oculta bool
	Vazia  bool
}

// Percentual vai de 0 a 100; Oculta = parte ou total oculto, ou total zero.
type Percentual struct {
	Valor  float64
	Oculta bool
}

// grupoPequeno diz se o número de clientes distintos revela pessoas: de 1 a MinimoGrupo-1.
func grupoPequeno(clientes int) bool { return clientes > 0 && clientes < MinimoGrupo }

// ProtegerContagem esconde o valor quando há de 1 a 4 clientes distintos por trás dele; zero
// aparece. Para contagem de clientes, valor e clientes são o mesmo número. Spec: RN-15.
func ProtegerContagem(valor, clientes int) Contagem {
	if grupoPequeno(clientes) {
		return Contagem{Oculta: true}
	}
	return Contagem{Valor: valor}
}

// ProtegerMedia calcula soma ÷ clientes; esconde a média de grupo pequeno e marca o grupo vazio.
// Spec: RN-15.
func ProtegerMedia(soma float64, clientes int) Media {
	switch {
	case clientes <= 0:
		return Media{Vazia: true}
	case grupoPequeno(clientes):
		return Media{Oculta: true}
	}
	return Media{Valor: soma / float64(clientes)}
}

// ProtegerPercentual calcula parte ÷ total em %; fica oculto se uma das contagens estiver oculta
// ou se o total for zero, para não reconstruir um grupo pequeno. Spec: RN-15.
func ProtegerPercentual(parte, total Contagem) Percentual {
	if parte.Oculta || total.Oculta || total.Valor <= 0 {
		return Percentual{Oculta: true}
	}
	return Percentual{Valor: 100 * float64(parte.Valor) / float64(total.Valor)}
}
