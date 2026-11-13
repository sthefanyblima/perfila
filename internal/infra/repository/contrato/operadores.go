package contrato

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/shared/errs"
)

func operador(t *testing.T, login string, papel entity.PapelOperador) entity.Operador {
	t.Helper()
	o, err := entity.NovoOperador(uuid.New(), login, "Operador Sintético", papel, "$2a$12$hashsintetico", agora)
	ok(t, err)
	return o
}

// Spec: 3.6 RN-12 Q-35 — operadores: login único sem diferenciar maiúsculas, upsert por ID,
// lista por login e criação atômica do administrador inicial.
func testarOperador(t *testing.T, r Repositorios) {
	ctx := context.Background()
	if r.Operadores == nil {
		t.Skip("backend sem OperadorRepository")
	}
	_, err := r.Operadores.BuscarPorID(ctx, uuid.New())
	erroEsperado(t, err, errs.ErrOperadorNaoEncontrado)

	inicial := operador(t, "inicial", entity.PapelAdministrador)
	criou, err := r.Operadores.CriarSeVazio(ctx, inicial)
	ok(t, err)
	igual(t, "criou o primeiro", criou, true)
	criou, err = r.Operadores.CriarSeVazio(ctx, operador(t, "segundo", entity.PapelAdministrador))
	ok(t, err)
	igual(t, "criou com operador existente", criou, false)

	rita := operador(t, "rita", entity.PapelAtendimento)
	ok(t, r.Operadores.Salvar(ctx, rita))
	lido, err := r.Operadores.BuscarPorLogin(ctx, " RITA ")
	ok(t, err)
	igual(t, "operador lido pelo login", lido, rita)

	outro := operador(t, "outra", entity.PapelAnalista)
	outro.Login = "Rita"
	erroEsperado(t, r.Operadores.Salvar(ctx, outro), errs.ErrLoginEmUso)

	ok(t, rita.MudarPapel(entity.PapelAnalista))
	rita.Desativar()
	ok(t, rita.DefinirSenha("$2a$12$outrohash", false, agora.Add(time.Hour)))
	ok(t, r.Operadores.Salvar(ctx, rita))
	lido, err = r.Operadores.BuscarPorID(ctx, rita.ID)
	ok(t, err)
	igual(t, "operador atualizado", lido, rita)

	todos, err := r.Operadores.Listar(ctx)
	ok(t, err)
	igual(t, "lista por login", []string{todos[0].Login, todos[1].Login}, []string{"inicial", "rita"})
}

func registro(t *testing.T, data time.Time, operador string, acao entity.AcaoAuditoria) entity.RegistroAuditoria {
	t.Helper()
	r, err := entity.NovoRegistroAuditoria(uuid.New(), data, operador, acao, entity.AlvoCliente, uuid.New(), entity.ResultadoSucesso)
	ok(t, err)
	return r
}

func ids(rs []entity.RegistroAuditoria) []uuid.UUID {
	out := make([]uuid.UUID, len(rs))
	for i, r := range rs {
		out[i] = r.ID
	}
	return out
}

// Spec: 3.6 RN-13 Q-39 — auditoria: só inclusão, ordem, filtros, página e expurgo.
func testarAuditoria(t *testing.T, r Repositorios) {
	ctx := context.Background()
	if r.Auditoria == nil {
		t.Skip("backend sem AuditoriaRepository")
	}
	a := registro(t, agora.Add(-2*time.Hour), "admin", entity.AcaoEntrar)
	b := registro(t, agora.Add(-time.Hour), "rita", entity.AcaoLerFicha)
	c := registro(t, agora.Add(-time.Hour), "rita", entity.AcaoProcessar) // mesmo instante: gravado depois
	d := registro(t, agora, "admin", entity.AcaoLerFicha)
	semAlvo, err := entity.NovoRegistroAuditoria(uuid.New(), agora.Add(-3*time.Hour), "", entity.AcaoEntrar,
		entity.AlvoNenhum, uuid.Nil, entity.ResultadoNegado)
	ok(t, err)
	for _, x := range []entity.RegistroAuditoria{semAlvo, a, b, c, d} {
		ok(t, r.Auditoria.Registrar(ctx, x))
	}
	if err := r.Auditoria.Registrar(ctx, a); err == nil {
		t.Fatal("registro com ID repetido foi aceito")
	}

	todos, total, err := r.Auditoria.Listar(ctx, port.FiltroAuditoria{})
	ok(t, err)
	igual(t, "total", total, 5)
	igual(t, "ordem", ids(todos), ids([]entity.RegistroAuditoria{d, c, b, a, semAlvo}))
	igual(t, "registro lido", todos[4], semAlvo)

	casos := []struct {
		nome     string
		f        port.FiltroAuditoria
		esperado []entity.RegistroAuditoria
		total    int
	}{
		{"operador", port.FiltroAuditoria{Operador: "RITA"}, []entity.RegistroAuditoria{c, b}, 2},
		{"ação", port.FiltroAuditoria{Acao: entity.AcaoLerFicha}, []entity.RegistroAuditoria{d, b}, 2},
		{"período", port.FiltroAuditoria{De: agora.Add(-time.Hour), Ate: agora}, []entity.RegistroAuditoria{c, b}, 2},
		{"página", port.FiltroAuditoria{Deslocamento: 1, Limite: 2}, []entity.RegistroAuditoria{c, b}, 5},
		{"além do fim", port.FiltroAuditoria{Deslocamento: 9, Limite: 2}, []entity.RegistroAuditoria{}, 5},
	}
	for _, tc := range casos {
		obtidos, n, err := r.Auditoria.Listar(ctx, tc.f)
		ok(t, err)
		igual(t, tc.nome+": total", n, tc.total)
		igual(t, tc.nome+": registros", ids(obtidos), ids(tc.esperado))
	}

	n, err := r.Auditoria.ExpurgarAntesDe(ctx, agora.Add(-time.Hour))
	ok(t, err)
	igual(t, "expurgados", n, 2)
	_, total, err = r.Auditoria.Listar(ctx, port.FiltroAuditoria{})
	ok(t, err)
	igual(t, "restantes", total, 3)
}
