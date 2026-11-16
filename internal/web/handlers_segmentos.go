package web

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"perfila/internal/application/usecase"
	"perfila/internal/domain/entity"
	"perfila/internal/shared/errs"
)

// Edição de segmentos pelo painel (Spec 3.7.1 v1.6, RN-14, Q-33, Q-42). Os handlers só montam o
// critério a partir das linhas do formulário e traduzem erros; as regras ficam no domínio.

// linhaCondicao é uma linha do formulário: campo, operador e valor (uma condição da RN-06).
type linhaCondicao struct{ Campo, Operador, Valor string }

// opcaoCriterio é um item de seletor.
type opcaoCriterio struct{ Valor, Rotulo string }

func camposCriterio() []opcaoCriterio {
	return []opcaoCriterio{{"frequenciaCompras", "Compras"}, {"ticketMedio", "Ticket médio"}, {"scoreEngajamento", "Engajamento"}}
}

func operadoresCriterio() []opcaoCriterio {
	return []opcaoCriterio{{">=", "≥ maior ou igual"}, {">", "> maior"}, {"<=", "≤ menor ou igual"}, {"<", "< menor"}, {"==", "= igual"}}
}

type testeCriterio struct {
	Frequencia, Ticket, Score string
	Feito, Casa               bool
}

type dadosSegmento struct {
	Seg        entity.Segmento
	Prioridade string // como digitada, para voltar ao formulário
	Novo       bool
	Original   string // critério gravado, visível acima das linhas
	Linhas     []linhaCondicao
	Erros      map[string]string // campo → mensagem (errs.ErroCampo e leitura do formulário)
	Erro       *alerta
	Teste      testeCriterio
	Campos     []opcaoCriterio
	Operadores []opcaoCriterio
	Previa     *usecase.PreviaSegmentacao
	Confirmar  string // na prévia: "salvar", "ativar" ou "desativar"
}

func (p *painel) rotasSegmentos(mux *http.ServeMux, sessao func(entity.Permissao, http.HandlerFunc) http.Handler) {
	editar := entity.PermEditarConfiguracao
	mux.Handle("GET /painel/segmentos/novo", sessao(editar, p.telaNovoSegmento))
	mux.Handle("GET /painel/segmentos/{id}", sessao(editar, p.telaSegmento))
	mux.Handle("POST /painel/segmentos/previa", sessao(editar, p.previaSegmento))
	mux.Handle("POST /painel/segmentos/testar", sessao(editar, p.testarCriterio))
	mux.Handle("POST /painel/segmentos", sessao(editar, p.salvarSegmento))
	for _, acao := range []string{"ativar", "desativar"} {
		mux.Handle("GET /painel/segmentos/{id}/"+acao, sessao(editar, p.confirmarSituacaoSegmento(acao)))
		mux.Handle("POST /painel/segmentos/{id}/"+acao, sessao(editar, p.mudarSituacaoSegmento(acao)))
	}
}

func novosDadosSegmento() dadosSegmento {
	return dadosSegmento{Erros: map[string]string{}, Campos: camposCriterio(), Operadores: operadoresCriterio()}
}

// linhasDoCriterio divide o critério gravado nas linhas do formulário (só exibição).
func linhasDoCriterio(criterio string) []linhaCondicao {
	var out []linhaCondicao
	if strings.TrimSpace(criterio) == "" {
		return out
	}
	for _, trecho := range strings.Split(criterio, "&&") {
		if m := reCondicao.FindStringSubmatch(trecho); m != nil {
			out = append(out, linhaCondicao{m[1], m[2], m[3]})
		} else {
			out = append(out, linhaCondicao{Campo: strings.TrimSpace(trecho)})
		}
	}
	return out
}

func (p *painel) mostrarSegmento(w http.ResponseWriter, r *http.Request, status int, d dadosSegmento) {
	titulo := "Novo segmento"
	if !d.Novo {
		titulo = "Segmento " + d.Seg.Nome
	}
	if len(d.Linhas) == 0 && (d.Novo || !d.Seg.EhPadrao()) {
		d.Linhas = []linhaCondicao{{}}
	}
	p.render(w, r, status, "segmento", pagina{Titulo: titulo, Nav: "segmentos", Dados: d})
}

// GET /painel/segmentos/novo
func (p *painel) telaNovoSegmento(w http.ResponseWriter, r *http.Request) {
	d := novosDadosSegmento()
	d.Novo, d.Seg.Ativo = true, true
	p.mostrarSegmento(w, r, http.StatusOK, d)
}

// GET /painel/segmentos/{id}
func (p *painel) telaSegmento(w http.ResponseWriter, r *http.Request) {
	s, ok := p.carregarSegmento(w, r, r.PathValue("id"))
	if !ok {
		return
	}
	d := novosDadosSegmento()
	d.Seg, d.Prioridade, d.Original, d.Linhas = s, strconv.Itoa(s.Prioridade), s.Criterio, linhasDoCriterio(s.Criterio)
	if !s.EhPadrao() {
		d.Linhas = append(d.Linhas, linhaCondicao{})
	}
	p.mostrarSegmento(w, r, http.StatusOK, d)
}

func (p *painel) carregarSegmento(w http.ResponseWriter, r *http.Request, texto string) (entity.Segmento, bool) {
	id, err := uuid.Parse(texto)
	if err == nil {
		var s entity.Segmento
		if s, err = p.d.ListarSegmentos.Consultar(r.Context(), id); err == nil {
			return s, true
		}
	} else {
		err = errs.ErrSegmentoNaoEncontrado
	}
	p.falha(w, r, err, "")
	return entity.Segmento{}, false
}

// lerSegmento lê o formulário: monta o critério juntando as linhas completas com "&&" e marca as
// linhas incompletas. O critério montado é validado pelo domínio (RN-14), não aqui.
func (p *painel) lerSegmento(w http.ResponseWriter, r *http.Request) (dadosSegmento, bool) {
	d := novosDadosSegmento()
	d.Seg.Nome, d.Seg.Descricao = strings.TrimSpace(r.PostFormValue("nome")), strings.TrimSpace(r.PostFormValue("descricao"))
	d.Prioridade = strings.TrimSpace(r.PostFormValue("prioridade"))
	d.Novo, d.Seg.Ativo = true, true
	if texto := r.PostFormValue("id"); texto != "" {
		atual, ok := p.carregarSegmento(w, r, texto)
		if !ok {
			return d, false
		}
		d.Novo, d.Seg.ID, d.Seg.Ativo, d.Original = false, atual.ID, atual.Ativo, atual.Criterio
	}
	if n, err := strconv.Atoi(d.Prioridade); err == nil {
		d.Seg.Prioridade = n
	} else {
		d.Erros["prioridade"] = "Informe a prioridade com um número inteiro."
	}
	campos, ops, valores := r.PostForm["campo"], r.PostForm["operador"], r.PostForm["valor"]
	var conds []string
	for i := range campos {
		l := linhaCondicao{Campo: strings.TrimSpace(campos[i])}
		if i < len(ops) {
			l.Operador = strings.TrimSpace(ops[i])
		}
		if i < len(valores) {
			l.Valor = strings.TrimSpace(valores[i])
		}
		if l == (linhaCondicao{}) {
			continue
		}
		d.Linhas = append(d.Linhas, l)
		if l.Campo == "" || l.Operador == "" || l.Valor == "" {
			d.Erros["criterio"] = fmt.Sprintf("Complete a condição %d: campo, operador e valor.", len(d.Linhas))
			continue
		}
		if v, err := lerDecimal(l.Valor); err == nil {
			l.Valor = strconv.FormatFloat(v, 'f', -1, 64)
		}
		conds = append(conds, l.Campo+" "+l.Operador+" "+l.Valor)
	}
	d.Seg.Criterio = strings.Join(conds, " && ")
	d.Teste = testeCriterio{Frequencia: r.PostFormValue("teste_frequencia"), Ticket: r.PostFormValue("teste_ticket"),
		Score: r.PostFormValue("teste_score")}
	return d, true
}

// erroNoFormulario põe a mensagem do errs.ErroCampo ao lado do campo; outro erro vira alerta.
func erroNoFormulario(r *http.Request, d *dadosSegmento, err error) {
	var ec *errs.ErroCampo
	if errors.As(err, &ec) && ec.Campo != "ativo" {
		d.Erros[ec.Campo] = ec.Mensagem
		return
	}
	d.Erro = alertaDoErro(r, err)
}

// alertaDoErro mostra a mensagem do errs.ErroCampo, ou o texto do catálogo pelo código.
func alertaDoErro(r *http.Request, err error) *alerta {
	var ec *errs.ErroCampo
	if errors.As(err, &ec) {
		return &alerta{Tom: "critical", Titulo: "Não foi possível salvar", Texto: ec.Mensagem, Agora: true}
	}
	titulo, texto := mensagem(err, "")
	return &alerta{Tom: "critical", Titulo: titulo, Texto: texto, Agora: true, RequestID: r.Header.Get("X-Request-ID")}
}

func alvoSegmento(id uuid.UUID) (entity.TipoAlvo, uuid.UUID) {
	if id == uuid.Nil {
		return entity.AlvoNenhum, uuid.Nil
	}
	return entity.AlvoSegmento, id
}

// POST /painel/segmentos/previa — "Adicionar condição" e "Voltar ao formulário" devolvem o
// formulário; senão, a prévia (Q-42) com o botão Salvar.
func (p *painel) previaSegmento(w http.ResponseWriter, r *http.Request) {
	d, ok := p.lerSegmento(w, r)
	if !ok {
		return
	}
	switch r.PostFormValue("acao") {
	case "linha":
		d.Linhas = append(d.Linhas, linhaCondicao{})
		p.mostrarSegmento(w, r, http.StatusOK, d)
		return
	case "editar":
		p.mostrarSegmento(w, r, http.StatusOK, d)
		return
	}
	if len(d.Erros) > 0 {
		p.mostrarSegmento(w, r, http.StatusUnprocessableEntity, d)
		return
	}
	p.mostrarPrevia(w, r, d, "salvar")
}

// mostrarPrevia calcula e registra a prévia; erro de validação volta ao formulário (ou, ao ativar
// e desativar, à prévia sem o botão de confirmar).
func (p *painel) mostrarPrevia(w http.ResponseWriter, r *http.Request, d dadosSegmento, confirmar string) {
	previa, err := p.d.PreverSegmentacao.Executar(r.Context(), operadorAtual(r), d.Seg)
	tipo, alvo := alvoSegmento(d.Seg.ID)
	if e := p.auditar(r, entity.AcaoPreverSegmentacao, tipo, alvo, err); e != nil && err == nil {
		p.falha(w, r, e, "")
		return
	}
	if err != nil {
		p.logarFalha(r, err)
		if confirmar == "salvar" {
			erroNoFormulario(r, &d, err)
			p.mostrarSegmento(w, r, statusDe(err), d)
			return
		}
		d.Erro = alertaDoErro(r, err)
		p.render(w, r, statusDe(err), "previa_segmento", pagina{Titulo: "Prévia", Nav: "segmentos", Dados: d})
		return
	}
	d.Previa, d.Confirmar = &previa, confirmar
	p.render(w, r, http.StatusOK, "previa_segmento", pagina{Titulo: "Prévia", Nav: "segmentos", Dados: d})
}

// POST /painel/segmentos — o botão Salvar da prévia (RN-14); campos alterados vão para a auditoria.
func (p *painel) salvarSegmento(w http.ResponseWriter, r *http.Request) {
	d, ok := p.lerSegmento(w, r)
	if !ok {
		return
	}
	if len(d.Erros) > 0 {
		p.mostrarSegmento(w, r, http.StatusUnprocessableEntity, d)
		return
	}
	alt, err := p.d.SalvarSegmento.Executar(r.Context(), operadorAtual(r), d.Seg)
	if err == nil && len(alt.Campos) == 0 {
		http.Redirect(w, r, "/painel/segmentos?aviso=sem-alteracoes", http.StatusSeeOther)
		return
	}
	acao, id := entity.AcaoEditarSegmento, d.Seg.ID
	if d.Novo {
		acao, id = entity.AcaoCriarSegmento, alt.Segmento.ID
	}
	tipo, alvo := alvoSegmento(id)
	p.auditar(r, acao, tipo, alvo, err, alt.Campos...)
	if err != nil {
		p.logarFalha(r, err)
		erroNoFormulario(r, &d, err)
		p.mostrarSegmento(w, r, statusDe(err), d)
		return
	}
	p.acao(r, string(acao), "")
	http.Redirect(w, r, "/painel/segmentos?aviso=segmento-salvo", http.StatusSeeOther)
}

// GET /painel/segmentos/{id}/ativar|desativar — prévia da mudança de situação, com confirmação.
func (p *painel) confirmarSituacaoSegmento(acao string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, ok := p.carregarSegmento(w, r, r.PathValue("id"))
		if !ok {
			return
		}
		d := novosDadosSegmento()
		d.Seg, d.Original = s, s.Criterio
		d.Seg.Ativo = acao == "ativar"
		p.mostrarPrevia(w, r, d, acao)
	}
}

// POST /painel/segmentos/{id}/ativar|desativar
func (p *painel) mudarSituacaoSegmento(acao string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			p.falha(w, r, errs.ErrSegmentoNaoEncontrado, "")
			return
		}
		var alt usecase.AlteracaoSegmento
		acaoAud := entity.AcaoAtivarSegmento
		if acao == "ativar" {
			alt, err = p.d.SalvarSegmento.Ativar(r.Context(), operadorAtual(r), id)
		} else {
			acaoAud = entity.AcaoDesativarSegmento
			alt, err = p.d.SalvarSegmento.Desativar(r.Context(), operadorAtual(r), id)
		}
		p.auditar(r, acaoAud, entity.AlvoSegmento, id, err, alt.Campos...)
		if err != nil {
			p.logarFalha(r, err)
			p.mostrarSegmentos(w, r, statusDe(err), alertaDoErro(r, err))
			return
		}
		p.acao(r, string(acaoAud), "")
		http.Redirect(w, r, "/painel/segmentos?aviso=segmento-"+acao[:len(acao)-1]+"do", http.StatusSeeOther)
	}
}

// POST /painel/segmentos/testar — avalia o critério das linhas com os três valores digitados.
func (p *painel) testarCriterio(w http.ResponseWriter, r *http.Request) {
	d, ok := p.lerSegmento(w, r)
	if !ok {
		return
	}
	freq, e1 := strconv.Atoi(strings.TrimSpace(d.Teste.Frequencia))
	ticket, e2 := lerDecimal(d.Teste.Ticket)
	score, e3 := lerDecimal(d.Teste.Score)
	if errors.Join(e1, e2, e3) != nil {
		d.Erros["teste"] = "Informe números nos três valores do teste."
	}
	if d.Erros["criterio"] != "" || d.Erros["teste"] != "" {
		p.mostrarSegmento(w, r, http.StatusUnprocessableEntity, d)
		return
	}
	casa, err := p.d.PreverSegmentacao.TestarCriterio(d.Seg.Criterio, freq, ticket, score)
	if err != nil {
		erroNoFormulario(r, &d, err)
		p.mostrarSegmento(w, r, statusDe(err), d)
		return
	}
	d.Teste.Feito, d.Teste.Casa = true, casa
	p.mostrarSegmento(w, r, http.StatusOK, d)
}
