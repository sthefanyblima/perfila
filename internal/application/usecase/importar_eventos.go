package usecase

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/shared/errs"
)

// Limites de um lote de eventos. Spec: RN-16, Q-47.
const (
	LimiteLinhasImportacao = 1000
	LimiteBytesImportacao  = 256 * 1024
)

// MotivoArquivo diz por que um lote foi recusado inteiro. Spec: RN-16.
type MotivoArquivo string

const (
	MotivoArquivoVazio        MotivoArquivo = "arquivo_vazio"
	MotivoArquivoGrande       MotivoArquivo = "arquivo_grande"
	MotivoLinhasDemais        MotivoArquivo = "linhas_demais"
	MotivoCodificacaoInvalida MotivoArquivo = "codificacao_invalida"
	MotivoCabecalhoInvalido   MotivoArquivo = "cabecalho_invalido"
	MotivoFormatoInvalido     MotivoArquivo = "formato_invalido"
)

// ErroArquivo recusa o lote inteiro; embrulha ErrDadosInvalidos (C-06). Spec: RN-16.
type ErroArquivo struct{ Motivo MotivoArquivo }

func (e *ErroArquivo) Error() string { return "lote de eventos recusado: " + string(e.Motivo) }

func (e *ErroArquivo) Unwrap() error { return errs.ErrDadosInvalidos }

// EntradaEvento é uma linha do lote, com os campos em texto como chegaram (CSV ou JSON).
// Numero é a linha no arquivo (cabeçalho = 1) ou a posição no lote JSON (a partir de 1).
type EntradaEvento struct {
	Numero                                               int
	ClienteID, Tipo, Origem, Timestamp, Valor, Categoria string
	ColunasErradas                                       bool
}

// SituacaoLinha de uma linha não aceita. Spec: RN-16.
type SituacaoLinha string

const (
	SituacaoDuplicada SituacaoLinha = "duplicada"
	SituacaoRejeitada SituacaoLinha = "rejeitada"
)

// Motivos de dados_invalidos numa linha: só detalham a mensagem (RN-16).
const (
	MotivoClienteID       = "cliente_id"
	MotivoTipo            = "tipo"
	MotivoOrigem          = "origem"
	MotivoTimestamp       = "timestamp"
	MotivoTimestampFuturo = "timestamp_futuro"
	MotivoValor           = "valor"
	MotivoValorCompra     = "valor_compra"
	MotivoColunas         = "colunas"
)

// LinhaImportacao é uma linha não aceita do lote: duplicada ou rejeitada, com o erro.
type LinhaImportacao struct {
	Numero    int
	ClienteID uuid.UUID // nulo se ilegível
	Situacao  SituacaoLinha
	Erro      error  // a sentinela (errs.Codigo); nil em duplicada
	Motivo    string // em dados_invalidos: o campo
}

// ResultadoImportacao é o relatório do lote. Spec: RN-16.
type ResultadoImportacao struct {
	Linhas                          []LinhaImportacao
	Aceitas, Duplicadas, Rejeitadas int
	Gravado                         bool
}

// ImportarEventosUseCase trata um lote linha a linha com as regras de UC-01. Spec: RN-16, Q-47.
type ImportarEventosUseCase struct {
	coletar *ColetarDadosUseCase
	pre     Precondicoes
	eventos port.EventoRepository
	clock   port.Clock
}

// NovoImportarEventosUseCase injeta as dependências (C-08).
func NovoImportarEventosUseCase(coletar *ColetarDadosUseCase, pre Precondicoes, eventos port.EventoRepository, clock port.Clock) *ImportarEventosUseCase {
	return &ImportarEventosUseCase{coletar: coletar, pre: pre, eventos: eventos, clock: clock}
}

// Previa aplica todas as regras sem gravar nada. Spec: RN-16.
func (uc *ImportarEventosUseCase) Previa(ctx context.Context, entradas []EntradaEvento) (ResultadoImportacao, error) {
	return uc.executar(ctx, entradas, false)
}

// Importar aplica as regras de novo e grava cada linha aceita com UC-01 Registrar. Spec: RN-16.
func (uc *ImportarEventosUseCase) Importar(ctx context.Context, entradas []EntradaEvento) (ResultadoImportacao, error) {
	return uc.executar(ctx, entradas, true)
}

// chaveEvento identifica um duplicado: cliente (pelo mapa), tipo, instante em µs e valor.
type chaveEvento struct {
	tipo     entity.TipoEvento
	instante int64
	valor    float64
}

func chaveDe(e entity.EventoComportamental) chaveEvento {
	return chaveEvento{tipo: e.Tipo, instante: e.Timestamp.UTC().UnixMicro(), valor: e.Valor}
}

func (uc *ImportarEventosUseCase) executar(ctx context.Context, entradas []EntradaEvento, gravar bool) (ResultadoImportacao, error) {
	switch {
	case len(entradas) == 0:
		return ResultadoImportacao{}, &ErroArquivo{Motivo: MotivoArquivoVazio}
	case len(entradas) > LimiteLinhasImportacao:
		return ResultadoImportacao{}, &ErroArquivo{Motivo: MotivoLinhasDemais}
	}
	agora := uc.clock.Agora()
	res := ResultadoImportacao{Gravado: gravar}
	precondicao := map[uuid.UUID]error{}               // RN-01 por cliente, uma vez por lote
	conhecidos := map[uuid.UUID]map[chaveEvento]bool{} // gravados e aceitos neste lote
	rejeitar := func(n int, id uuid.UUID, err error, motivo string) {
		res.Rejeitadas++
		res.Linhas = append(res.Linhas, LinhaImportacao{Numero: n, ClienteID: id, Situacao: SituacaoRejeitada, Erro: err, Motivo: motivo})
	}
	for _, en := range entradas {
		ev, motivo := converter(en)
		if motivo != "" {
			rejeitar(en.Numero, ev.ClienteID, errs.ErrDadosInvalidos, motivo)
			continue
		}
		err, visto := precondicao[ev.ClienteID]
		if !visto {
			_, err = uc.pre.Verificar(ctx, ev.ClienteID)
			if err != nil && errs.Codigo(err) == "" {
				return res, err // falha de infraestrutura interrompe o lote
			}
			precondicao[ev.ClienteID] = err
		}
		if err != nil {
			rejeitar(en.Numero, ev.ClienteID, err, "")
			continue
		}
		if err := ev.Validar(agora); err != nil {
			rejeitar(en.Numero, ev.ClienteID, err, motivoRN04(ev, agora))
			continue
		}
		conjunto, err := uc.conhecidos(ctx, ev.ClienteID, conhecidos)
		if err != nil {
			return res, err
		}
		if chave := chaveDe(ev); conjunto[chave] {
			res.Duplicadas++
			res.Linhas = append(res.Linhas, LinhaImportacao{Numero: en.Numero, ClienteID: ev.ClienteID, Situacao: SituacaoDuplicada})
			continue
		}
		if gravar {
			if _, err := uc.coletar.Registrar(ctx, ev); err != nil {
				if errs.Codigo(err) == "" {
					return res, err
				}
				rejeitar(en.Numero, ev.ClienteID, err, motivoRN04(ev, agora))
				continue
			}
		}
		conjunto[chaveDe(ev)] = true
		res.Aceitas++
	}
	return res, nil
}

// conhecidos carrega uma vez os eventos gravados do cliente (já com RN-01 conferida, C-03).
func (uc *ImportarEventosUseCase) conhecidos(ctx context.Context, id uuid.UUID, cache map[uuid.UUID]map[chaveEvento]bool) (map[chaveEvento]bool, error) {
	if c, ok := cache[id]; ok {
		return c, nil
	}
	evs, err := uc.eventos.BuscarPorCliente(ctx, id)
	if err != nil {
		return nil, err
	}
	c := make(map[chaveEvento]bool, len(evs))
	for _, e := range evs {
		c[chaveDe(e)] = true
	}
	cache[id] = c
	return c, nil
}

// converter lê o formato da linha (RN-16); devolve o motivo se algum campo não for legível.
func converter(en EntradaEvento) (entity.EventoComportamental, string) {
	ev := entity.EventoComportamental{Tipo: entity.TipoEvento(strings.TrimSpace(en.Tipo)),
		Origem: strings.TrimSpace(en.Origem), Categoria: strings.TrimSpace(en.Categoria)}
	id, err := uuid.Parse(strings.TrimSpace(en.ClienteID))
	if err == nil {
		ev.ClienteID = id
	}
	switch {
	case en.ColunasErradas:
		return ev, MotivoColunas
	case err != nil:
		return ev, MotivoClienteID
	}
	quando, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(en.Timestamp))
	if err != nil {
		return ev, MotivoTimestamp
	}
	ev.Timestamp = quando.UTC()
	valor := strings.TrimSpace(en.Valor)
	if valor != "" {
		if !decimalComPonto(valor) {
			return ev, MotivoValor
		}
		if ev.Valor, err = strconv.ParseFloat(valor, 64); err != nil {
			return ev, MotivoValor
		}
	}
	return ev, ""
}

// decimalComPonto aceita só [-]dígitos[.dígitos]: nada de vírgula, expoente, Inf ou NaN.
func decimalComPonto(s string) bool {
	s = strings.TrimPrefix(s, "-")
	inteira, fracao, temPonto := strings.Cut(s, ".")
	soDigitos := func(p string) bool {
		for _, r := range p {
			if r < '0' || r > '9' {
				return false
			}
		}
		return p != ""
	}
	return soDigitos(inteira) && (!temPonto || soDigitos(fracao))
}

// motivoRN04 diz qual campo RN-04 recusou, na mesma ordem de Validar, só para a mensagem.
func motivoRN04(e entity.EventoComportamental, agora time.Time) string {
	switch {
	case !e.Tipo.Valido():
		return MotivoTipo
	case strings.TrimSpace(e.Origem) == "":
		return MotivoOrigem
	case e.Timestamp.IsZero():
		return MotivoTimestamp
	case e.Timestamp.After(agora.Add(entity.ToleranciaRelogio)):
		return MotivoTimestampFuturo
	case !(e.Valor >= 0) || e.Valor > entity.ValorMaximoEvento:
		return MotivoValor
	case e.Tipo == entity.TipoCompra && e.Valor <= 0:
		return MotivoValorCompra
	}
	return ""
}
