package steps

import (
	"errors"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/service"
	"perfila/internal/shared/errs"
)

// Passos das correções Q-17 (falha de configuração) e Q-18 (uma recomendação ATIVA por cliente).
func registrarPassosConfiguracao(sc *godog.ScenarioContext, m *Mundo) {
	sc.Step(`^que existe o segmento ativo "([^"]*)" com prioridade (\d+) e critério "([^"]*)"$`, m.segmentoAtivo)
	sc.Step(`^a operação falha por erro de configuração$`, m.falhaDeConfiguracao)
	sc.Step(`^que o perfil do cliente "([^"]*)" aponta para um segmento inexistente$`, m.perfilComSegmentoInexistente)
	sc.Step(`^que uma recomendação do cliente "([^"]*)" já foi gerada$`, m.recomendacaoAnterior)
	sc.Step(`^o cliente "([^"]*)" tem (\d+) recomendaç(?:ão|ões) no estado "([^"]*)"$`, m.recomendacoesNoEstado)
	sc.Step(`^a recomendação anterior está no estado "([^"]*)"$`, m.recomendacaoAnteriorNoEstado)
}

// Spec: RN-06, Q-17.
func (m *Mundo) segmentoAtivo(nome string, prioridade int, criterio string) error {
	m.segmentos.Adicionar(entity.Segmento{ID: id("seg:" + nome), Nome: nome, Criterio: criterio, Prioridade: prioridade, Ativo: true})
	return nil
}

// Falha de configuração: erro do servidor (500), sem código de domínio que culpe o cliente. Spec: Q-02, Q-17.
func (m *Mundo) falhaDeConfiguracao() error {
	if m.err == nil {
		return errors.New("esperava falha de configuração, a operação teve sucesso")
	}
	if !errors.Is(m.err, service.ErrSegmentoMalConfigurado) && !errors.Is(m.err, service.ErrSemSegmentoAplicavel) {
		return fmt.Errorf("esperava falha de configuração, obtido %v", m.err)
	}
	if c := errs.Codigo(m.err); c != "" {
		return fmt.Errorf("falha de configuração não pode ter código de domínio, obtido %q", c)
	}
	return nil
}

// Spec: UC-04, Q-17.
func (m *Mundo) perfilComSegmentoInexistente(cliente string) error {
	p, err := m.perfis.BuscarPorCliente(m.ctx, id(cliente))
	if err != nil {
		return err
	}
	inexistente := uuid.New()
	p.SegmentoID = &inexistente
	_, err = m.perfis.Salvar(m.ctx, p)
	return err
}

// Spec: RN-08, Q-18.
func (m *Mundo) recomendacaoAnterior(cliente string) error {
	if err := m.geraRecomendacao(cliente); err != nil {
		return err
	}
	if err := m.semErro(); err != nil {
		return err
	}
	m.recomendacaoAnteriorID = m.recomendacao.ID
	return nil
}

func (m *Mundo) recomendacoesNoEstado(cliente string, n int, estado string) error {
	if err := m.semErro(); err != nil {
		return err
	}
	total := 0
	for _, r := range m.recomendacoes.Todas() {
		if r.ClienteID == id(cliente) && r.Status == entity.StatusRecomendacao(estado) {
			total++
		}
	}
	if total != n {
		return fmt.Errorf("recomendações %s do cliente %s: esperado %d, obtido %d", estado, cliente, n, total)
	}
	return nil
}

func (m *Mundo) recomendacaoAnteriorNoEstado(estado string) error {
	r, ok := m.recomendacoes.BuscarPorID(m.recomendacaoAnteriorID)
	if !ok {
		return errors.New("recomendação anterior não está no repositório")
	}
	if r.Status != entity.StatusRecomendacao(estado) {
		return fmt.Errorf("recomendação anterior: esperado %s, obtido %s", estado, r.Status)
	}
	return nil
}
