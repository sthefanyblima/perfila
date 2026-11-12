package steps

import (
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"perfila/internal/application/usecase"
	"perfila/internal/domain/entity"
)

// Estado dos cenários de leitura do painel (Q-21, Q-23).
type estadoLeitura struct {
	pagina    usecase.PaginaClientes
	ficha     usecase.FichaCliente
	resultado usecase.ResultadoCliente
	segmentos []entity.Segmento
	catalogo  []entity.Produto
}

func registrarPassosLeitura(sc *godog.ScenarioContext, m *Mundo) {
	sc.Step(`^que o cliente "([^"]*)" foi processado$`, m.foiProcessado)
	sc.Step(`^que o cliente "([^"]*)" foi processado de novo (\d+) horas? depois$`, func(c string, h int) error {
		m.clock.Avancar(time.Duration(h) * time.Hour)
		return m.foiProcessado(c)
	})
	sc.Step(`^que existe o cliente "([^"]*)" ativo com o nome "([^"]*)"$`, m.clienteComNome)
	sc.Step(`^que existem (\d+) clientes ativos cadastrados$`, m.variosClientes)
	sc.Step(`^os clientes são listados$`, func() error { return m.listar(usecase.FiltroListagem{}) })
	sc.Step(`^os clientes com consentimento "([^"]*)" são listados$`, func(e string) error {
		f := usecase.FiltroListagem{}
		if e != "TODOS" {
			s := entity.StatusConsentimento(e)
			f.Consentimento = &s
		}
		return m.listar(f)
	})
	sc.Step(`^a página (\d+) de clientes é listada$`, func(p int) error { return m.listar(usecase.FiltroListagem{Pagina: p}) })
	sc.Step(`^os clientes são buscados por "([^"]*)"$`, func(b string) error { return m.listar(usecase.FiltroListagem{Busca: b}) })
	sc.Step(`^a lista tem (\d+) clientes?$`, m.listaTem)
	sc.Step(`^a página tem (\d+) clientes de um total de (\d+)$`, m.paginaTem)
	sc.Step(`^a linha do cliente "([^"]*)" mostra o consentimento "([^"]*)" e o segmento "([^"]*)"$`, m.linhaMostra)
	sc.Step(`^a ficha do cliente "([^"]*)" é consultada$`, func(c string) error {
		m.leitura.ficha, m.err = m.uc().consultarCliente.Executar(m.ctx, id(c))
		return nil
	})
	sc.Step(`^a consulta não falha$`, m.semErro)
	sc.Step(`^a ficha mostra o consentimento "([^"]*)"$`, func(e string) error {
		return iguais("consentimento da ficha", string(m.leitura.ficha.Consentimento.Status), e)
	})
	sc.Step(`^o resultado do cliente "([^"]*)" é consultado$`, func(c string) error {
		m.leitura.resultado, m.err = m.uc().consultarResultado.Executar(m.ctx, id(c))
		return nil
	})
	sc.Step(`^o resultado traz (\d+) eventos, sem perfil e sem recomendações$`, m.resultadoSemPerfil)
	sc.Step(`^o resultado traz o segmento "([^"]*)" e (\d+) recomendações, a mais recente "([^"]*)"$`, m.resultadoProcessado)
	sc.Step(`^os segmentos são listados$`, func() error {
		m.leitura.segmentos, m.err = m.uc().listarSegmentos.Executar(m.ctx)
		return m.err
	})
	sc.Step(`^os segmentos listados são "([^"]*)" e "([^"]*)"$`, func(a, b string) error {
		nomes := []string{}
		for _, s := range m.leitura.segmentos {
			nomes = append(nomes, s.Nome)
		}
		return iguais("segmentos", strings.Join(nomes, ","), a+","+b)
	})
	sc.Step(`^o catálogo é listado$`, func() error {
		m.leitura.catalogo, m.err = m.uc().listarCatalogo.Executar(m.ctx)
		return m.err
	})
	sc.Step(`^o catálogo listado é "([^"]*)", "([^"]*)" e "([^"]*)"$`, func(a, b, c string) error {
		nomes := []string{}
		for _, p := range m.leitura.catalogo {
			nomes = append(nomes, p.Categoria+"/"+p.Nome)
		}
		return iguais("catálogo", strings.Join(nomes, ","), a+","+b+","+c)
	})
}

func iguais(oque, obtido, esperado string) error {
	if obtido != esperado {
		return fmt.Errorf("%s: esperado %q, obtido %q", oque, esperado, obtido)
	}
	return nil
}

func (m *Mundo) foiProcessado(c string) error {
	_, err := m.uc().processar.Executar(m.ctx, id(c))
	return err
}

func (m *Mundo) clienteComNome(c, nome string) error {
	return m.clientes.Salvar(m.ctx, entity.Cliente{ID: id(c), Nome: nome, Email: c + "@exemplo.test",
		DataCadastro: m.clock.Agora(), Ativo: true})
}

func (m *Mundo) variosClientes(n int) error {
	for i := 1; i <= n; i++ {
		if err := m.clienteAtivo(fmt.Sprintf("c-%03d", i)); err != nil {
			return err
		}
	}
	return nil
}

func (m *Mundo) listar(f usecase.FiltroListagem) error {
	m.leitura.pagina, m.err = m.uc().listarClientes.Executar(m.ctx, f)
	return m.err
}

func (m *Mundo) listaTem(n int) error {
	if got := len(m.leitura.pagina.Linhas); got != n || m.leitura.pagina.Total != n {
		return fmt.Errorf("lista: esperado %d, obtido %d linhas (total %d)", n, got, m.leitura.pagina.Total)
	}
	return nil
}

func (m *Mundo) paginaTem(linhas, total int) error {
	if len(m.leitura.pagina.Linhas) != linhas || m.leitura.pagina.Total != total {
		return fmt.Errorf("página: esperado %d de %d, obtido %d de %d", linhas, total,
			len(m.leitura.pagina.Linhas), m.leitura.pagina.Total)
	}
	return nil
}

func (m *Mundo) linhaMostra(c, consentimento, segmento string) error {
	for _, l := range m.leitura.pagina.Linhas {
		if l.Cliente.ID == id(c) {
			if string(l.Consentimento) != consentimento || l.Segmento != segmento {
				return fmt.Errorf("linha de %s: obtido %s / %q", c, l.Consentimento, l.Segmento)
			}
			return nil
		}
	}
	return fmt.Errorf("cliente %s fora da lista", c)
}

func (m *Mundo) resultadoSemPerfil(n int) error {
	r := m.leitura.resultado
	if len(r.Eventos) != n || r.Perfil != nil || len(r.Recomendacoes) != 0 {
		return fmt.Errorf("resultado: %d eventos, perfil %v, %d recomendações", len(r.Eventos), r.Perfil != nil, len(r.Recomendacoes))
	}
	return nil
}

func (m *Mundo) resultadoProcessado(segmento string, n int, estado string) error {
	r := m.leitura.resultado
	if m.err != nil {
		return m.err
	}
	if r.Segmento == nil || r.Segmento.Nome != segmento {
		return fmt.Errorf("segmento do resultado: %+v", r.Segmento)
	}
	if len(r.Recomendacoes) != n || string(r.Recomendacoes[0].Status) != estado {
		return fmt.Errorf("recomendações: %d, a primeira em %v", len(r.Recomendacoes), r.Recomendacoes)
	}
	if _, ok := r.Produtos[r.Recomendacoes[0].ProdutoID]; !ok {
		return fmt.Errorf("produto da recomendação ausente do resultado")
	}
	return nil
}
