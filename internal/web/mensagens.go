package web

import (
	"errors"
	"net/http"

	"perfila/internal/shared/errs"
)

// mensagem traduz o código estável do erro (errs.Codigo) no texto da interface
// (docs/design/03-textos.md). O texto técnico do servidor nunca aparece.
func mensagem(err error, contexto string) (titulo, texto string) {
	switch errs.Codigo(err) {
	case errs.ErrClienteNaoEncontrado.Codigo:
		return "Cliente não encontrado", "Não encontramos esse cliente. Confira o ID ou busque pelo nome."
	case errs.ErrClienteEliminado.Codigo:
		return "Dados eliminados", "Os dados deste titular foram eliminados. Nada mais pode ser alterado."
	case errs.ErrClienteInativo.Codigo:
		return "Cadastro inativo", "Este cadastro está inativo. Eventos e processamento ficam bloqueados enquanto ele estiver assim."
	case errs.ErrConsentimentoInvalido.Codigo:
		return "Sem consentimento válido", "Registre o consentimento do titular para usar os dados de comportamento."
	case errs.ErrTransicaoInvalida.Codigo:
		if contexto == "eliminar" {
			return "Consentimento concedido", "Revogue o consentimento antes de eliminar os dados."
		}
		if contexto == "revogar" {
			return "Nada para revogar", "Não há consentimento concedido para revogar."
		}
		return "Consentimento já concedido", "O consentimento já está concedido. Para mudar a versão do termo, revogue e conceda de novo."
	case errs.ErrPerfilNaoEncontrado.Codigo:
		return "Ainda não processado", "Este cliente ainda não foi processado. Use \"Processar agora\"."
	case errs.ErrDadosInvalidos.Codigo:
		return "Dados inválidos", "Confira os campos destacados e tente de novo."
	}
	return "Algo falhou do nosso lado", "As etapas concluídas antes da falha ficaram salvas. Tente de novo; se continuar, informe o código da requisição à equipe técnica."
}

// statusDe repete o mapeamento da Spec 3.5 para as respostas HTML.
func statusDe(err error) int {
	var e *errs.Erro
	if !errors.As(err, &e) {
		return http.StatusInternalServerError
	}
	switch e.Codigo {
	case errs.ErrClienteNaoEncontrado.Codigo:
		return http.StatusNotFound
	case errs.ErrClienteInativo.Codigo, errs.ErrClienteEliminado.Codigo, errs.ErrTransicaoInvalida.Codigo, errs.ErrPerfilNaoEncontrado.Codigo:
		return http.StatusConflict
	case errs.ErrConsentimentoInvalido.Codigo:
		return http.StatusForbidden
	case errs.ErrDadosInvalidos.Codigo:
		return http.StatusUnprocessableEntity
	}
	return http.StatusInternalServerError
}

// avisos são os toasts de confirmação, escolhidos por código na URL de redirecionamento (PRG):
// o texto nunca vem da URL.
func aviso(codigo, n string) string {
	switch codigo {
	case "cliente-cadastrado":
		return "Cliente cadastrado"
	case "consentimento-concedido":
		return "Consentimento concedido"
	case "consentimento-revogado":
		return "Consentimento revogado"
	case "evento-registrado":
		return "Evento registrado"
	case "processamento-concluido":
		return "Processamento concluído"
	case "processamento-sem-recomendacao":
		return "Processamento concluído sem recomendação"
	case "expiradas":
		switch n {
		case "", "0":
			return "Nenhuma recomendação vencida"
		case "1":
			return "1 recomendação expirada"
		}
		return n + " recomendações expiradas"
	case "cadastro-corrigido":
		return "Cadastro corrigido"
	case "dados-eliminados":
		return "Dados eliminados"
	case "retencao":
		switch n {
		case "", "0":
			return "Nenhum cliente com a retenção vencida"
		case "1":
			return "Retenção aplicada a 1 cliente"
		}
		return "Retenção aplicada a " + n + " clientes"
	}
	return ""
}
