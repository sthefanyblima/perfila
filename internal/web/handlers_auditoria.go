package web

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"perfila/internal/application/usecase"
	"perfila/internal/domain/entity"
)

// Tela Auditoria (Spec 3.7.1 v1.5, RN-13, Q-39): filtros por operador, ação e período, 50 por
// página, do mais recente para o mais antigo. Só ADMINISTRADOR (permissão "auditoria").
func (p *painel) rotasAuditoria(mux *http.ServeMux, sessao func(entity.Permissao, http.HandlerFunc) http.Handler) {
	mux.Handle("GET /painel/auditoria", sessao(entity.PermAuditoria, p.telaAuditoria))
}

type dadosAuditoria struct {
	Pagina            usecase.PaginaAuditoria
	Operador, De, Ate string
	Operadores, Acoes []opcao
	FiltroAtivo       bool
	ErroPeriodo       string
	Inicio, Fim       int
	Anterior, Proxima string
}

// lerDia interpreta "aaaa-mm-dd" como o início do dia em São Paulo; vazio = sem limite.
func lerDia(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, true
	}
	t, err := time.ParseInLocation("2006-01-02", s, fusoSaoPaulo())
	return t.UTC(), err == nil
}

func (p *painel) telaAuditoria(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	d := dadosAuditoria{Operador: strings.TrimSpace(q.Get("operador")), De: q.Get("de"), Ate: q.Get("ate")}
	f := usecase.FiltroListagemAuditoria{Operador: d.Operador, Acao: entity.AcaoAuditoria(q.Get("acao"))}
	if !f.Acao.Valida() {
		f.Acao = ""
	}
	de, okDe := lerDia(d.De)
	ate, okAte := lerDia(d.Ate)
	switch {
	case !okDe || !okAte:
		d.ErroPeriodo = "Use datas no formato do calendário."
	case !de.IsZero() && !ate.IsZero() && ate.Before(de):
		d.ErroPeriodo = "A data final vem antes da inicial."
	default:
		f.De = de
		if !ate.IsZero() {
			f.Ate = ate.AddDate(0, 0, 1) // o dia final inteiro
		}
	}
	f.Pagina, _ = strconv.Atoi(q.Get("pagina"))
	pg, err := p.d.ListarAuditoria.Executar(r.Context(), operadorAtual(r), f)
	if err != nil {
		p.falha(w, r, err, "")
		return
	}
	d.Pagina = pg
	d.FiltroAtivo = d.Operador != "" || f.Acao != "" || d.De != "" || d.Ate != ""
	d.Acoes = []opcao{{Valor: "", Rotulo: "Todas", Atual: f.Acao == ""}}
	for _, a := range entity.AcoesAuditoria() {
		d.Acoes = append(d.Acoes, opcao{Valor: string(a), Rotulo: rotuloAcao(a), Atual: a == f.Acao})
	}
	d.Operadores = []opcao{{Valor: "", Rotulo: "Todos", Atual: d.Operador == ""}}
	if ops, err := p.d.Operadores.Listar(r.Context(), operadorAtual(r)); err == nil {
		for _, o := range ops {
			d.Operadores = append(d.Operadores, opcao{Valor: o.Login, Rotulo: o.Login, Atual: o.Login == entity.NormalizarLogin(d.Operador)})
		}
	}
	if len(pg.Registros) > 0 {
		d.Inicio = (pg.Pagina-1)*usecase.PorPaginaAuditoria + 1
		d.Fim = d.Inicio + len(pg.Registros) - 1
	}
	link := func(n int) string {
		v := url.Values{}
		for k, x := range map[string]string{"operador": d.Operador, "acao": string(f.Acao), "de": d.De, "ate": d.Ate} {
			if x != "" {
				v.Set(k, x)
			}
		}
		v.Set("pagina", strconv.Itoa(n))
		return "/painel/auditoria?" + v.Encode()
	}
	if pg.Pagina > 1 {
		d.Anterior = link(pg.Pagina - 1)
	}
	if pg.Pagina*usecase.PorPaginaAuditoria < pg.Total {
		d.Proxima = link(pg.Pagina + 1)
	}
	p.render(w, r, http.StatusOK, "auditoria", pagina{Titulo: "Auditoria", Nav: "auditoria", Dados: d})
}

// rotuloAcao é o nome legível de cada ação da RN-13.
func rotuloAcao(a entity.AcaoAuditoria) string {
	switch a {
	case entity.AcaoEntrar:
		return "Entrar"
	case entity.AcaoSair:
		return "Sair"
	case entity.AcaoSessaoEncerrada:
		return "Sessão encerrada"
	case entity.AcaoAcessoNegado:
		return "Acesso negado"
	case entity.AcaoLerFicha:
		return "Abrir ficha"
	case entity.AcaoCadastrarCliente:
		return "Cadastrar cliente"
	case entity.AcaoConcederConsentimento:
		return "Conceder consentimento"
	case entity.AcaoRevogarConsentimento:
		return "Revogar consentimento"
	case entity.AcaoRegistrarEvento:
		return "Registrar evento"
	case entity.AcaoProcessar:
		return "Processar"
	case entity.AcaoExpirarRecomendacoes:
		return "Expirar recomendações"
	case entity.AcaoCriarOperador:
		return "Criar operador"
	case entity.AcaoMudarPapel:
		return "Mudar papel"
	case entity.AcaoDesativarOperador:
		return "Desativar operador"
	case entity.AcaoAtivarOperador:
		return "Ativar operador"
	case entity.AcaoRedefinirSenha:
		return "Redefinir senha"
	case entity.AcaoTrocarSenha:
		return "Trocar senha"
	}
	return string(a)
}
