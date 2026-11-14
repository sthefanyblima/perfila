// Comando omega-server: o binário do UML8. Subcomandos:
//
//	omega-server [-config config.env] serve                  sobe a API HTTP (Spec 3.7)
//	omega-server [-config config.env] expirar-recomendacoes  expira recomendações vencidas (RN-08)
//	omega-server hash-senha                                  lê a senha do operador e imprime o bcrypt (Q-28)
//	omega-server [-config config.env] carregar-exemplos      cria os clientes de exemplo se não houver nenhum (Q-30)
//
// Raiz de composição (C-08). DATABASE_URL preenchida = PostgreSQL, com as migrações aplicadas
// ao subir (T-19, Q-07, Q-09); vazia = repositórios em memória com dados sintéticos (D-23),
// só para desenvolvimento, BDD e demo local.
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/bcrypt"

	"perfila/internal/api"
	"perfila/internal/application/usecase"
	"perfila/internal/domain/port"
	"perfila/internal/domain/service"
	"perfila/internal/infra/config"
	"perfila/internal/infra/external"
	"perfila/internal/infra/memoria"
	"perfila/internal/infra/relogio"
	"perfila/internal/infra/repository/pg"
	"perfila/internal/web"
)

func main() {
	if err := executar(os.Args[1:], os.Getenv, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "omega-server:", err)
		os.Exit(1)
	}
}

const uso = "uso: omega-server [-config config.env] serve|expirar-recomendacoes|aplicar-retencao|carregar-exemplos|hash-senha"

func executar(args []string, getenv func(string) string, saida io.Writer) error {
	fs := flag.NewFlagSet("omega-server", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	caminho := fs.String("config", "", "arquivo KEY=VALUE (o ambiente prevalece)")
	if err := fs.Parse(args); err != nil || fs.NArg() != 1 {
		return errors.New(uso)
	}
	if fs.Arg(0) == "hash-senha" { // não precisa de configuração (Q-28)
		return hashSenha(os.Stdin, saida)
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
		if cfg.ModoMemoria() { // no modo memória, os exemplos sempre entram (Q-30)
			if err := carregarExemplos(ctx, app); err != nil {
				return err
			}
		}
		return servir(ctx, cfg, app)
	case "carregar-exemplos":
		return carregarExemplos(ctx, app)
	case "expirar-recomendacoes":
		n, err := app.expirar.Executar(ctx)
		if err != nil {
			return err
		}
		app.log.Info("recomendacoes_expiradas", "quantidade", n)
		return nil
	case "aplicar-retencao": // RN-18, Q-51: só a quantidade vai para o log (C-09)
		n, err := app.retencao.Executar(ctx)
		if err != nil {
			return err
		}
		app.log.Info("retencao_aplicada", "quantidade", n, "prazo_dias", app.retencao.PrazoDias())
		app.log.Info("direito_atendido", "pedido", "retencao", "origem", "sistema", "quantidade", n)
		return nil
	}
	return errors.New(uso)
}

// hashSenha lê a senha da primeira linha da entrada e imprime o hash bcrypt para
// PAINEL_SENHA_HASH (Q-28). A senha nunca é impressa nem registrada.
func hashSenha(entrada io.Reader, saida io.Writer) error {
	linha, err := bufio.NewReader(entrada).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	senha := strings.TrimRight(linha, "\r\n")
	if len(senha) < 8 {
		return errors.New("hash-senha: a senha precisa de pelo menos 8 caracteres")
	}
	h, err := bcrypt.GenerateFromPassword([]byte(senha), 12)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(saida, string(h))
	return err
}

func carregarExemplos(ctx context.Context, app aplicacao) error {
	n, err := app.exemplos.carregar(ctx)
	if err != nil {
		return err
	}
	app.log.Info("exemplos_carregados", "quantidade", n)
	return nil
}

// aplicacao é o grafo de dependências montado uma vez.
type aplicacao struct {
	log      *slog.Logger
	handler  http.Handler
	expirar  *usecase.ExpirarRecomendacoesUseCase
	retencao *usecase.AplicarRetencaoUseCase
	exemplos exemplos
	fechar   func()
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
	tokens         port.TokenAcessoRepository      // portal do titular (v1.10)
	pedidos        port.PedidoEliminacaoRepository // portal do titular (v1.10)
	caixaSaida     port.CaixaSaidaSimulada         // NotificadorSimulado (Q-57)
	saude          func(ctx context.Context) error // nil no modo memória (Q-19)
	migracoes      string                          // versão do esquema, para a tela de Operação
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
	processar := usecase.NovoProcessarClienteUseCase(coletar,
		usecase.NovoGerarPerfilUseCase(pre, coletar, service.NovoPerfilamentoService(clock, cfg.EngajamentoMetaEventos), repos.perfis),
		usecase.NovoSegmentarClienteUseCase(pre, repos.perfis, service.NovoSegmentacaoService(repos.segmentos)),
		usecase.NovoGerarRecomendacaoUseCase(pre, repos.perfis, repos.segmentos,
			service.NovoRecomendacaoService(repos.produtos, clock), repos.recomendacoes, clock, cfg.RecomendacaoValidadeDias, log))
	cadastrar := usecase.NovoCadastrarClienteUseCase(repos.clientes, clock)
	consentimento := usecase.NovoGerenciarConsentimentoUseCase(repos.clientes, repos.consentimentos, clock)
	expirar := usecase.NovoExpirarRecomendacoesUseCase(repos.recomendacoes, clock)
	retencao := usecase.NovoAplicarRetencaoUseCase(repos.consentimentos, repos.eventos, repos.perfis, repos.recomendacoes,
		clock, cfg.RetencaoDiasRevogado)
	deps := api.Dependencias{
		AuthToken:     cfg.AuthToken,
		Precondicoes:  pre,
		Cadastrar:     cadastrar,
		Consentimento: consentimento,
		Coletar:       coletar,
		Processar:     processar,
		Log:           log,
		Clock:         clock,
		Saude:         repos.saude,
	}
	mux := http.NewServeMux()
	mux.Handle("/", api.NovoRouter(deps))
	exportar := usecase.NovoExportarDadosTitularUseCase(repos.clientes, consentimento, repos.eventos, repos.perfis,
		repos.segmentos, repos.recomendacoes, repos.produtos, clock)
	meusDados := usecase.NovoConsultarMeusDadosUseCase(repos.clientes, consentimento, repos.pedidos)
	var caixaSaida *usecase.ListarCaixaSaidaUseCase
	if cfg.PortalLigado() { // portal do titular (v1.10, Q-55 a Q-60)
		portal, err := web.NovoPortal(web.DependenciasPortal{
			ChaveSessao: cfg.PortalChaveSessao, Finalidade: cfg.PainelFinalidade, VersaoTermo: cfg.PainelVersaoTermo,
			Solicitar: usecase.NovoSolicitarAcessoTitularUseCase(repos.clientes, repos.tokens,
				external.NovoNotificadorSimulado(repos.caixaSaida, clock), clock, cfg.PortalURLBase),
			Autenticar:    usecase.NovoAutenticarTitularUseCase(repos.clientes, repos.tokens, clock),
			MeusDados:     meusDados,
			Consentimento: consentimento,
			ExportarDados: exportar,
			Eliminacao:    usecase.NovoPedirEliminacaoUseCase(repos.clientes, consentimento, repos.pedidos, clock),
			Clock:         clock, Log: log,
		})
		if err != nil {
			repos.fechar()
			return aplicacao{}, err
		}
		mux.Handle("/titular/", portal)
		mux.Handle("/titular", http.RedirectHandler("/titular/entrar", http.StatusSeeOther))
		caixaSaida = usecase.NovoListarCaixaSaidaUseCase(repos.caixaSaida, clock)
	} else {
		log.Warn("portal_desligado", "detalhe", "defina PORTAL_URL_BASE e PORTAL_CHAVE_SESSAO para ligar o portal do titular (Q-60)")
	}
	if cfg.PainelLigado() {
		painel, err := web.NovoRouter(web.Dependencias{
			Usuario: cfg.PainelUsuario, SenhaHash: []byte(cfg.PainelSenhaHash), ChaveSessao: cfg.PainelChaveSessao,
			Finalidade: cfg.PainelFinalidade, VersaoTermo: cfg.PainelVersaoTermo,
			Cadastrar: cadastrar, Consentimento: consentimento, Coletar: coletar, Processar: processar,
			ListarClientes:     usecase.NovoListarClientesUseCase(repos.clientes, repos.consentimentos, repos.perfis, repos.segmentos),
			ConsultarCliente:   usecase.NovoConsultarClienteUseCase(repos.clientes, consentimento),
			ConsultarResultado: usecase.NovoConsultarResultadoUseCase(pre, repos.eventos, repos.perfis, repos.segmentos, repos.recomendacoes, repos.produtos),
			ListarSegmentos:    usecase.NovoListarSegmentosUseCase(repos.segmentos),
			ListarCatalogo:     usecase.NovoListarCatalogoUseCase(repos.produtos),
			Expirar:            expirar,
			AtualizarCadastro:  usecase.NovoAtualizarCadastroUseCase(repos.clientes),
			ExportarDados:      exportar,
			EliminarDados: usecase.NovoEliminarDadosTitularUseCase(repos.clientes, repos.consentimentos, repos.eventos,
				repos.perfis, repos.recomendacoes, repos.tokens, repos.pedidos, clock),
			AplicarRetencao: retencao,
			MeusDados:       meusDados,
			Pedidos:         usecase.NovoListarPedidosEliminacaoUseCase(repos.pedidos, repos.clientes, consentimento),
			CaixaSaida:      caixaSaida,
			MetaEventos:     cfg.EngajamentoMetaEventos,
			Clock:           clock, Log: log, Saude: repos.saude, Migracoes: repos.migracoes,
		})
		if err != nil {
			repos.fechar()
			return aplicacao{}, err
		}
		mux.Handle("/painel/", painel)
		mux.Handle("/painel", http.RedirectHandler("/painel/clientes", http.StatusSeeOther))
	} else {
		log.Warn("painel_desligado", "detalhe", "defina PAINEL_SENHA_HASH e PAINEL_CHAVE_SESSAO para ligar o painel (Q-29)")
	}
	return aplicacao{
		log:      log,
		handler:  mux,
		expirar:  expirar,
		retencao: retencao,
		exemplos: exemplos{repos: repos, clock: clock, metaEventos: cfg.EngajamentoMetaEventos, validadeDias: cfg.RecomendacaoValidadeDias},
		fechar:   repos.fechar,
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
		tokens:         memoria.NovoTokenAcessoRepository(),
		pedidos:        memoria.NovoPedidoEliminacaoRepository(),
		caixaSaida:     memoria.NovaCaixaSaidaSimulada(),
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
		tokens:         pg.NovoPgTokenAcessoRepository(db),
		pedidos:        pg.NovoPgPedidoEliminacaoRepository(db),
		caixaSaida:     pg.NovaPgCaixaSaidaSimulada(db),
		saude:          db.PingContext,
		migracoes:      pg.UltimaMigracao(),
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
