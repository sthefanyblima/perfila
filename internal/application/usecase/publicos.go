package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"time"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/domain/service"
	"perfila/internal/shared/errs"
)

// CandidatoPublico é um cliente que entra no público: só o ID (para o registro interno) e o
// hash do e-mail (o que vai no arquivo). Spec: RN-17.
type CandidatoPublico struct {
	ClienteID uuid.UUID
	HashEmail string
}

// PublicoGerado é a prévia do público de um segmento; não grava nada. Spec: UC-06 parcial, RN-17.
type PublicoGerado struct {
	Segmento          entity.Segmento
	Membros           []CandidatoPublico // em ordem de hash
	Fora              int                // perfis do segmento que ficaram de fora
	RemocoesPendentes int
}

// GerarPublicoUseCase aplica a seleção da RN-17. Spec: 3.5, RN-17, C-03.
type GerarPublicoUseCase struct {
	clientes  port.ClienteRepository
	perfis    port.PerfilRepository
	segmentos port.SegmentoRepository
	publicos  port.PublicoRepository
	validador *service.ValidadorLGPD
}

// NovoGerarPublicoUseCase injeta as dependências (C-08).
func NovoGerarPublicoUseCase(clientes port.ClienteRepository, perfis port.PerfilRepository, segmentos port.SegmentoRepository,
	publicos port.PublicoRepository, validador *service.ValidadorLGPD) *GerarPublicoUseCase {
	return &GerarPublicoUseCase{clientes: clientes, perfis: perfis, segmentos: segmentos, publicos: publicos, validador: validador}
}

// Executar: segmento inexistente ou inativo é ErrDadosInvalidos. O consentimento das duas
// finalidades é conferido antes de ler o cadastro (C-03). Spec: RN-17.
func (uc *GerarPublicoUseCase) Executar(ctx context.Context, segmentoID uuid.UUID) (PublicoGerado, error) {
	seg, err := uc.segmentoAtivo(ctx, segmentoID)
	if err != nil {
		return PublicoGerado{}, err
	}
	perfis, err := uc.perfis.ListarPorSegmento(ctx, seg.ID)
	if err != nil {
		return PublicoGerado{}, err
	}
	out := PublicoGerado{Segmento: seg, Membros: []CandidatoPublico{}}
	for _, p := range perfis {
		c, entra, err := uc.qualifica(ctx, p.ClienteID)
		if err != nil {
			return PublicoGerado{}, err
		}
		if !entra {
			out.Fora++
			continue
		}
		out.Membros = append(out.Membros, CandidatoPublico{ClienteID: c.ID, HashEmail: entity.HashEmail(c.Email)})
	}
	sort.Slice(out.Membros, func(i, j int) bool { return out.Membros[i].HashEmail < out.Membros[j].HashEmail })
	pendentes, err := uc.publicos.ListarMembrosPorSegmento(ctx, seg.ID, entity.MembroRemocaoPendente)
	if err != nil {
		return PublicoGerado{}, err
	}
	out.RemocoesPendentes = len(pendentes)
	return out, nil
}

func (uc *GerarPublicoUseCase) segmentoAtivo(ctx context.Context, id uuid.UUID) (entity.Segmento, error) {
	seg, err := uc.segmentos.BuscarPorID(ctx, id)
	if err != nil {
		return entity.Segmento{}, err
	}
	if !seg.Ativo {
		return entity.Segmento{}, fmt.Errorf("%w: segmento %s inativo", errs.ErrDadosInvalidos, id)
	}
	return seg, nil
}

// qualifica: segmentacao e publicidade_terceiros CONCEDIDO e cliente ativo. Spec: RN-17.
func (uc *GerarPublicoUseCase) qualifica(ctx context.Context, clienteID uuid.UUID) (entity.Cliente, bool, error) {
	for _, f := range entity.Finalidades() {
		ok, err := uc.validador.ConsentimentoValidoPara(ctx, clienteID, f)
		if err != nil || !ok {
			return entity.Cliente{}, false, err
		}
	}
	c, err := uc.clientes.BuscarPorID(ctx, clienteID)
	if errors.Is(err, errs.ErrClienteNaoEncontrado) {
		return entity.Cliente{}, false, nil
	}
	if err != nil {
		return entity.Cliente{}, false, err
	}
	return c, c.Ativo, nil
}

// ExportarPublicoUseCase gera os arquivos de inclusão e de remoção, guarda o registro e entrega o
// arquivo dentro do prazo. Spec: 3.5, RN-17, Q-50, Q-51.
type ExportarPublicoUseCase struct {
	gerar    *GerarPublicoUseCase
	publicos port.PublicoRepository
	destino  port.DestinoPublicitarioGateway
	clock    port.Clock
	prazo    time.Duration
	log      *slog.Logger
}

// NovoExportarPublicoUseCase injeta as dependências (C-08). horas é PUBLICOS_ARQUIVO_HORAS; log nil
// descarta os registros.
func NovoExportarPublicoUseCase(gerar *GerarPublicoUseCase, publicos port.PublicoRepository, destino port.DestinoPublicitarioGateway,
	clock port.Clock, horas int, log *slog.Logger) *ExportarPublicoUseCase {
	return &ExportarPublicoUseCase{gerar: gerar, publicos: publicos, destino: destino, clock: clock,
		prazo: time.Duration(horas) * time.Hour, log: logOuDescarte(log)}
}

// Exportar gera o arquivo de inclusão. Quem já estava EXPORTADO no segmento e não entra mais passa
// a REMOCAO_PENDENTE; quem entra e ainda não era membro vira membro novo. Spec: RN-17.
func (uc *ExportarPublicoUseCase) Exportar(ctx context.Context, segmentoID uuid.UUID) (entity.ExportacaoPublico, error) {
	agora := uc.clock.Agora().UTC()
	if _, err := uc.publicos.ApagarArquivosVencidos(ctx, agora); err != nil {
		return entity.ExportacaoPublico{}, err
	}
	pub, err := uc.gerar.Executar(ctx, segmentoID)
	if err != nil {
		return entity.ExportacaoPublico{}, err
	}
	if len(pub.Membros) == 0 {
		return entity.ExportacaoPublico{}, fmt.Errorf("%w: nenhum cliente entra no público %s", errs.ErrDadosInvalidos, segmentoID)
	}
	exportados, err := uc.publicos.ListarMembrosPorSegmento(ctx, pub.Segmento.ID, entity.MembroExportado)
	if err != nil {
		return entity.ExportacaoPublico{}, err
	}
	entram := map[uuid.UUID]bool{}
	hashes := make([]string, 0, len(pub.Membros))
	for _, c := range pub.Membros {
		entram[c.ClienteID] = true
		hashes = append(hashes, c.HashEmail)
	}
	var membros []entity.MembroPublico
	ja := map[uuid.UUID]bool{}
	for _, m := range exportados {
		if entram[m.ClienteID] {
			ja[m.ClienteID] = true
			continue
		}
		if err := m.MarcarRemocao(agora); err != nil {
			return entity.ExportacaoPublico{}, err
		}
		membros = append(membros, m)
	}
	e, conteudo, err := uc.enviar(ctx, pub.Segmento, entity.ExportacaoInclusao, hashes, agora)
	if err != nil {
		return entity.ExportacaoPublico{}, err
	}
	for _, c := range pub.Membros {
		if ja[c.ClienteID] {
			continue
		}
		m, err := entity.NovoMembro(uuid.New(), e.ID, pub.Segmento.ID, c.ClienteID, c.HashEmail, agora)
		if err != nil {
			return entity.ExportacaoPublico{}, err
		}
		membros = append(membros, m)
	}
	return uc.registrar(ctx, e, conteudo, membros)
}

// ExportarRemocoes gera o arquivo de remoção com os pendentes do segmento e os passa a REMOVIDO.
// Spec: RN-17, Q-50.
func (uc *ExportarPublicoUseCase) ExportarRemocoes(ctx context.Context, segmentoID uuid.UUID) (entity.ExportacaoPublico, error) {
	agora := uc.clock.Agora().UTC()
	if _, err := uc.publicos.ApagarArquivosVencidos(ctx, agora); err != nil {
		return entity.ExportacaoPublico{}, err
	}
	seg, err := uc.gerar.segmentoAtivo(ctx, segmentoID)
	if err != nil {
		return entity.ExportacaoPublico{}, err
	}
	pendentes, err := uc.publicos.ListarMembrosPorSegmento(ctx, seg.ID, entity.MembroRemocaoPendente)
	if err != nil {
		return entity.ExportacaoPublico{}, err
	}
	if len(pendentes) == 0 {
		return entity.ExportacaoPublico{}, fmt.Errorf("%w: sem remoção pendente no segmento %s", errs.ErrDadosInvalidos, segmentoID)
	}
	vistos := map[string]bool{}
	hashes := []string{}
	for _, m := range pendentes {
		if !vistos[m.HashEmail] {
			vistos[m.HashEmail] = true
			hashes = append(hashes, m.HashEmail)
		}
	}
	e, conteudo, err := uc.enviar(ctx, seg, entity.ExportacaoRemocao, hashes, agora)
	if err != nil {
		return entity.ExportacaoPublico{}, err
	}
	for i := range pendentes {
		if err := pendentes[i].Remover(e.ID, agora); err != nil {
			return entity.ExportacaoPublico{}, err
		}
	}
	return uc.registrar(ctx, e, conteudo, pendentes)
}

func (uc *ExportarPublicoUseCase) enviar(ctx context.Context, seg entity.Segmento, tipo entity.TipoExportacao, hashes []string, agora time.Time) (entity.ExportacaoPublico, []byte, error) {
	recibo, err := uc.destino.Enviar(ctx, entity.LotePublico{Tipo: tipo, SegmentoNome: seg.Nome, Data: agora, Hashes: hashes})
	if err != nil {
		return entity.ExportacaoPublico{}, nil, err
	}
	_, operador := origemDe(ctx)
	if operador == "" {
		operador = string(entity.OrigemSistema)
	}
	return entity.ExportacaoPublico{ID: uuid.New(), SegmentoID: seg.ID, SegmentoNome: seg.Nome, Tipo: tipo,
		Destino: recibo.Destino, Operador: operador, Data: agora, Quantidade: len(hashes), HashArquivo: recibo.Hash,
		NomeArquivo: recibo.NomeArquivo, ArquivoExpiraEm: agora.Add(uc.prazo)}, recibo.Conteudo, nil
}

func (uc *ExportarPublicoUseCase) registrar(ctx context.Context, e entity.ExportacaoPublico, conteudo []byte, membros []entity.MembroPublico) (entity.ExportacaoPublico, error) {
	if err := uc.publicos.RegistrarExportacao(ctx, e, conteudo, membros); err != nil {
		return entity.ExportacaoPublico{}, err
	}
	// C-09: só IDs e números; nunca e-mail, hash de e-mail nem nome.
	uc.log.Info("publico_exportado", "operador", e.Operador, "tipo", string(e.Tipo), "segmentoId", e.SegmentoID.String(),
		"exportacaoId", e.ID.String(), "quantidade", e.Quantidade, "hashArquivo", e.HashArquivo)
	return e, nil
}

// Baixar entrega o arquivo dentro do prazo; depois, ErrArquivoIndisponivel. Spec: Q-51.
func (uc *ExportarPublicoUseCase) Baixar(ctx context.Context, exportacaoID uuid.UUID) (entity.ReciboEnvio, error) {
	agora := uc.clock.Agora().UTC()
	if _, err := uc.publicos.ApagarArquivosVencidos(ctx, agora); err != nil {
		return entity.ReciboEnvio{}, err
	}
	e, conteudo, err := uc.publicos.BuscarArquivo(ctx, exportacaoID)
	if err != nil {
		return entity.ReciboEnvio{}, err
	}
	if !e.ArquivoExpiraEm.After(agora) {
		return entity.ReciboEnvio{}, errs.ErrArquivoIndisponivel
	}
	return entity.ReciboEnvio{Destino: e.Destino, NomeArquivo: e.NomeArquivo, Conteudo: conteudo, Hash: e.HashArquivo}, nil
}

// Historico devolve as exportações do segmento, da mais recente para a mais antiga.
func (uc *ExportarPublicoUseCase) Historico(ctx context.Context, segmentoID uuid.UUID) ([]entity.ExportacaoPublico, error) {
	return uc.publicos.ListarExportacoes(ctx, segmentoID)
}

// PropagarRevogacaoUseCase tira o cliente dos públicos já exportados (art. 18 §6º). Spec: RN-17, Q-50.
type PropagarRevogacaoUseCase struct {
	publicos port.PublicoRepository
	clock    port.Clock
	log      *slog.Logger
}

// NovoPropagarRevogacaoUseCase injeta as dependências (C-08); log nil descarta os registros.
func NovoPropagarRevogacaoUseCase(publicos port.PublicoRepository, clock port.Clock, log *slog.Logger) *PropagarRevogacaoUseCase {
	return &PropagarRevogacaoUseCase{publicos: publicos, clock: clock, log: logOuDescarte(log)}
}

// Executar passa os membros EXPORTADO do cliente a REMOCAO_PENDENTE e devolve quantos marcou.
// A falha vai para o log, para quem revogou não perder a revogação (Q-50).
func (uc *PropagarRevogacaoUseCase) Executar(ctx context.Context, clienteID uuid.UUID) (int, error) {
	n, err := uc.marcar(ctx, clienteID)
	if err != nil {
		uc.log.Error("remocao_propagada_falhou", "clienteId", clienteID.String(), "erro", err.Error())
		return 0, err
	}
	if n > 0 {
		uc.log.Info("remocao_propagada", "clienteId", clienteID.String(), "quantidade", n)
	}
	return n, nil
}

func (uc *PropagarRevogacaoUseCase) marcar(ctx context.Context, clienteID uuid.UUID) (int, error) {
	ms, err := uc.publicos.ListarMembrosPorCliente(ctx, clienteID, entity.MembroExportado)
	if err != nil || len(ms) == 0 {
		return 0, err
	}
	agora := uc.clock.Agora()
	for i := range ms {
		if err := ms[i].MarcarRemocao(agora); err != nil {
			return 0, err
		}
	}
	if err := uc.publicos.AtualizarMembros(ctx, ms); err != nil {
		return 0, err
	}
	return len(ms), nil
}

func logOuDescarte(log *slog.Logger) *slog.Logger {
	if log == nil {
		return slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return log
}
