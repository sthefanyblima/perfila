package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/shared/errs"
)

type operadoresFalso struct{ o map[uuid.UUID]entity.Operador }

func (f *operadoresFalso) Salvar(_ context.Context, o entity.Operador) error {
	f.o[o.ID] = o
	return nil
}
func (f *operadoresFalso) BuscarPorID(_ context.Context, id uuid.UUID) (entity.Operador, error) {
	o, ok := f.o[id]
	if !ok {
		return entity.Operador{}, errs.ErrOperadorNaoEncontrado
	}
	return o, nil
}
func (f *operadoresFalso) BuscarPorLogin(context.Context, string) (entity.Operador, error) {
	return entity.Operador{}, errs.ErrOperadorNaoEncontrado
}
func (f *operadoresFalso) Listar(context.Context) ([]entity.Operador, error) { return nil, nil }
func (f *operadoresFalso) CriarSeVazio(context.Context, entity.Operador) (bool, error) {
	return false, nil
}

type relogioFalso struct{ t time.Time }

func (r relogioFalso) Agora() time.Time { return r.t }

// Spec: RN-12 Q-38 Q-72 — sair muda o carimbo: o cookie emitido antes deixa de valer (Conferir
// recusa), e entrar de novo emite um cookie com o carimbo novo, que vale.
func TestSairInvalidaSessaoAntiga(t *testing.T) {
	ctx := context.Background()
	entrou := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	casos := []struct {
		nome  string
		agora time.Time
	}{
		{"mais tarde", entrou.Add(10 * time.Minute)},
		{"no mesmo instante", entrou},
		{"relógio atrasado", entrou.Add(-time.Hour)},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			op, err := entity.NovoOperador(uuid.New(), "analista", "Analista", entity.PapelAnalista, "hash", entrou)
			if err != nil {
				t.Fatal(err)
			}
			repo := &operadoresFalso{o: map[uuid.UUID]entity.Operador{op.ID: op}}
			ger := NovoGerenciarOperadoresUseCase(repo, nil, relogioFalso{c.agora})
			aut := NovoAutenticarOperadorUseCase(repo, hashFalso{})

			if _, err := aut.Conferir(ctx, op.ID, op.SenhaAtualizadaEm); err != nil {
				t.Fatalf("a sessão aberta deveria valer: %v", err)
			}
			if err := ger.Sair(ctx, op); err != nil {
				t.Fatal(err)
			}
			if _, err := aut.Conferir(ctx, op.ID, op.SenhaAtualizadaEm); err == nil {
				t.Fatal("o cookie antigo continua valendo depois de sair")
			}
			atual := repo.o[op.ID]
			if _, err := aut.Conferir(ctx, op.ID, atual.SenhaAtualizadaEm); err != nil {
				t.Fatalf("o carimbo novo deveria valer: %v", err)
			}
			if atual.SenhaHash != "hash" || !atual.Ativo || atual.TrocarSenha != op.TrocarSenha {
				t.Fatal("sair não pode mexer em senha, estado nem papel")
			}
		})
	}
}

type hashFalso struct{}

func (hashFalso) Gerar(s string) (string, error) { return s, nil }
func (hashFalso) Conferir(h, s string) bool      { return h == s }
