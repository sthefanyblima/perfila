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
	case errs.ErrClienteInativo.Codigo:
		return "Cadastro inativo", "Este cadastro está inativo. Eventos e processamento ficam bloqueados enquanto ele estiver assim."
	case errs.ErrConsentimentoInvalido.Codigo:
		return "Sem consentimento válido", "Registre o consentimento do titular para usar os dados de comportamento."
	case errs.ErrTransicaoInvalida.Codigo:
		if contexto == "revogar" {
			return "Nada para revogar", "Não há consentimento concedido para revogar."
		}
		return "Consentimento já concedido", "O consentimento já está concedido. Para mudar a versão do termo, revogue e conceda de novo."
	case errs.ErrPerfilNaoEncontrado.Codigo:
		return "Ainda não processado", "Este cliente ainda não foi processado. Use \"Processar agora\"."
	case errs.ErrDadosInvalidos.Codigo:
		return "Dados inválidos", "Confira os campos destacados e tente de novo."
	case errs.ErrSemPermissao.Codigo:
		return "Sem permissão", "Seu perfil de acesso não permite esta ação. Se precisar dela, fale com quem administra o painel."
	case errs.ErrOperadorNaoEncontrado.Codigo:
		return "Operador não encontrado", "Não encontramos esse operador. Volte para a lista de usuários."
	case errs.ErrLoginEmUso.Codigo:
		return "Login em uso", "Já existe um operador com esse login. Escolha outro: logins não se repetem, nem de operadores desativados."
	case errs.ErrPropriaConta.Codigo:
		return "Ação sobre a própria conta", "Você não pode mudar o próprio papel nem se desativar. Peça a outra pessoa administradora."
	case errs.ErrCredenciaisInvalidas.Codigo:
		return "Senha atual incorreta", "Confira a senha atual e tente de novo."
	case errs.ErrSegmentoNaoEncontrado.Codigo:
		return "Segmento não encontrado", "Não encontramos esse segmento. Volte para a lista de segmentos."
	case errs.ErrProdutoNaoEncontrado.Codigo:
		return "Produto não encontrado", "Não encontramos esse produto. Volte para o catálogo."
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
	case errs.ErrClienteNaoEncontrado.Codigo, errs.ErrOperadorNaoEncontrado.Codigo, errs.ErrSegmentoNaoEncontrado.Codigo,
		errs.ErrProdutoNaoEncontrado.Codigo:
		return http.StatusNotFound
	case errs.ErrClienteInativo.Codigo, errs.ErrTransicaoInvalida.Codigo, errs.ErrPerfilNaoEncontrado.Codigo,
		errs.ErrLoginEmUso.Codigo, errs.ErrPropriaConta.Codigo:
		return http.StatusConflict
	case errs.ErrConsentimentoInvalido.Codigo, errs.ErrSemPermissao.Codigo:
		return http.StatusForbidden
	case errs.ErrCredenciaisInvalidas.Codigo:
		return http.StatusUnauthorized
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
	case "operador-criado":
		return "Operador criado"
	case "papel-alterado":
		return "Papel alterado"
	case "operador-desativado":
		return "Operador desativado"
	case "operador-ativado":
		return "Operador ativado"
	case "senha-redefinida":
		return "Senha provisória definida"
	case "senha-trocada":
		return "Senha trocada"
	case "segmento-salvo":
		return "Segmento salvo. Os clientes mudam de segmento no próximo processamento."
	case "segmento-ativado":
		return "Segmento ativado"
	case "segmento-desativado":
		return "Segmento desativado"
	case "produto-salvo":
		return "Produto salvo"
	case "produto-ativado":
		return "Produto ativado"
	case "produto-desativado":
		return "Produto desativado"
	case "sem-alteracoes":
		return "Nada mudou"
	case "expiradas":
		switch n {
		case "", "0":
			return "Nenhuma recomendação vencida"
		case "1":
			return "1 recomendação expirada"
		}
		return n + " recomendações expiradas"
	}
	return ""
}
