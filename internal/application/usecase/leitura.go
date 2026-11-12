package usecase

import (
	"context"
	"errors"
	"sort"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/shared/errs"
)

// Casos de uso de leitura do painel do operador [ADIÇÃO]. Spec: 3.5, Q-21, Q-23.

// ClientesPorPagina é o tamanho da página da lista de clientes (Q-23).
const ClientesPorPagina = 20

// FiltroListagem é o que o operador escolhe na lista. Spec: Q-23.
type FiltroListagem struct {
	Busca         string                      // nome ou ID
	Consentimento *entity.StatusConsentimento // nil = todos
	Pagina        int                         // a partir de 1
}

// LinhaCliente é uma linha da lista. Segmento só vem com cliente ativo e consentimento válido (Q-21).
type LinhaCliente struct {
	Cliente       entity.Cliente
	Consentimento entity.StatusConsentimento
	Segmento      string
}

// PaginaClientes é uma página da lista com o total que casa com o filtro.
type PaginaClientes struct {
	Linhas []LinhaCliente
	Total  int
	Pagina int
}

// ListarClientesUseCase lista clientes com o estado do consentimento. Spec: 3.5, Q-21, Q-23.
type ListarClientesUseCase struct {
	clientes       port.ClienteRepository
	consentimentos port.ConsentimentoRepository
	perfis         port.PerfilRepository
	segmentos      port.SegmentoRepository
}

// NovoListarClientesUseCase injeta as dependências (C-08).
func NovoListarClientesUseCase(clientes port.ClienteRepository, consentimentos port.ConsentimentoRepository,
	perfis port.PerfilRepository, segmentos port.SegmentoRepository) *ListarClientesUseCase {
	return &ListarClientesUseCase{clientes: clientes, consentimentos: consentimentos, perfis: perfis, segmentos: segmentos}
}

// Executar devolve a página pedida. Com filtro de consentimento, o caso de uso filtra a busca
// inteira e pagina o resultado (volume do MVP). Spec: Q-23.
func (uc *ListarClientesUseCase) Executar(ctx context.Context, f FiltroListagem) (PaginaClientes, error) {
	pagina := max(f.Pagina, 1)
	desloc := (pagina - 1) * ClientesPorPagina
	filtro := port.FiltroClientes{Busca: f.Busca, Deslocamento: desloc, Limite: ClientesPorPagina}
	if f.Consentimento != nil {
		filtro = port.FiltroClientes{Busca: f.Busca}
	}
	clientes, total, err := uc.clientes.Listar(ctx, filtro)
	if err != nil {
		return PaginaClientes{}, err
	}
	linhas := make([]LinhaCliente, 0, len(clientes))
	for _, c := range clientes {
		status, err := uc.statusConsentimento(ctx, c.ID)
		if err != nil {
			return PaginaClientes{}, err
		}
		if f.Consentimento != nil && status != *f.Consentimento {
			continue
		}
		linhas = append(linhas, LinhaCliente{Cliente: c, Consentimento: status})
	}
	if f.Consentimento != nil {
		total = len(linhas)
		linhas = linhas[min(desloc, total):min(desloc+ClientesPorPagina, total)]
	}
	nomes := map[uuid.UUID]string{}
	for i, l := range linhas {
		if !l.Cliente.Ativo || l.Consentimento != entity.StatusConcedido {
			continue
		}
		if linhas[i].Segmento, err = uc.nomeDoSegmento(ctx, l.Cliente.ID, nomes); err != nil {
			return PaginaClientes{}, err
		}
	}
	return PaginaClientes{Linhas: linhas, Total: total, Pagina: pagina}, nil
}

func (uc *ListarClientesUseCase) statusConsentimento(ctx context.Context, clienteID uuid.UUID) (entity.StatusConsentimento, error) {
	c, err := uc.consentimentos.BuscarPorCliente(ctx, clienteID)
	if errors.Is(err, errs.ErrConsentimentoNaoEncontrado) {
		return entity.StatusNaoConcedido, nil // D-02
	}
	if err != nil {
		return "", err
	}
	return c.Status, nil
}

// nomeDoSegmento devolve "" se o cliente ainda não foi segmentado ou o segmento não existe mais.
func (uc *ListarClientesUseCase) nomeDoSegmento(ctx context.Context, clienteID uuid.UUID, cache map[uuid.UUID]string) (string, error) {
	p, err := uc.perfis.BuscarPorCliente(ctx, clienteID)
	if errors.Is(err, errs.ErrPerfilNaoEncontrado) || (err == nil && p.SegmentoID == nil) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if nome, ok := cache[*p.SegmentoID]; ok {
		return nome, nil
	}
	s, err := uc.segmentos.BuscarPorID(ctx, *p.SegmentoID)
	if errors.Is(err, errs.ErrDadosInvalidos) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	cache[s.ID] = s.Nome
	return s.Nome, nil
}

// FichaCliente é o cadastro e o consentimento: o que o painel mostra sem consentimento (Q-21).
type FichaCliente struct {
	Cliente       entity.Cliente
	Consentimento entity.ConsentimentoLGPD
}

// ConsultarClienteUseCase lê cadastro e consentimento sem exigir consentimento, como UC-08.
// Spec: 3.5, Q-21.
type ConsultarClienteUseCase struct {
	clientes      port.ClienteRepository
	consentimento *GerenciarConsentimentoUseCase
}

// NovoConsultarClienteUseCase injeta as dependências (C-08).
func NovoConsultarClienteUseCase(clientes port.ClienteRepository, consentimento *GerenciarConsentimentoUseCase) *ConsultarClienteUseCase {
	return &ConsultarClienteUseCase{clientes: clientes, consentimento: consentimento}
}

// Executar devolve ErrClienteNaoEncontrado se o cliente não existir. Spec: Q-21.
func (uc *ConsultarClienteUseCase) Executar(ctx context.Context, clienteID uuid.UUID) (FichaCliente, error) {
	c, err := uc.clientes.BuscarPorID(ctx, clienteID)
	if err != nil {
		return FichaCliente{}, err
	}
	cons, err := uc.consentimento.Consultar(ctx, clienteID)
	if err != nil {
		return FichaCliente{}, err
	}
	return FichaCliente{Cliente: c, Consentimento: cons}, nil
}

// ResultadoCliente é o comportamento do cliente: só sai com consentimento válido (RN-01).
type ResultadoCliente struct {
	Eventos       []entity.EventoComportamental // mais recente primeiro
	Perfil        *entity.PerfilCliente         // nil = ainda não processado
	Segmento      *entity.Segmento              // nil = ainda não segmentado
	Recomendacoes []entity.RecomendacaoProduto  // mais recente primeiro
	Produtos      map[uuid.UUID]entity.Produto  // os produtos citados nas recomendações
}

// ConsultarResultadoUseCase lê eventos, perfil, segmento e recomendações. Spec: 3.5, RN-01, C-03.
type ConsultarResultadoUseCase struct {
	pre           Precondicoes
	eventos       port.EventoRepository
	perfis        port.PerfilRepository
	segmentos     port.SegmentoRepository
	recomendacoes port.RecomendacaoRepository
	produtos      port.ProdutoRepository
}

// NovoConsultarResultadoUseCase injeta as dependências (C-08).
func NovoConsultarResultadoUseCase(pre Precondicoes, eventos port.EventoRepository, perfis port.PerfilRepository,
	segmentos port.SegmentoRepository, recomendacoes port.RecomendacaoRepository, produtos port.ProdutoRepository) *ConsultarResultadoUseCase {
	return &ConsultarResultadoUseCase{pre: pre, eventos: eventos, perfis: perfis, segmentos: segmentos,
		recomendacoes: recomendacoes, produtos: produtos}
}

// Executar confere cliente e consentimento antes de ler qualquer dado de comportamento; cliente
// inativo é só leitura (Q-31). Spec: RN-01, C-03.
func (uc *ConsultarResultadoUseCase) Executar(ctx context.Context, clienteID uuid.UUID) (ResultadoCliente, error) {
	if _, err := uc.pre.VerificarLeitura(ctx, clienteID); err != nil {
		return ResultadoCliente{}, err
	}
	var res ResultadoCliente
	var err error
	if res.Eventos, err = uc.eventos.BuscarPorCliente(ctx, clienteID); err != nil {
		return ResultadoCliente{}, err
	}
	sort.SliceStable(res.Eventos, func(i, j int) bool { return res.Eventos[i].Timestamp.After(res.Eventos[j].Timestamp) })

	p, err := uc.perfis.BuscarPorCliente(ctx, clienteID)
	switch {
	case err == nil:
		res.Perfil = &p
	case !errors.Is(err, errs.ErrPerfilNaoEncontrado):
		return ResultadoCliente{}, err
	}
	if res.Perfil != nil && res.Perfil.SegmentoID != nil {
		s, err := uc.segmentos.BuscarPorID(ctx, *res.Perfil.SegmentoID)
		switch {
		case err == nil:
			res.Segmento = &s
		case !errors.Is(err, errs.ErrDadosInvalidos):
			return ResultadoCliente{}, err
		}
	}
	if res.Recomendacoes, err = uc.recomendacoes.ListarPorCliente(ctx, clienteID); err != nil {
		return ResultadoCliente{}, err
	}
	res.Produtos = map[uuid.UUID]entity.Produto{}
	if len(res.Recomendacoes) > 0 {
		catalogo, err := uc.produtos.ListarTodos(ctx)
		if err != nil {
			return ResultadoCliente{}, err
		}
		citados := map[uuid.UUID]bool{}
		for _, r := range res.Recomendacoes {
			citados[r.ProdutoID] = true
		}
		for _, prod := range catalogo {
			if citados[prod.ID] {
				res.Produtos[prod.ID] = prod
			}
		}
	}
	return res, nil
}

// ListarSegmentosUseCase lista os segmentos ativos por Prioridade. Spec: 3.5, RN-06.
type ListarSegmentosUseCase struct{ segmentos port.SegmentoRepository }

// NovoListarSegmentosUseCase injeta o repositório (C-08).
func NovoListarSegmentosUseCase(segmentos port.SegmentoRepository) *ListarSegmentosUseCase {
	return &ListarSegmentosUseCase{segmentos: segmentos}
}

// Executar devolve os segmentos na ordem em que a segmentação os avalia.
func (uc *ListarSegmentosUseCase) Executar(ctx context.Context) ([]entity.Segmento, error) {
	return uc.segmentos.ListarAtivos(ctx)
}

// ListarCatalogoUseCase lista todos os produtos. Spec: 3.5, RN-07.
type ListarCatalogoUseCase struct{ produtos port.ProdutoRepository }

// NovoListarCatalogoUseCase injeta o repositório (C-08).
func NovoListarCatalogoUseCase(produtos port.ProdutoRepository) *ListarCatalogoUseCase {
	return &ListarCatalogoUseCase{produtos: produtos}
}

// Executar devolve o catálogo por Categoria e Nome.
func (uc *ListarCatalogoUseCase) Executar(ctx context.Context) ([]entity.Produto, error) {
	return uc.produtos.ListarTodos(ctx)
}
