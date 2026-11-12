package web

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strconv"
	"strings"
)

//go:embed templates/*.html
var arquivosTemplates embed.FS

//go:embed static
var arquivosEstaticos embed.FS

// pagina é o que toda tela recebe: moldura (operador, navegação, CSRF, toast) e o conteúdo.
type pagina struct {
	Titulo    string
	Nav       string // item ativo da navegação
	Operador  string
	CSRF      string
	Aviso     string // toast de confirmação
	RequestID string
	Dados     any
}

type telas map[string]*template.Template

// fichaCSRF e formCons levam o token CSRF às partes da ficha.
type fichaCSRF struct {
	D    dadosFicha
	CSRF string
}

type formCons struct {
	D                   dadosFicha
	CSRF, Acao, Prefixo string
	Dialogo             bool
}

// alerta alimenta a parte "alerta" (Alert): Tom critical, warn, ok ou info; Teste marca o
// aviso para os cenários @web; Agora pede role="alert" (erro que acabou de acontecer).
type alerta struct {
	Tom, Titulo, Texto, Teste        string
	Agora                            bool
	RequestID                        string
	AcaoHref, AcaoTexto, AcaoDialogo string
}

// funcoes recebe versoes, o hash de cada estático, usado na URL para o cache imutável.
func funcoes(versoes map[string]string) template.FuncMap {
	return template.FuncMap{
		"dataHora": dataHora, "data": soData, "dataHoraPtr": dataHoraPtr, "dinheiro": dinheiro,
		"decimal": decimal, "mascara": mascararEmail, "uuidCurto": uuidCurto, "rotulo": rotulo,
		"tom": tomDoEstado, "iconeEstado": iconeDoEstado, "condicoes": condicoes, "porcento": porcento, "icone": icone,
		"estatico":   func(nome string) string { return "/painel/static/" + nome + "?v=" + versoes[nome] },
		"estado":     func(v any) string { return fmt.Sprint(v) },
		"soma":       func(a, b int) int { return a + b },
		"alertaInfo": func(titulo, texto string) alerta { return alerta{Tom: "info", Titulo: titulo, Texto: texto} },
		"alertaAtencao": func(titulo, texto string) alerta {
			return alerta{Tom: "warn", Titulo: titulo, Texto: texto, Teste: "sem-recomendacao"}
		},
		"motivo":       motivoLegivel,
		"lista":        func(v ...string) []string { return v },
		"fichaComCSRF": func(d dadosFicha, csrf string) fichaCSRF { return fichaCSRF{D: d, CSRF: csrf} },
		"formConsentimento": func(d dadosFicha, csrf, acao string, dialogo bool) formCons {
			f := formCons{D: d, CSRF: csrf, Acao: acao, Dialogo: dialogo}
			if dialogo {
				f.Prefixo = "dlg-"
			}
			return f
		},
	}
}

// carregarTelas compila cada tela com o layout e as partes comuns.
func carregarTelas() (telas, error) {
	versoes := map[string]string{}
	_ = fs.WalkDir(arquivosEstaticos, "static", func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			b, _ := arquivosEstaticos.ReadFile(p)
			soma := sha256.Sum256(b)
			versoes[strings.TrimPrefix(p, "static/")] = hex.EncodeToString(soma[:])[:10]
		}
		return nil
	})
	nomes, err := fs.Glob(arquivosTemplates, "templates/tela_*.html")
	if err != nil {
		return nil, err
	}
	t := telas{}
	for _, n := range nomes {
		tpl, err := template.New("layout.html").Funcs(funcoes(versoes)).ParseFS(arquivosTemplates,
			"templates/layout.html", "templates/partes.html", n)
		if err != nil {
			return nil, err
		}
		t[strings.TrimSuffix(strings.TrimPrefix(n, "templates/tela_"), ".html")] = tpl
	}
	return t, nil
}

// render completa a moldura e escreve a tela. Requisições htmx de página recebem HX-Push-Url,
// para o histórico seguir o destino de um redirecionamento.
func (p *painel) render(w http.ResponseWriter, r *http.Request, status int, tela string, pg pagina) {
	s := sessaoDe(r)
	pg.Operador, pg.CSRF = s.Operador, s.CSRF
	if pg.Aviso == "" {
		n := r.URL.Query().Get("n")
		if _, err := strconv.Atoi(n); err != nil {
			n = ""
		}
		pg.Aviso = aviso(r.URL.Query().Get("aviso"), n)
	}
	pg.RequestID = r.Header.Get("X-Request-ID")
	var buf bytes.Buffer
	if err := p.telas[tela].Execute(&buf, pg); err != nil {
		p.d.Log.ErrorContext(r.Context(), "painel_template", "tela", tela, "erro", err.Error())
		http.Error(w, "Algo falhou do nosso lado.", http.StatusInternalServerError)
		return
	}
	if r.Header.Get("HX-Request") != "" && r.Method == http.MethodGet {
		w.Header().Set("HX-Push-Url", r.URL.RequestURI())
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
}

type dadosErro struct{ Titulo, Texto string }

func (p *painel) renderErro(w http.ResponseWriter, r *http.Request, status int, titulo, texto string) {
	p.render(w, r, status, "erro", pagina{Titulo: titulo, Dados: dadosErro{titulo, texto}})
}

func (p *painel) naoEncontrada(w http.ResponseWriter, r *http.Request) {
	p.renderErro(w, r, http.StatusNotFound, "Página não encontrada", "Confira o endereço ou volte para a lista de clientes.")
}

// estaticos serve CSS, JS e fontes embutidos, com cache de um ano (a URL leva a versão).
func estaticos() http.Handler {
	sub, _ := fs.Sub(arquivosEstaticos, "static")
	arquivos := http.StripPrefix("/painel/static/", http.FileServer(http.FS(sub)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		arquivos.ServeHTTP(w, r)
	})
}
