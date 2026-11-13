package usecase

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/domain/service"
	"perfila/internal/shared/errs"
)

// ---------- Leituras da configuração (v1.6) ----------

// Todos devolve os segmentos ativos em ordem de Prioridade e depois os inativos por Nome.
// Spec: 3.5, RN-14.
func (uc *ListarSegmentosUseCase) Todos(ctx context.Context) ([]entity.Segmento, error) {
	lista, err := uc.segmentos.ListarTodos(ctx)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(lista, func(i, j int) bool {
		a, b := lista[i], lista[j]
		if a.Ativo != b.Ativo {
			return a.Ativo
		}
		if a.Ativo {
			return a.Prioridade < b.Prioridade
		}
		return strings.ToLower(a.Nome) < strings.ToLower(b.Nome)
	})
	return lista, nil
}

// Consultar devolve um segmento, ativo ou não; ErrSegmentoNaoEncontrado. Spec: 3.5.
func (uc *ListarSegmentosUseCase) Consultar(ctx context.Context, id uuid.UUID) (entity.Segmento, error) {
	lista, err := uc.segmentos.ListarTodos(ctx)
	if err != nil {
		return entity.Segmento{}, err
	}
	for _, s := range lista {
		if s.ID == id {
			return s, nil
		}
	}
	return entity.Segmento{}, errs.ErrSegmentoNaoEncontrado
}

// Consultar devolve um produto; ErrProdutoNaoEncontrado. Spec: 3.5.
func (uc *ListarCatalogoUseCase) Consultar(ctx context.Context, id uuid.UUID) (entity.Produto, error) {
	return uc.produtos.BuscarPorID(ctx, id)
}

// Categorias devolve as categorias distintas do catálogo, em ordem alfabética. Spec: 3.5, Q-43.
func (uc *ListarCatalogoUseCase) Categorias(ctx context.Context) ([]string, error) {
	lista, err := uc.produtos.ListarTodos(ctx)
	if err != nil {
		return nil, err
	}
	vistas := map[string]bool{}
	out := []string{}
	for _, p := range lista {
		if !vistas[p.Categoria] {
			vistas[p.Categoria] = true
			out = append(out, p.Categoria)
		}
	}
	sort.Strings(out)
	return out, nil
}

// ---------- Segmentos ----------

// AlteracaoSegmento é o resultado de salvar: o segmento gravado, se é novo e os nomes dos campos
// que mudaram (para a auditoria, Q-44). Spec: 3.5, RN-14.
type AlteracaoSegmento struct {
	Segmento entity.Segmento
	Novo     bool
	Campos   []string
}

// SalvarSegmentoUseCase cria, edita, ativa e desativa segmentos (permissão editar_configuracao).
// Spec: 3.5, RN-11, RN-14, Q-33.
type SalvarSegmentoUseCase struct{ segmentos port.SegmentoRepository }

// NovoSalvarSegmentoUseCase injeta o repositório (C-08).
func NovoSalvarSegmentoUseCase(segmentos port.SegmentoRepository) *SalvarSegmentoUseCase {
	return &SalvarSegmentoUseCase{segmentos: segmentos}
}

// Executar cria (ID nulo, nasce ativo) ou edita nome, descrição, critério e prioridade, mantendo a
// situação. Sem mudança, nada é gravado. Spec: RN-14.
func (uc *SalvarSegmentoUseCase) Executar(ctx context.Context, ator entity.Operador, s entity.Segmento) (AlteracaoSegmento, error) {
	return uc.salvar(ctx, ator, s, false)
}

// Ativar devolve o segmento à avaliação; exige prioridade livre. Spec: RN-14.
func (uc *SalvarSegmentoUseCase) Ativar(ctx context.Context, ator entity.Operador, id uuid.UUID) (AlteracaoSegmento, error) {
	return uc.mudarSituacao(ctx, ator, id, true)
}

// Desativar tira o segmento da avaliação; o padrao é recusado. Spec: RN-14.
func (uc *SalvarSegmentoUseCase) Desativar(ctx context.Context, ator entity.Operador, id uuid.UUID) (AlteracaoSegmento, error) {
	return uc.mudarSituacao(ctx, ator, id, false)
}

func (uc *SalvarSegmentoUseCase) mudarSituacao(ctx context.Context, ator entity.Operador, id uuid.UUID, ativo bool) (AlteracaoSegmento, error) {
	if err := exigir(ator, entity.PermEditarConfiguracao); err != nil {
		return AlteracaoSegmento{}, err
	}
	lista, err := uc.segmentos.ListarTodos(ctx)
	if err != nil {
		return AlteracaoSegmento{}, err
	}
	atual, ok := acharSegmento(lista, id)
	if !ok {
		return AlteracaoSegmento{}, errs.ErrSegmentoNaoEncontrado
	}
	atual.Ativo = ativo
	return uc.salvar(ctx, ator, atual, true)
}

// salvar aplica a mudança à configuração inteira, valida (RN-14) e grava. comSituacao diz se o
// campo Ativo de s vale (ativar e desativar) ou se a situação gravada é mantida (edição).
func (uc *SalvarSegmentoUseCase) salvar(ctx context.Context, ator entity.Operador, s entity.Segmento, comSituacao bool) (AlteracaoSegmento, error) {
	if err := exigir(ator, entity.PermEditarConfiguracao); err != nil {
		return AlteracaoSegmento{}, err
	}
	lista, err := uc.segmentos.ListarTodos(ctx)
	if err != nil {
		return AlteracaoSegmento{}, err
	}
	nova, alvo, anterior, err := novaConfiguracao(lista, s, comSituacao)
	if err != nil {
		return AlteracaoSegmento{}, err
	}
	if err := entity.ValidarSegmentos(nova); err != nil {
		return AlteracaoSegmento{}, err
	}
	alt := AlteracaoSegmento{Segmento: alvo, Novo: anterior == nil}
	if anterior == nil {
		alt.Campos = []string{"nome", "descricao", "criterio", "prioridade", "ativo"}
	} else {
		alt.Campos = camposSegmento(*anterior, alvo)
	}
	if len(alt.Campos) == 0 {
		return alt, nil
	}
	if err := uc.segmentos.Salvar(ctx, alvo); err != nil {
		return AlteracaoSegmento{}, err
	}
	return alt, nil
}

// novaConfiguracao devolve a lista com a mudança aplicada, o segmento como ficaria e o gravado
// (nil se novo). Tira os espaços das pontas dos textos. Spec: RN-14.
func novaConfiguracao(lista []entity.Segmento, s entity.Segmento, comSituacao bool) ([]entity.Segmento, entity.Segmento, *entity.Segmento, error) {
	s.Nome, s.Descricao, s.Criterio = strings.TrimSpace(s.Nome), strings.TrimSpace(s.Descricao), strings.TrimSpace(s.Criterio)
	nova := make([]entity.Segmento, 0, len(lista)+1)
	if s.ID == uuid.Nil {
		s.ID, s.Ativo = uuid.New(), true
		return append(append(nova, lista...), s), s, nil, nil
	}
	atual, ok := acharSegmento(lista, s.ID)
	if !ok {
		return nil, entity.Segmento{}, nil, errs.ErrSegmentoNaoEncontrado
	}
	if !comSituacao {
		s.Ativo = atual.Ativo
	}
	for _, x := range lista {
		if x.ID == s.ID {
			x = s
		}
		nova = append(nova, x)
	}
	return nova, s, &atual, nil
}

func acharSegmento(lista []entity.Segmento, id uuid.UUID) (entity.Segmento, bool) {
	for _, s := range lista {
		if s.ID == id {
			return s, true
		}
	}
	return entity.Segmento{}, false
}

func camposSegmento(a, b entity.Segmento) []string {
	var c []string
	if a.Nome != b.Nome {
		c = append(c, "nome")
	}
	if a.Descricao != b.Descricao {
		c = append(c, "descricao")
	}
	if a.Criterio != b.Criterio {
		c = append(c, "criterio")
	}
	if a.Prioridade != b.Prioridade {
		c = append(c, "prioridade")
	}
	if a.Ativo != b.Ativo {
		c = append(c, "ativo")
	}
	return c
}

// ---------- Prévia ----------

// LinhaPrevia é a contagem de um segmento: quantos estão nele hoje e quantos ficariam. Spec: Q-42.
type LinhaPrevia struct {
	Segmento entity.Segmento
	Atual    int
	Previsto int
}

// PreviaSegmentacao só tem números: nunca nome, e-mail ou ID de cliente (C-09, Q-42).
type PreviaSegmentacao struct {
	Linhas    []LinhaPrevia
	SemPerfil int
	Total     int
}

// PreverSegmentacaoUseCase mostra o efeito de uma mudança antes de salvar. Spec: 3.5, RN-14, Q-42.
type PreverSegmentacaoUseCase struct {
	segmentos      port.SegmentoRepository
	consentimentos port.ConsentimentoRepository
	clientes       port.ClienteRepository
	perfis         port.PerfilRepository
}

// NovoPreverSegmentacaoUseCase injeta as dependências (C-08).
func NovoPreverSegmentacaoUseCase(segmentos port.SegmentoRepository, consentimentos port.ConsentimentoRepository,
	clientes port.ClienteRepository, perfis port.PerfilRepository) *PreverSegmentacaoUseCase {
	return &PreverSegmentacaoUseCase{segmentos: segmentos, consentimentos: consentimentos, clientes: clientes, perfis: perfis}
}

// Executar valida como SalvarSegmento (s.Ativo vale: prevê também ativar e desativar) e conta,
// pelos perfis salvos dos clientes ativos com consentimento válido, quantos ficam em cada
// segmento. Começa pelo consentimento: nenhum perfil de quem não consentiu é lido (C-03).
// Nada é gravado. Spec: RN-14, Q-42.
func (uc *PreverSegmentacaoUseCase) Executar(ctx context.Context, ator entity.Operador, s entity.Segmento) (PreviaSegmentacao, error) {
	if err := exigir(ator, entity.PermEditarConfiguracao); err != nil {
		return PreviaSegmentacao{}, err
	}
	lista, err := uc.segmentos.ListarTodos(ctx)
	if err != nil {
		return PreviaSegmentacao{}, err
	}
	nova, _, _, err := novaConfiguracao(lista, s, true)
	if err != nil {
		return PreviaSegmentacao{}, err
	}
	if err := entity.ValidarSegmentos(nova); err != nil {
		return PreviaSegmentacao{}, err
	}
	atual, previsto := map[uuid.UUID]int{}, map[uuid.UUID]int{}
	var previa PreviaSegmentacao
	concedidos, err := uc.consentimentos.ListarConcedidos(ctx)
	if err != nil {
		return PreviaSegmentacao{}, err
	}
	for _, c := range concedidos {
		if !c.EstaValido() {
			continue
		}
		cli, err := uc.clientes.BuscarPorID(ctx, c.ClienteID)
		if errors.Is(err, errs.ErrClienteNaoEncontrado) {
			continue
		}
		if err != nil {
			return PreviaSegmentacao{}, err
		}
		if !cli.Ativo {
			continue
		}
		previa.Total++
		p, err := uc.perfis.BuscarPorCliente(ctx, c.ClienteID)
		if errors.Is(err, errs.ErrPerfilNaoEncontrado) {
			previa.SemPerfil++
			continue
		}
		if err != nil {
			return PreviaSegmentacao{}, err
		}
		if p.SegmentoID != nil {
			atual[*p.SegmentoID]++
		}
		seg, err := service.EscolherSegmento(nova, p)
		if err != nil {
			return PreviaSegmentacao{}, err
		}
		previsto[seg.ID]++
	}
	previa.Linhas = linhasPrevia(lista, nova, atual, previsto)
	return previa, nil
}

// linhasPrevia: os ativos da nova configuração por Prioridade, depois os que deixam de ser ativos.
func linhasPrevia(antes, depois []entity.Segmento, atual, previsto map[uuid.UUID]int) []LinhaPrevia {
	ativos := []entity.Segmento{}
	for _, s := range depois {
		if s.Ativo {
			ativos = append(ativos, s)
		}
	}
	sort.SliceStable(ativos, func(i, j int) bool { return ativos[i].Prioridade < ativos[j].Prioridade })
	for _, s := range antes {
		if novo, ok := acharSegmento(depois, s.ID); s.Ativo && ok && !novo.Ativo {
			ativos = append(ativos, novo)
		}
	}
	out := make([]LinhaPrevia, 0, len(ativos))
	for _, s := range ativos {
		out = append(out, LinhaPrevia{Segmento: s, Atual: atual[s.ID], Previsto: previsto[s.ID]})
	}
	return out
}

// TestarCriterio avalia o critério com valores digitados: não lê nem grava dado de cliente.
// Spec: RN-14, Q-42.
func (uc *PreverSegmentacaoUseCase) TestarCriterio(criterio string, frequenciaCompras int, ticketMedio, scoreEngajamento float64) (bool, error) {
	return entity.AvaliarCriterio(criterio, entity.PerfilCliente{
		FrequenciaCompras: frequenciaCompras, TicketMedio: ticketMedio, ScoreEngajamento: scoreEngajamento,
	})
}

// ---------- Produtos ----------

// AlteracaoProduto é o resultado de salvar um produto. Spec: 3.5, Q-43, Q-44.
type AlteracaoProduto struct {
	Produto entity.Produto
	Novo    bool
	Campos  []string
}

// SalvarProdutoUseCase cria, edita, ativa e desativa produtos (permissão editar_configuracao).
// Spec: 3.5, RN-11, RN-14, Q-43.
type SalvarProdutoUseCase struct{ produtos port.ProdutoRepository }

// NovoSalvarProdutoUseCase injeta o repositório (C-08).
func NovoSalvarProdutoUseCase(produtos port.ProdutoRepository) *SalvarProdutoUseCase {
	return &SalvarProdutoUseCase{produtos: produtos}
}

// Executar cria (ID nulo, nasce ativo) ou edita nome, categoria e preço (AtualizarPreco), mantendo
// a situação. Sem mudança, nada é gravado. Spec: RN-14, Q-43.
func (uc *SalvarProdutoUseCase) Executar(ctx context.Context, ator entity.Operador, p entity.Produto) (AlteracaoProduto, error) {
	if err := exigir(ator, entity.PermEditarConfiguracao); err != nil {
		return AlteracaoProduto{}, err
	}
	alvo := entity.Produto{ID: p.ID, Ativo: true}
	var anterior *entity.Produto
	if p.ID == uuid.Nil {
		alvo.ID = uuid.New()
	} else {
		atual, err := uc.produtos.BuscarPorID(ctx, p.ID)
		if err != nil {
			return AlteracaoProduto{}, err
		}
		alvo, anterior = atual, &atual
	}
	alvo.Nome, alvo.Categoria = strings.TrimSpace(p.Nome), entity.NormalizarCategoria(p.Categoria)
	if err := alvo.AtualizarPreco(p.Preco); err != nil {
		return AlteracaoProduto{}, err
	}
	return uc.gravar(ctx, alvo, anterior)
}

// Ativar devolve o produto à recomendação (RN-07). Spec: Q-43.
func (uc *SalvarProdutoUseCase) Ativar(ctx context.Context, ator entity.Operador, id uuid.UUID) (AlteracaoProduto, error) {
	return uc.mudarSituacao(ctx, ator, id, true)
}

// Desativar tira o produto da recomendação a partir do próximo processamento. Spec: Q-43.
func (uc *SalvarProdutoUseCase) Desativar(ctx context.Context, ator entity.Operador, id uuid.UUID) (AlteracaoProduto, error) {
	return uc.mudarSituacao(ctx, ator, id, false)
}

func (uc *SalvarProdutoUseCase) mudarSituacao(ctx context.Context, ator entity.Operador, id uuid.UUID, ativo bool) (AlteracaoProduto, error) {
	if err := exigir(ator, entity.PermEditarConfiguracao); err != nil {
		return AlteracaoProduto{}, err
	}
	atual, err := uc.produtos.BuscarPorID(ctx, id)
	if err != nil {
		return AlteracaoProduto{}, err
	}
	alvo := atual
	alvo.Ativo = ativo
	return uc.gravar(ctx, alvo, &atual)
}

func (uc *SalvarProdutoUseCase) gravar(ctx context.Context, alvo entity.Produto, anterior *entity.Produto) (AlteracaoProduto, error) {
	if err := alvo.Validar(); err != nil {
		return AlteracaoProduto{}, err
	}
	alt := AlteracaoProduto{Produto: alvo, Novo: anterior == nil}
	if anterior == nil {
		alt.Campos = []string{"nome", "categoria", "preco", "ativo"}
	} else {
		alt.Campos = camposProduto(*anterior, alvo)
	}
	if len(alt.Campos) == 0 {
		return alt, nil
	}
	if err := uc.produtos.Salvar(ctx, alvo); err != nil {
		return AlteracaoProduto{}, err
	}
	return alt, nil
}

func camposProduto(a, b entity.Produto) []string {
	var c []string
	if a.Nome != b.Nome {
		c = append(c, "nome")
	}
	if a.Ativo != b.Ativo {
		c = append(c, "ativo")
	}
	if a.Categoria != b.Categoria {
		c = append(c, "categoria")
	}
	if a.Preco != b.Preco {
		c = append(c, "preco")
	}
	return c
}
