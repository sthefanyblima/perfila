package api

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	"perfila/internal/application/usecase"
	"perfila/internal/domain/entity"
	"perfila/internal/shared/errs"
)

// Dependencias reúne o que a API precisa; montada só na raiz de composição (C-08).
type Dependencias struct {
	AuthToken     string
	Precondicoes  usecase.Precondicoes
	Cadastrar     *usecase.CadastrarClienteUseCase
	Consentimento *usecase.GerenciarConsentimentoUseCase
	Coletar       *usecase.ColetarDadosUseCase
	GerarPerfil   *usecase.GerarPerfilUseCase
	Segmentar     *usecase.SegmentarClienteUseCase
	Recomendar    *usecase.GerarRecomendacaoUseCase
	Log           *slog.Logger
}

type handlers struct{ d Dependencias }

// NovoRouter monta as rotas da Spec 3.7 com seus middlewares.
func NovoRouter(d Dependencias) http.Handler {
	if d.Log == nil {
		d.Log = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	h := handlers{d: d}
	auth := func(f http.HandlerFunc) http.Handler { return Autenticacao(d.AuthToken, f) }
	authLGPD := func(f http.HandlerFunc) http.Handler {
		return Autenticacao(d.AuthToken, LGPD(d.Precondicoes, d.Log, f))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /saude", h.saude)
	mux.Handle("POST /v1/clientes", auth(h.cadastrarCliente))
	mux.Handle("POST /v1/clientes/{clienteID}/processamento", authLGPD(h.processar))
	mux.Handle("POST /v1/clientes/{clienteID}/eventos", authLGPD(h.registrarEvento))
	mux.Handle("PUT /v1/clientes/{clienteID}/consentimento", auth(h.concederConsentimento))
	mux.Handle("POST /v1/clientes/{clienteID}/consentimento/revogacao", auth(h.revogarConsentimento))
	mux.Handle("GET /v1/clientes/{clienteID}/consentimento", auth(h.consultarConsentimento))
	return Registro(d.Log, mux)
}

// GET /saude — usado pelo balanceador; sem autenticação.
func (h handlers) saude(w http.ResponseWriter, _ *http.Request) {
	escreverJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// POST /v1/clientes — UC-09.
func (h handlers) cadastrarCliente(w http.ResponseWriter, r *http.Request) {
	var req CadastrarClienteRequest
	if err := lerJSON(r, &req); err != nil {
		escreverErro(w, r, h.d.Log, err)
		return
	}
	c, err := h.d.Cadastrar.Executar(r.Context(), req.Nome, req.Email)
	if err != nil {
		escreverErro(w, r, h.d.Log, err)
		return
	}
	w.Header().Set("Location", "/v1/clientes/"+c.ID.String())
	escreverJSON(w, http.StatusCreated, clienteDTO(c))
}

// POST /v1/clientes/{clienteID}/processamento — RN-09: UC-02 («include» UC-01), UC-03 e UC-04,
// numa única resposta (sequência, passo 23). Erro interrompe; ErrSemProdutoElegivel não.
func (h handlers) processar(w http.ResponseWriter, r *http.Request) {
	id, err := clienteIDDaRota(r)
	if err != nil {
		escreverErro(w, r, h.d.Log, err)
		return
	}
	ctx := r.Context()
	perfil, err := h.d.GerarPerfil.Executar(ctx, id)
	if err != nil {
		escreverErro(w, r, h.d.Log, err)
		return
	}
	seg, err := h.d.Segmentar.Executar(ctx, id)
	if err != nil {
		escreverErro(w, r, h.d.Log, err)
		return
	}
	id2 := seg.ID
	perfil.SegmentoID = &id2
	resp := ProcessamentoResponse{Perfil: perfilDTO(perfil), Segmento: segmentoDTO(seg)}
	rec, err := h.d.Recomendar.Executar(ctx, id)
	switch {
	case err == nil:
		resp.Recomendacao = recomendacaoDTO(rec)
	case errors.Is(err, errs.ErrSemProdutoElegivel):
		resp.Recomendacao = nil
	default:
		escreverErro(w, r, h.d.Log, err)
		return
	}
	escreverJSON(w, http.StatusOK, resp)
}

// POST /v1/clientes/{clienteID}/eventos — UC-01 Registrar.
func (h handlers) registrarEvento(w http.ResponseWriter, r *http.Request) {
	id, err := clienteIDDaRota(r)
	if err != nil {
		escreverErro(w, r, h.d.Log, err)
		return
	}
	var req RegistrarEventoRequest
	if err := lerJSON(r, &req); err != nil {
		escreverErro(w, r, h.d.Log, err)
		return
	}
	ev, err := h.d.Coletar.Registrar(r.Context(), entity.EventoComportamental{
		ClienteID: id, Tipo: entity.TipoEvento(req.Tipo), Origem: req.Origem,
		Timestamp: req.Timestamp, Valor: req.Valor, Categoria: req.Categoria,
	})
	if err != nil {
		escreverErro(w, r, h.d.Log, err)
		return
	}
	escreverJSON(w, http.StatusCreated, eventoDTO(ev))
}

// PUT /v1/clientes/{clienteID}/consentimento — UC-08 Conceder.
func (h handlers) concederConsentimento(w http.ResponseWriter, r *http.Request) {
	id, err := clienteIDDaRota(r)
	if err != nil {
		escreverErro(w, r, h.d.Log, err)
		return
	}
	var req ConcederConsentimentoRequest
	if err := lerJSON(r, &req); err != nil {
		escreverErro(w, r, h.d.Log, err)
		return
	}
	if err := h.d.Consentimento.Conceder(r.Context(), id, req.Finalidade, req.VersaoTermo); err != nil {
		escreverErro(w, r, h.d.Log, err)
		return
	}
	h.responderConsentimento(w, r)
}

// POST /v1/clientes/{clienteID}/consentimento/revogacao — UC-08 Revogar.
func (h handlers) revogarConsentimento(w http.ResponseWriter, r *http.Request) {
	id, err := clienteIDDaRota(r)
	if err != nil {
		escreverErro(w, r, h.d.Log, err)
		return
	}
	if err := h.d.Consentimento.Revogar(r.Context(), id); err != nil {
		escreverErro(w, r, h.d.Log, err)
		return
	}
	h.responderConsentimento(w, r)
}

// GET /v1/clientes/{clienteID}/consentimento — UC-08 Consultar.
func (h handlers) consultarConsentimento(w http.ResponseWriter, r *http.Request) {
	h.responderConsentimento(w, r)
}

func (h handlers) responderConsentimento(w http.ResponseWriter, r *http.Request) {
	id, err := clienteIDDaRota(r)
	if err != nil {
		escreverErro(w, r, h.d.Log, err)
		return
	}
	c, err := h.d.Consentimento.Consultar(r.Context(), id)
	if err != nil {
		escreverErro(w, r, h.d.Log, err)
		return
	}
	escreverJSON(w, http.StatusOK, consentimentoDTO(c))
}
