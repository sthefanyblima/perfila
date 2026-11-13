package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/shared/errs"
)

// PeriodoPadraoRelatorio é o período do relatório quando nada é escolhido (6 meses). Spec: RN-15.
const PeriodoPadraoRelatorio = 6

// PeriodoRelatorioValido diz se o período é um dos aceitos: 6 ou 12 meses. Spec: RN-15.
func PeriodoRelatorioValido(meses int) bool { return meses == 6 || meses == 12 }

// LinhaConsentimento: clientes num estado do consentimento. Spec: RN-15.
type LinhaConsentimento struct {
	Status   entity.StatusConsentimento
	Clientes entity.Contagem
}

// LinhaSegmento: Nome vazio = clientes com consentimento válido ainda não segmentados. Spec: RN-15.
type LinhaSegmento struct {
	Nome                    string
	Clientes                entity.Contagem
	TicketMedio, ScoreMedio entity.Media
}

// LinhaCategoria: Outras = soma das categorias com menos de 5 compradores, sem nome; Categorias
// diz quantas foram somadas. Spec: RN-15.
type LinhaCategoria struct {
	Categoria         string
	Outras            bool
	Categorias        int
	Compras, Clientes entity.Contagem
}

// LinhaMes: um mês-calendário a partir de Inicio (no fuso do relatório). Spec: RN-15.
type LinhaMes struct {
	Inicio                         time.Time
	Cadastros, Eventos, Revogacoes entity.Contagem
}

// Relatorio é o relatório agregado de UC-05 e UC-07: só valores já protegidos. Spec: 3.5, RN-15.
type Relatorio struct {
	GeradoEm               time.Time
	PeriodoMeses           int
	Desde, Ate             time.Time
	Clientes, Ativos       entity.Contagem
	Consentimento          []LinhaConsentimento
	TaxaConcessao          entity.Percentual
	Segmentos              []LinhaSegmento
	RecomendacoesAtivas    entity.Contagem
	RecomendacoesExpiradas entity.Contagem
	Categorias             []LinhaCategoria
	Meses                  []LinhaMes
}

// GerarRelatorioUseCase monta o relatório agregado. Não lê nenhum cliente individual: a base de
// cada indicador está no contrato do RelatorioRepository. Spec: 3.5, RN-15, Q-46.
type GerarRelatorioUseCase struct {
	relatorios port.RelatorioRepository
	segmentos  port.SegmentoRepository
	clock      port.Clock
	fuso       *time.Location
}

// NovoGerarRelatorioUseCase injeta as dependências (C-08); fuso nil = UTC.
func NovoGerarRelatorioUseCase(relatorios port.RelatorioRepository, segmentos port.SegmentoRepository,
	clock port.Clock, fuso *time.Location) *GerarRelatorioUseCase {
	if fuso == nil {
		fuso = time.UTC
	}
	return &GerarRelatorioUseCase{relatorios: relatorios, segmentos: segmentos, clock: clock, fuso: fuso}
}

// Executar gera o relatório dos últimos periodoMeses meses-calendário, contando o mês atual.
// Spec: RN-15.
func (uc *GerarRelatorioUseCase) Executar(ctx context.Context, periodoMeses int) (Relatorio, error) {
	if !PeriodoRelatorioValido(periodoMeses) {
		return Relatorio{}, fmt.Errorf("%w: período de %d meses (aceitos: 6 ou 12)", errs.ErrDadosInvalidos, periodoMeses)
	}
	agora := uc.clock.Agora()
	limites := uc.limites(agora, periodoMeses)
	rel := Relatorio{GeradoEm: agora, PeriodoMeses: periodoMeses, Desde: limites[0], Ate: agora}

	if err := uc.clientes(ctx, &rel); err != nil {
		return Relatorio{}, err
	}
	var err error
	if rel.Segmentos, err = uc.linhasSegmentos(ctx); err != nil {
		return Relatorio{}, err
	}
	recs, err := uc.relatorios.ContarRecomendacoes(ctx)
	if err != nil {
		return Relatorio{}, err
	}
	rel.RecomendacoesAtivas = entity.ProtegerContagem(recs.Ativas, recs.ClientesAtivas)
	rel.RecomendacoesExpiradas = entity.ProtegerContagem(recs.Expiradas, recs.ClientesExpiradas)
	if rel.Categorias, err = uc.linhasCategorias(ctx, limites[0], limites[len(limites)-1]); err != nil {
		return Relatorio{}, err
	}
	serie, err := uc.relatorios.SerieMensal(ctx, limites)
	if err != nil {
		return Relatorio{}, err
	}
	for _, m := range serie {
		rel.Meses = append(rel.Meses, LinhaMes{Inicio: m.Inicio.In(uc.fuso),
			Cadastros:  entity.ProtegerContagem(m.Cadastros, m.Cadastros),
			Eventos:    entity.ProtegerContagem(m.Eventos, m.ClientesEventos),
			Revogacoes: entity.ProtegerContagem(m.Revogacoes, m.Revogacoes)})
	}
	return rel, nil
}

// limites devolve o início de cada mês do período e o início do mês seguinte ao atual (UTC).
func (uc *GerarRelatorioUseCase) limites(agora time.Time, meses int) []time.Time {
	local := agora.In(uc.fuso)
	atual := time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, uc.fuso)
	out := make([]time.Time, 0, meses+1)
	for i := meses - 1; i >= -1; i-- {
		out = append(out, atual.AddDate(0, -i, 0).UTC())
	}
	return out
}

// clientes preenche cadastro e consentimento, que contam todos os clientes. Spec: RN-15.
func (uc *GerarRelatorioUseCase) clientes(ctx context.Context, rel *Relatorio) error {
	c, err := uc.relatorios.ContarClientes(ctx)
	if err != nil {
		return err
	}
	protegida := func(n int) entity.Contagem { return entity.ProtegerContagem(n, n) }
	rel.Clientes, rel.Ativos = protegida(c.Total), protegida(c.Ativos)
	concedido := protegida(c.Concedido)
	rel.Consentimento = []LinhaConsentimento{
		{Status: entity.StatusConcedido, Clientes: concedido},
		{Status: entity.StatusNaoConcedido, Clientes: protegida(c.NaoConcedido)},
		{Status: entity.StatusRevogado, Clientes: protegida(c.Revogado)},
	}
	rel.TaxaConcessao = entity.ProtegerPercentual(concedido, rel.Clientes)
	return nil
}

// linhasSegmentos: ativos por prioridade, depois inativos que ainda têm perfis, depois os ainda
// não segmentados (se houver). Spec: RN-15.
func (uc *GerarRelatorioUseCase) linhasSegmentos(ctx context.Context) ([]LinhaSegmento, error) {
	resumos, err := uc.relatorios.ResumirSegmentos(ctx)
	if err != nil {
		return nil, err
	}
	porID := map[uuid.UUID]port.ResumoSegmento{}
	var semSegmento port.ResumoSegmento
	for _, r := range resumos {
		if r.SegmentoID == nil {
			semSegmento = r
			continue
		}
		porID[*r.SegmentoID] = r
	}
	linha := func(nome string, r port.ResumoSegmento) LinhaSegmento {
		return LinhaSegmento{Nome: nome, Clientes: entity.ProtegerContagem(r.Clientes, r.Clientes),
			TicketMedio: entity.ProtegerMedia(r.SomaTicket, r.Clientes), ScoreMedio: entity.ProtegerMedia(r.SomaScore, r.Clientes)}
	}
	ativos, err := uc.segmentos.ListarAtivos(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]LinhaSegmento, 0, len(ativos)+1)
	for _, s := range ativos {
		out = append(out, linha(s.Nome, porID[s.ID]))
		delete(porID, s.ID)
	}
	for _, r := range resumos { // na ordem do repositório, só os que sobraram (inativos)
		if r.SegmentoID == nil || r.Clientes == 0 {
			continue
		}
		if _, ok := porID[*r.SegmentoID]; !ok {
			continue
		}
		s, err := uc.segmentos.BuscarPorID(ctx, *r.SegmentoID)
		if err != nil {
			return nil, err
		}
		out = append(out, linha(s.Nome, r))
	}
	if semSegmento.Clientes > 0 {
		out = append(out, linha("", semSegmento))
	}
	return out, nil
}

// linhasCategorias: categorias com pelo menos MinimoGrupo compradores pelo nome; as demais somadas
// em "Outras categorias", protegidas pelos compradores distintos do conjunto. Spec: RN-15.
func (uc *GerarRelatorioUseCase) linhasCategorias(ctx context.Context, desde, ate time.Time) ([]LinhaCategoria, error) {
	compras, err := uc.relatorios.ContarCompras(ctx, desde, ate)
	if err != nil {
		return nil, err
	}
	out := []LinhaCategoria{}
	var pequenas []string
	somaPequenas := 0
	for _, c := range compras {
		if c.Clientes < entity.MinimoGrupo {
			pequenas = append(pequenas, c.Categoria)
			somaPequenas += c.Compras
			continue
		}
		out = append(out, LinhaCategoria{Categoria: c.Categoria, Categorias: 1,
			Compras: entity.ProtegerContagem(c.Compras, c.Clientes), Clientes: entity.ProtegerContagem(c.Clientes, c.Clientes)})
	}
	if len(pequenas) == 0 {
		return out, nil
	}
	compradores, err := uc.relatorios.ContarCompradores(ctx, desde, ate, pequenas)
	if err != nil {
		return nil, err
	}
	return append(out, LinhaCategoria{Outras: true, Categorias: len(pequenas),
		Compras: entity.ProtegerContagem(somaPequenas, compradores), Clientes: entity.ProtegerContagem(compradores, compradores)}), nil
}
