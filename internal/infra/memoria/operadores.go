package memoria

import (
	"context"
	"fmt"
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
	_ port.OperadorRepository  = (*OperadorRepository)(nil)
	_ port.AuditoriaRepository = (*AuditoriaRepository)(nil)
)

// ---------- Operador (v1.5, Q-35) ----------

type OperadorRepository struct {
	mu    sync.RWMutex
	dados map[uuid.UUID]entity.Operador
}

func NovoOperadorRepository() *OperadorRepository {
	return &OperadorRepository{dados: map[uuid.UUID]entity.Operador{}}
}

func (r *OperadorRepository) Salvar(_ context.Context, o entity.Operador) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.salvar(o)
}

func (r *OperadorRepository) salvar(o entity.Operador) error {
	o.Login = entity.NormalizarLogin(o.Login)
	for id, x := range r.dados {
		if x.Login == o.Login && id != o.ID {
			return fmt.Errorf("memoria: login de operador: %w", errs.ErrLoginEmUso)
		}
	}
	o.CriadoEm, o.SenhaAtualizadaEm = o.CriadoEm.UTC(), o.SenhaAtualizadaEm.UTC()
	r.dados[o.ID] = o
	return nil
}

func (r *OperadorRepository) BuscarPorID(_ context.Context, id uuid.UUID) (entity.Operador, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	o, ok := r.dados[id]
	if !ok {
		return entity.Operador{}, errs.ErrOperadorNaoEncontrado
	}
	return o, nil
}

func (r *OperadorRepository) BuscarPorLogin(_ context.Context, login string) (entity.Operador, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	login = entity.NormalizarLogin(login)
	for _, o := range r.dados {
		if o.Login == login {
			return o, nil
		}
	}
	return entity.Operador{}, errs.ErrOperadorNaoEncontrado
}

func (r *OperadorRepository) Listar(_ context.Context) ([]entity.Operador, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]entity.Operador, 0, len(r.dados))
	for _, o := range r.dados {
		out = append(out, o)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Login < out[j].Login })
	return out, nil
}

// CriarSeVazio grava sob a mesma trava da verificação: nunca dois administradores iniciais.
func (r *OperadorRepository) CriarSeVazio(_ context.Context, o entity.Operador) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.dados) > 0 {
		return false, nil
	}
	return true, r.salvar(o)
}

// ---------- Auditoria (v1.5, Q-39) ----------

type AuditoriaRepository struct {
	mu        sync.RWMutex
	registros []entity.RegistroAuditoria // ordem de gravação
	ids       map[uuid.UUID]bool
}

func NovoAuditoriaRepository() *AuditoriaRepository {
	return &AuditoriaRepository{ids: map[uuid.UUID]bool{}}
}

func (r *AuditoriaRepository) Registrar(_ context.Context, reg entity.RegistroAuditoria) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.ids[reg.ID] {
		return fmt.Errorf("memoria: registro de auditoria %s já existe", reg.ID)
	}
	r.ids[reg.ID] = true
	reg.Data = reg.Data.UTC()
	r.registros = append(r.registros, reg)
	return nil
}

// Listar filtra, ordena da Data mais recente para a mais antiga (no empate, a gravação mais
// recente primeiro) e pagina. Spec: 3.6, RN-13.
func (r *AuditoriaRepository) Listar(_ context.Context, f port.FiltroAuditoria) ([]entity.RegistroAuditoria, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var sel []entity.RegistroAuditoria
	for i := len(r.registros) - 1; i >= 0; i-- {
		x := r.registros[i]
		if (f.Operador != "" && x.Operador != entity.NormalizarLogin(f.Operador)) || (f.Acao != "" && x.Acao != f.Acao) ||
			(!f.De.IsZero() && x.Data.Before(f.De)) || (!f.Ate.IsZero() && !x.Data.Before(f.Ate)) {
			continue
		}
		sel = append(sel, x)
	}
	sort.SliceStable(sel, func(i, j int) bool { return sel[i].Data.After(sel[j].Data) })
	total := len(sel)
	ini := min(f.Deslocamento, total)
	fim := total
	if f.Limite > 0 {
		fim = min(ini+f.Limite, total)
	}
	return append([]entity.RegistroAuditoria{}, sel[ini:fim]...), total, nil
}

func (r *AuditoriaRepository) ExpurgarAntesDe(_ context.Context, limite time.Time) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	mantidos := r.registros[:0]
	n := 0
	for _, x := range r.registros {
		if x.Data.Before(limite) {
			delete(r.ids, x.ID)
			n++
			continue
		}
		mantidos = append(mantidos, x)
	}
	r.registros = mantidos
	return n, nil
}

// Todos devolve uma cópia em ordem de gravação (verificações dos testes).
func (r *AuditoriaRepository) Todos() []entity.RegistroAuditoria {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]entity.RegistroAuditoria{}, r.registros...)
}
