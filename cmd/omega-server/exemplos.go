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
)

// Clientes de exemplo (Q-30): um por situação que o painel mostra, todos sintéticos (B-08).
// Entram pelos próprios casos de uso, então perfil, segmento e recomendação são os que o sistema
// calcula de verdade. Só carregam com o banco sem nenhum cliente (idempotente).

type exemplos struct {
	clientes      port.ClienteRepository
	clock         port.Clock
	metaEventos   int
	cadastrar     *usecase.CadastrarClienteUseCase
	consentimento *usecase.GerenciarConsentimentoUseCase
	coletar       *usecase.ColetarDadosUseCase
	processar     *usecase.ProcessarClienteUseCase
}

type compra struct {
	categoria string
	valor     float64
}

type exemplo struct {
	nome, email string
	conceder    bool
	compras     []compra
	navegacoes  int // eventos de navegação e interação, para o score de engajamento
	processar   bool
	revogar     bool
	desativar   bool
}

func (e exemplos) lista() []exemplo {
	// Engajado: 80% da meta de eventos em 30 dias (RN-05), contando a compra.
	engajamento := int(math.Ceil(0.8*float64(e.metaEventos))) - 1
	return []exemplo{
		{nome: "Ana Sintética", email: "c-001@exemplo.test", conceder: true, processar: true,
			compras: []compra{{"livros", 520}, {"livros", 640}, {"livros", 710}}}, // premium
		{nome: "Bruno Sintético", email: "c-002@exemplo.test", conceder: true, processar: true,
			compras: []compra{{"games", 200}}, navegacoes: engajamento}, // engajado
		{nome: "Carla Sintética", email: "c-003@exemplo.test", conceder: true, processar: true,
			compras: []compra{{"eletronicos", 350}}}, // padrao
		{nome: "Davi Sintético", email: "c-004@exemplo.test", conceder: true, processar: true,
			compras: []compra{{"moveis", 120}}}, // sem produto elegível (RN-07)
		{nome: "Elisa Sintética", email: "c-005@exemplo.test"}, // consentimento não concedido
		{nome: "Fábio Sintético", email: "c-006@exemplo.test", conceder: true, processar: true, revogar: true,
			compras: []compra{{"livros", 90}, {"livros", 60}}}, // consentimento revogado
		{nome: "Gabi Sintética", email: "c-007@exemplo.test", conceder: true, processar: true, desativar: true,
			compras: []compra{{"games", 150}}}, // cadastro inativo
	}
}

// carregar devolve quantos clientes criou; 0 se já havia algum cliente.
func (e exemplos) carregar(ctx context.Context) (int, error) {
	_, total, err := e.clientes.Listar(ctx, port.FiltroClientes{Limite: 1})
	if err != nil || total > 0 {
		return 0, err
	}
	n := 0
	for _, ex := range e.lista() {
		if err := e.criar(ctx, ex); err != nil {
			return n, fmt.Errorf("exemplo %s: %w", ex.email, err)
		}
		n++
	}
	return n, nil
}

func (e exemplos) criar(ctx context.Context, ex exemplo) error {
	c, err := e.cadastrar.Executar(ctx, ex.nome, ex.email)
	if err != nil || !ex.conceder {
		return err
	}
	if err := e.consentimento.Conceder(ctx, c.ID, "segmentacao", "v1"); err != nil {
		return err
	}
	agora := e.clock.Agora()
	registrar := func(tipo entity.TipoEvento, valor float64, categoria string, horas int) error {
		_, err := e.coletar.Registrar(ctx, entity.EventoComportamental{ID: uuid.New(), ClienteID: c.ID, Tipo: tipo,
			Origem: "loja on-line", Valor: valor, Categoria: categoria, Timestamp: agora.Add(-time.Duration(horas) * time.Hour)})
		return err
	}
	for i, cp := range ex.compras {
		if err := registrar(entity.TipoCompra, cp.valor, cp.categoria, 24*(i+1)); err != nil {
			return err
		}
	}
	for i := 0; i < ex.navegacoes; i++ {
		tipo := entity.TipoNavegacao
		if i%3 == 2 {
			tipo = entity.TipoInteracao
		}
		if err := registrar(tipo, 0, "", 6*(i+1)); err != nil {
			return err
		}
	}
	if ex.processar {
		if _, err := e.processar.Executar(ctx, c.ID); err != nil {
			return err
		}
	}
	if ex.revogar {
		if err := e.consentimento.Revogar(ctx, c.ID); err != nil {
			return err
		}
	}
	if ex.desativar {
		c, err = e.clientes.BuscarPorID(ctx, c.ID)
		if err != nil {
			return err
		}
		c.Desativar()
		return e.clientes.Salvar(ctx, c)
	}
	return nil
}
