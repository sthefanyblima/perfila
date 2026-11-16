package steps

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
)

// Passos @web da edição de segmentos e catálogo (painel.feature). Spec: 3.7.1 v1.6, RN-14, Q-42 a Q-44.
func registrarPassosSegmentosWeb(sc *godog.ScenarioContext, m *Mundo) {
	sc.Step(`^o formulário do segmento "([^"]*)" é aberto$`, func(nome string) error {
		s, err := m.segmentoPorNome(nome)
		if err != nil {
			return err
		}
		return m.webGET("/painel/segmentos/"+s.ID.String(), true)
	})
	sc.Step(`^o formulário tem (\d+) condições preenchidas$`, func(n int) error {
		if got := strings.Count(m.web.pagina, `data-teste="condicao"`); got != n {
			return fmt.Errorf("o formulário tem %d condições preenchidas", got)
		}
		return nil
	})
	sc.Step(`^o operador envia o formulário do segmento "([^"]*)" com o critério "([^"]*)"$`, func(nome, criterio string) error {
		return m.enviarSegmento("/painel/segmentos/previa", nome, criterio)
	})
	sc.Step(`^o operador confirma o segmento "([^"]*)" com o critério "([^"]*)"$`, func(nome, criterio string) error {
		return m.enviarSegmento("/painel/segmentos", nome, criterio)
	})
	desativar := func(confirmar bool) func(string) error {
		return func(nome string) error {
			s, err := m.segmentoPorNome(nome)
			if err != nil {
				return err
			}
			caminho := "/painel/segmentos/" + s.ID.String() + "/desativar"
			if !confirmar {
				return m.webGET(caminho, true)
			}
			return m.webPOST(caminho, url.Values{}, true)
		}
	}
	sc.Step(`^o operador pede para desativar o segmento "([^"]*)"$`, desativar(false))
	sc.Step(`^o operador confirma a desativação do segmento "([^"]*)"$`, desativar(true))
	sc.Step(`^o operador testa o critério "([^"]*)" com frequência de compras (\d+), ticket médio ([\d.]+) e score ([\d.]+)$`,
		func(criterio, freq, ticket, score string) error {
			form := formCriterio(criterio)
			form.Set("nome", "teste")
			form.Set("prioridade", "1")
			form.Set("teste_frequencia", freq)
			form.Set("teste_ticket", ticket)
			form.Set("teste_score", score)
			return m.webPOST("/painel/segmentos/testar", form, true)
		})
	sc.Step(`^a página não mostra o nome nem o e-mail do cliente "([^"]*)"$`, func(c string) error {
		cli, err := m.clientes.BuscarPorID(m.ctx, id(c))
		if err != nil {
			return err
		}
		if strings.Contains(m.web.pagina, cli.Nome) || strings.Contains(m.web.pagina, cli.Email) {
			return fmt.Errorf("a página mostra dado pessoal do cliente %s", c)
		}
		return nil
	})
	sc.Step(`^o último registro da auditoria tem os campos "([^"]*)"$`, func(campos string) error {
		regs := m.auditoria.Todos()
		if len(regs) == 0 {
			return fmt.Errorf("auditoria vazia")
		}
		if got := strings.Join(regs[len(regs)-1].Campos, ","); got != campos {
			return fmt.Errorf("campos do último registro: %q, esperava %q", got, campos)
		}
		return nil
	})

	// ---- produtos ----
	sc.Step(`^o operador cria pelo painel o produto "([^"]*)" na categoria "([^"]*)" com preço "([^"]*)"$`, func(nome, categoria, preco string) error {
		return m.webPOST("/painel/catalogo", url.Values{"nome": {nome}, "categoria": {categoria}, "preco": {preco}}, true)
	})
	sc.Step(`^o operador confirma a desativação do produto "([^"]*)"$`, func(nome string) error {
		p, err := m.produtoPorNome(nome)
		if err != nil {
			return err
		}
		return m.webPOST("/painel/catalogo/"+p.ID.String()+"/desativar", url.Values{}, true)
	})
	sc.Step(`^o produto "([^"]*)" está gravado na categoria "([^"]*)" com o preço ([\d.]+)$`, func(nome, categoria string, preco float64) error {
		p, err := m.produtoPorNome(nome)
		if err != nil {
			return err
		}
		if p.Categoria != categoria || p.Preco != preco || !p.Ativo {
			return fmt.Errorf("produto gravado: %+v", p)
		}
		return nil
	})
	sc.Step(`^o produto "([^"]*)" está inativo$`, func(nome string) error {
		p, err := m.produtoPorNome(nome)
		if err != nil {
			return err
		}
		if p.Ativo {
			return fmt.Errorf("produto %s continua ativo", nome)
		}
		return nil
	})
}

// formCriterio divide o critério nas linhas do formulário (campo, operador e valor).
func formCriterio(criterio string) url.Values {
	form := url.Values{}
	for _, cond := range strings.Split(criterio, "&&") {
		partes := strings.Fields(cond)
		for len(partes) < 3 {
			partes = append(partes, "")
		}
		form.Add("campo", partes[0])
		form.Add("operador", partes[1])
		form.Add("valor", strings.Join(partes[2:], " "))
	}
	return form
}

// enviarSegmento manda o formulário do segmento existente com o critério trocado.
func (m *Mundo) enviarSegmento(caminho, nome, criterio string) error {
	s, err := m.segmentoPorNome(nome)
	if err != nil {
		return err
	}
	form := formCriterio(criterio)
	form.Set("id", s.ID.String())
	form.Set("nome", s.Nome)
	form.Set("descricao", s.Descricao)
	form.Set("prioridade", strconv.Itoa(s.Prioridade))
	return m.webPOST(caminho, form, true)
}
