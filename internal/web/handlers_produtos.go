package web

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/shared/errs"
)

// Edição do catálogo pelo painel (Spec 3.7.1 v1.6, RN-14, Q-43).

type dadosProduto struct {
	ID                     uuid.UUID // uuid.Nil = novo
	Nome, Categoria, Preco string
	Ativo, Confirmar       bool
	Categorias             []string
	Erros                  map[string]string
	Erro                   *alerta
}

func (p *painel) rotasProdutos(mux *http.ServeMux, sessao func(entity.Permissao, http.HandlerFunc) http.Handler) {
	editar := entity.PermEditarConfiguracao
	mux.Handle("GET /painel/catalogo/novo", sessao(editar, p.telaNovoProduto))
	mux.Handle("GET /painel/catalogo/{id}", sessao(editar, p.telaProduto(false)))
	mux.Handle("GET /painel/catalogo/{id}/desativar", sessao(editar, p.telaProduto(true)))
	mux.Handle("POST /painel/catalogo", sessao(editar, p.salvarProduto))
	mux.Handle("POST /painel/catalogo/{id}/ativar", sessao(editar, p.situacaoProduto(true)))
	mux.Handle("POST /painel/catalogo/{id}/desativar", sessao(editar, p.situacaoProduto(false)))
}

func (p *painel) mostrarProduto(w http.ResponseWriter, r *http.Request, status int, d dadosProduto) {
	if d.Erros == nil {
		d.Erros = map[string]string{}
	}
	d.Categorias, _ = p.d.ListarCatalogo.Categorias(r.Context()) // só sugestão: sem a lista, o campo continua livre
	titulo := "Novo produto"
	if d.ID != uuid.Nil {
		titulo = d.Nome
	}
	p.render(w, r, status, "produto", pagina{Titulo: titulo, Nav: "catalogo", Dados: d})
}

// GET /painel/catalogo/novo
func (p *painel) telaNovoProduto(w http.ResponseWriter, r *http.Request) {
	p.mostrarProduto(w, r, http.StatusOK, dadosProduto{Ativo: true})
}

// GET /painel/catalogo/{id} e /painel/catalogo/{id}/desativar (confirmação sem JavaScript).
func (p *painel) telaProduto(confirmar bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		prod, ok := p.carregarProduto(w, r)
		if !ok {
			return
		}
		p.mostrarProduto(w, r, http.StatusOK, dadosProduto{ID: prod.ID, Nome: prod.Nome, Categoria: prod.Categoria,
			Preco: decimal(prod.Preco, 2), Ativo: prod.Ativo, Confirmar: confirmar && prod.Ativo})
	}
}

func (p *painel) carregarProduto(w http.ResponseWriter, r *http.Request) (entity.Produto, bool) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		p.falha(w, r, errs.ErrProdutoNaoEncontrado, "")
		return entity.Produto{}, false
	}
	prod, err := p.d.ListarCatalogo.Consultar(r.Context(), id)
	if err != nil {
		p.falha(w, r, err, "")
		return entity.Produto{}, false
	}
	return prod, true
}

// POST /painel/catalogo — cria ou edita (SalvarProduto); campos alterados vão para a auditoria.
func (p *painel) salvarProduto(w http.ResponseWriter, r *http.Request) {
	d := dadosProduto{Nome: strings.TrimSpace(r.PostFormValue("nome")), Categoria: strings.TrimSpace(r.PostFormValue("categoria")),
		Preco: strings.TrimSpace(r.PostFormValue("preco")), Ativo: true, Erros: map[string]string{}}
	if texto := r.PostFormValue("id"); texto != "" {
		id, err := uuid.Parse(texto)
		if err != nil {
			p.falha(w, r, errs.ErrProdutoNaoEncontrado, "")
			return
		}
		d.ID = id
	}
	preco, err := lerDecimal(d.Preco)
	if err != nil {
		d.Erros["preco"] = "Informe o preço com números, por exemplo 90,50."
		p.mostrarProduto(w, r, http.StatusUnprocessableEntity, d)
		return
	}
	alt, err := p.d.SalvarProduto.Executar(r.Context(), operadorAtual(r),
		entity.Produto{ID: d.ID, Nome: d.Nome, Categoria: d.Categoria, Preco: preco})
	if err == nil && len(alt.Campos) == 0 {
		http.Redirect(w, r, "/painel/catalogo?aviso=sem-alteracoes", http.StatusSeeOther)
		return
	}
	acao, alvo := entity.AcaoEditarProduto, d.ID
	if d.ID == uuid.Nil {
		acao, alvo = entity.AcaoCriarProduto, alt.Produto.ID
	}
	tipo := entity.AlvoProduto
	if alvo == uuid.Nil {
		tipo = entity.AlvoNenhum
	}
	p.auditar(r, acao, tipo, alvo, err, alt.Campos...)
	if err != nil {
		p.logarFalha(r, err)
		var ec *errs.ErroCampo
		if errors.As(err, &ec) {
			d.Erros[ec.Campo] = ec.Mensagem
		} else {
			d.Erro = alertaDoErro(r, err)
		}
		p.mostrarProduto(w, r, statusDe(err), d)
		return
	}
	p.acao(r, string(acao), "")
	http.Redirect(w, r, "/painel/catalogo?aviso=produto-salvo", http.StatusSeeOther)
}

// POST /painel/catalogo/{id}/ativar|desativar
func (p *painel) situacaoProduto(ativar bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			p.falha(w, r, errs.ErrProdutoNaoEncontrado, "")
			return
		}
		acao, aviso := entity.AcaoDesativarProduto, "produto-desativado"
		var campos []string
		if ativar {
			acao, aviso = entity.AcaoAtivarProduto, "produto-ativado"
			alt, e := p.d.SalvarProduto.Ativar(r.Context(), operadorAtual(r), id)
			err, campos = e, alt.Campos
		} else {
			alt, e := p.d.SalvarProduto.Desativar(r.Context(), operadorAtual(r), id)
			err, campos = e, alt.Campos
		}
		p.auditar(r, acao, entity.AlvoProduto, id, err, campos...)
		if err != nil {
			p.falha(w, r, err, "")
			return
		}
		p.acao(r, string(acao), "")
		http.Redirect(w, r, "/painel/catalogo?aviso="+aviso, http.StatusSeeOther)
	}
}
