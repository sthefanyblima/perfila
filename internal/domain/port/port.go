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
	// ListarPorSegmento [ADIÇÃO] v1.11: os perfis do segmento, em ordem de ClienteID. Spec: RN-21, RN-22.
	ListarPorSegmento(ctx context.Context, segmentoID uuid.UUID) ([]entity.PerfilCliente, error)
}

// SegmentoRepository. ListarAtivos devolve em ordem crescente de Prioridade (RN-06).
type SegmentoRepository interface {
	ListarAtivos(ctx context.Context) ([]entity.Segmento, error)
	BuscarPorID(ctx context.Context, id uuid.UUID) (entity.Segmento, error) // [ADIÇÃO] errs.ErrDadosInvalidos se inexistente
	ListarTodos(ctx context.Context) ([]entity.Segmento, error)             // [ADIÇÃO] v1.6: ativos e inativos, Prioridade crescente, desempate por ID
	Salvar(ctx context.Context, s entity.Segmento) error                    // [ADIÇÃO] v1.6 RN-14: insere ou atualiza por ID; nome repetido ou prioridade repetida entre ativos: errs.ErroCampo
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

// ConsentimentoRepository guarda um registro por (ClienteID, Finalidade) desde a v1.11 (Q-61).
// Salvar faz upsert por (ClienteID, Finalidade), sem movimento (cargas e testes).
// BuscarPorCliente devolve errs.ErrConsentimentoNaoEncontrado se não existir.
// Spec: 3.6, RN-02, RN-10, Q-34, Q-61.
type ConsentimentoRepository interface {
	Salvar(ctx context.Context, c entity.ConsentimentoLGPD) error
	BuscarPorCliente(ctx context.Context, clienteID uuid.UUID, finalidade entity.Finalidade) (entity.ConsentimentoLGPD, error)
	// ListarPorCliente [ADIÇÃO] v1.11: só os registros que existem, em ordem de Finalidade.
	ListarPorCliente(ctx context.Context, clienteID uuid.UUID) ([]entity.ConsentimentoLGPD, error)
	// Registrar [ADIÇÃO] v1.4: upsert do registro e inclusão do movimento numa operação atômica;
	// ID de movimento repetido falha sem gravar nada.
	Registrar(ctx context.Context, c entity.ConsentimentoLGPD, m entity.MovimentoConsentimento) error
	// ListarMovimentos [ADIÇÃO] v1.4: Data crescente, desempate pela ordem de gravação.
	ListarMovimentos(ctx context.Context, clienteID uuid.UUID) ([]entity.MovimentoConsentimento, error)
	ListarConcedidos(ctx context.Context) ([]entity.ConsentimentoLGPD, error) // [ADIÇÃO] v1.6 Q-42: Status CONCEDIDO, por ClienteID
	// ListarRevogadosAte [ADIÇÃO] v1.9: clientes REVOGADO com DataRevogacao <= limite, da revogação
	// mais antiga para a mais recente. Spec: RN-18, Q-51.
	ListarRevogadosAte(ctx context.Context, limite time.Time) ([]uuid.UUID, error)
}

// ProdutoRepository [ADIÇÃO] (D-05).
type ProdutoRepository interface {
	ListarAtivosPorCategoria(ctx context.Context, categoria string) ([]entity.Produto, error)
	ListarTodos(ctx context.Context) ([]entity.Produto, error)             // [ADIÇÃO] Q-23: por Categoria e Nome
	BuscarPorID(ctx context.Context, id uuid.UUID) (entity.Produto, error) // [ADIÇÃO] v1.6: errs.ErrProdutoNaoEncontrado
	Salvar(ctx context.Context, p entity.Produto) error                    // [ADIÇÃO] v1.6: insere ou atualiza por ID
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

// PublicoRepository [ADIÇÃO] v1.11 guarda as exportações de públicos, o conteúdo dos arquivos
// (até ArquivoExpiraEm) e os membros. Spec: 3.6, RN-21, Q-64.
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

// DestinoPublicitarioGateway [ADIÇÃO] v1.11 entrega o lote a uma plataforma de anúncio. No MVP,
// gera um arquivo CSV (infra/external); trocar por API não muda os casos de uso. Spec: RN-21, Q-62.
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

// OperadorRepository [ADIÇÃO v1.5]. Salvar insere ou atualiza por ID; Login de outro ID dá
// errs.ErrLoginEmUso. Buscas devolvem errs.ErrOperadorNaoEncontrado. CriarSeVazio grava só se
// não houver nenhum operador, de forma atômica (duas réplicas, RN-12). Spec: 3.6, Q-35.
type OperadorRepository interface {
	Salvar(ctx context.Context, o entity.Operador) error
	BuscarPorID(ctx context.Context, id uuid.UUID) (entity.Operador, error)
	BuscarPorLogin(ctx context.Context, login string) (entity.Operador, error)
	Listar(ctx context.Context) ([]entity.Operador, error) // por Login
	CriarSeVazio(ctx context.Context, o entity.Operador) (bool, error)
}

// AuditoriaRepository [ADIÇÃO v1.5]. Registrar só inclui (ID repetido falha); Listar devolve
// da Data mais recente para a mais antiga (desempate pela gravação mais recente), com o total;
// ExpurgarAntesDe apaga Data < limite. Spec: 3.6, RN-13, Q-39.
type AuditoriaRepository interface {
	Registrar(ctx context.Context, r entity.RegistroAuditoria) error
	Listar(ctx context.Context, f FiltroAuditoria) ([]entity.RegistroAuditoria, int, error)
	ExpurgarAntesDe(ctx context.Context, limite time.Time) (int, error)
}

// FiltroAuditoria [ADIÇÃO v1.5]: campos vazios ou zero não filtram; De <= Data < Ate.
// Spec: 3.6, Q-39.
type FiltroAuditoria struct {
	Operador     string
	Acao         entity.AcaoAuditoria
	De, Ate      time.Time
	Deslocamento int
	Limite       int // 0 = sem limite
}

// HashSenha [ADIÇÃO v1.5] gera e confere o hash da senha do operador. Sem ctx: não faz I/O
// (C-07). Spec: 3.6, Q-36.
type HashSenha interface {
	Gerar(senha string) (string, error)
	Conferir(hash, senha string) bool
}

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
