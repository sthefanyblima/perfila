package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/shared/errs"
)

// validarSenha aplica a política da RN-12: de 12 caracteres a 72 bytes (limite do bcrypt).
// Spec: RN-12, Q-40.
func validarSenha(senha string) error {
	if utf8.RuneCountInString(senha) < 12 || len(senha) > 72 {
		return fmt.Errorf("%w: a senha precisa ter de 12 caracteres a 72 bytes", errs.ErrDadosInvalidos)
	}
	return nil
}

// AutenticarOperadorUseCase confere quem entra no painel e, a cada requisição, a sessão.
// Spec: 3.5, RN-12, Q-36, Q-38.
type AutenticarOperadorUseCase struct {
	operadores port.OperadorRepository
	hash       port.HashSenha
	ficticio   string // hash conferido quando o login não existe, para o tempo não revelar quem existe
}

// NovoAutenticarOperadorUseCase injeta as dependências (C-08).
func NovoAutenticarOperadorUseCase(operadores port.OperadorRepository, hash port.HashSenha) *AutenticarOperadorUseCase {
	ficticio, _ := hash.Gerar("senha-ficticia-para-tempo-constante")
	return &AutenticarOperadorUseCase{operadores: operadores, hash: hash, ficticio: ficticio}
}

// Executar devolve o operador ativo com aquele login e senha; senão, ErrCredenciaisInvalidas,
// sempre o mesmo erro (com o login existente, o operador vem junto para a auditoria). Spec: RN-12, RN-13.
func (uc *AutenticarOperadorUseCase) Executar(ctx context.Context, login, senha string) (entity.Operador, error) {
	o, err := uc.operadores.BuscarPorLogin(ctx, login)
	if errors.Is(err, errs.ErrOperadorNaoEncontrado) {
		uc.hash.Conferir(uc.ficticio, senha)
		return entity.Operador{}, errs.ErrCredenciaisInvalidas
	}
	if err != nil {
		return entity.Operador{}, err
	}
	if !uc.hash.Conferir(o.SenhaHash, senha) || !o.Ativo {
		// O login existe: devolve o operador junto com o erro só para a auditoria registrar quem
		// tentou (RN-13). Quem chama nunca o trata como autenticado.
		return o, errs.ErrCredenciaisInvalidas
	}
	return o, nil
}

// Conferir valida a sessão: operador existente, ativo e com a mesma senha de quando entrou
// (carimbo em microssegundos). Devolve o papel atual. Spec: Q-38.
func (uc *AutenticarOperadorUseCase) Conferir(ctx context.Context, id uuid.UUID, carimbo time.Time) (entity.Operador, error) {
	o, err := uc.operadores.BuscarPorID(ctx, id)
	if errors.Is(err, errs.ErrOperadorNaoEncontrado) {
		return entity.Operador{}, errs.ErrCredenciaisInvalidas
	}
	if err != nil {
		return entity.Operador{}, err
	}
	if !o.Ativo || o.SenhaAtualizadaEm.UnixMicro() != carimbo.UnixMicro() {
		return entity.Operador{}, errs.ErrCredenciaisInvalidas
	}
	return o, nil
}

// GerenciarOperadoresUseCase é a tela Usuários e Minha senha. O ator é o operador da sessão;
// a permissão "usuarios" é conferida de novo aqui (RN-11). Spec: 3.5, RN-12, Q-35, Q-40.
type GerenciarOperadoresUseCase struct {
	operadores port.OperadorRepository
	hash       port.HashSenha
	clock      port.Clock
}

// NovoGerenciarOperadoresUseCase injeta as dependências (C-08).
func NovoGerenciarOperadoresUseCase(operadores port.OperadorRepository, hash port.HashSenha, clock port.Clock) *GerenciarOperadoresUseCase {
	return &GerenciarOperadoresUseCase{operadores: operadores, hash: hash, clock: clock}
}

func exigir(ator entity.Operador, p entity.Permissao) error {
	if !ator.Ativo || !ator.Papel.Pode(p) {
		return errs.ErrSemPermissao
	}
	return nil
}

// Listar devolve os operadores por login. Spec: RN-11.
func (uc *GerenciarOperadoresUseCase) Listar(ctx context.Context, ator entity.Operador) ([]entity.Operador, error) {
	if err := exigir(ator, entity.PermUsuarios); err != nil {
		return nil, err
	}
	return uc.operadores.Listar(ctx)
}

// Consultar devolve um operador; ErrOperadorNaoEncontrado. Spec: RN-11.
func (uc *GerenciarOperadoresUseCase) Consultar(ctx context.Context, ator entity.Operador, id uuid.UUID) (entity.Operador, error) {
	if err := exigir(ator, entity.PermUsuarios); err != nil {
		return entity.Operador{}, err
	}
	return uc.operadores.BuscarPorID(ctx, id)
}

// Criar grava um operador novo com senha provisória (TrocarSenha). Spec: RN-12.
func (uc *GerenciarOperadoresUseCase) Criar(ctx context.Context, ator entity.Operador, login, nome string, papel entity.PapelOperador, senha string) (entity.Operador, error) {
	if err := exigir(ator, entity.PermUsuarios); err != nil {
		return entity.Operador{}, err
	}
	if err := validarSenha(senha); err != nil {
		return entity.Operador{}, err
	}
	if _, err := entity.NovoOperador(uuid.Nil, login, nome, papel, "-", uc.clock.Agora()); err != nil {
		return entity.Operador{}, err // valida antes de gastar o bcrypt
	}
	if _, err := uc.operadores.BuscarPorLogin(ctx, login); err == nil {
		return entity.Operador{}, errs.ErrLoginEmUso
	} else if !errors.Is(err, errs.ErrOperadorNaoEncontrado) {
		return entity.Operador{}, err
	}
	h, err := uc.hash.Gerar(senha)
	if err != nil {
		return entity.Operador{}, fmt.Errorf("gerando hash da senha: %w", err)
	}
	o, err := entity.NovoOperador(uuid.New(), login, nome, papel, h, uc.clock.Agora())
	if err != nil {
		return entity.Operador{}, err
	}
	return o, uc.operadores.Salvar(ctx, o)
}

// alterar carrega o operador, aplica a mudança e grava. propria indica se a mudança é
// proibida sobre a própria conta (ErrPropriaConta). Spec: RN-12.
func (uc *GerenciarOperadoresUseCase) alterar(ctx context.Context, ator entity.Operador, id uuid.UUID, propria bool, f func(*entity.Operador) error) error {
	if err := exigir(ator, entity.PermUsuarios); err != nil {
		return err
	}
	if propria && id == ator.ID {
		return errs.ErrPropriaConta
	}
	o, err := uc.operadores.BuscarPorID(ctx, id)
	if err != nil {
		return err
	}
	if err := f(&o); err != nil {
		return err
	}
	return uc.operadores.Salvar(ctx, o)
}

// MudarPapel vale na próxima requisição do operador (Q-38). Spec: RN-12.
func (uc *GerenciarOperadoresUseCase) MudarPapel(ctx context.Context, ator entity.Operador, id uuid.UUID, papel entity.PapelOperador) error {
	return uc.alterar(ctx, ator, id, true, func(o *entity.Operador) error { return o.MudarPapel(papel) })
}

// Desativar tira o acesso na hora (Q-38). Spec: RN-12.
func (uc *GerenciarOperadoresUseCase) Desativar(ctx context.Context, ator entity.Operador, id uuid.UUID) error {
	return uc.alterar(ctx, ator, id, true, func(o *entity.Operador) error { o.Desativar(); return nil })
}

// Ativar devolve o acesso. Spec: RN-12.
func (uc *GerenciarOperadoresUseCase) Ativar(ctx context.Context, ator entity.Operador, id uuid.UUID) error {
	return uc.alterar(ctx, ator, id, false, func(o *entity.Operador) error { o.Ativar(); return nil })
}

// RedefinirSenha grava uma senha provisória: TrocarSenha e fim das sessões abertas. Spec: RN-12, Q-40.
func (uc *GerenciarOperadoresUseCase) RedefinirSenha(ctx context.Context, ator entity.Operador, id uuid.UUID, senha string) error {
	if err := validarSenha(senha); err != nil {
		return err
	}
	return uc.alterar(ctx, ator, id, false, func(o *entity.Operador) error {
		h, err := uc.hash.Gerar(senha)
		if err != nil {
			return fmt.Errorf("gerando hash da senha: %w", err)
		}
		return o.DefinirSenha(h, true, uc.clock.Agora())
	})
}

// TrocarSenha é Minha senha: qualquer papel, só a própria, com a senha atual. Devolve o operador
// atualizado, para o painel renovar a sessão com o novo carimbo. Spec: RN-12, Q-40.
func (uc *GerenciarOperadoresUseCase) TrocarSenha(ctx context.Context, ator entity.Operador, atual, nova string) (entity.Operador, error) {
	o, err := uc.operadores.BuscarPorID(ctx, ator.ID)
	if err != nil {
		return entity.Operador{}, err
	}
	if !o.Ativo || !uc.hash.Conferir(o.SenhaHash, atual) {
		return entity.Operador{}, errs.ErrCredenciaisInvalidas
	}
	if err := validarSenha(nova); err != nil {
		return entity.Operador{}, err
	}
	h, err := uc.hash.Gerar(nova)
	if err != nil {
		return entity.Operador{}, fmt.Errorf("gerando hash da senha: %w", err)
	}
	if err := o.DefinirSenha(h, false, uc.clock.Agora()); err != nil {
		return entity.Operador{}, err
	}
	return o, uc.operadores.Salvar(ctx, o)
}

// CriarAdministradorInicial grava o ADMINISTRADOR de PAINEL_USUARIO e PAINEL_SENHA_HASH só se não
// houver nenhum operador (atômico no repositório). Sem TrocarSenha: a senha já foi gerada fora
// do sistema (Q-28). Spec: RN-12, Q-35.
func (uc *GerenciarOperadoresUseCase) CriarAdministradorInicial(ctx context.Context, login, senhaHash string) (bool, error) {
	o, err := entity.NovoOperador(uuid.New(), login, "Administrador", entity.PapelAdministrador, senhaHash, uc.clock.Agora())
	if err != nil {
		return false, err
	}
	o.TrocarSenha = false
	return uc.operadores.CriarSeVazio(ctx, o)
}
