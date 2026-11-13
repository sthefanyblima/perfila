package contrato

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/shared/errs"
)

func campoErro(t *testing.T, err error, campo string) {
	t.Helper()
	var ec *errs.ErroCampo
	if !errors.As(err, &ec) || !errors.Is(err, errs.ErrDadosInvalidos) || ec.Campo != campo {
		t.Fatalf("esperava ErroCampo no campo %q, obtido %v", campo, err)
	}
}

// Spec: 3.6, RN-14, Q-33 — gravação de segmentos e unicidades repetidas no armazenamento.
func testarSegmentoGravacao(t *testing.T, r Repositorios) {
	ctx := context.Background()
	padrao := entity.Segmento{ID: uuid.New(), Nome: "padrao", Prioridade: 99, Ativo: true}
	premium := entity.Segmento{ID: uuid.New(), Nome: "premium", Criterio: "ticketMedio >= 500", Prioridade: 1, Ativo: true}
	inativo := entity.Segmento{ID: uuid.New(), Nome: "antigo", Criterio: "frequenciaCompras > 1", Prioridade: 1, Ativo: false}
	for _, s := range []entity.Segmento{padrao, premium, inativo} {
		ok(t, r.Segmentos.Salvar(ctx, s))
	}
	todos, err := r.Segmentos.ListarTodos(ctx)
	ok(t, err)
	if len(todos) != 3 || todos[2] != padrao {
		t.Fatalf("ListarTodos: %v", todos)
	}

	premium.Descricao, premium.Criterio, premium.Prioridade = "alto valor", "ticketMedio >= 600", 2
	ok(t, r.Segmentos.Salvar(ctx, premium))
	obtido, err := r.Segmentos.BuscarPorID(ctx, premium.ID)
	ok(t, err)
	igual(t, "segmento atualizado", obtido, premium)

	// Prioridade repetida só conta entre os ativos.
	inativo.Prioridade = 2
	ok(t, r.Segmentos.Salvar(ctx, inativo))
	inativo.Ativo = true
	campoErro(t, r.Segmentos.Salvar(ctx, inativo), "prioridade")
	// Nome repetido sem diferenciar maiúsculas, mesmo inativo.
	campoErro(t, r.Segmentos.Salvar(ctx, entity.Segmento{ID: uuid.New(), Nome: "PREMIUM", Criterio: "ticketMedio > 1", Prioridade: 5}), "nome")

	ativos, err := r.Segmentos.ListarAtivos(ctx)
	ok(t, err)
	igual(t, "nada mudou com as recusas", ativos, []entity.Segmento{premium, padrao})
}

// Spec: 3.6, Q-43 — gravação e busca de produtos.
func testarProdutoGravacao(t *testing.T, r Repositorios) {
	ctx := context.Background()
	p := entity.Produto{ID: uuid.New(), Nome: "Atlas", Categoria: "livros", Preco: 90.5, Ativo: true}
	ok(t, r.Produtos.Salvar(ctx, p))
	p.Nome, p.Categoria, p.Preco, p.Ativo = "Atlas Escolar", "educacao", 1_000_000, false
	ok(t, r.Produtos.Salvar(ctx, p))
	obtido, err := r.Produtos.BuscarPorID(ctx, p.ID)
	ok(t, err)
	igual(t, "produto atualizado", obtido, p)
	ativos, err := r.Produtos.ListarAtivosPorCategoria(ctx, "educacao")
	ok(t, err)
	igual(t, "inativo fora da RN-07", len(ativos), 0)
	_, err = r.Produtos.BuscarPorID(ctx, uuid.New())
	erroEsperado(t, err, errs.ErrProdutoNaoEncontrado)
}

// Spec: 3.6, Q-42 — só os registros CONCEDIDO, por ClienteID.
func testarConcedidos(t *testing.T, r Repositorios) {
	ctx := context.Background()
	vazio, err := r.Consentimentos.ListarConcedidos(ctx)
	ok(t, err)
	igual(t, "sem registros", len(vazio), 0)
	var esperados []entity.ConsentimentoLGPD
	for _, status := range []entity.StatusConsentimento{entity.StatusConcedido, entity.StatusRevogado, entity.StatusConcedido, entity.StatusNaoConcedido} {
		c := cliente(t, r)
		reg := entity.ConsentimentoLGPD{ID: uuid.New(), ClienteID: c.ID, Status: status, Finalidade: "segmentacao",
			VersaoTermo: "v1", DataConsentimento: agora, DataAtualizacao: agora}
		if status == entity.StatusRevogado {
			reg.DataRevogacao = ptr(agora)
		}
		ok(t, r.Consentimentos.Salvar(ctx, reg))
		if status == entity.StatusConcedido {
			esperados = append(esperados, reg)
		}
	}
	if esperados[0].ClienteID.String() > esperados[1].ClienteID.String() {
		esperados[0], esperados[1] = esperados[1], esperados[0]
	}
	obtidos, err := r.Consentimentos.ListarConcedidos(ctx)
	ok(t, err)
	igual(t, "concedidos por ClienteID", obtidos, esperados)
}

// Spec: RN-13, Q-44 — os nomes dos campos alterados voltam como foram gravados.
func testarAuditoriaCampos(t *testing.T, r Repositorios) {
	ctx := context.Background()
	com, err := entity.NovoRegistroAuditoria(uuid.New(), agora, "caio", entity.AcaoEditarSegmento, entity.AlvoSegmento,
		uuid.New(), entity.ResultadoSucesso, "criterio", "prioridade")
	ok(t, err)
	sem, err := entity.NovoRegistroAuditoria(uuid.New(), agora.Add(-time.Minute), "caio", entity.AcaoPreverSegmentacao,
		entity.AlvoNenhum, uuid.Nil, entity.ResultadoSucesso)
	ok(t, err)
	ok(t, r.Auditoria.Registrar(ctx, com))
	ok(t, r.Auditoria.Registrar(ctx, sem))
	lista, _, err := r.Auditoria.Listar(ctx, port.FiltroAuditoria{})
	ok(t, err)
	igual(t, "auditoria com campos", lista, []entity.RegistroAuditoria{com, sem})
}
