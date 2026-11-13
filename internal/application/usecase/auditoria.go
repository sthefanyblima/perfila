package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
)

// RegistrarAuditoriaUseCase grava uma linha da trilha do painel. Spec: 3.5, RN-13, Q-39.
type RegistrarAuditoriaUseCase struct {
	auditoria port.AuditoriaRepository
	clock     port.Clock
}

// NovoRegistrarAuditoriaUseCase injeta as dependências (C-08).
func NovoRegistrarAuditoriaUseCase(auditoria port.AuditoriaRepository, clock port.Clock) *RegistrarAuditoriaUseCase {
	return &RegistrarAuditoriaUseCase{auditoria: auditoria, clock: clock}
}

// Executar monta o registro (ID novo, Data do relógio) e grava. Spec: RN-13.
func (uc *RegistrarAuditoriaUseCase) Executar(ctx context.Context, operador string, acao entity.AcaoAuditoria, alvoTipo entity.TipoAlvo, alvoID uuid.UUID, resultado entity.ResultadoAuditoria) error {
	r, err := entity.NovoRegistroAuditoria(uuid.New(), uc.clock.Agora(), operador, acao, alvoTipo, alvoID, resultado)
	if err != nil {
		return err
	}
	return uc.auditoria.Registrar(ctx, r)
}

// PorPaginaAuditoria é o tamanho da página da tela Auditoria. Spec: 3.5.
const PorPaginaAuditoria = 50

// FiltroListagemAuditoria é o que o administrador escolhe na tela. Campos vazios não filtram;
// De <= Data < Ate. Spec: 3.5, Q-39.
type FiltroListagemAuditoria struct {
	Operador string
	Acao     entity.AcaoAuditoria
	De, Ate  time.Time
	Pagina   int // a partir de 1
}

// PaginaAuditoria é uma página da trilha, do mais recente para o mais antigo. Spec: 3.5.
type PaginaAuditoria struct {
	Registros []entity.RegistroAuditoria
	Total     int
	Pagina    int
}

// ListarAuditoriaUseCase é a tela Auditoria (permissão "auditoria"). Spec: 3.5, RN-11, RN-13.
type ListarAuditoriaUseCase struct{ auditoria port.AuditoriaRepository }

// NovoListarAuditoriaUseCase injeta as dependências (C-08).
func NovoListarAuditoriaUseCase(auditoria port.AuditoriaRepository) *ListarAuditoriaUseCase {
	return &ListarAuditoriaUseCase{auditoria: auditoria}
}

// Executar confere a permissão do ator e devolve a página pedida. Spec: RN-11, RN-13.
func (uc *ListarAuditoriaUseCase) Executar(ctx context.Context, ator entity.Operador, f FiltroListagemAuditoria) (PaginaAuditoria, error) {
	if err := exigir(ator, entity.PermAuditoria); err != nil {
		return PaginaAuditoria{}, err
	}
	pagina := max(f.Pagina, 1)
	regs, total, err := uc.auditoria.Listar(ctx, port.FiltroAuditoria{Operador: f.Operador, Acao: f.Acao, De: f.De, Ate: f.Ate,
		Deslocamento: (pagina - 1) * PorPaginaAuditoria, Limite: PorPaginaAuditoria})
	if err != nil {
		return PaginaAuditoria{}, err
	}
	return PaginaAuditoria{Registros: regs, Total: total, Pagina: pagina}, nil
}

// ExpurgarAuditoriaUseCase apaga o que passou da retenção (subcomando expurgar-auditoria).
// Spec: 3.5, RN-13, Q-39.
type ExpurgarAuditoriaUseCase struct {
	auditoria    port.AuditoriaRepository
	clock        port.Clock
	retencaoDias int
}

// NovoExpurgarAuditoriaUseCase injeta as dependências; retencaoDias vem de AUDITORIA_RETENCAO_DIAS (C-08).
func NovoExpurgarAuditoriaUseCase(auditoria port.AuditoriaRepository, clock port.Clock, retencaoDias int) *ExpurgarAuditoriaUseCase {
	return &ExpurgarAuditoriaUseCase{auditoria: auditoria, clock: clock, retencaoDias: retencaoDias}
}

// Executar apaga os registros com Data anterior a agora menos o prazo; um registro com exatamente
// o prazo permanece. Devolve quantos apagou. Spec: RN-13.
func (uc *ExpurgarAuditoriaUseCase) Executar(ctx context.Context) (int, error) {
	limite := uc.clock.Agora().AddDate(0, 0, -uc.retencaoDias)
	return uc.auditoria.ExpurgarAntesDe(ctx, limite)
}
