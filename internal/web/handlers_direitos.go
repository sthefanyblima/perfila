package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"perfila/internal/application/usecase"
	"perfila/internal/shared/errs"
)

// Direitos do titular na ficha e retenção na Operação. Spec: v1.9, 3.7.1, RN-17, RN-18, Q-48 a Q-53.

func (p *painel) rotasDireitos(mux *http.ServeMux, sessao func(http.HandlerFunc) http.Handler) {
	mux.Handle("GET /painel/clientes/{id}/direitos", sessao(p.fichaDireitos))
	mux.Handle("POST /painel/clientes/{id}/cadastro", sessao(p.corrigirCadastro))
	mux.Handle("GET /painel/clientes/{id}/dados", sessao(p.baixarDados))
	mux.Handle("GET /painel/clientes/{id}/eliminar", sessao(p.confirmarEliminacao))
	mux.Handle("POST /painel/clientes/{id}/eliminar", sessao(p.eliminar))
	mux.Handle("GET /painel/operacao/retencao", sessao(p.confirmarRetencao))
	mux.Handle("POST /painel/operacao/retencao", sessao(p.aplicarRetencao))
}

// formCadastro é o painel "Corrigir cadastro" (Q-48).
type formCadastro struct {
	Nome, Email         string
	ErroNome, ErroEmail string
}

// direito registra o pedido atendido: operador, pedido e ID do cliente, nunca nome nem e-mail
// (C-09). É a auditoria dos pedidos até a v1.5 entrar na main (Q-53).
func (p *painel) direito(r *http.Request, pedido string, attrs ...slog.Attr) {
	attrs = append([]slog.Attr{slog.String("operador", operadorDe(r)), slog.String("pedido", pedido)}, attrs...)
	p.d.Log.LogAttrs(r.Context(), slog.LevelInfo, "direito_atendido", attrs...)
}

// GET /painel/clientes/{id}/direitos — vale com qualquer estado do consentimento.
func (p *painel) fichaDireitos(w http.ResponseWriter, r *http.Request) {
	if d, ok := p.carregarFicha(w, r, "direitos"); ok {
		p.renderFicha(w, r, http.StatusOK, d)
	}
}

// POST /painel/clientes/{id}/cadastro — AtualizarCadastro; sucesso volta à aba (PRG).
func (p *painel) corrigirCadastro(w http.ResponseWriter, r *http.Request) {
	f := formCadastro{Nome: strings.TrimSpace(r.PostFormValue("nome")), Email: strings.TrimSpace(r.PostFormValue("email"))}
	if f.Nome == "" {
		f.ErroNome = "Informe o nome do cliente."
	}
	if f.Email == "" {
		f.ErroEmail = "Informe o e-mail do cliente."
	}
	id, _ := idDaRota(r)
	var err error
	if f.ErroNome == "" && f.ErroEmail == "" {
		if _, err = p.d.AtualizarCadastro.Executar(r.Context(), id, f.Nome, f.Email); err == nil {
			p.direito(r, "correcao", slog.String("clienteId", id.String()))
			http.Redirect(w, r, fmt.Sprintf("/painel/clientes/%s/direitos?aviso=cadastro-corrigido", id), http.StatusSeeOther)
			return
		}
	}
	d, ok := p.carregarFicha(w, r, "direitos")
	if !ok {
		return
	}
	d.Cadastro = f
	status := http.StatusUnprocessableEntity
	if err != nil {
		p.logarFalha(r, err)
		titulo, texto := mensagem(err, "")
		d.Falha = &alerta{Tom: "critical", Titulo: titulo, Texto: texto, Agora: true, RequestID: r.Header.Get("X-Request-ID")}
		status = statusDe(err)
	}
	p.renderFicha(w, r, status, d)
}

// GET /painel/clientes/{id}/dados — o arquivo JSON do titular, como anexo e sem cache (Q-49).
func (p *painel) baixarDados(w http.ResponseWriter, r *http.Request) {
	id, ok := idDaRota(r)
	if !ok {
		p.falha(w, r, errs.ErrClienteNaoEncontrado, "")
		return
	}
	dados, err := p.d.ExportarDados.Executar(r.Context(), id)
	if err != nil {
		p.falha(w, r, err, "")
		return
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(exportacao(dados)); err != nil {
		p.falha(w, r, err, "")
		return
	}
	p.direito(r, "exportacao", slog.String("clienteId", id.String()))
	h := w.Header()
	h.Set("Content-Type", "application/json; charset=utf-8")
	h.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="perfila-dados-%s.json"`, id))
	_, _ = w.Write(buf.Bytes())
}

// GET /painel/clientes/{id}/eliminar — confirmação forte em página própria (RN-17).
func (p *painel) confirmarEliminacao(w http.ResponseWriter, r *http.Request) {
	d, ok := p.carregarFicha(w, r, "direitos")
	if !ok {
		return
	}
	if d.Eliminado {
		p.falha(w, r, errs.ErrClienteEliminado, "")
		return
	}
	d.Confirmar = "eliminar"
	p.renderFicha(w, r, http.StatusOK, d)
}

// POST /painel/clientes/{id}/eliminar — EliminarDadosTitular; sucesso vai à ficha (PRG).
func (p *painel) eliminar(w http.ResponseWriter, r *http.Request) {
	id, _ := idDaRota(r)
	confirmacao := r.PostFormValue("confirmacao")
	err := p.d.EliminarDados.Executar(r.Context(), id, confirmacao)
	if err == nil {
		p.direito(r, "eliminacao", slog.String("clienteId", id.String()))
		http.Redirect(w, r, fmt.Sprintf("/painel/clientes/%s?aviso=dados-eliminados", id), http.StatusSeeOther)
		return
	}
	p.logarFalha(r, err)
	d, ok := p.carregarFicha(w, r, "direitos")
	if !ok {
		return
	}
	if d.Eliminado {
		p.falha(w, r, err, "")
		return
	}
	d.Confirmar, d.Eliminacao = "eliminar", formEliminacao{Confirmacao: confirmacao}
	if errs.Codigo(err) == errs.ErrDadosInvalidos.Codigo {
		d.Eliminacao.Erro = "O nome digitado não confere com o cadastro."
	} else {
		titulo, texto := mensagem(err, "eliminar")
		d.ErroConfirmar = &alerta{Tom: "critical", Titulo: titulo, Texto: texto, Agora: true, RequestID: r.Header.Get("X-Request-ID")}
	}
	p.renderFicha(w, r, statusDe(err), d)
}

// formEliminacao guarda o nome digitado na confirmação e o erro do campo.
type formEliminacao struct{ Confirmacao, Erro string }

// GET /painel/operacao/retencao — confirmação sem JavaScript (RN-18).
func (p *painel) confirmarRetencao(w http.ResponseWriter, r *http.Request) {
	d := p.dadosOperacao(r.Context())
	d.ConfirmarRetencao = true
	p.render(w, r, http.StatusOK, "operacao", pagina{Titulo: "Aplicar retenção", Nav: "operacao", Dados: d})
}

// POST /painel/operacao/retencao — AplicarRetencao; o total vai no toast (PRG).
func (p *painel) aplicarRetencao(w http.ResponseWriter, r *http.Request) {
	n, err := p.d.AplicarRetencao.Executar(r.Context())
	if err != nil {
		p.falha(w, r, err, "")
		return
	}
	p.direito(r, "retencao", slog.Int("quantidade", n))
	http.Redirect(w, r, "/painel/operacao?aviso=retencao&n="+strconv.Itoa(n), http.StatusSeeOther)
}

// ---- formato do arquivo exportado (versaoFormato "1"): campos em camelCase, como na API ----

type exportacaoJSON struct {
	VersaoFormato string             `json:"versaoFormato"`
	GeradoEm      time.Time          `json:"geradoEm"`
	Cliente       clienteJSON        `json:"cliente"`
	Consentimento consentimentoJSON  `json:"consentimento"`
	Historico     []movimentoJSON    `json:"historico"`
	Eventos       []eventoJSON       `json:"eventos"`
	Perfil        *perfilJSON        `json:"perfil"`
	Segmento      *segmentoJSON      `json:"segmento"`
	Recomendacoes []recomendacaoJSON `json:"recomendacoes"`
	Produtos      []produtoJSON      `json:"produtos"`
}

type clienteJSON struct {
	ID             uuid.UUID  `json:"id"`
	Nome           string     `json:"nome"`
	Email          string     `json:"email"`
	DataCadastro   time.Time  `json:"dataCadastro"`
	Ativo          bool       `json:"ativo"`
	DataEliminacao *time.Time `json:"dataEliminacao"`
}

type consentimentoJSON struct {
	Status            string     `json:"status"`
	Finalidade        string     `json:"finalidade"`
	VersaoTermo       string     `json:"versaoTermo"`
	DataConsentimento *time.Time `json:"dataConsentimento"`
	DataRevogacao     *time.Time `json:"dataRevogacao"`
	DataAtualizacao   time.Time  `json:"dataAtualizacao"`
}

type movimentoJSON struct {
	Tipo        string    `json:"tipo"`
	Finalidade  string    `json:"finalidade"`
	VersaoTermo string    `json:"versaoTermo"`
	Data        time.Time `json:"data"`
	Origem      string    `json:"origem"`
}

type eventoJSON struct {
	ID        uuid.UUID `json:"id"`
	Tipo      string    `json:"tipo"`
	Origem    string    `json:"origem"`
	Timestamp time.Time `json:"timestamp"`
	Valor     float64   `json:"valor"`
	Categoria string    `json:"categoria"`
}

type perfilJSON struct {
	FrequenciaCompras  int     `json:"frequenciaCompras"`
	TicketMedio        float64 `json:"ticketMedio"`
	CategoriaPreferida string  `json:"categoriaPreferida"`
	ScoreEngajamento   float64 `json:"scoreEngajamento"`
}

type segmentoJSON struct {
	Nome      string `json:"nome"`
	Descricao string `json:"descricao"`
}

type recomendacaoJSON struct {
	ProdutoID     uuid.UUID  `json:"produtoId"`
	Motivo        string     `json:"motivo"`
	Status        string     `json:"status"`
	DataGeracao   time.Time  `json:"dataGeracao"`
	Validade      *time.Time `json:"validade"`
	DataExpiracao *time.Time `json:"dataExpiracao"`
}

type produtoJSON struct {
	ID        uuid.UUID `json:"id"`
	Nome      string    `json:"nome"`
	Categoria string    `json:"categoria"`
}

// exportacao monta o arquivo. O operador do movimento fica de fora: é o login da equipe, não dado
// do titular. Listas vazias saem como [] para o arquivo ter sempre as mesmas seções.
func exportacao(d usecase.DadosTitular) exportacaoJSON {
	c := d.Consentimento
	e := exportacaoJSON{VersaoFormato: "1", GeradoEm: d.GeradoEm,
		Cliente: clienteJSON{ID: d.Cliente.ID, Nome: d.Cliente.Nome, Email: d.Cliente.Email, DataCadastro: d.Cliente.DataCadastro,
			Ativo: d.Cliente.Ativo, DataEliminacao: d.Cliente.DataEliminacao},
		Consentimento: consentimentoJSON{Status: string(c.Status), Finalidade: c.Finalidade, VersaoTermo: c.VersaoTermo,
			DataRevogacao: c.DataRevogacao, DataAtualizacao: c.DataAtualizacao},
		Historico: []movimentoJSON{}, Eventos: []eventoJSON{}, Recomendacoes: []recomendacaoJSON{}, Produtos: []produtoJSON{}}
	if !c.DataConsentimento.IsZero() {
		t := c.DataConsentimento
		e.Consentimento.DataConsentimento = &t
	}
	for _, m := range d.Historico {
		e.Historico = append(e.Historico, movimentoJSON{Tipo: string(m.Tipo), Finalidade: m.Finalidade, VersaoTermo: m.VersaoTermo,
			Data: m.Data, Origem: string(m.Origem)})
	}
	for _, ev := range d.Eventos {
		e.Eventos = append(e.Eventos, eventoJSON{ID: ev.ID, Tipo: string(ev.Tipo), Origem: ev.Origem, Timestamp: ev.Timestamp,
			Valor: ev.Valor, Categoria: ev.Categoria})
	}
	if pf := d.Perfil; pf != nil {
		e.Perfil = &perfilJSON{FrequenciaCompras: pf.FrequenciaCompras, TicketMedio: pf.TicketMedio,
			CategoriaPreferida: pf.CategoriaPreferida, ScoreEngajamento: pf.ScoreEngajamento}
	}
	if s := d.Segmento; s != nil {
		e.Segmento = &segmentoJSON{Nome: s.Nome, Descricao: s.Descricao}
	}
	for _, rec := range d.Recomendacoes {
		e.Recomendacoes = append(e.Recomendacoes, recomendacaoJSON{ProdutoID: rec.ProdutoID, Motivo: rec.Motivo, Status: string(rec.Status),
			DataGeracao: rec.DataGeracao, Validade: rec.Validade, DataExpiracao: rec.DataExpiracao})
	}
	for _, prod := range d.Produtos {
		e.Produtos = append(e.Produtos, produtoJSON{ID: prod.ID, Nome: prod.Nome, Categoria: prod.Categoria})
	}
	sort.Slice(e.Produtos, func(i, j int) bool { return e.Produtos[i].Nome < e.Produtos[j].Nome })
	return e
}

// estadoEliminado é o estado mostrado no selo de um cadastro eliminado (03-textos.md).
const estadoEliminado = "ELIMINADO"
