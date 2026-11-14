package entity

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"perfila/internal/shared/errs"
)

// Spec: RN-17 Q-49 — normalização antes do hash.
func TestHashEmail(t *testing.T) {
	const esperado = "9c4cd21419f4d9f66ea792856912326980a93ed91c0f9b487dc9df386b13a7ef"
	for _, e := range []string{"c-001@exemplo.test", "  C-001@Exemplo.TEST ", "c-0 01@exemplo.test\t"} {
		if got := HashEmail(e); got != esperado {
			t.Errorf("HashEmail(%q) = %s", e, got)
		}
	}
	if HashEmail("c-002@exemplo.test") == esperado {
		t.Error("e-mails diferentes com o mesmo hash")
	}
}

// Spec: 3.4.3 — todas as combinações estado × evento do membro do público.
func TestMembroPublicoTransicoes(t *testing.T) {
	if _, err := NovoMembro(uuid.New(), uuid.New(), uuid.New(), uuid.New(), " ", agoraTeste); !errors.Is(err, errs.ErrDadosInvalidos) {
		t.Fatalf("membro sem hash: %v", err)
	}
	for _, e := range []StatusMembro{MembroExportado, MembroRemocaoPendente, MembroRemovido} {
		m := MembroPublico{Status: e}
		err := m.MarcarRemocao(agoraTeste)
		if (err == nil) != (e == MembroExportado) || (err != nil && !errors.Is(err, errs.ErrTransicaoInvalida)) {
			t.Errorf("MarcarRemocao a partir de %s: %v", e, err)
		}
		m = MembroPublico{Status: e}
		err = m.Remover(uuid.New(), agoraTeste)
		if (err == nil) != (e == MembroRemocaoPendente) || (err != nil && !errors.Is(err, errs.ErrTransicaoInvalida)) {
			t.Errorf("Remover a partir de %s: %v", e, err)
		}
	}
	m, err := NovoMembro(uuid.New(), uuid.New(), uuid.New(), uuid.New(), HashEmail("c-001@exemplo.test"), agoraTeste)
	if err != nil || m.Status != MembroExportado {
		t.Fatalf("novo membro: %v %s", err, m.Status)
	}
	if err := m.MarcarRemocao(agoraTeste); err != nil || m.DataRemocaoPendente == nil {
		t.Fatalf("marcar remoção: %v", err)
	}
	if err := m.Remover(uuid.Nil, agoraTeste); !errors.Is(err, errs.ErrDadosInvalidos) || m.Status != MembroRemocaoPendente {
		t.Fatalf("remover sem exportação: %v %s", err, m.Status)
	}
}

// Spec: 3.4.1 Q-48 — a guarda confere a finalidade do registro.
func TestConsentimentoPorFinalidade(t *testing.T) {
	c := NovoConsentimento(uuid.New(), uuid.New(), FinalidadePublicidadeTerceiros, agoraTeste)
	for _, f := range []Finalidade{"", "marketing", "Segmentacao", FinalidadeSegmentacao} {
		if err := c.Conceder(f, "v1", agoraTeste); !errors.Is(err, errs.ErrDadosInvalidos) || c.Status != StatusNaoConcedido {
			t.Errorf("Conceder %q num registro de publicidade: %v", f, err)
		}
	}
	if err := c.Conceder(FinalidadePublicidadeTerceiros, " ", agoraTeste); !errors.Is(err, errs.ErrDadosInvalidos) {
		t.Errorf("termo vazio: %v", err)
	}
	if err := c.Conceder(FinalidadePublicidadeTerceiros, "v1", agoraTeste); err != nil || !c.EstaValido() {
		t.Fatalf("conceder publicidade: %v", err)
	}
	if len(Finalidades()) != 2 || !Finalidades()[0].Valida() || Finalidade("x").Valida() {
		t.Error("enumerado de finalidades")
	}
	if got := (MovimentoConsentimento{Finalidade: "uso interno"}).FinalidadeEfetiva(); got != FinalidadeSegmentacao {
		t.Errorf("movimento antigo: %s", got)
	}
	if got := (MovimentoConsentimento{Finalidade: "publicidade_terceiros"}).FinalidadeEfetiva(); got != FinalidadePublicidadeTerceiros {
		t.Errorf("movimento de publicidade: %s", got)
	}
}
