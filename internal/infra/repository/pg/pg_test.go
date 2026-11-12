package pg_test

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"reflect"
	"sort"
	"testing"

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
		consentimentos_lgpd, clientes, produtos, segmentos`)
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
	if !reflect.DeepEqual(versoes, []string{"001_inicial", "002_dados_demonstracao"}) {
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
	for _, categoria := range []string{"livros", "games", "eletronicos"} {
		esperado, _ := memProd.ListarAtivosPorCategoria(ctx, categoria)
		obtido, err := pg.NovoPgProdutoRepository(db).ListarAtivosPorCategoria(ctx, categoria)
		sort.Slice(esperado, func(i, j int) bool { return esperado[i].Nome < esperado[j].Nome })
		if err != nil || !reflect.DeepEqual(obtido, esperado) {
			t.Fatalf("catálogo %s (%v):\n obtido   %+v\n esperado %+v", categoria, err, obtido, esperado)
		}
	}
}
