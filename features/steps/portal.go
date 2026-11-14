package steps

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/google/uuid"

	"perfila/internal/application/usecase"
	"perfila/internal/domain/entity"
)

// Passos do portal do titular nas camadas @aplicacao e @dominio. Spec: v1.10, RN-19, RN-20, Q-55 a Q-58.

type estadoPortal struct {
	urlBase   string
	clienteID uuid.UUID // quem entrou pelo link (aplicação)
	meus      usecase.MeusDados
	pedido    entity.PedidoEliminacao
	pedidos   []usecase.LinhaPedido
}

var reLink = regexp.MustCompile(`https?://\S+?/titular/acesso\?t=([0-9a-f]+)`)

func registrarPassosPortal(sc *godog.ScenarioContext, m *Mundo) {
	sc.Step(`^que o portal do titular está no ar no endereço "([^"]*)"$`, m.portalNoAr)
	sc.Step(`^que os dados do cliente "([^"]*)" estão cadastrados$`, func(string) error { return nil })
	sc.Step(`^que existe o cliente "([^"]*)" ativo com o nome "([^"]*)" e o e-mail "([^"]*)"$`, func(c, nome, email string) error {
		return m.clientes.Salvar(m.ctx, entity.Cliente{ID: id(c), Nome: nome, Email: email, DataCadastro: m.clock.Agora(), Ativo: true})
	})
	sc.Step(`^que se passaram (\d+) minutos$`, func(n int) error { m.clock.Avancar(time.Duration(n) * time.Minute); return nil })

	// ---- pedido e uso do link ----
	sc.Step(`^o titular pede acesso com o e-mail "([^"]*)"$`, func(e string) error {
		m.err = m.uc().solicitarAcesso.Executar(m.ctx, e)
		return nil
	})
	sc.Step(`^que o titular pediu (?:um link|(\d+) links) com o e-mail "([^"]*)"$`, m.pediuLinks)
	sc.Step(`^o pedido de acesso é aceito sem erro$`, m.semErro)
	sc.Step(`^a caixa de saída tem (\d+) mensage(?:m|ns)$`, func(n int) error { return m.caixaTem(n, "") })
	sc.Step(`^a caixa de saída tem (\d+) mensage(?:m|ns) para "([^"]*)"$`, m.caixaTem)
	sc.Step(`^o link enviado começa com "([^"]*)"$`, func(p string) error {
		link, _, err := m.linkEnviado()
		if err == nil && !strings.HasPrefix(link, p) {
			err = fmt.Errorf("link %q não começa com %q", link, p)
		}
		return err
	})
	sc.Step(`^o link enviado vale até "([^"]*)"$`, m.linkValeAte)
	sc.Step(`^o sistema guarda só o hash do token do link enviado$`, m.soHash)
	sc.Step(`^cada link enviado entra num cadastro diferente$`, m.linksEmCadastrosDiferentes)
	sc.Step(`^o titular usa o link enviado$`, func() error {
		_, token, err := m.linkEnviado()
		if err != nil {
			return err
		}
		m.usarToken(token)
		return nil
	})
	sc.Step(`^o titular usa o link com o token "([^"]*)"$`, func(t string) error { m.usarToken(t); return nil })
	sc.Step(`^o titular entra como o cliente "([^"]*)"$`, func(c string) error {
		if m.err != nil || m.portal.clienteID != id(c) {
			return fmt.Errorf("esperava entrar como %s, obtido %v (%v)", c, m.portal.clienteID, m.err)
		}
		return nil
	})
	sc.Step(`^usar o mesmo link de novo falha com o erro "([^"]*)"$`, func(codigo string) error {
		_, token, err := m.linkEnviado()
		if err != nil {
			return err
		}
		m.usarToken(token)
		return m.falhaComErro(codigo)
	})
	sc.Step(`^o resultado do acesso é "([^"]*)"$`, func(r string) error {
		if r == "entrou" {
			return m.semErro()
		}
		return m.falhaComErro(r)
	})

	// ---- meus dados e pedido de eliminação ----
	sc.Step(`^o titular do cliente "([^"]*)" consulta os próprios dados$`, func(c string) error {
		m.portal.meus, m.err = m.uc().meusDados.Executar(m.ctx, id(c))
		return nil
	})
	sc.Step(`^os dados consultados trazem o nome "([^"]*)" e o consentimento "([^"]*)" com (\d+) movimentos?$`, m.meusDadosTrazem)
	sc.Step(`^os dados consultados não trazem pedido de eliminação aberto$`, func() error {
		if m.err != nil || m.portal.meus.PedidoAberto != nil {
			return fmt.Errorf("pedido aberto inesperado: %+v (%v)", m.portal.meus.PedidoAberto, m.err)
		}
		return nil
	})
	sc.Step(`^o titular do cliente "([^"]*)" pede a eliminação dos dados$`, m.pedeEliminacao)
	sc.Step(`^que o titular do cliente "([^"]*)" pediu a eliminação dos dados$`, func(c string) error {
		if err := m.pedeEliminacao(c); err != nil {
			return err
		}
		return m.err
	})
	sc.Step(`^que os clientes "([^"]*)" e "([^"]*)" pediram a eliminação dos dados$`, func(a, b string) error {
		for _, c := range []string{a, b} {
			if _, err := m.uc().pedirEliminacao.Executar(m.ctx, id(c)); err != nil {
				return err
			}
		}
		return nil
	})
	sc.Step(`^o resultado do pedido de eliminação é "([^"]*)"$`, func(r string) error {
		if r == string(entity.PedidoAberto) {
			if m.err != nil || m.portal.pedido.Status != entity.PedidoAberto {
				return fmt.Errorf("esperava pedido ABERTO, obtido %+v (%v)", m.portal.pedido, m.err)
			}
			return nil
		}
		if err := m.falhaComErro(r); err != nil {
			return err
		}
		return m.pedidosDoCliente("c-001", 0, "")
	})
	sc.Step(`^o cliente "([^"]*)" tem (\d+) pedidos? de eliminação "([^"]*)"$`, func(c string, n int, s string) error {
		return m.pedidosDoCliente(c, n, entity.StatusPedido(s))
	})
	sc.Step(`^o cliente "([^"]*)" não tem links de acesso guardados$`, func(c string) error {
		for _, t := range m.tokens.Todos() {
			if t.ClienteID == id(c) {
				return fmt.Errorf("o cliente %s ainda tem o token %s", c, t.ID)
			}
		}
		return nil
	})
	sc.Step(`^a operação lista os pedidos de eliminação$`, func() error {
		m.portal.pedidos, m.err = m.uc().listarPedidos.Executar(m.ctx)
		return nil
	})
	sc.Step(`^a lista de pedidos é "([^"]*)"$`, m.listaDePedidos)

	// ---- domínio: estados do pedido ----
	sc.Step(`^que existe um pedido de eliminação "([^"]*)"$`, func(s string) error {
		p, err := entity.NovoPedidoEliminacao(id("pedido"), id("c-001"), m.clock.Agora().Add(-time.Hour))
		if err == nil && entity.StatusPedido(s) == entity.PedidoAtendido {
			err = p.Atender(m.clock.Agora().Add(-time.Minute))
		}
		m.portal.pedido = p
		return err
	})
	sc.Step(`^o pedido é atendido$`, func() error { m.err = m.portal.pedido.Atender(m.clock.Agora()); return nil })
	sc.Step(`^o pedido está no estado "([^"]*)" com a data de atendimento "([^"]*)"$`, func(s, data string) error {
		if err := m.semErro(); err != nil {
			return err
		}
		if m.portal.pedido.Status != entity.StatusPedido(s) {
			return fmt.Errorf("estado %s, esperado %s", m.portal.pedido.Status, s)
		}
		return m.confereData("atendimento", m.portal.pedido.DataAtendimento, data)
	})
}

func (m *Mundo) portalNoAr(base string) error {
	m.portal.urlBase = base
	return nil
}

func (m *Mundo) pediuLinks(n, email string) error {
	vezes := 1
	if n != "" {
		v, err := strconv.Atoi(n)
		if err != nil {
			return err
		}
		vezes = v
	}
	for i := 0; i < vezes; i++ {
		if err := m.uc().solicitarAcesso.Executar(m.ctx, email); err != nil {
			return err
		}
	}
	return nil
}

func (m *Mundo) mensagens() []entity.MensagemSimulada {
	msgs, _ := m.caixaSaida.Listar(m.ctx, time.Time{})
	return msgs
}

func (m *Mundo) caixaTem(n int, para string) error {
	msgs := m.mensagens()
	if len(msgs) != n {
		return fmt.Errorf("a caixa de saída tem %d mensagens, esperado %d", len(msgs), n)
	}
	for _, msg := range msgs {
		if para != "" && msg.Para != para {
			return fmt.Errorf("mensagem para %q, esperado %q", msg.Para, para)
		}
	}
	return nil
}

// linkEnviado devolve o link e o token da mensagem mais recente da caixa de saída simulada.
func (m *Mundo) linkEnviado() (string, string, error) {
	msgs := m.mensagens()
	if len(msgs) == 0 {
		return "", "", errors.New("nenhuma mensagem na caixa de saída")
	}
	s := reLink.FindStringSubmatch(msgs[0].Corpo)
	if s == nil {
		return "", "", fmt.Errorf("mensagem sem link: %q", msgs[0].Corpo)
	}
	return s[0], s[1], nil
}

func (m *Mundo) tokenDoLink() (entity.TokenAcessoTitular, error) {
	_, token, err := m.linkEnviado()
	if err != nil {
		return entity.TokenAcessoTitular{}, err
	}
	for _, t := range m.tokens.Todos() {
		if t.Hash == usecase.HashToken(token) {
			return t, nil
		}
	}
	return entity.TokenAcessoTitular{}, errors.New("o link enviado não corresponde a nenhum token guardado")
}

func (m *Mundo) linkValeAte(data string) error {
	t, err := m.tokenDoLink()
	if err != nil {
		return err
	}
	return m.confereData("validade do link", &t.ExpiraEm, data)
}

func (m *Mundo) soHash() error {
	_, token, err := m.linkEnviado()
	if err != nil {
		return err
	}
	if _, err := m.tokenDoLink(); err != nil {
		return err
	}
	for _, t := range m.tokens.Todos() {
		if strings.Contains(t.Hash, token) {
			return errors.New("o token em claro está guardado")
		}
	}
	return nil
}

func (m *Mundo) usarToken(token string) {
	m.portal.clienteID, m.err = m.uc().autenticar.Executar(m.ctx, token)
}

func (m *Mundo) linksEmCadastrosDiferentes() error {
	vistos := map[uuid.UUID]bool{}
	msgs := m.mensagens()
	for _, msg := range msgs {
		s := reLink.FindStringSubmatch(msg.Corpo)
		if s == nil {
			return fmt.Errorf("mensagem sem link: %q", msg.Corpo)
		}
		c, err := m.uc().autenticar.Executar(m.ctx, s[1])
		if err != nil {
			return err
		}
		vistos[c] = true
	}
	if len(vistos) != len(msgs) {
		return fmt.Errorf("%d links entraram em %d cadastros", len(msgs), len(vistos))
	}
	return nil
}

func (m *Mundo) meusDadosTrazem(nome, estado string, movs int) error {
	if err := m.semErro(); err != nil {
		return err
	}
	d := m.portal.meus
	if d.Cliente.Nome != nome || string(d.Consentimento.Status) != estado || len(d.Historico) != movs {
		return fmt.Errorf("meus dados: nome %q, consentimento %s, %d movimentos", d.Cliente.Nome, d.Consentimento.Status, len(d.Historico))
	}
	return nil
}

func (m *Mundo) pedeEliminacao(c string) error {
	m.portal.pedido, m.err = m.uc().pedirEliminacao.Executar(m.ctx, id(c))
	return nil
}

func (m *Mundo) pedidosDoCliente(c string, n int, status entity.StatusPedido) error {
	todos, err := m.pedidos.Listar(m.ctx)
	if err != nil {
		return err
	}
	achados := 0
	for _, p := range todos {
		if p.ClienteID == id(c) && (status == "" || p.Status == status) {
			achados++
		}
	}
	if achados != n {
		return fmt.Errorf("o cliente %s tem %d pedidos %s, esperado %d", c, achados, status, n)
	}
	return nil
}

func (m *Mundo) listaDePedidos(esperado string) error {
	if err := m.semErro(); err != nil {
		return err
	}
	var quer, obtido []string
	for _, item := range strings.Split(esperado, ", ") {
		c, s, _ := strings.Cut(item, " ")
		quer = append(quer, id(c).String()+" "+s)
	}
	for _, l := range m.portal.pedidos {
		obtido = append(obtido, l.Pedido.ClienteID.String()+" "+string(l.Pedido.Status))
	}
	return iguais("pedidos", strings.Join(obtido, ", "), strings.Join(quer, ", "))
}
