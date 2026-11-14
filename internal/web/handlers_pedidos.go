package web

import (
	"net/http"
)

// Pedidos do titular e caixa de saída simulada no painel. Spec: v1.10, 3.7.1, Q-57, Q-58.
// Hoje são da pessoa operadora única (Q-22); com os papéis da v1.5, ficam com ADMINISTRADOR.

func (p *painel) rotasPedidos(mux *http.ServeMux, sessao func(http.HandlerFunc) http.Handler) {
	mux.Handle("GET /painel/pedidos", sessao(p.telaPedidos))
	mux.Handle("GET /painel/operacao/caixa-saida", sessao(p.telaCaixaSaida))
}

// GET /painel/pedidos — abertos primeiro; cada linha leva à aba Direitos da ficha.
func (p *painel) telaPedidos(w http.ResponseWriter, r *http.Request) {
	if p.d.Pedidos == nil {
		p.naoEncontrada(w, r)
		return
	}
	lista, err := p.d.Pedidos.Executar(r.Context())
	if err != nil {
		p.falha(w, r, err, "")
		return
	}
	p.render(w, r, http.StatusOK, "pedidos", pagina{Titulo: "Pedidos do titular", Nav: "pedidos", Dados: lista})
}

// GET /painel/operacao/caixa-saida — só com o portal ligado; o link é uma credencial (Q-57).
func (p *painel) telaCaixaSaida(w http.ResponseWriter, r *http.Request) {
	if p.d.CaixaSaida == nil {
		p.naoEncontrada(w, r)
		return
	}
	msgs, err := p.d.CaixaSaida.Executar(r.Context())
	if err != nil {
		p.falha(w, r, err, "")
		return
	}
	p.render(w, r, http.StatusOK, "caixa", pagina{Titulo: "Caixa de saída (simulada)", Nav: "caixa", Dados: msgs})
}
