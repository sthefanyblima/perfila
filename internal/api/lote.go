package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"perfila/internal/application/usecase"
	"perfila/internal/shared/errs"
)

// EventoLoteRequest é uma linha do lote; timestamp em texto RFC 3339 e valor numérico (RN-16).
type EventoLoteRequest struct {
	ClienteID string      `json:"clienteId"`
	Tipo      string      `json:"tipo"`
	Origem    string      `json:"origem"`
	Timestamp string      `json:"timestamp"`
	Valor     json.Number `json:"valor"`
	Categoria string      `json:"categoria"`
}

type LoteEventosRequest struct {
	Eventos []EventoLoteRequest `json:"eventos"`
}

type LinhaLoteResponse struct {
	Numero    int    `json:"numero"`
	ClienteID string `json:"clienteId,omitempty"`
	Situacao  string `json:"situacao"`
	Codigo    string `json:"codigo,omitempty"`
	Motivo    string `json:"motivo,omitempty"`
}

type LoteEventosResponse struct {
	Gravado    bool                `json:"gravado"`
	Aceitas    int                 `json:"aceitas"`
	Duplicadas int                 `json:"duplicadas"`
	Rejeitadas int                 `json:"rejeitadas"`
	Linhas     []LinhaLoteResponse `json:"linhas"`
}

// POST /v1/eventos/lote — ImportarEventos (RN-16, Q-47). Sem cliente na rota, o middleware LGPD
// não se aplica: o caso de uso confere RN-01 em cada linha (C-03). ?previa=true só confere.
func (h handlers) importarLote(w http.ResponseWriter, r *http.Request) {
	var req LoteEventosRequest
	if err := lerJSON(r, &req); err != nil {
		escreverErro(w, r, h.d.Log, err)
		return
	}
	entradas := make([]usecase.EntradaEvento, len(req.Eventos))
	for i, e := range req.Eventos {
		entradas[i] = usecase.EntradaEvento{Numero: i + 1, ClienteID: e.ClienteID, Tipo: e.Tipo, Origem: e.Origem,
			Timestamp: e.Timestamp, Valor: e.Valor.String(), Categoria: e.Categoria}
	}
	executar := h.d.Importar.Importar
	if r.URL.Query().Get("previa") == "true" {
		executar = h.d.Importar.Previa
	}
	res, err := executar(r.Context(), entradas)
	if err != nil {
		escreverErro(w, r, h.d.Log, err)
		return
	}
	// Auditoria do lote: só quantidades, nunca IDs de cliente (C-09, RN-16).
	h.d.Log.LogAttrs(r.Context(), slog.LevelInfo, "importacao_eventos", slog.Bool("gravado", res.Gravado),
		slog.Int("aceitas", res.Aceitas), slog.Int("duplicadas", res.Duplicadas), slog.Int("rejeitadas", res.Rejeitadas))
	resp := LoteEventosResponse{Gravado: res.Gravado, Aceitas: res.Aceitas, Duplicadas: res.Duplicadas,
		Rejeitadas: res.Rejeitadas, Linhas: []LinhaLoteResponse{}}
	for _, l := range res.Linhas {
		linha := LinhaLoteResponse{Numero: l.Numero, Situacao: string(l.Situacao), Codigo: errs.Codigo(l.Erro), Motivo: l.Motivo}
		if l.ClienteID != uuid.Nil {
			linha.ClienteID = l.ClienteID.String()
		}
		resp.Linhas = append(resp.Linhas, linha)
	}
	escreverJSON(w, http.StatusOK, resp)
}
