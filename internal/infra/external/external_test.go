package external

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/service"
	"perfila/internal/infra/memoria"
)

func TestConsentimentoLocal(t *testing.T) {
	ctx := context.Background()
	repo := memoria.NovoConsentimentoRepository()
	g := NovoConsentimentoLocal(service.NovoValidadorLGPD(repo))
	id := uuid.New()
	if ok, err := g.Validar(ctx, id); ok || err != nil {
		t.Fatalf("sem registro deveria ser inválido: %v %v", ok, err)
	}
	c := entity.NovoConsentimento(uuid.New(), id, time.Now())
	if err := c.Conceder("segmentacao", "v1", time.Now()); err != nil {
		t.Fatal(err)
	}
	_ = repo.Salvar(ctx, c)
	if ok, _ := g.Validar(ctx, id); !ok {
		t.Fatal("concedido deveria ser válido")
	}
}

func TestIntegracaoStub(t *testing.T) {
	evs, err := IntegracaoStub{}.BuscarEventos(context.Background(), uuid.New(), time.Now())
	if err != nil || evs == nil || len(evs) != 0 {
		t.Fatalf("stub deveria devolver lista vazia: %v %v", evs, err)
	}
}
