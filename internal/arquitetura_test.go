// Teste de arquitetura (T-02): falha se alguma regra estrutural da Constituição for violada.
// Spec: C-01 (regra de dependência), C-05 (sem any/interface{}), C-07 (sem time.Now), C-08 (sem estado global).
package internal_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const modulo = "perfila/internal/"

type regra struct {
	camada    string   // diretório relativo a internal/
	proibidos []string // prefixos de import proibidos
	externos  []string // únicos imports não-stdlib permitidos fora do módulo
}

var regras = []regra{
	{"domain", []string{modulo + "api", modulo + "web", modulo + "application", modulo + "infra", "net/http", "database/sql"},
		[]string{"github.com/google/uuid"}},
	{"application", []string{modulo + "api", modulo + "web", modulo + "infra", "net/http", "database/sql"},
		[]string{"github.com/google/uuid"}},
	{"shared", []string{modulo + "api", modulo + "web", modulo + "application", modulo + "infra", modulo + "domain"}, nil},
	{"api", []string{modulo + "infra", modulo + "web", "database/sql"}, []string{"github.com/google/uuid"}},
	// web é um adaptador de apresentação como api (Q-20): nunca infra, banco nem a própria api.
	{"web", []string{modulo + "infra", modulo + "api", "database/sql"},
		[]string{"github.com/google/uuid"}}, // v1.5: a senha é conferida por AutenticarOperador (Q-36)
	// infra → domain (C-01, C-02); o driver PostgreSQL só entra aqui (Q-10).
	{"infra", []string{modulo + "api", modulo + "web", modulo + "application"},
		[]string{"github.com/google/uuid", "github.com/jackc/pgx/v5/stdlib", "golang.org/x/crypto/bcrypt"}}, // bcrypt só em infra/senha (Q-36)
}

func arquivosGo(t *testing.T, dir string, incluirTestes bool) []string {
	t.Helper()
	var out []string
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".go") &&
			(incluirTestes || !strings.HasSuffix(p, "_test.go")) {
			out = append(out, p)
		}
		return nil
	})
	return out
}

// C-01
func TestRegraDeDependencia(t *testing.T) {
	for _, r := range regras {
		for _, arq := range arquivosGo(t, r.camada, false) {
			f, err := parser.ParseFile(token.NewFileSet(), arq, nil, parser.ImportsOnly)
			if err != nil {
				t.Fatal(err)
			}
			for _, imp := range f.Imports {
				caminho, _ := strconv.Unquote(imp.Path.Value)
				for _, p := range r.proibidos {
					if strings.HasPrefix(caminho, p) {
						t.Errorf("C-01: %s importa %s (proibido em %s)", arq, caminho, r.camada)
					}
				}
				externo := strings.Contains(strings.Split(caminho, "/")[0], ".") && !strings.HasPrefix(caminho, modulo)
				if externo && !contem(r.externos, caminho) {
					t.Errorf("C-01: %s importa dependência externa %s", arq, caminho)
				}
			}
		}
	}
}

// C-05, C-07, C-08 no domínio e na aplicação.
func TestRegrasDeCodigo(t *testing.T) {
	for _, camada := range []string{"domain", "application"} {
		for _, arq := range arquivosGo(t, camada, false) {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, arq, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			for _, decl := range f.Decls {
				if g, ok := decl.(*ast.GenDecl); ok && g.Tok == token.VAR {
					for _, s := range g.Specs {
						for _, n := range s.(*ast.ValueSpec).Names {
							if n.Name != "_" {
								t.Errorf("C-08: %s declara variável de pacote %s", arq, n.Name)
							}
						}
					}
				}
			}
			ast.Inspect(f, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.InterfaceType:
					if x.Methods == nil || len(x.Methods.List) == 0 {
						t.Errorf("C-05: %s usa interface{} em %s", arq, fset.Position(x.Pos()))
					}
				case *ast.Ident:
					if x.Name == "any" && x.Obj == nil {
						t.Errorf("C-05: %s usa any em %s", arq, fset.Position(x.Pos()))
					}
				case *ast.SelectorExpr:
					if id, ok := x.X.(*ast.Ident); ok && id.Name == "time" && x.Sel.Name == "Now" {
						t.Errorf("C-07: %s chama time.Now em %s", arq, fset.Position(x.Pos()))
					}
				}
				return true
			})
		}
	}
}

// C-07 em todo internal/: time.Now só em infra/relogio (o domínio e a aplicação já são
// cobertos acima; aqui entram api e infra).
func TestRegraDoTempo(t *testing.T) {
	for _, arq := range arquivosGo(t, ".", false) {
		if strings.HasPrefix(filepath.ToSlash(arq), "infra/relogio/") {
			continue
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, arq, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if x, ok := n.(*ast.SelectorExpr); ok {
				if id, ok := x.X.(*ast.Ident); ok && id.Name == "time" && (x.Sel.Name == "Now" || x.Sel.Name == "Since") {
					t.Errorf("C-07: %s chama time.%s em %s", arq, x.Sel.Name, fset.Position(x.Pos()))
				}
			}
			return true
		})
	}
}

func contem(lista []string, s string) bool {
	for _, v := range lista {
		if v == s {
			return true
		}
	}
	return false
}
