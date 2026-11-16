package web

import (
	"strings"
	"unicode/utf8"

	"perfila/internal/domain/entity"
)

// Gráficos de barras desenhados no servidor em SVG: só geometria e classes CSS (sem atributo
// style, CSP 'self'). Um grupo oculto ("menos de 5") vira uma barra tracejada de tamanho fixo,
// que não revela o número. A tabela ao lado é a fonte para leitores de tela. Spec: 3.7.1, RN-15.

type dadoBarra struct {
	Rotulo string
	Valor  entity.Contagem
	Tom    string // accent, ok, neutral, critical
}

type grupoColunas struct {
	Rotulo string
	Series []dadoBarra
}

type legenda struct{ Rotulo, Tom string }

type barraSVG struct {
	X, Y, L, A float64 // retângulo
	Tom        string
	Oculta     bool
	Valor      string
	TX, TY     float64 // texto do valor
	Ancora     string  // text-anchor do valor
}

type textoSVG struct {
	X, Y   float64
	Texto  string
	Ancora string
}

type grafico struct {
	Largura, Altura int
	Barras          []barraSVG
	Rotulos         []textoSVG
	Base            float64 // linha de base das colunas (0 = sem linha)
	Legenda         []legenda
}

// escala devolve o maior valor visível, com piso no tamanho mínimo de grupo, para a barra
// tracejada ocupar sempre o mesmo espaço relativo.
func escala(dados []dadoBarra) float64 {
	maior := float64(entity.MinimoGrupo)
	for _, d := range dados {
		if !d.Valor.Oculta && float64(d.Valor.Valor) > maior {
			maior = float64(d.Valor.Valor)
		}
	}
	return maior
}

func valorTexto(c entity.Contagem) string {
	if c.Oculta {
		return "menos de 5"
	}
	return decimal(float64(c.Valor), 0)
}

// comprimento: oculta = tamanho de MinimoGrupo na escala (mínimo de 8 px), sem revelar o número.
func comprimento(c entity.Contagem, maior, maximo float64) float64 {
	v := float64(c.Valor)
	if c.Oculta {
		v = float64(entity.MinimoGrupo)
	}
	l := maximo * v / maior
	if (c.Oculta || c.Valor > 0) && l < 8 {
		l = 8
	}
	return l
}

func encurtar(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n-1]) + "…"
}

const (
	larguraGrafico = 440
	inicioBarra    = 128.0
	maximoBarra    = 232.0
	alturaLinha    = 30.0
)

// barrasHorizontais: uma linha por item, rótulo à esquerda e valor depois da barra.
func barrasHorizontais(dados []dadoBarra) grafico {
	g := grafico{Largura: larguraGrafico, Altura: int(alturaLinha*float64(len(dados))) + 8}
	maior := escala(dados)
	for i, d := range dados {
		y := 4 + alturaLinha*float64(i)
		l := comprimento(d.Valor, maior, maximoBarra)
		g.Rotulos = append(g.Rotulos, textoSVG{X: 0, Y: y + 20, Texto: encurtar(d.Rotulo, 16), Ancora: "start"})
		g.Barras = append(g.Barras, barraSVG{X: inicioBarra, Y: y + 6, L: l, A: 18, Tom: d.Tom, Oculta: d.Valor.Oculta,
			Valor: valorTexto(d.Valor), TX: inicioBarra + l + 8, TY: y + 20, Ancora: "start"})
	}
	return g
}

func umaSerie(dados []dadoBarra) []grupoColunas {
	out := make([]grupoColunas, len(dados))
	for i, d := range dados {
		out[i] = grupoColunas{Rotulo: d.Rotulo, Series: []dadoBarra{d}}
	}
	return out
}

// colunas: um grupo por mês, uma coluna por série, rótulo do mês embaixo.
func colunas(grupos []grupoColunas, leg []legenda) grafico {
	const topo, base, margem = 22.0, 168.0, 4.0
	g := grafico{Largura: larguraGrafico, Altura: 196, Base: base, Legenda: leg}
	if len(grupos) == 0 {
		return g
	}
	var todos []dadoBarra
	for _, gr := range grupos {
		todos = append(todos, gr.Series...)
	}
	maior := escala(todos)
	largGrupo := (larguraGrafico - 2*margem) / float64(len(grupos))
	for i, gr := range grupos {
		k := float64(len(gr.Series))
		larg := largGrupo * 0.7 / k
		if larg > 28 {
			larg = 28
		}
		centro := margem + largGrupo*(float64(i)+0.5)
		x := centro - larg*k/2
		for _, s := range gr.Series {
			a := comprimento(s.Valor, maior, base-topo)
			g.Barras = append(g.Barras, barraSVG{X: x, Y: base - a, L: larg, A: a, Tom: s.Tom, Oculta: s.Valor.Oculta,
				Valor: strings.Replace(valorTexto(s.Valor), "menos de 5", "<5", 1), TX: x + larg/2, TY: base - a - 6, Ancora: "middle"})
			x += larg
		}
		g.Rotulos = append(g.Rotulos, textoSVG{X: centro, Y: base + 18, Texto: gr.Rotulo, Ancora: "middle"})
	}
	return g
}
