package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	// Serializa antes de escrever o status: uma falha vira 500, nunca 200 com corpo vazio (Q-15).
	b, err := json.Marshal(corpo)
	if err != nil {
		status = http.StatusInternalServerError
		b = []byte(`{"codigo":"erro_interno","mensagem":"erro inesperado"}`)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(append(b, '\n'))
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
	// Um único objeto: conteúdo depois dele torna o corpo ambíguo (Q-19).
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: conteúdo depois do objeto JSON", errs.ErrDadosInvalidos)
	}
	return nil
}
