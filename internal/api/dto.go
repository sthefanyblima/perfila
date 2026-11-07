// Package api expõe os casos de uso por HTTP (net/http, padrões do Go 1.22).
// Handlers só traduzem HTTP ⇄ casos de uso; nenhuma regra de negócio mora aqui.
//
// Spec: 3.7, RN-09, C-01, C-03. Fonte: UML7 (componente API), UML4 (sequência).
package api

import (
	"time"

	"perfila/internal/domain/entity"
)

// ---- entrada ----

type CadastrarClienteRequest struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
}

type ConcederConsentimentoRequest struct {
	Finalidade  string `json:"finalidade"`
	VersaoTermo string `json:"versaoTermo"`
}

type RegistrarEventoRequest struct {
	Tipo      string    `json:"tipo"`
	Origem    string    `json:"origem"`
	Timestamp time.Time `json:"timestamp"`
	Valor     float64   `json:"valor"`
	Categoria string    `json:"categoria"`
}

// ---- saída ----

type ErroResponse struct {
	Codigo   string `json:"codigo"`
	Mensagem string `json:"mensagem"`
}

type ClienteResponse struct {
	ID           string    `json:"id"`
	Nome         string    `json:"nome"`
	Email        string    `json:"email"`
	DataCadastro time.Time `json:"dataCadastro"`
	Ativo        bool      `json:"ativo"`
}

type ConsentimentoResponse struct {
	ClienteID         string     `json:"clienteId"`
	Status            string     `json:"status"`
	Valido            bool       `json:"valido"`
	Finalidade        string     `json:"finalidade,omitempty"`
	VersaoTermo       string     `json:"versaoTermo,omitempty"`
	DataConsentimento *time.Time `json:"dataConsentimento,omitempty"`
	DataRevogacao     *time.Time `json:"dataRevogacao,omitempty"`
	DataAtualizacao   time.Time  `json:"dataAtualizacao"`
}

type EventoResponse struct {
	ID        string    `json:"id"`
	ClienteID string    `json:"clienteId"`
	Tipo      string    `json:"tipo"`
	Origem    string    `json:"origem"`
	Timestamp time.Time `json:"timestamp"`
	Valor     float64   `json:"valor"`
	Categoria string    `json:"categoria,omitempty"`
}

type PerfilResponse struct {
	ID                 string  `json:"id"`
	ClienteID          string  `json:"clienteId"`
	FrequenciaCompras  int     `json:"frequenciaCompras"`
	TicketMedio        float64 `json:"ticketMedio"`
	CategoriaPreferida string  `json:"categoriaPreferida"`
	ScoreEngajamento   float64 `json:"scoreEngajamento"`
	SegmentoID         string  `json:"segmentoId,omitempty"`
}

type SegmentoResponse struct {
	ID        string `json:"id"`
	Nome      string `json:"nome"`
	Descricao string `json:"descricao,omitempty"`
	Criterio  string `json:"criterio"`
}

type RecomendacaoResponse struct {
	ID           string     `json:"id"`
	ProdutoID    string     `json:"produtoId"`
	Motivo       string     `json:"motivo"`
	Status       string     `json:"status"`
	DataGeracao  time.Time  `json:"dataGeracao"`
	DataAtivacao *time.Time `json:"dataAtivacao,omitempty"`
	Validade     *time.Time `json:"validade,omitempty"`
}

// ProcessamentoResponse é a resposta consolidada do passo 23 da sequência (RN-09).
type ProcessamentoResponse struct {
	Perfil       PerfilResponse        `json:"perfil"`
	Segmento     SegmentoResponse      `json:"segmento"`
	Recomendacao *RecomendacaoResponse `json:"recomendacao"` // null quando ErrSemProdutoElegivel
}

// ---- conversões ----

func clienteDTO(c entity.Cliente) ClienteResponse {
	return ClienteResponse{ID: c.ID.String(), Nome: c.Nome, Email: c.Email, DataCadastro: c.DataCadastro, Ativo: c.Ativo}
}

func consentimentoDTO(c entity.ConsentimentoLGPD) ConsentimentoResponse {
	r := ConsentimentoResponse{
		ClienteID: c.ClienteID.String(), Status: string(c.Status), Valido: c.EstaValido(),
		Finalidade: c.Finalidade, VersaoTermo: c.VersaoTermo, DataRevogacao: c.DataRevogacao,
		DataAtualizacao: c.DataAtualizacao,
	}
	if !c.DataConsentimento.IsZero() {
		d := c.DataConsentimento
		r.DataConsentimento = &d
	}
	return r
}

func eventoDTO(e entity.EventoComportamental) EventoResponse {
	return EventoResponse{ID: e.ID.String(), ClienteID: e.ClienteID.String(), Tipo: string(e.Tipo),
		Origem: e.Origem, Timestamp: e.Timestamp, Valor: e.Valor, Categoria: e.Categoria}
}

func perfilDTO(p entity.PerfilCliente) PerfilResponse {
	r := PerfilResponse{ID: p.ID.String(), ClienteID: p.ClienteID.String(), FrequenciaCompras: p.FrequenciaCompras,
		TicketMedio: p.TicketMedio, CategoriaPreferida: p.CategoriaPreferida, ScoreEngajamento: p.ScoreEngajamento}
	if p.SegmentoID != nil {
		r.SegmentoID = p.SegmentoID.String()
	}
	return r
}

func segmentoDTO(s entity.Segmento) SegmentoResponse {
	return SegmentoResponse{ID: s.ID.String(), Nome: s.Nome, Descricao: s.Descricao, Criterio: s.Criterio}
}

func recomendacaoDTO(r entity.RecomendacaoProduto) *RecomendacaoResponse {
	return &RecomendacaoResponse{ID: r.ID.String(), ProdutoID: r.ProdutoID.String(), Motivo: r.Motivo,
		Status: string(r.Status), DataGeracao: r.DataGeracao, DataAtivacao: r.DataAtivacao, Validade: r.Validade}
}
