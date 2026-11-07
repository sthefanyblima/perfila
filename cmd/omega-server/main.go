// Comando omega-server: o binário do UML8. Subcomandos:
//
//	omega-server [-config config.env] serve                  sobe a API HTTP (Spec 3.7)
//	omega-server [-config config.env] expirar-recomendacoes  expira recomendações vencidas (RN-08)
//
// Raiz de composição (C-08). Até a fase Generic, DATABASE_URL vazia = repositórios em memória
// com dados sintéticos de demonstração (D-23).
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"perfila/internal/api"
	"perfila/internal/application/usecase"
	"perfila/internal/domain/service"
	"perfila/internal/infra/config"
	"perfila/internal/infra/memoria"
	"perfila/internal/infra/relogio"
)

func main() {
	if err := executar(os.Args[1:], os.Getenv, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "omega-server:", err)
		os.Exit(1)
	}
}

const uso = "uso: omega-server [-config config.env] serve|expirar-recomendacoes"

func executar(args []string, getenv func(string) string, saida io.Writer) error {
	fs := flag.NewFlagSet("omega-server", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	caminho := fs.String("config", "", "arquivo KEY=VALUE (o ambiente prevalece)")
	if err := fs.Parse(args); err != nil || fs.NArg() != 1 {
		return errors.New(uso)
	}
	cfg, err := config.Carregar(*caminho, getenv)
	if err != nil {
		return err
	}
	app, err := montar(cfg, saida)
	if err != nil {
		return err
	}
	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer parar()

	switch fs.Arg(0) {
	case "serve":
		return servir(ctx, cfg, app)
	case "expirar-recomendacoes":
		n, err := app.expirar.Executar(ctx)
		if err != nil {
			return err
		}
		app.log.Info("recomendacoes_expiradas", "quantidade", n)
		return nil
	}
	return errors.New(uso)
}

// aplicacao é o grafo de dependências montado uma vez.
type aplicacao struct {
	log     *slog.Logger
	handler http.Handler
	expirar *usecase.ExpirarRecomendacoesUseCase
}

func montar(cfg config.Config, saida io.Writer) (aplicacao, error) {
	nivel := map[string]slog.Level{"debug": slog.LevelDebug, "info": slog.LevelInfo, "warn": slog.LevelWarn, "error": slog.LevelError}[cfg.LogLevel]
	log := slog.New(slog.NewJSONHandler(saida, &slog.HandlerOptions{Level: nivel}))

	if !cfg.ModoMemoria() {
		return aplicacao{}, errors.New("PostgreSQL chega na fase Generic (T-19); deixe DATABASE_URL vazia para usar o modo memória (D-23)")
	}
	clientes := memoria.NovoClienteRepository()
	eventos := memoria.NovoEventoRepository()
	perfis := memoria.NovoPerfilRepository()
	segmentos := memoria.NovoSegmentoRepository()
	recomendacoes := memoria.NovoRecomendacaoRepository()
	consentimentos := memoria.NovoConsentimentoRepository()
	produtos := memoria.NovoProdutoRepository()
	memoria.CarregarDemonstracao(segmentos, produtos)
	log.Warn("modo_memoria", "detalhe", "dados não persistem entre execuções; segmentos e catálogo sintéticos (D-23)")

	clock := relogio.Sistema{}
	pre := usecase.NovasPrecondicoes(clientes, service.NovoValidadorLGPD(consentimentos))
	coletar := usecase.NovoColetarDadosUseCase(pre, eventos, clock)
	deps := api.Dependencias{
		AuthToken:     cfg.AuthToken,
		Precondicoes:  pre,
		Cadastrar:     usecase.NovoCadastrarClienteUseCase(clientes, clock),
		Consentimento: usecase.NovoGerenciarConsentimentoUseCase(clientes, consentimentos, clock),
		Coletar:       coletar,
		GerarPerfil:   usecase.NovoGerarPerfilUseCase(pre, coletar, service.NovoPerfilamentoService(clock, cfg.EngajamentoMetaEventos), perfis),
		Segmentar:     usecase.NovoSegmentarClienteUseCase(pre, perfis, service.NovoSegmentacaoService(segmentos)),
		Recomendar: usecase.NovoGerarRecomendacaoUseCase(pre, perfis, segmentos,
			service.NovoRecomendacaoService(produtos, clock), recomendacoes, clock, cfg.RecomendacaoValidadeDias, log),
		Log: log,
	}
	return aplicacao{
		log:     log,
		handler: api.NovoRouter(deps),
		expirar: usecase.NovoExpirarRecomendacoesUseCase(recomendacoes, clock),
	}, nil
}

func servir(ctx context.Context, cfg config.Config, app aplicacao) error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:           app.handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	erro := make(chan error, 1)
	go func() { erro <- srv.ListenAndServe() }()
	app.log.Info("servidor_no_ar", "porta", cfg.HTTPPort)
	select {
	case err := <-erro:
		return err
	case <-ctx.Done():
	}
	app.log.Info("encerrando")
	desligar, cancelar := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelar()
	return srv.Shutdown(desligar)
}
