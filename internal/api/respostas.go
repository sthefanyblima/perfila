package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"perfila/internal/shared/errs"
)

// statusPorCodigo é o mapeamento da Spec 3.5 (erro de domínio → HTTP).
func statusPorCodigo(codigo string) int {
	switch codigo {
	case errs.ErrClienteNaoEncontrado.Codigo:
		return http.StatusNotFound
	case errs.ErrClienteInativo.Codigo, errs.ErrTransicaoInvalida.Codigo, errs.ErrPerfilNaoEncontrado.Codigo:
		return http.StatusConflict
	case errs.ErrConsentimentoInvalido.Codigo:
		return http.StatusForbidden
	case errs.ErrDadosInvalidos.Codigo:
		return http.StatusUnprocessableEntity
	}
	return http.StatusInternalServerError
}

func escreverJSON(w http.ResponseWriter, status int, corpo any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(corpo)
}

// escreverErro responde {codigo, mensagem}; erros inesperados viram 500 sem detalhes internos.
func escreverErro(w http.ResponseWriter, r *http.Request, log *slog.Logger, err error) {
	var e *errs.Erro
	if errors.As(err, &e) {
		status := statusPorCodigo(e.Codigo)
		if status != http.StatusInternalServerError {
			escreverJSON(w, status, ErroResponse{Codigo: e.Codigo, Mensagem: e.Mensagem})
			return
		}
	}
	log.ErrorContext(r.Context(), "erro_interno", "caminho", r.URL.Path, "erro", err.Error())
	escreverJSON(w, http.StatusInternalServerError, ErroResponse{Codigo: "erro_interno", Mensagem: "erro inesperado"})
}

// lerJSON decodifica o corpo; JSON malformado ou com campos desconhecidos é ErrDadosInvalidos.
func lerJSON(r *http.Request, destino any) error {
	dec := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(destino); err != nil {
		return fmt.Errorf("%w: corpo JSON inválido", errs.ErrDadosInvalidos)
	}
	return nil
}
