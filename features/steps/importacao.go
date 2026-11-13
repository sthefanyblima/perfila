package steps

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"perfila/internal/application/usecase"
	"perfila/internal/shared/errs"
)

// estadoImportacao guarda o arquivo do cenário e o último relatório (RN-16, Q-47).
type estadoImportacao struct {
	arquivo   []byte
	resultado usecase.ResultadoImportacao
}

func registrarPassosImportacao(sc *godog.ScenarioContext, m *Mundo) {
	sc.Step(`^que o arquivo de eventos é:$`, func(d *godog.DocString) error { m.importacao.arquivo = csvSintetico(d.Content); return nil })
	sc.Step(`^que o arquivo de eventos começa com o BOM e é:$`, func(d *godog.DocString) error {
		m.importacao.arquivo = append([]byte("\xEF\xBB\xBF"), csvSintetico(d.Content)...)
		return nil
	})
	sc.Step(`^que o arquivo de eventos é (?:maior que (\d+) KiB|(.+))$`, m.arquivoDescrito)
	sc.Step(`^que o arquivo tem (\d+) linhas de navegação válidas do cliente "([^"]*)"$`, m.arquivoComLinhas)
	sc.Step(`^a prévia do arquivo é feita$`, func() error { return m.importar(false) })
	sc.Step(`^o arquivo é importado$`, func() error { return m.importar(true) })
	sc.Step(`^que o arquivo já foi importado$`, func() error { return m.semFalha(m.importar(true)) })
	sc.Step(`^que a prévia do arquivo mostrou (\d+) linhas? aceitas?$`, func(n int) error {
		if err := m.semFalha(m.importar(false)); err != nil {
			return err
		}
		return m.resultadoTem(n, 0, 0)
	})
	sc.Step(`^que o cliente "([^"]*)" revoga o consentimento$`, m.revoga)
	sc.Step(`^o resultado tem (\d+) aceitas?, (\d+) duplicadas? e (\d+) rejeitadas?$`, m.resultadoTem)
	sc.Step(`^a linha (\d+) é rejeitada com o erro "([^"]*)"(?: e o motivo "([^"]*)")?$`, m.linhaRejeitada)
	sc.Step(`^a linha (\d+) é duplicada$`, m.linhaDuplicada)
	sc.Step(`^o arquivo é recusado com o motivo "([^"]*)"$`, m.arquivoRecusado)
	sc.Step(`^o cliente "([^"]*)" tem (\d+) eventos? gravados?$`, m.eventosGravadosDo)

	sc.Step(`^o lote de eventos é enviado à API com o token "([^"]*)":$`, func(tok string, d *godog.DocString) error {
		return m.enviarLote(tok, d.Content, false)
	})
	sc.Step(`^o lote de eventos é enviado à API com o token "([^"]*)" só para conferir:$`, func(tok string, d *godog.DocString) error {
		return m.enviarLote(tok, d.Content, true)
	})
	sc.Step(`^o operador envia o arquivo de eventos pelo painel$`, m.enviarPeloPainel)
	sc.Step(`^que o operador enviou o arquivo de eventos pelo painel$`, m.enviarPeloPainel)
	sc.Step(`^o operador confirma a importação$`, m.confirmarImportacao)
	sc.Step(`^a página mostra (\d+) prontas?, (\d+) duplicadas? e (\d+) com erro$`, func(a, d, r int) error {
		return m.paginaTemMarcador(fmt.Sprintf(`data-teste="resumo-importacao" data-aceitas="%d" data-duplicadas="%d" data-rejeitadas="%d"`, a, d, r), true)
	})
	sc.Step(`^a resposta é um CSV que começa com "([^"]*)"$`, func(inicio string) error {
		if !strings.HasPrefix(m.api.resposta.Header().Get("Content-Type"), "text/csv") || !strings.HasPrefix(m.web.pagina, inicio) {
			return fmt.Errorf("esperava CSV começando com %q: %s %q", inicio, m.api.resposta.Header().Get("Content-Type"), m.web.pagina)
		}
		return nil
	})
	sc.Step(`^o log do painel registra a importação com (\d+) importados?, (\d+) duplicados? e (\d+) rejeitados?$`, func(i, d, r int) error {
		quer := fmt.Sprintf("operador=operador acao=importar_eventos importados=%d duplicados=%d rejeitados=%d", i, d, r)
		if !strings.Contains(m.web.log.String(), quer) {
			return fmt.Errorf("o log não registra %q: %s", quer, m.web.log.String())
		}
		return nil
	})
	sc.Step(`^o log do painel não contém o ID do cliente "([^"]*)"$`, func(c string) error {
		if strings.Contains(m.web.log.String(), id(c).String()) {
			return fmt.Errorf("o log do painel contém o ID do cliente %s", c)
		}
		return nil
	})

	sc.Step(`^o corpo informa (\d+) aceitas?, (\d+) duplicadas? e (\d+) rejeitadas?$`, m.corpoInforma)
	sc.Step(`^o corpo informa a linha (\d+) com o código "([^"]*)"$`, m.corpoInformaLinha)
}

// csvSintetico troca os IDs sintéticos ("c-001") pelos UUIDs do cenário (B-08).
func csvSintetico(s string) []byte {
	return []byte(reClienteSintetico.ReplaceAllStringFunc(s, func(n string) string { return id(n).String() }))
}

func (m *Mundo) arquivoDescrito(kib, descricao string) error {
	const cabecalho = "cliente_id,tipo,origem,timestamp,valor\n"
	switch {
	case kib != "":
		var n int
		_, _ = fmt.Sscan(kib, &n)
		linha := "# " + strings.Repeat("x", 1021) + "\n" // 1 KiB por linha; nunca chega a ser lida
		m.importacao.arquivo = []byte(cabecalho + strings.Repeat(linha, n+1))
	case descricao == "vazio":
		m.importacao.arquivo = nil
	case descricao == "só com o cabeçalho":
		m.importacao.arquivo = []byte(cabecalho)
	case descricao == "em Latin-1":
		m.importacao.arquivo = csvSintetico(cabecalho + "c-001,navegacao,promoção\xe7,2026-11-11T09:00:00Z,0\n")
	case descricao == "com o cabeçalho separado por ponto e vírgula":
		m.importacao.arquivo = csvSintetico("cliente_id;tipo;origem;timestamp;valor\nc-001;navegacao;app;2026-11-11T09:00:00Z;0\n")
	case descricao == `sem a coluna "valor"`:
		m.importacao.arquivo = csvSintetico("cliente_id,tipo,origem,timestamp\nc-001,navegacao,app,2026-11-11T09:00:00Z\n")
	case descricao == `com a coluna desconhecida "produto"`:
		m.importacao.arquivo = csvSintetico(cabecalho[:len(cabecalho)-1] + ",produto\nc-001,navegacao,app,2026-11-11T09:00:00Z,0,p-1\n")
	case descricao == "com aspas mal fechadas":
		m.importacao.arquivo = csvSintetico(cabecalho + "c-001,navegacao,\"app,2026-11-11T09:00:00Z,0\n")
	default:
		return fmt.Errorf("arquivo desconhecido: %q", descricao)
	}
	return nil
}

func (m *Mundo) arquivoComLinhas(n int, cliente string) error {
	var b strings.Builder
	b.WriteString("cliente_id,tipo,origem,timestamp,valor\n")
	inicio := m.clock.Agora().Add(-48 * time.Hour)
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "%s,navegacao,app,%s,0\n", id(cliente), inicio.Add(time.Duration(i)*time.Second).Format(time.RFC3339))
	}
	m.importacao.arquivo = []byte(b.String())
	return nil
}

func (m *Mundo) casoImportar() *usecase.ImportarEventosUseCase {
	return m.uc().importar
}

// importar lê o arquivo e faz a prévia ou a importação; a falha fica em m.err.
func (m *Mundo) importar(gravar bool) error {
	m.importacao.resultado, m.err = usecase.ResultadoImportacao{}, nil
	entradas, err := usecase.LerCSVEventos(m.importacao.arquivo)
	if err != nil {
		m.err = err
		return nil
	}
	if gravar {
		m.importacao.resultado, m.err = m.casoImportar().Importar(m.ctx, entradas)
	} else {
		m.importacao.resultado, m.err = m.casoImportar().Previa(m.ctx, entradas)
	}
	return nil
}

func (m *Mundo) semFalha(err error) error {
	if err != nil {
		return err
	}
	return m.err
}

func (m *Mundo) resultadoTem(aceitas, duplicadas, rejeitadas int) error {
	if m.err != nil {
		return fmt.Errorf("lote recusado: %v", m.err)
	}
	r := m.importacao.resultado
	if r.Aceitas != aceitas || r.Duplicadas != duplicadas || r.Rejeitadas != rejeitadas {
		return fmt.Errorf("esperava %d/%d/%d, obtido %d/%d/%d (%+v)", aceitas, duplicadas, rejeitadas,
			r.Aceitas, r.Duplicadas, r.Rejeitadas, r.Linhas)
	}
	return nil
}

func (m *Mundo) linha(n int) (usecase.LinhaImportacao, error) {
	for _, l := range m.importacao.resultado.Linhas {
		if l.Numero == n {
			return l, nil
		}
	}
	return usecase.LinhaImportacao{}, fmt.Errorf("a linha %d não está no relatório (%+v)", n, m.importacao.resultado.Linhas)
}

func (m *Mundo) linhaRejeitada(n int, codigo, motivo string) error {
	l, err := m.linha(n)
	if err != nil {
		return err
	}
	if l.Situacao != usecase.SituacaoRejeitada || errs.Codigo(l.Erro) != codigo || l.Motivo != motivo {
		return fmt.Errorf("linha %d: esperava rejeitada %q/%q, obtido %s %q/%q", n, codigo, motivo, l.Situacao, errs.Codigo(l.Erro), l.Motivo)
	}
	return nil
}

func (m *Mundo) linhaDuplicada(n int) error {
	l, err := m.linha(n)
	if err != nil {
		return err
	}
	if l.Situacao != usecase.SituacaoDuplicada {
		return fmt.Errorf("linha %d: esperava duplicada, obtido %s", n, l.Situacao)
	}
	return nil
}

func (m *Mundo) arquivoRecusado(motivo string) error {
	var e *usecase.ErroArquivo
	if !errors.As(m.err, &e) || string(e.Motivo) != motivo || !errors.Is(m.err, errs.ErrDadosInvalidos) {
		return fmt.Errorf("esperava o arquivo recusado por %q, obtido %v", motivo, m.err)
	}
	return nil
}

// eventosGravadosDo lê o repositório direto: o passo confere o que foi gravado, sem RN-01.
func (m *Mundo) eventosGravadosDo(nome string, n int) error {
	evs, err := m.eventos.BuscarPorCliente(m.ctx, id(nome))
	if err != nil {
		return err
	}
	if len(evs) != n {
		return fmt.Errorf("cliente %s: esperava %d eventos gravados, há %d", nome, n, len(evs))
	}
	return nil
}

// ---- API ----

func (m *Mundo) enviarLote(token, corpo string, previa bool) error {
	caminho := "/v1/eventos/lote"
	if previa {
		caminho += "?previa=true"
	}
	req := httptest.NewRequest(http.MethodPost, caminho, bytes.NewReader(csvSintetico(corpo)))
	req.Header.Set("Authorization", "Bearer "+token)
	m.api.resposta = httptest.NewRecorder()
	m.roteador().ServeHTTP(m.api.resposta, req)
	return nil
}

type loteResposta struct {
	Aceitas, Duplicadas, Rejeitadas int
	Linhas                          []struct {
		Numero int
		Codigo string
	}
}

func (m *Mundo) lote() (loteResposta, error) {
	var r loteResposta
	err := json.Unmarshal(m.api.resposta.Body.Bytes(), &r)
	return r, err
}

func (m *Mundo) corpoInforma(aceitas, duplicadas, rejeitadas int) error {
	r, err := m.lote()
	if err != nil || r.Aceitas != aceitas || r.Duplicadas != duplicadas || r.Rejeitadas != rejeitadas {
		return fmt.Errorf("esperava %d/%d/%d: %s", aceitas, duplicadas, rejeitadas, m.api.resposta.Body.String())
	}
	return nil
}

func (m *Mundo) corpoInformaLinha(n int, codigo string) error {
	r, err := m.lote()
	if err == nil {
		for _, l := range r.Linhas {
			if l.Numero == n && l.Codigo == codigo {
				return nil
			}
		}
	}
	return fmt.Errorf("linha %d sem o código %q: %s", n, codigo, m.api.resposta.Body.String())
}

// ---- painel ----

// postarMultipart envia um formulário multipart (arquivo ou conteúdo oculto) com o CSRF da página.
func (m *Mundo) postarMultipart(caminho, campo, nomeArquivo string, valor []byte) error {
	var corpo bytes.Buffer
	w := multipart.NewWriter(&corpo)
	_ = w.WriteField("_csrf", m.web.csrf)
	if nomeArquivo != "" {
		f, err := w.CreateFormFile(campo, nomeArquivo)
		if err != nil {
			return err
		}
		_, _ = f.Write(valor)
	} else {
		_ = w.WriteField(campo, string(valor))
	}
	if err := w.Close(); err != nil {
		return err
	}
	req := httptest.NewRequest(http.MethodPost, caminho, &corpo)
	req.Header.Set("Content-Type", w.FormDataContentType())
	m.requisitar(req, true)
	return nil
}

func (m *Mundo) enviarPeloPainel() error {
	if err := m.webGET("/painel/eventos/importar", true); err != nil {
		return err
	}
	return m.postarMultipart("/painel/eventos/importar", "arquivo", "eventos.csv", m.importacao.arquivo)
}

// confirmarImportacao reenvia o conteúdo oculto da prévia, como o navegador faria.
func (m *Mundo) confirmarImportacao() error {
	s := reConteudoOculto.FindStringSubmatch(m.web.pagina)
	if s == nil {
		return fmt.Errorf("a prévia não oferece a confirmação")
	}
	return m.postarMultipart("/painel/eventos/importar/confirmar", "conteudo", "", []byte(html.UnescapeString(s[1])))
}

var reConteudoOculto = regexp.MustCompile(`name="conteudo" value="([^"]*)"`)
