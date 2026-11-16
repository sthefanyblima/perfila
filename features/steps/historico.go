package steps

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cucumber/godog"

	"perfila/internal/application/usecase"
	"perfila/internal/domain/entity"
)

// Passos do histórico do consentimento. Spec: RN-10, Q-34.
func registrarPassosHistorico(sc *godog.ScenarioContext, m *Mundo) {
	var historico []entity.MovimentoConsentimento
	var ultimo entity.MovimentoConsentimento

	sc.Step(`^que o consentimento do cliente "([^"]*)" teve os movimentos:$`, func(nome string, t *godog.Table) error {
		return m.tevesMovimentos(nome, t)
	})
	sc.Step(`^o histórico do consentimento do cliente "([^"]*)" é consultado$`, func(nome string) error {
		historico, m.err = m.uc().consentimento.Historico(m.ctx, id(nome))
		return nil
	})
	sc.Step(`^o histórico tem os movimentos:$`, func(t *godog.Table) error {
		if err := m.semErro(); err != nil {
			return err
		}
		return confereMovimentos(historico, t)
	})
	sc.Step(`^o cliente "([^"]*)" concede consentimento para a finalidade "([^"]*)" com o termo "([^"]*)" pela origem "([^"]*)" com o operador "([^"]*)"$`,
		func(nome, finalidade, termo, origem, operador string) error {
			ctx := usecase.ComOrigem(m.ctx, entity.OrigemMovimento(origem), operador)
			m.err = m.uc().consentimento.Conceder(ctx, id(nome), entity.Finalidade(finalidade), termo)
			return nil
		})
	sc.Step(`^o último movimento do cliente "([^"]*)" é "([^"]*)" com a origem "([^"]*)" e o operador "([^"]*)"$`,
		func(nome, tipo, origem, operador string) error {
			movs, err := m.uc().consentimento.Historico(m.ctx, id(nome))
			if err != nil {
				return err
			}
			if len(movs) == 0 {
				return fmt.Errorf("o cliente %s não tem movimentos", nome)
			}
			ultimo = movs[len(movs)-1]
			if string(ultimo.Tipo) != tipo || string(ultimo.Origem) != origem || ultimo.Operador != operador {
				return fmt.Errorf("último movimento %+v; esperado %s, origem %q, operador %q", ultimo, tipo, origem, operador)
			}
			return nil
		})
	sc.Step(`^o movimento tem a data "([^"]*)"$`, func(s string) error {
		return m.confereData("movimento", &ultimo.Data, s)
	})
	sc.Step(`^o histórico do consentimento do cliente "([^"]*)" tem (\d+) movimentos?$`, func(nome string, n int) error {
		movs, err := m.uc().consentimento.Historico(m.ctx, id(nome))
		if err != nil {
			return err
		}
		if len(movs) != n {
			return fmt.Errorf("o histórico tem %d movimentos (%+v); esperado %d", len(movs), movs, n)
		}
		return nil
	})
}

// tevesMovimentos aplica UC-08 em cada data da tabela (colunas data, acao, termo e, opcionais,
// origem e operador); o relógio fica na data do último movimento.
func (m *Mundo) tevesMovimentos(nome string, t *godog.Table) error {
	for _, linha := range linhasPorColuna(t) {
		quando, err := instante(linha["data"])
		if err != nil {
			return err
		}
		m.clock.Definir(quando)
		ctx := m.ctx
		if o := linha["origem"]; o != "" {
			ctx = usecase.ComOrigem(ctx, entity.OrigemMovimento(o), linha["operador"])
		}
		switch linha["acao"] {
		case "conceder":
			err = m.uc().consentimento.Conceder(ctx, id(nome), entity.FinalidadeSegmentacao, linha["termo"])
		case "revogar":
			err = m.uc().consentimento.Revogar(ctx, id(nome), entity.FinalidadeSegmentacao)
		default:
			err = fmt.Errorf("ação de consentimento desconhecida: %q", linha["acao"])
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func confereMovimentos(movs []entity.MovimentoConsentimento, t *godog.Table) error {
	linhas := linhasPorColuna(t)
	if len(movs) != len(linhas) {
		return fmt.Errorf("o histórico tem %d movimentos (%+v); esperado %d", len(movs), movs, len(linhas))
	}
	for i, l := range linhas {
		quando, err := instante(l["data"])
		if err != nil {
			return err
		}
		mv := movs[i]
		if !mv.Data.Equal(quando) || string(mv.Tipo) != l["tipo"] || mv.Finalidade != l["finalidade"] ||
			mv.VersaoTermo != l["termo"] || string(mv.Origem) != l["origem"] {
			return fmt.Errorf("movimento %d: %+v; esperado %v", i+1, mv, l)
		}
	}
	return nil
}

// linhasPorColuna lê uma tabela com cabeçalho como mapas coluna → valor.
func linhasPorColuna(t *godog.Table) []map[string]string {
	var out []map[string]string
	for _, linha := range t.Rows[1:] {
		l := map[string]string{}
		for i, c := range linha.Cells {
			l[t.Rows[0].Cells[i].Value] = c.Value
		}
		out = append(out, l)
	}
	return out
}

var reMarco = regexp.MustCompile(`(?s)data-teste="marco">(.*?)</li>`)

// Passos @web do histórico: a linha do tempo da aba Consentimento. Spec: 3.7.1, Q-34.
func registrarPassosHistoricoWeb(sc *godog.ScenarioContext, m *Mundo) {
	sc.Step(`^a linha do tempo do consentimento mostra, do mais recente para o mais antigo:$`, func(t *godog.Table) error {
		marcos := reMarco.FindAllStringSubmatch(m.web.pagina, -1)
		linhas := linhasPorColuna(t)
		if len(marcos) != len(linhas) {
			return fmt.Errorf("a linha do tempo tem %d marcos; esperado %d", len(marcos), len(linhas))
		}
		for i, l := range linhas {
			if !strings.Contains(marcos[i][1], "<strong>"+l["titulo"]+"</strong>") || !strings.Contains(marcos[i][1], l["detalhe"]) {
				return fmt.Errorf("marco %d: %s; esperado %q com %q", i+1, marcos[i][1], l["titulo"], l["detalhe"])
			}
		}
		return nil
	})
	sc.Step(`^a página não mostra o aviso "([^"]*)"$`, func(a string) error {
		return m.paginaTemMarcador(`data-teste="aviso-`+a+`"`, false)
	})
}
