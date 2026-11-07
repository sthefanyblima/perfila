package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"perfila/internal/infra/config"
)

// Teste de fumaça ponta a ponta do binário montado (T-23): cadastro → consentimento →
// eventos → processamento, pelo HTTP real.
func TestFluxoPontaAPonta(t *testing.T) {
	var logs bytes.Buffer
	app, err := montar(config.Config{AuthToken: "segredo", RecomendacaoValidadeDias: 30, EngajamentoMetaEventos: 10, LogLevel: "info"}, &logs)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(app.handler)
	defer srv.Close()

	chamar := func(metodo, caminho, corpo string, quer int) map[string]any {
		t.Helper()
		req, _ := http.NewRequest(metodo, srv.URL+caminho, strings.NewReader(corpo))
		req.Header.Set("Authorization", "Bearer segredo")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != quer {
			t.Fatalf("%s %s: status %d, esperado %d: %s", metodo, caminho, resp.StatusCode, quer, b)
		}
		var out map[string]any
		_ = json.Unmarshal(b, &out)
		return out
	}

	c := chamar("POST", "/v1/clientes", `{"nome":"Cliente Sintético","email":"c-001@exemplo.test"}`, 201)
	id := c["id"].(string)
	base := "/v1/clientes/" + id
	chamar("POST", base+"/processamento", "", 403)
	chamar("PUT", base+"/consentimento", `{"finalidade":"segmentacao","versaoTermo":"v1"}`, 200)
	quando := time.Now().UTC().Add(-time.Hour).Format(time.RFC3339)
	for _, v := range []string{"520", "610", "700"} {
		chamar("POST", base+"/eventos", `{"tipo":"compra","origem":"ecommerce","valor":`+v+`,"categoria":"livros","timestamp":"`+quando+`"}`, 201)
	}
	r := chamar("POST", base+"/processamento", "", 200)
	if seg := r["segmento"].(map[string]any)["nome"]; seg != "premium" {
		t.Fatalf("segmento %v, esperado premium", seg)
	}
	if rec, ok := r["recomendacao"].(map[string]any); !ok || rec["status"] != "ATIVA" {
		t.Fatalf("recomendação inesperada: %v", r["recomendacao"])
	}
	chamar("POST", base+"/consentimento/revogacao", "", 200)
	chamar("POST", base+"/processamento", "", 403)

	// C-09: nenhum dado pessoal nos logs.
	if strings.Contains(logs.String(), "Sintético") || strings.Contains(logs.String(), "exemplo.test") {
		t.Fatal("log contém dado pessoal")
	}
}

func TestSaudeSemToken(t *testing.T) {
	app, _ := montar(config.Config{AuthToken: "x", LogLevel: "info", RecomendacaoValidadeDias: 30, EngajamentoMetaEventos: 30}, io.Discard)
	rec := httptest.NewRecorder()
	app.handler.ServeHTTP(rec, httptest.NewRequest("GET", "/saude", nil))
	if rec.Code != 200 {
		t.Fatalf("saude: %d", rec.Code)
	}
}

func TestArgumentosEConfig(t *testing.T) {
	semEnv := func(string) string { return "" }
	if err := executar(nil, semEnv, io.Discard); err == nil || !strings.Contains(err.Error(), "uso:") {
		t.Fatalf("sem subcomando: %v", err)
	}
	if err := executar([]string{"serve"}, semEnv, io.Discard); err == nil || !strings.Contains(err.Error(), "AUTH_TOKEN") {
		t.Fatalf("sem AUTH_TOKEN: %v", err)
	}
	comBanco := func(k string) string {
		return map[string]string{"AUTH_TOKEN": "x", "DATABASE_URL": "postgres://db"}[k]
	}
	if err := executar([]string{"serve"}, comBanco, io.Discard); err == nil || !strings.Contains(err.Error(), "Generic") {
		t.Fatalf("PostgreSQL ainda não suportado: %v", err)
	}
	soToken := func(k string) string { return map[string]string{"AUTH_TOKEN": "x"}[k] }
	if err := executar([]string{"expirar-recomendacoes"}, soToken, io.Discard); err != nil {
		t.Fatalf("expirar-recomendacoes: %v", err)
	}
}
