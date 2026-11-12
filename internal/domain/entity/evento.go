package entity

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"perfila/internal/shared/errs"
)

// Limites de RN-04 (constantes, não variáveis mutáveis — C-08).
const (
	ValorMaximoEvento = 1_000_000.0     // Q-15: impede o estouro das somas do perfil
	ToleranciaRelogio = 2 * time.Minute // Q-16: deriva aceitável do relógio do dispositivo
)

// EventoComportamental é uma interação do cliente (navegação, compra ou interação).
// Spec: 3.2, RN-04. Fonte: UML2.
type EventoComportamental struct {
	ID        uuid.UUID
	ClienteID uuid.UUID
	Tipo      TipoEvento
	Origem    string
	Timestamp time.Time
	Valor     float64
	Categoria string // [ADIÇÃO] opcional (D-05)
}

// Validar aplica RN-04 (com Q-15 e Q-16). Qualquer violação devolve ErrDadosInvalidos.
func (e EventoComportamental) Validar(agora time.Time) error {
	invalido := func(motivo string) error {
		return fmt.Errorf("%w: %s", errs.ErrDadosInvalidos, motivo)
	}
	switch {
	case e.ClienteID == uuid.Nil:
		return invalido("clienteId nulo")
	case !e.Tipo.Valido():
		return invalido("tipo de evento desconhecido")
	case strings.TrimSpace(e.Origem) == "":
		return invalido("origem vazia")
	case e.Timestamp.IsZero():
		return invalido("timestamp ausente")
	case e.Timestamp.After(agora.Add(ToleranciaRelogio)):
		return invalido("timestamp no futuro")
	case !(e.Valor >= 0): // também rejeita NaN
		return invalido("valor negativo")
	case e.Valor > ValorMaximoEvento:
		return invalido("valor acima do máximo")
	case e.Tipo == TipoCompra && e.Valor <= 0:
		return invalido("compra exige valor maior que zero")
	}
	return nil
}
