package web

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"perfila/internal/application/usecase"
	"perfila/internal/domain/entity"
	"perfila/internal/shared/errs"
)

func (p *painel) rotasFicha(mux *http.ServeMux, sessao func(http.HandlerFunc) http.Handler) {
	mux.Handle("GET /painel/clientes/{id}", sessao(p.fichaVisaoGeral))
	mux.Handle("POST /painel/clientes/{id}/processamento", sessao(p.processar))
	mux.Handle("GET /painel/clientes/{id}/eventos", sessao(p.fichaEventos))
	mux.Handle("POST /painel/clientes/{id}/eventos", sessao(p.registrarEvento))
	mux.Handle("GET /painel/clientes/{id}/consentimento", sessao(p.fichaConsentimento))
	mux.Handle("GET /painel/clientes/{id}/consentimento/{acao}", sessao(p.confirmarConsentimento))
	mux.Handle("POST /painel/clientes/{id}/consentimento/{acao}", sessao(p.alterarConsentimento))
}

// etapaVista é uma etapa do PipelineTrace.
type etapaVista struct {
	N              int
	Nome, Estado   string
	Rotulo, Resumo string
}

type formEvento struct {
	Tipo, Valor, Categoria, Origem, Quando string
	Erros                                  map[string]string
}

type formConsentimento struct {
	Finalidade, VersaoTermo    string
	ErroFinalidade, ErroVersao string
}

// dadosFicha alimenta as três abas da ficha (docs/design/02-telas-e-fluxos.md, Ficha do cliente).
type dadosFicha struct {
	Aba           string
	Cliente       entity.Cliente
	Consentimento entity.ConsentimentoLGPD
	Valido        bool    // consentimento CONCEDIDO: comportamento pode ser lido (inativo: só leitura, Q-31)
	Bloqueio      *alerta // sem consentimento: nada de comportamento é lido
	Inativo       *alerta // cadastro inativo com consentimento: histórico só para leitura
	Falha         *alerta // processamento ou registro que acabou de falhar

	Resultado     *usecase.ResultadoCliente
	Processado    bool
	Etapas        []etapaVista
	Condicoes     []condicao
	Recomendacao  *entity.RecomendacaoProduto
	Produto       string
	Historico     []entity.RecomendacaoProduto
	MetaEventos   int
	EventosScore  string
	Eventos       []entity.EventoComportamental
	MaisEventos   bool
	Evento        formEvento
	Concessao     formConsentimento
	Confirmar     string // "conceder" ou "revogar" na página de confirmação
	ErroConfirmar *alerta
}

func (d dadosFicha) PodeConceder() bool { return d.Consentimento.Status != entity.StatusConcedido }

// PodeAgir: processar e registrar eventos exigem RN-01 completa (consentimento e cliente ativo).
func (d dadosFicha) PodeAgir() bool { return d.Valido && d.Cliente.Ativo }

// marco é um ponto da linha do tempo do consentimento (aba Consentimento).
type marco struct {
	Data            time.Time
	Titulo, Detalhe string
	Icone           string
}

// LinhaConsentimento monta, a partir das datas guardadas (Q-01), o que aconteceu com o
// consentimento, do mais recente para o mais antigo.
func (d dadosFicha) LinhaConsentimento() []marco {
	c := d.Consentimento
	ms := []marco{{Data: d.Cliente.DataCadastro, Titulo: "Cadastro", Detalhe: "Nome e e-mail registrados, sem autorização de uso.", Icone: "vazado"}}
	if !c.DataConsentimento.IsZero() {
		m := marco{Data: c.DataConsentimento, Titulo: "Consentimento concedido", Icone: "ok",
			Detalhe: "Finalidade " + c.Finalidade + " · termo " + c.VersaoTermo}
		if c.DataRevogacao != nil && c.DataRevogacao.Before(c.DataConsentimento) {
			m.Titulo = "Consentimento concedido novamente"
		}
		ms = append(ms, m)
	}
	if c.DataRevogacao != nil {
		ms = append(ms, marco{Data: *c.DataRevogacao, Titulo: "Consentimento revogado", Icone: "barra",
			Detalhe: "O processamento parou na hora; os dados registrados até aqui continuam guardados."})
	}
	sort.SliceStable(ms, func(i, j int) bool { return ms[i].Data.After(ms[j].Data) })
	return ms
}

func (d dadosFicha) NomeProduto(id uuid.UUID) string {
	if d.Resultado != nil {
		if prod, ok := d.Resultado.Produtos[id]; ok {
			return prod.Nome
		}
	}
	return uuidCurto(id)
}

// carregarFicha lê cadastro e consentimento (sem exigir consentimento, Q-21) e, só com
// consentimento válido, o comportamento (ConsultarResultado aplica RN-01, C-03).
func (p *painel) carregarFicha(w http.ResponseWriter, r *http.Request, aba string) (dadosFicha, bool) {
	id, ok := idDaRota(r)
	if !ok {
		p.falha(w, r, errs.ErrClienteNaoEncontrado, "")
		return dadosFicha{}, false
	}
	ficha, err := p.d.ConsultarCliente.Executar(r.Context(), id)
	if err != nil {
		p.falha(w, r, err, "")
		return dadosFicha{}, false
	}
	d := dadosFicha{Aba: aba, Cliente: ficha.Cliente, Consentimento: ficha.Consentimento, MetaEventos: p.d.MetaEventos,
		Concessao: formConsentimento{Finalidade: p.d.Finalidade, VersaoTermo: p.d.VersaoTermo}}
	d.Valido = ficha.Consentimento.EstaValido()
	if !d.Valido {
		d.Bloqueio = p.alertaBloqueio(d)
		return d, true
	}
	if !ficha.Cliente.Ativo {
		d.Inativo = &alerta{Tom: "warn", Titulo: "Cadastro inativo", Teste: "cadastro-inativo",
			Texto: "Processamento e novos eventos ficam bloqueados enquanto o cadastro estiver inativo. O histórico abaixo é só para consulta."}
	}
	res, err := p.d.ConsultarResultado.Executar(r.Context(), id)
	if err != nil {
		if errors.Is(err, errs.ErrConsentimentoInvalido) {
			d.Valido, d.Inativo, d.Bloqueio = false, nil, p.alertaBloqueio(d)
			return d, true
		}
		p.falha(w, r, err, "")
		return dadosFicha{}, false
	}
	d.preencherResultado(res)
	return d, true
}

// alertaBloqueio explica a falta de consentimento, diferente para quem nunca concedeu e para
// quem revogou (RN-03: os dados ficam guardados, mas não podem ser usados).
func (p *painel) alertaBloqueio(d dadosFicha) *alerta {
	c := d.Consentimento
	if c.Status == entity.StatusRevogado && c.DataRevogacao != nil {
		return &alerta{Tom: "critical", Titulo: "Consentimento revogado", Teste: "bloqueio-consentimento",
			Texto: "O titular revogou o consentimento em " + dataHora(*c.DataRevogacao) + " (tinha concedido em " + dataHora(c.DataConsentimento) +
				"). Os dados registrados até a revogação continuam guardados, mas não podem ser lidos nem processados. Só uma nova concessão os libera."}
	}
	return &alerta{Tom: "critical", Titulo: "Sem consentimento", Teste: "bloqueio-consentimento",
		Texto: "O titular ainda não autorizou o uso dos dados. Só o cadastro (nome e e-mail) está guardado, e nenhum evento pode ser registrado. Conceda o consentimento no selo acima."}
}

func (d *dadosFicha) preencherResultado(res usecase.ResultadoCliente) {
	d.Resultado = &res
	d.Eventos = res.Eventos
	if len(d.Eventos) > 50 {
		d.Eventos, d.MaisEventos = d.Eventos[:50], true
	}
	d.Processado = res.Perfil != nil
	for i := range res.Recomendacoes {
		rec := res.Recomendacoes[i]
		if rec.Status == entity.StatusAtiva && d.Recomendacao == nil {
			d.Recomendacao = &rec
			continue
		}
		d.Historico = append(d.Historico, rec)
	}
	if res.Perfil != nil {
		crit := ""
		if res.Segmento != nil {
			crit = res.Segmento.Criterio
		}
		d.Condicoes = condicoes(crit, res.Perfil)
		if d.MetaEventos > 0 {
			n := int(math.Round(res.Perfil.ScoreEngajamento * float64(d.MetaEventos)))
			if res.Perfil.ScoreEngajamento >= 1 {
				d.EventosScore = fmt.Sprintf("Meta de %d eventos em 30 dias atingida", d.MetaEventos)
			} else {
				d.EventosScore = fmt.Sprintf("%d de %d eventos em 30 dias", n, d.MetaEventos)
			}
		}
	}
	d.Etapas = etapasSalvas(*d)
}

// etapasSalvas desenha o PipelineTrace a partir do que ficou gravado (depois de um PRG).
func etapasSalvas(d dadosFicha) []etapaVista {
	e := novasEtapas()
	res := d.Resultado
	if res == nil || res.Perfil == nil {
		return e
	}
	e[0].Estado, e[0].Resumo = "concluida", plural(len(res.Eventos), "evento", "eventos")
	e[1].Estado, e[1].Resumo = "concluida", "Ticket médio "+dinheiro(res.Perfil.TicketMedio)
	if res.Segmento == nil {
		return e
	}
	e[2].Estado, e[2].Resumo = "concluida", res.Segmento.Nome
	switch {
	case d.Recomendacao != nil:
		e[3].Estado, e[3].Resumo = "concluida", d.NomeProduto(d.Recomendacao.ProdutoID)+", válida até "+soData(derefData(d.Recomendacao.Validade))
	case len(d.Historico) > 0:
		e[3].Estado, e[3].Resumo = "atencao", "Recomendação expirada. Processe de novo para gerar outra."
	default:
		e[3].Estado, e[3].Resumo = "atencao", semProduto(res.Perfil.CategoriaPreferida)
	}
	return e
}

func novasEtapas() []etapaVista {
	return []etapaVista{
		{N: 1, Nome: "coletar", Estado: "pendente", Rotulo: "Coletar eventos"},
		{N: 2, Nome: "perfil", Estado: "pendente", Rotulo: "Gerar perfil"},
		{N: 3, Nome: "segmento", Estado: "pendente", Rotulo: "Segmentar"},
		{N: 4, Nome: "recomendacao", Estado: "pendente", Rotulo: "Recomendar"},
	}
}

func semProduto(categoria string) string {
	if categoria == "" {
		return "Sem categoria preferida: nenhuma compra com categoria ainda."
	}
	return "Nenhum produto ativo na categoria preferida (" + categoria + ")."
}

func plural(n int, um, varios string) string {
	if n == 1 {
		return "1 " + um
	}
	return fmt.Sprintf("%d %s", n, varios)
}

// motivoLegivel: "segmento:premium;categoria:livros" → "Segmento premium, categoria livros".
func motivoLegivel(m string) string {
	var partes []string
	for _, par := range strings.Split(m, ";") {
		k, v, ok := strings.Cut(par, ":")
		if !ok {
			continue
		}
		partes = append(partes, k+" "+v)
	}
	if len(partes) == 0 {
		return m
	}
	s := strings.Join(partes, ", ")
	return strings.ToUpper(s[:1]) + s[1:]
}

// GET /painel/clientes/{id} — visão geral.
func (p *painel) fichaVisaoGeral(w http.ResponseWriter, r *http.Request) {
	if d, ok := p.carregarFicha(w, r, "visao"); ok {
		p.renderFicha(w, r, http.StatusOK, d)
	}
}

func (p *painel) renderFicha(w http.ResponseWriter, r *http.Request, status int, d dadosFicha) {
	tela := "ficha"
	if d.Confirmar != "" {
		tela = "confirmar"
	}
	p.render(w, r, status, tela, pagina{Titulo: d.Cliente.Nome, Nav: "clientes", Dados: d})
}

// POST /painel/clientes/{id}/processamento — RN-09. Sucesso: PRG com o toast. Falha: a ficha com
// a etapa que parou e o aviso do que ficou salvo (fluxo C).
func (p *painel) processar(w http.ResponseWriter, r *http.Request) {
	id, ok := idDaRota(r)
	if !ok {
		p.falha(w, r, errs.ErrClienteNaoEncontrado, "")
		return
	}
	res, errProc := p.d.Processar.Executar(r.Context(), id)
	p.acao(r, "processar", id.String())
	if errProc == nil {
		codigo := "processamento-concluido"
		if res.Recomendacao == nil {
			codigo = "processamento-sem-recomendacao"
		}
		http.Redirect(w, r, fmt.Sprintf("/painel/clientes/%s?aviso=%s", id, codigo), http.StatusSeeOther)
		return
	}
	p.logarFalha(r, errProc)
	d, ok := p.carregarFicha(w, r, "visao")
	if !ok {
		return
	}
	if d.Valido {
		d.Etapas = etapasDoProcessamento(res, d)
		d.Falha = alertaProcessamento(res, errProc, r.Header.Get("X-Request-ID"))
	}
	p.renderFicha(w, r, statusDe(errProc), d)
}

// etapasDoProcessamento desenha o PipelineTrace direto do ResultadoProcessamento (Q-26).
func etapasDoProcessamento(res usecase.ResultadoProcessamento, d dadosFicha) []etapaVista {
	e := novasEtapas()
	for i, et := range res.Etapas {
		e[i].Estado = string(et.Estado)
		switch {
		case et.Estado == usecase.EtapaFalhou && statusDe(et.Erro) == http.StatusInternalServerError && et.Nome == usecase.EtapaSegmento:
			e[i].Resumo = "Falha de configuração dos segmentos. A equipe técnica foi avisada pelo log."
		case et.Estado == usecase.EtapaFalhou:
			_, e[i].Resumo = mensagem(et.Erro, "")
		case et.Estado == usecase.EtapaAtencao:
			e[i].Resumo = semProduto(res.Perfil.CategoriaPreferida)
		}
	}
	if res.Etapas[0].Estado == usecase.EtapaConcluida {
		e[0].Resumo = plural(res.Eventos, "evento", "eventos")
	}
	if res.Etapas[1].Estado == usecase.EtapaConcluida {
		e[1].Resumo = "Ticket médio " + dinheiro(res.Perfil.TicketMedio)
	}
	if res.Etapas[2].Estado == usecase.EtapaConcluida {
		e[2].Resumo = res.Segmento.Nome
	}
	return e
}

func alertaProcessamento(res usecase.ResultadoProcessamento, err error, requestID string) *alerta {
	a := &alerta{Tom: "critical", Teste: "falha-processamento", Agora: true, RequestID: requestID}
	switch {
	case res.Etapas[0].Estado == usecase.EtapaFalhou:
		a.Titulo, a.Texto = mensagem(err, "")
	case res.Etapas[1].Estado == usecase.EtapaFalhou:
		a.Titulo, a.Texto = "A geração do perfil falhou", "Os eventos continuam registrados. Nada mais foi alterado."
	case res.Etapas[2].Estado == usecase.EtapaFalhou:
		a.Titulo, a.Texto = "A segmentação falhou", "O perfil foi atualizado, mas a segmentação falhou. Nada mais foi alterado."
	default:
		a.Titulo, a.Texto = "A recomendação falhou", "O perfil e o segmento foram atualizados, mas a recomendação falhou. Nada mais foi alterado."
	}
	if statusDe(err) != http.StatusInternalServerError {
		a.RequestID = ""
	}
	return a
}

func derefData(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
