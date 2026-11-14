package external

import (
	"context"
	"strings"
	"testing"
	"time"

	"perfila/internal/infra/memoria"
	"perfila/internal/infra/relogio"
)

// Spec: Q-57 — destinatário mascarado, texto de 03-textos.md com o link e limpeza das de 24 h.
func TestNotificadorSimulado(t *testing.T) {
	ctx := context.Background()
	agora := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	relogioFixo := relogio.NovoFixo(agora)
	caixa := memoria.NovaCaixaSaidaSimulada()
	n := NovoNotificadorSimulado(caixa, relogioFixo)
	link := "https://perfila.exemplo.test/titular/acesso?t=abc"
	cadastro := time.Date(2026, 3, 2, 15, 0, 0, 0, time.UTC)
	if err := n.EnviarLinkAcesso(ctx, "c-001@exemplo.test", link, agora.Add(30*time.Minute), cadastro); err != nil {
		t.Fatal(err)
	}
	msgs, _ := caixa.Listar(ctx, time.Time{})
	if len(msgs) != 1 {
		t.Fatalf("mensagens: %d", len(msgs))
	}
	m := msgs[0]
	if m.Para != "c***@exemplo.test" || strings.Contains(m.Corpo, "c-001@exemplo.test") {
		t.Errorf("destinatário não mascarado: %q / %q", m.Para, m.Corpo)
	}
	for _, quer := range []string{link, "02/03/2026", "14/11/2026 09:30", "ele só pode ser usado uma vez."} {
		if !strings.Contains(m.Corpo, quer) {
			t.Errorf("corpo sem %q: %q", quer, m.Corpo)
		}
	}
	relogioFixo.Avancar(25 * time.Hour)
	if err := n.EnviarLinkAcesso(ctx, "c-001@exemplo.test", link, relogioFixo.Agora().Add(30*time.Minute), cadastro); err != nil {
		t.Fatal(err)
	}
	if msgs, _ = caixa.Listar(ctx, time.Time{}); len(msgs) != 1 {
		t.Fatalf("a mensagem de mais de 24 h deveria ter sido apagada: %d", len(msgs))
	}
}

// Spec: Q-57, 03-textos.md (e-mail mascarado).
func TestMascararEmail(t *testing.T) {
	for e, quer := range map[string]string{"c-001@exemplo.test": "c***@exemplo.test", " Ana@Exemplo.TEST": "A***@exemplo.test",
		"semarroba": "***", "@exemplo.test": "***"} {
		if got := MascararEmail(e); got != quer {
			t.Errorf("MascararEmail(%q) = %q, esperado %q", e, got, quer)
		}
	}
}
