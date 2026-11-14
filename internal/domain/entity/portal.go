package entity

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"

	"perfila/internal/shared/errs"
)

// ValidadeTokenAcesso é por quanto tempo o link do portal vale depois de criado. Spec: RN-19, Q-56.
const ValidadeTokenAcesso = 30 * time.Minute

// TokenAcessoTitular é o link de uso único do portal do titular. O token em claro só existe no
// link enviado; aqui fica o SHA-256 dele. Spec: 3.2, RN-19, Q-56. Fonte: UML2 (v1.10).
type TokenAcessoTitular struct {
	ID        uuid.UUID
	ClienteID uuid.UUID
	Hash      string // SHA-256 do token, 64 dígitos hexadecimais minúsculos
	CriadoEm  time.Time
	ExpiraEm  time.Time
	UsadoEm   *time.Time
}

// NovoTokenAcesso cria o token com ExpiraEm = agora + 30 minutos. Hash fora do formato ou
// cliente nulo: ErrDadosInvalidos. Spec: 3.2, RN-19.
func NovoTokenAcesso(id, clienteID uuid.UUID, hash string, agora time.Time) (TokenAcessoTitular, error) {
	if clienteID == uuid.Nil || !hashValido(hash) {
		return TokenAcessoTitular{}, fmt.Errorf("%w: token de acesso sem cliente ou com hash fora do formato", errs.ErrDadosInvalidos)
	}
	agora = agora.UTC()
	return TokenAcessoTitular{ID: id, ClienteID: clienteID, Hash: hash, CriadoEm: agora, ExpiraEm: agora.Add(ValidadeTokenAcesso)}, nil
}

func hashValido(h string) bool {
	if len(h) != 64 {
		return false
	}
	b, err := hex.DecodeString(h)
	return err == nil && hex.EncodeToString(b) == h // só minúsculas
}

// EstaValido: ainda não usado e com agora antes de ExpiraEm. Spec: RN-19.
func (t TokenAcessoTitular) EstaValido(agora time.Time) bool {
	return t.UsadoEm == nil && agora.Before(t.ExpiraEm)
}

// Usar marca o primeiro uso; token usado ou vencido: ErrAcessoInvalido. Spec: RN-19.
func (t *TokenAcessoTitular) Usar(agora time.Time) error {
	if !t.EstaValido(agora) {
		return fmt.Errorf("%w: token %s", errs.ErrAcessoInvalido, t.ID)
	}
	u := agora.UTC()
	t.UsadoEm = &u
	return nil
}

// StatusPedido é o estado do pedido de eliminação. Spec: 3.2, RN-20, Q-58.
type StatusPedido string

const (
	PedidoAberto   StatusPedido = "ABERTO"
	PedidoAtendido StatusPedido = "ATENDIDO"
)

// PedidoEliminacao é o pedido que o titular faz no portal e a operação atende no painel.
// Spec: 3.2, RN-20, Q-58. Fonte: UML2 (v1.10).
type PedidoEliminacao struct {
	ID              uuid.UUID
	ClienteID       uuid.UUID
	Status          StatusPedido
	DataPedido      time.Time
	DataAtendimento *time.Time
}

// NovoPedidoEliminacao nasce ABERTO, com DataPedido = agora; cliente nulo: ErrDadosInvalidos.
// Spec: RN-20.
func NovoPedidoEliminacao(id, clienteID uuid.UUID, agora time.Time) (PedidoEliminacao, error) {
	if clienteID == uuid.Nil {
		return PedidoEliminacao{}, fmt.Errorf("%w: pedido de eliminação sem cliente", errs.ErrDadosInvalidos)
	}
	return PedidoEliminacao{ID: id, ClienteID: clienteID, Status: PedidoAberto, DataPedido: agora.UTC()}, nil
}

// Atender aplica ABERTO → ATENDIDO; já atendido: ErrTransicaoInvalida. Spec: RN-17, RN-20.
func (p *PedidoEliminacao) Atender(agora time.Time) error {
	if p.Status != PedidoAberto {
		return fmt.Errorf("%w: pedido %s já está %s", errs.ErrTransicaoInvalida, p.ID, p.Status)
	}
	a := agora.UTC()
	p.Status, p.DataAtendimento = PedidoAtendido, &a
	return nil
}

// MensagemSimulada é o que o NotificadorSimulado grava na caixa de saída simulada: o destinatário
// vai mascarado, nunca o e-mail completo. Spec: 3.2, Q-57. Fonte: UML2 (v1.10).
type MensagemSimulada struct {
	ID        uuid.UUID
	Para      string
	Assunto   string
	Corpo     string
	EnviadaEm time.Time
}
