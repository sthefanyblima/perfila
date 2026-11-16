package entity

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"perfila/internal/shared/errs"
)

// Limites da configuração de segmentos e catálogo. Spec: RN-14, Q-33, Q-43.
const (
	SegmentoNomeMax       = 40
	SegmentoDescricaoMax  = 200
	SegmentoPrioridadeMin = 1
	SegmentoPrioridadeMax = 999
	ProdutoNomeMax        = 100
	ProdutoCategoriaMax   = 40
	ProdutoPrecoMax       = 1_000_000
)

// ValidarCriterio confere só a gramática da RN-06, sem perfil. Spec: RN-14.
func ValidarCriterio(criterio string) error {
	_, err := AvaliarCriterio(criterio, PerfilCliente{})
	return err
}

// EhPadrao informa se o segmento é o padrao (critério vazio). Spec: RN-06, RN-14.
func (s Segmento) EhPadrao() bool { return strings.TrimSpace(s.Criterio) == "" }

// Validar aplica as regras de um segmento: nome, descrição, prioridade e critério. Spec: RN-14.
func (s Segmento) Validar() error {
	nome := strings.TrimSpace(s.Nome)
	switch {
	case nome == "":
		return errs.NovoErroCampo("nome", "Informe o nome do segmento.")
	case utf8.RuneCountInString(nome) > SegmentoNomeMax:
		return errs.NovoErroCampo("nome", fmt.Sprintf("Use até %d caracteres no nome.", SegmentoNomeMax))
	case strings.ContainsAny(nome, ";:"):
		return errs.NovoErroCampo("nome", "O nome não pode ter ponto e vírgula nem dois-pontos.")
	case utf8.RuneCountInString(s.Descricao) > SegmentoDescricaoMax:
		return errs.NovoErroCampo("descricao", fmt.Sprintf("Use até %d caracteres na descrição.", SegmentoDescricaoMax))
	case s.Prioridade < SegmentoPrioridadeMin || s.Prioridade > SegmentoPrioridadeMax:
		return errs.NovoErroCampo("prioridade", fmt.Sprintf("A prioridade vai de %d a %d.", SegmentoPrioridadeMin, SegmentoPrioridadeMax))
	}
	return ValidarCriterio(s.Criterio)
}

// ValidarSegmentos aplica as regras do conjunto à configuração inteira, já com a mudança:
// cada segmento válido; exatamente um de critério vazio (o padrao), ativo e com a maior prioridade
// entre os ativos; prioridade única entre os ativos; nome único sem diferenciar maiúsculas.
// Spec: RN-14, Q-02, Q-17.
func ValidarSegmentos(lista []Segmento) error {
	var padrao *Segmento
	nomes := map[string]string{}
	prioridades := map[int]string{}
	for i := range lista {
		s := lista[i]
		if err := s.Validar(); err != nil {
			return err
		}
		chave := strings.ToLower(strings.TrimSpace(s.Nome))
		if outro, ok := nomes[chave]; ok {
			return errs.NovoErroCampo("nome", fmt.Sprintf("Já existe o segmento %q.", outro))
		}
		nomes[chave] = s.Nome
		if s.EhPadrao() {
			if padrao != nil {
				return errs.NovoErroCampo("criterio", "Só o segmento padrão pode ter o critério vazio. Monte pelo menos uma condição.")
			}
			padrao = &lista[i]
		}
		if !s.Ativo {
			continue
		}
		if outro, ok := prioridades[s.Prioridade]; ok {
			return errs.NovoErroCampo("prioridade", fmt.Sprintf("A prioridade %d já é do segmento %q.", s.Prioridade, outro))
		}
		prioridades[s.Prioridade] = s.Nome
	}
	if padrao == nil {
		return errs.NovoErroCampo("criterio", "O segmento padrão (critério vazio) precisa continuar existindo.")
	}
	if !padrao.Ativo {
		return errs.NovoErroCampo("ativo", fmt.Sprintf("O segmento %q não pode ser desativado.", padrao.Nome))
	}
	for _, s := range lista {
		if s.Ativo && !s.EhPadrao() && s.Prioridade >= padrao.Prioridade {
			return errs.NovoErroCampo("prioridade", fmt.Sprintf("A prioridade precisa ser menor que a do segmento %q (%d).", padrao.Nome, padrao.Prioridade))
		}
	}
	return nil
}

// NormalizarCategoria tira os espaços das pontas e passa a categoria para minúsculas, para casar
// com a categoria dos eventos (RN-05). Spec: RN-14, Q-43.
func NormalizarCategoria(categoria string) string {
	return strings.ToLower(strings.TrimSpace(categoria))
}

// Validar aplica as regras de um produto: nome, categoria e preço. Spec: RN-14, Q-43.
func (p Produto) Validar() error {
	nome := strings.TrimSpace(p.Nome)
	categoria := strings.TrimSpace(p.Categoria)
	switch {
	case nome == "":
		return errs.NovoErroCampo("nome", "Informe o nome do produto.")
	case utf8.RuneCountInString(nome) > ProdutoNomeMax:
		return errs.NovoErroCampo("nome", fmt.Sprintf("Use até %d caracteres no nome.", ProdutoNomeMax))
	case categoria == "":
		return errs.NovoErroCampo("categoria", "Informe a categoria.")
	case utf8.RuneCountInString(categoria) > ProdutoCategoriaMax:
		return errs.NovoErroCampo("categoria", fmt.Sprintf("Use até %d caracteres na categoria.", ProdutoCategoriaMax))
	}
	return validarPreco(p.Preco)
}

func validarPreco(preco float64) error {
	if math.IsNaN(preco) || preco < 0 || preco > ProdutoPrecoMax {
		return errs.NovoErroCampo("preco", "O preço vai de 0 a 1.000.000.")
	}
	return nil
}
