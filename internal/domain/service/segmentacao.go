package service

import (
	"context"
	"fmt"
	"sort"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
)

// SegmentacaoService escolhe o segmento de um perfil. Spec: UC-03 RN-06.
type SegmentacaoService struct {
	segmentos port.SegmentoRepository
}

// NovoSegmentacaoService injeta o repositório de segmentos (C-08).
func NovoSegmentacaoService(segmentos port.SegmentoRepository) *SegmentacaoService {
	return &SegmentacaoService{segmentos: segmentos}
}

// Segmentar avalia os segmentos ativos em ordem crescente de Prioridade; o primeiro
// cujo critério casa vence. Nenhum segmento aplicável é falha de configuração (falta o
// segmento padrao), não erro do cliente: devolve ErrSemSegmentoAplicavel, que não é sentinela
// de domínio e vira 500 na API (Q-02, decidida em 07/11). Spec: RN-06.
func (s *SegmentacaoService) Segmentar(ctx context.Context, p entity.PerfilCliente) (entity.Segmento, error) {
	lista, err := s.segmentos.ListarAtivos(ctx)
	if err != nil {
		return entity.Segmento{}, err
	}
	sort.SliceStable(lista, func(i, j int) bool { return lista[i].Prioridade < lista[j].Prioridade })
	for _, seg := range lista {
		if !seg.Ativo {
			continue
		}
		casa, err := seg.ClassificarPerfil(p)
		if err != nil {
			// O critério veio da configuração, não do cliente: falha de configuração (Q-17).
			return entity.Segmento{}, fmt.Errorf("%w: segmento %s: %v", ErrSegmentoMalConfigurado, seg.Nome, err)
		}
		if casa {
			return seg, nil
		}
	}
	return entity.Segmento{}, ErrSemSegmentoAplicavel
}

// ErrSemSegmentoAplicavel indica configuração incompleta: o segmento padrao (critério vazio) deve existir.
const ErrSemSegmentoAplicavel = erroConfiguracao("nenhum segmento ativo casa com o perfil: o segmento padrao (critério vazio) não está configurado")

// ErrSegmentoMalConfigurado indica segmento configurado com critério inválido ou inexistente
// (falha de configuração, HTTP 500). Spec: RN-06, Q-17.
const ErrSegmentoMalConfigurado = erroConfiguracao("segmento mal configurado")

type erroConfiguracao string

func (e erroConfiguracao) Error() string { return string(e) }
