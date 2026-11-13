package steps

import (
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/google/uuid"

	"perfila/internal/application/usecase"
	"perfila/internal/domain/entity"
)

// Estado dos cenários de relatórios (RN-15, Q-46).
type estadoRelatorio struct {
	relatorio usecase.Relatorio
	exibido   string
	criados   int // contador dos clientes sintéticos em grupo (r-001, r-002, ...)
}

// fusoRelatorio é o fuso de São Paulo sem horário de verão, fixo para os cenários (B-07).
var fusoRelatorio = time.FixedZone("BRT", -3*60*60)

func registrarPassosRelatorio(sc *godog.ScenarioContext, m *Mundo) {
	sc.Step(`^uma contagem de (\d+) com (\d+) clientes distintos é protegida$`, func(v, c int) error {
		m.rel.exibido = exibirContagem(entity.ProtegerContagem(v, c))
		return nil
	})
	sc.Step(`^a média de uma soma de `+num+` com (\d+) clientes é protegida$`, func(s float64, c int) error {
		m.rel.exibido = exibirMedia(entity.ProtegerMedia(s, c))
		return nil
	})
	sc.Step(`^o valor exibido é "([^"]*)"$`, func(quer string) error { return iguais("valor exibido", m.rel.exibido, quer) })

	sc.Step(`^que existem (\d+) clientes com consentimento concedido e compras na categoria "([^"]*)"$`,
		func(n int, cat string) error { return m.grupoDeClientes(n, cat, true, false, false) })
	sc.Step(`^que existem (\d+) clientes com consentimento concedido e compras na categoria "([^"]*)" há (\d+) meses$`,
		func(n int, cat string, meses int) error {
			agora := m.clock.Agora()
			m.clock.Definir(agora.AddDate(0, -meses, 0))
			defer m.clock.Definir(agora)
			return m.grupoDeClientes(n, cat, true, false, false)
		})
	sc.Step(`^que existem (\d+) clientes inativos com consentimento concedido e compras na categoria "([^"]*)"$`,
		func(n int, cat string) error { return m.grupoDeClientes(n, cat, true, false, true) })
	sc.Step(`^que existem (\d+) clientes que compraram na categoria "([^"]*)" e revogaram o consentimento$`,
		func(n int, cat string) error { return m.grupoDeClientes(n, cat, true, true, false) })
	sc.Step(`^que existem (\d+) clientes sem consentimento$`, func(n int) error { return m.grupoDeClientes(n, "", false, false, false) })

	sc.Step(`^o relatório dos últimos (\d+) meses é gerado$`, func(p int) error {
		m.rel.relatorio, m.err = m.uc().gerarRelatorio.Executar(m.ctx, p)
		return nil
	})
	sc.Step(`^o relatório mostra:$`, m.relatorioMostra)
	sc.Step(`^o relatório não mostra a categoria "([^"]*)"$`, func(cat string) error {
		if _, ok := m.indicadores()["categoria "+cat+": compras"]; ok {
			return fmt.Errorf("o relatório mostra a categoria %q", cat)
		}
		return nil
	})
	sc.Step(`^as compras da categoria "([^"]*)" no relatório são "([^"]*)"$`, func(cat, quer string) error {
		obtido, ok := m.indicadores()["categoria "+cat+": compras"]
		if !ok {
			obtido = "nenhuma"
		}
		return iguais("compras de "+cat, obtido, quer)
	})
	sc.Step(`^o relatório tem (\d+) meses, de "([^"]*)" a "([^"]*)"$`, func(n int, primeiro, ultimo string) error {
		if err := m.semErro(); err != nil {
			return err
		}
		meses := m.rel.relatorio.Meses
		if len(meses) != n {
			return fmt.Errorf("o relatório tem %d meses, esperado %d", len(meses), n)
		}
		return iguais("meses", mes(meses[0].Inicio)+" a "+mes(meses[n-1].Inicio), primeiro+" a "+ultimo)
	})
}

// grupoDeClientes cria n clientes sintéticos pelos casos de uso: cadastro, consentimento e uma
// compra de R$ 100 (processada), revogação ou inativação, conforme o cenário (B-08).
func (m *Mundo) grupoDeClientes(n int, categoria string, conceder, revogar, inativo bool) error {
	c := m.uc()
	for i := 0; i < n; i++ {
		m.rel.criados++
		nome := fmt.Sprintf("r-%03d", m.rel.criados)
		if err := m.clientes.Salvar(m.ctx, entity.Cliente{ID: id(nome), Nome: "Cliente Relatório " + nome,
			Email: nome + "@exemplo.test", DataCadastro: m.clock.Agora(), Ativo: true}); err != nil {
			return err
		}
		if !conceder {
			continue
		}
		if err := c.consentimento.Conceder(m.ctx, id(nome), "segmentacao", "v1"); err != nil {
			return err
		}
		if _, err := c.coletar.Registrar(m.ctx, entity.EventoComportamental{ID: uuid.New(), ClienteID: id(nome),
			Tipo: entity.TipoCompra, Origem: "loja on-line", Timestamp: m.clock.Agora().Add(-time.Hour),
			Valor: 100, Categoria: categoria}); err != nil {
			return err
		}
		if _, err := c.processar.Executar(m.ctx, id(nome)); err != nil {
			return err
		}
		if revogar {
			if err := c.consentimento.Revogar(m.ctx, id(nome)); err != nil {
				return err
			}
		}
		if inativo {
			if err := m.clienteInativo(nome); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Mundo) relatorioMostra(t *godog.Table) error {
	if err := m.semErro(); err != nil {
		return err
	}
	obtidos := m.indicadores()
	var erros []string
	for _, linha := range t.Rows[1:] {
		chave, quer := linha.Cells[0].Value, linha.Cells[1].Value
		obtido, ok := obtidos[chave]
		if !ok {
			obtido = "(ausente)"
		}
		if obtido != quer {
			erros = append(erros, fmt.Sprintf("%s: obtido %q, esperado %q", chave, obtido, quer))
		}
	}
	if len(erros) > 0 {
		return fmt.Errorf("indicadores diferentes:\n  %s", strings.Join(erros, "\n  "))
	}
	return nil
}

// indicadores achata o relatório em "nome do indicador" → valor exibido, a forma dos cenários.
func (m *Mundo) indicadores() map[string]string {
	r := m.rel.relatorio
	out := map[string]string{"clientes": exibirContagem(r.Clientes), "ativos": exibirContagem(r.Ativos),
		"recomendações ativas": exibirContagem(r.RecomendacoesAtivas), "recomendações expiradas": exibirContagem(r.RecomendacoesExpiradas)}
	out["taxa de concessão"] = "oculta"
	if !r.TaxaConcessao.Oculta {
		out["taxa de concessão"] = fmt.Sprintf("%.1f", r.TaxaConcessao.Valor)
	}
	for _, l := range r.Consentimento {
		out["consentimento "+string(l.Status)] = exibirContagem(l.Clientes)
	}
	for _, s := range r.Segmentos {
		nome := "segmento " + s.Nome
		if s.Nome == "" {
			nome = "ainda não segmentados"
		}
		out[nome+": clientes"] = exibirContagem(s.Clientes)
		out[nome+": ticket médio"] = exibirMedia(s.TicketMedio)
		out[nome+": score médio"] = exibirMedia(s.ScoreMedio)
	}
	for _, c := range r.Categorias {
		nome := "categoria " + c.Categoria
		if c.Outras {
			nome = "outras categorias"
			out[nome+": categorias"] = fmt.Sprint(c.Categorias)
		}
		out[nome+": compras"] = exibirContagem(c.Compras)
		out[nome+": clientes"] = exibirContagem(c.Clientes)
	}
	for _, l := range r.Meses {
		out["cadastros em "+mes(l.Inicio)] = exibirContagem(l.Cadastros)
		out["eventos em "+mes(l.Inicio)] = exibirContagem(l.Eventos)
		out["revogações em "+mes(l.Inicio)] = exibirContagem(l.Revogacoes)
	}
	return out
}

func mes(t time.Time) string { return t.In(fusoRelatorio).Format("2006-01") }

func exibirContagem(c entity.Contagem) string {
	if c.Oculta {
		return "menos de 5"
	}
	return fmt.Sprint(c.Valor)
}

func exibirMedia(md entity.Media) string {
	switch {
	case md.Vazia:
		return "sem dados"
	case md.Oculta:
		return "oculta"
	}
	return fmt.Sprintf("%.2f", md.Valor)
}
