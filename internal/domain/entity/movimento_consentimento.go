package entity

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"perfila/internal/shared/errs"
)

// TipoMovimento diz qual transição de 3.4.1 o movimento prova. Spec: 3.2, RN-10, Q-34.
type TipoMovimento string

const (
	MovimentoConcessao TipoMovimento = "CONCESSAO"
	MovimentoRevogacao TipoMovimento = "REVOGACAO"
)

// Valido informa se o tipo pertence ao conjunto permitido. Spec: RN-10.
func (t TipoMovimento) Valido() bool { return t == MovimentoConcessao || t == MovimentoRevogacao }

// OrigemMovimento diz por onde o pedido chegou. Spec: 3.2, RN-10, Q-34.
type OrigemMovimento string

const (
	OrigemAPI      OrigemMovimento = "api"
	OrigemPainel   OrigemMovimento = "painel"
	OrigemPortal   OrigemMovimento = "portal"
	OrigemSistema  OrigemMovimento = "sistema"  // sem origem informada: subcomandos, exemplos, testes
	OrigemMigracao OrigemMovimento = "migracao" // reconstruído das datas pela migração 006
)

// Valido informa se a origem pertence ao conjunto permitido. Spec: RN-10.
func (o OrigemMovimento) Valido() bool {
	switch o {
	case OrigemAPI, OrigemPainel, OrigemPortal, OrigemSistema, OrigemMigracao:
		return true
	}
	return false
}

// MovimentoConsentimento é a prova de uma concessão ou revogação. Só de inclusão: não tem
// métodos de alteração, e o estado continua vindo de ConsentimentoLGPD (RN-02).
// Spec: 3.2, RN-10, Q-34. Fonte: UML2 (v1.4).
type MovimentoConsentimento struct {
	ID          uuid.UUID
	ClienteID   uuid.UUID
	Tipo        TipoMovimento
	Finalidade  string // a concedida ou a que foi revogada
	VersaoTermo string // idem
	Data        time.Time
	Origem      OrigemMovimento
	Operador    string // login do operador; obrigatório só com Origem painel
}

// NovoMovimento monta o movimento logo depois da transição, com finalidade, termo e data
// (DataAtualizacao) do consentimento. Spec: 3.2, RN-10.
func NovoMovimento(id uuid.UUID, c ConsentimentoLGPD, tipo TipoMovimento, origem OrigemMovimento, operador string) (MovimentoConsentimento, error) {
	operador = strings.TrimSpace(operador)
	switch {
	case !tipo.Valido():
		return MovimentoConsentimento{}, fmt.Errorf("%w: tipo de movimento %q", errs.ErrDadosInvalidos, tipo)
	case !origem.Valido():
		return MovimentoConsentimento{}, fmt.Errorf("%w: origem de movimento %q", errs.ErrDadosInvalidos, origem)
	case origem == OrigemPainel && operador == "":
		return MovimentoConsentimento{}, fmt.Errorf("%w: movimento do painel sem operador", errs.ErrDadosInvalidos)
	}
	return MovimentoConsentimento{
		ID:          id,
		ClienteID:   c.ClienteID,
		Tipo:        tipo,
		Finalidade:  string(c.Finalidade),
		VersaoTermo: c.VersaoTermo,
		Data:        c.DataAtualizacao.UTC(),
		Origem:      origem,
		Operador:    operador,
	}, nil
}

// FinalidadeEfetiva é a finalidade a que o movimento pertence: os movimentos anteriores à v1.9,
// com texto livre fora do enumerado, contam como segmentacao (o texto original fica). Spec: RN-10, Q-48.
func (m MovimentoConsentimento) FinalidadeEfetiva() Finalidade {
	if f := Finalidade(m.Finalidade); f.Valida() {
		return f
	}
	return FinalidadeSegmentacao
}
