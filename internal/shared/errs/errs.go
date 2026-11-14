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
	ErrClienteEliminado           = &Erro{"cliente_eliminado", "dados do titular eliminados"} // v1.9 RN-17
	ErrConsentimentoInvalido      = &Erro{"consentimento_invalido", "consentimento diferente de CONCEDIDO"}
	ErrTransicaoInvalida          = &Erro{"transicao_invalida", "transição de estado não permitida"}
	ErrDadosInvalidos             = &Erro{"dados_invalidos", "argumento, evento ou critério inválido"}
	ErrPerfilNaoEncontrado        = &Erro{"perfil_nao_encontrado", "perfil ausente ou ainda sem segmento"}
	ErrSemProdutoElegivel         = &Erro{"sem_produto_elegivel", "nenhum produto elegível para recomendação"}
	ErrConsentimentoNaoEncontrado = &Erro{"consentimento_nao_encontrado", "registro de consentimento inexistente"}
	ErrAcessoInvalido             = &Erro{"acesso_invalido", "link de acesso inexistente, vencido ou já usado"} // v1.10 RN-19
	ErrPedidoNaoEncontrado        = &Erro{"pedido_nao_encontrado", "nenhum pedido de eliminação aberto"}        // v1.10, uso interno
)

// Codigo devolve o código do erro de domínio contido em err, ou "" se não houver.
func Codigo(err error) string {
	var e *Erro
	if errors.As(err, &e) {
		return e.Codigo
	}
	return ""
}
