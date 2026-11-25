package web

import (
	"net/http"
	"net/http/httptest"
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
