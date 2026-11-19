package service

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
)

type relogioFixo struct{ t time.Time }

func (r relogioFixo) Agora() time.Time { return r.t }

// Spec: RN-05, RN-14. A categoria do evento é comparada na mesma forma do catálogo.
func TestGerarPerfilNormalizaCategoria(t *testing.T) {
	agora := time.Date(2026, 11, 19, 12, 0, 0, 0, time.UTC)
	s := NovoPerfilamentoService(relogioFixo{agora}, 10)
	compra := func(cat string) entity.EventoComportamental {
		return entity.EventoComportamental{Tipo: entity.TipoCompra, Valor: 10, Categoria: cat, Timestamp: agora.Add(-time.Hour)}
	}
	p := s.GerarPerfil(uuid.New(), []entity.EventoComportamental{compra("Esportes "), compra("esportes"), compra("livros"), compra("   ")})
	if p.CategoriaPreferida != "esportes" {
		t.Errorf("CategoriaPreferida = %q, esperado %q", p.CategoriaPreferida, "esportes")
	}
	if p.FrequenciaCompras != 4 {
		t.Errorf("FrequenciaCompras = %d, esperado 4", p.FrequenciaCompras)
	}
}
