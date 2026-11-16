package web

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"

	"perfila/internal/application/usecase"
	"perfila/internal/domain/entity"
	"perfila/internal/shared/errs"
)

// Públicos para campanha (Spec 3.7.1, v1.11, UC-06 parcial, RN-21, RN-22, Q-66).
func (p *painel) rotasPublicos(mux *http.ServeMux, sessao func(entity.Permissao, http.HandlerFunc) http.Handler) {
	mux.Handle("GET /painel/publicos", sessao(entity.PermPublicos, p.telaPublicos))
	mux.Handle("GET /painel/publicos/{tipo}", sessao(entity.PermPublicos, p.confirmarExportacao))
	mux.Handle("POST /painel/publicos/{tipo}", sessao(entity.PermPublicos, p.exportar))
	mux.Handle("GET /painel/publicos/exportacoes/{id}/arquivo", sessao(entity.PermPublicos, p.baixarArquivo))
	mux.Handle("GET /painel/publicos/relatorio", sessao(entity.PermPublicos, p.telaRelatorio))
	mux.Handle("GET /painel/publicos/relatorio.csv", sessao(entity.PermPublicos, p.relatorioCSV))
}

// exportacaoVista é uma linha do histórico, com o arquivo disponível ou não.
type exportacaoVista struct {
	entity.ExportacaoPublico
	Disponivel bool
}

type dadosPublicos struct {
	Segmentos  []entity.Segmento
	Escolhido  *entity.Segmento
	Publico    usecase.PublicoGerado
	Historico  []exportacaoVista
	Gerada     *exportacaoVista // a exportação que acabou de ser gerada (PRG)
	Confirmar  string           // "exportar" ou "remocoes" na página de confirmação
	Horas      int
	Erro       *alerta
	Quantidade int // linhas do arquivo que a confirmação vai gerar
}

func tipoValido(t string) bool { return t == "exportar" || t == "remocoes" }

// carregarPublicos lê os segmentos e, com um segmento escolhido, a prévia e o histórico.
func (p *painel) carregarPublicos(w http.ResponseWriter, r *http.Request) (dadosPublicos, bool) {
	d := dadosPublicos{Horas: p.d.PublicosArquivoHoras}
	segs, err := p.d.ListarSegmentos.Executar(r.Context())
	if err != nil {
		p.falha(w, r, err, "")
		return d, false
	}
	d.Segmentos = segs
	bruto := r.FormValue("segmento")
	if bruto == "" {
		return d, true
	}
	segID, err := uuid.Parse(bruto)
	if err != nil {
		p.naoEncontrada(w, r)
		return d, false
	}
	if d.Publico, err = p.d.GerarPublico.Executar(r.Context(), segID); err != nil {
		p.falha(w, r, err, "")
		return d, false
	}
	d.Escolhido = &d.Publico.Segmento
	hist, err := p.d.ExportarPublico.Historico(r.Context(), segID)
	if err != nil {
		p.falha(w, r, err, "")
		return d, false
	}
	agora := p.d.Clock.Agora()
	for _, e := range hist {
		v := exportacaoVista{ExportacaoPublico: e, Disponivel: e.ArquivoExpiraEm.After(agora)}
		d.Historico = append(d.Historico, v)
		if r.URL.Query().Get("exportacao") == e.ID.String() {
			g := v
			d.Gerada = &g
		}
	}
	return d, true
}

// GET /painel/publicos?segmento= — prévia, remoções pendentes e histórico; não grava nada.
func (p *painel) telaPublicos(w http.ResponseWriter, r *http.Request) {
	if d, ok := p.carregarPublicos(w, r); ok {
		p.render(w, r, http.StatusOK, "publicos", pagina{Titulo: "Públicos", Nav: "publicos", Dados: d})
	}
}

// GET /painel/publicos/{exportar|remocoes}?segmento= — confirmação com o efeito escrito (Q-66).
func (p *painel) confirmarExportacao(w http.ResponseWriter, r *http.Request) {
	tipo := r.PathValue("tipo")
	if !tipoValido(tipo) {
		p.naoEncontrada(w, r)
		return
	}
	d, ok := p.carregarPublicos(w, r)
	if !ok {
		return
	}
	if d.Escolhido == nil {
		http.Redirect(w, r, "/painel/publicos", http.StatusSeeOther)
		return
	}
	d.Confirmar, d.Quantidade = tipo, len(d.Publico.Membros)
	if tipo == "remocoes" {
		d.Quantidade = d.Publico.RemocoesPendentes
	}
	p.render(w, r, http.StatusOK, "publicos", pagina{Titulo: "Gerar arquivo do público", Nav: "publicos", Dados: d})
}

// POST /painel/publicos/{exportar|remocoes} — ExportarPublico, depois PRG para a tela com o link.
func (p *painel) exportar(w http.ResponseWriter, r *http.Request) {
	tipo := r.PathValue("tipo")
	segID, err := uuid.Parse(r.FormValue("segmento"))
	if !tipoValido(tipo) || err != nil {
		p.naoEncontrada(w, r)
		return
	}
	ctx := withOperador(r)
	var e entity.ExportacaoPublico
	acao := entity.AcaoExportarPublico
	if tipo == "remocoes" {
		acao = entity.AcaoExportarRemocoesPublico
		e, err = p.d.ExportarPublico.ExportarRemocoes(ctx, segID)
	} else {
		e, err = p.d.ExportarPublico.Exportar(ctx, segID)
	}
	p.auditar(r, acao, entity.AlvoSegmento, segID, err) // RN-13, Q-68
	if err != nil {
		d, ok := p.carregarPublicos(w, r)
		if !ok {
			return
		}
		p.logarFalha(r, err)
		titulo, texto := mensagem(err, "publico-"+tipo)
		d.Confirmar, d.Erro = tipo, &alerta{Tom: "critical", Titulo: titulo, Texto: texto, Agora: true}
		p.render(w, r, statusDe(err), "publicos", pagina{Titulo: "Gerar arquivo do público", Nav: "publicos", Dados: d})
		return
	}
	p.acao(r, "exportar_publico_"+string(e.Tipo), "")
	http.Redirect(w, r, fmt.Sprintf("/painel/publicos?segmento=%s&exportacao=%s&aviso=arquivo-gerado", segID, e.ID), http.StatusSeeOther)
}

// GET /painel/publicos/exportacoes/{id}/arquivo — o CSV dentro do prazo (Q-64). É leitura: sem o
// registro na trilha, com o segmento como alvo, nenhum arquivo sai (RN-13, Q-68, Q-69).
func (p *painel) baixarArquivo(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		p.naoEncontrada(w, r)
		return
	}
	exportacao, recibo, err := p.d.ExportarPublico.Baixar(r.Context(), id)
	tipo, alvo := entity.AlvoNenhum, uuid.Nil
	if err == nil {
		tipo, alvo = entity.AlvoSegmento, exportacao.SegmentoID
	}
	if e := p.auditar(r, entity.AcaoBaixarArquivoPublico, tipo, alvo, err); e != nil && err == nil {
		p.falha(w, r, e, "")
		return
	}
	if err != nil {
		titulo, texto := mensagem(err, "")
		if errs.Codigo(err) == errs.ErrArquivoIndisponivel.Codigo {
			texto = fmt.Sprintf("Este arquivo não está mais disponível: ele fica guardado por %d horas. Gere um novo, se precisar.", p.d.PublicosArquivoHoras)
		} else {
			p.logarFalha(r, err)
		}
		p.renderErro(w, r, statusDe(err), titulo, texto)
		return
	}
	p.acao(r, "baixar_publico", "")
	escreverCSV(w, recibo.NomeArquivo, recibo.Conteudo)
}

// GET /painel/publicos/relatorio — agregado por segmento, sem identificador (RN-22).
func (p *painel) telaRelatorio(w http.ResponseWriter, r *http.Request) {
	linhas, err := p.d.Relatorio.Executar(r.Context())
	if err != nil {
		p.falha(w, r, err, "")
		return
	}
	p.render(w, r, http.StatusOK, "relatorio", pagina{Titulo: "Relatório agregado", Nav: "relatorio", Dados: linhas})
}

// GET /painel/publicos/relatorio.csv — o mesmo agregado em CSV (RN-22).
func (p *painel) relatorioCSV(w http.ResponseWriter, r *http.Request) {
	linhas, err := p.d.Relatorio.Executar(r.Context())
	if err != nil {
		p.falha(w, r, err, "")
		return
	}
	escreverCSV(w, "relatorio-segmentos-"+p.d.Clock.Agora().UTC().Format("20060102")+".csv", p.d.Relatorio.CSV(linhas))
}

func escreverCSV(w http.ResponseWriter, nome string, conteudo []byte) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+url.PathEscape(nome)+"\"")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(conteudo)
}

// withOperador põe a origem painel e o login da sessão no contexto (RN-10, RN-21).
func withOperador(r *http.Request) context.Context {
	return usecase.ComOrigem(r.Context(), entity.OrigemPainel, sessaoDe(r).Operador)
}
