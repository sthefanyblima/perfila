package web

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"perfila/internal/application/usecase"
	"perfila/internal/domain/entity"
)

func (p *painel) rotasClientes(mux *http.ServeMux, sessao func(entity.Permissao, http.HandlerFunc) http.Handler) {
	mux.Handle("GET /painel/clientes", sessao(entity.PermClientes, p.telaClientes))
	mux.Handle("GET /painel/clientes/novo", sessao(entity.PermClientes, p.telaNovoCliente))
	mux.Handle("POST /painel/clientes", sessao(entity.PermClientes, p.cadastrar))
}

type opcao struct {
	Valor, Rotulo string
	Atual         bool
}

type dadosClientes struct {
	Busca, Filtro     string
	BuscaCurta        bool
	Lista             usecase.PaginaClientes
	Inicio, Fim       int
	Anterior, Proxima string
	Filtros           []opcao
	FiltroAtivo       bool
}

// GET /painel/clientes — busca (2 caracteres ou mais, ou um ID), filtro de consentimento e
// páginas de 20. O termo fica só na query string, que ninguém registra (Q-24).
func (p *painel) telaClientes(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	d := dadosClientes{Busca: strings.TrimSpace(q.Get("q")), Filtro: q.Get("consentimento")}
	f := usecase.FiltroListagem{Busca: d.Busca}
	if utf8.RuneCountInString(d.Busca) == 1 {
		d.BuscaCurta, f.Busca = true, ""
	}
	for _, o := range []opcao{{"", "Todos", false}, {string(entity.StatusConcedido), "Concedido", false},
		{string(entity.StatusNaoConcedido), "Não concedido", false}, {string(entity.StatusRevogado), "Revogado", false}} {
		o.Atual = o.Valor == d.Filtro
		if o.Atual && o.Valor != "" {
			s := entity.StatusConsentimento(o.Valor)
			f.Consentimento = &s
		}
		d.Filtros = append(d.Filtros, o)
	}
	d.FiltroAtivo = f.Busca != "" || f.Consentimento != nil
	f.Pagina, _ = strconv.Atoi(q.Get("pagina"))
	lista, err := p.d.ListarClientes.Executar(r.Context(), f)
	if err != nil {
		p.falha(w, r, err, "")
		return
	}
	d.Lista = lista
	if n := len(lista.Linhas); n > 0 {
		d.Inicio = (lista.Pagina-1)*usecase.ClientesPorPagina + 1
		d.Fim = d.Inicio + n - 1
	}
	linkPagina := func(n int) string {
		v := url.Values{}
		if d.Busca != "" {
			v.Set("q", d.Busca)
		}
		if d.Filtro != "" {
			v.Set("consentimento", d.Filtro)
		}
		v.Set("pagina", strconv.Itoa(n))
		return "/painel/clientes?" + v.Encode()
	}
	if lista.Pagina > 1 {
		d.Anterior = linkPagina(lista.Pagina - 1)
	}
	if d.Fim < lista.Total {
		d.Proxima = linkPagina(lista.Pagina + 1)
	}
	p.render(w, r, http.StatusOK, "clientes", pagina{Titulo: "Clientes", Nav: "clientes", Dados: d})
}

type dadosNovo struct {
	Nome, Email         string
	ErroNome, ErroEmail string
	Erro                *alerta
}

// GET /painel/clientes/novo — formulário de uma coluna (UC-09).
func (p *painel) telaNovoCliente(w http.ResponseWriter, r *http.Request) {
	p.render(w, r, http.StatusOK, "novo", pagina{Titulo: "Cadastrar cliente", Nav: "clientes", Dados: dadosNovo{}})
}

// POST /painel/clientes — UC-09; no sucesso, vai para a ficha (PRG). Spec: 3.7.1.
func (p *painel) cadastrar(w http.ResponseWriter, r *http.Request) {
	d := dadosNovo{Nome: strings.TrimSpace(r.PostFormValue("nome")), Email: strings.TrimSpace(r.PostFormValue("email"))}
	if d.Nome == "" {
		d.ErroNome = "Informe o nome do cliente."
	}
	if d.Email == "" {
		d.ErroEmail = "Informe o e-mail do cliente."
	}
	if d.ErroNome != "" || d.ErroEmail != "" {
		p.render(w, r, http.StatusUnprocessableEntity, "novo", pagina{Titulo: "Cadastrar cliente", Nav: "clientes", Dados: d})
		return
	}
	c, err := p.d.Cadastrar.Executar(r.Context(), d.Nome, d.Email)
	if err != nil {
		titulo, texto := mensagem(err, "")
		d.Erro = &alerta{Tom: "critical", Titulo: titulo, Texto: texto, Agora: true, RequestID: r.Header.Get("X-Request-ID")}
		p.logarFalha(r, err)
		p.render(w, r, statusDe(err), "novo", pagina{Titulo: "Cadastrar cliente", Nav: "clientes", Dados: d})
		return
	}
	p.acao(r, "cadastrar_cliente", c.ID.String())
	http.Redirect(w, r, fmt.Sprintf("/painel/clientes/%s?aviso=cliente-cadastrado", c.ID), http.StatusSeeOther)
}

// falha mostra a página de erro com a mensagem do código estável; 500 vai para o log só com o
// caminho e a causa técnica (sem dado pessoal: os erros citam só IDs).
func (p *painel) falha(w http.ResponseWriter, r *http.Request, err error, contexto string) {
	p.logarFalha(r, err)
	titulo, texto := mensagem(err, contexto)
	p.renderErro(w, r, statusDe(err), titulo, texto)
}

func (p *painel) logarFalha(r *http.Request, err error) {
	if statusDe(err) == http.StatusInternalServerError {
		p.d.Log.ErrorContext(r.Context(), "painel_erro_interno", "caminho", r.URL.Path, "erro", err.Error())
	}
}

// idDaRota lê o {id} do caminho; inválido vira 404 de cliente.
func idDaRota(r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue("id"))
	return id, err == nil
}
