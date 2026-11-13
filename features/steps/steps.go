package steps

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/cucumber/godog"
	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/shared/errs"
)

const num = `(-?\d+(?:\.\d+)?)`

// InitializeScenario registra cada passo reutilizável uma única vez (B-12).
func InitializeScenario(sc *godog.ScenarioContext) {
	m := novoMundo()
	registrarPassosAPI(sc, m)
	registrarPassosProcessamento(sc, m)
	registrarPassosLeitura(sc, m)
	registrarPassosHistorico(sc, m)
	registrarPassosHistoricoWeb(sc, m)
	registrarPassosWeb(sc, m)
	registrarPassosImplantacao(sc)
	registrarPassosConfiguracao(sc, m)
	registrarPassosOperadores(sc, m)
	registrarPassosSegmentos(sc, m)
	registrarPassosSegmentosWeb(sc, m)

	// ---- comuns ----
	sc.Step(`^que a data atual é "([^"]*)"$`, m.dataAtual)
	sc.Step(`^que existe o cliente "([^"]*)" ativo$`, m.clienteAtivo)
	sc.Step(`^que o cliente "([^"]*)" está inativo$`, m.clienteInativo)
	sc.Step(`^que o consentimento do cliente "([^"]*)" está no estado "([^"]*)"$`, m.consentimentoNoEstado)
	sc.Step(`^a operação falha com o erro "([^"]*)"$`, m.falhaComErro)

	// ---- consentimento (UC-08) ----
	sc.Step(`^o cliente "([^"]*)" concede consentimento para a finalidade "([^"]*)" com o termo "([^"]*)"$`, m.concede)
	sc.Step(`^o cliente "([^"]*)" revoga o consentimento$`, m.revoga)
	sc.Step(`^o cliente "([^"]*)" tenta "([^"]*)" o consentimento$`, m.tentaConsentimento)
	sc.Step(`^o consentimento do cliente "([^"]*)" é consultado$`, m.consulta)
	sc.Step(`^o consentimento do cliente "([^"]*)" (?:está|permanece) no estado "([^"]*)"$`, m.confereEstadoConsentimento)
	sc.Step(`^o consentimento do cliente "([^"]*)" é válido$`, func(c string) error { return m.confereValidade(c, true) })
	sc.Step(`^o consentimento do cliente "([^"]*)" não é válido$`, func(c string) error { return m.confereValidade(c, false) })
	sc.Step(`^a validade retornada é "(verdadeiro|falso)"$`, m.validadeRetornada)
	sc.Step(`^a data de consentimento é "([^"]*)"$`, m.dataConsentimento)
	sc.Step(`^a data de revogação é "([^"]*)"$`, m.dataRevogacao)

	// ---- cadastro (UC-09) ----
	sc.Step(`^um cliente é cadastrado com nome "([^"]*)" e e-mail "([^"]*)"$`, m.cadastra)
	sc.Step(`^que um cliente foi cadastrado com nome "([^"]*)" e e-mail "([^"]*)"$`, func(n, e string) error {
		if err := m.cadastra(n, e); err != nil {
			return err
		}
		return m.semErro()
	})
	sc.Step(`^o cliente cadastrado está ativo$`, m.cadastradoAtivo)
	sc.Step(`^a data de cadastro é "([^"]*)"$`, func(s string) error {
		t := m.clienteCadastrado.DataCadastro
		return m.confereData("cadastro", &t, s)
	})
	sc.Step(`^o consentimento do cliente cadastrado está no estado "([^"]*)"$`, m.estadoConsentimentoCadastrado)
	sc.Step(`^que o cliente cadastrado concedeu consentimento para a finalidade "([^"]*)" com o termo "([^"]*)"$`, func(f, t string) error {
		return m.uc().consentimento.Conceder(m.ctx, m.clienteCadastrado.ID, f, t)
	})
	sc.Step(`^o perfil do cliente cadastrado é gerado$`, func() error {
		m.perfil, m.err = m.uc().gerarPerfil.Executar(m.ctx, m.clienteCadastrado.ID)
		return nil
	})

	// ---- eventos (UC-01) ----
	sc.Step(`^o sistema registra para o cliente "([^"]*)" um evento "([^"]*)" de origem "([^"]*)" com valor `+num+` em "([^"]*)"$`, m.registraEvento)
	sc.Step(`^o sistema tenta "([^"]*)" eventos do cliente "([^"]*)"$`, m.tentaEventos)
	sc.Step(`^o sistema coleta os eventos do cliente "([^"]*)"$`, m.coleta)
	sc.Step(`^que existem (\d+) eventos do cliente "([^"]*)" e (\d+) eventos do cliente "([^"]*)"$`, m.existemEventos)
	sc.Step(`^o evento é gravado$`, m.eventoGravado)
	sc.Step(`^nenhum evento é gravado$`, m.nenhumEvento)
	sc.Step(`^a coleta do cliente "([^"]*)" retorna (\d+) eventos?$`, m.coletaDoClienteRetorna)
	sc.Step(`^a coleta retorna (\d+) eventos?$`, m.coletaRetorna)

	// ---- perfil (UC-02) ----
	sc.Step(`^que a meta de engajamento é de (\d+) eventos$`, func(n int) error { m.metaEventos = n; return nil })
	sc.Step(`^que o cliente "([^"]*)" possui os eventos:$`, m.possuiEventos)
	sc.Step(`^que o cliente "([^"]*)" possui um evento "([^"]*)" ocorrido há (\d+) dias$`, m.eventoHaDias)
	sc.Step(`^que o cliente "([^"]*)" possui (\d+) eventos "([^"]*)" nos últimos (\d+) dias$`, m.eventosRecentes)
	sc.Step(`^que o cliente "([^"]*)" não possui eventos$`, func(string) error { return nil })
	sc.Step(`^que o cliente "([^"]*)" já possui um perfil com frequência de compras (\d+)$`, m.perfilExistente)
	sc.Step(`^o perfil do cliente "([^"]*)" é gerado$`, m.geraPerfil)
	sc.Step(`^a frequência de compras é (\d+)$`, m.frequencia)
	sc.Step(`^o ticket médio é `+num+`$`, m.ticket)
	sc.Step(`^a categoria preferida é "([^"]*)"$`, m.categoria)
	sc.Step(`^a categoria preferida é vazia$`, func() error { return m.categoria("") })
	sc.Step(`^o score de engajamento é `+num+`$`, m.score)
	sc.Step(`^existe um único perfil para o cliente "([^"]*)"$`, m.perfilUnico)
	sc.Step(`^o identificador do perfil é preservado$`, m.idPreservado)
	sc.Step(`^nenhum perfil é gravado para o cliente "([^"]*)"$`, m.nenhumPerfil)

	// ---- segmentação (UC-03) ----
	sc.Step(`^que existem os segmentos ativos:$`, m.segmentosAtivos)
	sc.Step(`^que o perfil do cliente "([^"]*)" tem ticket médio `+num+`, frequência de compras (\d+) e score de engajamento `+num+`$`, m.perfilComMetricas)
	sc.Step(`^que o cliente "([^"]*)" não possui perfil$`, func(string) error { return nil })
	sc.Step(`^o cliente "([^"]*)" é segmentado$`, m.segmenta)
	sc.Step(`^o segmento atribuído é "([^"]*)"$`, m.segmentoAtribuido)
	sc.Step(`^o perfil do cliente "([^"]*)" está associado ao segmento "([^"]*)"$`, m.perfilAssociado)
	sc.Step(`^o perfil do cliente "([^"]*)" continua sem segmento$`, m.perfilSemSegmento)
	sc.Step(`^o critério "([^"]*)" é avaliado$`, m.avaliaCriterio)
	sc.Step(`^a avaliação do critério não falha$`, m.semErro)

	// ---- recomendação (UC-04) ----
	sc.Step(`^que o perfil do cliente "([^"]*)" tem categoria preferida "([^"]*)" e segmento "([^"]*)"$`, m.perfilComCategoria)
	sc.Step(`^que o perfil do cliente "([^"]*)" tem categoria preferida vazia e segmento "([^"]*)"$`,
		func(c, s string) error { return m.perfilComCategoria(c, "", s) })
	sc.Step(`^que o perfil do cliente "([^"]*)" ainda não foi segmentado$`, m.perfilNaoSegmentado)
	sc.Step(`^que o catálogo possui os produtos:$`, m.catalogo)
	sc.Step(`^que todos os produtos da categoria "([^"]*)" estão inativos$`, m.inativaCategoria)
	sc.Step(`^a recomendação do cliente "([^"]*)" é gerada$`, m.geraRecomendacao)
	sc.Step(`^o produto recomendado é "([^"]*)"$`, m.produtoRecomendado)
	sc.Step(`^o motivo é "([^"]*)"$`, m.motivo)
	sc.Step(`^nenhuma recomendação é gravada$`, func() error { return m.gravacoes(0) })
	sc.Step(`^a recomendação é gravada (\d+) vez(?:es)?$`, m.gravacoes)
	sc.Step(`^a recomendação gravada está no estado "([^"]*)"$`, m.recomendacaoGravadaNoEstado)

	// ---- ciclo da recomendação (estados) ----
	sc.Step(`^que a validade padrão é de (\d+) dias$`, func(n int) error { m.validadeDias = n; return nil })
	sc.Step(`^que existe uma recomendação no estado "([^"]*)"$`, m.recomendacaoNoEstado)
	sc.Step(`^que existe uma recomendação ativada em "([^"]*)" com validade de (\d+) dias$`, m.recomendacaoAtivada)
	sc.Step(`^a recomendação é disponibilizada$`, func() error { return m.acaoRecomendacao("disponibilizar") })
	sc.Step(`^a recomendação é expirada$`, func() error { return m.acaoRecomendacao("expirar") })
	sc.Step(`^tenta-se "([^"]*)" a recomendação$`, m.acaoRecomendacao)
	sc.Step(`^passam (\d+) dias e o comando "expirar-recomendacoes" é executado$`, m.passamDiasEExpira)
	sc.Step(`^a recomendação está no estado "([^"]*)"$`, m.recomendacaoEstado)
	sc.Step(`^a validade é "([^"]*)"$`, m.validade)
	sc.Step(`^a data de ativação é "([^"]*)"$`, func(s string) error { return m.confereData("ativação", m.recomendacao.DataAtivacao, s) })
	sc.Step(`^a data de expiração é "([^"]*)"$`, func(s string) error { return m.confereData("expiração", m.recomendacao.DataExpiracao, s) })
}

// ======================= comuns =======================

func (m *Mundo) dataAtual(s string) error {
	t, err := instante(s)
	if err != nil {
		return err
	}
	m.clock.Definir(t)
	return nil
}

func (m *Mundo) clienteAtivo(nome string) error {
	return m.clientes.Salvar(m.ctx, entity.Cliente{
		ID: id(nome), Nome: "Cliente " + nome, Email: nome + "@exemplo.test",
		DataCadastro: m.clock.Agora(), Ativo: true,
	})
}

func (m *Mundo) clienteInativo(nome string) error {
	c, err := m.clientes.BuscarPorID(m.ctx, id(nome))
	if err != nil {
		return err
	}
	c.Desativar()
	return m.clientes.Salvar(m.ctx, c)
}

// consentimentoNoEstado monta o estado inicial usando as próprias transições do domínio.
func (m *Mundo) consentimentoNoEstado(nome, estado string) error {
	cid := id(nome)
	agora := m.clock.Agora()
	switch entity.StatusConsentimento(estado) {
	case entity.StatusNaoConcedido:
		m.consentimentos.Remover(cid)
		return nil
	case entity.StatusConcedido, entity.StatusRevogado:
		c := entity.NovoConsentimento(uuid.New(), cid, agora)
		if err := c.Conceder("segmentacao", "v1", agora); err != nil {
			return err
		}
		if estado == string(entity.StatusRevogado) {
			if err := c.Revogar(agora); err != nil {
				return err
			}
		}
		return m.consentimentos.Salvar(m.ctx, c)
	}
	return fmt.Errorf("estado de consentimento desconhecido: %q", estado)
}

func (m *Mundo) falhaComErro(codigo string) error {
	if m.err == nil {
		return fmt.Errorf("esperava erro %q, mas a operação teve sucesso", codigo)
	}
	if got := errs.Codigo(m.err); got != codigo {
		return fmt.Errorf("esperava erro %q, obtido %q (%v)", codigo, got, m.err)
	}
	return nil
}

func (m *Mundo) semErro() error {
	if m.err != nil {
		return fmt.Errorf("erro inesperado: %v", m.err)
	}
	return nil
}

// ======================= consentimento =======================

func (m *Mundo) concede(nome, finalidade, termo string) error {
	m.err = m.uc().consentimento.Conceder(m.ctx, id(nome), finalidade, termo)
	return nil
}

func (m *Mundo) revoga(nome string) error {
	m.err = m.uc().consentimento.Revogar(m.ctx, id(nome))
	return nil
}

func (m *Mundo) tentaConsentimento(nome, acao string) error {
	switch acao {
	case "conceder":
		return m.concede(nome, "segmentacao", "v1")
	case "revogar":
		return m.revoga(nome)
	}
	return fmt.Errorf("ação desconhecida: %q", acao)
}

func (m *Mundo) consulta(nome string) error {
	m.consentimento, m.err = m.uc().consentimento.Consultar(m.ctx, id(nome))
	return nil
}

func (m *Mundo) recarregaConsentimento(nome string) error {
	c, err := m.uc().consentimento.Consultar(m.ctx, id(nome))
	if err != nil {
		return err
	}
	m.consentimento = c
	return nil
}

func (m *Mundo) confereEstadoConsentimento(nome, estado string) error {
	if err := m.recarregaConsentimento(nome); err != nil {
		return err
	}
	if string(m.consentimento.Status) != estado {
		return fmt.Errorf("estado: esperado %s, obtido %s", estado, m.consentimento.Status)
	}
	return nil
}

func (m *Mundo) confereValidade(nome string, quer bool) error {
	if err := m.recarregaConsentimento(nome); err != nil {
		return err
	}
	if m.consentimento.EstaValido() != quer {
		return fmt.Errorf("validade: esperado %v, obtido %v", quer, m.consentimento.EstaValido())
	}
	return nil
}

func (m *Mundo) validadeRetornada(v string) error {
	if m.err != nil {
		return m.err
	}
	if quer := v == "verdadeiro"; m.consentimento.EstaValido() != quer {
		return fmt.Errorf("validade: esperado %v, obtido %v", quer, !quer)
	}
	return nil
}

func (m *Mundo) dataConsentimento(s string) error {
	t := m.consentimento.DataConsentimento
	return m.confereData("consentimento", &t, s)
}

func (m *Mundo) dataRevogacao(s string) error {
	return m.confereData("revogação", m.consentimento.DataRevogacao, s)
}

func (m *Mundo) confereData(campo string, got *time.Time, s string) error {
	quer, err := instante(s)
	if err != nil {
		return err
	}
	if got == nil || !got.Equal(quer) {
		return fmt.Errorf("data de %s: esperado %s, obtido %v", campo, quer, got)
	}
	return nil
}

// ======================= eventos =======================

func (m *Mundo) registraEvento(nome, tipo, origem string, valor float64, quando string) error {
	t, err := instante(quando)
	if err != nil {
		return err
	}
	_, m.err = m.uc().coletar.Registrar(m.ctx, entity.EventoComportamental{
		ClienteID: id(nome), Tipo: entity.TipoEvento(tipo), Origem: origem, Timestamp: t, Valor: valor,
	})
	return nil
}

func (m *Mundo) tentaEventos(acao, nome string) error {
	switch acao {
	case "registrar":
		return m.registraEvento(nome, "navegacao", "web", 0, "2026-03-09T10:00:00Z")
	case "coletar":
		return m.coleta(nome)
	}
	return fmt.Errorf("ação desconhecida: %q", acao)
}

func (m *Mundo) coleta(nome string) error {
	m.eventosColeta, m.err = m.uc().coletar.Executar(m.ctx, id(nome))
	return nil
}

func (m *Mundo) gravaEvento(nome string, tipo entity.TipoEvento, valor float64, categoria string, t time.Time) error {
	return m.eventos.Salvar(m.ctx, entity.EventoComportamental{
		ID: uuid.New(), ClienteID: id(nome), Tipo: tipo, Origem: "ecommerce",
		Timestamp: t, Valor: valor, Categoria: categoria,
	})
}

func (m *Mundo) existemEventos(n1 int, c1 string, n2 int, c2 string) error {
	t := m.clock.Agora().Add(-time.Hour)
	for i := 0; i < n1; i++ {
		if err := m.gravaEvento(c1, entity.TipoNavegacao, 0, "", t); err != nil {
			return err
		}
	}
	for i := 0; i < n2; i++ {
		if err := m.gravaEvento(c2, entity.TipoNavegacao, 0, "", t); err != nil {
			return err
		}
	}
	return nil
}

func (m *Mundo) eventoGravado() error {
	if m.err != nil {
		return fmt.Errorf("evento não gravado: %v", m.err)
	}
	if m.eventos.Total() == 0 {
		return errors.New("nenhum evento no repositório")
	}
	return nil
}

func (m *Mundo) nenhumEvento() error {
	if n := m.eventos.Total(); n != 0 {
		return fmt.Errorf("esperava 0 eventos gravados, há %d", n)
	}
	return nil
}

func (m *Mundo) coletaDoClienteRetorna(nome string, n int) error {
	evs, err := m.uc().coletar.Executar(m.ctx, id(nome))
	if err != nil {
		return err
	}
	if len(evs) != n {
		return fmt.Errorf("coleta: esperado %d eventos, obtido %d", n, len(evs))
	}
	return nil
}

func (m *Mundo) coletaRetorna(n int) error {
	if m.err != nil {
		return m.err
	}
	if len(m.eventosColeta) != n {
		return fmt.Errorf("coleta: esperado %d eventos, obtido %d", n, len(m.eventosColeta))
	}
	return nil
}

// ======================= perfil =======================

func (m *Mundo) possuiEventos(nome string, tabela *godog.Table) error {
	for _, linha := range tabela.Rows[1:] {
		valor, err := strconv.ParseFloat(linha.Cells[1].Value, 64)
		if err != nil {
			return err
		}
		t, err := instante(linha.Cells[3].Value)
		if err != nil {
			return err
		}
		if err := m.gravaEvento(nome, entity.TipoEvento(linha.Cells[0].Value), valor, linha.Cells[2].Value, t); err != nil {
			return err
		}
	}
	return nil
}

func (m *Mundo) eventoHaDias(nome, tipo string, dias int) error {
	t := m.clock.Agora().AddDate(0, 0, -dias)
	return m.gravaEvento(nome, entity.TipoEvento(tipo), 0, "", t)
}

func (m *Mundo) eventosRecentes(nome string, n int, tipo string, dias int) error {
	for i := 0; i < n; i++ {
		t := m.clock.Agora().Add(-time.Duration(i%dias+1) * 12 * time.Hour)
		if err := m.gravaEvento(nome, entity.TipoEvento(tipo), 0, "", t); err != nil {
			return err
		}
	}
	return nil
}

func (m *Mundo) perfilExistente(nome string, freq int) error {
	p, err := m.perfis.Salvar(m.ctx, entity.PerfilCliente{ID: uuid.New(), ClienteID: id(nome), FrequenciaCompras: freq})
	m.perfilIDOrigem = p.ID
	return err
}

func (m *Mundo) geraPerfil(nome string) error {
	m.perfil, m.err = m.uc().gerarPerfil.Executar(m.ctx, id(nome))
	return nil
}

func (m *Mundo) frequencia(n int) error {
	if err := m.semErro(); err != nil {
		return err
	}
	if m.perfil.FrequenciaCompras != n {
		return fmt.Errorf("frequência: esperado %d, obtido %d", n, m.perfil.FrequenciaCompras)
	}
	return nil
}

func (m *Mundo) ticket(v float64) error {
	if !quase(m.perfil.TicketMedio, v) {
		return fmt.Errorf("ticket médio: esperado %v, obtido %v", v, m.perfil.TicketMedio)
	}
	return m.semErro()
}

func (m *Mundo) categoria(c string) error {
	if m.perfil.CategoriaPreferida != c {
		return fmt.Errorf("categoria preferida: esperado %q, obtido %q", c, m.perfil.CategoriaPreferida)
	}
	return m.semErro()
}

func (m *Mundo) score(v float64) error {
	if !quase(m.perfil.ScoreEngajamento, v) {
		return fmt.Errorf("score: esperado %v, obtido %v", v, m.perfil.ScoreEngajamento)
	}
	return m.semErro()
}

func (m *Mundo) perfilUnico(nome string) error {
	if _, err := m.perfis.BuscarPorCliente(m.ctx, id(nome)); err != nil {
		return err
	}
	if n := m.perfis.Total(); n != 1 {
		return fmt.Errorf("esperava 1 perfil, há %d", n)
	}
	return nil
}

func (m *Mundo) idPreservado() error {
	if m.perfil.ID != m.perfilIDOrigem {
		return fmt.Errorf("ID do perfil mudou: %s → %s", m.perfilIDOrigem, m.perfil.ID)
	}
	return nil
}

func (m *Mundo) nenhumPerfil(nome string) error {
	if _, err := m.perfis.BuscarPorCliente(m.ctx, id(nome)); !errors.Is(err, errs.ErrPerfilNaoEncontrado) {
		return fmt.Errorf("esperava nenhum perfil para %s", nome)
	}
	return nil
}

// ======================= segmentação =======================

func (m *Mundo) segmentosAtivos(tabela *godog.Table) error {
	for _, linha := range tabela.Rows[1:] {
		prio, err := strconv.Atoi(linha.Cells[0].Value)
		if err != nil {
			return err
		}
		nome := linha.Cells[1].Value
		m.segmentos.Adicionar(entity.Segmento{
			ID: id("seg:" + nome), Nome: nome, Criterio: linha.Cells[2].Value, Prioridade: prio, Ativo: true,
		})
	}
	return nil
}

func (m *Mundo) perfilComMetricas(nome string, ticket float64, freq int, score float64) error {
	_, err := m.perfis.Salvar(m.ctx, entity.PerfilCliente{
		ID: uuid.New(), ClienteID: id(nome), TicketMedio: ticket, FrequenciaCompras: freq, ScoreEngajamento: score,
	})
	return err
}

func (m *Mundo) segmenta(nome string) error {
	m.segmento, m.err = m.uc().segmentar.Executar(m.ctx, id(nome))
	return nil
}

func (m *Mundo) segmentoAtribuido(nome string) error {
	if err := m.semErro(); err != nil {
		return err
	}
	if m.segmento.Nome != nome {
		return fmt.Errorf("segmento: esperado %q, obtido %q", nome, m.segmento.Nome)
	}
	return nil
}

func (m *Mundo) perfilAssociado(cliente, segmento string) error {
	p, err := m.perfis.BuscarPorCliente(m.ctx, id(cliente))
	if err != nil {
		return err
	}
	if p.SegmentoID == nil || *p.SegmentoID != id("seg:"+segmento) {
		return fmt.Errorf("perfil não associado ao segmento %q", segmento)
	}
	return nil
}

func (m *Mundo) perfilSemSegmento(cliente string) error {
	p, err := m.perfis.BuscarPorCliente(m.ctx, id(cliente))
	if err != nil {
		return err
	}
	if p.SegmentoID != nil {
		return errors.New("perfil recebeu segmento indevidamente")
	}
	return nil
}

func (m *Mundo) avaliaCriterio(criterio string) error {
	_, m.err = entity.Segmento{Nome: "teste", Criterio: criterio}.ClassificarPerfil(entity.PerfilCliente{})
	return nil
}

// ======================= recomendação =======================

func (m *Mundo) garanteSegmento(nome string) uuid.UUID {
	sid := id("seg:" + nome)
	if _, err := m.segmentos.BuscarPorID(m.ctx, sid); err != nil {
		m.segmentos.Adicionar(entity.Segmento{ID: sid, Nome: nome, Prioridade: 1, Ativo: true})
	}
	return sid
}

func (m *Mundo) perfilComCategoria(cliente, categoria, segmento string) error {
	sid := m.garanteSegmento(segmento)
	_, err := m.perfis.Salvar(m.ctx, entity.PerfilCliente{
		ID: uuid.New(), ClienteID: id(cliente), CategoriaPreferida: categoria, SegmentoID: &sid,
	})
	return err
}

func (m *Mundo) perfilNaoSegmentado(cliente string) error {
	p, err := m.perfis.BuscarPorCliente(m.ctx, id(cliente))
	if err != nil {
		return err
	}
	p.SegmentoID = nil
	_, err = m.perfis.Salvar(m.ctx, p)
	return err
}

func (m *Mundo) catalogo(tabela *godog.Table) error {
	for _, linha := range tabela.Rows[1:] {
		preco, err := strconv.ParseFloat(linha.Cells[2].Value, 64)
		if err != nil {
			return err
		}
		nome := linha.Cells[0].Value
		m.produtos.Adicionar(entity.Produto{
			ID: id("prod:" + nome), Nome: nome, Categoria: linha.Cells[1].Value,
			Preco: preco, Ativo: linha.Cells[3].Value == "sim",
		})
	}
	return nil
}

func (m *Mundo) inativaCategoria(categoria string) error {
	for _, p := range m.produtos.Todos() {
		if p.Categoria == categoria {
			p.Ativo = false
			m.produtos.Adicionar(p)
		}
	}
	return nil
}

func (m *Mundo) geraRecomendacao(cliente string) error {
	m.recomendacao, m.err = m.uc().recomendar.Executar(m.ctx, id(cliente))
	m.temRecomendacao = m.err == nil
	return nil
}

func (m *Mundo) produtoRecomendado(nome string) error {
	if err := m.semErro(); err != nil {
		return err
	}
	// Compara pelo nome do produto gravado: os criados pelo painel (v1.6) têm ID novo.
	p, err := m.produtos.BuscarPorID(m.ctx, m.recomendacao.ProdutoID)
	if err != nil || p.Nome != nome {
		return fmt.Errorf("produto recomendado %q diferente de %q", p.Nome, nome)
	}
	return nil
}

func (m *Mundo) motivo(s string) error {
	if m.recomendacao.Motivo != s {
		return fmt.Errorf("motivo: esperado %q, obtido %q", s, m.recomendacao.Motivo)
	}
	return nil
}

func (m *Mundo) gravacoes(n int) error {
	if g := m.recomendacoes.Gravacoes(); g != n {
		return fmt.Errorf("gravações de recomendação: esperado %d, obtido %d", n, g)
	}
	return nil
}

func (m *Mundo) recomendacaoGravadaNoEstado(estado string) error {
	rec, ok := m.recomendacoes.BuscarPorID(m.recomendacao.ID)
	if !ok {
		return errors.New("recomendação não está no repositório")
	}
	if string(rec.Status) != estado {
		return fmt.Errorf("estado gravado: esperado %s, obtido %s", estado, rec.Status)
	}
	return nil
}

// ======================= ciclo da recomendação =======================

func (m *Mundo) recomendacaoNoEstado(estado string) error {
	agora := m.clock.Agora()
	r := entity.NovaRecomendacao(uuid.New(), id("c-001"), id("prod:qualquer"), "teste", agora)
	switch entity.StatusRecomendacao(estado) {
	case entity.StatusGerada:
	case entity.StatusAtiva:
		if err := r.Disponibilizar(agora, m.validadeDias); err != nil {
			return err
		}
	case entity.StatusExpirada:
		if err := r.Disponibilizar(agora, m.validadeDias); err != nil {
			return err
		}
		if err := r.Expirar(agora); err != nil {
			return err
		}
	default:
		return fmt.Errorf("estado de recomendação desconhecido: %q", estado)
	}
	m.recomendacao = r
	return nil
}

func (m *Mundo) acaoRecomendacao(acao string) error {
	switch acao {
	case "disponibilizar":
		m.err = m.recomendacao.Disponibilizar(m.clock.Agora(), m.validadeDias)
	case "expirar":
		m.err = m.recomendacao.Expirar(m.clock.Agora())
	default:
		return fmt.Errorf("ação desconhecida: %q", acao)
	}
	return nil
}

func (m *Mundo) recomendacaoAtivada(quando string, dias int) error {
	t, err := instante(quando)
	if err != nil {
		return err
	}
	r := entity.NovaRecomendacao(uuid.New(), id("c-001"), id("prod:qualquer"), "teste", t)
	if err := r.Disponibilizar(t, dias); err != nil {
		return err
	}
	m.recomendacao = r
	return m.recomendacoes.Salvar(m.ctx, r)
}

func (m *Mundo) passamDiasEExpira(dias int) error {
	m.clock.Avancar(time.Duration(dias) * 24 * time.Hour)
	if _, err := m.uc().expirar.Executar(m.ctx); err != nil {
		return err
	}
	rec, ok := m.recomendacoes.BuscarPorID(m.recomendacao.ID)
	if !ok {
		return errors.New("recomendação sumiu do repositório")
	}
	m.recomendacao = rec
	return nil
}

func (m *Mundo) recomendacaoEstado(estado string) error {
	if string(m.recomendacao.Status) != estado {
		return fmt.Errorf("estado: esperado %s, obtido %s", estado, m.recomendacao.Status)
	}
	return nil
}

func (m *Mundo) validade(s string) error {
	return m.confereData("validade", m.recomendacao.Validade, s)
}

// ======================= cadastro (UC-09) =======================

func (m *Mundo) cadastra(nome, email string) error {
	m.clienteCadastrado, m.err = m.uc().cadastrar.Executar(m.ctx, nome, email)
	return nil
}

func (m *Mundo) cadastradoAtivo() error {
	if err := m.semErro(); err != nil {
		return err
	}
	c, err := m.clientes.BuscarPorID(m.ctx, m.clienteCadastrado.ID)
	if err != nil {
		return err
	}
	if !c.Ativo {
		return errors.New("cliente cadastrado não está ativo")
	}
	return nil
}

func (m *Mundo) estadoConsentimentoCadastrado(estado string) error {
	c, err := m.uc().consentimento.Consultar(m.ctx, m.clienteCadastrado.ID)
	if err != nil {
		return err
	}
	if string(c.Status) != estado {
		return fmt.Errorf("estado: esperado %s, obtido %s", estado, c.Status)
	}
	return nil
}
