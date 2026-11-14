// Package memoria implementa os ports em memória (fakes) para testes BDD, contrato e demonstração.
// Cada repositório é um valor independente, criado por construtor (sem estado global — C-08).
//
// Spec: T-13, 3.6.
package memoria

import (
	"context"
	"fmt"
	"sort"
	"strings"
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

// Listar aplica FiltroClientes: ID exato quando a busca é um UUID; senão, nome contém a busca,
// sem diferenciar maiúsculas e acentos (a mesma normalização do PostgreSQL). Spec: Q-23.
func (r *ClienteRepository) Listar(_ context.Context, f port.FiltroClientes) ([]entity.Cliente, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	busca := strings.TrimSpace(f.Busca)
	idBusca, errID := uuid.Parse(busca)
	termo := NormalizarBusca(busca)
	todos := []entity.Cliente{}
	for _, c := range r.dados {
		switch {
		case busca == "":
		case errID == nil:
			if c.ID != idBusca {
				continue
			}
		case !strings.Contains(NormalizarBusca(c.Nome), termo):
			continue
		}
		todos = append(todos, c)
	}
	sort.Slice(todos, func(i, j int) bool {
		if !todos[i].DataCadastro.Equal(todos[j].DataCadastro) {
			return todos[i].DataCadastro.After(todos[j].DataCadastro)
		}
		return todos[i].ID.String() < todos[j].ID.String()
	})
	total := len(todos)
	ini := min(max(f.Deslocamento, 0), total)
	fim := total
	if f.Limite > 0 {
		fim = min(ini+f.Limite, total)
	}
	return todos[ini:fim], total, nil
}

// NormalizarBusca põe em minúsculas e tira os acentos do português, como o translate() da
// migração 004 faz no PostgreSQL. Spec: Q-23.
func NormalizarBusca(s string) string {
	return strings.NewReplacer(
		"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"í", "i", "ì", "i", "î", "i", "ï", "i",
		"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
		"ú", "u", "ù", "u", "û", "u", "ü", "u", "ç", "c", "ñ", "n",
	).Replace(strings.ToLower(s))
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

// ListarPorCliente devolve todas as recomendações do cliente, a mais recente primeiro (Q-23).
func (r *RecomendacaoRepository) ListarPorCliente(_ context.Context, clienteID uuid.UUID) ([]entity.RecomendacaoProduto, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []entity.RecomendacaoProduto{}
	for _, rec := range r.dados {
		if rec.ClienteID == clienteID {
			out = append(out, rec)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].DataGeracao.Equal(out[j].DataGeracao) {
			return out[i].DataGeracao.After(out[j].DataGeracao)
		}
		return out[i].ID.String() < out[j].ID.String()
	})
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
	mu         sync.RWMutex
	dados      map[chaveConsentimento]entity.ConsentimentoLGPD // chave: (ClienteID, Finalidade), Q-48
	movimentos map[uuid.UUID][]entity.MovimentoConsentimento   // chave: ClienteID; ordem de gravação
	idsMov     map[uuid.UUID]bool
}

type chaveConsentimento struct {
	cliente    uuid.UUID
	finalidade entity.Finalidade
}

func NovoConsentimentoRepository() *ConsentimentoRepository {
	return &ConsentimentoRepository{dados: map[chaveConsentimento]entity.ConsentimentoLGPD{},
		movimentos: map[uuid.UUID][]entity.MovimentoConsentimento{}, idsMov: map[uuid.UUID]bool{}}
}

func (r *ConsentimentoRepository) Salvar(_ context.Context, c entity.ConsentimentoLGPD) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.salvar(c)
	return nil
}

func (r *ConsentimentoRepository) salvar(c entity.ConsentimentoLGPD) {
	k := chaveConsentimento{c.ClienteID, c.Finalidade}
	if atual, ok := r.dados[k]; ok {
		c.ID = atual.ID
	}
	r.dados[k] = c
}

// Registrar grava registro e movimento juntos, sob a mesma trava; nada muda se falhar. Spec: RN-10.
func (r *ConsentimentoRepository) Registrar(_ context.Context, c entity.ConsentimentoLGPD, m entity.MovimentoConsentimento) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.idsMov[m.ID] {
		return fmt.Errorf("memoria: movimento de consentimento %s já existe", m.ID)
	}
	r.salvar(c)
	r.idsMov[m.ID] = true
	m.Data = m.Data.UTC()
	r.movimentos[m.ClienteID] = append(r.movimentos[m.ClienteID], m)
	return nil
}

// ListarMovimentos devolve uma cópia em ordem cronológica; a ordenação estável mantém a ordem de
// gravação no empate. Spec: RN-10.
func (r *ConsentimentoRepository) ListarMovimentos(_ context.Context, clienteID uuid.UUID) ([]entity.MovimentoConsentimento, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := append([]entity.MovimentoConsentimento{}, r.movimentos[clienteID]...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Data.Before(out[j].Data) })
	return out, nil
}

func (r *ConsentimentoRepository) BuscarPorCliente(_ context.Context, clienteID uuid.UUID, finalidade entity.Finalidade) (entity.ConsentimentoLGPD, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.dados[chaveConsentimento{clienteID, finalidade}]
	if !ok {
		return entity.ConsentimentoLGPD{}, errs.ErrConsentimentoNaoEncontrado
	}
	return c, nil
}

// ListarPorCliente devolve os registros existentes, em ordem de Finalidade. Spec: 3.6, Q-48.
func (r *ConsentimentoRepository) ListarPorCliente(_ context.Context, clienteID uuid.UUID) ([]entity.ConsentimentoLGPD, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []entity.ConsentimentoLGPD{}
	for k, c := range r.dados {
		if k.cliente == clienteID {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Finalidade < out[j].Finalidade })
	return out, nil
}

// Remover apaga o registro de uma finalidade (apoio a testes: NAO_CONCEDIDO sem registro).
func (r *ConsentimentoRepository) Remover(clienteID uuid.UUID, finalidade entity.Finalidade) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.dados, chaveConsentimento{clienteID, finalidade})
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

// ListarTodos devolve o catálogo completo, por Categoria, Nome e ID (Q-23).
func (r *ProdutoRepository) ListarTodos(_ context.Context) ([]entity.Produto, error) {
	out := r.Todos()
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Categoria != b.Categoria {
			return a.Categoria < b.Categoria
		}
		if a.Nome != b.Nome {
			return a.Nome < b.Nome
		}
		return a.ID.String() < b.ID.String()
	})
	return out, nil
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
