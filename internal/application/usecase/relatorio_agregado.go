package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"sort"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/domain/service"
	"perfila/internal/shared/errs"
)

// RelatorioAgregadoSegmentosUseCase conta, por segmento ativo, quem tem segmentacao válida, sem
// nenhum identificador; grupos com menos de entity.TamanhoMinimoGrupo ficam ocultos. Spec: RN-18, Q-52.
type RelatorioAgregadoSegmentosUseCase struct {
	segmentos port.SegmentoRepository
	perfis    port.PerfilRepository
	clientes  port.ClienteRepository
	validador *service.ValidadorLGPD
}

// NovoRelatorioAgregadoSegmentosUseCase injeta as dependências (C-08).
func NovoRelatorioAgregadoSegmentosUseCase(segmentos port.SegmentoRepository, perfis port.PerfilRepository,
	clientes port.ClienteRepository, validador *service.ValidadorLGPD) *RelatorioAgregadoSegmentosUseCase {
	return &RelatorioAgregadoSegmentosUseCase{segmentos: segmentos, perfis: perfis, clientes: clientes, validador: validador}
}

// Executar devolve uma linha por segmento ativo, em ordem de Prioridade. Spec: RN-18.
func (uc *RelatorioAgregadoSegmentosUseCase) Executar(ctx context.Context) ([]entity.LinhaAgregada, error) {
	segs, err := uc.segmentos.ListarAtivos(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]entity.LinhaAgregada, 0, len(segs))
	for _, s := range segs {
		perfis, err := uc.perfis.ListarPorSegmento(ctx, s.ID)
		if err != nil {
			return nil, err
		}
		n, soma, categorias := 0, 0.0, map[string]int{}
		for _, p := range perfis {
			conta, err := uc.conta(ctx, p)
			if err != nil {
				return nil, err
			}
			if !conta {
				continue
			}
			n++
			soma += p.TicketMedio
			if p.CategoriaPreferida != "" {
				categorias[p.CategoriaPreferida]++
			}
		}
		linha := entity.LinhaAgregada{SegmentoNome: s.Nome}
		if n < entity.TamanhoMinimoGrupo {
			linha.Oculta = true
		} else {
			linha.Quantidade, linha.TicketMedio, linha.CategoriaMaisComum = n, soma/float64(n), maisComum(categorias)
		}
		out = append(out, linha)
	}
	return out, nil
}

func (uc *RelatorioAgregadoSegmentosUseCase) conta(ctx context.Context, p entity.PerfilCliente) (bool, error) {
	ok, err := uc.validador.ConsentimentoValido(ctx, p.ClienteID)
	if err != nil || !ok {
		return false, err
	}
	c, err := uc.clientes.BuscarPorID(ctx, p.ClienteID)
	if errors.Is(err, errs.ErrClienteNaoEncontrado) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return c.Ativo, nil
}

// maisComum: a categoria mais frequente; empate por ordem alfabética.
func maisComum(contagem map[string]int) string {
	nomes := make([]string, 0, len(contagem))
	for k := range contagem {
		nomes = append(nomes, k)
	}
	sort.Strings(nomes)
	melhor := ""
	for _, k := range nomes {
		if melhor == "" || contagem[k] > contagem[melhor] {
			melhor = k
		}
	}
	return melhor
}

// CSV monta o arquivo do agregado; grupo oculto sai como "menos de 5" com os outros campos vazios.
// Spec: RN-18.
func (uc *RelatorioAgregadoSegmentosUseCase) CSV(linhas []entity.LinhaAgregada) []byte {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"segmento", "clientes", "ticket_medio", "categoria_mais_comum"})
	for _, l := range linhas {
		if l.Oculta {
			_ = w.Write([]string{l.SegmentoNome, fmt.Sprintf("menos de %d", entity.TamanhoMinimoGrupo), "", ""})
			continue
		}
		_ = w.Write([]string{l.SegmentoNome, fmt.Sprint(l.Quantidade), fmt.Sprintf("%.2f", l.TicketMedio), l.CategoriaMaisComum})
	}
	w.Flush()
	return buf.Bytes()
}
