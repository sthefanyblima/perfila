package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/shared/errs"
)

type segmentosFixos []entity.Segmento

func (s segmentosFixos) ListarAtivos(context.Context) ([]entity.Segmento, error) { return s, nil }
func (s segmentosFixos) BuscarPorID(context.Context, uuid.UUID) (entity.Segmento, error) {
	return entity.Segmento{}, nil
}

// Spec: RN-06, Q-02 — sem segmento padrao é falha de configuração, não dado inválido do cliente.
func TestSegmentarSemPadrao(t *testing.T) {
	svc := NovoSegmentacaoService(segmentosFixos{{Nome: "premium", Criterio: "ticketMedio >= 500", Ativo: true}})
	_, err := svc.Segmentar(context.Background(), entity.PerfilCliente{TicketMedio: 10})
	if !errors.Is(err, ErrSemSegmentoAplicavel) {
		t.Fatalf("esperava ErrSemSegmentoAplicavel, obtido %v", err)
	}
	if errs.Codigo(err) != "" {
		t.Fatalf("não deve ser erro de domínio com código HTTP 4xx: %q", errs.Codigo(err))
	}
}

func TestSegmentarCriterioInvalido(t *testing.T) {
	svc := NovoSegmentacaoService(segmentosFixos{{Nome: "x", Criterio: "idade > 1", Ativo: true}})
	_, err := svc.Segmentar(context.Background(), entity.PerfilCliente{})
	if !errors.Is(err, ErrSegmentoMalConfigurado) || errs.Codigo(err) != "" {
		t.Fatalf("critério inválido configurado deveria ser falha de configuração (Q-17): %v", err)
	}
}
