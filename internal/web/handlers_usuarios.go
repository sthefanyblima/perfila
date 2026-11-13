package web

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/shared/errs"
)

// Telas Usuários e Minha senha (Spec 3.7.1 v1.5, RN-12, Q-35, Q-40). Os handlers só traduzem
// formulários; papel, conta própria e política de senha são decididos por GerenciarOperadores.
func (p *painel) rotasUsuarios(mux *http.ServeMux, sessao func(entity.Permissao, http.HandlerFunc) http.Handler) {
	mux.Handle("GET /painel/minha-senha", sessao("", p.telaMinhaSenha))
	mux.Handle("POST /painel/minha-senha", sessao("", p.trocarSenha))
	mux.Handle("GET /painel/usuarios", sessao(entity.PermUsuarios, p.telaUsuarios))
	mux.Handle("GET /painel/usuarios/novo", sessao(entity.PermUsuarios, p.telaNovoOperador))
	mux.Handle("POST /painel/usuarios", sessao(entity.PermUsuarios, p.criarOperador))
	mux.Handle("GET /painel/usuarios/{id}", sessao(entity.PermUsuarios, p.telaOperador))
	mux.Handle("POST /painel/usuarios/{id}/papel", sessao(entity.PermUsuarios, p.mudarPapel))
	mux.Handle("GET /painel/usuarios/{id}/desativar", sessao(entity.PermUsuarios, p.confirmarDesativar))
	mux.Handle("POST /painel/usuarios/{id}/desativar", sessao(entity.PermUsuarios, p.desativar))
	mux.Handle("POST /painel/usuarios/{id}/ativar", sessao(entity.PermUsuarios, p.ativar))
	mux.Handle("POST /painel/usuarios/{id}/senha", sessao(entity.PermUsuarios, p.redefinirSenha))
}

// papeis é a lista do seletor, na ordem da tela.
func papeis() []entity.PapelOperador {
	return []entity.PapelOperador{entity.PapelAtendimento, entity.PapelAnalista, entity.PapelAdministrador}
}

const erroSenha = "A senha precisa ter pelo menos 12 caracteres."

// senhaCurta repete no painel a regra da RN-12 só para mostrar o erro ao lado do campo; quem
// decide é o caso de uso.
func senhaCurta(s string) bool { return utf8.RuneCountInString(s) < 12 || len(s) > 72 }

// ---------- Usuários ----------

func (p *painel) telaUsuarios(w http.ResponseWriter, r *http.Request) {
	ops, err := p.d.Operadores.Listar(r.Context(), operadorAtual(r))
	if err != nil {
		p.falha(w, r, err, "")
		return
	}
	p.render(w, r, http.StatusOK, "usuarios", pagina{Titulo: "Usuários", Nav: "usuarios", Dados: ops})
}

type dadosNovoOperador struct {
	Login, Nome, Papel             string
	ErroLogin, ErroNome, ErroSenha string
	Papeis                         []entity.PapelOperador
	Erro                           *alerta
}

func (p *painel) telaNovoOperador(w http.ResponseWriter, r *http.Request) {
	p.render(w, r, http.StatusOK, "novo_operador", pagina{Titulo: "Criar operador", Nav: "usuarios",
		Dados: dadosNovoOperador{Papel: string(entity.PapelAtendimento), Papeis: papeis()}})
}

func (p *painel) criarOperador(w http.ResponseWriter, r *http.Request) {
	senha := r.PostFormValue("senha")
	d := dadosNovoOperador{Login: strings.TrimSpace(r.PostFormValue("login")), Nome: strings.TrimSpace(r.PostFormValue("nome")),
		Papel: r.PostFormValue("papel"), Papeis: papeis()}
	if d.Login == "" {
		d.ErroLogin = "Informe o login."
	}
	if d.Nome == "" {
		d.ErroNome = "Informe o nome de quem vai operar."
	}
	if senhaCurta(senha) {
		d.ErroSenha = erroSenha
	}
	if d.ErroLogin == "" && d.ErroNome == "" && d.ErroSenha == "" {
		o, err := p.d.Operadores.Criar(r.Context(), operadorAtual(r), d.Login, d.Nome, entity.PapelOperador(d.Papel), senha)
		alvo := entity.AlvoOperador
		if err != nil {
			alvo = entity.AlvoNenhum
		}
		p.auditar(r, entity.AcaoCriarOperador, alvo, o.ID, err)
		if err == nil {
			http.Redirect(w, r, fmt.Sprintf("/painel/usuarios/%s?aviso=operador-criado", o.ID), http.StatusSeeOther)
			return
		}
		p.logarFalha(r, err)
		switch {
		case errors.Is(err, errs.ErrLoginEmUso):
			d.ErroLogin = "Já existe um operador com esse login. Logins não se repetem, nem de operadores desativados."
		case errors.Is(err, errs.ErrDadosInvalidos):
			d.ErroLogin = "Use de 3 a 32 letras minúsculas sem acento, números, ponto, hífen ou sublinhado, começando por letra."
		default:
			titulo, texto := mensagem(err, "")
			d.Erro = &alerta{Tom: "critical", Titulo: titulo, Texto: texto, Agora: true, RequestID: r.Header.Get("X-Request-ID")}
		}
		p.render(w, r, statusDe(err), "novo_operador", pagina{Titulo: "Criar operador", Nav: "usuarios", Dados: d})
		return
	}
	p.render(w, r, http.StatusUnprocessableEntity, "novo_operador", pagina{Titulo: "Criar operador", Nav: "usuarios", Dados: d})
}

type dadosOperador struct {
	Op        entity.Operador
	Propria   bool
	Papeis    []entity.PapelOperador
	Confirmar bool   // página de confirmação de desativar, sem JavaScript
	ErroSenha string // redefinição
	Erro      *alerta
}

func (p *painel) carregarOperador(w http.ResponseWriter, r *http.Request) (dadosOperador, bool) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		p.falha(w, r, errs.ErrOperadorNaoEncontrado, "")
		return dadosOperador{}, false
	}
	o, err := p.d.Operadores.Consultar(r.Context(), operadorAtual(r), id)
	if err != nil {
		p.falha(w, r, err, "")
		return dadosOperador{}, false
	}
	return dadosOperador{Op: o, Propria: o.ID == operadorAtual(r).ID, Papeis: papeis()}, true
}

func (p *painel) mostrarOperador(w http.ResponseWriter, r *http.Request, status int, d dadosOperador) {
	p.render(w, r, status, "operador", pagina{Titulo: d.Op.Login, Nav: "usuarios", Dados: d})
}

// GET /painel/usuarios/{id} — papel, situação e senha do operador.
func (p *painel) telaOperador(w http.ResponseWriter, r *http.Request) {
	if d, ok := p.carregarOperador(w, r); ok {
		p.mostrarOperador(w, r, http.StatusOK, d)
	}
}

func (p *painel) confirmarDesativar(w http.ResponseWriter, r *http.Request) {
	if d, ok := p.carregarOperador(w, r); ok {
		d.Confirmar = true
		p.mostrarOperador(w, r, http.StatusOK, d)
	}
}

// alterarOperador aplica a mudança, registra na auditoria e volta para a página do operador
// com o toast (PRG); erro de domínio mostra a página com o motivo.
func (p *painel) alterarOperador(w http.ResponseWriter, r *http.Request, acao entity.AcaoAuditoria, aviso string,
	f func(ator entity.Operador, id uuid.UUID) error) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		p.falha(w, r, errs.ErrOperadorNaoEncontrado, "")
		return
	}
	err = f(operadorAtual(r), id)
	p.auditar(r, acao, entity.AlvoOperador, id, err)
	if err == nil {
		http.Redirect(w, r, fmt.Sprintf("/painel/usuarios/%s?aviso=%s", id, aviso), http.StatusSeeOther)
		return
	}
	p.logarFalha(r, err)
	d, ok := p.carregarOperador(w, r)
	if !ok {
		return
	}
	if errors.Is(err, errs.ErrDadosInvalidos) && acao == entity.AcaoRedefinirSenha {
		d.ErroSenha = erroSenha
	} else {
		titulo, texto := mensagem(err, "")
		d.Erro = &alerta{Tom: "critical", Titulo: titulo, Texto: texto, Agora: true, RequestID: r.Header.Get("X-Request-ID")}
	}
	p.mostrarOperador(w, r, statusDe(err), d)
}

func (p *painel) mudarPapel(w http.ResponseWriter, r *http.Request) {
	papel := entity.PapelOperador(r.PostFormValue("papel"))
	p.alterarOperador(w, r, entity.AcaoMudarPapel, "papel-alterado", func(a entity.Operador, id uuid.UUID) error {
		return p.d.Operadores.MudarPapel(r.Context(), a, id, papel)
	})
}

func (p *painel) desativar(w http.ResponseWriter, r *http.Request) {
	p.alterarOperador(w, r, entity.AcaoDesativarOperador, "operador-desativado", func(a entity.Operador, id uuid.UUID) error {
		return p.d.Operadores.Desativar(r.Context(), a, id)
	})
}

func (p *painel) ativar(w http.ResponseWriter, r *http.Request) {
	p.alterarOperador(w, r, entity.AcaoAtivarOperador, "operador-ativado", func(a entity.Operador, id uuid.UUID) error {
		return p.d.Operadores.Ativar(r.Context(), a, id)
	})
}

func (p *painel) redefinirSenha(w http.ResponseWriter, r *http.Request) {
	senha := r.PostFormValue("senha")
	p.alterarOperador(w, r, entity.AcaoRedefinirSenha, "senha-redefinida", func(a entity.Operador, id uuid.UUID) error {
		return p.d.Operadores.RedefinirSenha(r.Context(), a, id, senha)
	})
}

// ---------- Minha senha ----------

type dadosMinhaSenha struct {
	Obrigatoria                        bool // senha provisória: troca antes de continuar (Q-40)
	ErroAtual, ErroNova, ErroConfirmar string
	Erro                               string
}

func (p *painel) telaMinhaSenha(w http.ResponseWriter, r *http.Request) {
	p.render(w, r, http.StatusOK, "minha_senha", pagina{Titulo: "Minha senha", Nav: "minha-senha",
		Dados: dadosMinhaSenha{Obrigatoria: operadorAtual(r).TrocarSenha}})
}

// POST /painel/minha-senha — senha atual, nova e confirmação. A senha atual errada conta no
// limite de tentativas (Spec 3.7.1). No sucesso, a sessão é renovada com o novo carimbo.
func (p *painel) trocarSenha(w http.ResponseWriter, r *http.Request) {
	atual, nova, confirmacao := r.PostFormValue("atual"), r.PostFormValue("nova"), r.PostFormValue("confirmacao")
	ator := operadorAtual(r)
	d := dadosMinhaSenha{Obrigatoria: ator.TrocarSenha}
	mostrar := func(status int) {
		p.render(w, r, status, "minha_senha", pagina{Titulo: "Minha senha", Nav: "minha-senha", Dados: d})
	}
	ip := ipDe(r)
	if p.tentativas.bloqueado(ip) {
		d.Erro = "Muitas tentativas. Aguarde 5 minutos."
		mostrar(http.StatusTooManyRequests)
		return
	}
	switch {
	case senhaCurta(nova):
		d.ErroNova = erroSenha
	case nova != confirmacao:
		d.ErroConfirmar = "A confirmação não é igual à senha nova."
	case nova == atual:
		d.ErroNova = "A senha nova precisa ser diferente da atual."
	}
	if d.ErroNova != "" || d.ErroConfirmar != "" {
		mostrar(http.StatusUnprocessableEntity)
		return
	}
	o, err := p.d.Operadores.TrocarSenha(r.Context(), ator, atual, nova)
	p.auditar(r, entity.AcaoTrocarSenha, entity.AlvoOperador, ator.ID, err)
	if err != nil {
		if errors.Is(err, errs.ErrCredenciaisInvalidas) {
			p.tentativas.falhou(ip)
			d.ErroAtual = "Senha atual incorreta."
			mostrar(http.StatusUnauthorized)
			return
		}
		p.falha(w, r, err, "")
		return
	}
	s := novaSessao(o, p.d.Clock.Agora().Add(duracaoSessao))
	gravarCookie(w, cookieSessao, assinar(p.d.ChaveSessao, s), s.Expira)
	http.Redirect(w, r, inicio(o)+"?aviso=senha-trocada", http.StatusSeeOther)
}
