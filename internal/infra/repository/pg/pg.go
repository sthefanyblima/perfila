// Package pg implementa os ports de repositório do domínio sobre PostgreSQL, via database/sql
// com o driver pgx (Q-10), e aplica as migrações SQL embutidas no binário (Q-09).
//
// Spec: T-19, 3.6, 3.8. Fonte: UML7 (Repositórios) e UML8 (omega_db).
package pg

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // registra o driver "pgx" no database/sql
)

//go:embed migracoes/*.sql
var migracoes embed.FS

// chaveTravaMigracao identifica a trava consultiva que serializa as migrações entre réplicas.
const chaveTravaMigracao = 7_280_190_019 // "perfila" + T-19

// Abrir conecta ao PostgreSQL e confere a conexão. Spec: 3.8.
func Abrir(ctx context.Context, url string) (*sql.DB, error) {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, fmt.Errorf("pg: abrindo conexão: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxIdleTime(5 * time.Minute)

	ping, cancelar := context.WithTimeout(ctx, 10*time.Second)
	defer cancelar()
	if err := db.PingContext(ping); err != nil {
		db.Close()
		return nil, fmt.Errorf("pg: conectando ao banco: %w", err)
	}
	return db, nil
}

// Migrar aplica, em ordem, as migrações ainda não registradas em schema_migracoes.
// Uma trava consultiva impede que duas réplicas migrem ao mesmo tempo; cada migração
// roda na própria transação. Devolve as versões aplicadas nesta chamada. Spec: Q-09.
func Migrar(ctx context.Context, db *sql.DB) ([]string, error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("pg: migração: %w", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, chaveTravaMigracao); err != nil {
		return nil, fmt.Errorf("pg: travando migração: %w", err)
	}
	defer conn.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, chaveTravaMigracao)

	if _, err := conn.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migracoes (
		versao      text        PRIMARY KEY,
		aplicada_em timestamptz NOT NULL DEFAULT now())`); err != nil {
		return nil, fmt.Errorf("pg: criando schema_migracoes: %w", err)
	}

	arquivos, err := fs.Glob(migracoes, "migracoes/*.sql")
	if err != nil {
		return nil, fmt.Errorf("pg: listando migrações: %w", err)
	}
	sort.Strings(arquivos)

	aplicadas := []string{}
	for _, arq := range arquivos {
		versao := strings.TrimSuffix(strings.TrimPrefix(arq, "migracoes/"), ".sql")
		var existe bool
		if err := conn.QueryRowContext(ctx,
			`SELECT EXISTS (SELECT 1 FROM schema_migracoes WHERE versao = $1)`, versao).Scan(&existe); err != nil {
			return aplicadas, fmt.Errorf("pg: consultando migração %s: %w", versao, err)
		}
		if existe {
			continue
		}
		sqlTexto, err := migracoes.ReadFile(arq)
		if err != nil {
			return aplicadas, fmt.Errorf("pg: lendo migração %s: %w", versao, err)
		}
		if err := aplicar(ctx, conn, versao, string(sqlTexto)); err != nil {
			return aplicadas, err
		}
		aplicadas = append(aplicadas, versao)
	}
	return aplicadas, nil
}

// UltimaMigracao devolve a versão mais recente embutida no binário: o esquema depois de Migrar.
// O painel a mostra na tela de Operação (Spec 3.7.1).
func UltimaMigracao() string {
	arquivos, _ := fs.Glob(migracoes, "migracoes/*.sql")
	if len(arquivos) == 0 {
		return ""
	}
	sort.Strings(arquivos)
	return strings.TrimSuffix(strings.TrimPrefix(arquivos[len(arquivos)-1], "migracoes/"), ".sql")
}

func aplicar(ctx context.Context, conn *sql.Conn, versao, sqlTexto string) error {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("pg: migração %s: %w", versao, err)
	}
	if _, err := tx.ExecContext(ctx, sqlTexto); err != nil {
		tx.Rollback()
		return fmt.Errorf("pg: aplicando migração %s: %w", versao, err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migracoes (versao) VALUES ($1)`, versao); err != nil {
		tx.Rollback()
		return fmt.Errorf("pg: registrando migração %s: %w", versao, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("pg: confirmando migração %s: %w", versao, err)
	}
	return nil
}
