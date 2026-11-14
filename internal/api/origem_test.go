package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"perfila/internal/api"
	"perfila/internal/application/usecase"
	"perfila/internal/domain/entity"
	"perfila/internal/infra/memoria"
	"perfila/internal/infra/relogio"
)

// Spec: 3.7 RN-10 Q-34 — conceder e revogar pela API gravam o movimento com origem api.
func TestConsentimentoPelaAPIRegistraOrigem(t *testing.T) {
	ctx := context.Background()
	clock := relogio.NovoFixo(time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC))
	clientes, cons := memoria.NovoClienteRepository(), memoria.NovoConsentimentoRepository()
	c, _ := entity.NovoCliente(uuid.New(), "Cliente Sintético", "c-001@exemplo.test", clock.Agora())
	if err := clientes.Salvar(ctx, c); err != nil {
		t.Fatal(err)
	}
	h := api.NovoRouter(api.Dependencias{AuthToken: "teste",
		Consentimento: usecase.NovoGerenciarConsentimentoUseCase(clientes, cons, clock, nil)})
	base := "/v1/clientes/" + c.ID.String() + "/consentimento"
	for _, req := range []*http.Request{
		httptest.NewRequest(http.MethodPut, base, strings.NewReader(`{"finalidade":"segmentacao","versaoTermo":"v1"}`)),
		httptest.NewRequest(http.MethodPost, base+"/revogacao", nil),
	} {
		req.Header.Set("Authorization", "Bearer teste")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s %s: status %d %s", req.Method, req.URL.Path, rec.Code, rec.Body)
		}
	}
	movs, err := cons.ListarMovimentos(ctx, c.ID)
	if err != nil || len(movs) != 2 {
		t.Fatalf("movimentos %+v (%v)", movs, err)
	}
	for i, tipo := range []entity.TipoMovimento{entity.MovimentoConcessao, entity.MovimentoRevogacao} {
		if movs[i].Tipo != tipo || movs[i].Origem != entity.OrigemAPI || movs[i].Operador != "" {
			t.Fatalf("movimento %d: %+v", i, movs[i])
		}
	}
}
