// Package config lê a configuração do omega-server (Spec 3.8): variáveis de ambiente e,
// opcionalmente, um arquivo config.env (KEY=VALUE). O ambiente prevalece sobre o arquivo.
//
// Spec: T-20, 3.8, D-23.
package config

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
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
	RetencaoDiasRevogado     int // RN-18, Q-51: prazo da retenção após a revogação, em dias

	// Painel web do operador (Spec 3.8, Q-22, Q-25, Q-29).
	PainelUsuario     string
	PainelSenhaHash   string // bcrypt; vazia = painel desligado
	PainelChaveSessao []byte // 32 bytes (64 dígitos hexadecimais); vazia = painel desligado
	PainelFinalidade  string
	PainelVersaoTermo string

	// Portal do titular (Spec 3.8, v1.10, Q-60).
	PortalURLBase     string // endereço público usado no link; vazia = portal desligado
	PortalChaveSessao []byte // 32 bytes, diferente da do painel; vazia = portal desligado
}

// PortalLigado informa se o portal /titular deve ser montado (Q-60).
func (c Config) PortalLigado() bool { return c.PortalURLBase != "" && len(c.PortalChaveSessao) > 0 }

// PainelLigado informa se o painel /painel deve ser montado (Q-29).
func (c Config) PainelLigado() bool { return c.PainelSenhaHash != "" && len(c.PainelChaveSessao) > 0 }

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
		RetencaoDiasRevogado:     inteiro("RETENCAO_DIAS_REVOGADO", 180, 1),
		LogLevel:                 strings.ToLower(valor("LOG_LEVEL")),
	}
	if c.LogLevel == "" {
		c.LogLevel = "info"
	}
	problemas = append(problemas, c.lerPainel(valor)...)
	problemas = append(problemas, c.lerPortal(valor)...)
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

// lerPainel lê PAINEL_*: hash e chave andam juntos (Q-29); a chave tem 32 bytes em hexadecimal.
// Os valores nunca entram nas mensagens de erro.
func (c *Config) lerPainel(valor func(string) string) []string {
	padrao := func(k, p string) string {
		if v := valor(k); v != "" {
			return v
		}
		return p
	}
	c.PainelUsuario = padrao("PAINEL_USUARIO", "operador")
	c.PainelFinalidade = padrao("PAINEL_FINALIDADE", "segmentacao")
	c.PainelVersaoTermo = padrao("PAINEL_VERSAO_TERMO", "v1")
	c.PainelSenhaHash = valor("PAINEL_SENHA_HASH")
	chave := valor("PAINEL_CHAVE_SESSAO")
	var problemas []string
	if (c.PainelSenhaHash == "") != (chave == "") {
		return append(problemas, "PAINEL_SENHA_HASH e PAINEL_CHAVE_SESSAO devem ser definidas juntas (ou nenhuma, painel desligado)")
	}
	if chave != "" {
		b, err := hex.DecodeString(chave)
		if err != nil || len(b) != 32 {
			problemas = append(problemas, "PAINEL_CHAVE_SESSAO deve ter 64 dígitos hexadecimais (32 bytes)")
		}
		c.PainelChaveSessao = b
	}
	if c.PainelSenhaHash != "" && !strings.HasPrefix(c.PainelSenhaHash, "$2") {
		problemas = append(problemas, "PAINEL_SENHA_HASH deve ser um hash bcrypt (gere com: omega-server hash-senha)")
	}
	return problemas
}

// lerPortal lê PORTAL_*: endereço e chave andam juntos (Q-60); o endereço é http(s), sem barra no
// fim; a chave tem 32 bytes em hexadecimal e não pode ser a do painel. Valores fora das mensagens.
func (c *Config) lerPortal(valor func(string) string) []string {
	base, chave := strings.TrimSpace(valor("PORTAL_URL_BASE")), valor("PORTAL_CHAVE_SESSAO")
	if (base == "") != (chave == "") {
		return []string{"PORTAL_URL_BASE e PORTAL_CHAVE_SESSAO devem ser definidas juntas (ou nenhuma, portal desligado)"}
	}
	if base == "" {
		return nil
	}
	var problemas []string
	if u, err := url.Parse(base); err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" ||
		strings.HasSuffix(base, "/") || u.RawQuery != "" || u.Fragment != "" {
		problemas = append(problemas, "PORTAL_URL_BASE deve ser um endereço http(s) sem barra no fim (por exemplo https://localhost:8443)")
	}
	b, err := hex.DecodeString(chave)
	if err != nil || len(b) != 32 {
		problemas = append(problemas, "PORTAL_CHAVE_SESSAO deve ter 64 dígitos hexadecimais (32 bytes)")
	} else if bytes.Equal(b, c.PainelChaveSessao) {
		problemas = append(problemas, "PORTAL_CHAVE_SESSAO deve ser diferente de PAINEL_CHAVE_SESSAO")
	}
	c.PortalURLBase, c.PortalChaveSessao = base, b
	return problemas
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
