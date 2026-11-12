package web

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
)

type relogioFixo struct{ t time.Time }

func (r *relogioFixo) Agora() time.Time { return r.t }

// Spec: 3.7.1, docs/design/03-textos.md (Formatos).
func TestFormatos(t *testing.T) {
	utc := time.Date(2026, 11, 12, 1, 7, 0, 0, time.UTC) // 22:07 do dia 11 em São Paulo
	casos := []struct{ nome, obtido, esperado string }{
		{"data e hora em São Paulo", dataHora(utc), "11/11/2026 22:07"},
		{"só data", soData(utc), "11/11/2026"},
		{"data zero", dataHora(time.Time{}), "—"},
		{"data nula", dataHoraPtr(nil), "—"},
		{"dinheiro", dinheiro(1234.56), "R$ 1.234,56"},
		{"dinheiro milhão", dinheiro(1000000), "R$ 1.000.000,00"},
		{"dinheiro zero", dinheiro(0), "R$ 0,00"},
		{"decimal", decimal(0.4, 2), "0,40"},
		{"negativo", decimal(-1500.5, 1), "-1.500,5"},
		{"e-mail mascarado", mascararEmail("cliente@exemplo.test"), "c***@exemplo.test"},
		{"e-mail sem arroba", mascararEmail("invalido"), "***"},
		{"ID truncado", uuidCurto(uuid.MustParse("6200ae2f-1111-2222-3333-44445555a4c9")), "6200ae2f…a4c9"},
		{"estado legível", rotulo("NAO_CONCEDIDO"), "Não concedido"},
		{"tipo legível", rotulo("navegacao"), "Navegação"},
		{"motivo legível", motivoLegivel("segmento:premium;categoria:livros"), "Segmento premium, categoria livros"},
		{"datetime-local", valorDataLocal(utc), "2026-11-11T22:07"},
		{"porcento limitado", porcento(1.7), "100"},
	}
	for _, c := range casos {
		if c.obtido != c.esperado {
			t.Errorf("%s: esperado %q, obtido %q", c.nome, c.esperado, c.obtido)
		}
	}
	for _, c := range []struct {
		texto string
		valor float64
	}{{"1.234,56", 1234.56}, {"120,50", 120.5}, {"120.50", 120.5}, {"R$ 10", 10}} {
		if v, err := lerDecimal(c.texto); err != nil || v != c.valor {
			t.Errorf("lerDecimal(%q) = %v, %v", c.texto, v, err)
		}
	}
	if v, err := lerDataLocal("2026-11-11T22:07"); err != nil || !v.Equal(utc) {
		t.Errorf("lerDataLocal = %v, %v", v, err)
	}
}

// Spec: RN-06 — cada condição é avaliada pelo domínio; a tela só traduz.
func TestCondicoes(t *testing.T) {
	p := &entity.PerfilCliente{TicketMedio: 610, FrequenciaCompras: 2}
	cs := condicoes("ticketMedio >= 500 && frequenciaCompras >= 3", p)
	if len(cs) != 2 || cs[0].Campo != "Ticket médio" || cs[0].Operador != "≥" || cs[0].Numero != "500" || !cs[0].Atendida || cs[1].Atendida {
		t.Fatalf("condições: %+v", cs)
	}
	if cs := condicoes("scoreEngajamento >= 0.7", nil); cs[0].Numero != "0,7" || cs[0].Avaliada {
		t.Fatalf("sem perfil: %+v", cs)
	}
	if condicoes("", p) != nil {
		t.Fatal("critério vazio deveria virar a etiqueta 'casa sempre'")
	}
}

// Spec: Q-22 — assinatura, validade e adulteração da sessão.
func TestSessao(t *testing.T) {
	chave := []byte(strings.Repeat("k", 32))
	agora := time.Date(2026, 11, 12, 12, 0, 0, 0, time.UTC)
	s := sessao{Operador: "operador", Expira: agora.Add(duracaoSessao), CSRF: "abc123"}
	valor := assinar(chave, s)
	if got, err := validar(chave, valor, agora); err != nil || got.Operador != "operador" || got.CSRF != "abc123" {
		t.Fatalf("sessão válida: %+v %v", got, err)
	}
	corpo, assinatura, _ := strings.Cut(valor, ".")
	casos := []struct {
		nome  string
		valor string
		agora time.Time
		erro  error
	}{
		{"vencida", valor, agora.Add(duracaoSessao), errSessaoVencida},
		{"outra chave", assinar([]byte(strings.Repeat("x", 32)), s), agora, errSessaoAusente},
		{"corpo adulterado", strings.Replace(corpo, "abc123", "zzz999", 1) + "." + assinatura, agora, errSessaoAusente},
		{"sem assinatura", corpo, agora, errSessaoAusente},
		{"vazia", "", agora, errSessaoAusente},
	}
	for _, c := range casos {
		if _, err := validar(chave, c.valor, c.agora); !errors.Is(err, c.erro) {
			t.Errorf("%s: esperado %v, obtido %v", c.nome, c.erro, err)
		}
	}
}

// Spec: Q-22 — 5 erros em 5 minutos bloqueiam; a janela passa.
func TestLimitador(t *testing.T) {
	r := &relogioFixo{t: time.Date(2026, 11, 12, 12, 0, 0, 0, time.UTC)}
	l := novoLimitador(r)
	for i := 0; i < maxTentativas; i++ {
		if l.bloqueado("192.0.2.1") {
			t.Fatalf("bloqueou cedo, na tentativa %d", i+1)
		}
		l.falhou("192.0.2.1")
	}
	if !l.bloqueado("192.0.2.1") || l.bloqueado("192.0.2.2") {
		t.Fatal("o bloqueio deveria valer só para o IP que errou")
	}
	r.t = r.t.Add(janelaBloqueio + time.Second)
	if l.bloqueado("192.0.2.1") {
		t.Fatal("a janela de 5 minutos deveria ter passado")
	}
}

// Spec: 3.7.1 — o destino depois de Entrar fica dentro do painel.
func TestDestinoSeguro(t *testing.T) {
	for d, quer := range map[string]bool{
		"/painel/clientes": true, "/painel/clientes/x?y=1": true, "//evil.test/painel/": false,
		"https://evil.test": false, "/v1/clientes": false, `/painel/\evil`: false, "": false,
	} {
		if destinoSeguro(d) != quer {
			t.Errorf("destinoSeguro(%q) deveria ser %v", d, quer)
		}
	}
	r := httptest.NewRequest("POST", "/painel/clientes/x/eventos", nil)
	r.Header.Set("Referer", "https://localhost:8443/painel/clientes/x/eventos")
	if got := voltaPara(r); got != "/painel/clientes/x/eventos" {
		t.Errorf("voltaPara de POST: %q", got)
	}
}

// As telas compilam com o layout e as partes (falha cedo se um template quebrar).
func TestTelasCarregam(t *testing.T) {
	ts, err := carregarTelas()
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"entrar", "erro", "clientes", "novo", "ficha", "confirmar", "segmentos", "catalogo", "operacao"} {
		if ts[n] == nil {
			t.Errorf("tela %s não carregou", n)
		}
	}
}
