package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/shared/errs"
)

// Casos de uso do portal do titular. Não exigem consentimento: o próprio titular vê, entrega a si
// mesmo e pede a eliminação dos seus dados (exceção de C-03). Spec: v1.10, 3.1, 3.5, RN-19, RN-20,
// Q-55 a Q-59.

// LimiteLinksPorHora é quantos links um cadastro recebe por hora. Spec: RN-19, Q-56.
const LimiteLinksPorHora = 3

// HashToken é o SHA-256 do token do link, em hexadecimal minúsculo: só ele é guardado. Spec: RN-19.
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// SolicitarAcessoTitularUseCase envia o link de uso único para cada cadastro com o e-mail.
// Spec: RN-19 (passos 1 e 2).
type SolicitarAcessoTitularUseCase struct {
	clientes    port.ClienteRepository
	tokens      port.TokenAcessoRepository
	notificador port.Notificador
	clock       port.Clock
	urlBase     string // PORTAL_URL_BASE
}

// NovoSolicitarAcessoTitularUseCase injeta as dependências (C-08); urlBase vem da configuração,
// nunca do cabeçalho Host (Q-56).
func NovoSolicitarAcessoTitularUseCase(clientes port.ClienteRepository, tokens port.TokenAcessoRepository,
	notificador port.Notificador, clock port.Clock, urlBase string) *SolicitarAcessoTitularUseCase {
	return &SolicitarAcessoTitularUseCase{clientes: clientes, tokens: tokens, notificador: notificador, clock: clock,
		urlBase: strings.TrimRight(urlBase, "/")}
}

// Executar só falha com e-mail vazio ou sem arroba, ou com falha de infraestrutura: e-mail
// desconhecido, cadastro eliminado e limite atingido têm o mesmo resultado (RN-19, passo 2).
func (uc *SolicitarAcessoTitularUseCase) Executar(ctx context.Context, email string) error {
	email = strings.TrimSpace(email)
	if email == "" || !strings.Contains(email, "@") {
		return fmt.Errorf("%w: e-mail do pedido de acesso", errs.ErrDadosInvalidos)
	}
	clientes, err := uc.clientes.BuscarPorEmail(ctx, email)
	if err != nil {
		return err
	}
	agora := uc.clock.Agora()
	for _, c := range clientes {
		if c.EstaEliminado() {
			continue
		}
		n, err := uc.tokens.ContarDesde(ctx, c.ID, agora.Add(-time.Hour))
		if err != nil {
			return err
		}
		if n >= LimiteLinksPorHora {
			continue
		}
		if err := uc.enviar(ctx, c, agora); err != nil {
			return err
		}
	}
	return nil
}

func (uc *SolicitarAcessoTitularUseCase) enviar(ctx context.Context, c entity.Cliente, agora time.Time) error {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return fmt.Errorf("gerando token de acesso: %w", err)
	}
	token := hex.EncodeToString(b)
	t, err := entity.NovoTokenAcesso(uuid.New(), c.ID, HashToken(token), agora)
	if err != nil {
		return err
	}
	if err := uc.tokens.Salvar(ctx, t); err != nil {
		return err
	}
	return uc.notificador.EnviarLinkAcesso(ctx, c.Email, uc.urlBase+"/titular/acesso?t="+token, t.ExpiraEm, c.DataCadastro)
}

// AutenticarTitularUseCase consome o link e devolve o cliente da sessão. Spec: RN-19 (passo 3).
type AutenticarTitularUseCase struct {
	clientes port.ClienteRepository
	tokens   port.TokenAcessoRepository
	clock    port.Clock
}

// NovoAutenticarTitularUseCase injeta as dependências (C-08).
func NovoAutenticarTitularUseCase(clientes port.ClienteRepository, tokens port.TokenAcessoRepository, clock port.Clock) *AutenticarTitularUseCase {
	return &AutenticarTitularUseCase{clientes: clientes, tokens: tokens, clock: clock}
}

// Executar: token inexistente, vencido, já usado, ou cliente inexistente ou eliminado, é
// ErrAcessoInvalido. Spec: RN-19.
func (uc *AutenticarTitularUseCase) Executar(ctx context.Context, token string) (uuid.UUID, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return uuid.Nil, errs.ErrAcessoInvalido
	}
	t, err := uc.tokens.Consumir(ctx, HashToken(token), uc.clock.Agora())
	if err != nil {
		return uuid.Nil, err
	}
	c, err := uc.clientes.BuscarPorID(ctx, t.ClienteID)
	if errors.Is(err, errs.ErrClienteNaoEncontrado) || (err == nil && c.EstaEliminado()) {
		return uuid.Nil, fmt.Errorf("%w: cliente do link %s", errs.ErrAcessoInvalido, t.ClienteID)
	}
	if err != nil {
		return uuid.Nil, err
	}
	return c.ID, nil
}

// MeusDados é o que o titular vê no portal. Spec: 3.5, Q-55.
type MeusDados struct {
	Cliente       entity.Cliente
	Consentimento entity.ConsentimentoLGPD        // sintético em NAO_CONCEDIDO sem registro
	Historico     []entity.MovimentoConsentimento // ordem cronológica (RN-10)
	PedidoAberto  *entity.PedidoEliminacao        // nil sem pedido aberto
}

// ConsultarMeusDadosUseCase monta Meus dados. Spec: 3.5, Q-55.
type ConsultarMeusDadosUseCase struct {
	clientes      port.ClienteRepository
	consentimento *GerenciarConsentimentoUseCase
	pedidos       port.PedidoEliminacaoRepository
}

// NovoConsultarMeusDadosUseCase injeta as dependências (C-08).
func NovoConsultarMeusDadosUseCase(clientes port.ClienteRepository, consentimento *GerenciarConsentimentoUseCase,
	pedidos port.PedidoEliminacaoRepository) *ConsultarMeusDadosUseCase {
	return &ConsultarMeusDadosUseCase{clientes: clientes, consentimento: consentimento, pedidos: pedidos}
}

// Executar: cliente inexistente ou eliminado encerra a sessão do portal. Spec: 3.5, 3.7.2.
func (uc *ConsultarMeusDadosUseCase) Executar(ctx context.Context, clienteID uuid.UUID) (MeusDados, error) {
	c, err := uc.clientes.BuscarPorID(ctx, clienteID)
	if err != nil {
		return MeusDados{}, err
	}
	if c.EstaEliminado() {
		return MeusDados{}, fmt.Errorf("%w: %s", errs.ErrClienteEliminado, clienteID)
	}
	d := MeusDados{Cliente: c}
	if d.Consentimento, err = uc.consentimento.Consultar(ctx, clienteID); err != nil {
		return MeusDados{}, err
	}
	if d.Historico, err = uc.consentimento.Historico(ctx, clienteID); err != nil {
		return MeusDados{}, err
	}
	p, err := uc.pedidos.BuscarAbertoPorCliente(ctx, clienteID)
	switch {
	case err == nil:
		d.PedidoAberto = &p
	case !errors.Is(err, errs.ErrPedidoNaoEncontrado):
		return MeusDados{}, err
	}
	return d, nil
}

// PedirEliminacaoUseCase registra o pedido de eliminação do titular. Spec: RN-20, Q-58.
type PedirEliminacaoUseCase struct {
	clientes      port.ClienteRepository
	consentimento *GerenciarConsentimentoUseCase
	pedidos       port.PedidoEliminacaoRepository
	clock         port.Clock
}

// NovoPedirEliminacaoUseCase injeta as dependências (C-08).
func NovoPedirEliminacaoUseCase(clientes port.ClienteRepository, consentimento *GerenciarConsentimentoUseCase,
	pedidos port.PedidoEliminacaoRepository, clock port.Clock) *PedirEliminacaoUseCase {
	return &PedirEliminacaoUseCase{clientes: clientes, consentimento: consentimento, pedidos: pedidos, clock: clock}
}

// Executar segue a ordem da RN-20; com pedido aberto, devolve o mesmo pedido.
func (uc *PedirEliminacaoUseCase) Executar(ctx context.Context, clienteID uuid.UUID) (entity.PedidoEliminacao, error) {
	c, err := uc.clientes.BuscarPorID(ctx, clienteID)
	if err != nil {
		return entity.PedidoEliminacao{}, err
	}
	if c.EstaEliminado() {
		return entity.PedidoEliminacao{}, fmt.Errorf("%w: %s", errs.ErrClienteEliminado, clienteID)
	}
	cons, err := uc.consentimento.Consultar(ctx, clienteID)
	if err != nil {
		return entity.PedidoEliminacao{}, err
	}
	if cons.Status == entity.StatusConcedido {
		return entity.PedidoEliminacao{}, fmt.Errorf("%w: retire a autorização do cliente %s antes de pedir a eliminação",
			errs.ErrTransicaoInvalida, clienteID)
	}
	aberto, err := uc.pedidos.BuscarAbertoPorCliente(ctx, clienteID)
	if err == nil {
		return aberto, nil
	}
	if !errors.Is(err, errs.ErrPedidoNaoEncontrado) {
		return entity.PedidoEliminacao{}, err
	}
	p, err := entity.NovoPedidoEliminacao(uuid.New(), clienteID, uc.clock.Agora())
	if err != nil {
		return entity.PedidoEliminacao{}, err
	}
	if err := uc.pedidos.Salvar(ctx, p); err != nil {
		return entity.PedidoEliminacao{}, err
	}
	return p, nil
}

// LinhaPedido é uma linha de Pedidos do titular no painel. Spec: 3.5, Q-58.
type LinhaPedido struct {
	Pedido        entity.PedidoEliminacao
	Cliente       entity.Cliente
	Consentimento entity.StatusConsentimento
}

// ListarPedidosEliminacaoUseCase lista os pedidos para a operação. Spec: 3.5, Q-58.
type ListarPedidosEliminacaoUseCase struct {
	pedidos       port.PedidoEliminacaoRepository
	clientes      port.ClienteRepository
	consentimento *GerenciarConsentimentoUseCase
}

// NovoListarPedidosEliminacaoUseCase injeta as dependências (C-08).
func NovoListarPedidosEliminacaoUseCase(pedidos port.PedidoEliminacaoRepository, clientes port.ClienteRepository,
	consentimento *GerenciarConsentimentoUseCase) *ListarPedidosEliminacaoUseCase {
	return &ListarPedidosEliminacaoUseCase{pedidos: pedidos, clientes: clientes, consentimento: consentimento}
}

// Executar: abertos primeiro, depois DataPedido decrescente (a ordem do repositório).
func (uc *ListarPedidosEliminacaoUseCase) Executar(ctx context.Context) ([]LinhaPedido, error) {
	pedidos, err := uc.pedidos.Listar(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]LinhaPedido, 0, len(pedidos))
	for _, p := range pedidos {
		c, err := uc.clientes.BuscarPorID(ctx, p.ClienteID)
		if err != nil {
			return nil, err
		}
		cons, err := uc.consentimento.Consultar(ctx, p.ClienteID)
		if err != nil {
			return nil, err
		}
		out = append(out, LinhaPedido{Pedido: p, Cliente: c, Consentimento: cons.Status})
	}
	return out, nil
}

// ListarCaixaSaidaUseCase mostra a caixa de saída simulada das últimas 24 horas. Spec: 3.5, Q-57.
type ListarCaixaSaidaUseCase struct {
	caixa port.CaixaSaidaSimulada
	clock port.Clock
}

// NovoListarCaixaSaidaUseCase injeta as dependências (C-08).
func NovoListarCaixaSaidaUseCase(caixa port.CaixaSaidaSimulada, clock port.Clock) *ListarCaixaSaidaUseCase {
	return &ListarCaixaSaidaUseCase{caixa: caixa, clock: clock}
}

// Executar devolve da mais recente para a mais antiga.
func (uc *ListarCaixaSaidaUseCase) Executar(ctx context.Context) ([]entity.MensagemSimulada, error) {
	return uc.caixa.Listar(ctx, uc.clock.Agora().Add(-24*time.Hour))
}
