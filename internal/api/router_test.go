package api

import (
	"context"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Spec: 3.7, Q-19 — /saude reflete o banco.
func TestSaudeReflete(t *testing.T) {
	casos := []struct {
		nome   string
		saude  func(context.Context) error
		status int
		corpo  string
	}{
		{"sem verificação (modo memória)", nil, http.StatusOK, `"ok"`},
		{"banco respondendo", func(context.Context) error { return nil }, http.StatusOK, `"ok"`},
		{"banco fora do ar", func(context.Context) error { return errors.New("conexão recusada") }, http.StatusServiceUnavailable, `"indisponivel"`},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			rec := httptest.NewRecorder()
			NovoRouter(Dependencias{AuthToken: "x", Saude: c.saude}).ServeHTTP(rec, httptest.NewRequest("GET", "/saude", nil))
			if rec.Code != c.status || !strings.Contains(rec.Body.String(), c.corpo) {
				t.Fatalf("status %d corpo %s; esperado %d com %s", rec.Code, rec.Body, c.status, c.corpo)
			}
			if strings.Contains(rec.Body.String(), "recusada") {
				t.Fatal("a causa interna não pode ir para a resposta")
			}
		})
	}
}

// Spec: Q-15 — resposta que não serializa vira 500, nunca 200 com corpo vazio.
func TestEscreverJSONFalhaVira500(t *testing.T) {
	rec := httptest.NewRecorder()
	escreverJSON(rec, http.StatusOK, map[string]float64{"ticketMedio": math.Inf(1)})
	if rec.Code != http.StatusInternalServerError || !strings.Contains(rec.Body.String(), "erro_interno") {
		t.Fatalf("status %d corpo %q", rec.Code, rec.Body)
	}
}

// Spec: 3.7, Q-19 — o corpo é um único objeto JSON.
func TestLerJSONUmUnicoObjeto(t *testing.T) {
	casos := map[string]bool{
		`{"nome":"A"}`:                  true,
		"{\"nome\":\"A\"}\n  ":          true,
		`{"nome":"A"}{"nome":"B"}`:      false,
		`{"nome":"A"} lixo`:             false,
		`{"nome":"A","desconhecido":1}`: false,
	}
	for corpo, valido := range casos {
		var destino CadastrarClienteRequest
		err := lerJSON(httptest.NewRequest("POST", "/", strings.NewReader(corpo)), &destino)
		if (err == nil) != valido {
			t.Errorf("%q: erro %v, esperado válido=%v", corpo, err, valido)
		}
	}
}
