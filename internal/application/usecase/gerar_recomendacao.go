package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/domain/service"
	"perfila/internal/shared/errs"
)

// GerarRecomendacaoUseCase — UC-04 («extend» de UC-03). Spec: 3.5, RN-07, RN-08.
type GerarRecomendacaoUseCase struct {
	pre           Precondicoes
	perfis        port.PerfilRepository
	segmentos     port.SegmentoRepository
	recomendacao  *service.RecomendacaoService
	recomendacoes port.RecomendacaoRepository
	clock         port.Clock
	validadeDias  int
	log           *slog.Logger
}

// NovoGerarRecomendacaoUseCase injeta as dependências (C-08). log nil descarta os registros.
func NovoGerarRecomendacaoUseCase(pre Precondicoes, perfis port.PerfilRepository, segmentos port.SegmentoRepository,
	recomendacao *service.RecomendacaoService, recomendacoes port.RecomendacaoRepository,
	clock port.Clock, validadeDias int, log *slog.Logger) *GerarRecomendacaoUseCase {
	if log == nil {
		log = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &GerarRecomendacaoUseCase{pre: pre, perfis: perfis, segmentos: segmentos, recomendacao: recomendacao,
		recomendacoes: recomendacoes, clock: clock, validadeDias: validadeDias, log: log}
}

// Executar gera, disponibiliza e grava a recomendação uma única vez, já em ATIVA, expirando antes
// as ATIVA anteriores do cliente. Spec: UC-04 RN-08, Q-17, Q-18.
func (uc *GerarRecomendacaoUseCase) Executar(ctx context.Context, clienteID uuid.UUID) (entity.RecomendacaoProduto, error) {
	cliente, err := uc.pre.Verificar(ctx, clienteID)
	if err != nil {
		return entity.RecomendacaoProduto{}, err
	}
	perfil, err := uc.perfis.BuscarPorCliente(ctx, clienteID)
	if err != nil {
		return entity.RecomendacaoProduto{}, err
	}
	if perfil.SegmentoID == nil {
		return entity.RecomendacaoProduto{}, fmt.Errorf("%w: perfil ainda sem segmento", errs.ErrPerfilNaoEncontrado)
	}
	seg, err := uc.segmentos.BuscarPorID(ctx, *perfil.SegmentoID)
	if errors.Is(err, errs.ErrDadosInvalidos) {
		// O perfil aponta para um segmento que a configuração não tem: falha de configuração (Q-17).
		return entity.RecomendacaoProduto{}, fmt.Errorf("%w: segmento %s do perfil: %v", service.ErrSegmentoMalConfigurado, *perfil.SegmentoID, err)
	}
	if err != nil {
		return entity.RecomendacaoProduto{}, err
	}
	rec, err := uc.recomendacao.Recomendar(ctx, cliente, perfil, seg)
	if err != nil {
		return entity.RecomendacaoProduto{}, err
	}
	agora := uc.clock.Agora()
	if err := rec.Disponibilizar(agora, uc.validadeDias); err != nil {
		return entity.RecomendacaoProduto{}, err
	}
	// RN-08 (Q-18): no máximo uma ATIVA por cliente; as anteriores expiram antes da nova.
	anteriores, err := uc.recomendacoes.ListarAtivasPorCliente(ctx, clienteID)
	if err != nil {
		return entity.RecomendacaoProduto{}, err
	}
	for _, ant := range anteriores {
		if err := ant.Expirar(agora); err != nil {
			return entity.RecomendacaoProduto{}, err
		}
		if err := uc.recomendacoes.Salvar(ctx, ant); err != nil {
			return entity.RecomendacaoProduto{}, err
		}
	}
	if err := uc.recomendacoes.Salvar(ctx, rec); err != nil {
		return entity.RecomendacaoProduto{}, err
	}
	// Spec 3.4.2: notificação reduzida a log; somente IDs (C-09).
	uc.log.InfoContext(ctx, "recomendacao_ativada",
		"recomendacaoId", rec.ID.String(), "clienteId", rec.ClienteID.String(), "produtoId", rec.ProdutoID.String())
	return rec, nil
}
