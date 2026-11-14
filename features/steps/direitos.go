package steps

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"perfila/internal/application/usecase"
	"perfila/internal/domain/entity"
	"perfila/internal/shared/errs"
)

// estadoDireitos guarda os resultados dos cenários de direitos do titular e retenção.
// Spec: v1.9, RN-17, RN-18, Q-48 a Q-51.
type estadoDireitos struct {
	prazoDias int
	dados     usecase.DadosTitular
	atingidos int
}

// prazoRetencaoPadrao é o RETENCAO_DIAS_REVOGADO padrão (Spec 3.8).
const prazoRetencaoPadrao = 180

func registrarPassosDireitos(sc *godog.ScenarioContext, m *Mundo) {
	// ---- Dado ----
	sc.Step(`^que o cliente "([^"]*)" tem dados de comportamento desde "([^"]*)"$`, m.temComportamentoDesde)
	sc.Step(`^que o cliente "([^"]*)" revogou o consentimento em "([^"]*)"$`, m.revogouEm)
	sc.Step(`^que os dados do cliente "([^"]*)" foram eliminados$`, m.foiEliminado)
	sc.Step(`^que o prazo de retenção é de (\d+) dias$`, func(n int) error { m.direitos.prazoDias = n; return nil })
	sc.Step(`^que a retenção foi aplicada$`, func() error { _, err := m.uc().retencao.Executar(m.ctx); return err })

	// ---- Quando ----
	sc.Step(`^o cadastro do cliente "([^"]*)" é corrigido para o nome "([^"]*)" e o e-mail "([^"]*)"$`, m.corrigeCadastro)
	sc.Step(`^os dados do cliente "([^"]*)" são exportados$`, m.exporta)
	sc.Step(`^os dados do cliente "([^"]*)" são eliminados com a confirmação "([^"]*)"$`, m.elimina)
	sc.Step(`^a retenção é aplicada$`, func() error {
		m.direitos.atingidos, m.err = m.uc().retencao.Executar(m.ctx)
		return nil
	})
	sc.Step(`^a ação "([^"]*)" é tentada para o cliente "([^"]*)"$`, m.tentaAcao)

	// ---- Então ----
	sc.Step(`^o cadastro do cliente "([^"]*)" tem o nome "([^"]*)"$`, func(c, nome string) error {
		return m.confereCadastro(c, func(cli entity.Cliente) string { return cli.Nome }, nome)
	})
	sc.Step(`^o cadastro do cliente "([^"]*)" tem o e-mail "([^"]*)"$`, func(c, email string) error {
		return m.confereCadastro(c, func(cli entity.Cliente) string { return cli.Email }, email)
	})
	sc.Step(`^o cadastro do cliente "([^"]*)" está anonimizado, inativo e eliminado em "([^"]*)"$`, m.anonimizado)
	sc.Step(`^o cliente "([^"]*)" não tem eventos, perfil nem recomendações$`, func(c string) error { return m.comportamento(c, false) })
	sc.Step(`^o cliente "([^"]*)" continua com os dados de comportamento$`, func(c string) error { return m.comportamento(c, true) })
	sc.Step(`^a exportação traz o cadastro com o nome "([^"]*)"$`, m.exportacaoNome)
	sc.Step(`^a exportação traz o consentimento "([^"]*)" com (\d+) movimentos?$`, m.exportacaoConsentimento)
	sc.Step(`^a exportação traz (\d+) eventos?, o perfil no segmento "([^"]*)" e (\d+) recomendaç(?:ão|ões)$`, m.exportacaoCompleta)
	sc.Step(`^a exportação traz (\d+) eventos?, sem perfil e sem recomendações$`, m.exportacaoSemPerfil)
	sc.Step(`^a retenção atinge (\d+) clientes?$`, m.retencaoAtinge)

	// ---- painel (@web) ----
	sc.Step(`^o operador corrige o cadastro do cliente "([^"]*)" para o nome "([^"]*)" e o e-mail "([^"]*)"$`, func(c, nome, email string) error {
		return m.webPOST("/painel/clientes/"+c+"/cadastro", url.Values{"nome": {nome}, "email": {email}}, true)
	})
	sc.Step(`^o operador baixa os dados do cliente "([^"]*)"$`, func(c string) error { return m.webGET("/painel/clientes/"+c+"/dados", true) })
	sc.Step(`^a resposta é um arquivo JSON anexo com o cadastro de "([^"]*)" e (\d+) eventos?$`, m.arquivoExportado)
	sc.Step(`^o operador confirma a eliminação dos dados do cliente "([^"]*)" digitando "([^"]*)"$`, func(c, nome string) error {
		return m.webPOST("/painel/clientes/"+c+"/eliminar", url.Values{"confirmacao": {nome}}, true)
	})
	sc.Step(`^o operador confirma a aplicação da retenção$`, func() error {
		return m.webPOST("/painel/operacao/retencao", url.Values{}, true)
	})
	sc.Step(`^o log registra o pedido "([^"]*)" atendido para o cliente "([^"]*)"$`, m.logRegistraPedido)
}

// arquivoExportado confere o anexo JSON: cabeçalhos, nome do titular e quantidade de eventos (Q-49).
func (m *Mundo) arquivoExportado(nome string, eventos int) error {
	r := m.api.resposta
	if cd := r.Header().Get("Content-Disposition"); !strings.HasPrefix(cd, "attachment;") || !strings.HasSuffix(cd, `.json"`) {
		return fmt.Errorf("Content-Disposition inesperado: %q", cd)
	}
	if ct := r.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		return fmt.Errorf("Content-Type inesperado: %q", ct)
	}
	var arq struct {
		VersaoFormato string `json:"versaoFormato"`
		Cliente       struct {
			Nome string `json:"nome"`
		} `json:"cliente"`
		Eventos []json.RawMessage `json:"eventos"`
	}
	if err := json.Unmarshal(r.Body.Bytes(), &arq); err != nil {
		return fmt.Errorf("arquivo não é JSON: %v", err)
	}
	if arq.VersaoFormato != "1" || arq.Cliente.Nome != nome || len(arq.Eventos) != eventos {
		return fmt.Errorf("arquivo: versão %q, nome %q, %d eventos", arq.VersaoFormato, arq.Cliente.Nome, len(arq.Eventos))
	}
	return nil
}

// logRegistraPedido confere a auditoria do pedido: direito_atendido com o pedido e o ID (Q-53).
func (m *Mundo) logRegistraPedido(pedido, c string) error {
	for _, linha := range strings.Split(m.web.log.String(), "\n") {
		if strings.Contains(linha, "msg=direito_atendido") && strings.Contains(linha, "pedido="+pedido) &&
			strings.Contains(linha, "clienteId="+id(c).String()) {
			return nil
		}
	}
	return fmt.Errorf("log sem direito_atendido do pedido %q para %s", pedido, c)
}

// emData executa f com o relógio do cenário em t e depois o devolve ao instante atual (B-07).
func (m *Mundo) emData(data string, f func() error) error {
	t, err := instante(data)
	if err != nil {
		return err
	}
	agora := m.clock.Agora()
	m.clock.Definir(t)
	defer m.clock.Definir(agora)
	return f()
}

// temComportamentoDesde concede (se preciso), grava duas compras em livros e processa, tudo na data
// informada: eventos, perfil no segmento padrao e uma recomendação ativa.
func (m *Mundo) temComportamentoDesde(c, data string) error {
	m.segmentos.Adicionar(entity.Segmento{ID: id("seg:padrao"), Nome: "padrao", Prioridade: 99, Ativo: true})
	return m.emData(data, func() error {
		uc := m.uc()
		cons, err := uc.consentimento.Consultar(m.ctx, id(c))
		if err != nil {
			return err
		}
		if cons.Status != entity.StatusConcedido {
			if err := uc.consentimento.Conceder(m.ctx, id(c), "segmentacao", "v1"); err != nil {
				return err
			}
		}
		agora := m.clock.Agora()
		for i, valor := range []float64{120, 80} {
			if err := m.gravaEvento(c, entity.TipoCompra, valor, "livros", agora.Add(-time.Duration(i+1)*time.Hour)); err != nil {
				return err
			}
		}
		_, err = uc.processar.Executar(m.ctx, id(c))
		return err
	})
}

// revogouEm revoga na data informada; quem ainda não tinha concedido concede antes, no mesmo
// instante (revogar pressupõe a concessão): o histórico fica com os dois movimentos.
func (m *Mundo) revogouEm(c, data string) error {
	return m.emData(data, func() error {
		uc := m.uc()
		cons, err := uc.consentimento.Consultar(m.ctx, id(c))
		if err != nil {
			return err
		}
		if cons.Status != entity.StatusConcedido {
			if err := uc.consentimento.Conceder(m.ctx, id(c), "segmentacao", "v1"); err != nil {
				return err
			}
		}
		return uc.consentimento.Revogar(m.ctx, id(c))
	})
}

func (m *Mundo) foiEliminado(c string) error {
	cli, err := m.clientes.BuscarPorID(m.ctx, id(c))
	if err != nil {
		return err
	}
	return m.uc().eliminar.Executar(m.ctx, id(c), cli.Nome)
}

func (m *Mundo) corrigeCadastro(c, nome, email string) error {
	_, m.err = m.uc().atualizarCadastro.Executar(m.ctx, id(c), nome, email)
	return nil
}

func (m *Mundo) exporta(c string) error {
	m.direitos.dados, m.err = m.uc().exportar.Executar(m.ctx, id(c))
	return nil
}

func (m *Mundo) elimina(c, confirmacao string) error {
	m.err = m.uc().eliminar.Executar(m.ctx, id(c), confirmacao)
	return nil
}

// tentaAcao tenta cada escrita possível sobre o cliente; o resultado fica em m.err.
func (m *Mundo) tentaAcao(acao, c string) error {
	uc, cid := m.uc(), id(c)
	switch acao {
	case "corrigir o cadastro":
		_, m.err = uc.atualizarCadastro.Executar(m.ctx, cid, "Nome Novo Sintético", "novo@exemplo.test")
	case "conceder o consentimento":
		m.err = uc.consentimento.Conceder(m.ctx, cid, "segmentacao", "v1")
	case "revogar o consentimento":
		m.err = uc.consentimento.Revogar(m.ctx, cid)
	case "registrar um evento":
		_, m.err = uc.coletar.Registrar(m.ctx, entity.EventoComportamental{ClienteID: cid, Tipo: entity.TipoNavegacao,
			Origem: "web", Timestamp: m.clock.Agora()})
	case "processar":
		_, m.err = uc.processar.Executar(m.ctx, cid)
	case "eliminar os dados":
		m.err = uc.eliminar.Executar(m.ctx, cid, entity.NomeEliminado)
	default:
		return fmt.Errorf("ação desconhecida: %q", acao)
	}
	return nil
}

func (m *Mundo) confereCadastro(c string, campo func(entity.Cliente) string, quer string) error {
	cli, err := m.clientes.BuscarPorID(m.ctx, id(c))
	if err != nil {
		return err
	}
	if got := campo(cli); got != quer {
		return fmt.Errorf("cadastro: esperado %q, obtido %q", quer, got)
	}
	return nil
}

func (m *Mundo) anonimizado(c, data string) error {
	quando, err := instante(data)
	if err != nil {
		return err
	}
	cli, err := m.clientes.BuscarPorID(m.ctx, id(c))
	if err != nil {
		return err
	}
	if cli.Nome != entity.NomeEliminado || cli.Email != entity.EmailEliminado || cli.Ativo ||
		cli.DataEliminacao == nil || !cli.DataEliminacao.Equal(quando) {
		return fmt.Errorf("cadastro não anonimizado como esperado: %+v", cli)
	}
	return nil
}

// comportamento confere se o cliente tem (ou não tem) eventos, perfil e recomendações gravados.
func (m *Mundo) comportamento(c string, tem bool) error {
	eventos, err := m.eventos.BuscarPorCliente(m.ctx, id(c))
	if err != nil {
		return err
	}
	_, err = m.perfis.BuscarPorCliente(m.ctx, id(c))
	temPerfil := err == nil
	if err != nil && !errors.Is(err, errs.ErrPerfilNaoEncontrado) {
		return err
	}
	recs, err := m.recomendacoes.ListarPorCliente(m.ctx, id(c))
	if err != nil {
		return err
	}
	achou := [3]bool{len(eventos) > 0, temPerfil, len(recs) > 0}
	if achou != [3]bool{tem, tem, tem} {
		return fmt.Errorf("comportamento (eventos, perfil, recomendações): esperado %v, obtido %v", tem, achou)
	}
	return nil
}

func (m *Mundo) exportacaoNome(nome string) error {
	if m.err != nil {
		return m.err
	}
	if got := m.direitos.dados.Cliente.Nome; got != nome {
		return fmt.Errorf("nome exportado: esperado %q, obtido %q", nome, got)
	}
	return nil
}

func (m *Mundo) exportacaoConsentimento(estado string, n int) error {
	d := m.direitos.dados
	if string(d.Consentimento.Status) != estado || len(d.Historico) != n {
		return fmt.Errorf("consentimento exportado: esperado %s com %d movimentos, obtido %s com %d",
			estado, n, d.Consentimento.Status, len(d.Historico))
	}
	return nil
}

func (m *Mundo) exportacaoCompleta(eventos int, segmento string, recs int) error {
	d := m.direitos.dados
	if len(d.Eventos) != eventos || d.Perfil == nil || d.Segmento == nil || d.Segmento.Nome != segmento || len(d.Recomendacoes) != recs {
		return fmt.Errorf("exportação: %d eventos, perfil %v, segmento %v, %d recomendações", len(d.Eventos), d.Perfil, d.Segmento, len(d.Recomendacoes))
	}
	for i := 1; i < len(d.Eventos); i++ {
		if d.Eventos[i].Timestamp.Before(d.Eventos[i-1].Timestamp) {
			return fmt.Errorf("eventos exportados fora da ordem cronológica")
		}
	}
	for _, r := range d.Recomendacoes {
		if _, ok := d.Produtos[r.ProdutoID]; !ok {
			return fmt.Errorf("produto %s da recomendação ausente da exportação", r.ProdutoID)
		}
	}
	return nil
}

func (m *Mundo) exportacaoSemPerfil(eventos int) error {
	d := m.direitos.dados
	if len(d.Eventos) != eventos || d.Perfil != nil || len(d.Recomendacoes) != 0 {
		return fmt.Errorf("exportação: %d eventos, perfil %v, %d recomendações", len(d.Eventos), d.Perfil, len(d.Recomendacoes))
	}
	return nil
}

func (m *Mundo) retencaoAtinge(n int) error {
	if m.err != nil {
		return m.err
	}
	if m.direitos.atingidos != n {
		return fmt.Errorf("retenção: esperado %d clientes, obtido %d", n, m.direitos.atingidos)
	}
	return nil
}
