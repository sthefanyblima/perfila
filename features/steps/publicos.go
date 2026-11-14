package steps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/cucumber/godog"
	"github.com/google/uuid"

	"perfila/internal/application/usecase"
	"perfila/internal/domain/entity"
)

// Estado dos cenários de públicos (Spec v1.9, RN-17, RN-18).
type estadoPublicos struct {
	log        *bytes.Buffer
	gerado     usecase.PublicoGerado
	exportacao entity.ExportacaoPublico
	arquivo    []byte // o conteúdo da última exportação
	baixado    entity.ReciboEnvio
	hash       string
	agregado   []entity.LinhaAgregada
	csv        []byte
	membro     entity.MembroPublico
}

// logger devolve o log dos casos de uso de públicos, guardado para os cenários de C-09.
func (e *estadoPublicos) logger() *slog.Logger {
	if e.log == nil {
		e.log = &bytes.Buffer{}
	}
	return slog.New(slog.NewTextHandler(e.log, nil))
}

func registrarPassosPublicos(sc *godog.ScenarioContext, m *Mundo) {
	sc.Step(`^que o cliente "([^"]*)" com o e-mail "([^"]*)" está no segmento "([^"]*)" com as finalidades "([^"]*)"$`, m.clienteNoSegmento)
	sc.Step(`^que o cadastro está inativo do cliente "([^"]*)"$`, m.clienteInativo)
	sc.Step(`^que a finalidade "([^"]*)" foi revogada do cliente "([^"]*)"$`, func(f, nome string) error {
		return m.uc().consentimento.Revogar(m.ctx, id(nome), entity.Finalidade(f))
	})
	sc.Step(`^que o perfil do cliente "([^"]*)" passou para o segmento "([^"]*)"$`, func(nome, seg string) error {
		p, err := m.perfis.BuscarPorCliente(m.ctx, id(nome))
		if err != nil {
			return err
		}
		s := id("seg:" + seg)
		p.SegmentoID = &s
		_, err = m.perfis.Salvar(m.ctx, p)
		return err
	})
	sc.Step(`^o público do segmento "([^"]*)" é gerado$`, func(seg string) error {
		m.pub.gerado, m.err = m.uc().gerarPublico.Executar(m.ctx, id("seg:"+seg))
		return nil
	})
	sc.Step(`^entram no público (\d+) clientes? e ficam de fora (\d+)$`, func(n, fora int) error {
		if err := m.semErro(); err != nil {
			return err
		}
		if len(m.pub.gerado.Membros) != n || m.pub.gerado.Fora != fora {
			return fmt.Errorf("público: entram %d e ficam de fora %d; esperado %d e %d", len(m.pub.gerado.Membros), m.pub.gerado.Fora, n, fora)
		}
		return nil
	})
	sc.Step(`^o público (não )?contém o cliente "([^"]*)"$`, func(nao, nome string) error {
		achou := false
		for _, c := range m.pub.gerado.Membros {
			achou = achou || c.ClienteID == id(nome)
		}
		if achou != (nao == "") {
			return fmt.Errorf("cliente %s no público: %v", nome, achou)
		}
		return nil
	})
	sc.Step(`^(?:que )?o operador "([^"]*)" (?:exporta|exportou) o público do segmento "([^"]*)"$`, func(op, seg string) error {
		return m.exportar(op, seg, false)
	})
	sc.Step(`^o operador "([^"]*)" exporta as remoções do segmento "([^"]*)"$`, func(op, seg string) error {
		return m.exportar(op, seg, true)
	})
	sc.Step(`^nenhuma exportação é registrada$`, func() error {
		for _, seg := range []string{"vip", "fiel"} {
			h, err := m.uc().exportarPublico.Historico(m.ctx, id("seg:"+seg))
			if err != nil || len(h) > 0 {
				return fmt.Errorf("exportações do %s: %d (%v)", seg, len(h), err)
			}
		}
		return nil
	})
	sc.Step(`^o arquivo exportado é:$`, func(d *godog.DocString) error {
		if got := strings.TrimSpace(string(m.pub.arquivo)); got != strings.TrimSpace(d.Content) {
			return fmt.Errorf("arquivo:\n%s", got)
		}
		return nil
	})
	sc.Step(`^o arquivo exportado não contém "([^"]*)" nem o nome do cliente "([^"]*)"$`, func(email, nome string) error {
		c, err := m.clientes.BuscarPorID(m.ctx, id(nome))
		if err != nil {
			return err
		}
		for _, s := range []string{email, c.Nome, id(nome).String()} {
			if bytes.Contains(m.pub.arquivo, []byte(s)) {
				return fmt.Errorf("o arquivo contém %q", s)
			}
		}
		return nil
	})
	sc.Step(`^o arquivo exportado tem o hash do e-mail "([^"]*)"$`, func(email string) error {
		if !bytes.Contains(m.pub.arquivo, []byte(entity.HashEmail(email))) {
			return fmt.Errorf("o arquivo não tem o hash de %s:\n%s", email, m.pub.arquivo)
		}
		return nil
	})
	sc.Step(`^o hash do e-mail "([^"]*)" é calculado$`, func(e string) error { m.pub.hash = entity.HashEmail(e); return nil })
	sc.Step(`^o hash é "([^"]*)"$`, func(h string) error {
		if m.pub.hash != h {
			return fmt.Errorf("hash %s", m.pub.hash)
		}
		return nil
	})
	sc.Step(`^a exportação registra o operador "([^"]*)", o segmento "([^"]*)", o tipo "([^"]*)", (\d+) clientes? e o hash do arquivo$`,
		func(op, seg, tipo string, n int) error {
			e := m.pub.exportacao
			if e.Operador != op || e.SegmentoNome != seg || string(e.Tipo) != tipo || e.Quantidade != n || len(e.HashArquivo) != 64 {
				return fmt.Errorf("exportação registrada: %+v", e)
			}
			h, err := m.uc().exportarPublico.Historico(m.ctx, id("seg:"+seg))
			if err != nil || len(h) != 1 || h[0] != e {
				return fmt.Errorf("histórico: %+v %v", h, err)
			}
			return nil
		})
	sc.Step(`^o log não contém o e-mail, o hash do e-mail nem o nome do cliente "([^"]*)"$`, func(nome string) error {
		c, err := m.clientes.BuscarPorID(m.ctx, id(nome))
		if err != nil {
			return err
		}
		log := m.pub.log.String()
		if !strings.Contains(log, "publico_exportado") {
			return fmt.Errorf("a exportação não foi registrada no log")
		}
		for _, s := range []string{c.Email, entity.HashEmail(c.Email), c.Nome} {
			if strings.Contains(log, s) {
				return fmt.Errorf("o log contém %q", s)
			}
		}
		return nil
	})
	sc.Step(`^o arquivo da última exportação é baixado$`, func() error {
		m.pub.baixado, m.err = m.uc().exportarPublico.Baixar(m.ctx, m.pub.exportacao.ID)
		return nil
	})
	sc.Step(`^o arquivo baixado é igual ao exportado$`, func() error {
		if err := m.semErro(); err != nil {
			return err
		}
		if !bytes.Equal(m.pub.baixado.Conteudo, m.pub.arquivo) || m.pub.baixado.Hash != m.pub.exportacao.HashArquivo {
			return fmt.Errorf("arquivo baixado diferente do exportado")
		}
		return nil
	})
	sc.Step(`^o segmento "([^"]*)" tem (\d+) remoç(?:ão|ões) pendentes?$`, func(seg string, n int) error {
		p, err := m.publicos.ListarMembrosPorSegmento(m.ctx, id("seg:"+seg), entity.MembroRemocaoPendente)
		if err != nil || len(p) != n {
			return fmt.Errorf("remoções pendentes: %d (%v), esperado %d", len(p), err, n)
		}
		return nil
	})
	sc.Step(`^o membro do cliente "([^"]*)" no segmento "([^"]*)" está no estado "([^"]*)"$`, func(nome, seg, estado string) error {
		ms, err := m.publicos.ListarMembrosPorCliente(m.ctx, id(nome), entity.StatusMembro(estado))
		if err != nil {
			return err
		}
		for _, x := range ms {
			if x.SegmentoID == id("seg:"+seg) {
				return nil
			}
		}
		return fmt.Errorf("o cliente %s não tem membro %s no segmento %s", nome, estado, seg)
	})
	sc.Step(`^que existe um membro de público no estado "([^"]*)"$`, func(estado string) error {
		m.pub.membro = entity.MembroPublico{ID: uuid.New(), Status: entity.StatusMembro(estado)}
		return nil
	})
	sc.Step(`^tenta-se "([^"]*)" o membro$`, func(acao string) error {
		switch acao {
		case "marcar remoção":
			m.err = m.pub.membro.MarcarRemocao(m.clock.Agora())
		case "remover":
			m.err = m.pub.membro.Remover(uuid.New(), m.clock.Agora())
		default:
			return fmt.Errorf("ação desconhecida: %q", acao)
		}
		return nil
	})
	sc.Step(`^que (\d+) clientes estão no segmento "([^"]*)" com ticket médio (\d+) e a categoria preferida "([^"]*)"$`, m.clientesNoSegmento)
	sc.Step(`^que 1 dos clientes do segmento "([^"]*)" revogou a segmentação$`, func(seg string) error {
		return m.uc().consentimento.Revogar(m.ctx, id(seg+"-1"), entity.FinalidadeSegmentacao)
	})
	sc.Step(`^o relatório agregado por segmento é gerado$`, func() error {
		m.pub.agregado, m.err = m.uc().relatorio.Executar(m.ctx)
		return m.err
	})
	sc.Step(`^o relatório agregado por segmento é exportado em CSV$`, func() error {
		linhas, err := m.uc().relatorio.Executar(m.ctx)
		m.pub.csv = m.uc().relatorio.CSV(linhas)
		return err
	})
	sc.Step(`^a linha do segmento "([^"]*)" mostra (\d+) clientes, ticket médio (\d+) e a categoria "([^"]*)"$`,
		func(seg string, n, ticket int, cat string) error {
			l, err := m.linhaAgregada(seg)
			if err != nil {
				return err
			}
			if l.Oculta || l.Quantidade != n || !quase(l.TicketMedio, float64(ticket)) || l.CategoriaMaisComum != cat {
				return fmt.Errorf("linha do %s: %+v", seg, l)
			}
			return nil
		})
	sc.Step(`^a linha do segmento "([^"]*)" mostra "menos de 5" sem ticket nem categoria$`, func(seg string) error {
		l, err := m.linhaAgregada(seg)
		if err != nil {
			return err
		}
		if !l.Oculta || l.Quantidade != 0 || l.TicketMedio != 0 || l.CategoriaMaisComum != "" {
			return fmt.Errorf("linha do %s deveria estar oculta: %+v", seg, l)
		}
		return nil
	})
	sc.Step(`^o CSV do agregado é:$`, func(d *godog.DocString) error {
		if got := strings.TrimSpace(string(m.pub.csv)); got != strings.TrimSpace(d.Content) {
			return fmt.Errorf("CSV:\n%s", got)
		}
		return nil
	})
	sc.Step(`^o corpo lista a finalidade "([^"]*)" no estado "([^"]*)" e "([^"]*)" no estado "([^"]*)"$`, m.corpoListaFinalidades)
}

// clienteNoSegmento cria o cliente ativo com perfil no segmento e as finalidades concedidas.
func (m *Mundo) clienteNoSegmento(nome, email, seg, finalidades string) error {
	if err := m.clientes.Salvar(m.ctx, entity.Cliente{ID: id(nome), Nome: "Cliente " + nome, Email: email,
		DataCadastro: m.clock.Agora(), Ativo: true}); err != nil {
		return err
	}
	s := id("seg:" + seg)
	if _, err := m.perfis.Salvar(m.ctx, entity.PerfilCliente{ID: uuid.New(), ClienteID: id(nome), FrequenciaCompras: 3,
		TicketMedio: 600, CategoriaPreferida: "livros", SegmentoID: &s}); err != nil {
		return err
	}
	for _, f := range strings.Split(finalidades, ",") {
		if err := m.consentimentoNoEstadoPara(nome, strings.TrimSpace(f), string(entity.StatusConcedido)); err != nil {
			return err
		}
	}
	return nil
}

// clientesNoSegmento cria n clientes "<seg>-1".."<seg>-n" com segmentacao concedida.
func (m *Mundo) clientesNoSegmento(n int, seg string, ticket int, cat string) error {
	s := id("seg:" + seg)
	for i := 1; i <= n; i++ {
		nome := fmt.Sprintf("%s-%d", seg, i)
		if err := m.clientes.Salvar(m.ctx, entity.Cliente{ID: id(nome), Nome: "Cliente " + nome, Email: nome + "@exemplo.test",
			DataCadastro: m.clock.Agora(), Ativo: true}); err != nil {
			return err
		}
		if _, err := m.perfis.Salvar(m.ctx, entity.PerfilCliente{ID: uuid.New(), ClienteID: id(nome), FrequenciaCompras: 3,
			TicketMedio: float64(ticket), CategoriaPreferida: cat, SegmentoID: &s}); err != nil {
			return err
		}
		if err := m.consentimentoNoEstadoPara(nome, string(entity.FinalidadeSegmentacao), string(entity.StatusConcedido)); err != nil {
			return err
		}
	}
	return nil
}

func (m *Mundo) exportar(operador, seg string, remocoes bool) error {
	ctx := usecase.ComOrigem(m.ctx, entity.OrigemPainel, operador)
	uc := m.uc().exportarPublico
	if remocoes {
		m.pub.exportacao, m.err = uc.ExportarRemocoes(ctx, id("seg:"+seg))
	} else {
		m.pub.exportacao, m.err = uc.Exportar(ctx, id("seg:"+seg))
	}
	m.pub.arquivo = nil
	if m.err == nil {
		r, err := uc.Baixar(m.ctx, m.pub.exportacao.ID)
		if err != nil {
			return err
		}
		m.pub.arquivo = r.Conteudo
	}
	return nil
}

func (m *Mundo) linhaAgregada(seg string) (entity.LinhaAgregada, error) {
	for _, l := range m.pub.agregado {
		if l.SegmentoNome == seg {
			return l, nil
		}
	}
	return entity.LinhaAgregada{}, fmt.Errorf("o relatório não tem o segmento %s", seg)
}

func (m *Mundo) corpoListaFinalidades(f1, e1, f2, e2 string) error {
	corpo, _ := io.ReadAll(bytes.NewReader(m.api.resposta.Body.Bytes()))
	var r struct {
		Finalidades []struct {
			Finalidade string `json:"finalidade"`
			Status     string `json:"status"`
		} `json:"finalidades"`
	}
	if err := json.Unmarshal(corpo, &r); err != nil {
		return err
	}
	if len(r.Finalidades) != 2 || r.Finalidades[0].Finalidade != f1 || r.Finalidades[0].Status != e1 ||
		r.Finalidades[1].Finalidade != f2 || r.Finalidades[1].Status != e2 {
		return fmt.Errorf("finalidades no corpo: %s", corpo)
	}
	return nil
}
