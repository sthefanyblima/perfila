package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"perfila/internal/shared/errs"
)

// AcaoAuditoria é o que o operador fez no painel. Spec: 3.2, RN-13, Q-39.
type AcaoAuditoria string

const (
	AcaoEntrar                AcaoAuditoria = "entrar"
	AcaoSair                  AcaoAuditoria = "sair"
	AcaoSessaoEncerrada       AcaoAuditoria = "sessao_encerrada"
	AcaoAcessoNegado          AcaoAuditoria = "acesso_negado"
	AcaoLerFicha              AcaoAuditoria = "ler_ficha"
	AcaoCadastrarCliente      AcaoAuditoria = "cadastrar_cliente"
	AcaoConcederConsentimento AcaoAuditoria = "conceder_consentimento"
	AcaoRevogarConsentimento  AcaoAuditoria = "revogar_consentimento"
	AcaoRegistrarEvento       AcaoAuditoria = "registrar_evento"
	AcaoProcessar             AcaoAuditoria = "processar"
	AcaoExpirarRecomendacoes  AcaoAuditoria = "expirar_recomendacoes"
	AcaoCriarOperador         AcaoAuditoria = "criar_operador"
	AcaoMudarPapel            AcaoAuditoria = "mudar_papel"
	AcaoDesativarOperador     AcaoAuditoria = "desativar_operador"
	AcaoAtivarOperador        AcaoAuditoria = "ativar_operador"
	AcaoRedefinirSenha        AcaoAuditoria = "redefinir_senha"
	AcaoTrocarSenha           AcaoAuditoria = "trocar_senha"
	// v1.6 (Q-44): configuração de segmentos e catálogo.
	AcaoCriarSegmento     AcaoAuditoria = "criar_segmento"
	AcaoEditarSegmento    AcaoAuditoria = "editar_segmento"
	AcaoAtivarSegmento    AcaoAuditoria = "ativar_segmento"
	AcaoDesativarSegmento AcaoAuditoria = "desativar_segmento"
	AcaoPreverSegmentacao AcaoAuditoria = "prever_segmentacao"
	AcaoCriarProduto      AcaoAuditoria = "criar_produto"
	AcaoEditarProduto     AcaoAuditoria = "editar_produto"
	AcaoAtivarProduto     AcaoAuditoria = "ativar_produto"
	AcaoDesativarProduto  AcaoAuditoria = "desativar_produto"
)

// CamposAuditaveis é a lista fechada de nomes de campo que a auditoria aceita, na ordem em que
// são gravados. Nunca valores. Spec: RN-13, Q-44.
func CamposAuditaveis() []string {
	return []string{"nome", "descricao", "criterio", "prioridade", "ativo", "categoria", "preco"}
}

// AcoesAuditoria devolve todas as ações, na ordem da RN-13 (filtro da tela Auditoria).
func AcoesAuditoria() []AcaoAuditoria {
	return []AcaoAuditoria{AcaoEntrar, AcaoSair, AcaoSessaoEncerrada, AcaoAcessoNegado, AcaoLerFicha,
		AcaoCadastrarCliente, AcaoConcederConsentimento, AcaoRevogarConsentimento, AcaoRegistrarEvento,
		AcaoProcessar, AcaoExpirarRecomendacoes, AcaoCriarOperador, AcaoMudarPapel, AcaoDesativarOperador,
		AcaoAtivarOperador, AcaoRedefinirSenha, AcaoTrocarSenha, AcaoCriarSegmento, AcaoEditarSegmento,
		AcaoAtivarSegmento, AcaoDesativarSegmento, AcaoPreverSegmentacao, AcaoCriarProduto, AcaoEditarProduto,
		AcaoAtivarProduto, AcaoDesativarProduto}
}

// Valida informa se a ação pertence à lista da RN-13.
func (a AcaoAuditoria) Valida() bool {
	for _, x := range AcoesAuditoria() {
		if a == x {
			return true
		}
	}
	return false
}

// TipoAlvo diz sobre o que a ação foi; vazio = sem alvo. Spec: 3.2, RN-13.
type TipoAlvo string

const (
	AlvoNenhum   TipoAlvo = ""
	AlvoCliente  TipoAlvo = "cliente"
	AlvoSegmento TipoAlvo = "segmento"
	AlvoProduto  TipoAlvo = "produto"
	AlvoOperador TipoAlvo = "operador"
)

// Valido informa se o tipo pertence ao conjunto permitido (vazio incluso).
func (t TipoAlvo) Valido() bool {
	switch t {
	case AlvoNenhum, AlvoCliente, AlvoSegmento, AlvoProduto, AlvoOperador:
		return true
	}
	return false
}

// ResultadoAuditoria diz como a ação terminou. Spec: 3.2, RN-13.
type ResultadoAuditoria string

const (
	ResultadoSucesso ResultadoAuditoria = "SUCESSO"
	ResultadoNegado  ResultadoAuditoria = "NEGADO"
	ResultadoFalha   ResultadoAuditoria = "FALHA"
)

// Valido informa se o resultado pertence ao conjunto permitido.
func (r ResultadoAuditoria) Valido() bool {
	return r == ResultadoSucesso || r == ResultadoNegado || r == ResultadoFalha
}

// RegistroAuditoria é uma linha da trilha do painel: só login, ação, alvo (tipo e ID) e
// resultado, nunca dado pessoal (C-09). Só de inclusão. Spec: 3.2, RN-13, Q-39. Fonte: UML2 (v1.5).
type RegistroAuditoria struct {
	ID        uuid.UUID
	Data      time.Time
	Operador  string // login; vazio em entrada com login desconhecido
	Acao      AcaoAuditoria
	AlvoTipo  TipoAlvo
	AlvoID    uuid.UUID // uuid.Nil quando AlvoTipo é vazio
	Resultado ResultadoAuditoria
	Campos    []string // v1.6 (Q-44): nomes dos campos alterados, da lista CamposAuditaveis; nil sem campos
}

// NovoRegistroAuditoria valida os enumerados, o par tipo/ID do alvo e os campos (lista fechada,
// sem repetição, gravados na ordem da lista). Spec: 3.2, RN-13, Q-44.
func NovoRegistroAuditoria(id uuid.UUID, data time.Time, operador string, acao AcaoAuditoria, alvoTipo TipoAlvo, alvoID uuid.UUID, resultado ResultadoAuditoria, campos ...string) (RegistroAuditoria, error) {
	ordenados, err := ordenarCampos(campos)
	if err != nil {
		return RegistroAuditoria{}, err
	}
	switch {
	case !acao.Valida():
		return RegistroAuditoria{}, fmt.Errorf("%w: ação de auditoria %q", errs.ErrDadosInvalidos, acao)
	case !alvoTipo.Valido():
		return RegistroAuditoria{}, fmt.Errorf("%w: tipo de alvo %q", errs.ErrDadosInvalidos, alvoTipo)
	case (alvoTipo == AlvoNenhum) != (alvoID == uuid.Nil):
		return RegistroAuditoria{}, fmt.Errorf("%w: alvo sem tipo ou tipo sem alvo", errs.ErrDadosInvalidos)
	case !resultado.Valido():
		return RegistroAuditoria{}, fmt.Errorf("%w: resultado %q", errs.ErrDadosInvalidos, resultado)
	}
	return RegistroAuditoria{ID: id, Data: data.UTC(), Operador: NormalizarLogin(operador), Acao: acao,
		AlvoTipo: alvoTipo, AlvoID: alvoID, Resultado: resultado, Campos: ordenados}, nil
}

func ordenarCampos(campos []string) ([]string, error) {
	pedidos := map[string]bool{}
	for _, c := range campos {
		pedidos[c] = true
	}
	var out []string
	for _, c := range CamposAuditaveis() {
		if pedidos[c] {
			out = append(out, c)
			delete(pedidos, c)
		}
	}
	for c := range pedidos {
		return nil, fmt.Errorf("%w: campo de auditoria %q fora da lista", errs.ErrDadosInvalidos, c)
	}
	return out, nil
}
