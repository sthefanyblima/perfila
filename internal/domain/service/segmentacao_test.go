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
func (s segmentosFixos) ListarTodos(context.Context) ([]entity.Segmento, error) { return s, nil }
func (s segmentosFixos) Salvar(context.Context, entity.Segmento) error          { return nil }

// Spec: RN-06, RN-14 — EscolherSegmento ignora inativos e respeita a prioridade, sem depender da ordem.
func TestEscolherSegmento(t *testing.T) {
	lista := []entity.Segmento{
		{Nome: "padrao", Prioridade: 99, Ativo: true},
		{Nome: "engajado", Criterio: "scoreEngajamento >= 0.7", Prioridade: 2, Ativo: true},
		{Nome: "desligado", Criterio: "scoreEngajamento >= 0.1", Prioridade: 1, Ativo: false},
	}
	casos := map[float64]string{0.9: "engajado", 0.5: "padrao"}
	for score, esperado := range casos {
		seg, err := EscolherSegmento(lista, entity.PerfilCliente{ScoreEngajamento: score})
		if err != nil || seg.Nome != esperado {
			t.Errorf("score %v: obtido %q (%v), esperava %q", score, seg.Nome, err, esperado)
		}
	}
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
