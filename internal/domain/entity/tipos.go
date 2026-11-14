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

// Finalidade é o uso autorizado pelo consentimento; há um registro por (cliente, finalidade).
// Spec: 3.2, RN-02, Q-48 (v1.9).
type Finalidade string

const (
	FinalidadeSegmentacao          Finalidade = "segmentacao"           // perfil, segmento e recomendação (uso interno)
	FinalidadePublicidadeTerceiros Finalidade = "publicidade_terceiros" // públicos para plataformas de anúncio (RN-17)
)

// Finalidades devolve as finalidades na ordem de exibição. Spec: Q-48.
func Finalidades() []Finalidade {
	return []Finalidade{FinalidadeSegmentacao, FinalidadePublicidadeTerceiros}
}

// Valida informa se a finalidade pertence ao enumerado (sem diferenciar nada: o texto é exato).
func (f Finalidade) Valida() bool {
	return f == FinalidadeSegmentacao || f == FinalidadePublicidadeTerceiros
}

// TipoExportacao diz se o arquivo de público inclui ou remove clientes. Spec: RN-17.
type TipoExportacao string

const (
	ExportacaoInclusao TipoExportacao = "INCLUSAO"
	ExportacaoRemocao  TipoExportacao = "REMOCAO"
)

// StatusMembro é o estado da máquina 3.4.3. Spec: RN-17, Q-50.
type StatusMembro string

const (
	MembroExportado       StatusMembro = "EXPORTADO"
	MembroRemocaoPendente StatusMembro = "REMOCAO_PENDENTE"
	MembroRemovido        StatusMembro = "REMOVIDO"
)

// TipoDestino identifica o destino publicitário; no MVP, só o arquivo CSV. Spec: RN-17, Q-49.
type TipoDestino string

const DestinoArquivoCSV TipoDestino = "arquivo_csv"
