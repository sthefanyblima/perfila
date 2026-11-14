package memoria

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/shared/errs"
)

// PublicoRepository guarda exportações, conteúdo dos arquivos e membros dos públicos.
// Spec: 3.6, RN-17, Q-51.
type PublicoRepository struct {
	mu          sync.RWMutex
	exportacoes map[uuid.UUID]entity.ExportacaoPublico
	arquivos    map[uuid.UUID][]byte
	membros     map[uuid.UUID]entity.MembroPublico
	ordem       []uuid.UUID // ordem de gravação dos membros
}

func NovoPublicoRepository() *PublicoRepository {
	return &PublicoRepository{exportacoes: map[uuid.UUID]entity.ExportacaoPublico{},
		arquivos: map[uuid.UUID][]byte{}, membros: map[uuid.UUID]entity.MembroPublico{}}
}

// RegistrarExportacao grava exportação, conteúdo e membros sob a mesma trava: insere os membros
// novos e atualiza os existentes. Exportação repetida falha sem gravar nada.
func (r *PublicoRepository) RegistrarExportacao(_ context.Context, e entity.ExportacaoPublico, conteudo []byte, membros []entity.MembroPublico) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.exportacoes[e.ID]; ok {
		return fmt.Errorf("memoria: exportação %s já existe", e.ID)
	}
	e.Data, e.ArquivoExpiraEm = e.Data.UTC(), e.ArquivoExpiraEm.UTC()
	r.exportacoes[e.ID] = e
	r.arquivos[e.ID] = append([]byte{}, conteudo...)
	for _, m := range membros {
		if _, ok := r.membros[m.ID]; !ok {
			r.ordem = append(r.ordem, m.ID)
		}
		r.membros[m.ID] = normalizarMembro(m)
	}
	return nil
}

// ListarExportacoes devolve as exportações do segmento, da mais recente para a mais antiga.
func (r *PublicoRepository) ListarExportacoes(_ context.Context, segmentoID uuid.UUID) ([]entity.ExportacaoPublico, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []entity.ExportacaoPublico{}
	for _, e := range r.exportacoes {
		if e.SegmentoID == segmentoID {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].Data.Equal(out[j].Data) {
			return out[i].Data.After(out[j].Data)
		}
		return out[i].ID.String() < out[j].ID.String()
	})
	return out, nil
}

// BuscarArquivo devolve a exportação e o conteúdo; sem conteúdo, ErrArquivoIndisponivel.
func (r *PublicoRepository) BuscarArquivo(_ context.Context, exportacaoID uuid.UUID) (entity.ExportacaoPublico, []byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.arquivos[exportacaoID]
	if !ok {
		return entity.ExportacaoPublico{}, nil, errs.ErrArquivoIndisponivel
	}
	return r.exportacoes[exportacaoID], append([]byte{}, c...), nil
}

// ApagarArquivosVencidos apaga o conteúdo com ArquivoExpiraEm até agora; o registro fica. Spec: Q-51.
func (r *PublicoRepository) ApagarArquivosVencidos(_ context.Context, agora time.Time) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 0
	for id := range r.arquivos {
		if !r.exportacoes[id].ArquivoExpiraEm.After(agora) {
			delete(r.arquivos, id)
			n++
		}
	}
	return n, nil
}

// ListarMembrosPorSegmento devolve os membros do segmento no estado, em ordem de HashEmail.
func (r *PublicoRepository) ListarMembrosPorSegmento(_ context.Context, segmentoID uuid.UUID, status entity.StatusMembro) ([]entity.MembroPublico, error) {
	return r.filtrar(func(m entity.MembroPublico) bool { return m.SegmentoID == segmentoID && m.Status == status }), nil
}

// ListarMembrosPorCliente devolve os membros do cliente no estado, em ordem de HashEmail.
func (r *PublicoRepository) ListarMembrosPorCliente(_ context.Context, clienteID uuid.UUID, status entity.StatusMembro) ([]entity.MembroPublico, error) {
	return r.filtrar(func(m entity.MembroPublico) bool { return m.ClienteID == clienteID && m.Status == status }), nil
}

// AtualizarMembros grava o estado e as datas; ID inexistente falha sem gravar nada.
func (r *PublicoRepository) AtualizarMembros(_ context.Context, ms []entity.MembroPublico) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, m := range ms {
		if _, ok := r.membros[m.ID]; !ok {
			return fmt.Errorf("memoria: membro de público %s inexistente", m.ID)
		}
	}
	for _, m := range ms {
		atual := r.membros[m.ID]
		atual.Status, atual.DataRemocaoPendente, atual.RemocaoID, atual.DataRemocao = m.Status, m.DataRemocaoPendente, m.RemocaoID, m.DataRemocao
		r.membros[m.ID] = normalizarMembro(atual)
	}
	return nil
}

func (r *PublicoRepository) filtrar(ok func(entity.MembroPublico) bool) []entity.MembroPublico {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []entity.MembroPublico{}
	for _, id := range r.ordem {
		if m := r.membros[id]; ok(m) {
			out = append(out, m)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].HashEmail < out[j].HashEmail })
	return out
}

func normalizarMembro(m entity.MembroPublico) entity.MembroPublico {
	m.DataInclusao = m.DataInclusao.UTC()
	m.DataRemocaoPendente = utcPtr(m.DataRemocaoPendente)
	m.DataRemocao = utcPtr(m.DataRemocao)
	return m
}

func utcPtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	u := t.UTC()
	return &u
}
