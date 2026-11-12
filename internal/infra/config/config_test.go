package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func env(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

// Spec: 3.8 — padrões e obrigatórias.
func TestCarregarPadroes(t *testing.T) {
	c, err := Carregar("", env(map[string]string{"AUTH_TOKEN": "t"}))
	if err != nil {
		t.Fatal(err)
	}
	if c.HTTPPort != 8080 || c.RecomendacaoValidadeDias != 30 || c.EngajamentoMetaEventos != 30 || c.LogLevel != "info" || !c.ModoMemoria() {
		t.Fatalf("padrões errados: %+v", c)
	}
}

func TestCarregarErros(t *testing.T) {
	casos := []struct {
		env  map[string]string
		quer string
	}{
		{map[string]string{}, "AUTH_TOKEN"},
		{map[string]string{"AUTH_TOKEN": "t", "HTTP_PORT": "x"}, "HTTP_PORT"},
		{map[string]string{"AUTH_TOKEN": "t", "ENGAJAMENTO_META_EVENTOS": "0"}, "ENGAJAMENTO_META_EVENTOS"},
		{map[string]string{"AUTH_TOKEN": "t", "LOG_LEVEL": "verbose"}, "LOG_LEVEL"},
	}
	for _, c := range casos {
		if _, err := Carregar("", env(c.env)); err == nil || !strings.Contains(err.Error(), c.quer) {
			t.Errorf("%v: esperava erro citando %s, obtido %v", c.env, c.quer, err)
		}
	}
}

// O ambiente prevalece sobre o arquivo.
func TestCarregarArquivo(t *testing.T) {
	p := filepath.Join(t.TempDir(), "config.env")
	conteudo := "# comentário\nAUTH_TOKEN=do-arquivo\nHTTP_PORT=9090\nDATABASE_URL=\"postgres://x\"\n"
	if err := os.WriteFile(p, []byte(conteudo), 0o600); err != nil {
		t.Fatal(err)
	}
	c, err := Carregar(p, env(map[string]string{"HTTP_PORT": "7070"}))
	if err != nil {
		t.Fatal(err)
	}
	if c.AuthToken != "do-arquivo" || c.HTTPPort != 7070 || c.DatabaseURL != "postgres://x" || c.ModoMemoria() {
		t.Fatalf("leitura do arquivo: %+v", c)
	}
	if err := os.WriteFile(p, []byte("SEM_IGUAL\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Carregar(p, env(nil)); err == nil {
		t.Fatal("linha sem '=' deveria falhar")
	}
}

// Spec: 3.8, Q-29 — painel desligado sem PAINEL_*; hash e chave juntos; chave de 32 bytes.
func TestCarregarPainel(t *testing.T) {
	chave := strings.Repeat("ab", 32)
	c, err := Carregar("", env(map[string]string{"AUTH_TOKEN": "t"}))
	if err != nil || c.PainelLigado() || c.PainelUsuario != "operador" || c.PainelFinalidade != "segmentacao" || c.PainelVersaoTermo != "v1" {
		t.Fatalf("padrões do painel: %+v (%v)", c, err)
	}
	c, err = Carregar("", env(map[string]string{"AUTH_TOKEN": "t", "PAINEL_SENHA_HASH": "$2a$10$x", "PAINEL_CHAVE_SESSAO": chave}))
	if err != nil || !c.PainelLigado() || len(c.PainelChaveSessao) != 32 {
		t.Fatalf("painel ligado: %+v (%v)", c, err)
	}
	casos := []map[string]string{
		{"AUTH_TOKEN": "t", "PAINEL_SENHA_HASH": "$2a$10$x"},
		{"AUTH_TOKEN": "t", "PAINEL_CHAVE_SESSAO": chave},
		{"AUTH_TOKEN": "t", "PAINEL_SENHA_HASH": "$2a$10$x", "PAINEL_CHAVE_SESSAO": "curta"},
		{"AUTH_TOKEN": "t", "PAINEL_SENHA_HASH": "senha-em-claro", "PAINEL_CHAVE_SESSAO": chave},
	}
	for _, e := range casos {
		_, err := Carregar("", env(e))
		if err == nil || !strings.Contains(err.Error(), "PAINEL_") || strings.Contains(err.Error(), "senha-em-claro") {
			t.Errorf("%v: esperava erro de PAINEL_* sem o valor, obtido %v", e, err)
		}
	}
}
