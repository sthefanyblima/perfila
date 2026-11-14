package entity

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"

	"perfila/internal/shared/errs"
)

// TamanhoMinimoGrupo é o menor grupo detalhado no relatório agregado. Spec: RN-18, Q-52.
const TamanhoMinimoGrupo = 5

// HashEmail é o SHA-256, em hexadecimal minúsculo, do e-mail sem nenhum espaço e em minúsculas.
// Pseudonimiza, não anonimiza: o resultado continua sendo dado pessoal. Spec: RN-17, Q-49.
func HashEmail(email string) string {
	normalizado := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return unicode.ToLower(r)
	}, email)
	soma := sha256.Sum256([]byte(normalizado))
	return hex.EncodeToString(soma[:])
}

// ExportacaoPublico registra cada arquivo de público gerado, sem dado pessoal. Spec: 3.2, RN-17, Q-51.
type ExportacaoPublico struct {
	ID              uuid.UUID
	SegmentoID      uuid.UUID
	SegmentoNome    string
	Tipo            TipoExportacao
	Destino         TipoDestino
	Operador        string
	Data            time.Time
	Quantidade      int
	HashArquivo     string
	NomeArquivo     string
	ArquivoExpiraEm time.Time
}

// MembroPublico liga um cliente a um arquivo de inclusão pelo hash do e-mail (pseudônimo, ainda
// dado pessoal). Spec: 3.2, 3.4.3, RN-17.
type MembroPublico struct {
	ID                  uuid.UUID
	ExportacaoID        uuid.UUID
	SegmentoID          uuid.UUID
	ClienteID           uuid.UUID
	HashEmail           string
	Status              StatusMembro
	DataInclusao        time.Time
	DataRemocaoPendente *time.Time
	RemocaoID           *uuid.UUID
	DataRemocao         *time.Time
}

// NovoMembro inclui o cliente no estado EXPORTADO. Spec: 3.4.3.
func NovoMembro(id, exportacaoID, segmentoID, clienteID uuid.UUID, hashEmail string, agora time.Time) (MembroPublico, error) {
	if strings.TrimSpace(hashEmail) == "" {
		return MembroPublico{}, fmt.Errorf("%w: membro de público sem hash", errs.ErrDadosInvalidos)
	}
	return MembroPublico{ID: id, ExportacaoID: exportacaoID, SegmentoID: segmentoID, ClienteID: clienteID,
		HashEmail: hashEmail, Status: MembroExportado, DataInclusao: agora.UTC()}, nil
}

// MarcarRemocao aplica EXPORTADO → REMOCAO_PENDENTE. Spec: 3.4.3, Q-50.
func (m *MembroPublico) MarcarRemocao(agora time.Time) error {
	if m.Status != MembroExportado {
		return fmt.Errorf("%w: marcar remoção a partir de %s", errs.ErrTransicaoInvalida, m.Status)
	}
	agora = agora.UTC()
	m.Status = MembroRemocaoPendente
	m.DataRemocaoPendente = &agora
	return nil
}

// Remover aplica REMOCAO_PENDENTE → REMOVIDO, com a exportação de remoção que o tirou. Spec: 3.4.3.
func (m *MembroPublico) Remover(remocaoID uuid.UUID, agora time.Time) error {
	if m.Status != MembroRemocaoPendente {
		return fmt.Errorf("%w: remover a partir de %s", errs.ErrTransicaoInvalida, m.Status)
	}
	if remocaoID == uuid.Nil {
		return fmt.Errorf("%w: remoção sem exportação", errs.ErrDadosInvalidos)
	}
	agora = agora.UTC()
	m.Status = MembroRemovido
	m.RemocaoID = &remocaoID
	m.DataRemocao = &agora
	return nil
}

// LotePublico é o que vai para o destino: só hashes e o nome do segmento. Spec: RN-17.
type LotePublico struct {
	Tipo         TipoExportacao
	SegmentoNome string
	Data         time.Time
	Hashes       []string
}

// ReciboEnvio é o que o destino devolve. Spec: RN-17, Q-49.
type ReciboEnvio struct {
	Destino     TipoDestino
	NomeArquivo string
	Conteudo    []byte
	Hash        string
}

// LinhaAgregada é uma linha do relatório por segmento, sem identificador. Spec: RN-18, Q-52.
type LinhaAgregada struct {
	SegmentoNome       string
	Quantidade         int
	TicketMedio        float64
	CategoriaMaisComum string
	Oculta             bool
}
