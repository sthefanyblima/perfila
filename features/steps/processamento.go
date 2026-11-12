package steps

import (
	"fmt"

	"github.com/cucumber/godog"
)

// Passos do processamento etapa por etapa (RN-09, Q-26).
func registrarPassosProcessamento(sc *godog.ScenarioContext, m *Mundo) {
	sc.Step(`^o cliente "([^"]*)" é processado$`, m.processa)
	sc.Step(`^as etapas do processamento são "([^"]*)", "([^"]*)", "([^"]*)" e "([^"]*)"$`, m.etapas)
	sc.Step(`^o processamento coletou (\d+) eventos?$`, m.eventosProcessados)
	sc.Step(`^o processamento não falha$`, m.semErro)
}

func (m *Mundo) processa(cliente string) error {
	m.processamento, m.err = m.uc().processar.Executar(m.ctx, id(cliente))
	if m.processamento.Recomendacao != nil {
		m.recomendacao = *m.processamento.Recomendacao
	}
	return nil
}

func (m *Mundo) etapas(a, b, c, d string) error {
	quer := [4]string{a, b, c, d}
	for i, e := range m.processamento.Etapas {
		if string(e.Estado) != quer[i] {
			return fmt.Errorf("etapa %s: esperado %q, obtido %q (erro: %v)", e.Nome, quer[i], e.Estado, e.Erro)
		}
	}
	return nil
}

func (m *Mundo) eventosProcessados(n int) error {
	if m.processamento.Eventos != n {
		return fmt.Errorf("eventos coletados: esperado %d, obtido %d", n, m.processamento.Eventos)
	}
	return nil
}
