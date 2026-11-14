package external

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
	"perfila/internal/shared/errs"
)

var _ port.DestinoPublicitarioGateway = DestinoArquivoCSV{}

// DestinoArquivoCSV é o destino publicitário do MVP: gera o arquivo de público personalizado
// (identificador com hash) para a equipe subir na plataforma. Não envia nada a terceiros.
// Spec: RN-17, Q-49.
type DestinoArquivoCSV struct{}

// Enviar monta o CSV `email_sha256,segmento`, uma linha por hash, em ordem crescente e sem
// repetição, e devolve o conteúdo com o seu SHA-256 e o nome do arquivo.
func (DestinoArquivoCSV) Enviar(_ context.Context, lote entity.LotePublico) (entity.ReciboEnvio, error) {
	if len(lote.Hashes) == 0 {
		return entity.ReciboEnvio{}, fmt.Errorf("%w: lote de público vazio", errs.ErrDadosInvalidos)
	}
	hashes := append([]string{}, lote.Hashes...)
	sort.Strings(hashes)
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"email_sha256", "segmento"})
	anterior := ""
	for _, h := range hashes {
		if h == anterior {
			continue
		}
		anterior = h
		_ = w.Write([]string{h, lote.SegmentoNome})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return entity.ReciboEnvio{}, fmt.Errorf("external: gerando CSV do público: %w", err)
	}
	soma := sha256.Sum256(buf.Bytes())
	return entity.ReciboEnvio{
		Destino:     entity.DestinoArquivoCSV,
		NomeArquivo: fmt.Sprintf("publico-%s-%s-%s.csv", nomeSeguro(lote.SegmentoNome), strings.ToLower(string(lote.Tipo)), lote.Data.UTC().Format("20060102T150405Z")),
		Conteudo:    buf.Bytes(),
		Hash:        hex.EncodeToString(soma[:]),
	}, nil
}

// nomeSeguro deixa no nome do arquivo só letras, dígitos e hífen.
func nomeSeguro(s string) string {
	s = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
			return r
		case r >= 'A' && r <= 'Z':
			return r + ('a' - 'A')
		}
		return '-'
	}, s)
	if s == "" {
		return "segmento"
	}
	return s
}
