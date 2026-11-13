package pg_test

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/infra/memoria"
	"perfila/internal/infra/repository/contrato"
	"perfila/internal/infra/repository/pg"
)

// bancoDeTeste conecta ao PostgreSQL de PERFILA_TEST_DATABASE_URL num esquema próprio,
// recriado do zero, e aplica as migrações. Esquemas separados isolam estes testes de outros
// pacotes que usem o mesmo banco ao mesmo tempo (go test roda pacotes em paralelo).
// Sem a variável o teste é pulado: testes unitários não usam banco (C-10); o CI a define.
func bancoDeTeste(t *testing.T, esquema string) *sql.DB {
	t.Helper()
	base := os.Getenv("PERFILA_TEST_DATABASE_URL")
	if base == "" {
		t.Skip("PERFILA_TEST_DATABASE_URL não definida: suíte PostgreSQL pulada")
	}
	ctx := context.Background()
	admin, err := pg.Abrir(ctx, base)
	if err != nil {
		t.Fatal(err)
	}
	_, err = admin.Exec(`DROP SCHEMA IF EXISTS ` + esquema + ` CASCADE; CREATE SCHEMA ` + esquema)
	admin.Close()
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(base)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	q.Set("search_path", esquema)
	u.RawQuery = q.Encode()
	db, err := pg.Abrir(ctx, u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err := pg.Migrar(ctx, db); err != nil {
		t.Fatal(err)
	}
	return db
}

func esvaziar(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`TRUNCATE recomendacoes_produto, perfis_cliente, eventos_comportamentais,
		movimentos_consentimento, consentimentos_lgpd, clientes, produtos, segmentos`)
	if err != nil {
		t.Fatal(err)
	}
}

// Spec: T-19 — o PostgreSQL cumpre o mesmo contrato dos fakes em memória.
func TestContratoPostgres(t *testing.T) {
	db := bancoDeTeste(t, "teste_contrato")
	contrato.Executar(t, func(t *testing.T) contrato.Repositorios {
		esvaziar(t, db)
		return contrato.Repositorios{
			Clientes:       pg.NovoPgClienteRepository(db),
			Eventos:        pg.NovoPgEventoRepository(db),
			Perfis:         pg.NovoPgPerfilRepository(db),
			Segmentos:      pg.NovoPgSegmentoRepository(db),
			Recomendacoes:  pg.NovoPgRecomendacaoRepository(db),
			Consentimentos: pg.NovoPgConsentimentoRepository(db),
			Produtos:       pg.NovoPgProdutoRepository(db),
			AdicionarSegmento: func(t *testing.T, s entity.Segmento) {
				if _, err := db.Exec(`INSERT INTO segmentos (id, nome, descricao, criterio, prioridade, ativo)
					VALUES ($1, $2, $3, $4, $5, $6)`, s.ID, s.Nome, s.Descricao, s.Criterio, s.Prioridade, s.Ativo); err != nil {
					t.Fatal(err)
				}
			},
			AdicionarProduto: func(t *testing.T, p entity.Produto) {
				if _, err := db.Exec(`INSERT INTO produtos (id, nome, categoria, preco, ativo)
					VALUES ($1, $2, $3, $4, $5)`, p.ID, p.Nome, p.Categoria, p.Preco, p.Ativo); err != nil {
					t.Fatal(err)
				}
			},
		}
	})
}

// Spec: Q-09 — migrar de novo não reaplica nada; a 001 cria o padrao e a 002 os dados de demonstração.
func TestMigracoes(t *testing.T) {
	db := bancoDeTeste(t, "teste_migracoes")
	ctx := context.Background()
	var versoes []string
	linhas, err := db.Query(`SELECT versao FROM schema_migracoes ORDER BY versao`)
	if err != nil {
		t.Fatal(err)
	}
	for linhas.Next() {
		var v string
		_ = linhas.Scan(&v)
		versoes = append(versoes, v)
	}
	linhas.Close()
	if !reflect.DeepEqual(versoes, []string{"001_inicial", "002_dados_demonstracao", "003_recomendacoes_ativas_por_cliente", "004_listagem_de_clientes", "005_demonstracao_ampliada", "006_movimentos_consentimento"}) {
		t.Fatalf("migrações registradas: %v", versoes)
	}
	aplicadas, err := pg.Migrar(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	if len(aplicadas) != 0 {
		t.Fatalf("segunda execução reaplicou %v", aplicadas)
	}
	// Q-08: o banco recém-migrado tem os mesmos segmentos e catálogo do modo memória.
	memSeg, memProd := memoria.NovoSegmentoRepository(), memoria.NovoProdutoRepository()
	memoria.CarregarDemonstracao(memSeg, memProd)
	esperados, _ := memSeg.ListarAtivos(ctx)
	ativos, err := pg.NovoPgSegmentoRepository(db).ListarAtivos(ctx)
	if err != nil || !reflect.DeepEqual(ativos, esperados) {
		t.Fatalf("segmentos da migração (%v):\n obtido   %+v\n esperado %+v", err, ativos, esperados)
	}
	todosMem, _ := memProd.ListarTodos(ctx)
	if todos, err := pg.NovoPgProdutoRepository(db).ListarTodos(ctx); err != nil || !reflect.DeepEqual(todos, todosMem) {
		t.Fatalf("catálogo completo (%v):\n obtido   %+v\n esperado %+v", err, todos, todosMem)
	}
	for _, categoria := range []string{"livros", "games", "eletronicos", "moda", "casa", "beleza", "esportes"} {
		esperado, _ := memProd.ListarAtivosPorCategoria(ctx, categoria)
		obtido, err := pg.NovoPgProdutoRepository(db).ListarAtivosPorCategoria(ctx, categoria)
		sort.Slice(esperado, func(i, j int) bool { return esperado[i].Nome < esperado[j].Nome })
		if err != nil || !reflect.DeepEqual(obtido, esperado) {
			t.Fatalf("catálogo %s (%v):\n obtido   %+v\n esperado %+v", categoria, err, obtido, esperado)
		}
	}
}

// Spec: RN-10, Q-34 — a 006 reconstrói só o que as datas provam, com origem migracao, e o banco
// recusa alterar ou apagar um movimento.
func TestMigracaoReconstroiMovimentos(t *testing.T) {
	db := bancoDeTeste(t, "teste_movimentos")
	ctx := context.Background()
	dia := func(d int) time.Time { return time.Date(2026, 1, d, 10, 0, 0, 0, time.UTC) }
	clientes := pg.NovoPgClienteRepository(db)
	cons := pg.NovoPgConsentimentoRepository(db)
	revogado, reconcedido, nunca := uuid.New(), uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{revogado, reconcedido, nunca} {
		c := entity.Cliente{ID: id, Nome: "Cliente Sintético", Email: "c-001@exemplo.test", DataCadastro: dia(1), Ativo: true}
		if err := clientes.Salvar(ctx, c); err != nil {
			t.Fatal(err)
		}
	}
	a := entity.NovoConsentimento(uuid.New(), revogado, dia(1))
	_ = a.Conceder("segmentacao", "v1", dia(2))
	_ = a.Revogar(dia(3))
	b := entity.NovoConsentimento(uuid.New(), reconcedido, dia(1))
	_ = b.Conceder("segmentacao", "v1", dia(2))
	_ = b.Revogar(dia(3))
	_ = b.Conceder("segmentacao", "v2", dia(4))
	for _, c := range []entity.ConsentimentoLGPD{a, b, entity.NovoConsentimento(uuid.New(), nunca, dia(1))} {
		if err := cons.Salvar(ctx, c); err != nil {
			t.Fatal(err)
		}
	}
	// Simula um banco anterior à 006 e aplica a migração de novo sobre esses registros.
	if _, err := db.Exec(`DROP TABLE movimentos_consentimento; DROP FUNCTION movimentos_consentimento_so_inclusao();
		DELETE FROM schema_migracoes WHERE versao = '006_movimentos_consentimento'`); err != nil {
		t.Fatal(err)
	}
	if _, err := pg.Migrar(ctx, db); err != nil {
		t.Fatal(err)
	}
	mov := func(cliente uuid.UUID, tipo entity.TipoMovimento, termo string, d int) entity.MovimentoConsentimento {
		fin := "segmentacao"
		if termo == "" {
			fin = ""
		}
		return entity.MovimentoConsentimento{ClienteID: cliente, Tipo: tipo, Finalidade: fin, VersaoTermo: termo,
			Data: dia(d), Origem: entity.OrigemMigracao}
	}
	esperado := map[uuid.UUID][]entity.MovimentoConsentimento{
		revogado:    {mov(revogado, entity.MovimentoConcessao, "v1", 2), mov(revogado, entity.MovimentoRevogacao, "v1", 3)},
		reconcedido: {mov(reconcedido, entity.MovimentoRevogacao, "", 3), mov(reconcedido, entity.MovimentoConcessao, "v2", 4)},
		nunca:       {},
	}
	for id, quer := range esperado {
		obtido, err := cons.ListarMovimentos(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		for i := range obtido {
			obtido[i].ID = uuid.Nil
		}
		if !reflect.DeepEqual(obtido, quer) {
			t.Fatalf("movimentos reconstruídos:\n obtido   %+v\n esperado %+v", obtido, quer)
		}
	}
	for _, sqlTexto := range []string{`UPDATE movimentos_consentimento SET versao_termo = 'v9'`, `DELETE FROM movimentos_consentimento`} {
		if _, err := db.Exec(sqlTexto); err == nil {
			t.Fatalf("o banco aceitou %q em movimentos_consentimento", sqlTexto)
		}
	}
}
