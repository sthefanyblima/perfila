package main

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"perfila/internal/application/usecase"
	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/domain/service"
	"perfila/internal/infra/config"
	"perfila/internal/infra/relogio"
	"perfila/internal/shared/errs"
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
		// Histórico (Q-34): concede v1 pela API, revoga e concede v2 pelo painel; os outros, pelo sistema.
		movs, _ := r.consentimentos.ListarMovimentos(ctx, porEmail["c-011@exemplo.test"].ID)
		var resumo []string
		for _, m := range movs {
			resumo = append(resumo, string(m.Tipo)+" "+m.VersaoTermo+" "+string(m.Origem)+" "+m.Operador)
		}
		quer := []string{"CONCESSAO v1 api ", "REVOGACAO v1 painel operador", "CONCESSAO v2 painel operador"}
		if strings.Join(resumo, "|") != strings.Join(quer, "|") {
			t.Errorf("Karina: histórico %q, esperado %q", resumo, quer)
		}
		if fabio, _ := r.consentimentos.ListarMovimentos(ctx, porEmail["c-006@exemplo.test"].ID); len(fabio) != 2 || fabio[1].Origem != entity.OrigemSistema {
			t.Errorf("Fábio: histórico %+v", fabio)
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

// Spec: RN-16, Q-47 — o arquivo de exemplo cita os IDs fixos dos clientes de exemplo e mostra
// cada situação da importação: aceitas, duplicada e rejeitadas por consentimento, cadastro e formato.
func TestArquivoDeExemploImporta(t *testing.T) {
	ctx := context.Background()
	app, err := montar(ctx, config.Config{AuthToken: "x", LogLevel: "info", RecomendacaoValidadeDias: 30, EngajamentoMetaEventos: 10}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.exemplos.carregar(ctx); err != nil {
		t.Fatal(err)
	}
	conteudo, err := os.ReadFile("../../docs/exemplos/eventos-lote.csv")
	if err != nil {
		t.Fatal(err)
	}
	entradas, err := usecase.LerCSVEventos(conteudo)
	if err != nil {
		t.Fatal(err)
	}
	r := app.exemplos.repos
	pre := usecase.NovasPrecondicoes(r.clientes, service.NovoValidadorLGPD(r.consentimentos))
	importar := usecase.NovoImportarEventosUseCase(usecase.NovoColetarDadosUseCase(pre, r.eventos, relogio.Sistema{}), pre, r.eventos, relogio.Sistema{})
	res, err := importar.Previa(ctx, entradas)
	if err != nil || res.Aceitas != 4 || res.Duplicadas != 1 || res.Rejeitadas != 5 {
		t.Fatalf("prévia do exemplo: %d/%d/%d, %v (%+v)", res.Aceitas, res.Duplicadas, res.Rejeitadas, err, res.Linhas)
	}
	codigos := map[int]string{}
	for _, l := range res.Linhas {
		codigos[l.Numero] = errs.Codigo(l.Erro) + "/" + l.Motivo
	}
	quer := map[int]string{6: "/", 7: "consentimento_invalido/", 8: "consentimento_invalido/", 9: "cliente_inativo/",
		10: "dados_invalidos/valor_compra", 11: "dados_invalidos/timestamp"}
	for n, q := range quer {
		if codigos[n] != q {
			t.Errorf("linha %d: obtido %q, esperado %q", n, codigos[n], q)
		}
	}
}
