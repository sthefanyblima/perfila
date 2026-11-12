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

// Spec: Q-30 — um cliente de exemplo por situação do painel, carregados uma única vez.
func TestCarregarExemplos(t *testing.T) {
	ctx := context.Background()
	app, err := montar(ctx, config.Config{AuthToken: "x", LogLevel: "info", RecomendacaoValidadeDias: 30, EngajamentoMetaEventos: 10}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	n, err := app.exemplos.carregar(ctx)
	if err != nil || n != 7 {
		t.Fatalf("carga: %d clientes, %v", n, err)
	}
	if n, err := app.exemplos.carregar(ctx); err != nil || n != 0 {
		t.Fatalf("segunda carga deveria ignorar: %d, %v", n, err)
	}
	e := app.exemplos
	repos := e.clientes
	lista, _, err := repos.Listar(ctx, port.FiltroClientes{})
	if err != nil {
		t.Fatal(err)
	}
	porEmail := map[string]entity.Cliente{}
	for _, c := range lista {
		porEmail[c.Email] = c
	}
	estado := func(email string) entity.StatusConsentimento {
		c, err := e.consentimento.Consultar(ctx, porEmail[email].ID)
		if err != nil {
			t.Fatal(err)
		}
		return c.Status
	}
	if estado("c-005@exemplo.test") != entity.StatusNaoConcedido || estado("c-006@exemplo.test") != entity.StatusRevogado {
		t.Fatal("Elisa deveria estar sem consentimento e Fábio revogado")
	}
	if porEmail["c-007@exemplo.test"].Ativo {
		t.Fatal("Gabi deveria estar inativa")
	}
	for email, segmento := range map[string]string{"c-001@exemplo.test": "premium", "c-002@exemplo.test": "engajado",
		"c-003@exemplo.test": "padrao", "c-004@exemplo.test": "padrao"} {
		res, err := e.processar.Executar(ctx, porEmail[email].ID)
		if err != nil || res.Segmento.Nome != segmento {
			t.Fatalf("%s: segmento %q, esperado %q (%v)", email, res.Segmento.Nome, segmento, err)
		}
		semProduto := res.Etapas[3].Estado == usecase.EtapaAtencao
		if semProduto != (email == "c-004@exemplo.test") {
			t.Fatalf("%s: etapa de recomendação %s", email, res.Etapas[3].Estado)
		}
	}
}
