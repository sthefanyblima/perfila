package web

import (
	"testing"
	"time"

	"perfila/internal/domain/entity"
)

// Spec: 3.7.1 RN-10 Q-34 — cada origem tem o seu texto; movimento reconstruído sem termo diz isso.
func TestDetalheMovimento(t *testing.T) {
	casos := []struct {
		m    entity.MovimentoConsentimento
		quer string
	}{
		{entity.MovimentoConsentimento{Finalidade: "segmentacao", VersaoTermo: "v1", Origem: entity.OrigemAPI}, "Finalidade segmentacao · termo v1 · pela API"},
		{entity.MovimentoConsentimento{Finalidade: "segmentacao", VersaoTermo: "v2", Origem: entity.OrigemPainel, Operador: "op-01"}, "Finalidade segmentacao · termo v2 · pelo painel, por op-01"},
		{entity.MovimentoConsentimento{Finalidade: "segmentacao", VersaoTermo: "v2", Origem: entity.OrigemPortal}, "Finalidade segmentacao · termo v2 · pelo portal do titular"},
		{entity.MovimentoConsentimento{Finalidade: "segmentacao", VersaoTermo: "v1", Origem: entity.OrigemSistema}, "Finalidade segmentacao · termo v1 · pelo sistema"},
		{entity.MovimentoConsentimento{Origem: entity.OrigemMigracao}, "termo não registrado · reconstruído das datas"},
	}
	for _, c := range casos {
		if got := detalheMovimento(c.m, textoOrigem); got != c.quer {
			t.Errorf("detalhe %q; esperado %q", got, c.quer)
		}
	}
}

// Spec: Q-34 — reconstrução (revogação antes da última concessão) e empate de datas na linha do tempo.
func TestLinhaPeloHistorico(t *testing.T) {
	dia := func(d int) time.Time { return time.Date(2026, 1, d, 10, 0, 0, 0, time.UTC) }
	d := dadosFicha{Cliente: entity.Cliente{DataCadastro: dia(1)}, Consentimento: entity.ConsentimentoLGPD{Status: entity.StatusConcedido},
		Movimentos: []entity.MovimentoConsentimento{
			{Tipo: entity.MovimentoRevogacao, Data: dia(3), Origem: entity.OrigemMigracao},
			{Tipo: entity.MovimentoConcessao, Data: dia(3), VersaoTermo: "v2", Origem: entity.OrigemMigracao},
		}}
	ms := d.LinhaConsentimento()
	titulos := []string{"Consentimento concedido novamente", "Consentimento revogado", "Cadastro"}
	if len(ms) != len(titulos) {
		t.Fatalf("marcos %+v", ms)
	}
	for i, quer := range titulos {
		if ms[i].Titulo != quer {
			t.Fatalf("marco %d: %q; esperado %q", i, ms[i].Titulo, quer)
		}
	}
	if d.PelasDatas() {
		t.Fatal("com movimentos, a linha não é a montada pelas datas")
	}
}
