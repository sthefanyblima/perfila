package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/shared/errs"
)

// EstadoEtapa é o resultado de uma etapa do processamento. Spec: RN-09, Q-26.
type EstadoEtapa string

const (
	EtapaConcluida EstadoEtapa = "concluida"
	EtapaAtencao   EstadoEtapa = "atencao" // concluída sem recomendação (ErrSemProdutoElegivel)
	EtapaFalhou    EstadoEtapa = "falhou"
	EtapaPendente  EstadoEtapa = "pendente" // não executada porque uma anterior falhou
)

// NomeEtapa identifica as quatro etapas da sequência (UML4), na ordem. Spec: RN-09.
type NomeEtapa string

const (
	EtapaColetar      NomeEtapa = "coletar"
	EtapaPerfil       NomeEtapa = "perfil"
	EtapaSegmento     NomeEtapa = "segmento"
	EtapaRecomendacao NomeEtapa = "recomendacao"
)

// Etapa é o estado de uma etapa e, se falhou ou ficou em atenção, o erro que a explica.
type Etapa struct {
	Nome   NomeEtapa
	Estado EstadoEtapa
	Erro   error
}

// ResultadoProcessamento é a resposta consolidada do passo 23 da sequência. Spec: RN-09, Q-26.
type ResultadoProcessamento struct {
	Eventos      int
	Perfil       entity.PerfilCliente
	Segmento     entity.Segmento
	Recomendacao *entity.RecomendacaoProduto // nula sem produto elegível (RN-07)
	Etapas       [4]Etapa
}

// ProcessarClienteUseCase orquestra UC-01, UC-02, UC-03 e UC-04 (RN-09). Antes vivia no handler
// da API; agora API e painel usam o mesmo caso de uso. Spec: 3.5, RN-09, Q-26.
type ProcessarClienteUseCase struct {
	coletar    *ColetarDadosUseCase
	perfil     *GerarPerfilUseCase
	segmentar  *SegmentarClienteUseCase
	recomendar *GerarRecomendacaoUseCase
}

// NovoProcessarClienteUseCase injeta os quatro casos de uso (C-08).
func NovoProcessarClienteUseCase(coletar *ColetarDadosUseCase, perfil *GerarPerfilUseCase,
	segmentar *SegmentarClienteUseCase, recomendar *GerarRecomendacaoUseCase) *ProcessarClienteUseCase {
	return &ProcessarClienteUseCase{coletar: coletar, perfil: perfil, segmentar: segmentar, recomendar: recomendar}
}

// Executar roda as etapas na ordem; a primeira que falha interrompe, fica "falhou" e as seguintes
// "pendente". O que já foi persistido permanece. ErrSemProdutoElegivel deixa a última etapa em
// "atencao" sem erro. O error devolvido é o da etapa que falhou. Spec: RN-09.
func (uc *ProcessarClienteUseCase) Executar(ctx context.Context, clienteID uuid.UUID) (ResultadoProcessamento, error) {
	res := ResultadoProcessamento{Etapas: [4]Etapa{
		{Nome: EtapaColetar, Estado: EtapaPendente},
		{Nome: EtapaPerfil, Estado: EtapaPendente},
		{Nome: EtapaSegmento, Estado: EtapaPendente},
		{Nome: EtapaRecomendacao, Estado: EtapaPendente},
	}}
	falhar := func(i int, err error) (ResultadoProcessamento, error) {
		res.Etapas[i].Estado, res.Etapas[i].Erro = EtapaFalhou, err
		return res, err
	}

	eventos, err := uc.coletar.Executar(ctx, clienteID)
	if err != nil {
		return falhar(0, err)
	}
	res.Eventos = len(eventos)
	res.Etapas[0].Estado = EtapaConcluida

	if res.Perfil, err = uc.perfil.Executar(ctx, clienteID); err != nil {
		return falhar(1, err)
	}
	res.Etapas[1].Estado = EtapaConcluida

	if res.Segmento, err = uc.segmentar.Executar(ctx, clienteID); err != nil {
		return falhar(2, err)
	}
	segID := res.Segmento.ID
	res.Perfil.SegmentoID = &segID
	res.Etapas[2].Estado = EtapaConcluida

	rec, err := uc.recomendar.Executar(ctx, clienteID)
	switch {
	case err == nil:
		res.Recomendacao = &rec
		res.Etapas[3].Estado = EtapaConcluida
	case errors.Is(err, errs.ErrSemProdutoElegivel):
		res.Etapas[3].Estado, res.Etapas[3].Erro = EtapaAtencao, err
	default:
		return falhar(3, err)
	}
	return res, nil
}
