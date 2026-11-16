package entity

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"perfila/internal/shared/errs"
)

// Spec: RN-11 Q-37 — a matriz de permissões, linha por linha.
func TestPapelPode(t *testing.T) {
	todas := []Permissao{PermClientes, PermConsentimento, PermEventos, PermProcessamento, PermConfiguracao,
		PermEditarConfiguracao, PermOperacao, PermUsuarios, PermAuditoria, PermRelatorios, PermEliminacao}
	matriz := map[PapelOperador][]Permissao{
		PapelAdministrador:     todas,
		PapelAtendimento:       {PermClientes, PermConsentimento, PermEventos, PermProcessamento, PermConfiguracao},
		PapelAnalista:          {PermConfiguracao, PermEditarConfiguracao, PermRelatorios},
		PapelOperador("CHEFE"): nil,
	}
	for papel, permitidas := range matriz {
		for _, perm := range todas {
			esperado := false
			for _, p := range permitidas {
				esperado = esperado || p == perm
			}
			if got := papel.Pode(perm); got != esperado {
				t.Errorf("%s.Pode(%s) = %v, esperado %v", papel, perm, got, esperado)
			}
		}
		if papel.Pode(Permissao("tudo")) {
			t.Errorf("%s pode uma permissão desconhecida", papel)
		}
	}
}

// Spec: 3.2 RN-12 Q-35 — invariantes do operador.
func TestNovoOperador(t *testing.T) {
	agora := time.Date(2026, 3, 10, 9, 0, 0, 0, time.FixedZone("BRT", -3*3600))
	casos := []struct {
		nome, login, nomeOp string
		papel               PapelOperador
		hash                string
		erro                error
		loginGravado        string
	}{
		{"válido", " Rita.Silva_1 ", "Rita", PapelAtendimento, "$2a$hash", nil, "rita.silva_1"},
		{"login mínimo", "abc", "A", PapelAnalista, "h", nil, "abc"},
		{"login com 32", "a" + strings.Repeat("b", 31), "A", PapelAnalista, "h", nil, "a" + strings.Repeat("b", 31)},
		{"login com 33", "a" + strings.Repeat("b", 32), "A", PapelAnalista, "h", errs.ErrDadosInvalidos, ""},
		{"login curto", "ab", "A", PapelAnalista, "h", errs.ErrDadosInvalidos, ""},
		{"começa por dígito", "1rita", "A", PapelAnalista, "h", errs.ErrDadosInvalidos, ""},
		{"acento", "rité", "A", PapelAnalista, "h", errs.ErrDadosInvalidos, ""},
		{"e-mail como login", "rita@exemplo.test", "A", PapelAnalista, "h", errs.ErrDadosInvalidos, ""},
		{"nome vazio", "rita", "  ", PapelAnalista, "h", errs.ErrDadosInvalidos, ""},
		{"nome longo", "rita", strings.Repeat("n", 101), PapelAnalista, "h", errs.ErrDadosInvalidos, ""},
		{"papel desconhecido", "rita", "A", PapelOperador("CHEFE"), "h", errs.ErrDadosInvalidos, ""},
		{"sem hash", "rita", "A", PapelAnalista, " ", errs.ErrDadosInvalidos, ""},
	}
	for _, tc := range casos {
		t.Run(tc.nome, func(t *testing.T) {
			o, err := NovoOperador(uuid.New(), tc.login, tc.nomeOp, tc.papel, tc.hash, agora)
			if !errors.Is(err, tc.erro) {
				t.Fatalf("erro %v, esperado %v", err, tc.erro)
			}
			if tc.erro != nil {
				return
			}
			if o.Login != tc.loginGravado || !o.Ativo || !o.TrocarSenha || !o.CriadoEm.Equal(agora) ||
				o.CriadoEm.Location() != time.UTC || !o.SenhaAtualizadaEm.Equal(agora) {
				t.Fatalf("operador inesperado: %+v", o)
			}
		})
	}
}

// Spec: RN-12 Q-38 Q-40 — papel, situação e senha.
func TestOperadorAlteracoes(t *testing.T) {
	agora := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	o, err := NovoOperador(uuid.New(), "rita", "Rita", PapelAtendimento, "h1", agora)
	if err != nil {
		t.Fatal(err)
	}
	if err := o.MudarPapel(PapelOperador("X")); !errors.Is(err, errs.ErrDadosInvalidos) || o.Papel != PapelAtendimento {
		t.Fatalf("papel inválido aceito: %v %s", err, o.Papel)
	}
	if err := o.MudarPapel(PapelAnalista); err != nil || o.Papel != PapelAnalista {
		t.Fatalf("mudar papel: %v %s", err, o.Papel)
	}
	o.Desativar()
	if o.Ativo {
		t.Fatal("continua ativo")
	}
	o.Ativar()
	if !o.Ativo {
		t.Fatal("continua inativo")
	}
	depois := agora.Add(time.Hour)
	if err := o.DefinirSenha("", false, depois); !errors.Is(err, errs.ErrDadosInvalidos) || o.SenhaHash != "h1" {
		t.Fatalf("hash vazio aceito: %v", err)
	}
	if err := o.DefinirSenha("h2", false, depois); err != nil || o.SenhaHash != "h2" || o.TrocarSenha || !o.SenhaAtualizadaEm.Equal(depois) {
		t.Fatalf("definir senha: %v %+v", err, o)
	}
}
