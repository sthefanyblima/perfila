package api

import (
	"crypto/subtle"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"perfila/internal/application/usecase"
	"perfila/internal/shared/errs"
)

// Autenticacao exige "Authorization: Bearer <AUTH_TOKEN>"; ausente ou inválido → 401 (D-16).
func Autenticacao(token string, next http.Handler) http.Handler {
	esperado := []byte("Bearer " + token)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recebido := []byte(r.Header.Get("Authorization"))
		if subtle.ConstantTimeCompare(recebido, esperado) != 1 {
			w.Header().Set("WWW-Authenticate", `Bearer realm="perfila"`)
			escreverJSON(w, http.StatusUnauthorized, ErroResponse{Codigo: "nao_autenticado", Mensagem: "token ausente ou inválido"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// LGPD aplica as pré-condições RN-01 antes do handler: cliente existe (404), ativo (409),
// consentimento válido (403). Os casos de uso repetem a checagem (C-03).
func LGPD(pre usecase.Precondicoes, log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := clienteIDDaRota(r)
		if err != nil {
			escreverErro(w, r, log, err)
			return
		}
		if _, err := pre.Verificar(r.Context(), id); err != nil {
			escreverErro(w, r, log, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Registro loga método, caminho (que só carrega UUIDs), status e duração — nunca dados pessoais (C-09).
func Registro(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		rw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		log.InfoContext(r.Context(), "http", "metodo", r.Method, "caminho", r.URL.Path,
			"status", rw.status, "duracaoMs", time.Since(inicio).Milliseconds())
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func clienteIDDaRota(r *http.Request) (uuid.UUID, error) {
	bruto := strings.TrimSpace(r.PathValue("clienteID"))
	id, err := uuid.Parse(bruto)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: clienteID não é um UUID", errs.ErrDadosInvalidos)
	}
	return id, nil
}
