package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/shared/errs"
)

var _ port.PublicoRepository = (*PgPublicoRepository)(nil)

// PgPublicoRepository grava `exportacoes_publico` e `membros_publico` (migração 009). Spec: RN-17, Q-51.
type PgPublicoRepository struct{ db *sql.DB }

func NovoPgPublicoRepository(db *sql.DB) *PgPublicoRepository { return &PgPublicoRepository{db} }

// RegistrarExportacao grava exportação, conteúdo e membros numa transação.
func (r *PgPublicoRepository) RegistrarExportacao(ctx context.Context, e entity.ExportacaoPublico, conteudo []byte, membros []entity.MembroPublico) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("pg: abrindo transação da exportação %s: %w", e.ID, err)
	}
	defer tx.Rollback() //nolint:errcheck // sem efeito depois do Commit
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO exportacoes_publico (id, segmento_id, segmento_nome, tipo, destino, operador, data,
			quantidade, hash_arquivo, nome_arquivo, arquivo_expira_em, arquivo)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		e.ID, e.SegmentoID, e.SegmentoNome, string(e.Tipo), string(e.Destino), e.Operador, e.Data.UTC(),
		e.Quantidade, e.HashArquivo, e.NomeArquivo, e.ArquivoExpiraEm.UTC(), conteudo); err != nil {
		return fmt.Errorf("pg: gravando exportação %s: %w", e.ID, err)
	}
	for _, m := range membros {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO membros_publico (id, exportacao_id, segmento_id, cliente_id, hash_email, status,
				data_inclusao, data_remocao_pendente, remocao_id, data_remocao)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (id) DO UPDATE SET status = EXCLUDED.status, data_remocao_pendente = EXCLUDED.data_remocao_pendente,
				remocao_id = EXCLUDED.remocao_id, data_remocao = EXCLUDED.data_remocao`,
			m.ID, m.ExportacaoID, m.SegmentoID, m.ClienteID, m.HashEmail, string(m.Status), m.DataInclusao.UTC(),
			nulo(m.DataRemocaoPendente), uuidNulo(m.RemocaoID), nulo(m.DataRemocao)); err != nil {
			return fmt.Errorf("pg: gravando membro %s da exportação %s: %w", m.ID, e.ID, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("pg: confirmando exportação %s: %w", e.ID, err)
	}
	return nil
}

const colunasExportacao = `id, segmento_id, segmento_nome, tipo, destino, operador, data, quantidade,
	hash_arquivo, nome_arquivo, arquivo_expira_em`

func lerExportacao(l escaneavel, extra ...any) (entity.ExportacaoPublico, error) {
	var e entity.ExportacaoPublico
	var tipo, destino string
	dest := append([]any{&e.ID, &e.SegmentoID, &e.SegmentoNome, &tipo, &destino, &e.Operador, &e.Data,
		&e.Quantidade, &e.HashArquivo, &e.NomeArquivo, &e.ArquivoExpiraEm}, extra...)
	if err := l.Scan(dest...); err != nil {
		return entity.ExportacaoPublico{}, err
	}
	e.Tipo, e.Destino = entity.TipoExportacao(tipo), entity.TipoDestino(destino)
	e.Data, e.ArquivoExpiraEm = e.Data.UTC(), e.ArquivoExpiraEm.UTC()
	return e, nil
}

// ListarExportacoes devolve as exportações do segmento, da mais recente para a mais antiga.
func (r *PgPublicoRepository) ListarExportacoes(ctx context.Context, segmentoID uuid.UUID) ([]entity.ExportacaoPublico, error) {
	linhas, err := r.db.QueryContext(ctx, `SELECT `+colunasExportacao+`
		FROM exportacoes_publico WHERE segmento_id = $1 ORDER BY data DESC, id`, segmentoID)
	if err != nil {
		return nil, fmt.Errorf("pg: listando exportações do segmento %s: %w", segmentoID, err)
	}
	defer linhas.Close()
	out := []entity.ExportacaoPublico{}
	for linhas.Next() {
		e, err := lerExportacao(linhas)
		if err != nil {
			return nil, fmt.Errorf("pg: lendo exportação do segmento %s: %w", segmentoID, err)
		}
		out = append(out, e)
	}
	if err := linhas.Err(); err != nil {
		return nil, fmt.Errorf("pg: listando exportações do segmento %s: %w", segmentoID, err)
	}
	return out, nil
}

// BuscarArquivo devolve errs.ErrArquivoIndisponivel se a exportação não existir ou o conteúdo já
// tiver sido apagado.
func (r *PgPublicoRepository) BuscarArquivo(ctx context.Context, exportacaoID uuid.UUID) (entity.ExportacaoPublico, []byte, error) {
	var conteudo []byte
	e, err := lerExportacao(r.db.QueryRowContext(ctx, `SELECT `+colunasExportacao+`, arquivo
		FROM exportacoes_publico WHERE id = $1`, exportacaoID), &conteudo)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && conteudo == nil) {
		return entity.ExportacaoPublico{}, nil, errs.ErrArquivoIndisponivel
	}
	if err != nil {
		return entity.ExportacaoPublico{}, nil, fmt.Errorf("pg: buscando arquivo da exportação %s: %w", exportacaoID, err)
	}
	return e, conteudo, nil
}

// ApagarArquivosVencidos apaga o conteúdo vencido; o registro fica como prova. Spec: Q-51.
func (r *PgPublicoRepository) ApagarArquivosVencidos(ctx context.Context, agora time.Time) (int, error) {
	res, err := r.db.ExecContext(ctx, `UPDATE exportacoes_publico SET arquivo = NULL
		WHERE arquivo IS NOT NULL AND arquivo_expira_em <= $1`, agora.UTC())
	if err != nil {
		return 0, fmt.Errorf("pg: apagando arquivos vencidos: %w", err)
	}
	n, err := res.RowsAffected()
	return int(n), err
}

const colunasMembro = `id, exportacao_id, segmento_id, cliente_id, hash_email, status, data_inclusao,
	data_remocao_pendente, remocao_id, data_remocao`

// ListarMembrosPorSegmento devolve os membros do segmento no estado, por hash.
func (r *PgPublicoRepository) ListarMembrosPorSegmento(ctx context.Context, segmentoID uuid.UUID, status entity.StatusMembro) ([]entity.MembroPublico, error) {
	return r.membros(ctx, `segmento_id = $1`, segmentoID, status)
}

// ListarMembrosPorCliente devolve os membros do cliente no estado, por hash.
func (r *PgPublicoRepository) ListarMembrosPorCliente(ctx context.Context, clienteID uuid.UUID, status entity.StatusMembro) ([]entity.MembroPublico, error) {
	return r.membros(ctx, `cliente_id = $1`, clienteID, status)
}

func (r *PgPublicoRepository) membros(ctx context.Context, filtro string, id uuid.UUID, status entity.StatusMembro) ([]entity.MembroPublico, error) {
	linhas, err := r.db.QueryContext(ctx, `SELECT `+colunasMembro+` FROM membros_publico
		WHERE `+filtro+` AND status = $2 ORDER BY hash_email, id`, id, string(status))
	if err != nil {
		return nil, fmt.Errorf("pg: listando membros de público (%s): %w", id, err)
	}
	defer linhas.Close()
	out := []entity.MembroPublico{}
	for linhas.Next() {
		var m entity.MembroPublico
		var st string
		var pendente, removido sql.NullTime
		var remocao uuid.NullUUID
		if err := linhas.Scan(&m.ID, &m.ExportacaoID, &m.SegmentoID, &m.ClienteID, &m.HashEmail, &st, &m.DataInclusao,
			&pendente, &remocao, &removido); err != nil {
			return nil, fmt.Errorf("pg: lendo membro de público (%s): %w", id, err)
		}
		m.Status, m.DataInclusao = entity.StatusMembro(st), m.DataInclusao.UTC()
		m.DataRemocaoPendente, m.DataRemocao = ptrUTC(pendente), ptrUTC(removido)
		if remocao.Valid {
			u := remocao.UUID
			m.RemocaoID = &u
		}
		out = append(out, m)
	}
	if err := linhas.Err(); err != nil {
		return nil, fmt.Errorf("pg: listando membros de público (%s): %w", id, err)
	}
	return out, nil
}

// AtualizarMembros grava estado e datas numa transação; ID inexistente desfaz tudo.
func (r *PgPublicoRepository) AtualizarMembros(ctx context.Context, ms []entity.MembroPublico) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("pg: abrindo transação dos membros de público: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // sem efeito depois do Commit
	for _, m := range ms {
		res, err := tx.ExecContext(ctx, `UPDATE membros_publico SET status = $2, data_remocao_pendente = $3,
			remocao_id = $4, data_remocao = $5 WHERE id = $1`,
			m.ID, string(m.Status), nulo(m.DataRemocaoPendente), uuidNulo(m.RemocaoID), nulo(m.DataRemocao))
		if err != nil {
			return fmt.Errorf("pg: atualizando membro de público %s: %w", m.ID, err)
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return fmt.Errorf("pg: membro de público %s inexistente", m.ID)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("pg: confirmando membros de público: %w", err)
	}
	return nil
}

func uuidNulo(u *uuid.UUID) any {
	if u == nil {
		return nil
	}
	return *u
}
