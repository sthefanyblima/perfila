package usecase

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/shared/errs"
)

// Casos de uso dos direitos do titular (art. 18 da LGPD) e da retenção. Não exigem consentimento:
// corrigem, entregam ou apagam os dados sem usá-los para perfilamento (exceção de C-03).
// Spec: v1.9, 3.1, 3.5, RN-17, RN-18, Q-48 a Q-53.

// AtualizarCadastroUseCase corrige nome e e-mail a pedido do titular. Spec: Q-48.
type AtualizarCadastroUseCase struct{ clientes port.ClienteRepository }

// NovoAtualizarCadastroUseCase injeta as dependências (C-08).
func NovoAtualizarCadastroUseCase(clientes port.ClienteRepository) *AtualizarCadastroUseCase {
	return &AtualizarCadastroUseCase{clientes: clientes}
}

// Executar vale com qualquer estado do consentimento e com cliente inativo; cliente eliminado
// recusa com ErrClienteEliminado. Spec: Q-48, RN-17.
func (uc *AtualizarCadastroUseCase) Executar(ctx context.Context, clienteID uuid.UUID, nome, email string) (entity.Cliente, error) {
	c, err := uc.clientes.BuscarPorID(ctx, clienteID)
	if err != nil {
		return entity.Cliente{}, err
	}
	if err := c.AtualizarDados(nome, email); err != nil {
		return entity.Cliente{}, err
	}
	if err := uc.clientes.Salvar(ctx, c); err != nil {
		return entity.Cliente{}, err
	}
	return c, nil
}

// DadosTitular é tudo o que o sistema guarda de um titular. Spec: Q-49.
type DadosTitular struct {
	GeradoEm      time.Time
	Cliente       entity.Cliente
	Consentimento entity.ConsentimentoLGPD        // sintético em NAO_CONCEDIDO sem registro
	Historico     []entity.MovimentoConsentimento // ordem cronológica
	Eventos       []entity.EventoComportamental   // Timestamp crescente
	Perfil        *entity.PerfilCliente           // nil sem perfil
	Segmento      *entity.Segmento                // nil sem segmento
	Recomendacoes []entity.RecomendacaoProduto    // DataGeracao decrescente
	Produtos      map[uuid.UUID]entity.Produto    // os citados pelas recomendações
}

// ExportarDadosTitularUseCase atende acesso e portabilidade (art. 18, II e V). Spec: Q-49.
type ExportarDadosTitularUseCase struct {
	consentimento *GerenciarConsentimentoUseCase
	clientes      port.ClienteRepository
	leitor        leitorComportamento
	clock         port.Clock
}

// NovoExportarDadosTitularUseCase injeta as dependências (C-08).
func NovoExportarDadosTitularUseCase(clientes port.ClienteRepository, consentimento *GerenciarConsentimentoUseCase,
	eventos port.EventoRepository, perfis port.PerfilRepository, segmentos port.SegmentoRepository,
	recomendacoes port.RecomendacaoRepository, produtos port.ProdutoRepository, clock port.Clock) *ExportarDadosTitularUseCase {
	return &ExportarDadosTitularUseCase{consentimento: consentimento, clientes: clientes, clock: clock,
		leitor: leitorComportamento{eventos: eventos, perfis: perfis, segmentos: segmentos, recomendacoes: recomendacoes, produtos: produtos}}
}

// Executar lê tudo com qualquer estado do consentimento, inclusive do cliente eliminado: é o
// próprio titular recebendo os seus dados, não um uso para perfilamento. Spec: Q-49, C-03.
func (uc *ExportarDadosTitularUseCase) Executar(ctx context.Context, clienteID uuid.UUID) (DadosTitular, error) {
	c, err := uc.clientes.BuscarPorID(ctx, clienteID)
	if err != nil {
		return DadosTitular{}, err
	}
	d := DadosTitular{GeradoEm: uc.clock.Agora().UTC(), Cliente: c}
	if d.Consentimento, err = uc.consentimento.Consultar(ctx, clienteID); err != nil {
		return DadosTitular{}, err
	}
	if d.Historico, err = uc.consentimento.Historico(ctx, clienteID); err != nil {
		return DadosTitular{}, err
	}
	r, err := uc.leitor.ler(ctx, clienteID)
	if err != nil {
		return DadosTitular{}, err
	}
	sort.SliceStable(r.Eventos, func(i, j int) bool { return r.Eventos[i].Timestamp.Before(r.Eventos[j].Timestamp) })
	d.Eventos, d.Perfil, d.Segmento, d.Recomendacoes, d.Produtos = r.Eventos, r.Perfil, r.Segmento, r.Recomendacoes, r.Produtos
	return d, nil
}

// remocoes apaga o comportamento de um cliente: eventos, perfil e recomendações. Nunca toca o
// consentimento nem os movimentos (RN-10). Spec: RN-17 (passo 4), RN-18, Q-52.
type remocoes struct {
	eventos       port.EventoRepository
	perfis        port.PerfilRepository
	recomendacoes port.RecomendacaoRepository
}

// apagar devolve quantos registros removeu ao todo; sem dados, 0 e nenhum erro.
func (r remocoes) apagar(ctx context.Context, clienteID uuid.UUID) (int, error) {
	total := 0
	for _, remover := range []func(context.Context, uuid.UUID) (int, error){
		r.eventos.RemoverPorCliente, r.perfis.RemoverPorCliente, r.recomendacoes.RemoverPorCliente,
	} {
		n, err := remover(ctx, clienteID)
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}

// EliminarDadosTitularUseCase atende a eliminação (art. 18, VI). Spec: RN-17, Q-50.
type EliminarDadosTitularUseCase struct {
	clientes       port.ClienteRepository
	consentimentos port.ConsentimentoRepository
	remocoes       remocoes
	clock          port.Clock
}

// NovoEliminarDadosTitularUseCase injeta as dependências (C-08).
func NovoEliminarDadosTitularUseCase(clientes port.ClienteRepository, consentimentos port.ConsentimentoRepository,
	eventos port.EventoRepository, perfis port.PerfilRepository, recomendacoes port.RecomendacaoRepository,
	clock port.Clock) *EliminarDadosTitularUseCase {
	return &EliminarDadosTitularUseCase{clientes: clientes, consentimentos: consentimentos, clock: clock,
		remocoes: remocoes{eventos: eventos, perfis: perfis, recomendacoes: recomendacoes}}
}

// Executar segue a ordem da RN-17: cliente existe e não eliminado; confirmação igual ao nome;
// consentimento diferente de CONCEDIDO; comportamento apagado; por último, cadastro anonimizado.
func (uc *EliminarDadosTitularUseCase) Executar(ctx context.Context, clienteID uuid.UUID, confirmacao string) error {
	c, err := uc.clientes.BuscarPorID(ctx, clienteID)
	if err != nil {
		return err
	}
	if c.EstaEliminado() {
		return fmt.Errorf("%w: %s", errs.ErrClienteEliminado, clienteID)
	}
	if !ConfirmacaoConfere(confirmacao, c.Nome) {
		return fmt.Errorf("%w: a confirmação não confere com o nome do cliente %s", errs.ErrDadosInvalidos, clienteID)
	}
	cons, err := uc.consentimentos.BuscarPorCliente(ctx, clienteID)
	switch {
	case err == nil && cons.Status == entity.StatusConcedido:
		return fmt.Errorf("%w: revogue o consentimento do cliente %s antes de eliminar", errs.ErrTransicaoInvalida, clienteID)
	case err != nil && !errors.Is(err, errs.ErrConsentimentoNaoEncontrado):
		return err
	}
	if _, err := uc.remocoes.apagar(ctx, clienteID); err != nil {
		return err
	}
	if err := c.Anonimizar(uc.clock.Agora()); err != nil {
		return err
	}
	return uc.clientes.Salvar(ctx, c)
}

// ConfirmacaoConfere compara a confirmação digitada com o nome do cadastro sem diferenciar
// maiúsculas, acentos e espaços nas pontas. Confirmação vazia nunca confere. Spec: RN-17.
func ConfirmacaoConfere(confirmacao, nome string) bool {
	a, b := normalizarNome(confirmacao), normalizarNome(nome)
	return a != "" && a == b
}

// normalizarNome põe em minúsculas, tira os acentos do português e os espaços nas pontas.
func normalizarNome(s string) string {
	return strings.NewReplacer(
		"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"í", "i", "ì", "i", "î", "i", "ï", "i",
		"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
		"ú", "u", "ù", "u", "û", "u", "ü", "u", "ç", "c", "ñ", "n",
	).Replace(strings.ToLower(strings.TrimSpace(s)))
}

// AplicarRetencaoUseCase apaga o comportamento de quem revogou há o prazo ou mais. Spec: RN-18, Q-51.
type AplicarRetencaoUseCase struct {
	consentimentos port.ConsentimentoRepository
	remocoes       remocoes
	clock          port.Clock
	prazoDias      int
}

// NovoAplicarRetencaoUseCase injeta as dependências (C-08); prazoDias é RETENCAO_DIAS_REVOGADO.
func NovoAplicarRetencaoUseCase(consentimentos port.ConsentimentoRepository, eventos port.EventoRepository,
	perfis port.PerfilRepository, recomendacoes port.RecomendacaoRepository, clock port.Clock, prazoDias int) *AplicarRetencaoUseCase {
	return &AplicarRetencaoUseCase{consentimentos: consentimentos, clock: clock, prazoDias: prazoDias,
		remocoes: remocoes{eventos: eventos, perfis: perfis, recomendacoes: recomendacoes}}
}

// PrazoDias devolve o prazo configurado, para o painel explicar a regra.
func (uc *AplicarRetencaoUseCase) PrazoDias() int { return uc.prazoDias }

// Executar devolve quantos clientes tiveram algum dado apagado; sem nada vencido, 0. Cadastro,
// consentimento e histórico ficam. Prazo menor que 1 é configuração inválida. Spec: RN-18.
func (uc *AplicarRetencaoUseCase) Executar(ctx context.Context) (int, error) {
	if uc.prazoDias < 1 {
		return 0, fmt.Errorf("%w: prazo de retenção %d", errs.ErrDadosInvalidos, uc.prazoDias)
	}
	limite := uc.clock.Agora().UTC().AddDate(0, 0, -uc.prazoDias)
	ids, err := uc.consentimentos.ListarRevogadosAte(ctx, limite)
	if err != nil {
		return 0, err
	}
	atingidos := 0
	for _, id := range ids {
		n, err := uc.remocoes.apagar(ctx, id)
		if err != nil {
			return atingidos, err
		}
		if n > 0 {
			atingidos++
		}
	}
	return atingidos, nil
}
