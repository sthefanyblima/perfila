package memoria

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
)

var _ port.RelatorioRepository = (*RelatorioRepository)(nil)

// RelatorioRepository agrega os mesmos fakes em memória que os casos de uso gravam, com a mesma
// base do PostgreSQL: comportamento só de quem tem consentimento CONCEDIDO agora. Spec: 3.6,
// RN-15, Q-46.
type RelatorioRepository struct {
	clientes       *ClienteRepository
	eventos        *EventoRepository
	perfis         *PerfilRepository
	recomendacoes  *RecomendacaoRepository
	consentimentos *ConsentimentoRepository
}

// NovoRelatorioRepository lê os repositórios em memória informados (C-08).
func NovoRelatorioRepository(clientes *ClienteRepository, eventos *EventoRepository, perfis *PerfilRepository,
	recomendacoes *RecomendacaoRepository, consentimentos *ConsentimentoRepository) *RelatorioRepository {
	return &RelatorioRepository{clientes: clientes, eventos: eventos, perfis: perfis,
		recomendacoes: recomendacoes, consentimentos: consentimentos}
}

// validos devolve os clientes com consentimento CONCEDIDO (RN-02).
func (r *RelatorioRepository) validos() map[uuid.UUID]bool {
	r.consentimentos.mu.RLock()
	defer r.consentimentos.mu.RUnlock()
	v := map[uuid.UUID]bool{}
	for id, c := range r.consentimentos.dados {
		if c.Status == entity.StatusConcedido {
			v[id] = true
		}
	}
	return v
}

func (r *RelatorioRepository) ContarClientes(_ context.Context) (port.ContagemClientes, error) {
	r.consentimentos.mu.RLock()
	estados := map[uuid.UUID]entity.StatusConsentimento{}
	for id, c := range r.consentimentos.dados {
		estados[id] = c.Status
	}
	r.consentimentos.mu.RUnlock()
	r.clientes.mu.RLock()
	defer r.clientes.mu.RUnlock()
	var c port.ContagemClientes
	for id, cli := range r.clientes.dados {
		c.Total++
		if cli.Ativo {
			c.Ativos++
		}
		switch estados[id] {
		case entity.StatusConcedido:
			c.Concedido++
		case entity.StatusRevogado:
			c.Revogado++
		default: // sem registro = NAO_CONCEDIDO (D-02)
			c.NaoConcedido++
		}
	}
	return c, nil
}

func (r *RelatorioRepository) ResumirSegmentos(_ context.Context) ([]port.ResumoSegmento, error) {
	validos := r.validos()
	r.perfis.mu.RLock()
	defer r.perfis.mu.RUnlock()
	grupos := map[uuid.UUID]*port.ResumoSegmento{}
	var semSegmento *port.ResumoSegmento
	for cid, p := range r.perfis.dados {
		if !validos[cid] {
			continue
		}
		g := semSegmento
		if p.SegmentoID != nil {
			g = grupos[*p.SegmentoID]
		}
		if g == nil {
			g = &port.ResumoSegmento{}
			if p.SegmentoID != nil {
				sid := *p.SegmentoID
				g.SegmentoID = &sid
				grupos[sid] = g
			} else {
				semSegmento = g
			}
		}
		g.Clientes++
		g.SomaTicket += p.TicketMedio
		g.SomaScore += p.ScoreEngajamento
	}
	out := []port.ResumoSegmento{}
	for _, g := range grupos {
		out = append(out, *g)
	}
	if semSegmento != nil {
		out = append(out, *semSegmento)
	}
	return out, nil
}

func (r *RelatorioRepository) ContarRecomendacoes(_ context.Context) (port.ContagemRecomendacoes, error) {
	validos := r.validos()
	r.recomendacoes.mu.RLock()
	defer r.recomendacoes.mu.RUnlock()
	ativos, expirados := map[uuid.UUID]bool{}, map[uuid.UUID]bool{}
	var c port.ContagemRecomendacoes
	for _, rec := range r.recomendacoes.dados {
		if !validos[rec.ClienteID] {
			continue
		}
		switch rec.Status {
		case entity.StatusAtiva:
			c.Ativas++
			ativos[rec.ClienteID] = true
		case entity.StatusExpirada:
			c.Expiradas++
			expirados[rec.ClienteID] = true
		}
	}
	c.ClientesAtivas, c.ClientesExpiradas = len(ativos), len(expirados)
	return c, nil
}

// compras percorre as compras com categoria de clientes com consentimento válido em [desde, ate).
func (r *RelatorioRepository) compras(desde, ate time.Time, fn func(e entity.EventoComportamental)) {
	validos := r.validos()
	r.eventos.mu.RLock()
	defer r.eventos.mu.RUnlock()
	for _, e := range r.eventos.dados {
		if e.Tipo == entity.TipoCompra && e.Categoria != "" && validos[e.ClienteID] && dentro(e.Timestamp, desde, ate) {
			fn(e)
		}
	}
}

func dentro(t, desde, ate time.Time) bool { return !t.Before(desde) && t.Before(ate) }

func (r *RelatorioRepository) ContarCompras(_ context.Context, desde, ate time.Time) ([]port.ComprasCategoria, error) {
	compras := map[string]int{}
	clientes := map[string]map[uuid.UUID]bool{}
	r.compras(desde, ate, func(e entity.EventoComportamental) {
		compras[e.Categoria]++
		if clientes[e.Categoria] == nil {
			clientes[e.Categoria] = map[uuid.UUID]bool{}
		}
		clientes[e.Categoria][e.ClienteID] = true
	})
	out := []port.ComprasCategoria{}
	for cat, n := range compras {
		out = append(out, port.ComprasCategoria{Categoria: cat, Compras: n, Clientes: len(clientes[cat])})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Compras != out[j].Compras {
			return out[i].Compras > out[j].Compras
		}
		return out[i].Categoria < out[j].Categoria
	})
	return out, nil
}

func (r *RelatorioRepository) ContarCompradores(_ context.Context, desde, ate time.Time, categorias []string) (int, error) {
	quer := map[string]bool{}
	for _, c := range categorias {
		quer[c] = true
	}
	clientes := map[uuid.UUID]bool{}
	r.compras(desde, ate, func(e entity.EventoComportamental) {
		if quer[e.Categoria] {
			clientes[e.ClienteID] = true
		}
	})
	return len(clientes), nil
}

func (r *RelatorioRepository) SerieMensal(_ context.Context, limites []time.Time) ([]port.ContagemMes, error) {
	if len(limites) < 2 {
		return []port.ContagemMes{}, nil
	}
	out := make([]port.ContagemMes, len(limites)-1)
	clientesEventos := make([]map[uuid.UUID]bool, len(out))
	for i := range out {
		out[i].Inicio = limites[i].UTC()
		clientesEventos[i] = map[uuid.UUID]bool{}
	}
	indice := func(t time.Time) int {
		for i := range out {
			if dentro(t, limites[i], limites[i+1]) {
				return i
			}
		}
		return -1
	}
	r.clientes.mu.RLock()
	for _, c := range r.clientes.dados {
		if i := indice(c.DataCadastro); i >= 0 {
			out[i].Cadastros++
		}
	}
	r.clientes.mu.RUnlock()
	r.consentimentos.mu.RLock()
	for _, movs := range r.consentimentos.movimentos {
		for _, m := range movs {
			if i := indice(m.Data); m.Tipo == entity.MovimentoRevogacao && i >= 0 {
				out[i].Revogacoes++
			}
		}
	}
	r.consentimentos.mu.RUnlock()
	validos := r.validos()
	r.eventos.mu.RLock()
	for _, e := range r.eventos.dados {
		if i := indice(e.Timestamp); validos[e.ClienteID] && i >= 0 {
			out[i].Eventos++
			clientesEventos[i][e.ClienteID] = true
		}
	}
	r.eventos.mu.RUnlock()
	for i := range out {
		out[i].ClientesEventos = len(clientesEventos[i])
	}
	return out, nil
}
