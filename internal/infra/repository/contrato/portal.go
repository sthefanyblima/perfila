package contrato

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/shared/errs"
)

// Casos do contrato dos ports do portal do titular. Spec: v1.10, 3.6, RN-19, RN-20, Q-56 a Q-59.

func hashDe(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// Spec: Q-59 — sem diferenciar maiúsculas e espaços nas pontas; mais recente primeiro; vazio sem erro.
func testarBuscarPorEmail(t *testing.T, r Repositorios) {
	ctx := context.Background()
	antigo := entity.Cliente{ID: uuid.New(), Nome: "Ana Sintética", Email: "ana@exemplo.test", DataCadastro: agora.Add(-time.Hour), Ativo: true}
	novo := entity.Cliente{ID: uuid.New(), Nome: "Ana Sintética", Email: " Ana@Exemplo.TEST", DataCadastro: agora, Ativo: true}
	outro := entity.Cliente{ID: uuid.New(), Nome: "Bruno Sintético", Email: "bruno@exemplo.test", DataCadastro: agora, Ativo: true}
	for _, c := range []entity.Cliente{antigo, novo, outro} {
		ok(t, r.Clientes.Salvar(ctx, c))
	}
	achados, err := r.Clientes.BuscarPorEmail(ctx, "  ANA@exemplo.test ")
	ok(t, err)
	igual(t, "clientes com o e-mail", achados, []entity.Cliente{novo, antigo})
	for _, e := range []string{"ninguem@exemplo.test", "", "   ", "ana"} {
		achados, err = r.Clientes.BuscarPorEmail(ctx, e)
		ok(t, err)
		igual(t, "busca por "+e, len(achados), 0)
	}
}

// Spec: RN-19, Q-56 — uso único e atômico, validade de 30 minutos, contagem e remoção por cliente.
func testarTokens(t *testing.T, r Repositorios) {
	ctx := context.Background()
	c, outro := cliente(t, r), cliente(t, r)
	tok, err := entity.NovoTokenAcesso(uuid.New(), c.ID, hashDe("token-1"), agora)
	ok(t, err)
	ok(t, r.Tokens.Salvar(ctx, tok))

	_, err = r.Tokens.Consumir(ctx, hashDe("outro"), agora)
	erroEsperado(t, err, errs.ErrAcessoInvalido)
	_, err = r.Tokens.Consumir(ctx, tok.Hash, tok.ExpiraEm) // vence no instante de ExpiraEm
	erroEsperado(t, err, errs.ErrAcessoInvalido)
	usado, err := r.Tokens.Consumir(ctx, tok.Hash, agora.Add(29*time.Minute))
	ok(t, err)
	quer := tok
	quer.UsadoEm = ptr(agora.Add(29 * time.Minute))
	igual(t, "token consumido", usado, quer)
	_, err = r.Tokens.Consumir(ctx, tok.Hash, agora.Add(29*time.Minute))
	erroEsperado(t, err, errs.ErrAcessoInvalido)

	for i, criado := range []time.Time{agora.Add(10 * time.Minute), agora.Add(61 * time.Minute)} {
		t2, err := entity.NovoTokenAcesso(uuid.New(), c.ID, hashDe("token-"+string(rune('a'+i))), criado)
		ok(t, err)
		ok(t, r.Tokens.Salvar(ctx, t2))
	}
	t3, err := entity.NovoTokenAcesso(uuid.New(), outro.ID, hashDe("token-outro"), agora)
	ok(t, err)
	ok(t, r.Tokens.Salvar(ctx, t3))
	n, err := r.Tokens.ContarDesde(ctx, c.ID, agora)
	ok(t, err)
	igual(t, "tokens desde agora (inclusive)", n, 3)
	n, err = r.Tokens.ContarDesde(ctx, c.ID, agora.Add(time.Minute))
	ok(t, err)
	igual(t, "tokens depois de agora", n, 2)

	n, err = r.Tokens.RemoverPorCliente(ctx, c.ID)
	ok(t, err)
	igual(t, "tokens removidos", n, 3)
	n, err = r.Tokens.ContarDesde(ctx, c.ID, agora.Add(-time.Hour))
	ok(t, err)
	igual(t, "tokens depois da remoção", n, 0)
	n, err = r.Tokens.ContarDesde(ctx, outro.ID, agora.Add(-time.Hour))
	ok(t, err)
	igual(t, "tokens do outro cliente", n, 1)
}

// Spec: RN-20, Q-58 — um ABERTO por cliente, atender, abertos primeiro na lista.
func testarPedidos(t *testing.T, r Repositorios) {
	ctx := context.Background()
	c, outro := cliente(t, r), cliente(t, r)
	_, err := r.Pedidos.BuscarAbertoPorCliente(ctx, c.ID)
	erroEsperado(t, err, errs.ErrPedidoNaoEncontrado)

	p, err := entity.NovoPedidoEliminacao(uuid.New(), c.ID, agora)
	ok(t, err)
	ok(t, r.Pedidos.Salvar(ctx, p))
	aberto, err := r.Pedidos.BuscarAbertoPorCliente(ctx, c.ID)
	ok(t, err)
	igual(t, "pedido aberto", aberto, p)

	dup, err := entity.NovoPedidoEliminacao(uuid.New(), c.ID, agora.Add(time.Minute))
	ok(t, err)
	erroEsperado(t, r.Pedidos.Salvar(ctx, dup), errs.ErrTransicaoInvalida)

	ok(t, p.Atender(agora.Add(time.Hour)))
	ok(t, r.Pedidos.Salvar(ctx, p))
	_, err = r.Pedidos.BuscarAbertoPorCliente(ctx, c.ID)
	erroEsperado(t, err, errs.ErrPedidoNaoEncontrado)

	novo, err := entity.NovoPedidoEliminacao(uuid.New(), outro.ID, agora.Add(-time.Hour)) // mais antigo, mas aberto
	ok(t, err)
	ok(t, r.Pedidos.Salvar(ctx, novo))
	lista, err := r.Pedidos.Listar(ctx)
	ok(t, err)
	igual(t, "pedidos (abertos primeiro)", lista, []entity.PedidoEliminacao{novo, p})
}

// Spec: Q-57 — mais recente primeiro, filtro por data e remoção das antigas.
func testarCaixaSaida(t *testing.T, r Repositorios) {
	ctx := context.Background()
	msg := func(n int, quando time.Time) entity.MensagemSimulada {
		return entity.MensagemSimulada{ID: uuid.New(), Para: "c***@exemplo.test", Assunto: "Seu link de acesso ao Perfila",
			Corpo: "link " + string(rune('0'+n)), EnviadaEm: quando}
	}
	velha, ontem, hoje1, hoje2 := msg(1, agora.Add(-25*time.Hour)), msg(2, agora.Add(-23*time.Hour)), msg(3, agora), msg(4, agora)
	for _, m := range []entity.MensagemSimulada{velha, ontem, hoje1, hoje2} {
		ok(t, r.CaixaSaida.Guardar(ctx, m))
	}
	lista, err := r.CaixaSaida.Listar(ctx, agora.Add(-24*time.Hour))
	ok(t, err)
	igual(t, "mensagens das últimas 24 horas", lista, []entity.MensagemSimulada{hoje2, hoje1, ontem})
	n, err := r.CaixaSaida.RemoverAte(ctx, agora.Add(-24*time.Hour))
	ok(t, err)
	igual(t, "mensagens apagadas", n, 1)
	lista, err = r.CaixaSaida.Listar(ctx, time.Time{})
	ok(t, err)
	igual(t, "mensagens que ficaram", len(lista), 3)
}
