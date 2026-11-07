package steps

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"

	"github.com/cucumber/godog"

	"perfila/internal/api"
	"perfila/internal/application/usecase"
	"perfila/internal/domain/entity"
	"perfila/internal/domain/service"
)

// Estado dos cenários @api (httptest, sem rede — Spec B-04).
type estadoAPI struct {
	token      string
	finalidade string
	termo      string
	resposta   *httptest.ResponseRecorder
}

var reClienteSintetico = regexp.MustCompile(`c-\d{3}`)

func registrarPassosAPI(sc *godog.ScenarioContext, m *Mundo) {
	sc.Step(`^que a API está no ar e exige o token "([^"]*)"$`, m.apiNoAr)
	sc.Step(`^que existe o cliente "([^"]*)" ativo com eventos de compra na categoria "([^"]*)"$`, m.clienteComCompras)
	sc.Step(`^que o catálogo possui produtos ativos na categoria "([^"]*)"$`, m.catalogoAtivo)
	sc.Step(`^que as requisições que exigem corpo usam a finalidade "([^"]*)" e o termo "([^"]*)"$`,
		func(f, t string) error { m.api.finalidade, m.api.termo = f, t; return nil })
	sc.Step(`^"([^"]*)" é chamado com o token "([^"]*)"$`, func(req, tok string) error { return m.chamar(req, tok, "") })
	sc.Step(`^"([^"]*)" é chamado com o token "([^"]*)" e o corpo:$`,
		func(req, tok string, corpo *godog.DocString) error { return m.chamar(req, tok, corpo.Content) })
	sc.Step(`^a resposta tem status (\d+)$`, m.status)
	sc.Step(`^o corpo contém "([^"]*)", "([^"]*)" e "([^"]*)"$`, m.corpoContem)
	sc.Step(`^o corpo tem o código "([^"]*)"$`, m.codigo)
	sc.Step(`^a recomendação no corpo é nula$`, m.recomendacaoNula)
}

func (m *Mundo) apiNoAr(token string) error {
	m.api.token = token
	m.segmentos.Adicionar(entity.Segmento{ID: id("seg:padrao"), Nome: "padrao", Prioridade: 99, Ativo: true})
	return nil
}

func (m *Mundo) clienteComCompras(nome, categoria string) error {
	if err := m.clienteAtivo(nome); err != nil {
		return err
	}
	t := m.clock.Agora().AddDate(0, 0, -1)
	for _, v := range []float64{100, 200, 300} {
		if err := m.gravaEvento(nome, entity.TipoCompra, v, categoria, t); err != nil {
			return err
		}
	}
	return nil
}

func (m *Mundo) catalogoAtivo(categoria string) error {
	for _, nome := range []string{"Item 1", "Item 2"} {
		m.produtos.Adicionar(entity.Produto{ID: id("prod:" + nome), Nome: nome, Categoria: categoria, Preco: 10, Ativo: true})
	}
	return nil
}

// roteador monta a API com os mesmos fakes do cenário (raiz de composição de teste).
func (m *Mundo) roteador() http.Handler {
	c := m.uc()
	return api.NovoRouter(api.Dependencias{
		AuthToken:     m.api.token,
		Precondicoes:  usecase.NovasPrecondicoes(m.clientes, service.NovoValidadorLGPD(m.consentimentos)),
		Cadastrar:     c.cadastrar,
		Consentimento: c.consentimento,
		Coletar:       c.coletar,
		GerarPerfil:   c.gerarPerfil,
		Segmentar:     c.segmentar,
		Recomendar:    c.recomendar,
	})
}

// chamar recebe "MÉTODO /caminho"; "c-001" no caminho vira o UUID sintético do cliente.
func (m *Mundo) chamar(requisicao, token, corpo string) error {
	metodo, caminho, ok := strings.Cut(strings.TrimSpace(requisicao), " ")
	if !ok {
		return fmt.Errorf("requisição deve ser \"MÉTODO /caminho\": %q", requisicao)
	}
	caminho = reClienteSintetico.ReplaceAllStringFunc(caminho, func(n string) string { return id(n).String() })
	if corpo == "" && metodo == http.MethodPut && strings.HasSuffix(caminho, "/consentimento") {
		b, _ := json.Marshal(api.ConcederConsentimentoRequest{Finalidade: m.api.finalidade, VersaoTermo: m.api.termo})
		corpo = string(b)
	}
	req := httptest.NewRequest(metodo, caminho, strings.NewReader(corpo))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	m.api.resposta = httptest.NewRecorder()
	m.roteador().ServeHTTP(m.api.resposta, req)
	return nil
}

func (m *Mundo) status(quer int) error {
	if m.api.resposta == nil {
		return fmt.Errorf("nenhuma requisição feita")
	}
	if got := m.api.resposta.Code; got != quer {
		return fmt.Errorf("status: esperado %d, obtido %d (%s)", quer, got, strings.TrimSpace(m.api.resposta.Body.String()))
	}
	return nil
}

func (m *Mundo) corpoJSON() (map[string]json.RawMessage, error) {
	var corpo map[string]json.RawMessage
	if err := json.Unmarshal(m.api.resposta.Body.Bytes(), &corpo); err != nil {
		return nil, fmt.Errorf("corpo não é JSON: %v", err)
	}
	return corpo, nil
}

func (m *Mundo) corpoContem(a, b, c string) error {
	corpo, err := m.corpoJSON()
	if err != nil {
		return err
	}
	for _, k := range []string{a, b, c} {
		if _, ok := corpo[k]; !ok {
			return fmt.Errorf("corpo sem a chave %q: %s", k, m.api.resposta.Body.String())
		}
	}
	return nil
}

func (m *Mundo) codigo(quer string) error {
	corpo, err := m.corpoJSON()
	if err != nil {
		return err
	}
	var got string
	_ = json.Unmarshal(corpo["codigo"], &got)
	if got != quer {
		return fmt.Errorf("código: esperado %q, obtido %q", quer, got)
	}
	return nil
}

func (m *Mundo) recomendacaoNula() error {
	corpo, err := m.corpoJSON()
	if err != nil {
		return err
	}
	if v, ok := corpo["recomendacao"]; !ok || string(v) != "null" {
		return fmt.Errorf("recomendação deveria ser null: %s", string(v))
	}
	return nil
}
