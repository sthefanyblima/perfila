// Package service contém os serviços de domínio: calculam, e só leem ports quando precisam.
// Quem persiste resultados são os casos de uso (D-09).
//
// Spec: 3.6. Fonte: UML7 (Serviços de Domínio).
package service

import (
	"sort"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
)

// PerfilamentoService é puro: depende só do Clock e da meta de engajamento. Spec: RN-05.
type PerfilamentoService struct {
	clock       port.Clock
	metaEventos int
}

// NovoPerfilamentoService recebe ENGAJAMENTO_META_EVENTOS por injeção (C-08).
func NovoPerfilamentoService(clock port.Clock, metaEventos int) *PerfilamentoService {
	return &PerfilamentoService{clock: clock, metaEventos: metaEventos}
}

// GerarPerfil calcula o perfil a partir dos eventos do cliente. Spec: UC-02 RN-05.
func (s *PerfilamentoService) GerarPerfil(clienteID uuid.UUID, eventos []entity.EventoComportamental) entity.PerfilCliente {
	p := entity.PerfilCliente{ID: uuid.New(), ClienteID: clienteID}

	var soma float64
	contagem := map[string]int{}
	for _, e := range eventos {
		if e.Tipo != entity.TipoCompra {
			continue
		}
		p.FrequenciaCompras++
		soma += e.Valor
		if e.Categoria != "" {
			contagem[e.Categoria]++
		}
	}
	if p.FrequenciaCompras > 0 {
		p.TicketMedio = soma / float64(p.FrequenciaCompras)
	}
	p.CategoriaPreferida = categoriaMaisFrequente(contagem)
	p.CalcularScore(eventos, s.clock.Agora(), s.metaEventos)
	return p
}

// categoriaMaisFrequente resolve empates por ordem alfabética. Spec: RN-05.
func categoriaMaisFrequente(contagem map[string]int) string {
	categorias := make([]string, 0, len(contagem))
	for c := range contagem {
		categorias = append(categorias, c)
	}
	sort.Strings(categorias)
	melhor, maior := "", 0
	for _, c := range categorias {
		if contagem[c] > maior {
			melhor, maior = c, contagem[c]
		}
	}
	return melhor
}
