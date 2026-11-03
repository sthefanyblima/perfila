package entity

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"perfila/internal/shared/errs"
)

// Segmento agrupa perfis por um critério declarativo. Spec: 3.2, RN-06, D-06. Fonte: UML2.
type Segmento struct {
	ID         uuid.UUID
	Nome       string
	Descricao  string
	Criterio   string // gramática em RN-06
	Prioridade int    // [ADIÇÃO] menor valor = avaliado primeiro
	Ativo      bool   // [ADIÇÃO]
}

// ClassificarPerfil informa se o perfil satisfaz o critério do segmento.
// Critério vazio casa sempre; critério fora da gramática devolve ErrDadosInvalidos.
// Spec: RN-06. Fonte: UML2 (ClassificarPerfil(perfil): bool).
func (s Segmento) ClassificarPerfil(p PerfilCliente) (bool, error) {
	return AvaliarCriterio(s.Criterio, p)
}

// AvaliarCriterio interpreta a gramática de RN-06 sem reflexão e sem execução dinâmica:
//
//	criterio := vazio | condicao { "&&" condicao }
//	condicao := campo operador numero
//	campo    := frequenciaCompras | ticketMedio | scoreEngajamento
//	operador := > | >= | < | <= | ==
//	numero   := decimal com ponto
//
// Todas as condições são validadas, mesmo após uma condição falsa. Spec: RN-06.
func AvaliarCriterio(criterio string, p PerfilCliente) (bool, error) {
	if strings.TrimSpace(criterio) == "" {
		return true, nil
	}
	reCondicao := regexp.MustCompile(`^\s*([A-Za-z]+)\s*(>=|<=|==|>|<)\s*(\d+(?:\.\d+)?)\s*$`)
	resultado := true
	for _, trecho := range strings.Split(criterio, "&&") {
		m := reCondicao.FindStringSubmatch(trecho)
		if m == nil {
			return false, fmt.Errorf("%w: condição fora da gramática: %q", errs.ErrDadosInvalidos, strings.TrimSpace(trecho))
		}
		valorCampo, ok := valorDoCampo(m[1], p)
		if !ok {
			return false, fmt.Errorf("%w: campo desconhecido: %q", errs.ErrDadosInvalidos, m[1])
		}
		limite, err := strconv.ParseFloat(m[3], 64)
		if err != nil {
			return false, fmt.Errorf("%w: número inválido: %q", errs.ErrDadosInvalidos, m[3])
		}
		if !comparar(valorCampo, m[2], limite) {
			resultado = false
		}
	}
	return resultado, nil
}

func valorDoCampo(campo string, p PerfilCliente) (float64, bool) {
	switch campo {
	case "frequenciaCompras":
		return float64(p.FrequenciaCompras), true
	case "ticketMedio":
		return p.TicketMedio, true
	case "scoreEngajamento":
		return p.ScoreEngajamento, true
	}
	return 0, false
}

func comparar(a float64, op string, b float64) bool {
	switch op {
	case ">":
		return a > b
	case ">=":
		return a >= b
	case "<":
		return a < b
	case "<=":
		return a <= b
	default: // "==" — único restante aceito pela expressão regular
		return a == b
	}
}
