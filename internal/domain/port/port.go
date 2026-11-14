// Package port declara os contratos (ports) que o domínio exige da infraestrutura.
// Implementações ficam em internal/infra e provam conformidade em tempo de compilação.
//
// Spec: 3.6, C-02, C-07. Fonte: UML7 (Diagrama de Componentes — Interfaces/Ports).
package port

import (
	"context"
	"time"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
)

// ClienteRepository. BuscarPorID devolve errs.ErrClienteNaoEncontrado se não existir.
type ClienteRepository interface {
	Salvar(ctx context.Context, c entity.Cliente) error
	BuscarPorID(ctx context.Context, id uuid.UUID) (entity.Cliente, error)
	// Listar [ADIÇÃO] devolve a página pedida e o total que casa com a busca. Spec: 3.6, Q-23.
	Listar(ctx context.Context, f FiltroClientes) ([]entity.Cliente, int, error)
	// BuscarPorEmail [ADIÇÃO] v1.10 compara sem maiúsculas e sem espaços nas pontas; cadastro mais
	// recente primeiro; vazio sem erro. Spec: RN-19, Q-59.
	BuscarPorEmail(ctx context.Context, email string) ([]entity.Cliente, error)
}

// FiltroClientes [ADIÇÃO] seleciona e pagina clientes, do cadastro mais recente para o mais
// antigo (desempate por ID). Spec: 3.6, Q-23.
type FiltroClientes struct {
	Busca        string // vazio = todos; UUID = ID exato; senão, nome contém (sem maiúsculas nem acentos)
	Deslocamento int
	Limite       int // 0 = sem limite
}

// EventoRepository. Spec: UC-01.
type EventoRepository interface {
	Salvar(ctx context.Context, e entity.EventoComportamental) error
	BuscarPorCliente(ctx context.Context, clienteID uuid.UUID) ([]entity.EventoComportamental, error)
	// RemoverPorCliente [ADIÇÃO] v1.9 apaga os eventos do cliente e diz quantos; 0 sem erro. Spec: RN-17, Q-52.
	RemoverPorCliente(ctx context.Context, clienteID uuid.UUID) (int, error)
}

// PerfilRepository. Salvar faz upsert por ClienteID e preserva o ID original (RN-05).
// BuscarPorCliente devolve errs.ErrPerfilNaoEncontrado se não existir.
type PerfilRepository interface {
	Salvar(ctx context.Context, p entity.PerfilCliente) (entity.PerfilCliente, error)
	BuscarPorCliente(ctx context.Context, clienteID uuid.UUID) (entity.PerfilCliente, error)
	// RemoverPorCliente [ADIÇÃO] v1.9 apaga o perfil do cliente; devolve 0 ou 1. Spec: RN-17, Q-52.
	RemoverPorCliente(ctx context.Context, clienteID uuid.UUID) (int, error)
}

// SegmentoRepository. ListarAtivos devolve em ordem crescente de Prioridade (RN-06).
type SegmentoRepository interface {
	ListarAtivos(ctx context.Context) ([]entity.Segmento, error)
	BuscarPorID(ctx context.Context, id uuid.UUID) (entity.Segmento, error) // [ADIÇÃO] errs.ErrDadosInvalidos se inexistente
}

// RecomendacaoRepository. Salvar insere ou atualiza por ID.
type RecomendacaoRepository interface {
	Salvar(ctx context.Context, r entity.RecomendacaoProduto) error
	ListarAtivasVencidas(ctx context.Context, agora time.Time) ([]entity.RecomendacaoProduto, error)       // [ADIÇÃO]
	ListarAtivasPorCliente(ctx context.Context, clienteID uuid.UUID) ([]entity.RecomendacaoProduto, error) // [ADIÇÃO] Q-18
	ListarPorCliente(ctx context.Context, clienteID uuid.UUID) ([]entity.RecomendacaoProduto, error)       // [ADIÇÃO] Q-23: DataGeracao decrescente
	// RemoverPorCliente [ADIÇÃO] v1.9 apaga as recomendações do cliente e diz quantas. Spec: RN-17, Q-52.
	RemoverPorCliente(ctx context.Context, clienteID uuid.UUID) (int, error)
}

// ConsentimentoRepository. Salvar faz upsert por ClienteID, sem movimento (cargas e testes).
// BuscarPorCliente devolve errs.ErrConsentimentoNaoEncontrado se não existir.
// Spec: 3.6, RN-10, Q-34.
type ConsentimentoRepository interface {
	Salvar(ctx context.Context, c entity.ConsentimentoLGPD) error
	BuscarPorCliente(ctx context.Context, clienteID uuid.UUID) (entity.ConsentimentoLGPD, error)
	// Registrar [ADIÇÃO] v1.4: upsert do registro e inclusão do movimento numa operação atômica;
	// ID de movimento repetido falha sem gravar nada.
	Registrar(ctx context.Context, c entity.ConsentimentoLGPD, m entity.MovimentoConsentimento) error
	// ListarMovimentos [ADIÇÃO] v1.4: Data crescente, desempate pela ordem de gravação.
	ListarMovimentos(ctx context.Context, clienteID uuid.UUID) ([]entity.MovimentoConsentimento, error)
	// ListarRevogadosAte [ADIÇÃO] v1.9: clientes REVOGADO com DataRevogacao <= limite, da revogação
	// mais antiga para a mais recente. Spec: RN-18, Q-51.
	ListarRevogadosAte(ctx context.Context, limite time.Time) ([]uuid.UUID, error)
}

// ProdutoRepository [ADIÇÃO] (D-05).
type ProdutoRepository interface {
	ListarAtivosPorCategoria(ctx context.Context, categoria string) ([]entity.Produto, error)
	ListarTodos(ctx context.Context) ([]entity.Produto, error) // [ADIÇÃO] Q-23: por Categoria e Nome
}

// ConsentimentoGateway — MVP: implementação local sobre ConsentimentoRepository (fase Support).
type ConsentimentoGateway interface {
	Validar(ctx context.Context, clienteID uuid.UUID) (bool, error)
}

// IntegracaoGateway — e-commerce/CRM; MVP: stub que devolve lista vazia (D-12, fase Support).
type IntegracaoGateway interface {
	BuscarEventos(ctx context.Context, clienteID uuid.UUID, desde time.Time) ([]entity.EventoComportamental, error)
}

// Clock é a única fonte de tempo do sistema (C-07).
type Clock interface{ Agora() time.Time }

// TokenAcessoRepository [ADIÇÃO] v1.10 guarda os links de uso único do portal (só o hash).
// Spec: 3.6, RN-19, Q-56, Q-59.
type TokenAcessoRepository interface {
	Salvar(ctx context.Context, t entity.TokenAcessoTitular) error
	// Consumir marca UsadoEm de forma atômica (vale entre réplicas); inexistente, vencido ou já
	// usado: errs.ErrAcessoInvalido.
	Consumir(ctx context.Context, hash string, agora time.Time) (entity.TokenAcessoTitular, error)
	// ContarDesde conta os tokens do cliente com CriadoEm >= desde (limite por cadastro).
	ContarDesde(ctx context.Context, clienteID uuid.UUID, desde time.Time) (int, error)
	// RemoverPorCliente apaga os tokens do cliente e diz quantos. Spec: RN-17.
	RemoverPorCliente(ctx context.Context, clienteID uuid.UUID) (int, error)
}

// PedidoEliminacaoRepository [ADIÇÃO] v1.10 guarda os pedidos feitos no portal. Spec: RN-20, Q-58.
type PedidoEliminacaoRepository interface {
	Salvar(ctx context.Context, p entity.PedidoEliminacao) error // insere ou atualiza por ID; um ABERTO por cliente
	// BuscarAbertoPorCliente devolve errs.ErrPedidoNaoEncontrado sem pedido aberto.
	BuscarAbertoPorCliente(ctx context.Context, clienteID uuid.UUID) (entity.PedidoEliminacao, error)
	Listar(ctx context.Context) ([]entity.PedidoEliminacao, error) // ABERTO primeiro, depois DataPedido decrescente
}

// Notificador [ADIÇÃO] v1.10 avisa o titular; no MVP, NotificadorSimulado (Q-57).
type Notificador interface {
	EnviarLinkAcesso(ctx context.Context, para, link string, expiraEm, dataCadastro time.Time) error
}

// CaixaSaidaSimulada [ADIÇÃO] v1.10 guarda as mensagens do NotificadorSimulado. Spec: Q-57.
type CaixaSaidaSimulada interface {
	Guardar(ctx context.Context, m entity.MensagemSimulada) error
	Listar(ctx context.Context, desde time.Time) ([]entity.MensagemSimulada, error) // EnviadaEm >= desde, mais recente primeiro
	RemoverAte(ctx context.Context, limite time.Time) (int, error)                  // EnviadaEm < limite
}
