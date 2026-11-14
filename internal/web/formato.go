package web

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata" // fuso de São Paulo mesmo em imagens sem zoneinfo (UML8)

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
)

// Formatos brasileiros da interface (docs/design/03-textos.md, seção Formatos). O servidor
// guarda UTC; a tela mostra o horário de São Paulo.

func fusoSaoPaulo() *time.Location {
	if loc, err := time.LoadLocation("America/Sao_Paulo"); err == nil {
		return loc
	}
	return time.FixedZone("BRT", -3*60*60)
}

func dataHora(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	return t.In(fusoSaoPaulo()).Format("02/01/2006 15:04")
}

func soData(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	return t.In(fusoSaoPaulo()).Format("02/01/2006")
}

// dataHoraPtr formata um *time.Time; nulo vira "—".
func dataHoraPtr(t *time.Time) string {
	if t == nil {
		return "—"
	}
	return dataHora(*t)
}

// valorDataLocal é o valor de <input type="datetime-local"> no fuso de São Paulo.
func valorDataLocal(t time.Time) string { return t.In(fusoSaoPaulo()).Format("2006-01-02T15:04") }

// lerDataLocal interpreta o datetime-local no fuso de São Paulo e devolve UTC.
func lerDataLocal(s string) (time.Time, error) {
	t, err := time.ParseInLocation("2006-01-02T15:04", strings.TrimSpace(s), fusoSaoPaulo())
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

// decimal formata com vírgula e separador de milhar: decimal(1234.5, 2) = "1.234,50".
func decimal(v float64, casas int) string {
	s := strconv.FormatFloat(math.Abs(v), 'f', casas, 64)
	inteiro, frac, _ := strings.Cut(s, ".")
	var b strings.Builder
	for i, c := range inteiro {
		if i > 0 && (len(inteiro)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(c)
	}
	out := b.String()
	if frac != "" {
		out += "," + frac
	}
	if v < 0 && strings.Trim(out, "0.,") != "" {
		out = "-" + out
	}
	return out
}

func dinheiro(v float64) string { return "R$ " + decimal(v, 2) }

// lerDecimal aceita "1.234,56", "1234,56" e "1234.56".
func lerDecimal(s string) (float64, error) {
	s = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(s), "R$"))
	if strings.Contains(s, ",") {
		s = strings.ReplaceAll(strings.ReplaceAll(s, ".", ""), ",", ".")
	}
	return strconv.ParseFloat(s, 64)
}

// mascararEmail mostra só a primeira letra e o domínio: "c***@exemplo.test" (minimização).
func mascararEmail(e string) string {
	local, dominio, ok := strings.Cut(e, "@")
	if !ok || local == "" {
		return "***"
	}
	r := []rune(local)
	return string(r[0]) + "***@" + dominio
}

// uuidCurto: 8 primeiros + "…" + 4 últimos (UuidTag).
func uuidCurto(id uuid.UUID) string {
	s := id.String()
	return s[:8] + "…" + s[len(s)-4:]
}

// rotulos legíveis dos estados e tipos do domínio (03-textos.md).
func rotulo(v string) string {
	switch v {
	case string(entity.StatusNaoConcedido):
		return "Não concedido"
	case string(entity.StatusConcedido):
		return "Concedido"
	case string(entity.StatusRevogado):
		return "Revogado"
	case estadoEliminado:
		return "Eliminado"
	case string(entity.PedidoAberto):
		return "Aberto"
	case string(entity.PedidoAtendido):
		return "Atendido"
	case string(entity.StatusGerada):
		return "Gerada"
	case string(entity.StatusAtiva):
		return "Ativa"
	case string(entity.StatusExpirada):
		return "Expirada"
	case string(entity.TipoCompra):
		return "Compra"
	case string(entity.TipoNavegacao):
		return "Navegação"
	case string(entity.TipoInteracao):
		return "Interação"
	}
	return v
}

// tomDoEstado é a variante semântica da etiqueta de estado (StatusTag).
func tomDoEstado(v string) string {
	switch v {
	case string(entity.StatusConcedido), string(entity.StatusAtiva), string(entity.PedidoAtendido):
		return "ok"
	case string(entity.StatusRevogado), estadoEliminado:
		return "critical"
	case string(entity.StatusGerada), string(entity.PedidoAberto):
		return "warn"
	}
	return "neutral"
}

// condicao é uma condição do critério de um segmento, legível (CriterionChips).
type condicao struct {
	Original string
	Campo    string
	Operador string
	Numero   string
	Avaliada bool // o perfil foi comparado
	Atendida bool
}

var reCondicao = regexp.MustCompile(`^\s*([A-Za-z]+)\s*(>=|<=|==|>|<)\s*(\d+(?:\.\d+)?)\s*$`)

// condicoes divide o critério para exibição; cada condição é avaliada pelo próprio domínio
// (entity.AvaliarCriterio), então a tela nunca decide por conta própria. Spec: RN-06.
func condicoes(criterio string, perfil *entity.PerfilCliente) []condicao {
	if strings.TrimSpace(criterio) == "" {
		return nil
	}
	campos := map[string]string{"frequenciaCompras": "Compras", "ticketMedio": "Ticket médio", "scoreEngajamento": "Engajamento"}
	ops := map[string]string{">=": "≥", "<=": "≤", "==": "=", ">": ">", "<": "<"}
	var out []condicao
	for _, trecho := range strings.Split(criterio, "&&") {
		c := condicao{Original: strings.TrimSpace(trecho), Campo: strings.TrimSpace(trecho)}
		if m := reCondicao.FindStringSubmatch(trecho); m != nil {
			if nome, ok := campos[m[1]]; ok {
				c.Campo = nome
			}
			c.Operador = ops[m[2]]
			if n, err := strconv.ParseFloat(m[3], 64); err == nil {
				c.Numero = strings.TrimSuffix(strings.TrimRight(decimal(n, 2), "0"), ",")
			}
		}
		if perfil != nil {
			ok, err := entity.AvaliarCriterio(trecho, *perfil)
			c.Avaliada, c.Atendida = err == nil, ok && err == nil
		}
		out = append(out, c)
	}
	return out
}

func porcento(v float64) string { return fmt.Sprintf("%.0f", math.Max(0, math.Min(1, v))*100) }
