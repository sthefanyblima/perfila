package entity

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"perfila/internal/shared/errs"
)

// Spec: 3.2 RN-10 Q-34 — o movimento copia finalidade, termo e data do consentimento.
func TestNovoMovimento(t *testing.T) {
	agora := time.Date(2026, 3, 10, 12, 0, 0, 0, time.FixedZone("BRT", -3*3600))
	c := NovoConsentimento(uuid.New(), uuid.New(), FinalidadeSegmentacao, agora)
	if err := c.Conceder(FinalidadeSegmentacao, "v1", agora); err != nil {
		t.Fatal(err)
	}
	casos := []struct {
		nome     string
		tipo     TipoMovimento
		origem   OrigemMovimento
		operador string
		erro     error
	}{
		{"api", MovimentoConcessao, OrigemAPI, "", nil},
		{"painel com operador", MovimentoRevogacao, OrigemPainel, " op-01 ", nil},
		{"portal", MovimentoConcessao, OrigemPortal, "", nil},
		{"sistema", MovimentoConcessao, OrigemSistema, "", nil},
		{"migracao", MovimentoRevogacao, OrigemMigracao, "", nil},
		{"painel sem operador", MovimentoConcessao, OrigemPainel, "  ", errs.ErrDadosInvalidos},
		{"tipo desconhecido", TipoMovimento("ALTERACAO"), OrigemAPI, "", errs.ErrDadosInvalidos},
		{"origem desconhecida", MovimentoConcessao, OrigemMovimento("email"), "", errs.ErrDadosInvalidos},
	}
	for _, tc := range casos {
		t.Run(tc.nome, func(t *testing.T) {
			id := uuid.New()
			m, err := NovoMovimento(id, c, tc.tipo, tc.origem, tc.operador)
			if !errors.Is(err, tc.erro) {
				t.Fatalf("erro %v, esperado %v", err, tc.erro)
			}
			if tc.erro != nil {
				return
			}
			esperado := MovimentoConsentimento{ID: id, ClienteID: c.ClienteID, Tipo: tc.tipo, Finalidade: "segmentacao",
				VersaoTermo: "v1", Data: agora.UTC(), Origem: tc.origem}
			if tc.origem == OrigemPainel {
				esperado.Operador = "op-01"
			}
			if m != esperado {
				t.Fatalf("movimento\n obtido   %+v\n esperado %+v", m, esperado)
			}
		})
	}
}
