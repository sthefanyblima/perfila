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
	// ListarPorSegmento [ADIÇÃO] v1.9: os perfis do segmento, em ordem de ClienteID. Spec: RN-17, RN-18.
	ListarPorSegmento(ctx context.Context, segmentoID uuid.UUID) ([]entity.PerfilCliente, error)
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

// ConsentimentoRepository guarda um registro por (ClienteID, Finalidade) desde a v1.9 (Q-48).
// Salvar faz upsert por (ClienteID, Finalidade), sem movimento (cargas e testes).
// BuscarPorCliente devolve errs.ErrConsentimentoNaoEncontrado se não existir.
// Spec: 3.6, RN-02, RN-10, Q-34, Q-48.
type ConsentimentoRepository interface {
	Salvar(ctx context.Context, c entity.ConsentimentoLGPD) error
	BuscarPorCliente(ctx context.Context, clienteID uuid.UUID, finalidade entity.Finalidade) (entity.ConsentimentoLGPD, error)
	// ListarPorCliente [ADIÇÃO] v1.9: só os registros que existem, em ordem de Finalidade.
	ListarPorCliente(ctx context.Context, clienteID uuid.UUID) ([]entity.ConsentimentoLGPD, error)
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

// PublicoRepository [ADIÇÃO] v1.9 guarda as exportações de públicos, o conteúdo dos arquivos
// (até ArquivoExpiraEm) e os membros. Spec: 3.6, RN-17, Q-51.
type PublicoRepository interface {
	// RegistrarExportacao é atômica: insere a exportação e o conteúdo, insere os membros novos e
	// atualiza os existentes; exportação repetida falha sem gravar nada.
	RegistrarExportacao(ctx context.Context, e entity.ExportacaoPublico, conteudo []byte, membros []entity.MembroPublico) error
	ListarExportacoes(ctx context.Context, segmentoID uuid.UUID) ([]entity.ExportacaoPublico, error) // Data decrescente
	// BuscarArquivo devolve errs.ErrArquivoIndisponivel se não houver conteúdo.
	BuscarArquivo(ctx context.Context, exportacaoID uuid.UUID) (entity.ExportacaoPublico, []byte, error)
	ApagarArquivosVencidos(ctx context.Context, agora time.Time) (int, error)                                                       // ArquivoExpiraEm <= agora
	ListarMembrosPorSegmento(ctx context.Context, segmentoID uuid.UUID, status entity.StatusMembro) ([]entity.MembroPublico, error) // por HashEmail
	ListarMembrosPorCliente(ctx context.Context, clienteID uuid.UUID, status entity.StatusMembro) ([]entity.MembroPublico, error)   // por HashEmail
	// AtualizarMembros grava estado e datas; ID inexistente falha sem gravar nada.
	AtualizarMembros(ctx context.Context, ms []entity.MembroPublico) error
}

// DestinoPublicitarioGateway [ADIÇÃO] v1.9 entrega o lote a uma plataforma de anúncio. No MVP,
// gera um arquivo CSV (infra/external); trocar por API não muda os casos de uso. Spec: RN-17, Q-49.
type DestinoPublicitarioGateway interface {
	Enviar(ctx context.Context, lote entity.LotePublico) (entity.ReciboEnvio, error)
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
