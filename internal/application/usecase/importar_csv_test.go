package usecase

import (
	"errors"
	"strings"
	"testing"

	"perfila/internal/shared/errs"
)

// Spec: RN-16 — formato do valor: só decimal com ponto.
func TestDecimalComPonto(t *testing.T) {
	for s, quer := range map[string]bool{
		"0": true, "150.90": true, "-1": true, "1000000": true,
		"150,90": false, "1e3": false, "Inf": false, "NaN": false, ".5": false, "5.": false, "0x10": false, "": false, "-": false,
	} {
		if got := decimalComPonto(s); got != quer {
			t.Errorf("decimalComPonto(%q) = %v, quer %v", s, got, quer)
		}
	}
}

// Spec: RN-16 — cabeçalho por nome, linhas numeradas pelo arquivo, colunas erradas por linha.
func TestLerCSVEventos(t *testing.T) {
	csv := "valor,cliente_id,tipo,origem,timestamp\n\n10,a,compra,web,2026-11-10T14:30:00Z\n" +
		"\"1\nx\",b,navegacao,app,t\nsó,duas\n"
	entradas, err := LerCSVEventos([]byte(csv))
	if err != nil {
		t.Fatal(err)
	}
	if len(entradas) != 3 {
		t.Fatalf("esperava 3 entradas, obtido %d", len(entradas))
	}
	if e := entradas[0]; e.Numero != 3 || e.ClienteID != "a" || e.Valor != "10" || e.Categoria != "" || e.ColunasErradas {
		t.Errorf("primeira entrada: %+v", e)
	}
	if e := entradas[1]; e.Numero != 4 || e.Valor != "1\nx" {
		t.Errorf("campo com quebra de linha: %+v", e)
	}
	if e := entradas[2]; e.Numero != 6 || !e.ColunasErradas {
		t.Errorf("linha com colunas a menos: %+v", e)
	}
}

// Spec: RN-16 — arquivo recusado inteiro, com o motivo, embrulhando ErrDadosInvalidos (C-06).
func TestLerCSVEventosRecusa(t *testing.T) {
	cab := "cliente_id,tipo,origem,timestamp,valor\n"
	for motivo, conteudo := range map[MotivoArquivo]string{
		MotivoArquivoVazio:        "\xEF\xBB\xBF",
		MotivoArquivoGrande:       cab + strings.Repeat("x", LimiteBytesImportacao),
		MotivoLinhasDemais:        cab + strings.Repeat("a,b,c,d,e\n", LimiteLinhasImportacao+1),
		MotivoCodificacaoInvalida: cab + "a,b,\xe7,d,e\n",
		MotivoCabecalhoInvalido:   "cliente_id,tipo,origem,timestamp,valor,valor\n",
		MotivoFormatoInvalido:     cab + "a,\"b,c,d,e\n",
	} {
		_, err := LerCSVEventos([]byte(conteudo))
		var e *ErroArquivo
		if !errors.As(err, &e) || e.Motivo != motivo || !errors.Is(err, errs.ErrDadosInvalidos) {
			t.Errorf("%s: obtido %v", motivo, err)
		}
	}
}
