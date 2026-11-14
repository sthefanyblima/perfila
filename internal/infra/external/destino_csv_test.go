package external

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"perfila/internal/domain/entity"
	"perfila/internal/shared/errs"
)

// Spec: RN-17 Q-49 — só hashes e o nome do segmento, em ordem, sem repetição.
func TestDestinoCSV(t *testing.T) {
	quando := time.Date(2026, 11, 13, 12, 0, 0, 0, time.UTC)
	b, a := entity.HashEmail("c-002@exemplo.test"), entity.HashEmail("c-001@exemplo.test")
	r, err := DestinoArquivoCSV{}.Enviar(context.Background(), entity.LotePublico{Tipo: entity.ExportacaoInclusao,
		SegmentoNome: "Vip Ouro", Data: quando, Hashes: []string{b, a, b}})
	if err != nil {
		t.Fatal(err)
	}
	primeiro, segundo := a, b
	if b < a {
		primeiro, segundo = b, a
	}
	esperado := "email_sha256,segmento\n" + primeiro + ",Vip Ouro\n" + segundo + ",Vip Ouro\n"
	if string(r.Conteudo) != esperado {
		t.Fatalf("conteúdo:\n%s", r.Conteudo)
	}
	soma := sha256.Sum256(r.Conteudo)
	if r.Hash != hex.EncodeToString(soma[:]) || r.Destino != entity.DestinoArquivoCSV {
		t.Fatalf("recibo: %+v", r)
	}
	if r.NomeArquivo != "publico-vip-ouro-inclusao-20261113T120000Z.csv" {
		t.Fatalf("nome: %s", r.NomeArquivo)
	}
	if _, err := (DestinoArquivoCSV{}).Enviar(context.Background(), entity.LotePublico{}); !errors.Is(err, errs.ErrDadosInvalidos) {
		t.Fatalf("lote vazio: %v", err)
	}
	if nomeSeguro("") != "segmento" {
		t.Fatal("nome vazio")
	}
}
