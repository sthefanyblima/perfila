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
}

// PerfilRepository. Salvar faz upsert por ClienteID e preserva o ID original (RN-05).
// BuscarPorCliente devolve errs.ErrPerfilNaoEncontrado se não existir.
type PerfilRepository interface {
	Salvar(ctx context.Context, p entity.PerfilCliente) (entity.PerfilCliente, error)
	BuscarPorCliente(ctx context.Context, clienteID uuid.UUID) (entity.PerfilCliente, error)
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
}

// ProdutoRepository [ADIÇÃO] (D-05).
type ProdutoRepository interface {
	ListarAtivosPorCategoria(ctx context.Context, categoria string) ([]entity.Produto, error)
	ListarTodos(ctx context.Context) ([]entity.Produto, error) // [ADIÇÃO] Q-23: por Categoria e Nome
}

// RelatorioRepository [ADIÇÃO] v1.7: leitura agregada dos relatórios. Devolve só contagens e
// somas, nunca linhas de cliente. "Consentimento válido" é o registro CONCEDIDO no momento da
// leitura (RN-02); cliente sem registro conta como NAO_CONCEDIDO (D-02). Intervalos são
// fechados no início e abertos no fim. A proteção de grupos pequenos fica no domínio.
// Spec: 3.6, RN-15, Q-46.
type RelatorioRepository interface {
	// ContarClientes conta todos os clientes, ativos e por estado do consentimento.
	ContarClientes(ctx context.Context) (ContagemClientes, error)
	// ResumirSegmentos agrupa os perfis de quem tem consentimento válido por SegmentoID
	// (nil = sem segmento), em qualquer ordem.
	ResumirSegmentos(ctx context.Context) ([]ResumoSegmento, error)
	// ContarRecomendacoes conta as ATIVA e EXPIRADA de quem tem consentimento válido.
	ContarRecomendacoes(ctx context.Context) (ContagemRecomendacoes, error)
	// ContarCompras agrupa por categoria as compras com categoria preenchida em [desde, ate) de
	// quem tem consentimento válido, por Compras decrescente e Categoria.
	ContarCompras(ctx context.Context, desde, ate time.Time) ([]ComprasCategoria, error)
	// ContarCompradores conta os clientes distintos com consentimento válido que compraram em
	// alguma das categorias em [desde, ate).
	ContarCompradores(ctx context.Context, desde, ate time.Time, categorias []string) (int, error)
	// SerieMensal devolve um item por intervalo [limites[i], limites[i+1]), em ordem.
	SerieMensal(ctx context.Context, limites []time.Time) ([]ContagemMes, error)
}

// ContagemClientes: todos os clientes. Spec: RN-15.
type ContagemClientes struct {
	Total, Ativos                     int
	Concedido, NaoConcedido, Revogado int
}

// ResumoSegmento: perfis de clientes com consentimento válido num segmento. Spec: RN-15.
type ResumoSegmento struct {
	SegmentoID            *uuid.UUID
	Clientes              int
	SomaTicket, SomaScore float64
}

// ContagemRecomendacoes: recomendações e clientes distintos por estado. Spec: RN-15.
type ContagemRecomendacoes struct {
	Ativas, ClientesAtivas       int
	Expiradas, ClientesExpiradas int
}

// ComprasCategoria: compras e clientes distintos de uma categoria. Spec: RN-15.
type ComprasCategoria struct {
	Categoria         string
	Compras, Clientes int
}

// ContagemMes: cadastros e revogações (movimentos REVOGACAO) de todos os clientes; eventos e
// seus clientes distintos só de quem tem consentimento válido. Spec: RN-15.
type ContagemMes struct {
	Inicio                   time.Time
	Cadastros, Revogacoes    int
	Eventos, ClientesEventos int
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
