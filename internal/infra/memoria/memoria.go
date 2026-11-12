// Package memoria implementa os ports em memória (fakes) para testes BDD, contrato e demonstração.
// Cada repositório é um valor independente, criado por construtor (sem estado global — C-08).
//
// Spec: T-13, 3.6.
package memoria

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/shared/errs"
)

// Conformidade em tempo de compilação (C-02).
var (
	_ port.ClienteRepository       = (*ClienteRepository)(nil)
	_ port.EventoRepository        = (*EventoRepository)(nil)
	_ port.PerfilRepository        = (*PerfilRepository)(nil)
	_ port.SegmentoRepository      = (*SegmentoRepository)(nil)
	_ port.RecomendacaoRepository  = (*RecomendacaoRepository)(nil)
	_ port.ConsentimentoRepository = (*ConsentimentoRepository)(nil)
	_ port.ProdutoRepository       = (*ProdutoRepository)(nil)
)

// ---------- Cliente ----------

type ClienteRepository struct {
	mu    sync.RWMutex
	dados map[uuid.UUID]entity.Cliente
}

func NovoClienteRepository() *ClienteRepository {
	return &ClienteRepository{dados: map[uuid.UUID]entity.Cliente{}}
}

func (r *ClienteRepository) Salvar(_ context.Context, c entity.Cliente) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.dados[c.ID] = c
	return nil
}

func (r *ClienteRepository) BuscarPorID(_ context.Context, id uuid.UUID) (entity.Cliente, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.dados[id]
	if !ok {
		return entity.Cliente{}, errs.ErrClienteNaoEncontrado
	}
	return c, nil
}

// ---------- Evento ----------

type EventoRepository struct {
	mu    sync.RWMutex
	dados []entity.EventoComportamental
}

func NovoEventoRepository() *EventoRepository { return &EventoRepository{} }

func (r *EventoRepository) Salvar(_ context.Context, e entity.EventoComportamental) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.dados = append(r.dados, e)
	return nil
}

func (r *EventoRepository) BuscarPorCliente(_ context.Context, clienteID uuid.UUID) ([]entity.EventoComportamental, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []entity.EventoComportamental{}
	for _, e := range r.dados {
		if e.ClienteID == clienteID {
			out = append(out, e)
		}
	}
	return out, nil
}

// Total devolve o número de eventos gravados (apoio a testes).
func (r *EventoRepository) Total() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.dados)
}

// ---------- Perfil ----------

type PerfilRepository struct {
	mu    sync.RWMutex
	dados map[uuid.UUID]entity.PerfilCliente // chave: ClienteID
}

func NovoPerfilRepository() *PerfilRepository {
	return &PerfilRepository{dados: map[uuid.UUID]entity.PerfilCliente{}}
}

// Salvar faz upsert por ClienteID e preserva o ID original (RN-05).
func (r *PerfilRepository) Salvar(_ context.Context, p entity.PerfilCliente) (entity.PerfilCliente, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if atual, ok := r.dados[p.ClienteID]; ok {
		p.ID = atual.ID
	}
	r.dados[p.ClienteID] = p
	return p, nil
}

func (r *PerfilRepository) BuscarPorCliente(_ context.Context, clienteID uuid.UUID) (entity.PerfilCliente, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.dados[clienteID]
	if !ok {
		return entity.PerfilCliente{}, errs.ErrPerfilNaoEncontrado
	}
	return p, nil
}

// Total devolve o número de perfis gravados (apoio a testes).
func (r *PerfilRepository) Total() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.dados)
}

// ---------- Segmento ----------

type SegmentoRepository struct {
	mu    sync.RWMutex
	dados map[uuid.UUID]entity.Segmento
}

func NovoSegmentoRepository() *SegmentoRepository {
	return &SegmentoRepository{dados: map[uuid.UUID]entity.Segmento{}}
}

// Adicionar insere um segmento (no PostgreSQL, segmentos entram por migração).
func (r *SegmentoRepository) Adicionar(s entity.Segmento) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.dados[s.ID] = s
}

func (r *SegmentoRepository) ListarAtivos(_ context.Context) ([]entity.Segmento, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []entity.Segmento{}
	for _, s := range r.dados {
		if s.Ativo {
			out = append(out, s)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Prioridade != out[j].Prioridade {
			return out[i].Prioridade < out[j].Prioridade
		}
		return out[i].Nome < out[j].Nome
	})
	return out, nil
}

func (r *SegmentoRepository) BuscarPorID(_ context.Context, id uuid.UUID) (entity.Segmento, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.dados[id]
	if !ok {
		return entity.Segmento{}, errs.ErrDadosInvalidos
	}
	return s, nil
}

// ---------- Recomendação ----------

type RecomendacaoRepository struct {
	mu        sync.RWMutex
	dados     map[uuid.UUID]entity.RecomendacaoProduto
	gravacoes int
}

func NovoRecomendacaoRepository() *RecomendacaoRepository {
	return &RecomendacaoRepository{dados: map[uuid.UUID]entity.RecomendacaoProduto{}}
}

func (r *RecomendacaoRepository) Salvar(_ context.Context, rec entity.RecomendacaoProduto) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.dados[rec.ID] = rec
	r.gravacoes++
	return nil
}

func (r *RecomendacaoRepository) ListarAtivasVencidas(_ context.Context, agora time.Time) ([]entity.RecomendacaoProduto, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []entity.RecomendacaoProduto{}
	for _, rec := range r.dados {
		if rec.Status == entity.StatusAtiva && rec.Validade != nil && !rec.Validade.After(agora) {
			out = append(out, rec)
		}
	}
	return out, nil
}

// ListarAtivasPorCliente devolve as recomendações ATIVA do cliente (Q-18).
func (r *RecomendacaoRepository) ListarAtivasPorCliente(_ context.Context, clienteID uuid.UUID) ([]entity.RecomendacaoProduto, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []entity.RecomendacaoProduto{}
	for _, rec := range r.dados {
		if rec.Status == entity.StatusAtiva && rec.ClienteID == clienteID {
			out = append(out, rec)
		}
	}
	return out, nil
}

// BuscarPorID e Gravacoes apoiam testes; não fazem parte do port.
func (r *RecomendacaoRepository) BuscarPorID(id uuid.UUID) (entity.RecomendacaoProduto, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rec, ok := r.dados[id]
	return rec, ok
}

func (r *RecomendacaoRepository) Gravacoes() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.gravacoes
}

// Todas devolve as recomendações gravadas (apoio a testes).
func (r *RecomendacaoRepository) Todas() []entity.RecomendacaoProduto {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]entity.RecomendacaoProduto, 0, len(r.dados))
	for _, rec := range r.dados {
		out = append(out, rec)
	}
	return out
}

// ---------- Consentimento ----------

type ConsentimentoRepository struct {
	mu    sync.RWMutex
	dados map[uuid.UUID]entity.ConsentimentoLGPD // chave: ClienteID
}

func NovoConsentimentoRepository() *ConsentimentoRepository {
	return &ConsentimentoRepository{dados: map[uuid.UUID]entity.ConsentimentoLGPD{}}
}

func (r *ConsentimentoRepository) Salvar(_ context.Context, c entity.ConsentimentoLGPD) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if atual, ok := r.dados[c.ClienteID]; ok {
		c.ID = atual.ID
	}
	r.dados[c.ClienteID] = c
	return nil
}

func (r *ConsentimentoRepository) BuscarPorCliente(_ context.Context, clienteID uuid.UUID) (entity.ConsentimentoLGPD, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.dados[clienteID]
	if !ok {
		return entity.ConsentimentoLGPD{}, errs.ErrConsentimentoNaoEncontrado
	}
	return c, nil
}

// Remover apaga o registro (apoio a testes: estado NAO_CONCEDIDO sem registro).
func (r *ConsentimentoRepository) Remover(clienteID uuid.UUID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.dados, clienteID)
}

// ---------- Produto ----------

type ProdutoRepository struct {
	mu    sync.RWMutex
	dados map[uuid.UUID]entity.Produto
}

func NovoProdutoRepository() *ProdutoRepository {
	return &ProdutoRepository{dados: map[uuid.UUID]entity.Produto{}}
}

// Adicionar insere ou substitui um produto do catálogo.
func (r *ProdutoRepository) Adicionar(p entity.Produto) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.dados[p.ID] = p
}

// Todos devolve o catálogo completo (apoio a testes).
func (r *ProdutoRepository) Todos() []entity.Produto {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]entity.Produto, 0, len(r.dados))
	for _, p := range r.dados {
		out = append(out, p)
	}
	return out
}

func (r *ProdutoRepository) ListarAtivosPorCategoria(_ context.Context, categoria string) ([]entity.Produto, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []entity.Produto{}
	for _, p := range r.dados {
		if p.Ativo && p.Categoria == categoria {
			out = append(out, p)
		}
	}
	return out, nil
}
