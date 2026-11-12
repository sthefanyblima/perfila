package web

import (
	"crypto/subtle"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type dadosEntrar struct {
	Usuario string
	Volta   string
	Token   string // CSRF do formulário Entrar (cookie perfila_entrar)
	Erro    string
	Expirou bool
}

// GET /painel/entrar — formulário único; guarda o destino e avisa sessão vencida (fluxo E).
func (p *painel) telaEntrar(w http.ResponseWriter, r *http.Request) {
	d := dadosEntrar{Volta: r.URL.Query().Get("volta"), Expirou: r.URL.Query().Get("expirou") == "1"}
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

// POST /painel/entrar — bcrypt, erro genérico e 5 tentativas por IP em 5 minutos (Q-22).
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
	usuarioOK := subtle.ConstantTimeCompare([]byte(usuario), []byte(p.d.Usuario)) == 1
	senhaOK := bcrypt.CompareHashAndPassword(p.d.SenhaHash, []byte(senha)) == nil
	if !usuarioOK || !senhaOK {
		p.tentativas.falhou(ip)
		d.Erro = "Usuário ou senha incorretos."
		p.mostrarEntrar(w, r, http.StatusUnauthorized, d)
		return
	}
	s := sessao{Operador: p.d.Usuario, Expira: p.d.Clock.Agora().Add(duracaoSessao), CSRF: tokenAleatorio()}
	gravarCookie(w, cookieSessao, assinar(p.d.ChaveSessao, s), s.Expira)
	gravarCookie(w, cookieEntrar, "", time.Unix(1, 0))
	r = r.WithContext(comSessao(r.Context(), s))
	p.acao(r, "entrar", "")
	destino := d.Volta
	if !destinoSeguro(destino) {
		destino = "/painel/clientes"
	}
	http.Redirect(w, r, destino, http.StatusSeeOther)
}

// POST /painel/sair — apaga o cookie (a sessão assinada perde a validade no navegador).
func (p *painel) sair(w http.ResponseWriter, r *http.Request) {
	p.acao(r, "sair", "")
	gravarCookie(w, cookieSessao, "", time.Unix(1, 0))
	http.Redirect(w, r, "/painel/entrar", http.StatusSeeOther)
}
