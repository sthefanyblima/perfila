package usecase

import (
	"context"

	"perfila/internal/domain/port"
)

// ExpirarRecomendacoesUseCase [ADIÇÃO, D-08] é o que o subcomando CMD
// `expirar-recomendacoes` executa. Spec: 3.4.2, RN-08.
type ExpirarRecomendacoesUseCase struct {
	recomendacoes port.RecomendacaoRepository
	clock         port.Clock
}

// NovoExpirarRecomendacoesUseCase injeta as dependências (C-08).
func NovoExpirarRecomendacoesUseCase(recomendacoes port.RecomendacaoRepository, clock port.Clock) *ExpirarRecomendacoesUseCase {
	return &ExpirarRecomendacoesUseCase{recomendacoes: recomendacoes, clock: clock}
}

// Executar expira as recomendações ATIVA com Validade <= agora e devolve quantas expirou.
func (uc *ExpirarRecomendacoesUseCase) Executar(ctx context.Context) (int, error) {
	agora := uc.clock.Agora()
	vencidas, err := uc.recomendacoes.ListarAtivasVencidas(ctx, agora)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, r := range vencidas {
		if err := r.Expirar(agora); err != nil {
			return n, err
		}
		if err := uc.recomendacoes.Salvar(ctx, r); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
