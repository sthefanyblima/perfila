// Comando omega-server: o binário do UML8. Subcomandos:
//
//	omega-server [-config config.env] serve                  sobe a API HTTP (Spec 3.7)
//	omega-server [-config config.env] expirar-recomendacoes  expira recomendações vencidas (RN-08)
//
// Raiz de composição (C-08). DATABASE_URL preenchida = PostgreSQL, com as migrações aplicadas
// ao subir (T-19, Q-07, Q-09); vazia = repositórios em memória com dados sintéticos (D-23),
// só para desenvolvimento, BDD e demo local.
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
	"perfila/internal/domain/port"
	"perfila/internal/domain/service"
	"perfila/internal/infra/config"
	"perfila/internal/infra/memoria"
	"perfila/internal/infra/relogio"
	"perfila/internal/infra/repository/pg"
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
	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer parar()
	app, err := montar(ctx, cfg, saida)
	if err != nil {
		return err
	}
	defer app.fechar()

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
	fechar  func()
}

// repositorios agrupa as implementações dos ports de persistência de um backend.
type repositorios struct {
	clientes       port.ClienteRepository
	eventos        port.EventoRepository
	perfis         port.PerfilRepository
	segmentos      port.SegmentoRepository
	recomendacoes  port.RecomendacaoRepository
	consentimentos port.ConsentimentoRepository
	produtos       port.ProdutoRepository
	saude          func(ctx context.Context) error // nil no modo memória (Q-19)
	fechar         func()
}

func montar(ctx context.Context, cfg config.Config, saida io.Writer) (aplicacao, error) {
	nivel := map[string]slog.Level{"debug": slog.LevelDebug, "info": slog.LevelInfo, "warn": slog.LevelWarn, "error": slog.LevelError}[cfg.LogLevel]
	log := slog.New(slog.NewJSONHandler(saida, &slog.HandlerOptions{Level: nivel}))

	var repos repositorios
	var err error
	if cfg.ModoMemoria() {
		repos = repositoriosEmMemoria()
		log.Warn("modo_memoria", "detalhe", "dados não persistem entre execuções; segmentos e catálogo sintéticos (D-23)")
	} else if repos, err = repositoriosPostgres(ctx, cfg.DatabaseURL, log); err != nil {
		return aplicacao{}, err
	}

	clock := relogio.Sistema{}
	pre := usecase.NovasPrecondicoes(repos.clientes, service.NovoValidadorLGPD(repos.consentimentos))
	coletar := usecase.NovoColetarDadosUseCase(pre, repos.eventos, clock)
	deps := api.Dependencias{
		AuthToken:     cfg.AuthToken,
		Precondicoes:  pre,
		Cadastrar:     usecase.NovoCadastrarClienteUseCase(repos.clientes, clock),
		Consentimento: usecase.NovoGerenciarConsentimentoUseCase(repos.clientes, repos.consentimentos, clock),
		Coletar:       coletar,
		GerarPerfil:   usecase.NovoGerarPerfilUseCase(pre, coletar, service.NovoPerfilamentoService(clock, cfg.EngajamentoMetaEventos), repos.perfis),
		Segmentar:     usecase.NovoSegmentarClienteUseCase(pre, repos.perfis, service.NovoSegmentacaoService(repos.segmentos)),
		Recomendar: usecase.NovoGerarRecomendacaoUseCase(pre, repos.perfis, repos.segmentos,
			service.NovoRecomendacaoService(repos.produtos, clock), repos.recomendacoes, clock, cfg.RecomendacaoValidadeDias, log),
		Log:   log,
		Clock: clock,
		Saude: repos.saude,
	}
	return aplicacao{
		log:     log,
		handler: api.NovoRouter(deps),
		expirar: usecase.NovoExpirarRecomendacoesUseCase(repos.recomendacoes, clock),
		fechar:  repos.fechar,
	}, nil
}

// repositoriosEmMemoria monta os fakes com segmentos e catálogo sintéticos (D-23).
func repositoriosEmMemoria() repositorios {
	segmentos := memoria.NovoSegmentoRepository()
	produtos := memoria.NovoProdutoRepository()
	memoria.CarregarDemonstracao(segmentos, produtos)
	return repositorios{
		clientes:       memoria.NovoClienteRepository(),
		eventos:        memoria.NovoEventoRepository(),
		perfis:         memoria.NovoPerfilRepository(),
		segmentos:      segmentos,
		recomendacoes:  memoria.NovoRecomendacaoRepository(),
		consentimentos: memoria.NovoConsentimentoRepository(),
		produtos:       produtos,
		fechar:         func() {},
	}
}

// repositoriosPostgres conecta ao omega_db, aplica as migrações pendentes (Q-09) e monta os
// Pg*Repository (T-19). A URL nunca vai para o log: pode conter senha.
func repositoriosPostgres(ctx context.Context, url string, log *slog.Logger) (repositorios, error) {
	db, err := pg.Abrir(ctx, url)
	if err != nil {
		return repositorios{}, err
	}
	aplicadas, err := pg.Migrar(ctx, db)
	if err != nil {
		db.Close()
		return repositorios{}, err
	}
	log.Info("postgres_conectado", "migracoes_aplicadas", aplicadas)
	return repositorios{
		clientes:       pg.NovoPgClienteRepository(db),
		eventos:        pg.NovoPgEventoRepository(db),
		perfis:         pg.NovoPgPerfilRepository(db),
		segmentos:      pg.NovoPgSegmentoRepository(db),
		recomendacoes:  pg.NovoPgRecomendacaoRepository(db),
		consentimentos: pg.NovoPgConsentimentoRepository(db),
		produtos:       pg.NovoPgProdutoRepository(db),
		saude:          db.PingContext,
		fechar:         func() { db.Close() },
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
