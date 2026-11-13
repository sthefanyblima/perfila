package web

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"perfila/internal/domain/entity"
)

func (p *painel) rotasConfiguracao(mux *http.ServeMux, sessao func(entity.Permissao, http.HandlerFunc) http.Handler) {
	mux.Handle("GET /painel/segmentos", sessao(entity.PermConfiguracao, p.telaSegmentos))
	mux.Handle("GET /painel/catalogo", sessao(entity.PermConfiguracao, p.telaCatalogo))
	mux.Handle("GET /painel/operacao", sessao(entity.PermOperacao, p.telaOperacao))
	mux.Handle("GET /painel/operacao/expirar", sessao(entity.PermOperacao, p.confirmarExpiracao))
	mux.Handle("POST /painel/operacao/expirar", sessao(entity.PermOperacao, p.expirar))
}

// GET /painel/segmentos — somente leitura, por prioridade (RN-06).
func (p *painel) telaSegmentos(w http.ResponseWriter, r *http.Request) {
	lista, err := p.d.ListarSegmentos.Executar(r.Context())
	if err != nil {
		p.falha(w, r, err, "")
		return
	}
	p.render(w, r, http.StatusOK, "segmentos", pagina{Titulo: "Segmentos", Nav: "segmentos", Dados: lista})
}

type grupoCatalogo struct {
	Categoria string
	Produtos  []entity.Produto
}

// GET /painel/catalogo — produtos agrupados por categoria (RN-07).
func (p *painel) telaCatalogo(w http.ResponseWriter, r *http.Request) {
	lista, err := p.d.ListarCatalogo.Executar(r.Context())
	if err != nil {
		p.falha(w, r, err, "")
		return
	}
	var grupos []grupoCatalogo
	for _, prod := range lista {
		if len(grupos) == 0 || grupos[len(grupos)-1].Categoria != prod.Categoria {
			grupos = append(grupos, grupoCatalogo{Categoria: prod.Categoria})
		}
		grupos[len(grupos)-1].Produtos = append(grupos[len(grupos)-1].Produtos, prod)
	}
	p.render(w, r, http.StatusOK, "catalogo", pagina{Titulo: "Catálogo", Nav: "catalogo", Dados: grupos})
}

type dadosOperacao struct {
	BancoOK     bool
	ModoMemoria bool
	Migracoes   string
	Confirmar   bool
}

func (p *painel) dadosOperacao(ctx context.Context) dadosOperacao {
	d := dadosOperacao{BancoOK: true, ModoMemoria: p.d.Saude == nil, Migracoes: p.d.Migracoes}
	if p.d.Saude != nil {
		c, cancelar := context.WithTimeout(ctx, 2*time.Second)
		defer cancelar()
		d.BancoOK = p.d.Saude(c) == nil
	}
	return d
}

// GET /painel/operacao — saúde do banco e expiração de recomendações (RN-08).
func (p *painel) telaOperacao(w http.ResponseWriter, r *http.Request) {
	p.render(w, r, http.StatusOK, "operacao", pagina{Titulo: "Operação", Nav: "operacao", Dados: p.dadosOperacao(r.Context())})
}

// GET /painel/operacao/expirar — confirmação sem JavaScript.
func (p *painel) confirmarExpiracao(w http.ResponseWriter, r *http.Request) {
	d := p.dadosOperacao(r.Context())
	d.Confirmar = true
	p.render(w, r, http.StatusOK, "operacao", pagina{Titulo: "Expirar recomendações vencidas", Nav: "operacao", Dados: d})
}

// POST /painel/operacao/expirar — ExpirarRecomendacoesUseCase; o total vai no toast (PRG).
func (p *painel) expirar(w http.ResponseWriter, r *http.Request) {
	n, err := p.d.Expirar.Executar(r.Context())
	if err != nil {
		p.falha(w, r, err, "")
		return
	}
	p.acao(r, "expirar_recomendacoes", "")
	http.Redirect(w, r, "/painel/operacao?aviso=expiradas&n="+strconv.Itoa(n), http.StatusSeeOther)
}
