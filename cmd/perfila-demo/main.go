// Comando perfila-demo: executa, passo a passo, o fluxo do Diagrama de Sequência (UML4)
// com repositórios em memória e dados sintéticos. É a demonstração do core; o servidor
// HTTP (omega-server) chega na fase Support e chama exatamente os mesmos casos de uso.
//
// Raiz de composição (C-08): único lugar, junto de omega-server, que monta dependências.
package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
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
	if err := executar(os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
}

func passo(w io.Writer, n int, uc, titulo string) {
	fmt.Fprintf(w, "\n[%d] %s — %s\n", n, uc, titulo)
}

func linha(w io.Writer, formato string, args ...any) {
	fmt.Fprintf(w, "    "+formato+"\n", args...)
}

func executar(w io.Writer) error {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(io.Discard, nil)) // logs vão para o arquivo/coletor na fase Support
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
	cadastrar := usecase.NovoCadastrarClienteUseCase(clientes, clock)
	consentimento := usecase.NovoGerenciarConsentimentoUseCase(clientes, consentimentos, clock)
	coletar := usecase.NovoColetarDadosUseCase(pre, eventos, clock)
	gerarPerfil := usecase.NovoGerarPerfilUseCase(pre, coletar, service.NovoPerfilamentoService(clock, 10), perfis)
	segmentar := usecase.NovoSegmentarClienteUseCase(pre, perfis, service.NovoSegmentacaoService(segmentos))
	recomendar := usecase.NovoGerarRecomendacaoUseCase(pre, perfis, segmentos,
		service.NovoRecomendacaoService(produtos, clock), recomendacoes, clock, 30, log)

	fmt.Fprintln(w, "Perfila — demonstração do core (fluxo do Diagrama de Sequência, em memória)")
	fmt.Fprintln(w, strings.Repeat("─", 72))

	passo(w, 1, "UC-09", "Cadastrar cliente")
	cliente, err := cadastrar.Executar(ctx, "Cliente Sintético", "c-001@exemplo.test")
	if err != nil {
		return err
	}
	linha(w, "cliente %s… criado, ativo, sem consentimento", cliente.ID.String()[:8])

	passo(w, 2, "UC-02", "Tentar gerar perfil SEM consentimento (C-03)")
	_, err = gerarPerfil.Executar(ctx, cliente.ID)
	linha(w, "bloqueado → %s", errs.Codigo(err))

	passo(w, 3, "UC-08", "Conceder consentimento LGPD")
	if err := consentimento.Conceder(ctx, cliente.ID, entity.FinalidadeSegmentacao, "v1"); err != nil {
		return err
	}
	c, _ := consentimento.Consultar(ctx, cliente.ID, entity.FinalidadeSegmentacao)
	linha(w, "NAO_CONCEDIDO → %s (finalidade %q, termo %q)", c.Status, c.Finalidade, c.VersaoTermo)

	passo(w, 4, "UC-01", "Registrar eventos comportamentais")
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
		if _, err := coletar.Registrar(ctx, ev); err != nil {
			return err
		}
		linha(w, "%-9s valor %6.2f  categoria %q", e.tipo, e.valor, e.categoria)
	}
	_, err = gerarPerfil.Executar(ctx, uuid.New())
	linha(w, "(cliente inexistente → %s)", errs.Codigo(err))
	_, err = coletar.Registrar(ctx, entity.EventoComportamental{ClienteID: cliente.ID, Tipo: entity.TipoCompra,
		Origem: "ecommerce", Timestamp: agora, Valor: 0})
	linha(w, "(compra com valor 0 → %s, RN-04)", errs.Codigo(err))

	passo(w, 5, "UC-02", "Gerar perfil («include» Coletar)")
	perfil, err := gerarPerfil.Executar(ctx, cliente.ID)
	if err != nil {
		return err
	}
	linha(w, "frequência de compras  %d", perfil.FrequenciaCompras)
	linha(w, "ticket médio           %.2f", perfil.TicketMedio)
	linha(w, "categoria preferida    %s", perfil.CategoriaPreferida)
	linha(w, "score de engajamento   %.2f (5 eventos em 30 dias / meta 10)", perfil.ScoreEngajamento)

	passo(w, 6, "UC-03", "Segmentar cliente (primeiro critério que casa)")
	seg, err := segmentar.Executar(ctx, cliente.ID)
	if err != nil {
		return err
	}
	linha(w, "segmento %q ← critério: %s", seg.Nome, seg.Criterio)

	passo(w, 7, "UC-04", "Recomendar produto («extend» Segmentar)")
	rec, err := recomendar.Executar(ctx, cliente.ID)
	if err != nil {
		return err
	}
	linha(w, "status %s · motivo %q", rec.Status, rec.Motivo)
	linha(w, "válida até %s (30 dias)", rec.Validade.Format("02/01/2006"))

	passo(w, 8, "UC-08", "Revogar consentimento (RN-03: efeito imediato)")
	if err := consentimento.Revogar(ctx, cliente.ID, entity.FinalidadeSegmentacao); err != nil {
		return err
	}
	_, err = recomendar.Executar(ctx, cliente.ID)
	linha(w, "nova recomendação → %s", errs.Codigo(err))

	fmt.Fprintln(w, "\n"+strings.Repeat("─", 72))
	fmt.Fprintln(w, "Mesmos casos de uso que a API chamará na fase Support; só a infraestrutura muda.")
	return nil
}
