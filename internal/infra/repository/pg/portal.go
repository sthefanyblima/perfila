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

// Repositórios do portal do titular (migração 010). Spec: v1.10, 3.6, RN-19, RN-20, Q-56 a Q-59.
var (
	_ port.TokenAcessoRepository      = (*PgTokenAcessoRepository)(nil)
	_ port.PedidoEliminacaoRepository = (*PgPedidoEliminacaoRepository)(nil)
	_ port.CaixaSaidaSimulada         = (*PgCaixaSaidaSimulada)(nil)
)

// BuscarPorEmail usa o índice de lower(btrim(email)) da migração 010. Spec: RN-19, Q-59.
func (r *PgClienteRepository) BuscarPorEmail(ctx context.Context, email string) ([]entity.Cliente, error) {
	linhas, err := r.db.QueryContext(ctx, `SELECT `+colunasCliente+` FROM clientes
		WHERE lower(btrim(email)) = lower(btrim($1)) AND btrim($1) <> ''
		ORDER BY data_cadastro DESC, id::text`, email)
	if err != nil {
		return nil, fmt.Errorf("pg: buscando clientes por e-mail: %w", err) // sem o e-mail (C-09)
	}
	defer linhas.Close()
	out := []entity.Cliente{}
	for linhas.Next() {
		c, err := lerCliente(linhas)
		if err != nil {
			return nil, fmt.Errorf("pg: lendo cliente: %w", err)
		}
		out = append(out, c)
	}
	if err := linhas.Err(); err != nil {
		return nil, fmt.Errorf("pg: lendo clientes: %w", err)
	}
	return out, nil
}

// ---------- Tokens de acesso ----------

type PgTokenAcessoRepository struct{ db *sql.DB }

func NovoPgTokenAcessoRepository(db *sql.DB) *PgTokenAcessoRepository {
	return &PgTokenAcessoRepository{db}
}

func (r *PgTokenAcessoRepository) Salvar(ctx context.Context, t entity.TokenAcessoTitular) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tokens_acesso_titular (id, cliente_id, hash, criado_em, expira_em, usado_em) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET usado_em = EXCLUDED.usado_em`,
		t.ID, t.ClienteID, t.Hash, t.CriadoEm.UTC(), t.ExpiraEm.UTC(), nulo(t.UsadoEm))
	if err != nil {
		return fmt.Errorf("pg: salvando token de acesso %s: %w", t.ID, err)
	}
	return nil
}

// Consumir marca o uso numa única instrução: só a primeira de duas réplicas consegue (RN-19).
func (r *PgTokenAcessoRepository) Consumir(ctx context.Context, hash string, agora time.Time) (entity.TokenAcessoTitular, error) {
	var t entity.TokenAcessoTitular
	var usado sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		UPDATE tokens_acesso_titular SET usado_em = $2
		WHERE hash = $1 AND usado_em IS NULL AND expira_em > $2
		RETURNING id, cliente_id, hash, criado_em, expira_em, usado_em`, hash, agora.UTC()).
		Scan(&t.ID, &t.ClienteID, &t.Hash, &t.CriadoEm, &t.ExpiraEm, &usado)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.TokenAcessoTitular{}, errs.ErrAcessoInvalido
	}
	if err != nil {
		return entity.TokenAcessoTitular{}, fmt.Errorf("pg: consumindo token de acesso: %w", err) // sem o hash
	}
	t.CriadoEm, t.ExpiraEm, t.UsadoEm = utc(t.CriadoEm), utc(t.ExpiraEm), ptrUTC(usado)
	return t, nil
}

func (r *PgTokenAcessoRepository) ContarDesde(ctx context.Context, clienteID uuid.UUID, desde time.Time) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT count(*) FROM tokens_acesso_titular WHERE cliente_id = $1 AND criado_em >= $2`,
		clienteID, desde.UTC()).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("pg: contando tokens do cliente %s: %w", clienteID, err)
	}
	return n, nil
}

func (r *PgTokenAcessoRepository) RemoverPorCliente(ctx context.Context, clienteID uuid.UUID) (int, error) {
	return remover(ctx, r.db, "tokens_acesso_titular", clienteID)
}

// ---------- Pedidos de eliminação ----------

type PgPedidoEliminacaoRepository struct{ db *sql.DB }

func NovoPgPedidoEliminacaoRepository(db *sql.DB) *PgPedidoEliminacaoRepository {
	return &PgPedidoEliminacaoRepository{db}
}

// Salvar: um segundo pedido ABERTO do mesmo cliente não é gravado e vira ErrTransicaoInvalida,
// como na memória. O índice único parcial da migração 010 cobre a corrida entre réplicas.
func (r *PgPedidoEliminacaoRepository) Salvar(ctx context.Context, p entity.PedidoEliminacao) error {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO pedidos_eliminacao (id, cliente_id, status, data_pedido, data_atendimento)
		SELECT $1, $2, $3, $4, $5
		WHERE $3 <> 'ABERTO' OR NOT EXISTS (
			SELECT 1 FROM pedidos_eliminacao WHERE cliente_id = $2 AND status = 'ABERTO' AND id <> $1)
		ON CONFLICT (id) DO UPDATE SET status = EXCLUDED.status, data_atendimento = EXCLUDED.data_atendimento`,
		p.ID, p.ClienteID, string(p.Status), p.DataPedido.UTC(), nulo(p.DataAtendimento))
	if err != nil {
		return fmt.Errorf("pg: salvando pedido de eliminação %s: %w", p.ID, err)
	}
	if n, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("pg: conferindo o pedido de eliminação %s: %w", p.ID, err)
	} else if n == 0 {
		return fmt.Errorf("%w: o cliente %s já tem um pedido aberto", errs.ErrTransicaoInvalida, p.ClienteID)
	}
	return nil
}

const colunasPedido = `id, cliente_id, status, data_pedido, data_atendimento`

func lerPedido(l interface{ Scan(...any) error }) (entity.PedidoEliminacao, error) {
	var p entity.PedidoEliminacao
	var status string
	var atendimento sql.NullTime
	if err := l.Scan(&p.ID, &p.ClienteID, &status, &p.DataPedido, &atendimento); err != nil {
		return entity.PedidoEliminacao{}, err
	}
	p.Status, p.DataPedido, p.DataAtendimento = entity.StatusPedido(status), utc(p.DataPedido), ptrUTC(atendimento)
	return p, nil
}

func (r *PgPedidoEliminacaoRepository) BuscarAbertoPorCliente(ctx context.Context, clienteID uuid.UUID) (entity.PedidoEliminacao, error) {
	p, err := lerPedido(r.db.QueryRowContext(ctx, `SELECT `+colunasPedido+` FROM pedidos_eliminacao
		WHERE cliente_id = $1 AND status = 'ABERTO'`, clienteID))
	if errors.Is(err, sql.ErrNoRows) {
		return entity.PedidoEliminacao{}, errs.ErrPedidoNaoEncontrado
	}
	if err != nil {
		return entity.PedidoEliminacao{}, fmt.Errorf("pg: buscando pedido do cliente %s: %w", clienteID, err)
	}
	return p, nil
}

func (r *PgPedidoEliminacaoRepository) Listar(ctx context.Context) ([]entity.PedidoEliminacao, error) {
	linhas, err := r.db.QueryContext(ctx, `SELECT `+colunasPedido+` FROM pedidos_eliminacao
		ORDER BY (status = 'ABERTO') DESC, data_pedido DESC, id::text`)
	if err != nil {
		return nil, fmt.Errorf("pg: listando pedidos de eliminação: %w", err)
	}
	defer linhas.Close()
	out := []entity.PedidoEliminacao{}
	for linhas.Next() {
		p, err := lerPedido(linhas)
		if err != nil {
			return nil, fmt.Errorf("pg: lendo pedido de eliminação: %w", err)
		}
		out = append(out, p)
	}
	if err := linhas.Err(); err != nil {
		return nil, fmt.Errorf("pg: lendo pedidos de eliminação: %w", err)
	}
	return out, nil
}

// ---------- Caixa de saída simulada ----------

type PgCaixaSaidaSimulada struct{ db *sql.DB }

func NovaPgCaixaSaidaSimulada(db *sql.DB) *PgCaixaSaidaSimulada { return &PgCaixaSaidaSimulada{db} }

func (c *PgCaixaSaidaSimulada) Guardar(ctx context.Context, m entity.MensagemSimulada) error {
	_, err := c.db.ExecContext(ctx, `INSERT INTO caixa_saida_simulada (id, para, assunto, corpo, enviada_em) VALUES ($1, $2, $3, $4, $5)`,
		m.ID, m.Para, m.Assunto, m.Corpo, m.EnviadaEm.UTC())
	if err != nil {
		return fmt.Errorf("pg: guardando mensagem simulada %s: %w", m.ID, err) // sem o texto, que tem o link
	}
	return nil
}

func (c *PgCaixaSaidaSimulada) Listar(ctx context.Context, desde time.Time) ([]entity.MensagemSimulada, error) {
	linhas, err := c.db.QueryContext(ctx, `SELECT id, para, assunto, corpo, enviada_em FROM caixa_saida_simulada
		WHERE enviada_em >= $1 ORDER BY enviada_em DESC, ordem DESC`, desde.UTC())
	if err != nil {
		return nil, fmt.Errorf("pg: listando a caixa de saída simulada: %w", err)
	}
	defer linhas.Close()
	out := []entity.MensagemSimulada{}
	for linhas.Next() {
		var m entity.MensagemSimulada
		if err := linhas.Scan(&m.ID, &m.Para, &m.Assunto, &m.Corpo, &m.EnviadaEm); err != nil {
			return nil, fmt.Errorf("pg: lendo mensagem simulada: %w", err)
		}
		m.EnviadaEm = utc(m.EnviadaEm)
		out = append(out, m)
	}
	if err := linhas.Err(); err != nil {
		return nil, fmt.Errorf("pg: lendo a caixa de saída simulada: %w", err)
	}
	return out, nil
}

func (c *PgCaixaSaidaSimulada) RemoverAte(ctx context.Context, limite time.Time) (int, error) {
	res, err := c.db.ExecContext(ctx, `DELETE FROM caixa_saida_simulada WHERE enviada_em < $1`, limite.UTC())
	if err != nil {
		return 0, fmt.Errorf("pg: limpando a caixa de saída simulada: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("pg: contando mensagens apagadas: %w", err)
	}
	return int(n), nil
}
