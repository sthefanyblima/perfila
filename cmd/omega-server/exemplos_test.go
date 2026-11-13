package main

import (
	"context"
	"io"
	"testing"

	"perfila/internal/application/usecase"
	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/infra/config"
)

// Spec: Q-30, Q-32 — clientes de exemplo com história, um por situação, carregados uma única vez.
func TestCarregarExemplos(t *testing.T) {
	for _, meta := range []int{10, 30} {
		ctx := context.Background()
		app, err := montar(ctx, config.Config{AuthToken: "x", LogLevel: "info", RecomendacaoValidadeDias: 30, EngajamentoMetaEventos: meta}, io.Discard)
		if err != nil {
			t.Fatal(err)
		}
		e := app.exemplos
		n, err := e.carregar(ctx)
		if err != nil || n != 14 {
			t.Fatalf("meta %d: carga criou %d clientes, %v", meta, n, err)
		}
		if n, err := e.carregar(ctx); err != nil || n != 0 {
			t.Fatalf("segunda carga deveria ignorar: %d, %v", n, err)
		}
		r := e.repos
		lista, _, _ := r.clientes.Listar(ctx, port.FiltroClientes{})
		porEmail := map[string]entity.Cliente{}
		for _, c := range lista {
			porEmail[c.Email] = c
		}
		listar := usecase.NovoListarClientesUseCase(r.clientes, r.consentimentos, r.perfis, r.segmentos)
		pagina, err := listar.Executar(ctx, usecase.FiltroListagem{})
		if err != nil {
			t.Fatal(err)
		}
		linhas := map[string]usecase.LinhaCliente{}
		for _, l := range pagina.Linhas {
			linhas[l.Cliente.Email] = l
		}
		esperado := map[string]string{
			"c-001@exemplo.test": "CONCEDIDO premium", "c-002@exemplo.test": "CONCEDIDO engajado",
			"c-003@exemplo.test": "CONCEDIDO ocasional", "c-004@exemplo.test": "CONCEDIDO ocasional",
			"c-005@exemplo.test": "NAO_CONCEDIDO ", "c-006@exemplo.test": "REVOGADO ",
			"c-007@exemplo.test": "CONCEDIDO ", "c-008@exemplo.test": "CONCEDIDO vip",
			"c-009@exemplo.test": "CONCEDIDO fiel", "c-010@exemplo.test": "CONCEDIDO explorador",
			"c-011@exemplo.test": "CONCEDIDO padrao", "c-012@exemplo.test": "CONCEDIDO ",
			"c-013@exemplo.test": "CONCEDIDO ocasional", "c-014@exemplo.test": "CONCEDIDO padrao",
		}
		for email, quer := range esperado {
			l := linhas[email]
			if got := string(l.Consentimento) + " " + l.Segmento; got != quer {
				t.Errorf("meta %d, %s: obtido %q, esperado %q", meta, email, got, quer)
			}
		}
		if porEmail["c-007@exemplo.test"].Ativo {
			t.Error("Gabi deveria estar inativa")
		}
		// Datas diferentes: cadastro, concessão e revogação não caem no mesmo instante.
		cons, _ := r.consentimentos.BuscarPorCliente(ctx, porEmail["c-006@exemplo.test"].ID)
		if cons.DataRevogacao == nil || !cons.DataConsentimento.Before(*cons.DataRevogacao) ||
			!porEmail["c-006@exemplo.test"].DataCadastro.Before(cons.DataConsentimento) {
			t.Errorf("Fábio: datas fora de ordem: %+v", cons)
		}
		karina, _ := r.consentimentos.BuscarPorCliente(ctx, porEmail["c-011@exemplo.test"].ID)
		if karina.VersaoTermo != "v2" || karina.DataRevogacao == nil || !karina.DataRevogacao.Before(karina.DataConsentimento) {
			t.Errorf("Karina deveria ter concedido de novo, termo v2: %+v", karina)
		}
		// Histórico de recomendações: Ana tem expiradas e uma ativa; Marina, só expirada.
		recs, _ := r.recomendacoes.ListarPorCliente(ctx, porEmail["c-001@exemplo.test"].ID)
		if len(recs) != 3 || recs[0].Status != entity.StatusAtiva || recs[2].Status != entity.StatusExpirada {
			t.Errorf("Ana: histórico de recomendações inesperado: %d", len(recs))
		}
		marina, _ := r.recomendacoes.ListarPorCliente(ctx, porEmail["c-013@exemplo.test"].ID)
		if len(marina) != 1 || marina[0].Status != entity.StatusExpirada {
			t.Errorf("Marina deveria ter só uma recomendação expirada: %+v", marina)
		}
	}
}
