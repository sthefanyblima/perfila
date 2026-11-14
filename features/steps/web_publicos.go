package steps

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"

	"github.com/cucumber/godog"
)

// Passos @web dos públicos e do consentimento por finalidade (Spec 3.7.1, v1.9).
var reBaixar = regexp.MustCompile(`href="(/painel/publicos/exportacoes/[0-9a-f-]{36}/arquivo)"[^>]*data-teste="acao-baixar"`)

func registrarPassosPublicosWeb(sc *godog.ScenarioContext, m *Mundo) {
	sc.Step(`^a página mostra o selo "([^"]*)" da finalidade "([^"]*)"$`, func(estado, f string) error {
		return m.paginaTemMarcador(`data-teste="selo-finalidade" data-finalidade="`+f+`" data-estado="`+estado+`"`, true)
	})
	sc.Step(`^a página não mostra "([^"]*)"$`, func(t string) error {
		if strings.Contains(m.web.pagina, t) {
			return fmt.Errorf("a página mostra %q", t)
		}
		return nil
	})
	sc.Step(`^(?:que )?o operador (?:confirma|confirmou) "(conceder|revogar)" a finalidade "([^"]*)" do cliente "([^"]*)"$`, m.confirmaFinalidade)
	sc.Step(`^que o operador revogou a finalidade "([^"]*)" do cliente "([^"]*)"$`, func(f, c string) error {
		return m.confirmaFinalidade("revogar", f, c)
	})
	sc.Step(`^o operador abre o público do segmento "([^"]*)"$`, func(seg string) error {
		return m.webGET("/painel/publicos?segmento="+id("seg:"+seg).String(), true)
	})
	sc.Step(`^o operador abre a confirmação do público do segmento "([^"]*)"$`, func(seg string) error {
		return m.webGET("/painel/publicos/exportar?segmento="+id("seg:"+seg).String(), true)
	})
	sc.Step(`^(?:que )?o operador (?:confirma gerar|gerou) o arquivo do público do segmento "([^"]*)"$`, m.gerarPeloPainel)
	sc.Step(`^a página mostra que entram (\d+) e ficam de fora (\d+)$`, func(n, fora int) error {
		if err := m.paginaTemMarcador(fmt.Sprintf(`data-teste="publico-entram" data-n="%d"`, n), true); err != nil {
			return err
		}
		return m.paginaTemMarcador(fmt.Sprintf(`data-teste="publico-fora" data-n="%d"`, fora), true)
	})
	sc.Step(`^a página mostra (\d+) remoç(?:ão|ões) pendentes?$`, func(n int) error {
		return m.paginaTemMarcador(fmt.Sprintf(`data-teste="publico-pendentes" data-n="%d"`, n), true)
	})
	sc.Step(`^o operador baixa o arquivo gerado$`, func() error {
		if m.web.download == "" {
			return fmt.Errorf("nenhum arquivo foi gerado pelo painel")
		}
		// A sessão dura 8 h (Q-22): quem volta depois do prazo entra de novo antes de baixar.
		if err := m.entrouNoPainel(); err != nil {
			return err
		}
		m.requisitar(httptest.NewRequest(http.MethodGet, m.web.download, nil), true) // UUID real: sem troca de c-NNN
		return nil
	})
	sc.Step(`^o download é um CSV com o cabeçalho "([^"]*)" e (\d+) linhas?$`, func(cab string, n int) error {
		r := m.api.resposta
		if r.Code != http.StatusOK || !strings.HasPrefix(r.Header().Get("Content-Type"), "text/csv") ||
			!strings.Contains(r.Header().Get("Content-Disposition"), "attachment") {
			return fmt.Errorf("download: %d %q %q", r.Code, r.Header().Get("Content-Type"), r.Header().Get("Content-Disposition"))
		}
		linhas := strings.Split(strings.TrimSpace(m.web.pagina), "\n")
		if linhas[0] != cab || len(linhas)-1 != n {
			return fmt.Errorf("CSV:\n%s", m.web.pagina)
		}
		return nil
	})
	sc.Step(`^o download não contém "([^"]*)"$`, func(t string) error {
		if strings.Contains(m.web.pagina, t) {
			return fmt.Errorf("o download contém %q", t)
		}
		return nil
	})
}

func (m *Mundo) confirmaFinalidade(acao, finalidade, cliente string) error {
	base := "/painel/clientes/" + cliente + "/consentimento/" + acao
	if err := m.webGET(base+"?finalidade="+finalidade, true); err != nil {
		return err
	}
	return m.webPOST(base, url.Values{"finalidade": {finalidade}, "versaoTermo": {"v1"}}, true)
}

func (m *Mundo) gerarPeloPainel(seg string) error {
	sid := id("seg:" + seg).String()
	if err := m.webGET("/painel/publicos/exportar?segmento="+sid, true); err != nil {
		return err
	}
	if err := m.webPOST("/painel/publicos/exportar", url.Values{"segmento": {sid}}, true); err != nil {
		return err
	}
	s := reBaixar.FindStringSubmatch(m.web.pagina)
	if s == nil {
		return fmt.Errorf("a página não oferece baixar o arquivo gerado")
	}
	m.web.download = s[1]
	return nil
}
