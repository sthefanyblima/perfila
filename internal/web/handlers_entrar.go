package web

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/shared/errs"
)

type dadosEntrar struct {
	Usuario   string
	Volta     string
	Token     string // CSRF do formulário Entrar (cookie perfila_entrar)
	Erro      string
	Expirou   bool
	Encerrada bool // sessão recusada por operador inativo ou senha trocada (Q-38)
}

// GET /painel/entrar — formulário único; guarda o destino e avisa sessão vencida ou encerrada.
func (p *painel) telaEntrar(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	d := dadosEntrar{Volta: q.Get("volta"), Expirou: q.Get("expirou") == "1", Encerrada: q.Get("encerrada") == "1"}
	p.mostrarEntrar(w, r, http.StatusOK, d)
}

func (p *painel) mostrarEntrar(w http.ResponseWriter, r *http.Request, status int, d dadosEntrar) {
	if !destinoSeguro(d.Volta) {
		d.Volta = ""
	}
	d.Token = tokenAleatorio()
	gravarCookie(w, cookieEntrar, d.Token, p.d.Clock.Agora().Add(time.Hour))
	p.render(w, r, status, "entrar", pagina{Titulo: "Entrar", Dados: d})
}

// POST /painel/entrar — AutenticarOperador (RN-12), erro genérico e 5 tentativas por IP em
// 5 minutos (Q-22). Cada tentativa vai para a auditoria (RN-13).
func (p *painel) entrar(w http.ResponseWriter, r *http.Request) {
	usuario, senha := r.PostFormValue("usuario"), r.PostFormValue("senha")
	d := dadosEntrar{Usuario: usuario, Volta: r.PostFormValue("volta")}
	c, err := r.Cookie(cookieEntrar)
	if err != nil || subtle.ConstantTimeCompare([]byte(c.Value), []byte(r.PostFormValue("_csrf"))) != 1 {
		d.Erro = "O formulário expirou. Tente entrar de novo."
		p.mostrarEntrar(w, r, http.StatusForbidden, d)
		return
	}
	ip := ipDe(r)
	if p.tentativas.bloqueado(ip) {
		d.Erro = "Muitas tentativas. Aguarde 5 minutos."
		p.mostrarEntrar(w, r, http.StatusTooManyRequests, d)
		return
	}
	op, err := p.d.Autenticar.Executar(r.Context(), usuario, senha)
	if err != nil {
		if !errors.Is(err, errs.ErrCredenciaisInvalidas) {
			p.falha(w, r, err, "")
			return
		}
		p.tentativas.falhou(ip)
		// Só o login de um operador existente vai para a auditoria: o texto digitado pode ser qualquer coisa.
		ra := r.WithContext(comSessao(r.Context(), sessao{Operador: op.Login}, entity.Operador{}))
		p.auditar(ra, entity.AcaoEntrar, entity.AlvoNenhum, uuid.Nil, err)
		d.Erro = "Usuário ou senha incorretos."
		p.mostrarEntrar(w, r, http.StatusUnauthorized, d)
		return
	}
	s := novaSessao(op, p.d.Clock.Agora().Add(duracaoSessao))
	gravarCookie(w, cookieSessao, assinar(p.d.ChaveSessao, s), s.Expira)
	gravarCookie(w, cookieEntrar, "", time.Unix(1, 0))
	r = r.WithContext(comSessao(r.Context(), s, op))
	p.acao(r, "entrar", "")
	p.auditar(r, entity.AcaoEntrar, entity.AlvoNenhum, uuid.Nil, nil)
	destino := d.Volta
	switch {
	case op.TrocarSenha:
		destino = "/painel/minha-senha"
	case !destinoSeguro(destino):
		destino = inicio(op)
	}
	http.Redirect(w, r, destino, http.StatusSeeOther)
}

// POST /painel/sair — apaga o cookie (a sessão assinada perde a validade no navegador).
func (p *painel) sair(w http.ResponseWriter, r *http.Request) {
	p.acao(r, "sair", "")
	p.auditar(r, entity.AcaoSair, entity.AlvoNenhum, uuid.Nil, nil)
	gravarCookie(w, cookieSessao, "", time.Unix(1, 0))
	http.Redirect(w, r, "/painel/entrar", http.StatusSeeOther)
}
