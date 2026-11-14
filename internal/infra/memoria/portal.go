package memoria

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/shared/errs"
)

// Fakes do portal do titular. Spec: v1.10, 3.6, RN-19, RN-20, Q-56 a Q-59.
var (
	_ port.TokenAcessoRepository      = (*TokenAcessoRepository)(nil)
	_ port.PedidoEliminacaoRepository = (*PedidoEliminacaoRepository)(nil)
	_ port.CaixaSaidaSimulada         = (*CaixaSaidaSimulada)(nil)
)

// BuscarPorEmail compara sem maiúsculas e sem espaços nas pontas; cadastro mais recente primeiro.
// Spec: RN-19, Q-59.
func (r *ClienteRepository) BuscarPorEmail(_ context.Context, email string) ([]entity.Cliente, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	alvo := strings.ToLower(strings.TrimSpace(email))
	out := []entity.Cliente{}
	for _, c := range r.dados {
		if alvo != "" && strings.ToLower(strings.TrimSpace(c.Email)) == alvo {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].DataCadastro.Equal(out[j].DataCadastro) {
			return out[i].DataCadastro.After(out[j].DataCadastro)
		}
		return out[i].ID.String() < out[j].ID.String()
	})
	return out, nil
}

// ---------- Tokens de acesso ----------

type TokenAcessoRepository struct {
	mu    sync.Mutex
	dados map[uuid.UUID]entity.TokenAcessoTitular
}

func NovoTokenAcessoRepository() *TokenAcessoRepository {
	return &TokenAcessoRepository{dados: map[uuid.UUID]entity.TokenAcessoTitular{}}
}

func (r *TokenAcessoRepository) Salvar(_ context.Context, t entity.TokenAcessoTitular) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.dados[t.ID] = copiaToken(t)
	return nil
}

// Consumir procura pelo hash e marca o uso sob a trava: dois usos simultâneos, só um vale.
func (r *TokenAcessoRepository) Consumir(_ context.Context, hash string, agora time.Time) (entity.TokenAcessoTitular, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, t := range r.dados {
		if t.Hash != hash {
			continue
		}
		if err := t.Usar(agora); err != nil {
			return entity.TokenAcessoTitular{}, err
		}
		r.dados[id] = t
		return copiaToken(t), nil
	}
	return entity.TokenAcessoTitular{}, errs.ErrAcessoInvalido
}

func (r *TokenAcessoRepository) ContarDesde(_ context.Context, clienteID uuid.UUID, desde time.Time) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 0
	for _, t := range r.dados {
		if t.ClienteID == clienteID && !t.CriadoEm.Before(desde) {
			n++
		}
	}
	return n, nil
}

func (r *TokenAcessoRepository) RemoverPorCliente(_ context.Context, clienteID uuid.UUID) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 0
	for id, t := range r.dados {
		if t.ClienteID == clienteID {
			delete(r.dados, id)
			n++
		}
	}
	return n, nil
}

// Todos devolve os tokens gravados (asserções dos testes: só o hash é guardado).
func (r *TokenAcessoRepository) Todos() []entity.TokenAcessoTitular {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]entity.TokenAcessoTitular, 0, len(r.dados))
	for _, t := range r.dados {
		out = append(out, copiaToken(t))
	}
	return out
}

func copiaToken(t entity.TokenAcessoTitular) entity.TokenAcessoTitular {
	t.CriadoEm, t.ExpiraEm = t.CriadoEm.UTC(), t.ExpiraEm.UTC()
	if t.UsadoEm != nil {
		u := t.UsadoEm.UTC()
		t.UsadoEm = &u
	}
	return t
}

// ---------- Pedidos de eliminação ----------

type PedidoEliminacaoRepository struct {
	mu    sync.RWMutex
	dados map[uuid.UUID]entity.PedidoEliminacao
}

func NovoPedidoEliminacaoRepository() *PedidoEliminacaoRepository {
	return &PedidoEliminacaoRepository{dados: map[uuid.UUID]entity.PedidoEliminacao{}}
}

// Salvar recusa um segundo pedido ABERTO do mesmo cliente, como o índice único da migração 010.
func (r *PedidoEliminacaoRepository) Salvar(_ context.Context, p entity.PedidoEliminacao) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if p.Status == entity.PedidoAberto {
		for id, outro := range r.dados {
			if id != p.ID && outro.ClienteID == p.ClienteID && outro.Status == entity.PedidoAberto {
				return errs.ErrTransicaoInvalida
			}
		}
	}
	r.dados[p.ID] = copiaPedido(p)
	return nil
}

func (r *PedidoEliminacaoRepository) BuscarAbertoPorCliente(_ context.Context, clienteID uuid.UUID) (entity.PedidoEliminacao, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.dados {
		if p.ClienteID == clienteID && p.Status == entity.PedidoAberto {
			return copiaPedido(p), nil
		}
	}
	return entity.PedidoEliminacao{}, errs.ErrPedidoNaoEncontrado
}

func (r *PedidoEliminacaoRepository) Listar(_ context.Context) ([]entity.PedidoEliminacao, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]entity.PedidoEliminacao, 0, len(r.dados))
	for _, p := range r.dados {
		out = append(out, copiaPedido(p))
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if (a.Status == entity.PedidoAberto) != (b.Status == entity.PedidoAberto) {
			return a.Status == entity.PedidoAberto
		}
		if !a.DataPedido.Equal(b.DataPedido) {
			return a.DataPedido.After(b.DataPedido)
		}
		return a.ID.String() < b.ID.String()
	})
	return out, nil
}

func copiaPedido(p entity.PedidoEliminacao) entity.PedidoEliminacao {
	p.DataPedido = p.DataPedido.UTC()
	if p.DataAtendimento != nil {
		a := p.DataAtendimento.UTC()
		p.DataAtendimento = &a
	}
	return p
}

// ---------- Caixa de saída simulada ----------

type CaixaSaidaSimulada struct {
	mu        sync.RWMutex
	mensagens []entity.MensagemSimulada
}

func NovaCaixaSaidaSimulada() *CaixaSaidaSimulada { return &CaixaSaidaSimulada{} }

func (c *CaixaSaidaSimulada) Guardar(_ context.Context, m entity.MensagemSimulada) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	m.EnviadaEm = m.EnviadaEm.UTC()
	c.mensagens = append(c.mensagens, m)
	return nil
}

func (c *CaixaSaidaSimulada) Listar(_ context.Context, desde time.Time) ([]entity.MensagemSimulada, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := []entity.MensagemSimulada{}
	for i := len(c.mensagens) - 1; i >= 0; i-- { // ordem de gravação desempata o mesmo instante
		if m := c.mensagens[i]; !m.EnviadaEm.Before(desde) {
			out = append(out, m)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].EnviadaEm.After(out[j].EnviadaEm) })
	return out, nil
}

func (c *CaixaSaidaSimulada) RemoverAte(_ context.Context, limite time.Time) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	fica := c.mensagens[:0]
	for _, m := range c.mensagens {
		if !m.EnviadaEm.Before(limite) {
			fica = append(fica, m)
		}
	}
	n := len(c.mensagens) - len(fica)
	c.mensagens = fica
	return n, nil
}
