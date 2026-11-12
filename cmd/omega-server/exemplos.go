package main

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"perfila/internal/application/usecase"
	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/domain/service"
	"perfila/internal/infra/relogio"
)

// Clientes de exemplo com história (Q-30, Q-32): sintéticos (B-08), criados pelos próprios casos
// de uso com um relógio de carga que anda pelo passado. Assim cadastro, consentimento, revogação,
// eventos e recomendações ficam com datas diferentes, e perfil, segmento e recomendação são os que
// as regras calculam. Só carregam com o banco sem nenhum cliente (idempotente).

type exemplos struct {
	repos        repositorios
	clock        port.Clock // relógio real: "agora" da carga
	metaEventos  int
	validadeDias int
}

// passo é uma ação na linha do tempo de um cliente, "dias" antes de agora.
type passo struct {
	dias      float64
	acao      string // cadastrar, conceder, compra, navegar, processar, revogar, expirar, desativar
	categoria string
	valor     float64
	termo     string
	n         int // navegar: quantos eventos (espalhados nas horas seguintes)
}

type exemplo struct {
	nome, email string
	passos      []passo
}

func cad(d float64) passo                { return passo{dias: d, acao: "cadastrar"} }
func conc(d float64, termo string) passo { return passo{dias: d, acao: "conceder", termo: termo} }
func comp(d float64, c string, v float64) passo {
	return passo{dias: d, acao: "compra", categoria: c, valor: v}
}
func nav(d float64, n int) passo { return passo{dias: d, acao: "navegar", n: n} }
func proc(d float64) passo       { return passo{dias: d, acao: "processar"} }
func rev(d float64) passo        { return passo{dias: d, acao: "revogar"} }
func exp(d float64) passo        { return passo{dias: d, acao: "expirar"} }
func desat(d float64) passo      { return passo{dias: d, acao: "desativar"} }

func (e exemplos) lista() []exemplo {
	meta := float64(e.metaEventos)
	engajado := int(math.Ceil(0.8 * meta))   // 80% da meta em 30 dias: engajado
	explorador := int(math.Ceil(0.5 * meta)) // 50% da meta, sem compras: explorador
	return []exemplo{
		{"Ana Sintética", "c-001@exemplo.test", []passo{ // premium, com histórico de recomendações
			cad(150), conc(148, "v1"), comp(140, "livros", 520), nav(139, 2), proc(139),
			comp(100, "livros", 610), comp(60, "livros", 700), proc(59), comp(20, "livros", 640), comp(5, "livros", 550), nav(4, 3), proc(4)}},
		{"Bruno Sintético", "c-002@exemplo.test", []passo{ // engajado
			cad(90), conc(89, "v1"), comp(80, "games", 200), proc(79), nav(25, engajado), proc(2)}},
		{"Carla Sintética", "c-003@exemplo.test", []passo{ // ocasional
			cad(45), conc(44, "v1"), comp(30, "eletronicos", 350), nav(29, 2), proc(1)}},
		{"Davi Sintético", "c-004@exemplo.test", []passo{ // ocasional, sem produto elegível (moveis)
			cad(40), conc(39, "v1"), comp(12, "moveis", 120), proc(11)}},
		{"Elisa Sintética", "c-005@exemplo.test", []passo{cad(12)}}, // nunca concedeu
		{"Fábio Sintético", "c-006@exemplo.test", []passo{ // revogado: dados guardados e bloqueados
			cad(200), conc(198, "v1"), comp(190, "livros", 90), comp(150, "livros", 60), proc(149), exp(110), rev(45)}},
		{"Gabi Sintética", "c-007@exemplo.test", []passo{ // inativa com consentimento: histórico só leitura
			cad(300), conc(299, "v1"), comp(280, "games", 150), comp(250, "games", 320), proc(249), exp(200), desat(180)}},
		{"Heitor Sintético", "c-008@exemplo.test", []passo{ // vip
			cad(130), conc(129, "v1"), comp(120, "eletronicos", 1200), comp(90, "eletronicos", 1500), proc(89),
			comp(60, "eletronicos", 980), comp(25, "eletronicos", 1100), comp(8, "eletronicos", 1300), proc(7)}},
		{"Isabela Sintética", "c-009@exemplo.test", []passo{ // fiel
			cad(110), conc(109, "v1"), comp(100, "moda", 90), comp(80, "moda", 140), comp(55, "moda", 210),
			proc(54), comp(35, "moda", 120), comp(15, "moda", 160), comp(3, "moda", 75), proc(2)}},
		{"João Sintético", "c-010@exemplo.test", []passo{ // explorador: navega, não compra
			cad(20), conc(19, "v1"), nav(18, explorador), proc(1)}},
		{"Karina Sintética", "c-011@exemplo.test", []passo{ // revogou e concedeu de novo, termo v2
			cad(160), conc(158, "v1"), comp(150, "beleza", 180), comp(130, "beleza", 240), proc(120), rev(90),
			conc(30, "v2"), comp(20, "beleza", 150), nav(19, 2), proc(3)}},
		{"Lucas Sintético", "c-012@exemplo.test", []passo{ // concedeu e tem eventos, ainda não processado
			cad(6), conc(5, "v2"), nav(4, 2), comp(3, "esportes", 70)}},
		{"Marina Sintética", "c-013@exemplo.test", []passo{ // só recomendação expirada
			cad(80), conc(79, "v1"), comp(70, "casa", 290), proc(69), exp(35)}},
		{"Nina Sintética", "c-014@exemplo.test", []passo{ // padrao
			cad(60), conc(59, "v1"), comp(50, "esportes", 70), comp(35, "esportes", 120), comp(10, "esportes", 90), proc(9)}},
	}
}

// casos monta os casos de uso sobre os mesmos repositórios, com o relógio de carga.
func (e exemplos) casos(c port.Clock) (cad *usecase.CadastrarClienteUseCase, cons *usecase.GerenciarConsentimentoUseCase,
	col *usecase.ColetarDadosUseCase, proc *usecase.ProcessarClienteUseCase, exp *usecase.ExpirarRecomendacoesUseCase) {
	r := e.repos
	pre := usecase.NovasPrecondicoes(r.clientes, service.NovoValidadorLGPD(r.consentimentos))
	col = usecase.NovoColetarDadosUseCase(pre, r.eventos, c)
	proc = usecase.NovoProcessarClienteUseCase(col,
		usecase.NovoGerarPerfilUseCase(pre, col, service.NovoPerfilamentoService(c, e.metaEventos), r.perfis),
		usecase.NovoSegmentarClienteUseCase(pre, r.perfis, service.NovoSegmentacaoService(r.segmentos)),
		usecase.NovoGerarRecomendacaoUseCase(pre, r.perfis, r.segmentos, service.NovoRecomendacaoService(r.produtos, c),
			r.recomendacoes, c, e.validadeDias, nil))
	return usecase.NovoCadastrarClienteUseCase(r.clientes, c), usecase.NovoGerenciarConsentimentoUseCase(r.clientes, r.consentimentos, c),
		col, proc, usecase.NovoExpirarRecomendacoesUseCase(r.recomendacoes, c)
}

// carregar devolve quantos clientes criou; 0 se já havia algum cliente.
func (e exemplos) carregar(ctx context.Context) (int, error) {
	_, total, err := e.repos.clientes.Listar(ctx, port.FiltroClientes{Limite: 1})
	if err != nil || total > 0 {
		return 0, err
	}
	agora := e.clock.Agora()
	carga := relogio.NovoFixo(agora)
	cadastrar, consentimento, coletar, processar, expirar := e.casos(carga)
	n := 0
	for _, ex := range e.lista() {
		var id uuid.UUID
		for _, p := range ex.passos {
			carga.Definir(agora.Add(-time.Duration(p.dias * 24 * float64(time.Hour))))
			if err := e.executar(ctx, p, ex, &id, carga, cadastrar, consentimento, coletar, processar, expirar); err != nil {
				return n, fmt.Errorf("exemplo %s, passo %s: %w", ex.email, p.acao, err)
			}
		}
		n++
	}
	return n, nil
}

func (e exemplos) executar(ctx context.Context, p passo, ex exemplo, id *uuid.UUID, carga *relogio.Fixo,
	cadastrar *usecase.CadastrarClienteUseCase, consentimento *usecase.GerenciarConsentimentoUseCase,
	coletar *usecase.ColetarDadosUseCase, processar *usecase.ProcessarClienteUseCase, expirar *usecase.ExpirarRecomendacoesUseCase) error {
	evento := func(tipo entity.TipoEvento, valor float64, categoria string, quando time.Time) error {
		_, err := coletar.Registrar(ctx, entity.EventoComportamental{ID: uuid.New(), ClienteID: *id, Tipo: tipo,
			Origem: origem(quando), Valor: valor, Categoria: categoria, Timestamp: quando})
		return err
	}
	switch p.acao {
	case "cadastrar":
		c, err := cadastrar.Executar(ctx, ex.nome, ex.email)
		*id = c.ID
		return err
	case "conceder":
		return consentimento.Conceder(ctx, *id, "segmentacao", p.termo)
	case "compra":
		return evento(entity.TipoCompra, p.valor, p.categoria, carga.Agora())
	case "navegar":
		base := carga.Agora()
		for i := 0; i < p.n; i++ {
			tipo := entity.TipoNavegacao
			if i%3 == 2 {
				tipo = entity.TipoInteracao
			}
			quando := base.Add(time.Duration(i*7) * time.Hour)
			if quando.After(e.clock.Agora()) {
				quando = e.clock.Agora()
			}
			carga.Definir(quando)
			if err := evento(tipo, 0, "", quando); err != nil {
				return err
			}
		}
		return nil
	case "processar":
		_, err := processar.Executar(ctx, *id)
		return err
	case "revogar":
		return consentimento.Revogar(ctx, *id)
	case "expirar":
		_, err := expirar.Executar(ctx)
		return err
	case "desativar":
		c, err := e.repos.clientes.BuscarPorID(ctx, *id)
		if err != nil {
			return err
		}
		c.Desativar()
		return e.repos.clientes.Salvar(ctx, c)
	}
	return fmt.Errorf("ação desconhecida %q", p.acao)
}

// origem varia o canal do evento para a linha do tempo não ficar repetitiva.
func origem(t time.Time) string {
	return []string{"loja on-line", "aplicativo", "loja física", "central de atendimento"}[t.YearDay()%4]
}
