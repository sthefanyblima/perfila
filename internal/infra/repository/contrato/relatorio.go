package contrato

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
)

// Spec: 3.6, RN-15, Q-46 — os agregados contam o comportamento só de quem tem consentimento
// CONCEDIDO; cadastro e estado do consentimento contam todos; intervalos [desde, ate).
func testarRelatorio(t *testing.T, r Repositorios) {
	ctx := context.Background()
	seg := entity.Segmento{ID: uuid.New(), Nome: "premium", Criterio: "ticketMedio >= 1", Prioridade: 1, Ativo: true}
	r.AdicionarSegmento(t, seg)
	prod := entity.Produto{ID: uuid.New(), Nome: "Livro A", Categoria: "livros", Preco: 80, Ativo: true}
	r.AdicionarProduto(t, prod)

	novo := func(ativo bool, estado entity.StatusConsentimento) uuid.UUID {
		c := entity.Cliente{ID: uuid.New(), Nome: "Cliente Sintético", Email: "c-001@exemplo.test", DataCadastro: agora, Ativo: ativo}
		ok(t, r.Clientes.Salvar(ctx, c))
		if estado != entity.StatusNaoConcedido {
			cons := entity.ConsentimentoLGPD{ID: uuid.New(), ClienteID: c.ID, Status: estado, Finalidade: "segmentacao",
				VersaoTermo: "v1", DataConsentimento: agora.Add(-3 * time.Hour), DataAtualizacao: agora.Add(-2 * time.Hour)}
			if estado == entity.StatusRevogado {
				cons.DataRevogacao = ptr(agora.Add(-2 * time.Hour))
				ok(t, r.Consentimentos.Registrar(ctx, cons, entity.MovimentoConsentimento{ID: uuid.New(), ClienteID: c.ID,
					Tipo: entity.MovimentoRevogacao, Finalidade: "segmentacao", VersaoTermo: "v1", Data: agora.Add(-2 * time.Hour),
					Origem: entity.OrigemSistema}))
			} else {
				ok(t, r.Consentimentos.Salvar(ctx, cons))
			}
		}
		return c.ID
	}
	perfil := func(cid uuid.UUID, segmento *uuid.UUID, ticket, score float64) {
		_, err := r.Perfis.Salvar(ctx, entity.PerfilCliente{ID: uuid.New(), ClienteID: cid, FrequenciaCompras: 1,
			TicketMedio: ticket, CategoriaPreferida: "livros", ScoreEngajamento: score, SegmentoID: segmento})
		ok(t, err)
	}
	rec := func(cid uuid.UUID, status entity.StatusRecomendacao) {
		v := agora.AddDate(0, 0, 30)
		x := entity.RecomendacaoProduto{ID: uuid.New(), ClienteID: cid, ProdutoID: prod.ID, Motivo: "segmento:premium;categoria:livros",
			Status: status, DataGeracao: agora, DataAtivacao: ptr(agora), Validade: &v}
		if status == entity.StatusExpirada {
			x.DataExpiracao = ptr(agora)
		}
		ok(t, r.Recomendacoes.Salvar(ctx, x))
	}
	evento := func(cid uuid.UUID, tipo entity.TipoEvento, categoria string, quando time.Time) {
		valor := 0.0
		if tipo == entity.TipoCompra {
			valor = 100
		}
		ok(t, r.Eventos.Salvar(ctx, entity.EventoComportamental{ID: uuid.New(), ClienteID: cid, Tipo: tipo, Origem: "loja on-line",
			Timestamp: quando, Valor: valor, Categoria: categoria}))
	}

	// c1: válido, segmentado; c2: revogado (fica de fora do comportamento); c3: sem consentimento e
	// inativo; c4: válido, sem segmento.
	c1 := novo(true, entity.StatusConcedido)
	perfil(c1, &seg.ID, 100, 0.5)
	rec(c1, entity.StatusAtiva)
	evento(c1, entity.TipoCompra, "livros", agora.Add(-time.Hour))
	evento(c1, entity.TipoCompra, "livros", agora.Add(-time.Hour))
	evento(c1, entity.TipoNavegacao, "", agora.Add(-time.Hour))
	evento(c1, entity.TipoCompra, "livros", agora.AddDate(0, 0, -40)) // fora do intervalo

	c2 := novo(true, entity.StatusRevogado)
	perfil(c2, &seg.ID, 900, 1)
	rec(c2, entity.StatusAtiva)
	evento(c2, entity.TipoCompra, "livros", agora.Add(-time.Hour))

	c3 := novo(false, entity.StatusNaoConcedido)
	perfil(c3, nil, 300, 1)
	evento(c3, entity.TipoCompra, "games", agora.Add(-time.Hour))

	c4 := novo(true, entity.StatusConcedido)
	perfil(c4, nil, 50, 0.25)
	rec(c4, entity.StatusExpirada)
	evento(c4, entity.TipoCompra, "games", agora.Add(-time.Hour))
	evento(c4, entity.TipoCompra, "", agora.Add(-time.Hour)) // sem categoria: não entra em compras

	clientes, err := r.Relatorios.ContarClientes(ctx)
	ok(t, err)
	igual(t, "clientes", clientes, port.ContagemClientes{Total: 4, Ativos: 3, Concedido: 2, NaoConcedido: 1, Revogado: 1})

	segs, err := r.Relatorios.ResumirSegmentos(ctx)
	ok(t, err)
	sort.Slice(segs, func(i, j int) bool { return segs[i].SegmentoID != nil && segs[j].SegmentoID == nil })
	igual(t, "segmentos", segs, []port.ResumoSegmento{
		{SegmentoID: &seg.ID, Clientes: 1, SomaTicket: 100, SomaScore: 0.5},
		{Clientes: 1, SomaTicket: 50, SomaScore: 0.25},
	})

	recs, err := r.Relatorios.ContarRecomendacoes(ctx)
	ok(t, err)
	igual(t, "recomendações", recs, port.ContagemRecomendacoes{Ativas: 1, ClientesAtivas: 1, Expiradas: 1, ClientesExpiradas: 1})

	desde, ate := agora.AddDate(0, 0, -1), agora
	compras, err := r.Relatorios.ContarCompras(ctx, desde, ate)
	ok(t, err)
	igual(t, "compras", compras, []port.ComprasCategoria{{Categoria: "livros", Compras: 2, Clientes: 1}, {Categoria: "games", Compras: 1, Clientes: 1}})
	for _, caso := range []struct {
		cats []string
		quer int
	}{{[]string{"livros", "games"}, 2}, {[]string{"livros"}, 1}, {[]string{"moda"}, 0}, {nil, 0}} {
		n, err := r.Relatorios.ContarCompradores(ctx, desde, ate, caso.cats)
		ok(t, err)
		igual(t, "compradores", n, caso.quer)
	}

	limites := []time.Time{agora.AddDate(0, 0, -2), agora.Add(time.Second), agora.AddDate(0, 0, 1)}
	serie, err := r.Relatorios.SerieMensal(ctx, limites)
	ok(t, err)
	igual(t, "série", serie, []port.ContagemMes{
		{Inicio: limites[0], Cadastros: 4, Revogacoes: 1, Eventos: 5, ClientesEventos: 2},
		{Inicio: limites[1]},
	})
}
