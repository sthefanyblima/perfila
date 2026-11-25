package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Spec: 3.7.1, 3.7.2 (cabeçalhos de segurança): o HSTS só vai quando o TLS termina num proxy.
func TestCabecalhosSegurancaHSTS(t *testing.T) {
	casos := []struct {
		nome, proto, esperado string
	}{
		{"http direto (desenvolvimento)", "", ""},
		{"atrás de proxy HTTPS", "https", "max-age=31536000"},
		{"proxy declarando http", "http", ""},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			h := cabecalhosSeguranca(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
			req := httptest.NewRequest(http.MethodGet, "/painel/entrar", nil)
			if c.proto != "" {
				req.Header.Set("X-Forwarded-Proto", c.proto)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if got := rec.Header().Get("Strict-Transport-Security"); got != c.esperado {
				t.Fatalf("HSTS = %q, esperado %q", got, c.esperado)
			}
		})
	}
}

// Spec: 3.7.1 — formulário com bytes fora de UTF-8 é recusado com 400, sem chegar à tela (no
// PostgreSQL daria erro 500).
func TestFormularioUTF8(t *testing.T) {
	casos := []struct {
		nome, tipo, corpo string
		metodo            string
		esperado          int
	}{
		{"valor válido com acento", "application/x-www-form-urlencoded", "nome=Demonstra%C3%A7%C3%A3o", http.MethodPost, http.StatusOK},
		{"valor em latin-1", "application/x-www-form-urlencoded", "nome=Demonstra%E7%E3o", http.MethodPost, http.StatusBadRequest},
		{"chave em latin-1", "application/x-www-form-urlencoded", "n%E3me=a", http.MethodPost, http.StatusBadRequest},
		{"GET não é conferido", "", "", http.MethodGet, http.StatusOK},
		{"multipart passa (validação própria)", "multipart/form-data; boundary=x", "--x--", http.MethodPost, http.StatusOK},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			h := cabecalhosSeguranca(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
			req := httptest.NewRequest(c.metodo, "/painel/usuarios", strings.NewReader(c.corpo))
			if c.tipo != "" {
				req.Header.Set("Content-Type", c.tipo)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != c.esperado {
				t.Fatalf("status %d, esperado %d", rec.Code, c.esperado)
			}
		})
	}
}
