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
	_ port.ClienteRepository       = (*PgClienteRepository)(nil)
	_ port.EventoRepository        = (*PgEventoRepository)(nil)
	_ port.PerfilRepository        = (*PgPerfilRepository)(nil)
	_ port.SegmentoRepository      = (*PgSegmentoRepository)(nil)
	_ port.RecomendacaoRepository  = (*PgRecomendacaoRepository)(nil)
	_ port.ConsentimentoRepository = (*PgConsentimentoRepository)(nil)
	_ port.ProdutoRepository       = (*PgProdutoRepository)(nil)
)

// ---------- auxiliares de tempo (C-07: datas em UTC) ----------

func utc(t time.Time) time.Time { return t.UTC() }

func ptrUTC(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	v := t.Time.UTC()
	return &v
}

func nulo(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t.UTC(), Valid: true}
}

func nuloSeZero(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t.UTC(), Valid: true}
}

// ---------- Cliente ----------

// PgClienteRepository persiste clientes em `clientes`. Spec: UC-09, 3.6.
type PgClienteRepository struct{ db *sql.DB }

func NovoPgClienteRepository(db *sql.DB) *PgClienteRepository { return &PgClienteRepository{db} }

func (r *PgClienteRepository) Salvar(ctx context.Context, c entity.Cliente) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO clientes (id, nome, email, data_cadastro, ativo) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET nome = EXCLUDED.nome, email = EXCLUDED.email,
			data_cadastro = EXCLUDED.data_cadastro, ativo = EXCLUDED.ativo`,
		c.ID, c.Nome, c.Email, c.DataCadastro.UTC(), c.Ativo)
	if err != nil {
		return fmt.Errorf("pg: salvando cliente %s: %w", c.ID, err)
	}
	return nil
}

func (r *PgClienteRepository) BuscarPorID(ctx context.Context, id uuid.UUID) (entity.Cliente, error) {
	var c entity.Cliente
	err := r.db.QueryRowContext(ctx,
		`SELECT id, nome, email, data_cadastro, ativo FROM clientes WHERE id = $1`, id).
		Scan(&c.ID, &c.Nome, &c.Email, &c.DataCadastro, &c.Ativo)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Cliente{}, errs.ErrClienteNaoEncontrado
	}
	if err != nil {
		return entity.Cliente{}, fmt.Errorf("pg: buscando cliente %s: %w", id, err)
	}
	c.DataCadastro = utc(c.DataCadastro)
	return c, nil
}

// normalizado repete memoria.NormalizarBusca em SQL (minúsculas, sem acentos); a migração 004
// indexa a mesma expressão sobre nome. Spec: Q-23.
const normalizado = `translate(lower(%s), 'áàâãäéèêëíìîïóòôõöúùûüçñ', 'aaaaaeeeeiiiiooooouuuucn')`

// Listar aplica FiltroClientes (Q-23): ID exato quando a busca é um UUID; senão, nome contém.
func (r *PgClienteRepository) Listar(ctx context.Context, f port.FiltroClientes) ([]entity.Cliente, int, error) {
	busca := strings.TrimSpace(f.Busca)
	where, args := "", []any{}
	if id, err := uuid.Parse(busca); err == nil {
		where, args = "WHERE id = $1", append(args, id)
	} else if busca != "" {
		where = "WHERE " + fmt.Sprintf(normalizado, "nome") + " LIKE '%' || " + fmt.Sprintf(normalizado, "$1::text") + " || '%'"
		args = append(args, escaparLike(busca))
	}
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT count(*) FROM clientes `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("pg: contando clientes: %w", err)
	}
	limite := "ALL"
	if f.Limite > 0 {
		limite = strconv.Itoa(f.Limite)
	}
	linhas, err := r.db.QueryContext(ctx, `SELECT id, nome, email, data_cadastro, ativo FROM clientes `+where+
		` ORDER BY data_cadastro DESC, id::text LIMIT `+limite+` OFFSET `+strconv.Itoa(max(f.Deslocamento, 0)), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("pg: listando clientes: %w", err)
	}
	defer linhas.Close()
	out := []entity.Cliente{}
	for linhas.Next() {
		var c entity.Cliente
		if err := linhas.Scan(&c.ID, &c.Nome, &c.Email, &c.DataCadastro, &c.Ativo); err != nil {
			return nil, 0, fmt.Errorf("pg: lendo cliente: %w", err)
		}
		c.DataCadastro = utc(c.DataCadastro)
		out = append(out, c)
	}
	if err := linhas.Err(); err != nil {
		return nil, 0, fmt.Errorf("pg: lendo clientes: %w", err)
	}
	return out, total, nil
}

// escaparLike impede que % e _ digitados virem curingas do LIKE (escape padrão: barra invertida).
func escaparLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(s)
}

// ---------- Evento ----------

// PgEventoRepository persiste eventos em `eventos_comportamentais`. Spec: UC-01, D-12.
type PgEventoRepository struct{ db *sql.DB }

func NovoPgEventoRepository(db *sql.DB) *PgEventoRepository { return &PgEventoRepository{db} }

func (r *PgEventoRepository) Salvar(ctx context.Context, e entity.EventoComportamental) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO eventos_comportamentais (id, cliente_id, tipo, origem, "timestamp", valor, categoria)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		e.ID, e.ClienteID, string(e.Tipo), e.Origem, e.Timestamp.UTC(), e.Valor, e.Categoria)
	if err != nil {
		return fmt.Errorf("pg: salvando evento %s: %w", e.ID, err)
	}
	return nil
}

func (r *PgEventoRepository) BuscarPorCliente(ctx context.Context, clienteID uuid.UUID) ([]entity.EventoComportamental, error) {
	linhas, err := r.db.QueryContext(ctx, `
		SELECT id, cliente_id, tipo, origem, "timestamp", valor, categoria
		FROM eventos_comportamentais WHERE cliente_id = $1 ORDER BY "timestamp", id`, clienteID)
	if err != nil {
		return nil, fmt.Errorf("pg: buscando eventos do cliente %s: %w", clienteID, err)
	}
	defer linhas.Close()
	out := []entity.EventoComportamental{}
	for linhas.Next() {
		var e entity.EventoComportamental
		var tipo string
		if err := linhas.Scan(&e.ID, &e.ClienteID, &tipo, &e.Origem, &e.Timestamp, &e.Valor, &e.Categoria); err != nil {
			return nil, fmt.Errorf("pg: lendo evento: %w", err)
		}
		e.Tipo = entity.TipoEvento(tipo)
		e.Timestamp = utc(e.Timestamp)
		out = append(out, e)
	}
	if err := linhas.Err(); err != nil {
		return nil, fmt.Errorf("pg: lendo eventos: %w", err)
	}
	return out, nil
}

// ---------- Perfil ----------

// PgPerfilRepository persiste perfis em `perfis_cliente`. Spec: UC-02, UC-03, RN-05, D-06.
type PgPerfilRepository struct{ db *sql.DB }

func NovoPgPerfilRepository(db *sql.DB) *PgPerfilRepository { return &PgPerfilRepository{db} }

// Salvar faz upsert por ClienteID e preserva o ID original (RN-05).
func (r *PgPerfilRepository) Salvar(ctx context.Context, p entity.PerfilCliente) (entity.PerfilCliente, error) {
	segmento := uuid.NullUUID{}
	if p.SegmentoID != nil {
		segmento = uuid.NullUUID{UUID: *p.SegmentoID, Valid: true}
	}
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO perfis_cliente (id, cliente_id, frequencia_compras, ticket_medio, categoria_preferida,
			score_engajamento, segmento_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (cliente_id) DO UPDATE SET frequencia_compras = EXCLUDED.frequencia_compras,
			ticket_medio = EXCLUDED.ticket_medio, categoria_preferida = EXCLUDED.categoria_preferida,
			score_engajamento = EXCLUDED.score_engajamento, segmento_id = EXCLUDED.segmento_id
		RETURNING id`,
		p.ID, p.ClienteID, p.FrequenciaCompras, p.TicketMedio, p.CategoriaPreferida, p.ScoreEngajamento, segmento).
		Scan(&p.ID)
	if err != nil {
		return entity.PerfilCliente{}, fmt.Errorf("pg: salvando perfil do cliente %s: %w", p.ClienteID, err)
	}
	return p, nil
}

func (r *PgPerfilRepository) BuscarPorCliente(ctx context.Context, clienteID uuid.UUID) (entity.PerfilCliente, error) {
	var p entity.PerfilCliente
	var segmento uuid.NullUUID
	err := r.db.QueryRowContext(ctx, `
		SELECT id, cliente_id, frequencia_compras, ticket_medio, categoria_preferida, score_engajamento, segmento_id
		FROM perfis_cliente WHERE cliente_id = $1`, clienteID).
		Scan(&p.ID, &p.ClienteID, &p.FrequenciaCompras, &p.TicketMedio, &p.CategoriaPreferida, &p.ScoreEngajamento, &segmento)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.PerfilCliente{}, errs.ErrPerfilNaoEncontrado
	}
	if err != nil {
		return entity.PerfilCliente{}, fmt.Errorf("pg: buscando perfil do cliente %s: %w", clienteID, err)
	}
	if segmento.Valid {
		id := segmento.UUID
		p.SegmentoID = &id
	}
	return p, nil
}

// ---------- Segmento ----------

// PgSegmentoRepository lê `segmentos` (populados por migração). Spec: UC-03, RN-06.
type PgSegmentoRepository struct{ db *sql.DB }

func NovoPgSegmentoRepository(db *sql.DB) *PgSegmentoRepository { return &PgSegmentoRepository{db} }

// ListarAtivos devolve os segmentos ativos em ordem crescente de Prioridade (desempate por Nome).
func (r *PgSegmentoRepository) ListarAtivos(ctx context.Context) ([]entity.Segmento, error) {
	linhas, err := r.db.QueryContext(ctx, `
		SELECT id, nome, descricao, criterio, prioridade, ativo
		FROM segmentos WHERE ativo ORDER BY prioridade, nome`)
	if err != nil {
		return nil, fmt.Errorf("pg: listando segmentos: %w", err)
	}
	defer linhas.Close()
	out := []entity.Segmento{}
	for linhas.Next() {
		var s entity.Segmento
		if err := linhas.Scan(&s.ID, &s.Nome, &s.Descricao, &s.Criterio, &s.Prioridade, &s.Ativo); err != nil {
			return nil, fmt.Errorf("pg: lendo segmento: %w", err)
		}
		out = append(out, s)
	}
	if err := linhas.Err(); err != nil {
		return nil, fmt.Errorf("pg: lendo segmentos: %w", err)
	}
	return out, nil
}

// BuscarPorID devolve errs.ErrDadosInvalidos se o segmento não existir (mesmo contrato do fake).
func (r *PgSegmentoRepository) BuscarPorID(ctx context.Context, id uuid.UUID) (entity.Segmento, error) {
	var s entity.Segmento
	err := r.db.QueryRowContext(ctx,
		`SELECT id, nome, descricao, criterio, prioridade, ativo FROM segmentos WHERE id = $1`, id).
		Scan(&s.ID, &s.Nome, &s.Descricao, &s.Criterio, &s.Prioridade, &s.Ativo)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Segmento{}, fmt.Errorf("segmento %s inexistente: %w", id, errs.ErrDadosInvalidos)
	}
	if err != nil {
		return entity.Segmento{}, fmt.Errorf("pg: buscando segmento %s: %w", id, err)
	}
	return s, nil
}

// ---------- Recomendação ----------

// PgRecomendacaoRepository persiste `recomendacoes_produto`. Spec: UC-04, RN-08, D-08.
type PgRecomendacaoRepository struct{ db *sql.DB }

func NovoPgRecomendacaoRepository(db *sql.DB) *PgRecomendacaoRepository {
	return &PgRecomendacaoRepository{db}
}

// Salvar insere ou atualiza por ID.
func (r *PgRecomendacaoRepository) Salvar(ctx context.Context, rec entity.RecomendacaoProduto) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO recomendacoes_produto (id, cliente_id, produto_id, motivo, status, data_geracao,
			data_ativacao, validade, data_expiracao)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET motivo = EXCLUDED.motivo, status = EXCLUDED.status,
			data_ativacao = EXCLUDED.data_ativacao, validade = EXCLUDED.validade,
			data_expiracao = EXCLUDED.data_expiracao`,
		rec.ID, rec.ClienteID, rec.ProdutoID, rec.Motivo, string(rec.Status), rec.DataGeracao.UTC(),
		nulo(rec.DataAtivacao), nulo(rec.Validade), nulo(rec.DataExpiracao))
	if err != nil {
		return fmt.Errorf("pg: salvando recomendação %s: %w", rec.ID, err)
	}
	return nil
}

// ListarAtivasVencidas devolve as recomendações ATIVA com Validade <= agora (RN-08).
func (r *PgRecomendacaoRepository) ListarAtivasVencidas(ctx context.Context, agora time.Time) ([]entity.RecomendacaoProduto, error) {
	return r.listar(ctx, `WHERE status = 'ATIVA' AND validade IS NOT NULL AND validade <= $1 ORDER BY validade, id`, agora.UTC())
}

// ListarAtivasPorCliente devolve as recomendações ATIVA do cliente (Q-18).
func (r *PgRecomendacaoRepository) ListarAtivasPorCliente(ctx context.Context, clienteID uuid.UUID) ([]entity.RecomendacaoProduto, error) {
	return r.listar(ctx, `WHERE status = 'ATIVA' AND cliente_id = $1 ORDER BY data_geracao, id`, clienteID)
}

// ListarPorCliente devolve todas as recomendações do cliente, a mais recente primeiro (Q-23).
func (r *PgRecomendacaoRepository) ListarPorCliente(ctx context.Context, clienteID uuid.UUID) ([]entity.RecomendacaoProduto, error) {
	return r.listar(ctx, `WHERE cliente_id = $1 ORDER BY data_geracao DESC, id::text`, clienteID)
}

func (r *PgRecomendacaoRepository) listar(ctx context.Context, filtro string, arg any) ([]entity.RecomendacaoProduto, error) {
	linhas, err := r.db.QueryContext(ctx, `
		SELECT id, cliente_id, produto_id, motivo, status, data_geracao, data_ativacao, validade, data_expiracao
		FROM recomendacoes_produto `+filtro, arg)
	if err != nil {
		return nil, fmt.Errorf("pg: listando recomendações: %w", err)
	}
	defer linhas.Close()
	out := []entity.RecomendacaoProduto{}
	for linhas.Next() {
		var rec entity.RecomendacaoProduto
		var status string
		var ativacao, validade, expiracao sql.NullTime
		if err := linhas.Scan(&rec.ID, &rec.ClienteID, &rec.ProdutoID, &rec.Motivo, &status, &rec.DataGeracao,
			&ativacao, &validade, &expiracao); err != nil {
			return nil, fmt.Errorf("pg: lendo recomendação: %w", err)
		}
		rec.Status = entity.StatusRecomendacao(status)
		rec.DataGeracao = utc(rec.DataGeracao)
		rec.DataAtivacao, rec.Validade, rec.DataExpiracao = ptrUTC(ativacao), ptrUTC(validade), ptrUTC(expiracao)
		out = append(out, rec)
	}
	if err := linhas.Err(); err != nil {
		return nil, fmt.Errorf("pg: lendo recomendações: %w", err)
	}
	return out, nil
}

// ---------- Consentimento ----------

// PgConsentimentoRepository persiste `consentimentos_lgpd`. Spec: UC-08, 3.4.1, D-02.
type PgConsentimentoRepository struct{ db *sql.DB }

func NovoPgConsentimentoRepository(db *sql.DB) *PgConsentimentoRepository {
	return &PgConsentimentoRepository{db}
}

// executor é o que *sql.DB e *sql.Tx têm em comum para gravar.
type executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// Salvar faz upsert por ClienteID e preserva o ID original.
func (r *PgConsentimentoRepository) Salvar(ctx context.Context, c entity.ConsentimentoLGPD) error {
	return salvarConsentimento(ctx, r.db, c)
}

// Registrar faz o upsert do registro e a inclusão do movimento numa transação. Spec: RN-10, Q-34.
func (r *PgConsentimentoRepository) Registrar(ctx context.Context, c entity.ConsentimentoLGPD, m entity.MovimentoConsentimento) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("pg: abrindo transação do consentimento do cliente %s: %w", c.ClienteID, err)
	}
	defer tx.Rollback() //nolint:errcheck // sem efeito depois do Commit
	if err := salvarConsentimento(ctx, tx, c); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO movimentos_consentimento (id, cliente_id, tipo, finalidade, versao_termo, data, origem, operador)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		m.ID, m.ClienteID, string(m.Tipo), m.Finalidade, m.VersaoTermo, m.Data.UTC(), string(m.Origem), m.Operador); err != nil {
		return fmt.Errorf("pg: gravando movimento do consentimento do cliente %s: %w", m.ClienteID, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("pg: confirmando consentimento do cliente %s: %w", c.ClienteID, err)
	}
	return nil
}

// ListarMovimentos devolve os movimentos em ordem cronológica, com a ordem de gravação no empate.
func (r *PgConsentimentoRepository) ListarMovimentos(ctx context.Context, clienteID uuid.UUID) ([]entity.MovimentoConsentimento, error) {
	linhas, err := r.db.QueryContext(ctx, `
		SELECT id, cliente_id, tipo, finalidade, versao_termo, data, origem, operador
		FROM movimentos_consentimento WHERE cliente_id = $1 ORDER BY data, ordem`, clienteID)
	if err != nil {
		return nil, fmt.Errorf("pg: listando movimentos do cliente %s: %w", clienteID, err)
	}
	defer linhas.Close()
	out := []entity.MovimentoConsentimento{}
	for linhas.Next() {
		var m entity.MovimentoConsentimento
		var tipo, origem string
		if err := linhas.Scan(&m.ID, &m.ClienteID, &tipo, &m.Finalidade, &m.VersaoTermo, &m.Data, &origem, &m.Operador); err != nil {
			return nil, fmt.Errorf("pg: lendo movimento do cliente %s: %w", clienteID, err)
		}
		m.Tipo, m.Origem, m.Data = entity.TipoMovimento(tipo), entity.OrigemMovimento(origem), m.Data.UTC()
		out = append(out, m)
	}
	if err := linhas.Err(); err != nil {
		return nil, fmt.Errorf("pg: listando movimentos do cliente %s: %w", clienteID, err)
	}
	return out, nil
}

func salvarConsentimento(ctx context.Context, ex executor, c entity.ConsentimentoLGPD) error {
	_, err := ex.ExecContext(ctx, `
		INSERT INTO consentimentos_lgpd (id, cliente_id, status, finalidade, versao_termo,
			data_consentimento, data_revogacao, data_atualizacao)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (cliente_id) DO UPDATE SET status = EXCLUDED.status, finalidade = EXCLUDED.finalidade,
			versao_termo = EXCLUDED.versao_termo, data_consentimento = EXCLUDED.data_consentimento,
			data_revogacao = EXCLUDED.data_revogacao, data_atualizacao = EXCLUDED.data_atualizacao`,
		c.ID, c.ClienteID, string(c.Status), c.Finalidade, c.VersaoTermo,
		nuloSeZero(c.DataConsentimento), nulo(c.DataRevogacao), c.DataAtualizacao.UTC())
	if err != nil {
		return fmt.Errorf("pg: salvando consentimento do cliente %s: %w", c.ClienteID, err)
	}
	return nil
}

// BuscarPorCliente devolve errs.ErrConsentimentoNaoEncontrado se não houver registro.
func (r *PgConsentimentoRepository) BuscarPorCliente(ctx context.Context, clienteID uuid.UUID) (entity.ConsentimentoLGPD, error) {
	var c entity.ConsentimentoLGPD
	var status string
	var concessao, revogacao sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT id, cliente_id, status, finalidade, versao_termo, data_consentimento, data_revogacao, data_atualizacao
		FROM consentimentos_lgpd WHERE cliente_id = $1`, clienteID).
		Scan(&c.ID, &c.ClienteID, &status, &c.Finalidade, &c.VersaoTermo, &concessao, &revogacao, &c.DataAtualizacao)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.ConsentimentoLGPD{}, errs.ErrConsentimentoNaoEncontrado
	}
	if err != nil {
		return entity.ConsentimentoLGPD{}, fmt.Errorf("pg: buscando consentimento do cliente %s: %w", clienteID, err)
	}
	c.Status = entity.StatusConsentimento(status)
	if concessao.Valid {
		c.DataConsentimento = concessao.Time.UTC()
	}
	c.DataRevogacao = ptrUTC(revogacao)
	c.DataAtualizacao = utc(c.DataAtualizacao)
	return c, nil
}

// ---------- Produto ----------

// PgProdutoRepository lê o catálogo em `produtos` (populado por migração, Q-08). Spec: D-05, RN-07.
type PgProdutoRepository struct{ db *sql.DB }

func NovoPgProdutoRepository(db *sql.DB) *PgProdutoRepository { return &PgProdutoRepository{db} }

func (r *PgProdutoRepository) ListarAtivosPorCategoria(ctx context.Context, categoria string) ([]entity.Produto, error) {
	return r.listar(ctx, `WHERE ativo AND categoria = $1 ORDER BY nome, id`, categoria)
}

// ListarTodos devolve o catálogo completo, por Categoria, Nome e ID (Q-23).
func (r *PgProdutoRepository) ListarTodos(ctx context.Context) ([]entity.Produto, error) {
	return r.listar(ctx, `ORDER BY categoria, nome, id::text`)
}

func (r *PgProdutoRepository) listar(ctx context.Context, filtro string, args ...any) ([]entity.Produto, error) {
	linhas, err := r.db.QueryContext(ctx, `SELECT id, nome, categoria, preco, ativo FROM produtos `+filtro, args...)
	if err != nil {
		return nil, fmt.Errorf("pg: listando produtos da categoria: %w", err)
	}
	defer linhas.Close()
	out := []entity.Produto{}
	for linhas.Next() {
		var p entity.Produto
		if err := linhas.Scan(&p.ID, &p.Nome, &p.Categoria, &p.Preco, &p.Ativo); err != nil {
			return nil, fmt.Errorf("pg: lendo produto: %w", err)
		}
		out = append(out, p)
	}
	if err := linhas.Err(); err != nil {
		return nil, fmt.Errorf("pg: lendo produtos: %w", err)
	}
	return out, nil
}
