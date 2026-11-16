package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/shared/errs"
)

// Conformidade em tempo de compilação (C-02).
var (
	_ port.OperadorRepository  = (*PgOperadorRepository)(nil)
	_ port.AuditoriaRepository = (*PgAuditoriaRepository)(nil)
)

// chaveTravaOperadores serializa a criação do administrador inicial entre as réplicas (RN-12).
const chaveTravaOperadores = 7_352_001

// violacaoUnica reconhece a violação de UNIQUE (23505) sem importar o pacote do driver.
func violacaoUnica(err error) bool {
	var e interface{ SQLState() string }
	return errors.As(err, &e) && e.SQLState() == "23505"
}

// ---------- Operador (v1.5, Q-35) ----------

// PgOperadorRepository grava em `operadores` (migração 007). Spec: 3.6, RN-12.
type PgOperadorRepository struct{ db *sql.DB }

func NovoPgOperadorRepository(db *sql.DB) *PgOperadorRepository { return &PgOperadorRepository{db} }

const colunasOperador = `id, login, nome, papel, senha_hash, ativo, criado_em, trocar_senha, senha_atualizada_em`

func (r *PgOperadorRepository) Salvar(ctx context.Context, o entity.Operador) error {
	return salvarOperador(ctx, r.db, o)
}

func salvarOperador(ctx context.Context, ex executor, o entity.Operador) error {
	_, err := ex.ExecContext(ctx, `INSERT INTO operadores (`+colunasOperador+`)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET login = EXCLUDED.login, nome = EXCLUDED.nome, papel = EXCLUDED.papel,
			senha_hash = EXCLUDED.senha_hash, ativo = EXCLUDED.ativo, trocar_senha = EXCLUDED.trocar_senha,
			senha_atualizada_em = EXCLUDED.senha_atualizada_em`,
		o.ID, entity.NormalizarLogin(o.Login), o.Nome, string(o.Papel), o.SenhaHash, o.Ativo, o.CriadoEm.UTC(),
		o.TrocarSenha, o.SenhaAtualizadaEm.UTC())
	if violacaoUnica(err) {
		return fmt.Errorf("pg: operador %s: %w", o.ID, errs.ErrLoginEmUso)
	}
	if err != nil {
		return fmt.Errorf("pg: salvando operador %s: %w", o.ID, err)
	}
	return nil
}

type varredor interface{ Scan(dest ...any) error }

func lerOperador(s varredor) (entity.Operador, error) {
	var o entity.Operador
	var papel string
	if err := s.Scan(&o.ID, &o.Login, &o.Nome, &papel, &o.SenhaHash, &o.Ativo, &o.CriadoEm, &o.TrocarSenha, &o.SenhaAtualizadaEm); err != nil {
		return entity.Operador{}, err
	}
	o.Papel, o.CriadoEm, o.SenhaAtualizadaEm = entity.PapelOperador(papel), utc(o.CriadoEm), utc(o.SenhaAtualizadaEm)
	return o, nil
}

func (r *PgOperadorRepository) buscar(ctx context.Context, where string, arg any) (entity.Operador, error) {
	o, err := lerOperador(r.db.QueryRowContext(ctx, `SELECT `+colunasOperador+` FROM operadores WHERE `+where, arg))
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Operador{}, errs.ErrOperadorNaoEncontrado
	}
	if err != nil {
		return entity.Operador{}, fmt.Errorf("pg: buscando operador: %w", err)
	}
	return o, nil
}

func (r *PgOperadorRepository) BuscarPorID(ctx context.Context, id uuid.UUID) (entity.Operador, error) {
	return r.buscar(ctx, "id = $1", id)
}

func (r *PgOperadorRepository) BuscarPorLogin(ctx context.Context, login string) (entity.Operador, error) {
	return r.buscar(ctx, "login = $1", entity.NormalizarLogin(login))
}

func (r *PgOperadorRepository) Listar(ctx context.Context) ([]entity.Operador, error) {
	linhas, err := r.db.QueryContext(ctx, `SELECT `+colunasOperador+` FROM operadores ORDER BY login`)
	if err != nil {
		return nil, fmt.Errorf("pg: listando operadores: %w", err)
	}
	defer linhas.Close()
	out := []entity.Operador{}
	for linhas.Next() {
		o, err := lerOperador(linhas)
		if err != nil {
			return nil, fmt.Errorf("pg: lendo operador: %w", err)
		}
		out = append(out, o)
	}
	return out, linhas.Err()
}

// CriarSeVazio confere e grava na mesma transação, sob trava consultiva: com duas réplicas
// subindo juntas, só uma cria o administrador inicial. Spec: RN-12, Q-35.
func (r *PgOperadorRepository) CriarSeVazio(ctx context.Context, o entity.Operador) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("pg: abrindo transação do administrador inicial: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // sem efeito depois do Commit
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, chaveTravaOperadores); err != nil {
		return false, fmt.Errorf("pg: travando operadores: %w", err)
	}
	var existe bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM operadores)`).Scan(&existe); err != nil {
		return false, fmt.Errorf("pg: contando operadores: %w", err)
	}
	if existe {
		return false, nil
	}
	if err := salvarOperador(ctx, tx, o); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("pg: confirmando administrador inicial: %w", err)
	}
	return true, nil
}

// ---------- Auditoria (v1.5, Q-39) ----------

// PgAuditoriaRepository grava em `registros_auditoria` (migração 007). Spec: 3.6, RN-13.
type PgAuditoriaRepository struct{ db *sql.DB }

func NovoPgAuditoriaRepository(db *sql.DB) *PgAuditoriaRepository { return &PgAuditoriaRepository{db} }

func (r *PgAuditoriaRepository) Registrar(ctx context.Context, a entity.RegistroAuditoria) error {
	var alvo any
	if a.AlvoID != uuid.Nil {
		alvo = a.AlvoID
	}
	if _, err := r.db.ExecContext(ctx, `INSERT INTO registros_auditoria (id, data, operador, acao, alvo_tipo, alvo_id, resultado, campos)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		a.ID, a.Data.UTC(), a.Operador, string(a.Acao), string(a.AlvoTipo), alvo, string(a.Resultado), strings.Join(a.Campos, ",")); err != nil {
		return fmt.Errorf("pg: gravando auditoria %s: %w", a.ID, err)
	}
	return nil
}

func (r *PgAuditoriaRepository) Listar(ctx context.Context, f port.FiltroAuditoria) ([]entity.RegistroAuditoria, int, error) {
	var cond []string
	var args []any
	add := func(c string, v any) {
		args = append(args, v)
		cond = append(cond, fmt.Sprintf(c, len(args)))
	}
	if f.Operador != "" {
		add("operador = $%d", entity.NormalizarLogin(f.Operador))
	}
	if f.Acao != "" {
		add("acao = $%d", string(f.Acao))
	}
	if !f.De.IsZero() {
		add("data >= $%d", f.De.UTC())
	}
	if !f.Ate.IsZero() {
		add("data < $%d", f.Ate.UTC())
	}
	where := ""
	if len(cond) > 0 {
		where = "WHERE " + strings.Join(cond, " AND ")
	}
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT count(*) FROM registros_auditoria `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("pg: contando auditoria: %w", err)
	}
	limite := "ALL"
	if f.Limite > 0 {
		limite = strconv.Itoa(f.Limite)
	}
	linhas, err := r.db.QueryContext(ctx, `SELECT id, data, operador, acao, alvo_tipo, alvo_id, resultado, campos
		FROM registros_auditoria `+where+` ORDER BY data DESC, ordem DESC LIMIT `+limite+
		` OFFSET `+strconv.Itoa(max(f.Deslocamento, 0)), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("pg: listando auditoria: %w", err)
	}
	defer linhas.Close()
	out := []entity.RegistroAuditoria{}
	for linhas.Next() {
		var a entity.RegistroAuditoria
		var acao, tipo, resultado, campos string
		var alvo uuid.NullUUID
		if err := linhas.Scan(&a.ID, &a.Data, &a.Operador, &acao, &tipo, &alvo, &resultado, &campos); err != nil {
			return nil, 0, fmt.Errorf("pg: lendo auditoria: %w", err)
		}
		a.Data, a.Acao, a.AlvoTipo, a.Resultado = utc(a.Data), entity.AcaoAuditoria(acao), entity.TipoAlvo(tipo), entity.ResultadoAuditoria(resultado)
		if alvo.Valid {
			a.AlvoID = alvo.UUID
		}
		if campos != "" {
			a.Campos = strings.Split(campos, ",")
		}
		out = append(out, a)
	}
	return out, total, linhas.Err()
}

func (r *PgAuditoriaRepository) ExpurgarAntesDe(ctx context.Context, limite time.Time) (int, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM registros_auditoria WHERE data < $1`, limite.UTC())
	if err != nil {
		return 0, fmt.Errorf("pg: expurgando auditoria: %w", err)
	}
	n, err := res.RowsAffected()
	return int(n), err
}
