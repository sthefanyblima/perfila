// Package entity contém as entidades e os tipos enumerados do domínio.
//
// Spec: 3.2 (modelo de domínio), C-05 (tipos explícitos).
// Fonte: UML2 (Diagrama de Classes), UML5 e UML6 (Diagramas de Estados).
package entity

// TipoEvento classifica um EventoComportamental. Spec: 3.2, RN-04.
type TipoEvento string

const (
	TipoNavegacao TipoEvento = "navegacao"
	TipoCompra    TipoEvento = "compra"
	TipoInteracao TipoEvento = "interacao"
)

// Valido informa se o tipo pertence ao conjunto permitido. Spec: RN-04.
func (t TipoEvento) Valido() bool {
	switch t {
	case TipoNavegacao, TipoCompra, TipoInteracao:
		return true
	}
	return false
}

// StatusConsentimento é o estado da máquina 3.4.1. Spec: D-02.
type StatusConsentimento string

const (
	StatusNaoConcedido StatusConsentimento = "NAO_CONCEDIDO"
	StatusConcedido    StatusConsentimento = "CONCEDIDO"
	StatusRevogado     StatusConsentimento = "REVOGADO"
)

// StatusRecomendacao é o estado da máquina 3.4.2. Spec: D-08.
type StatusRecomendacao string

const (
	StatusGerada   StatusRecomendacao = "GERADA"
	StatusAtiva    StatusRecomendacao = "ATIVA"
	StatusExpirada StatusRecomendacao = "EXPIRADA"
)
