package web

import (
	"encoding/base64"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"perfila/internal/application/usecase"
)

// limitePedidoImportacao corta o pedido antes de ler o corpo: o arquivo (256 KiB) em base64 na
// confirmação, com folga para o multipart. Spec: 3.7.1, RN-16.
const limitePedidoImportacao = 512 << 10

func (p *painel) rotasImportacao(mux *http.ServeMux, sessao func(http.HandlerFunc) http.Handler) {
	mux.Handle("GET /painel/eventos/importar", sessao(p.telaImportar))
	mux.Handle("POST /painel/eventos/importar", limitarPedido(sessao(p.previaImportacao)))
	mux.Handle("POST /painel/eventos/importar/confirmar", limitarPedido(sessao(p.confirmarImportacao)))
	mux.Handle("GET /painel/eventos/importar/confirmar", http.RedirectHandler("/painel/eventos/importar", http.StatusSeeOther))
	mux.Handle("GET /painel/eventos/modelo.csv", sessao(p.modeloCSV))
}

// limitarPedido recusa o pedido grande pelo Content-Length, antes da sessão ler o formulário,
// e devolve a tela com o aviso pelo código (o texto nunca vem da URL).
func limitarPedido(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > limitePedidoImportacao {
			http.Redirect(w, r, "/painel/eventos/importar?erro="+string(usecase.MotivoArquivoGrande), http.StatusSeeOther)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, limitePedidoImportacao)
		next.ServeHTTP(w, r)
	})
}

type dadosImportacao struct {
	Etapa     string // enviar, previa ou relatorio
	Erro      *alerta
	Resultado usecase.ResultadoImportacao
	Conteudo  string // o arquivo em base64, só na prévia com linhas aceitas (não fica no servidor)
}

// GET /painel/eventos/importar — enviar o arquivo; ?erro= escolhe o aviso por código.
func (p *painel) telaImportar(w http.ResponseWriter, r *http.Request) {
	d := dadosImportacao{Etapa: "enviar"}
	if motivo := usecase.MotivoArquivo(r.URL.Query().Get("erro")); motivo != "" {
		if titulo, texto, ok := mensagemArquivo(motivo); ok {
			d.Erro = &alerta{Tom: "critical", Titulo: titulo, Texto: texto, Teste: "arquivo-recusado"}
		}
	}
	p.renderImportar(w, r, http.StatusOK, d)
}

// POST /painel/eventos/importar — prévia: lê o arquivo e aplica RN-16 sem gravar nada.
func (p *painel) previaImportacao(w http.ResponseWriter, r *http.Request) {
	var conteudo []byte
	if arquivo, _, err := r.FormFile("arquivo"); err == nil {
		conteudo, err = io.ReadAll(io.LimitReader(arquivo, usecase.LimiteBytesImportacao+1))
		_ = arquivo.Close()
		if err != nil {
			p.falha(w, r, err, "")
			return
		}
	}
	p.importar(w, r, conteudo, false)
}

// POST /painel/eventos/importar/confirmar — aplica RN-16 de novo e grava; mostra o relatório
// direto (sem PRG: reenviar só gera duplicados).
func (p *painel) confirmarImportacao(w http.ResponseWriter, r *http.Request) {
	conteudo, err := base64.StdEncoding.DecodeString(r.PostFormValue("conteudo"))
	if err != nil {
		conteudo = nil // vira arquivo vazio
	}
	p.importar(w, r, conteudo, true)
}

func (p *painel) importar(w http.ResponseWriter, r *http.Request, conteudo []byte, gravar bool) {
	d := dadosImportacao{Etapa: "previa"}
	entradas, err := usecase.LerCSVEventos(conteudo)
	if err == nil {
		if gravar {
			d.Etapa = "relatorio"
			d.Resultado, err = p.d.Importar.Importar(r.Context(), entradas)
		} else {
			d.Resultado, err = p.d.Importar.Previa(r.Context(), entradas)
		}
	}
	var recusa *usecase.ErroArquivo
	switch {
	case errors.As(err, &recusa):
		titulo, texto, _ := mensagemArquivo(recusa.Motivo)
		d = dadosImportacao{Etapa: "enviar", Erro: &alerta{Tom: "critical", Titulo: titulo, Texto: texto, Agora: true, Teste: "arquivo-recusado"}}
		p.renderImportar(w, r, http.StatusUnprocessableEntity, d)
		return
	case err != nil:
		p.falha(w, r, err, "")
		return
	}
	if gravar {
		// Auditoria do lote: operador e quantidades, nunca IDs de cliente (C-09, RN-16).
		p.acao(r, "importar_eventos", "", slog.Int("importados", d.Resultado.Aceitas),
			slog.Int("duplicados", d.Resultado.Duplicadas), slog.Int("rejeitados", d.Resultado.Rejeitadas))
	} else if d.Resultado.Aceitas > 0 {
		d.Conteudo = base64.StdEncoding.EncodeToString(conteudo)
	}
	p.renderImportar(w, r, http.StatusOK, d)
}

func (p *painel) renderImportar(w http.ResponseWriter, r *http.Request, status int, d dadosImportacao) {
	if r.Method == http.MethodPost {
		w.Header().Set("HX-Push-Url", "/painel/eventos/importar") // recarregar volta ao envio, sem reenviar
	}
	p.render(w, r, status, "importar", pagina{Titulo: "Importar eventos", Nav: "importar", Dados: d})
}

// GET /painel/eventos/modelo.csv — o cabeçalho de RN-16 e duas linhas de exemplo sintéticas.
func (p *painel) modeloCSV(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="modelo-eventos.csv"`)
	_, _ = io.WriteString(w, usecase.ColunasImportacao+"\n"+
		"00000000-0000-0000-0000-000000000000,compra,loja on-line,2026-11-10T14:30:00-03:00,150.90,livros\n"+
		"00000000-0000-0000-0000-000000000000,navegacao,app,2026-11-10T14:35:00-03:00,0,\n")
}
