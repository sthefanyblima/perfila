package web

import (
	"errors"
	"net/http"

	"perfila/internal/application/usecase"
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
	case errs.ErrArquivoIndisponivel.Codigo:
		return "Arquivo indisponível", "Este arquivo não está mais disponível. Gere um novo, se precisar."
	case errs.ErrDadosInvalidos.Codigo:
		switch contexto {
		case "publico-exportar":
			return "Público vazio", "Nenhum cliente entra neste público. Nada foi gerado."
		case "publico-remocoes":
			return "Sem remoção pendente", "Não há remoção pendente neste segmento. Nada foi gerado."
		}
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
	case errs.ErrClienteInativo.Codigo, errs.ErrClienteEliminado.Codigo, errs.ErrTransicaoInvalida.Codigo, errs.ErrPerfilNaoEncontrado.Codigo,
		errs.ErrLoginEmUso.Codigo, errs.ErrPropriaConta.Codigo:
		return http.StatusConflict
	case errs.ErrConsentimentoInvalido.Codigo, errs.ErrSemPermissao.Codigo:
		return http.StatusForbidden
	case errs.ErrCredenciaisInvalidas.Codigo:
		return http.StatusUnauthorized
	case errs.ErrDadosInvalidos.Codigo:
		return http.StatusUnprocessableEntity
	case errs.ErrArquivoIndisponivel.Codigo:
		return http.StatusGone
	}
	return http.StatusInternalServerError
}

// mensagemLinha é o motivo de uma linha não aceita na importação (03-textos.md, RN-16).
func mensagemLinha(l usecase.LinhaImportacao) string {
	if l.Situacao == usecase.SituacaoDuplicada {
		return "Já registrado: mesmo cliente, tipo, data e valor."
	}
	switch errs.Codigo(l.Erro) {
	case errs.ErrClienteNaoEncontrado.Codigo:
		return "Cliente não encontrado. Confira o ID."
	case errs.ErrClienteInativo.Codigo:
		return "Cadastro inativo. Eventos ficam bloqueados."
	case errs.ErrConsentimentoInvalido.Codigo:
		return "Sem consentimento válido. O evento não pode ser usado."
	}
	switch l.Motivo {
	case usecase.MotivoClienteID:
		return "Informe o ID do cliente no formato UUID."
	case usecase.MotivoTipo:
		return "Use compra, navegacao ou interacao."
	case usecase.MotivoOrigem:
		return "Informe de onde veio o evento."
	case usecase.MotivoTimestamp:
		return "Informe data, hora e fuso, por exemplo 2026-11-10T14:30:00-03:00."
	case usecase.MotivoTimestampFuturo:
		return "A data não pode estar no futuro."
	case usecase.MotivoValor:
		return "O valor deve estar entre 0 e 1000000, com ponto decimal."
	case usecase.MotivoValorCompra:
		return "Uma compra precisa ter valor maior que zero."
	case usecase.MotivoColunas:
		return "A linha não tem as mesmas colunas do cabeçalho."
	}
	_, texto := mensagem(l.Erro, "")
	return texto
}

// mensagemArquivo explica por que o arquivo foi recusado inteiro (03-textos.md, RN-16).
func mensagemArquivo(m usecase.MotivoArquivo) (titulo, texto string, ok bool) {
	textos := map[usecase.MotivoArquivo]string{
		usecase.MotivoArquivoVazio:        "O arquivo não tem nenhuma linha de evento. Confira se ele tem o cabeçalho e pelo menos uma linha.",
		usecase.MotivoArquivoGrande:       "O arquivo passa de 256 KiB. Divida em arquivos menores.",
		usecase.MotivoLinhasDemais:        "O arquivo passa de 1.000 linhas. Divida em arquivos menores.",
		usecase.MotivoCodificacaoInvalida: "O arquivo não está em UTF-8. Salve de novo como \"CSV UTF-8\".",
		usecase.MotivoCabecalhoInvalido:   "O cabeçalho deve ter as colunas cliente_id, tipo, origem, timestamp e valor (categoria é opcional), separadas por vírgula.",
		usecase.MotivoFormatoInvalido:     "O arquivo tem aspas mal fechadas. Confira o arquivo e envie de novo.",
	}
	texto, ok = textos[m]
	return "Arquivo recusado", texto, ok
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
	case "arquivo-gerado":
		return "Arquivo gerado"
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
