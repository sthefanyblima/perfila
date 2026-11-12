package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"perfila/internal/infra/config"
)

// Teste de fumaça ponta a ponta do binário montado (T-23): cadastro → consentimento →
// eventos → processamento, pelo HTTP real. Roda no modo memória e, se
// PERFILA_TEST_DATABASE_URL estiver definida, também sobre o PostgreSQL (T-19).
func TestFluxoPontaAPonta(t *testing.T) {
	t.Run("memoria", func(t *testing.T) { fluxoPontaAPonta(t, "") })
	t.Run("postgres", func(t *testing.T) {
		url := os.Getenv("PERFILA_TEST_DATABASE_URL")
		if url == "" {
			t.Skip("PERFILA_TEST_DATABASE_URL não definida")
		}
		fluxoPontaAPonta(t, url)
	})
}

func fluxoPontaAPonta(t *testing.T, databaseURL string) {
	var logs bytes.Buffer
	cfg := config.Config{DatabaseURL: databaseURL, AuthToken: "segredo", RecomendacaoValidadeDias: 30, EngajamentoMetaEventos: 10, LogLevel: "info"}
	app, err := montar(context.Background(), cfg, &logs)
	if err != nil {
		t.Fatal(err)
	}
	defer app.fechar()
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
	// C-09: a query string nunca vai para o log (o teste de dados pessoais no fim confere).
	chamar("GET", base+"/consentimento?email=c-001@exemplo.test", "", 200)
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
	app, _ := montar(context.Background(), config.Config{AuthToken: "x", LogLevel: "info", RecomendacaoValidadeDias: 30, EngajamentoMetaEventos: 30}, io.Discard)
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
	// Banco inacessível aborta a inicialização sem expor a senha da URL.
	bancoFora := func(k string) string {
		return map[string]string{"AUTH_TOKEN": "x",
			"DATABASE_URL": "postgres://omega:senha-secreta@127.0.0.1:1/omega_db?connect_timeout=2"}[k]
	}
	err := executar([]string{"serve"}, bancoFora, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "pg: conectando") || strings.Contains(err.Error(), "senha-secreta") {
		t.Fatalf("banco inacessível: %v", err)
	}
	soToken := func(k string) string { return map[string]string{"AUTH_TOKEN": "x"}[k] }
	if err := executar([]string{"expirar-recomendacoes"}, soToken, io.Discard); err != nil {
		t.Fatalf("expirar-recomendacoes: %v", err)
	}
}

// Spec: Q-28 — hash-senha imprime um bcrypt que confere com a senha e nunca a senha.
func TestHashSenha(t *testing.T) {
	var saida strings.Builder
	if err := hashSenha(strings.NewReader("senha-sintetica-1\n"), &saida); err != nil {
		t.Fatal(err)
	}
	h := strings.TrimSpace(saida.String())
	if strings.Contains(h, "senha-sintetica-1") || bcrypt.CompareHashAndPassword([]byte(h), []byte("senha-sintetica-1")) != nil {
		t.Fatalf("hash inválido: %q", h)
	}
	if err := hashSenha(strings.NewReader("curta\n"), io.Discard); err == nil {
		t.Fatal("senha curta deveria falhar")
	}
}
