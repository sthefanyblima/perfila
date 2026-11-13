package entity

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"perfila/internal/shared/errs"
)

// PapelOperador define o que um operador do painel alcança. Spec: 3.2, RN-11, Q-35, Q-37.
type PapelOperador string

const (
	PapelAdministrador PapelOperador = "ADMINISTRADOR"
	PapelAtendimento   PapelOperador = "ATENDIMENTO"
	PapelAnalista      PapelOperador = "ANALISTA"
)

// Valido informa se o papel pertence ao conjunto permitido. Spec: RN-12.
func (p PapelOperador) Valido() bool {
	switch p {
	case PapelAdministrador, PapelAtendimento, PapelAnalista:
		return true
	}
	return false
}

// Permissao é uma linha da matriz da RN-11. Spec: 3.2, RN-11, Q-37.
type Permissao string

const (
	PermClientes      Permissao = "clientes"
	PermConsentimento Permissao = "consentimento"
	PermEventos       Permissao = "eventos"
	PermProcessamento Permissao = "processamento"
	PermConfiguracao  Permissao = "configuracao"
	PermOperacao      Permissao = "operacao"
	PermUsuarios      Permissao = "usuarios"
	PermAuditoria     Permissao = "auditoria"
	PermRelatorios    Permissao = "relatorios" // reservada para UC-05 a UC-07 (D-04): sem rota no MVP
	PermEliminacao    Permissao = "eliminacao" // reservada para a eliminação de dados (D-13): sem rota no MVP
)

// Pode aplica a matriz da RN-11: função pura, sem estado. Papel fora do enumerado não pode nada.
// Spec: RN-11, Q-37.
func (p PapelOperador) Pode(perm Permissao) bool {
	switch p {
	case PapelAdministrador:
		switch perm {
		case PermClientes, PermConsentimento, PermEventos, PermProcessamento, PermConfiguracao,
			PermOperacao, PermUsuarios, PermAuditoria, PermRelatorios, PermEliminacao:
			return true
		}
	case PapelAtendimento:
		switch perm {
		case PermClientes, PermConsentimento, PermEventos, PermProcessamento, PermConfiguracao:
			return true
		}
	case PapelAnalista:
		switch perm {
		case PermConfiguracao, PermRelatorios:
			return true
		}
	}
	return false
}

// Operador é a pessoa que opera o painel. Não é excluído: é desativado (RN-12).
// Spec: 3.2, RN-12, Q-35. Fonte: UML2 (v1.5).
type Operador struct {
	ID                uuid.UUID
	Login             string // único e imutável; minúsculas
	Nome              string // nunca vai para log nem auditoria (C-09)
	Papel             PapelOperador
	SenhaHash         string // gerado pelo port HashSenha; nunca a senha
	Ativo             bool
	CriadoEm          time.Time
	TrocarSenha       bool      // senha provisória (Q-40)
	SenhaAtualizadaEm time.Time // carimbo da sessão (Q-38)
}

// NormalizarLogin devolve o login como é gravado e comparado: sem espaços nas pontas e em
// minúsculas. Spec: RN-12.
func NormalizarLogin(login string) string { return strings.ToLower(strings.TrimSpace(login)) }

// loginValido: de 3 a 32 caracteres entre a-z, 0-9, ".", "-" e "_", começando por letra. Spec: RN-12.
func loginValido(login string) bool {
	if len(login) < 3 || len(login) > 32 || login[0] < 'a' || login[0] > 'z' {
		return false
	}
	for i := 0; i < len(login); i++ {
		c := login[i]
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '.' || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

// NovoOperador cria um operador ativo, com senha provisória (TrocarSenha) e CriadoEm e
// SenhaAtualizadaEm = agora. Spec: 3.2, RN-12.
func NovoOperador(id uuid.UUID, login, nome string, papel PapelOperador, senhaHash string, agora time.Time) (Operador, error) {
	login, nome = NormalizarLogin(login), strings.TrimSpace(nome)
	switch {
	case !loginValido(login):
		return Operador{}, fmt.Errorf("%w: login fora do formato", errs.ErrDadosInvalidos)
	case nome == "" || utf8.RuneCountInString(nome) > 100:
		return Operador{}, fmt.Errorf("%w: nome do operador vazio ou longo demais", errs.ErrDadosInvalidos)
	case !papel.Valido():
		return Operador{}, fmt.Errorf("%w: papel %q", errs.ErrDadosInvalidos, papel)
	case strings.TrimSpace(senhaHash) == "":
		return Operador{}, fmt.Errorf("%w: hash de senha vazio", errs.ErrDadosInvalidos)
	}
	agora = agora.UTC()
	return Operador{ID: id, Login: login, Nome: nome, Papel: papel, SenhaHash: senhaHash, Ativo: true,
		CriadoEm: agora, TrocarSenha: true, SenhaAtualizadaEm: agora}, nil
}

// MudarPapel troca o papel; vale na próxima requisição (Q-38). Spec: RN-12.
func (o *Operador) MudarPapel(p PapelOperador) error {
	if !p.Valido() {
		return fmt.Errorf("%w: papel %q", errs.ErrDadosInvalidos, p)
	}
	o.Papel = p
	return nil
}

// Desativar tira o acesso na hora (Q-38). Spec: RN-12.
func (o *Operador) Desativar() { o.Ativo = false }

// Ativar devolve o acesso. Spec: RN-12.
func (o *Operador) Ativar() { o.Ativo = true }

// DefinirSenha grava um novo hash e atualiza o carimbo, o que encerra as sessões abertas.
// trocar marca a senha como provisória. Spec: RN-12, Q-38, Q-40.
func (o *Operador) DefinirSenha(senhaHash string, trocar bool, agora time.Time) error {
	if strings.TrimSpace(senhaHash) == "" {
		return fmt.Errorf("%w: hash de senha vazio", errs.ErrDadosInvalidos)
	}
	o.SenhaHash, o.TrocarSenha, o.SenhaAtualizadaEm = senhaHash, trocar, agora.UTC()
	return nil
}
