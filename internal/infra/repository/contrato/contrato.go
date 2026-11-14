// Package contrato define a suíte de contrato dos sete ports de repositório: o comportamento
// observável que toda implementação (fakes em memória e PostgreSQL) precisa ter para que os
// casos de uso funcionem igual nos dois backends. Usado apenas por testes.
//
// Spec: T-19, 3.6, RN-05, RN-06, RN-08, D-02. Fonte: UML7 (ports).
package contrato

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/shared/errs"
)

// Repositorios é um backend completo e vazio. Segmentos e produtos não têm escrita no port
// (entram por migração), então o backend informa como inseri-los para o teste.
type Repositorios struct {
	Clientes       port.ClienteRepository
	Eventos        port.EventoRepository
	Perfis         port.PerfilRepository
	Segmentos      port.SegmentoRepository
	Recomendacoes  port.RecomendacaoRepository
	Consentimentos port.ConsentimentoRepository
	Produtos       port.ProdutoRepository
	Publicos       port.PublicoRepository // v1.9, RN-17

	AdicionarSegmento func(t *testing.T, s entity.Segmento)
	AdicionarProduto  func(t *testing.T, p entity.Produto)
}

// Fabrica devolve um backend vazio a cada chamada.
type Fabrica func(t *testing.T) Repositorios

// agora tem precisão de microssegundo, a do timestamptz.
var agora = time.Date(2026, 11, 10, 12, 30, 45, 123456000, time.UTC)

// Executar roda todos os casos do contrato contra o backend criado por nova.
func Executar(t *testing.T, nova Fabrica) {
	casos := []struct {
		nome string
		fn   func(*testing.T, Repositorios)
	}{
		{"Cliente", testarCliente},
		{"Evento", testarEvento},
		{"Perfil", testarPerfil},
		{"Segmento", testarSegmento},
		{"Recomendacao", testarRecomendacao},
		{"Consentimento", testarConsentimento},
		{"MovimentoConsentimento", testarMovimentos},
		{"Produto", testarProduto},
		{"PerfilPorSegmento", testarPerfisPorSegmento},
		{"Publico", testarPublico},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) { c.fn(t, nova(t)) })
	}
}

func cliente(t *testing.T, r Repositorios) entity.Cliente {
	t.Helper()
	c := entity.Cliente{ID: uuid.New(), Nome: "Cliente Sintético", Email: "c-001@exemplo.test", DataCadastro: agora, Ativo: true}
	ok(t, r.Clientes.Salvar(context.Background(), c))
	return c
}

func ok(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func igual(t *testing.T, oque string, obtido, esperado interface{}) {
	t.Helper()
	if !reflect.DeepEqual(obtido, esperado) {
		t.Fatalf("%s:\n obtido   %+v\n esperado %+v", oque, obtido, esperado)
	}
}

func erroEsperado(t *testing.T, err, alvo error) {
	t.Helper()
	if !errors.Is(err, alvo) {
		t.Fatalf("erro %v, esperado %v", err, alvo)
	}
}

func ptr(t time.Time) *time.Time { return &t }

// Spec: UC-09, 3.6.
func testarCliente(t *testing.T, r Repositorios) {
	ctx := context.Background()
	_, err := r.Clientes.BuscarPorID(ctx, uuid.New())
	erroEsperado(t, err, errs.ErrClienteNaoEncontrado)

	c := cliente(t, r)
	obtido, err := r.Clientes.BuscarPorID(ctx, c.ID)
	ok(t, err)
	igual(t, "cliente salvo", obtido, c)

	c.Ativo = false
	c.Nome = "Outro Nome Sintético"
	ok(t, r.Clientes.Salvar(ctx, c))
	obtido, err = r.Clientes.BuscarPorID(ctx, c.ID)
	ok(t, err)
	igual(t, "cliente atualizado", obtido, c)
	testarListarClientes(t, r, c)
}

// Spec: Q-23 — busca por nome (sem maiúsculas nem acentos) ou ID exato, paginação e ordem.
func testarListarClientes(t *testing.T, r Repositorios, existente entity.Cliente) {
	ctx := context.Background()
	novo := func(nome string, dias int) entity.Cliente {
		c := entity.Cliente{ID: uuid.New(), Nome: nome, Email: "c-100@exemplo.test", DataCadastro: agora.AddDate(0, 0, dias), Ativo: true}
		ok(t, r.Clientes.Salvar(ctx, c))
		return c
	}
	joana := novo("Joana Sintética", 1)
	ines := novo("INÊS SINTÉTICA", 2)
	novo("Pedro 100% Sintético", 3)
	todos, total, err := r.Clientes.Listar(ctx, port.FiltroClientes{})
	ok(t, err)
	igual(t, "total sem filtro", total, 4)
	igual(t, "mais recente primeiro", todos[1].ID, ines.ID)
	igual(t, "mais antigo por último", todos[3].ID, existente.ID)

	pagina, total, err := r.Clientes.Listar(ctx, port.FiltroClientes{Deslocamento: 1, Limite: 2})
	ok(t, err)
	igual(t, "total paginado", total, 4)
	igual(t, "página", []uuid.UUID{pagina[0].ID, pagina[1].ID}, []uuid.UUID{ines.ID, joana.ID})

	achados, total, err := r.Clientes.Listar(ctx, port.FiltroClientes{Busca: "sintetica"})
	ok(t, err)
	igual(t, "busca sem acento", total, 2)
	igual(t, "ordem da busca", []uuid.UUID{achados[0].ID, achados[1].ID}, []uuid.UUID{ines.ID, joana.ID})

	_, total, err = r.Clientes.Listar(ctx, port.FiltroClientes{Busca: "%"})
	ok(t, err)
	igual(t, "% é literal", total, 1)

	porID, total, err := r.Clientes.Listar(ctx, port.FiltroClientes{Busca: " " + joana.ID.String() + " "})
	ok(t, err)
	igual(t, "busca por ID", total, 1)
	igual(t, "cliente por ID", porID[0], joana)

	vazio, total, err := r.Clientes.Listar(ctx, port.FiltroClientes{Busca: "inexistente"})
	ok(t, err)
	igual(t, "sem resultado", total, 0)
	igual(t, "lista vazia", len(vazio), 0)
}

// Spec: UC-01, RN-04, D-05.
func testarEvento(t *testing.T, r Repositorios) {
	ctx := context.Background()
	c1, c2 := cliente(t, r), cliente(t, r)
	vazio, err := r.Eventos.BuscarPorCliente(ctx, c1.ID)
	ok(t, err)
	igual(t, "sem eventos", len(vazio), 0)

	esperados := []entity.EventoComportamental{
		{ID: uuid.New(), ClienteID: c1.ID, Tipo: entity.TipoCompra, Origem: "ecommerce", Timestamp: agora.Add(-time.Hour), Valor: 520.5, Categoria: "livros"},
		{ID: uuid.New(), ClienteID: c1.ID, Tipo: entity.TipoNavegacao, Origem: "app", Timestamp: agora, Valor: 0},
	}
	for _, e := range esperados {
		ok(t, r.Eventos.Salvar(ctx, e))
	}
	ok(t, r.Eventos.Salvar(ctx, entity.EventoComportamental{ID: uuid.New(), ClienteID: c2.ID, Tipo: entity.TipoInteracao, Origem: "app", Timestamp: agora}))

	obtidos, err := r.Eventos.BuscarPorCliente(ctx, c1.ID)
	ok(t, err)
	igual(t, "eventos do cliente", porID(obtidos), porID(esperados))
}

func porID(es []entity.EventoComportamental) []entity.EventoComportamental {
	out := append([]entity.EventoComportamental(nil), es...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID.String() < out[j].ID.String() })
	return out
}

// Spec: RN-05 (upsert por ClienteID preservando o ID), D-06.
func testarPerfil(t *testing.T, r Repositorios) {
	ctx := context.Background()
	c := cliente(t, r)
	_, err := r.Perfis.BuscarPorCliente(ctx, c.ID)
	erroEsperado(t, err, errs.ErrPerfilNaoEncontrado)

	p := entity.PerfilCliente{ID: uuid.New(), ClienteID: c.ID, FrequenciaCompras: 3, TicketMedio: 610, CategoriaPreferida: "livros", ScoreEngajamento: 0.3}
	salvo, err := r.Perfis.Salvar(ctx, p)
	ok(t, err)
	igual(t, "primeiro salvamento", salvo, p)

	seg := entity.Segmento{ID: uuid.New(), Nome: "premium", Criterio: "ticketMedio >= 500", Prioridade: 1, Ativo: true}
	r.AdicionarSegmento(t, seg)
	novo := entity.PerfilCliente{ID: uuid.New(), ClienteID: c.ID, FrequenciaCompras: 4, TicketMedio: 700, CategoriaPreferida: "games", ScoreEngajamento: 1, SegmentoID: &seg.ID}
	salvo, err = r.Perfis.Salvar(ctx, novo)
	ok(t, err)
	novo.ID = p.ID // o ID original é preservado
	igual(t, "upsert", salvo, novo)

	obtido, err := r.Perfis.BuscarPorCliente(ctx, c.ID)
	ok(t, err)
	igual(t, "perfil lido", obtido, novo)
}

// Spec: RN-06 (ativos em ordem crescente de Prioridade).
func testarSegmento(t *testing.T, r Repositorios) {
	ctx := context.Background()
	padrao := entity.Segmento{ID: uuid.New(), Nome: "padrao", Descricao: "segmento padrão", Prioridade: 99, Ativo: true}
	premium := entity.Segmento{ID: uuid.New(), Nome: "premium", Descricao: "alto valor", Criterio: "ticketMedio >= 500", Prioridade: 1, Ativo: true}
	inativo := entity.Segmento{ID: uuid.New(), Nome: "antigo", Criterio: "frequenciaCompras > 1", Prioridade: 0, Ativo: false}
	for _, s := range []entity.Segmento{padrao, inativo, premium} {
		r.AdicionarSegmento(t, s)
	}
	ativos, err := r.Segmentos.ListarAtivos(ctx)
	ok(t, err)
	igual(t, "ativos por prioridade", ativos, []entity.Segmento{premium, padrao})

	obtido, err := r.Segmentos.BuscarPorID(ctx, inativo.ID)
	ok(t, err)
	igual(t, "busca por ID", obtido, inativo)
	_, err = r.Segmentos.BuscarPorID(ctx, uuid.New())
	erroEsperado(t, err, errs.ErrDadosInvalidos)
}

// Spec: RN-08 (ATIVA com Validade <= agora), D-08.
func testarRecomendacao(t *testing.T, r Repositorios) {
	ctx := context.Background()
	c := cliente(t, r)
	prod := entity.Produto{ID: uuid.New(), Nome: "Livro A", Categoria: "livros", Preco: 80, Ativo: true}
	r.AdicionarProduto(t, prod)

	nova := func(status entity.StatusRecomendacao, validade time.Time) entity.RecomendacaoProduto {
		rec := entity.RecomendacaoProduto{ID: uuid.New(), ClienteID: c.ID, ProdutoID: prod.ID, Motivo: "segmento:premium;categoria:livros",
			Status: status, DataGeracao: validade.AddDate(0, 0, -30), DataAtivacao: ptr(validade.AddDate(0, 0, -30)), Validade: ptr(validade)}
		if status == entity.StatusExpirada {
			rec.DataExpiracao = ptr(validade)
		}
		ok(t, r.Recomendacoes.Salvar(ctx, rec))
		return rec
	}
	vencida := nova(entity.StatusAtiva, agora.Add(-time.Hour))
	noLimite := nova(entity.StatusAtiva, agora)
	nova(entity.StatusAtiva, agora.Add(time.Second))
	nova(entity.StatusExpirada, agora.Add(-48*time.Hour))

	lista, err := r.Recomendacoes.ListarAtivasVencidas(ctx, agora)
	ok(t, err)
	igual(t, "vencidas", porIDRec(lista), porIDRec([]entity.RecomendacaoProduto{vencida, noLimite}))

	vencida.Status, vencida.DataExpiracao = entity.StatusExpirada, ptr(agora)
	ok(t, r.Recomendacoes.Salvar(ctx, vencida)) // atualização por ID
	lista, err = r.Recomendacoes.ListarAtivasVencidas(ctx, agora)
	ok(t, err)
	igual(t, "após expirar", lista, []entity.RecomendacaoProduto{noLimite})

	// Q-18: ATIVA do cliente, de qualquer validade; EXPIRADA e de outros clientes ficam de fora.
	outro := cliente(t, r)
	ok(t, r.Recomendacoes.Salvar(ctx, entity.RecomendacaoProduto{ID: uuid.New(), ClienteID: outro.ID, ProdutoID: prod.ID,
		Motivo: "segmento:padrao;categoria:livros", Status: entity.StatusAtiva, DataGeracao: agora}))
	ativas, err := r.Recomendacoes.ListarAtivasPorCliente(ctx, c.ID)
	ok(t, err)
	igual(t, "ativas do cliente", len(ativas), 2) // noLimite e a de validade futura
	for _, a := range ativas {
		if a.Status != entity.StatusAtiva || a.ClienteID != c.ID {
			t.Fatalf("ativa inesperada: %+v", a)
		}
	}

	// Q-23: histórico completo do cliente, a mais recente primeiro.
	historico, err := r.Recomendacoes.ListarPorCliente(ctx, c.ID)
	ok(t, err)
	igual(t, "histórico do cliente", len(historico), 4)
	for i := 1; i < len(historico); i++ {
		if historico[i].DataGeracao.After(historico[i-1].DataGeracao) {
			t.Fatalf("histórico fora de ordem: %v antes de %v", historico[i-1].DataGeracao, historico[i].DataGeracao)
		}
	}
	igual(t, "mais recente primeiro", historico[0].Validade, ptr(agora.Add(time.Second)))
}

func porIDRec(rs []entity.RecomendacaoProduto) []entity.RecomendacaoProduto {
	out := append([]entity.RecomendacaoProduto(nil), rs...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID.String() < out[j].ID.String() })
	return out
}

// Spec: UC-08, 3.4.1, D-02, Q-01 (DataRevogacao mantida).
func testarConsentimento(t *testing.T, r Repositorios) {
	ctx := context.Background()
	c := cliente(t, r)
	seg, pub := entity.FinalidadeSegmentacao, entity.FinalidadePublicidadeTerceiros
	_, err := r.Consentimentos.BuscarPorCliente(ctx, c.ID, seg)
	erroEsperado(t, err, errs.ErrConsentimentoNaoEncontrado)
	lista, err := r.Consentimentos.ListarPorCliente(ctx, c.ID)
	ok(t, err)
	igual(t, "sem registros", lista, []entity.ConsentimentoLGPD{})

	cons := entity.NovoConsentimento(uuid.New(), c.ID, seg, agora)
	ok(t, r.Consentimentos.Salvar(ctx, cons))
	obtido, err := r.Consentimentos.BuscarPorCliente(ctx, c.ID, seg)
	ok(t, err)
	igual(t, "NAO_CONCEDIDO", obtido, cons)

	ok(t, cons.Conceder(seg, "v1", agora.Add(time.Minute)))
	ok(t, cons.Revogar(agora.Add(2*time.Minute)))
	ok(t, cons.Conceder(seg, "v2", agora.Add(3*time.Minute)))
	original := cons.ID
	cons.ID = uuid.New() // upsert por (ClienteID, Finalidade) preserva o ID original
	ok(t, r.Consentimentos.Salvar(ctx, cons))
	obtido, err = r.Consentimentos.BuscarPorCliente(ctx, c.ID, seg)
	ok(t, err)
	cons.ID = original
	igual(t, "CONCEDIDO após revogação", obtido, cons)
	movs, err := r.Consentimentos.ListarMovimentos(ctx, c.ID)
	ok(t, err)
	igual(t, "Salvar não gera movimento", movs, []entity.MovimentoConsentimento{})

	// Q-48: a outra finalidade é outro registro, independente.
	_, err = r.Consentimentos.BuscarPorCliente(ctx, c.ID, pub)
	erroEsperado(t, err, errs.ErrConsentimentoNaoEncontrado)
	publicidade := entity.NovoConsentimento(uuid.New(), c.ID, pub, agora)
	ok(t, publicidade.Conceder(pub, "v1", agora.Add(4*time.Minute)))
	ok(t, r.Consentimentos.Salvar(ctx, publicidade))
	obtido, err = r.Consentimentos.BuscarPorCliente(ctx, c.ID, seg)
	ok(t, err)
	igual(t, "segmentação intacta", obtido, cons)
	lista, err = r.Consentimentos.ListarPorCliente(ctx, c.ID)
	ok(t, err)
	igual(t, "os dois registros, por finalidade", lista, []entity.ConsentimentoLGPD{publicidade, cons})
}

// Spec: RN-10, Q-34 — registro e movimento gravados juntos; histórico em ordem cronológica.
func testarMovimentos(t *testing.T, r Repositorios) {
	ctx := context.Background()
	c, outro := cliente(t, r), cliente(t, r)
	seg := entity.FinalidadeSegmentacao
	cons := entity.NovoConsentimento(uuid.New(), c.ID, seg, agora)
	var esperados []entity.MovimentoConsentimento
	registrar := func(tipo entity.TipoMovimento, origem entity.OrigemMovimento, operador string) {
		t.Helper()
		m, err := entity.NovoMovimento(uuid.New(), cons, tipo, origem, operador)
		ok(t, err)
		ok(t, r.Consentimentos.Registrar(ctx, cons, m))
		esperados = append(esperados, m)
	}
	ok(t, cons.Conceder(seg, "v1", agora.Add(time.Minute)))
	registrar(entity.MovimentoConcessao, entity.OrigemAPI, "")
	ok(t, cons.Revogar(agora.Add(2*time.Minute)))
	registrar(entity.MovimentoRevogacao, entity.OrigemPainel, "op-01")
	ok(t, cons.Conceder(seg, "v2", agora.Add(2*time.Minute))) // mesma data: vale a ordem de gravação
	registrar(entity.MovimentoConcessao, entity.OrigemPortal, "")

	obtido, err := r.Consentimentos.BuscarPorCliente(ctx, c.ID, seg)
	ok(t, err)
	igual(t, "registro gravado por Registrar", obtido, cons)
	movs, err := r.Consentimentos.ListarMovimentos(ctx, c.ID)
	ok(t, err)
	igual(t, "histórico", movs, esperados)
	vazio, err := r.Consentimentos.ListarMovimentos(ctx, outro.ID)
	ok(t, err)
	igual(t, "histórico de outro cliente", vazio, []entity.MovimentoConsentimento{})

	// Atomicidade: movimento com ID repetido falha e não altera o registro nem o histórico.
	antes := cons
	ok(t, cons.Revogar(agora.Add(time.Hour)))
	repetido, err := entity.NovoMovimento(esperados[0].ID, cons, entity.MovimentoRevogacao, entity.OrigemAPI, "")
	ok(t, err)
	if err := r.Consentimentos.Registrar(ctx, cons, repetido); err == nil {
		t.Fatal("Registrar com ID de movimento repetido deveria falhar")
	}
	obtido, err = r.Consentimentos.BuscarPorCliente(ctx, c.ID, seg)
	ok(t, err)
	igual(t, "registro depois da falha", obtido, antes)
	movs, err = r.Consentimentos.ListarMovimentos(ctx, c.ID)
	ok(t, err)
	igual(t, "histórico depois da falha", movs, esperados)
}

// Spec: D-05, RN-07.
func testarProduto(t *testing.T, r Repositorios) {
	ctx := context.Background()
	a := entity.Produto{ID: uuid.New(), Nome: "Livro A", Categoria: "livros", Preco: 80, Ativo: true}
	b := entity.Produto{ID: uuid.New(), Nome: "Livro B", Categoria: "livros", Preco: 49.9, Ativo: true}
	for _, p := range []entity.Produto{
		b, a,
		{ID: uuid.New(), Nome: "Livro C", Categoria: "livros", Preco: 10, Ativo: false},
		{ID: uuid.New(), Nome: "Jogo X", Categoria: "games", Preco: 200, Ativo: true},
	} {
		r.AdicionarProduto(t, p)
	}
	lista, err := r.Produtos.ListarAtivosPorCategoria(ctx, "livros")
	ok(t, err)
	sort.Slice(lista, func(i, j int) bool { return lista[i].Nome < lista[j].Nome })
	igual(t, "ativos da categoria", lista, []entity.Produto{a, b})

	vazio, err := r.Produtos.ListarAtivosPorCategoria(ctx, "inexistente")
	ok(t, err)
	igual(t, "categoria sem produtos", len(vazio), 0)

	// Q-23: catálogo completo, inativos incluídos, por categoria e nome.
	todos, err := r.Produtos.ListarTodos(ctx)
	ok(t, err)
	nomes := []string{}
	for _, p := range todos {
		nomes = append(nomes, p.Categoria+"/"+p.Nome)
	}
	igual(t, "catálogo completo", nomes, []string{"games/Jogo X", "livros/Livro A", "livros/Livro B", "livros/Livro C"})
}
