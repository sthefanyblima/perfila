package usecase

import (
	"bytes"
	"encoding/csv"
	"errors"
	"io"
	"strings"
	"unicode/utf8"
)

// ColunasImportacao é o cabeçalho do modelo, na ordem em que o modelo o escreve. Spec: RN-16.
const ColunasImportacao = "cliente_id,tipo,origem,timestamp,valor,categoria"

// LerCSVEventos lê o arquivo de RN-16: UTF-8 (BOM aceito), vírgula, cabeçalho por nome.
// Devolve uma entrada por linha de dados; arquivo recusado: *ErroArquivo. Spec: RN-16, Q-47.
func LerCSVEventos(conteudo []byte) ([]EntradaEvento, error) {
	if len(conteudo) > LimiteBytesImportacao {
		return nil, &ErroArquivo{Motivo: MotivoArquivoGrande}
	}
	conteudo = bytes.TrimPrefix(conteudo, []byte("\xEF\xBB\xBF"))
	if !utf8.Valid(conteudo) {
		return nil, &ErroArquivo{Motivo: MotivoCodificacaoInvalida}
	}
	r := csv.NewReader(bytes.NewReader(conteudo))
	r.FieldsPerRecord = -1 // a linha com colunas a mais ou a menos é recusada sozinha
	cabecalho, err := r.Read()
	if errors.Is(err, io.EOF) {
		return nil, &ErroArquivo{Motivo: MotivoArquivoVazio}
	}
	if err != nil {
		return nil, &ErroArquivo{Motivo: MotivoFormatoInvalido}
	}
	posicao, ok := lerCabecalho(cabecalho)
	if !ok {
		return nil, &ErroArquivo{Motivo: MotivoCabecalhoInvalido}
	}
	var entradas []EntradaEvento
	for {
		campos, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, &ErroArquivo{Motivo: MotivoFormatoInvalido}
		}
		if len(entradas) == LimiteLinhasImportacao {
			return nil, &ErroArquivo{Motivo: MotivoLinhasDemais}
		}
		linha, _ := r.FieldPos(0)
		en := EntradaEvento{Numero: linha, ColunasErradas: len(campos) != len(cabecalho)}
		if !en.ColunasErradas {
			valor := func(coluna string) string {
				if i, ok := posicao[coluna]; ok {
					return campos[i]
				}
				return ""
			}
			en.ClienteID, en.Tipo, en.Origem = valor("cliente_id"), valor("tipo"), valor("origem")
			en.Timestamp, en.Valor, en.Categoria = valor("timestamp"), valor("valor"), valor("categoria")
		}
		entradas = append(entradas, en)
	}
	if len(entradas) == 0 {
		return nil, &ErroArquivo{Motivo: MotivoArquivoVazio}
	}
	return entradas, nil
}

// lerCabecalho exige as colunas obrigatórias, aceita categoria e recusa repetida ou desconhecida.
func lerCabecalho(cabecalho []string) (map[string]int, bool) {
	posicao := map[string]int{}
	for i, nome := range cabecalho {
		nome = strings.ToLower(strings.TrimSpace(nome))
		if _, repetida := posicao[nome]; repetida || !strings.Contains(","+ColunasImportacao+",", ","+nome+",") || nome == "" {
			return nil, false
		}
		posicao[nome] = i
	}
	for _, obrigatoria := range strings.Split(strings.TrimSuffix(ColunasImportacao, ",categoria"), ",") {
		if _, ok := posicao[obrigatoria]; !ok {
			return nil, false
		}
	}
	return posicao, true
}
