package steps

import (
	"fmt"

	"github.com/cucumber/godog"
)

// Passos do consentimento por finalidade (Spec v1.9, Q-48). Os passos sem finalidade continuam
// valendo para "segmentacao".
func registrarPassosFinalidade(sc *godog.ScenarioContext, m *Mundo) {
	sc.Step(`^que o consentimento do cliente "([^"]*)" para "([^"]*)" está no estado "([^"]*)"$`, m.consentimentoNoEstadoPara)
	sc.Step(`^o cliente "([^"]*)" revoga o consentimento para "([^"]*)"$`, m.revogaPara)
	sc.Step(`^o consentimento do cliente "([^"]*)" para "([^"]*)" (?:está|permanece) no estado "([^"]*)"$`, m.confereEstadoPara)
	sc.Step(`^o último movimento do cliente "([^"]*)" é da finalidade "([^"]*)"$`, func(nome, finalidade string) error {
		movs, err := m.uc().consentimento.Historico(m.ctx, id(nome))
		if err != nil {
			return err
		}
		if len(movs) == 0 {
			return fmt.Errorf("o cliente %s não tem movimentos", nome)
		}
		if got := movs[len(movs)-1].Finalidade; got != finalidade {
			return fmt.Errorf("finalidade do último movimento: esperado %s, obtido %s", finalidade, got)
		}
		return nil
	})
}

func (m *Mundo) confereEstadoPara(nome, finalidade, estado string) error {
	if err := m.recarregaConsentimentoPara(nome, finalidade); err != nil {
		return err
	}
	if string(m.consentimento.Status) != estado {
		return fmt.Errorf("estado de %s: esperado %s, obtido %s", finalidade, estado, m.consentimento.Status)
	}
	return nil
}
