package web

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"perfila/internal/application/usecase"
	"perfila/internal/domain/entity"
	"perfila/internal/shared/errs"
)

// Tela Relatórios e CSV agregado (UC-05 e UC-07 parciais). Só números protegidos pelo caso de
// uso; nenhum nome, e-mail ou ID. Spec: 3.7.1, RN-15, Q-46.

const textoPeriodoInvalido = "Escolha o período de 6 ou 12 meses."

func (p *painel) rotasRelatorios(mux *http.ServeMux, sessao func(http.HandlerFunc) http.Handler) {
	mux.Handle("GET /painel/relatorios", sessao(p.telaRelatorios))
	mux.Handle("GET /painel/relatorios.csv", sessao(p.exportarRelatorio))
}

// gerar lê o período da query (vazio = padrão) e chama o caso de uso; período inválido vira 422.
func (p *painel) gerar(w http.ResponseWriter, r *http.Request, acao string) (usecase.Relatorio, bool) {
	periodo := usecase.PeriodoPadraoRelatorio
	if v := r.URL.Query().Get("periodo"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			n = -1
		}
		periodo = n
	}
	rel, err := p.d.GerarRelatorio.Executar(r.Context(), periodo)
	if errors.Is(err, errs.ErrDadosInvalidos) {
		p.renderErro(w, r, http.StatusUnprocessableEntity, "Período inválido", textoPeriodoInvalido)
		return rel, false
	}
	if err != nil {
		p.falha(w, r, err, "")
		return rel, false
	}
	// C-09: só operador, ação e período; nenhum número do relatório.
	p.d.Log.LogAttrs(r.Context(), slog.LevelInfo, "painel_acao", slog.String("operador", operadorDe(r)),
		slog.String("acao", acao), slog.Int("periodo", periodo))
	return rel, true
}

// GET /painel/relatorios — indicadores, gráficos SVG e tabelas.
func (p *painel) telaRelatorios(w http.ResponseWriter, r *http.Request) {
	rel, ok := p.gerar(w, r, "ver_relatorio")
	if !ok {
		return
	}
	p.render(w, r, http.StatusOK, "relatorios", pagina{Titulo: "Relatórios", Nav: "relatorios", Dados: montarTelaRelatorio(rel)})
}

// ---------- modelo da tela ----------

type celula struct {
	Texto  string
	Oculta bool
}

type linhaTabela struct {
	Rotulo  string
	Valores []celula
}

type tabela struct {
	Titulo  string
	Colunas []string
	Linhas  []linhaTabela
}

type indicador struct {
	Rotulo, Nota string
	Valor        celula
	Teste        string
}

type telaRelatorio struct {
	Periodo     int
	Periodos    []opcaoPeriodo
	Intervalo   string
	GeradoEm    string
	Indicadores []indicador
	Graficos    []graficoComTabela
}

type opcaoPeriodo struct {
	Valor  int
	Rotulo string
	Atual  bool
}

type graficoComTabela struct {
	ID, Titulo, Nota string
	Grafico          grafico
	Tabela           tabela
	Vazio            string // texto quando não há dados
}

func exibirContagem(c entity.Contagem) celula {
	if c.Oculta {
		return celula{Texto: "menos de 5", Oculta: true}
	}
	return celula{Texto: decimal(float64(c.Valor), 0)}
}

func exibirMedia(m entity.Media, formato func(float64) string) celula {
	switch {
	case m.Vazia:
		return celula{Texto: "—"}
	case m.Oculta:
		return celula{Texto: "oculta", Oculta: true}
	}
	return celula{Texto: formato(m.Valor)}
}

func exibirPercentual(v entity.Percentual) celula {
	if v.Oculta {
		return celula{Texto: "oculta", Oculta: true}
	}
	return celula{Texto: decimal(v.Valor, 1) + "%"}
}

func score(v float64) string { return decimal(v, 2) }

func nomeSegmento(s usecase.LinhaSegmento) string {
	if s.Nome == "" {
		return "Ainda não segmentados"
	}
	return s.Nome
}

func nomeCategoria(c usecase.LinhaCategoria) string {
	if c.Outras {
		return fmt.Sprintf("Outras categorias (%d)", c.Categorias)
	}
	return c.Categoria
}

func rotuloMes(t time.Time) string {
	t = t.In(fusoSaoPaulo())
	meses := strings.Fields("jan fev mar abr mai jun jul ago set out nov dez")
	return fmt.Sprintf("%s/%02d", meses[t.Month()-1], t.Year()%100)
}

func montarTelaRelatorio(rel usecase.Relatorio) telaRelatorio {
	t := telaRelatorio{Periodo: rel.PeriodoMeses, GeradoEm: dataHora(rel.GeradoEm),
		Intervalo: soData(rel.Desde) + " a " + soData(rel.Ate)}
	for _, n := range []int{6, 12} {
		t.Periodos = append(t.Periodos, opcaoPeriodo{Valor: n, Rotulo: fmt.Sprintf("Últimos %d meses", n), Atual: n == rel.PeriodoMeses})
	}
	t.Indicadores = []indicador{
		{Rotulo: "Clientes", Valor: exibirContagem(rel.Clientes), Nota: "todos os cadastros", Teste: "clientes"},
		{Rotulo: "Ativos", Valor: exibirContagem(rel.Ativos), Nota: "cadastros ativos", Teste: "ativos"},
		{Rotulo: "Taxa de concessão", Valor: exibirPercentual(rel.TaxaConcessao), Nota: "consentimento concedido ÷ clientes", Teste: "taxa"},
		{Rotulo: "Recomendações ativas", Valor: exibirContagem(rel.RecomendacoesAtivas),
			Nota: "expiradas: " + exibirContagem(rel.RecomendacoesExpiradas).Texto, Teste: "recomendacoes"},
	}

	// Consentimento por estado: cores de estado do design system.
	cons := graficoComTabela{ID: "consentimento", Titulo: "Consentimento por estado",
		Nota: "Conta todos os clientes; quem nunca concedeu aparece como Não concedido."}
	cons.Tabela = tabela{Titulo: "Clientes por estado do consentimento", Colunas: []string{"Estado", "Clientes"}}
	var barrasCons []dadoBarra
	for _, l := range rel.Consentimento {
		cons.Tabela.Linhas = append(cons.Tabela.Linhas, linhaTabela{Rotulo: rotulo(string(l.Status)), Valores: []celula{exibirContagem(l.Clientes)}})
		barrasCons = append(barrasCons, dadoBarra{Rotulo: rotulo(string(l.Status)), Valor: l.Clientes, Tom: tomDoEstado(string(l.Status))})
	}
	cons.Grafico = barrasHorizontais(barrasCons)

	// Segmentos: dado derivado, acento ameixa.
	segs := graficoComTabela{ID: "segmentos", Titulo: "Clientes por segmento",
		Nota: "Só clientes com consentimento válido. Médias de grupos com menos de 5 clientes ficam ocultas."}
	segs.Tabela = tabela{Titulo: "Distribuição por segmento", Colunas: []string{"Segmento", "Clientes", "Ticket médio", "Score médio"}}
	var barrasSeg []dadoBarra
	for _, s := range rel.Segmentos {
		segs.Tabela.Linhas = append(segs.Tabela.Linhas, linhaTabela{Rotulo: nomeSegmento(s), Valores: []celula{
			exibirContagem(s.Clientes), exibirMedia(s.TicketMedio, dinheiro), exibirMedia(s.ScoreMedio, score)}})
		barrasSeg = append(barrasSeg, dadoBarra{Rotulo: nomeSegmento(s), Valor: s.Clientes, Tom: "accent"})
	}
	segs.Grafico = barrasHorizontais(barrasSeg)
	if len(barrasSeg) == 0 {
		segs.Vazio = "Nenhum segmento ativo."
	}

	cats := graficoComTabela{ID: "categorias", Titulo: "Categorias mais compradas",
		Nota: "Compras no período, só de clientes com consentimento válido. Categorias com menos de 5 compradores entram em Outras, sem nome."}
	cats.Tabela = tabela{Titulo: "Compras por categoria", Colunas: []string{"Categoria", "Compras", "Clientes"}}
	var barrasCat []dadoBarra
	for _, c := range rel.Categorias {
		cats.Tabela.Linhas = append(cats.Tabela.Linhas, linhaTabela{Rotulo: nomeCategoria(c), Valores: []celula{
			exibirContagem(c.Compras), exibirContagem(c.Clientes)}})
		barrasCat = append(barrasCat, dadoBarra{Rotulo: nomeCategoria(c), Valor: c.Compras, Tom: "accent"})
	}
	cats.Grafico = barrasHorizontais(barrasCat)
	if len(barrasCat) == 0 {
		cats.Vazio = "Nenhuma compra com categoria no período."
	}

	meses := graficoComTabela{ID: "meses", Titulo: "Cadastros e revogações por mês",
		Nota: "Contam todos os clientes. Revogações vêm do histórico do consentimento."}
	meses.Tabela = tabela{Titulo: "Evolução mensal", Colunas: []string{"Mês", "Cadastros", "Revogações", "Eventos"}}
	eventos := graficoComTabela{ID: "eventos", Titulo: "Eventos por mês",
		Nota: "Navegação, compra e interação de clientes com consentimento válido. A tabela está em Evolução mensal."}
	var grupos []grupoColunas
	var barrasEv []dadoBarra
	for _, l := range rel.Meses {
		meses.Tabela.Linhas = append(meses.Tabela.Linhas, linhaTabela{Rotulo: rotuloMes(l.Inicio), Valores: []celula{
			exibirContagem(l.Cadastros), exibirContagem(l.Revogacoes), exibirContagem(l.Eventos)}})
		grupos = append(grupos, grupoColunas{Rotulo: rotuloMes(l.Inicio), Series: []dadoBarra{
			{Rotulo: "Cadastros", Valor: l.Cadastros, Tom: "neutral"}, {Rotulo: "Revogações", Valor: l.Revogacoes, Tom: "critical"}}})
		barrasEv = append(barrasEv, dadoBarra{Rotulo: rotuloMes(l.Inicio), Valor: l.Eventos, Tom: "accent"})
	}
	meses.Grafico = colunas(grupos, []legenda{{"Cadastros", "neutral"}, {"Revogações", "critical"}})
	eventos.Grafico = colunas(umaSerie(barrasEv), nil)

	t.Graficos = []graficoComTabela{cons, segs, cats, meses, eventos}
	return t
}

// ---------- CSV ----------

// GET /painel/relatorios.csv — o mesmo relatório, protegido, em CSV para planilha (`;`, vírgula
// decimal, UTF-8 com BOM).
func (p *painel) exportarRelatorio(w http.ResponseWriter, r *http.Request) {
	rel, ok := p.gerar(w, r, "exportar_relatorio")
	if !ok {
		return
	}
	nome := fmt.Sprintf("perfila-relatorio-%d-meses-%s.csv", rel.PeriodoMeses, rel.GeradoEm.In(fusoSaoPaulo()).Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+nome+`"`)
	_, _ = w.Write([]byte("\ufeff"))
	cw := csv.NewWriter(w)
	cw.Comma, cw.UseCRLF = ';', true
	for _, l := range linhasCSV(rel) {
		_ = cw.Write([]string{seguro(l[0]), seguro(l[1]), seguro(l[2]), seguro(l[3])})
	}
	cw.Flush()
}

// seguro impede que a planilha interprete uma célula como fórmula (categoria é texto livre).
func seguro(s string) string {
	if s != "" && strings.ContainsRune("=+-@\t\r", rune(s[0])) {
		return "'" + s
	}
	return s
}

func linhasCSV(rel usecase.Relatorio) [][4]string {
	out := [][4]string{{"secao", "item", "indicador", "valor"},
		{"relatorio", "período", "meses", strconv.Itoa(rel.PeriodoMeses)},
		{"relatorio", "intervalo", "datas", soData(rel.Desde) + " a " + soData(rel.Ate)},
		{"relatorio", "gerado em", "data e hora", dataHora(rel.GeradoEm)},
		{"clientes", "todos", "total", exibirContagem(rel.Clientes).Texto},
		{"clientes", "todos", "ativos", exibirContagem(rel.Ativos).Texto},
	}
	for _, l := range rel.Consentimento {
		out = append(out, [4]string{"consentimento", rotulo(string(l.Status)), "clientes", exibirContagem(l.Clientes).Texto})
	}
	out = append(out, [4]string{"consentimento", "todos", "taxa de concessão (%)", strings.TrimSuffix(exibirPercentual(rel.TaxaConcessao).Texto, "%")})
	for _, s := range rel.Segmentos {
		n := nomeSegmento(s)
		out = append(out, [4]string{"segmentos", n, "clientes", exibirContagem(s.Clientes).Texto},
			[4]string{"segmentos", n, "ticket médio (R$)", exibirMedia(s.TicketMedio, func(v float64) string { return decimal(v, 2) }).Texto},
			[4]string{"segmentos", n, "score médio", exibirMedia(s.ScoreMedio, score).Texto})
	}
	out = append(out, [4]string{"recomendacoes", "Ativa", "recomendações", exibirContagem(rel.RecomendacoesAtivas).Texto},
		[4]string{"recomendacoes", "Expirada", "recomendações", exibirContagem(rel.RecomendacoesExpiradas).Texto})
	for _, c := range rel.Categorias {
		n := nomeCategoria(c)
		out = append(out, [4]string{"categorias", n, "compras", exibirContagem(c.Compras).Texto},
			[4]string{"categorias", n, "clientes", exibirContagem(c.Clientes).Texto})
	}
	for _, l := range rel.Meses {
		m := l.Inicio.In(fusoSaoPaulo()).Format("2006-01")
		out = append(out, [4]string{"meses", m, "cadastros", exibirContagem(l.Cadastros).Texto},
			[4]string{"meses", m, "eventos", exibirContagem(l.Eventos).Texto},
			[4]string{"meses", m, "revogações", exibirContagem(l.Revogacoes).Texto})
	}
	return out
}
