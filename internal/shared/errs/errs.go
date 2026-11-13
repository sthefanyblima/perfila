// Package errs contém os erros de domínio (sentinelas) do Perfila.
//
// Spec: 3.5 (catálogo de erros), C-06 (erros por valor).
package errs

import "errors"

// Erro é um erro de domínio com código estável, usado pela API para montar
// o corpo {codigo, mensagem}. Sentinelas são comparadas com errors.Is.
type Erro struct {
	Codigo   string
	Mensagem string
}

func (e *Erro) Error() string { return e.Codigo + ": " + e.Mensagem }

// Sentinelas do catálogo (Spec 3.5).
var (
	ErrClienteNaoEncontrado       = &Erro{"cliente_nao_encontrado", "cliente inexistente"}
	ErrClienteInativo             = &Erro{"cliente_inativo", "cliente desativado"}
	ErrConsentimentoInvalido      = &Erro{"consentimento_invalido", "consentimento diferente de CONCEDIDO"}
	ErrTransicaoInvalida          = &Erro{"transicao_invalida", "transição de estado não permitida"}
	ErrDadosInvalidos             = &Erro{"dados_invalidos", "argumento, evento ou critério inválido"}
	ErrPerfilNaoEncontrado        = &Erro{"perfil_nao_encontrado", "perfil ausente ou ainda sem segmento"}
	ErrSemProdutoElegivel         = &Erro{"sem_produto_elegivel", "nenhum produto elegível para recomendação"}
	ErrConsentimentoNaoEncontrado = &Erro{"consentimento_nao_encontrado", "registro de consentimento inexistente"}

	// Operadores do painel (v1.5, Spec 3.5, RN-11, RN-12).
	ErrCredenciaisInvalidas  = &Erro{"nao_autenticado", "login ou senha incorretos, ou operador inativo"}
	ErrSemPermissao          = &Erro{"sem_permissao", "o papel do operador não tem esta permissão"}
	ErrOperadorNaoEncontrado = &Erro{"operador_nao_encontrado", "operador inexistente"}
	ErrLoginEmUso            = &Erro{"login_em_uso", "login já usado por outro operador"}
	ErrPropriaConta          = &Erro{"propria_conta", "o administrador não muda o próprio papel nem se desativa"}
)

// Codigo devolve o código do erro de domínio contido em err, ou "" se não houver.
func Codigo(err error) string {
	var e *Erro
	if errors.As(err, &e) {
		return e.Codigo
	}
	return ""
}

// Configuração de segmentos e catálogo (v1.6, Spec 3.5, RN-14).
var (
	ErrSegmentoNaoEncontrado = &Erro{"segmento_nao_encontrado", "segmento inexistente"}
	ErrProdutoNaoEncontrado  = &Erro{"produto_nao_encontrado", "produto inexistente"}
)

// ErroCampo é um erro de validação de um campo de formulário, com a mensagem em português para a
// tela. Embrulha ErrDadosInvalidos: o código continua dados_invalidos. Spec: 3.5, RN-14.
type ErroCampo struct {
	Campo    string
	Mensagem string
}

func (e *ErroCampo) Error() string {
	return ErrDadosInvalidos.Codigo + ": " + e.Campo + ": " + e.Mensagem
}

// Unwrap liga o erro à sentinela ErrDadosInvalidos (C-06).
func (e *ErroCampo) Unwrap() error { return ErrDadosInvalidos }

// NovoErroCampo monta um ErroCampo.
func NovoErroCampo(campo, mensagem string) error { return &ErroCampo{Campo: campo, Mensagem: mensagem} }
