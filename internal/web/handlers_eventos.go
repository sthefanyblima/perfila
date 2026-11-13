package web

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"perfila/internal/domain/entity"
	"perfila/internal/shared/errs"
)

// GET /painel/clientes/{id}/eventos — formulário em linha e linha do tempo (UC-01).
func (p *painel) fichaEventos(w http.ResponseWriter, r *http.Request) {
	d, ok := p.carregarFicha(w, r, "eventos")
	if !ok {
		return
	}
	d.Evento = formEvento{Quando: valorDataLocal(p.d.Clock.Agora())}
	p.renderFicha(w, r, http.StatusOK, d)
}

// POST /painel/clientes/{id}/eventos — confere o formato no painel (retorno por campo, textos de
// 03-textos.md) e deixa a decisão para UC-01 Registrar, que aplica RN-01 e RN-04. Spec: 3.7.1.
func (p *painel) registrarEvento(w http.ResponseWriter, r *http.Request) {
	id, ok := idDaRota(r)
	if !ok {
		p.falha(w, r, errs.ErrClienteNaoEncontrado, "")
		return
	}
	f := formEvento{Tipo: r.PostFormValue("tipo"), Valor: r.PostFormValue("valor"), Categoria: strings.TrimSpace(r.PostFormValue("categoria")),
		Origem: strings.TrimSpace(r.PostFormValue("origem")), Quando: r.PostFormValue("quando"), Erros: map[string]string{}}
	ev := entity.EventoComportamental{ClienteID: id, Tipo: entity.TipoEvento(f.Tipo), Origem: f.Origem, Categoria: f.Categoria}
	if !ev.Tipo.Valido() {
		f.Erros["tipo"] = "Escolha Compra, Navegação ou Interação."
	}
	if f.Origem == "" {
		f.Erros["origem"] = "Informe de onde veio o evento (por exemplo, loja on-line)."
	}
	valor, err := lerDecimal(f.Valor)
	switch {
	case strings.TrimSpace(f.Valor) == "" && ev.Tipo != entity.TipoCompra:
		valor = 0
	case err != nil || valor < 0 || valor > entity.ValorMaximoEvento:
		f.Erros["valor"] = "O valor deve estar entre R$ 0,00 e R$ 1.000.000,00."
	case ev.Tipo == entity.TipoCompra && valor <= 0:
		f.Erros["valor"] = "Uma compra precisa ter valor maior que zero."
	}
	ev.Valor = valor
	quando, err := lerDataLocal(f.Quando)
	switch {
	case err != nil:
		f.Erros["quando"] = "Informe a data e a hora do evento."
	case quando.After(p.d.Clock.Agora().Add(entity.ToleranciaRelogio)):
		f.Erros["quando"] = "A data não pode estar no futuro."
	}
	ev.Timestamp = quando

	if len(f.Erros) == 0 {
		if _, err = p.d.Coletar.Registrar(r.Context(), ev); err == nil {
			p.acao(r, "registrar_evento", id.String())
			http.Redirect(w, r, fmt.Sprintf("/painel/clientes/%s/eventos?aviso=evento-registrado", id), http.StatusSeeOther)
			return
		}
		p.logarFalha(r, err)
	}
	d, ok := p.carregarFicha(w, r, "eventos")
	if !ok {
		return
	}
	d.Evento = f
	status := http.StatusUnprocessableEntity
	if err != nil && len(f.Erros) == 0 {
		status = statusDe(err)
		if d.Valido && !errors.Is(err, errs.ErrConsentimentoInvalido) {
			titulo, texto := mensagem(err, "")
			d.Falha = &alerta{Tom: "critical", Titulo: titulo, Texto: texto, Agora: true, Teste: "falha-evento"}
		}
	}
	p.renderFicha(w, r, status, d)
}

// GET /painel/clientes/{id}/consentimento — o ConsentRecord completo (UC-08), sempre disponível.
func (p *painel) fichaConsentimento(w http.ResponseWriter, r *http.Request) {
	if d, ok := p.carregarFicha(w, r, "consentimento"); ok {
		p.renderFicha(w, r, http.StatusOK, d)
	}
}

func acaoValida(a string) bool { return a == "conceder" || a == "revogar" }

// GET /painel/clientes/{id}/consentimento/{acao} — confirmação sem JavaScript (o diálogo, com JS).
func (p *painel) confirmarConsentimento(w http.ResponseWriter, r *http.Request) {
	acao := r.PathValue("acao")
	if !acaoValida(acao) {
		p.naoEncontrada(w, r)
		return
	}
	d, ok := p.carregarFicha(w, r, "consentimento")
	if !ok {
		return
	}
	d.Confirmar = acao
	p.renderFicha(w, r, http.StatusOK, d)
}

// POST /painel/clientes/{id}/consentimento/{acao} — UC-08 Conceder ou Revogar, depois PRG para a
// ficha com o selo novo. Transição recusada pelo domínio volta à confirmação com o motivo.
func (p *painel) alterarConsentimento(w http.ResponseWriter, r *http.Request) {
	acao := r.PathValue("acao")
	id, ok := idDaRota(r)
	if !ok || !acaoValida(acao) {
		p.naoEncontrada(w, r)
		return
	}
	f := formConsentimento{Finalidade: strings.TrimSpace(r.PostFormValue("finalidade")), VersaoTermo: strings.TrimSpace(r.PostFormValue("versaoTermo"))}
	var err error
	if acao == "conceder" {
		if f.Finalidade == "" {
			f.ErroFinalidade = "Informe para que os dados serão usados."
		}
		if f.VersaoTermo == "" {
			f.ErroVersao = "Informe a versão do termo que o titular aceitou."
		}
		if f.ErroFinalidade == "" && f.ErroVersao == "" {
			err = p.d.Consentimento.Conceder(r.Context(), id, f.Finalidade, f.VersaoTermo)
		}
	} else {
		err = p.d.Consentimento.Revogar(r.Context(), id)
	}
	if err == nil && f.ErroFinalidade == "" && f.ErroVersao == "" {
		p.acao(r, acao+"_consentimento", id.String())
		aviso := map[string]string{"conceder": "consentimento-concedido", "revogar": "consentimento-revogado"}[acao]
		http.Redirect(w, r, fmt.Sprintf("/painel/clientes/%s?aviso=%s", id, aviso), http.StatusSeeOther)
		return
	}
	d, ok := p.carregarFicha(w, r, "consentimento")
	if !ok {
		return
	}
	d.Confirmar, d.Concessao = acao, f
	status := http.StatusUnprocessableEntity
	if err != nil {
		p.logarFalha(r, err)
		status = statusDe(err)
		titulo, texto := mensagem(err, acao)
		d.ErroConfirmar = &alerta{Tom: "critical", Titulo: titulo, Texto: texto, Agora: true}
	}
	p.renderFicha(w, r, status, d)
}
