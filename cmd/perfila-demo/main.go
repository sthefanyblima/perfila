// Comando perfila-demo: executa o fluxo do Diagrama de Sequência (UC-08 → UC-01 → UC-02 →
// UC-03 → UC-04) com repositórios em memória e dados sintéticos. Serve de demonstração do
// core na apresentação parcial; o servidor HTTP (omega-server) chega na fase Support.
//
// Raiz de composição (C-08): é o único lugar, junto de omega-server, que monta dependências.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"

	"perfila/internal/application/usecase"
	"perfila/internal/domain/entity"
	"perfila/internal/domain/service"
	"perfila/internal/infra/memoria"
	"perfila/internal/infra/relogio"
	"perfila/internal/shared/errs"
)

func main() {
	if err := executar(); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
}

func executar() error {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	clock := relogio.Sistema{}
	agora := clock.Agora()

	// Infraestrutura em memória (substituída por PostgreSQL na fase Generic).
	clientes := memoria.NovoClienteRepository()
	eventos := memoria.NovoEventoRepository()
	perfis := memoria.NovoPerfilRepository()
	segmentos := memoria.NovoSegmentoRepository()
	recomendacoes := memoria.NovoRecomendacaoRepository()
	consentimentos := memoria.NovoConsentimentoRepository()
	produtos := memoria.NovoProdutoRepository()

	// Segmentos (equivalente à migração inicial) e catálogo sintético.
	for _, s := range []entity.Segmento{
		{ID: uuid.New(), Nome: "premium", Criterio: "ticketMedio >= 500 && frequenciaCompras >= 3", Prioridade: 1, Ativo: true},
		{ID: uuid.New(), Nome: "engajado", Criterio: "scoreEngajamento >= 0.7", Prioridade: 2, Ativo: true},
		{ID: uuid.New(), Nome: "padrao", Criterio: "", Prioridade: 99, Ativo: true},
	} {
		segmentos.Adicionar(s)
	}
	for _, p := range []entity.Produto{
		{ID: uuid.New(), Nome: "Livro A", Categoria: "livros", Preco: 80, Ativo: true},
		{ID: uuid.New(), Nome: "Livro B", Categoria: "livros", Preco: 50, Ativo: true},
		{ID: uuid.New(), Nome: "Jogo X", Categoria: "games", Preco: 200, Ativo: true},
	} {
		produtos.Adicionar(p)
	}

	// Composição dos casos de uso.
	pre := usecase.NovasPrecondicoes(clientes, service.NovoValidadorLGPD(consentimentos))
	consentimento := usecase.NovoGerenciarConsentimentoUseCase(clientes, consentimentos, clock)
	coletar := usecase.NovoColetarDadosUseCase(pre, eventos, clock)
	gerarPerfil := usecase.NovoGerarPerfilUseCase(pre, coletar, service.NovoPerfilamentoService(clock, 10), perfis)
	segmentar := usecase.NovoSegmentarClienteUseCase(pre, perfis, service.NovoSegmentacaoService(segmentos))
	recomendar := usecase.NovoGerarRecomendacaoUseCase(pre, perfis, segmentos,
		service.NovoRecomendacaoService(produtos, clock), recomendacoes, clock, 30, log)

	cliente := entity.Cliente{ID: uuid.New(), Nome: "Cliente Sintético", Email: "c-001@exemplo.test", DataCadastro: agora, Ativo: true}
	if err := clientes.Salvar(ctx, cliente); err != nil {
		return err
	}

	// 0) Sem consentimento, nada é processado (C-03).
	_, err := gerarPerfil.Executar(ctx, cliente.ID)
	fmt.Printf("Sem consentimento → %s\n", errs.Codigo(err))

	// UC-08: conceder consentimento.
	if err := consentimento.Conceder(ctx, cliente.ID, "segmentacao", "v1"); err != nil {
		return err
	}

	// UC-01: registrar eventos sintéticos.
	for i, e := range []struct {
		tipo      entity.TipoEvento
		valor     float64
		categoria string
	}{
		{entity.TipoCompra, 520, "livros"}, {entity.TipoCompra, 610, "livros"}, {entity.TipoCompra, 480, "games"},
		{entity.TipoNavegacao, 0, "livros"}, {entity.TipoInteracao, 0, ""},
	} {
		ev := entity.EventoComportamental{ClienteID: cliente.ID, Tipo: e.tipo, Origem: "ecommerce",
			Timestamp: agora.Add(-time.Duration(i+1) * 24 * time.Hour), Valor: e.valor, Categoria: e.categoria}
		if err := coletar.Registrar(ctx, ev); err != nil {
			return err
		}
	}

	// Fluxo RN-09: UC-02 → UC-03 → UC-04.
	perfil, err := gerarPerfil.Executar(ctx, cliente.ID)
	if err != nil {
		return err
	}
	seg, err := segmentar.Executar(ctx, cliente.ID)
	if err != nil {
		return err
	}
	rec, err := recomendar.Executar(ctx, cliente.ID)
	if err != nil {
		return err
	}

	saida := map[string]map[string]string{
		"perfil": {
			"frequenciaCompras":  fmt.Sprint(perfil.FrequenciaCompras),
			"ticketMedio":        fmt.Sprintf("%.2f", perfil.TicketMedio),
			"categoriaPreferida": perfil.CategoriaPreferida,
			"scoreEngajamento":   fmt.Sprintf("%.2f", perfil.ScoreEngajamento),
		},
		"segmento":     {"nome": seg.Nome, "criterio": seg.Criterio},
		"recomendacao": {"status": string(rec.Status), "motivo": rec.Motivo, "validade": rec.Validade.Format(time.RFC3339)},
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(saida); err != nil {
		return err
	}

	// RN-03: revogação tem efeito imediato.
	if err := consentimento.Revogar(ctx, cliente.ID); err != nil {
		return err
	}
	_, err = recomendar.Executar(ctx, cliente.ID)
	fmt.Printf("Após revogar → %s\n", errs.Codigo(err))
	return nil
}
