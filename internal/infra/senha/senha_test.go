package senha

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// Spec: Q-36 — gera um hash bcrypt que confere só com a mesma senha.
func TestBcrypt(t *testing.T) {
	for _, b := range []Bcrypt{NovoBcrypt(bcrypt.MinCost), {}} {
		h, err := b.Gerar("senha-sintetica-1")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(h, "$2") || strings.Contains(h, "senha-sintetica-1") {
			t.Fatalf("hash inesperado: %q", h)
		}
		if !b.Conferir(h, "senha-sintetica-1") || b.Conferir(h, "senha-sintetica-2") || b.Conferir("lixo", "x") {
			t.Fatal("conferência errada")
		}
	}
	if c, _ := bcrypt.Cost([]byte(mustHash(t, Bcrypt{}))); c != Custo {
		t.Fatalf("custo padrão %d, esperado %d", c, Custo)
	}
}

func mustHash(t *testing.T, b Bcrypt) string {
	h, err := b.Gerar("senha-sintetica-1")
	if err != nil {
		t.Fatal(err)
	}
	return h
}
