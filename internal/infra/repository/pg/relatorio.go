package pg

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"perfila/internal/domain/port"
)

var _ port.RelatorioRepository = (*PgRelatorioRepository)(nil)

// PgRelatorioRepository agrega no banco e devolve só números (COUNT, SUM, COUNT DISTINCT): nenhum
// dado de cliente sai do PostgreSQL. Comportamento só de quem tem consentimento CONCEDIDO agora.
// Spec: 3.6, 3.8, RN-15, Q-46.
type PgRelatorioRepository struct{ db *sql.DB }

func NovoPgRelatorioRepository(db *sql.DB) *PgRelatorioRepository { return &PgRelatorioRepository{db} }

// comConsentimento junta a tabela de alias "x" com o consentimento válido do mesmo cliente (RN-02).
const comConsentimento = ` JOIN consentimentos_lgpd cl ON cl.cliente_id = x.cliente_id AND cl.status = 'CONCEDIDO'`

func (r *PgRelatorioRepository) ContarClientes(ctx context.Context) (port.ContagemClientes, error) {
	var c port.ContagemClientes
	err := r.db.QueryRowContext(ctx, `SELECT count(*), count(*) FILTER (WHERE c.ativo),
		count(*) FILTER (WHERE cl.status = 'CONCEDIDO'), count(*) FILTER (WHERE cl.status = 'REVOGADO')
		FROM clientes c LEFT JOIN consentimentos_lgpd cl ON cl.cliente_id = c.id`).
		Scan(&c.Total, &c.Ativos, &c.Concedido, &c.Revogado)
	if err != nil {
		return port.ContagemClientes{}, fmt.Errorf("pg: contar clientes: %w", err)
	}
	c.NaoConcedido = c.Total - c.Concedido - c.Revogado // sem registro = NAO_CONCEDIDO (D-02)
	return c, nil
}

func (r *PgRelatorioRepository) ResumirSegmentos(ctx context.Context) ([]port.ResumoSegmento, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT x.segmento_id, count(*), sum(x.ticket_medio), sum(x.score_engajamento)
		FROM perfis_cliente x`+comConsentimento+` GROUP BY x.segmento_id`)
	if err != nil {
		return nil, fmt.Errorf("pg: resumir segmentos: %w", err)
	}
	defer rows.Close()
	out := []port.ResumoSegmento{}
	for rows.Next() {
		var s port.ResumoSegmento
		var id uuid.NullUUID
		if err := rows.Scan(&id, &s.Clientes, &s.SomaTicket, &s.SomaScore); err != nil {
			return nil, fmt.Errorf("pg: resumir segmentos: %w", err)
		}
		if id.Valid {
			s.SegmentoID = &id.UUID
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *PgRelatorioRepository) ContarRecomendacoes(ctx context.Context) (port.ContagemRecomendacoes, error) {
	var c port.ContagemRecomendacoes
	err := r.db.QueryRowContext(ctx, `SELECT
		count(*) FILTER (WHERE x.status = 'ATIVA'), count(DISTINCT x.cliente_id) FILTER (WHERE x.status = 'ATIVA'),
		count(*) FILTER (WHERE x.status = 'EXPIRADA'), count(DISTINCT x.cliente_id) FILTER (WHERE x.status = 'EXPIRADA')
		FROM recomendacoes_produto x`+comConsentimento).
		Scan(&c.Ativas, &c.ClientesAtivas, &c.Expiradas, &c.ClientesExpiradas)
	if err != nil {
		return port.ContagemRecomendacoes{}, fmt.Errorf("pg: contar recomendações: %w", err)
	}
	return c, nil
}

// filtroCompras: compras com categoria, de clientes com consentimento válido, em [$1, $2).
const filtroCompras = ` FROM eventos_comportamentais x` + comConsentimento + `
	WHERE x.tipo = 'compra' AND x.categoria <> '' AND x."timestamp" >= $1 AND x."timestamp" < $2`

func (r *PgRelatorioRepository) ContarCompras(ctx context.Context, desde, ate time.Time) ([]port.ComprasCategoria, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT x.categoria, count(*), count(DISTINCT x.cliente_id)`+filtroCompras+`
		GROUP BY x.categoria ORDER BY count(*) DESC, x.categoria COLLATE "C"`, utc(desde), utc(ate))
	if err != nil {
		return nil, fmt.Errorf("pg: contar compras: %w", err)
	}
	defer rows.Close()
	out := []port.ComprasCategoria{}
	for rows.Next() {
		var c port.ComprasCategoria
		if err := rows.Scan(&c.Categoria, &c.Compras, &c.Clientes); err != nil {
			return nil, fmt.Errorf("pg: contar compras: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *PgRelatorioRepository) ContarCompradores(ctx context.Context, desde, ate time.Time, categorias []string) (int, error) {
	if len(categorias) == 0 {
		return 0, nil
	}
	args := []any{utc(desde), utc(ate)}
	marcas := make([]string, len(categorias))
	for i, c := range categorias {
		args = append(args, c)
		marcas[i] = fmt.Sprintf("$%d", i+3)
	}
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT count(DISTINCT x.cliente_id)`+filtroCompras+
		` AND x.categoria IN (`+strings.Join(marcas, ", ")+`)`, args...).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("pg: contar compradores: %w", err)
	}
	return n, nil
}

func (r *PgRelatorioRepository) SerieMensal(ctx context.Context, limites []time.Time) ([]port.ContagemMes, error) {
	out := []port.ContagemMes{}
	for i := 0; i+1 < len(limites); i++ {
		m := port.ContagemMes{Inicio: utc(limites[i])}
		err := r.db.QueryRowContext(ctx, `SELECT
			(SELECT count(*) FROM clientes WHERE data_cadastro >= $1 AND data_cadastro < $2),
			(SELECT count(*) FROM movimentos_consentimento WHERE tipo = 'REVOGACAO' AND data >= $1 AND data < $2),
			(SELECT count(*) FROM eventos_comportamentais x`+comConsentimento+` WHERE x."timestamp" >= $1 AND x."timestamp" < $2),
			(SELECT count(DISTINCT x.cliente_id) FROM eventos_comportamentais x`+comConsentimento+`
				WHERE x."timestamp" >= $1 AND x."timestamp" < $2)`, utc(limites[i]), utc(limites[i+1])).
			Scan(&m.Cadastros, &m.Revogacoes, &m.Eventos, &m.ClientesEventos)
		if err != nil {
			return nil, fmt.Errorf("pg: série mensal: %w", err)
		}
		out = append(out, m)
	}
	return out, nil
}
