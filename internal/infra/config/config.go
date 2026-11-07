// Package config lê a configuração do omega-server (Spec 3.8): variáveis de ambiente e,
// opcionalmente, um arquivo config.env (KEY=VALUE). O ambiente prevalece sobre o arquivo.
//
// Spec: T-20, 3.8, D-23.
package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config é a configuração validada do servidor.
type Config struct {
	HTTPPort                 int
	DatabaseURL              string // vazia = repositórios em memória (D-23, até a fase Generic)
	AuthToken                string
	RecomendacaoValidadeDias int
	EngajamentoMetaEventos   int
	LogLevel                 string
}

// ModoMemoria informa se o servidor roda sem PostgreSQL (D-23).
func (c Config) ModoMemoria() bool { return c.DatabaseURL == "" }

// Carregar lê o arquivo (se path não for vazio) e depois o ambiente via getenv.
// Variável obrigatória ausente ou inválida aborta com mensagem clara.
func Carregar(path string, getenv func(string) string) (Config, error) {
	arquivo := map[string]string{}
	if path != "" {
		var err error
		if arquivo, err = lerArquivo(path); err != nil {
			return Config{}, err
		}
	}
	valor := func(k string) string {
		if v := strings.TrimSpace(getenv(k)); v != "" {
			return v
		}
		return arquivo[k]
	}

	var problemas []string
	inteiro := func(k string, padrao, minimo int) int {
		v := valor(k)
		if v == "" {
			return padrao
		}
		n, err := strconv.Atoi(v)
		if err != nil || n < minimo {
			problemas = append(problemas, fmt.Sprintf("%s deve ser inteiro >= %d (recebido %q)", k, minimo, v))
		}
		return n
	}

	c := Config{
		HTTPPort:                 inteiro("HTTP_PORT", 8080, 1),
		DatabaseURL:              valor("DATABASE_URL"),
		AuthToken:                valor("AUTH_TOKEN"),
		RecomendacaoValidadeDias: inteiro("RECOMENDACAO_VALIDADE_DIAS", 30, 1),
		EngajamentoMetaEventos:   inteiro("ENGAJAMENTO_META_EVENTOS", 30, 1),
		LogLevel:                 strings.ToLower(valor("LOG_LEVEL")),
	}
	if c.LogLevel == "" {
		c.LogLevel = "info"
	}
	if c.AuthToken == "" {
		problemas = append(problemas, "AUTH_TOKEN é obrigatória")
	}
	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		problemas = append(problemas, fmt.Sprintf("LOG_LEVEL inválido: %q (use debug, info, warn ou error)", c.LogLevel))
	}
	if len(problemas) > 0 {
		return Config{}, errors.New("configuração inválida: " + strings.Join(problemas, "; "))
	}
	return c, nil
}

func lerArquivo(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("lendo %s: %w", path, err)
	}
	defer f.Close()
	out := map[string]string{}
	sc := bufio.NewScanner(f)
	for n := 1; sc.Scan(); n++ {
		linha := strings.TrimSpace(sc.Text())
		if linha == "" || strings.HasPrefix(linha, "#") {
			continue
		}
		k, v, ok := strings.Cut(linha, "=")
		if !ok {
			return nil, fmt.Errorf("%s:%d: esperado CHAVE=VALOR", path, n)
		}
		out[strings.TrimSpace(k)] = strings.Trim(strings.TrimSpace(v), `"'`)
	}
	return out, sc.Err()
}
