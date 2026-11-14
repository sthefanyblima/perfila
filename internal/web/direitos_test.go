package web

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"perfila/internal/application/usecase"
	"perfila/internal/domain/entity"
)

// Spec: Q-49 — o arquivo tem sempre as mesmas seções (listas vazias como []), datas ausentes como
// null e nunca o login do operador, que não é dado do titular.
func TestExportacaoFormato(t *testing.T) {
	agora := time.Date(2026, 11, 13, 12, 0, 0, 0, time.UTC)
	cli := entity.Cliente{ID: uuid.New(), Nome: "Ana Sintética", Email: "c-001@exemplo.test", DataCadastro: agora, Ativo: true}
	cons := entity.NovoConsentimento(uuid.New(), cli.ID, agora)
	vazio, err := json.Marshal(exportacao(usecase.DadosTitular{GeradoEm: agora, Cliente: cli, Consentimento: cons}))
	if err != nil {
		t.Fatal(err)
	}
	for _, trecho := range []string{`"versaoFormato":"1"`, `"historico":[]`, `"eventos":[]`, `"recomendacoes":[]`, `"produtos":[]`,
		`"perfil":null`, `"segmento":null`, `"dataConsentimento":null`, `"dataEliminacao":null`, `"status":"NAO_CONCEDIDO"`} {
		if !strings.Contains(string(vazio), trecho) {
			t.Errorf("arquivo sem %s: %s", trecho, vazio)
		}
	}

	_ = cons.Conceder("segmentacao", "v1", agora)
	mov, _ := entity.NovoMovimento(uuid.New(), cons, entity.MovimentoConcessao, entity.OrigemPainel, "op-sintetico")
	cheio, err := json.Marshal(exportacao(usecase.DadosTitular{GeradoEm: agora, Cliente: cli, Consentimento: cons,
		Historico: []entity.MovimentoConsentimento{mov}}))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(cheio), "op-sintetico") || !strings.Contains(string(cheio), `"origem":"painel"`) {
		t.Errorf("movimento exportado: %s", cheio)
	}
}
