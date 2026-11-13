package steps

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
	"github.com/google/uuid"

	"perfila/internal/application/usecase"
	"perfila/internal/domain/entity"
	"perfila/internal/shared/errs"
)

// Passos da edição de segmentos e catálogo (@aplicacao, @dominio). Spec: RN-14, Q-33, Q-42, Q-43.
type estadoConfig struct {
	previa     usecase.PreviaSegmentacao
	testeCasa  bool
	alteracoes []string // campos da última alteração
}

const atorRe = `o operador "([^"]*)" com o papel "([^"]*)"`

func registrarPassosSegmentos(sc *godog.ScenarioContext, m *Mundo) {
	// ---- segmentos ----
	sc.Step(`^que `+atorRe+` criou o segmento "([^"]*)" com prioridade (\d+) e critério "([^"]*)"$`,
		func(login, papel, nome string, prio int, criterio string) error {
			return m.exigirSucesso(m.criarSegmento(login, papel, nome, prio, criterio))
		})
	sc.Step(`^`+atorRe+` cria o segmento "([^"]*)" com prioridade (-?\d+) e critério "([^"]*)"$`,
		func(login, papel, nome string, prio int, criterio string) error {
			m.criarSegmento(login, papel, nome, prio, criterio)
			return nil
		})
	sc.Step(`^`+atorRe+` muda o critério do segmento "([^"]*)" para "([^"]*)"$`, func(login, papel, nome, criterio string) error {
		return m.editarSegmento(login, papel, nome, func(s *entity.Segmento) { s.Criterio = criterio })
	})
	sc.Step(`^`+atorRe+` muda a prioridade do segmento "([^"]*)" para (\d+)$`, func(login, papel, nome string, prio int) error {
		return m.editarSegmento(login, papel, nome, func(s *entity.Segmento) { s.Prioridade = prio })
	})
	sc.Step(`^`+atorRe+` (ativa|desativa) o segmento "([^"]*)"$`, func(login, papel, acao, nome string) error {
		return m.situacaoSegmento(login, papel, acao == "ativa", nome)
	})
	sc.Step(`^que `+atorRe+` desativou o segmento "([^"]*)"$`, func(login, papel, nome string) error {
		if err := m.situacaoSegmento(login, papel, false, nome); err != nil {
			return err
		}
		return m.exigirSucesso(m.err)
	})
	sc.Step(`^que o cliente "([^"]*)" foi segmentado$`, func(c string) error {
		_, err := m.uc().segmentar.Executar(m.ctx, id(c))
		return err
	})
	sc.Step(`^a operação falha com o erro "([^"]*)" no campo "([^"]*)"$`, func(codigo, campo string) error {
		if err := m.falhaComErro(codigo); err != nil {
			return err
		}
		var ec *errs.ErroCampo
		if !errors.As(m.err, &ec) || ec.Campo != campo {
			return fmt.Errorf("esperava erro no campo %q, obtido %v", campo, m.err)
		}
		return nil
	})
	sc.Step(`^a mensagem do erro cita "([^"]*)"$`, func(trecho string) error {
		var ec *errs.ErroCampo
		if !errors.As(m.err, &ec) || !strings.Contains(ec.Mensagem, trecho) {
			return fmt.Errorf("esperava mensagem citando %q, obtido %v", trecho, m.err)
		}
		return nil
	})
	sc.Step(`^o segmento "([^"]*)" continua com o critério "([^"]*)"$`, func(nome, criterio string) error {
		s, err := m.segmentoPorNome(nome)
		if err != nil {
			return err
		}
		if s.Criterio != criterio {
			return fmt.Errorf("segmento %s com critério %q, esperava %q", nome, s.Criterio, criterio)
		}
		return nil
	})
	sc.Step(`^os segmentos ativos são (.+)$`, func(lista string) error {
		ativos, err := m.segmentos.ListarAtivos(m.ctx)
		if err != nil {
			return err
		}
		obtidos := make([]string, len(ativos))
		for i, s := range ativos {
			obtidos[i] = s.Nome
		}
		if esperados := citados(lista); strings.Join(obtidos, ",") != strings.Join(esperados, ",") {
			return fmt.Errorf("segmentos ativos %v, esperava %v", obtidos, esperados)
		}
		return nil
	})

	// ---- prévia e teste de critério ----
	sc.Step(`^que existem os clientes com perfil salvo:$`, m.clientesComPerfil)
	sc.Step(`^que existe o cliente "([^"]*)" ativo, com consentimento válido e sem perfil$`, func(c string) error {
		return m.clienteComConsentimento(c, true, entity.StatusConcedido)
	})
	sc.Step(`^`+atorRe+` pede a prévia do critério "([^"]*)" para o segmento "([^"]*)"$`, func(login, papel, criterio, nome string) error {
		s, err := m.segmentoPorNome(nome)
		if err != nil {
			return err
		}
		s.Criterio = criterio
		m.cfg.previa, m.err = m.uc().prever.Executar(m.ctx, ator(login, papel), s)
		return nil
	})
	sc.Step(`^a prévia mostra:$`, func(tabela *godog.Table) error {
		var obtido, esperado []string
		for _, l := range m.cfg.previa.Linhas {
			obtido = append(obtido, fmt.Sprintf("%s:%d:%d", l.Segmento.Nome, l.Atual, l.Previsto))
		}
		for _, r := range tabela.Rows[1:] {
			esperado = append(esperado, r.Cells[0].Value+":"+r.Cells[1].Value+":"+r.Cells[2].Value)
		}
		if m.err != nil || strings.Join(obtido, " ") != strings.Join(esperado, " ") {
			return fmt.Errorf("prévia %v (erro %v), esperava %v", obtido, m.err, esperado)
		}
		return nil
	})
	sc.Step(`^a prévia considera (\d+) clientes, (\d+) deles? sem perfil$`, func(total, sem int) error {
		if p := m.cfg.previa; p.Total != total || p.SemPerfil != sem {
			return fmt.Errorf("prévia com %d clientes e %d sem perfil", p.Total, p.SemPerfil)
		}
		return nil
	})
	sc.Step(`^o critério "([^"]*)" é testado com frequência de compras (\d+), ticket médio ([\d.]+) e score ([\d.]+)$`,
		func(criterio string, freq int, ticket, score float64) error {
			m.cfg.testeCasa, m.err = m.uc().prever.TestarCriterio(criterio, freq, ticket, score)
			return nil
		})
	sc.Step(`^o teste do critério dá "(casa|não casa)"$`, func(r string) error {
		if m.err != nil || m.cfg.testeCasa != (r == "casa") {
			return fmt.Errorf("teste deu casa=%v (erro %v), esperava %q", m.cfg.testeCasa, m.err, r)
		}
		return nil
	})

	// ---- produtos ----
	sc.Step(`^que `+atorRe+` desativou o produto "([^"]*)"$`, func(login, papel, nome string) error {
		p, err := m.produtoPorNome(nome)
		if err != nil {
			return err
		}
		_, err = m.uc().salvarProduto.Desativar(m.ctx, ator(login, papel), p.ID)
		return err
	})
	sc.Step(`^que `+atorRe+` criou o produto "([^"]*)" na categoria "([^"]*)" com preço ([\d.]+)$`,
		func(login, papel, nome, categoria string, preco float64) error {
			_, err := m.uc().salvarProduto.Executar(m.ctx, ator(login, papel), entity.Produto{Nome: nome, Categoria: categoria, Preco: preco})
			return err
		})
	sc.Step(`^`+atorRe+` cria o produto "([^"]*)" na categoria "([^"]*)" com preço (-?[\d.]+)$`,
		func(login, papel, nome, categoria string, preco float64) error {
			_, m.err = m.uc().salvarProduto.Executar(m.ctx, ator(login, papel), entity.Produto{Nome: nome, Categoria: categoria, Preco: preco})
			return nil
		})
	sc.Step(`^`+atorRe+` muda o preço do produto "([^"]*)" para ([\d.]+)$`, func(login, papel, nome string, preco float64) error {
		p, err := m.produtoPorNome(nome)
		if err != nil {
			return err
		}
		p.Preco = preco
		_, m.err = m.uc().salvarProduto.Executar(m.ctx, ator(login, papel), p)
		return nil
	})
	sc.Step(`^o produto "([^"]*)" está gravado com o preço ([\d.]+)$`, func(nome string, preco float64) error {
		p, err := m.produtoPorNome(nome)
		if err != nil {
			return err
		}
		if m.err != nil || p.Preco != preco {
			return fmt.Errorf("produto %s com preço %v (erro %v), esperava %v", nome, p.Preco, m.err, preco)
		}
		return nil
	})
}

// ator monta o operador da sessão; os casos de uso só olham papel e situação (RN-11).
func ator(login, papel string) entity.Operador {
	return entity.Operador{ID: id("op:" + login), Login: login, Papel: entity.PapelOperador(papel), Ativo: true}
}

func (m *Mundo) exigirSucesso(err error) error {
	if err != nil {
		return fmt.Errorf("preparação falhou: %w", err)
	}
	return nil
}

func (m *Mundo) criarSegmento(login, papel, nome string, prio int, criterio string) error {
	var alt usecase.AlteracaoSegmento
	alt, m.err = m.uc().salvarSegmento.Executar(m.ctx, ator(login, papel), entity.Segmento{Nome: nome, Criterio: criterio, Prioridade: prio})
	m.cfg.alteracoes = alt.Campos
	return m.err
}

func (m *Mundo) editarSegmento(login, papel, nome string, mudar func(*entity.Segmento)) error {
	s, err := m.segmentoPorNome(nome)
	if err != nil {
		return err
	}
	mudar(&s)
	var alt usecase.AlteracaoSegmento
	alt, m.err = m.uc().salvarSegmento.Executar(m.ctx, ator(login, papel), s)
	m.cfg.alteracoes = alt.Campos
	return nil
}

func (m *Mundo) situacaoSegmento(login, papel string, ativar bool, nome string) error {
	s, err := m.segmentoPorNome(nome)
	if err != nil {
		return err
	}
	uc := m.uc().salvarSegmento
	if ativar {
		_, m.err = uc.Ativar(m.ctx, ator(login, papel), s.ID)
	} else {
		_, m.err = uc.Desativar(m.ctx, ator(login, papel), s.ID)
	}
	return nil
}

func (m *Mundo) segmentoPorNome(nome string) (entity.Segmento, error) {
	todos, err := m.segmentos.ListarTodos(m.ctx)
	if err != nil {
		return entity.Segmento{}, err
	}
	for _, s := range todos {
		if s.Nome == nome {
			return s, nil
		}
	}
	return entity.Segmento{}, fmt.Errorf("segmento %q não existe no cenário", nome)
}

func (m *Mundo) produtoPorNome(nome string) (entity.Produto, error) {
	for _, p := range m.produtos.Todos() {
		if p.Nome == nome {
			return p, nil
		}
	}
	return entity.Produto{}, fmt.Errorf("produto %q não existe no cenário", nome)
}

var reCitado = regexp.MustCompile(`"([^"]*)"`)

// citados extrai os textos entre aspas de `"a", "b" e "c"`.
func citados(s string) []string {
	var out []string
	for _, m := range reCitado.FindAllStringSubmatch(s, -1) {
		out = append(out, m[1])
	}
	return out
}

func (m *Mundo) clienteComConsentimento(c string, ativo bool, status entity.StatusConsentimento) error {
	cli := entity.Cliente{ID: id(c), Nome: "Cliente " + c, Email: c + "@exemplo.test", DataCadastro: m.clock.Agora(), Ativo: ativo}
	if err := m.clientes.Salvar(m.ctx, cli); err != nil {
		return err
	}
	return m.consentimentos.Salvar(m.ctx, entity.ConsentimentoLGPD{ID: uuid.New(), ClienteID: cli.ID, Status: status,
		Finalidade: "segmentacao", VersaoTermo: "v1", DataConsentimento: m.clock.Agora(), DataAtualizacao: m.clock.Agora()})
}

// clientesComPerfil: | cliente | consentimento | ativo | ticket | frequencia | score | segmento |
func (m *Mundo) clientesComPerfil(tabela *godog.Table) error {
	for _, r := range tabela.Rows[1:] {
		v := make([]string, len(r.Cells))
		for i, c := range r.Cells {
			v[i] = c.Value
		}
		if err := m.clienteComConsentimento(v[0], v[2] == "sim", entity.StatusConsentimento(v[1])); err != nil {
			return err
		}
		ticket, err1 := strconv.ParseFloat(v[3], 64)
		freq, err2 := strconv.Atoi(v[4])
		score, err3 := strconv.ParseFloat(v[5], 64)
		if err := errors.Join(err1, err2, err3); err != nil {
			return err
		}
		seg := id("seg:" + v[6])
		if _, err := m.perfis.Salvar(m.ctx, entity.PerfilCliente{ID: uuid.New(), ClienteID: id(v[0]), TicketMedio: ticket,
			FrequenciaCompras: freq, ScoreEngajamento: score, SegmentoID: &seg}); err != nil {
			return err
		}
	}
	return nil
}
