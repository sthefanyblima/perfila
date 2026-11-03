package entity

import (
	"time"

	"github.com/google/uuid"
)

// JanelaEngajamento é a janela de RN-05 (constante, não variável mutável — C-08).
const JanelaEngajamento = 30 * 24 * time.Hour

// PerfilCliente resume o comportamento do cliente. Spec: 3.2, RN-05, D-06. Fonte: UML2.
type PerfilCliente struct {
	ID                 uuid.UUID
	ClienteID          uuid.UUID
	FrequenciaCompras  int
	TicketMedio        float64
	CategoriaPreferida string
	ScoreEngajamento   float64    // entre 0 e 1
	SegmentoID         *uuid.UUID // [ADIÇÃO] (D-06)
}

// CalcularScore define ScoreEngajamento = min(1, N/meta), onde N é o número de eventos
// com Timestamp em (agora-30d, agora]. meta <= 0 resulta em score 0. Spec: RN-05.
func (p *PerfilCliente) CalcularScore(eventos []EventoComportamental, agora time.Time, meta int) {
	if meta <= 0 {
		p.ScoreEngajamento = 0
		return
	}
	inicio := agora.Add(-JanelaEngajamento)
	n := 0
	for _, e := range eventos {
		if e.Timestamp.After(inicio) && !e.Timestamp.After(agora) {
			n++
		}
	}
	score := float64(n) / float64(meta)
	if score > 1 {
		score = 1
	}
	p.ScoreEngajamento = score
}

// AtualizarPerfil copia as métricas de novo e preserva ID, ClienteID e SegmentoID.
// Spec: 3.2, RN-05 (um perfil por cliente).
func (p *PerfilCliente) AtualizarPerfil(novo PerfilCliente) {
	p.FrequenciaCompras = novo.FrequenciaCompras
	p.TicketMedio = novo.TicketMedio
	p.CategoriaPreferida = novo.CategoriaPreferida
	p.ScoreEngajamento = novo.ScoreEngajamento
}
