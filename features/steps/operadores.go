package steps

import (
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/google/uuid"

	"perfila/internal/application/usecase"
	"perfila/internal/domain/entity"
)

// Passos de usuarios.feature (@aplicacao) e os comuns aos cenários @web de papéis e auditoria.
// Spec: RN-11 a RN-13, Q-35 a Q-41.
type estadoOperadores struct {
	entrou  entity.Operador
	sessoes map[string]time.Time // login → carimbo da sessão aberta
}

const retencaoPadrao = 180 // AUDITORIA_RETENCAO_DIAS (Spec 3.8)

func registrarPassosOperadores(sc *godog.ScenarioContext, m *Mundo) {
	sc.Step(`^que existe o operador "([^"]*)" com o papel "([^"]*)" e a senha "([^"]*)"$`, m.existeOperador)
	sc.Step(`^o operador "([^"]*)" cria o operador "([^"]*)" com o papel "([^"]*)" e a senha "([^"]*)"$`,
		func(ator, login, papel, senha string) error {
			a, err := m.operador(ator)
			if err != nil {
				return err
			}
			_, m.err = m.uc().operadores.Criar(m.ctx, a, login, "Operador Sintético", entity.PapelOperador(papel), senha)
			return nil
		})
	sc.Step(`^o operador "([^"]*)" está ativo com o papel "([^"]*)"$`, func(login, papel string) error {
		o, err := m.operador(login)
		if err != nil {
			return err
		}
		if !o.Ativo || o.Papel != entity.PapelOperador(papel) {
			return fmt.Errorf("operador %s: ativo=%v papel=%s", login, o.Ativo, o.Papel)
		}
		return nil
	})
	sc.Step(`^o operador "([^"]*)" (não )?precisa trocar a senha$`, func(login, nao string) error {
		o, err := m.operador(login)
		if err != nil {
			return err
		}
		if o.TrocarSenha != (nao == "") {
			return fmt.Errorf("operador %s: TrocarSenha=%v", login, o.TrocarSenha)
		}
		return nil
	})
	sc.Step(`^que o operador "([^"]*)" precisa trocar a senha$`, func(login string) error {
		return m.mudarOperador(login, func(o *entity.Operador) { o.TrocarSenha = true })
	})
	sc.Step(`^que o operador "([^"]*)" foi desativado$`, func(login string) error {
		return m.mudarOperador(login, func(o *entity.Operador) { o.Desativar() })
	})
	sc.Step(`^que o papel do operador "([^"]*)" mudou para "([^"]*)"$`, func(login, papel string) error {
		return m.mudarOperador(login, func(o *entity.Operador) { o.Papel = entity.PapelOperador(papel) })
	})
	sc.Step(`^que o operador "([^"]*)" está ativo$`, func(login string) error { _, err := m.operador(login); return err })
	sc.Step(`^o resultado da criação é "([^"]*)"$`, func(r string) error {
		if r == "ok" {
			return m.semErro()
		}
		return m.falhaComErro(r)
	})
	sc.Step(`^o operador "([^"]*)" desativa a própria conta$`, func(login string) error {
		return m.comAtor(login, func(a entity.Operador) error { return m.uc().operadores.Desativar(m.ctx, a, a.ID) })
	})
	sc.Step(`^o operador "([^"]*)" muda o próprio papel para "([^"]*)"$`, func(login, papel string) error {
		return m.comAtor(login, func(a entity.Operador) error {
			return m.uc().operadores.MudarPapel(m.ctx, a, a.ID, entity.PapelOperador(papel))
		})
	})
	sc.Step(`^o operador "([^"]*)" desativa "([^"]*)"$`, func(ator, alvo string) error {
		return m.sobre(ator, alvo, func(a, o entity.Operador) error { return m.uc().operadores.Desativar(m.ctx, a, o.ID) })
	})
	sc.Step(`^o operador "([^"]*)" redefine a senha provisória de "([^"]*)"$`, func(ator, alvo string) error {
		return m.sobre(ator, alvo, func(a, o entity.Operador) error {
			return m.uc().operadores.RedefinirSenha(m.ctx, a, o.ID, "senha-provisoria-2")
		})
	})
	sc.Step(`^o operador "([^"]*)" muda o papel de "([^"]*)" para "([^"]*)"$`, func(ator, alvo, papel string) error {
		return m.sobre(ator, alvo, func(a, o entity.Operador) error {
			return m.uc().operadores.MudarPapel(m.ctx, a, o.ID, entity.PapelOperador(papel))
		})
	})
	sc.Step(`^alguém entra com o login "([^"]*)" e a senha "([^"]*)"$`, func(login, senha string) error {
		m.ops.entrou, m.err = m.uc().autenticar.Executar(m.ctx, login, senha)
		return nil
	})
	sc.Step(`^a entrada devolve o operador "([^"]*)" com o papel "([^"]*)"$`, func(login, papel string) error {
		if err := m.semErro(); err != nil {
			return err
		}
		if m.ops.entrou.Login != login || m.ops.entrou.Papel != entity.PapelOperador(papel) {
			return fmt.Errorf("entrou %s com %s", m.ops.entrou.Login, m.ops.entrou.Papel)
		}
		return nil
	})
	sc.Step(`^que o operador "([^"]*)" tem uma sessão aberta$`, func(login string) error {
		o, err := m.operador(login)
		if err != nil {
			return err
		}
		m.ops.sessoes[login] = o.SenhaAtualizadaEm
		m.clock.Avancar(time.Minute) // o pedido seguinte acontece depois da entrada
		return nil
	})
	sc.Step(`^a sessão aberta de "([^"]*)" é recusada$`, func(login string) error {
		_, m.err = m.conferirSessao(login)
		return m.falhaComErro("nao_autenticado")
	})
	sc.Step(`^a sessão aberta de "([^"]*)" devolve o papel "([^"]*)"$`, func(login, papel string) error {
		o, err := m.conferirSessao(login)
		if err != nil {
			return err
		}
		if o.Papel != entity.PapelOperador(papel) {
			return fmt.Errorf("papel da sessão %s", o.Papel)
		}
		return nil
	})
	sc.Step(`^o operador "([^"]*)" troca a própria senha informando a atual "([^"]*)"$`, func(login, atual string) error {
		return m.comAtor(login, func(a entity.Operador) error {
			_, err := m.uc().operadores.TrocarSenha(m.ctx, a, atual, "senha-nova-123456")
			return err
		})
	})
	sc.Step(`^que não existe nenhum operador$`, func() error {
		if os, _ := m.operadores.Listar(m.ctx); len(os) > 0 {
			return fmt.Errorf("há %d operadores", len(os))
		}
		return nil
	})
	sc.Step(`^que já existem operadores gravados$`, func() error {
		return m.existeOperador("admin", "ADMINISTRADOR", "senha-admin-123")
	})
	sc.Step(`^o servidor cria o administrador inicial "([^"]*)"$`, func(login string) error {
		h, err := m.hash.Gerar("senha-inicial-123")
		if err != nil {
			return err
		}
		_, m.err = m.uc().operadores.CriarAdministradorInicial(m.ctx, login, h)
		return m.err
	})
	sc.Step(`^o operador "([^"]*)" não existe$`, func(login string) error {
		if _, err := m.operadores.BuscarPorLogin(m.ctx, login); err == nil {
			return fmt.Errorf("operador %s existe", login)
		}
		return nil
	})

	// Auditoria (RN-13).
	sc.Step(`^o painel registra a ação "([^"]*)" do operador "([^"]*)" sobre o cliente "([^"]*)" com o resultado "([^"]*)"$`,
		func(acao, login, cliente, resultado string) error {
			m.err = m.uc().registrarAuditoria.Executar(m.ctx, login, entity.AcaoAuditoria(acao), entity.AlvoCliente,
				id(cliente), entity.ResultadoAuditoria(resultado))
			return m.err
		})
	sc.Step(`^a auditoria tem (\d+) registros?$`, func(n int) error {
		if got := len(m.auditoria.Todos()); got != n {
			return fmt.Errorf("a auditoria tem %d registros, esperado %d", got, n)
		}
		return nil
	})
	sc.Step(`^o último registro da auditoria é "([^"]*)" do operador "([^"]*)" sobre o (cliente|operador) "([^"]*)" com o resultado "([^"]*)"$`,
		m.ultimoRegistro)
	// Depois de uma ação, o painel redireciona para a ficha (PRG), e abrir a ficha também é registrado.
	sc.Step(`^a auditoria tem um registro "([^"]*)" do operador "([^"]*)" sobre o (cliente|operador) "([^"]*)" com o resultado "([^"]*)"$`,
		func(acao, login, tipo, alvo, resultado string) error {
			regs := m.auditoria.Todos()
			for i := len(regs) - 1; i >= 0; i-- {
				if string(regs[i].Acao) == acao {
					return m.conferirRegistro(regs[i], acao, login, tipo, alvo, resultado)
				}
			}
			return fmt.Errorf("nenhum registro %q na auditoria", acao)
		})
	sc.Step(`^a auditoria tem registros das ações "([^"]*)" e "([^"]*)"$`, func(a, b string) error {
		achou := map[string]bool{}
		for _, r := range m.auditoria.Todos() {
			achou[string(r.Acao)] = true
		}
		if !achou[a] || !achou[b] {
			return fmt.Errorf("ações gravadas: %v", achou)
		}
		return nil
	})
	sc.Step(`^nenhum registro da auditoria contém "([^"]*)" nem "([^"]*)"$`, func(a, b string) error {
		for _, r := range m.auditoria.Todos() {
			linha := fmt.Sprintf("%+v", r)
			if strings.Contains(linha, a) || strings.Contains(linha, b) {
				return fmt.Errorf("registro com dado pessoal: %s", linha)
			}
		}
		return nil
	})
	sc.Step(`^o operador "([^"]*)" lista a auditoria$`, func(login string) error {
		return m.comAtor(login, func(a entity.Operador) error {
			_, err := m.uc().listarAuditoria.Executar(m.ctx, a, usecase.FiltroListagemAuditoria{})
			return err
		})
	})
	sc.Step(`^que existe um registro de auditoria de (\d+) dias atrás$`, func(dias int) error {
		r, err := entity.NovoRegistroAuditoria(uuid.New(), m.clock.Agora().AddDate(0, 0, -dias), "admin",
			entity.AcaoEntrar, entity.AlvoNenhum, uuid.Nil, entity.ResultadoSucesso)
		if err != nil {
			return err
		}
		return m.auditoria.Registrar(m.ctx, r)
	})
	sc.Step(`^a auditoria é expurgada$`, func() error { _, m.err = m.uc().expurgarAuditoria.Executar(m.ctx); return m.err })
}

func (m *Mundo) existeOperador(login, papel, senha string) error {
	h, err := m.hash.Gerar(senha)
	if err != nil {
		return err
	}
	o, err := entity.NovoOperador(id("op:"+login), login, "Operador Sintético", entity.PapelOperador(papel), h, m.clock.Agora())
	if err != nil {
		return err
	}
	o.TrocarSenha = false
	return m.operadores.Salvar(m.ctx, o)
}

func (m *Mundo) operador(login string) (entity.Operador, error) {
	o, err := m.operadores.BuscarPorLogin(m.ctx, login)
	if err != nil {
		return entity.Operador{}, fmt.Errorf("operador %q: %w", login, err)
	}
	return o, nil
}

func (m *Mundo) mudarOperador(login string, f func(*entity.Operador)) error {
	o, err := m.operador(login)
	if err != nil {
		return err
	}
	f(&o)
	return m.operadores.Salvar(m.ctx, o)
}

// comAtor executa a ação com o operador como ator e guarda o erro do caso de uso.
func (m *Mundo) comAtor(login string, f func(entity.Operador) error) error {
	a, err := m.operador(login)
	if err != nil {
		return err
	}
	m.err = f(a)
	return nil
}

func (m *Mundo) sobre(ator, alvo string, f func(a, o entity.Operador) error) error {
	o, err := m.operador(alvo)
	if err != nil {
		return err
	}
	return m.comAtor(ator, func(a entity.Operador) error { return f(a, o) })
}

func (m *Mundo) conferirSessao(login string) (entity.Operador, error) {
	carimbo, ok := m.ops.sessoes[login]
	if !ok {
		return entity.Operador{}, fmt.Errorf("%s não tem sessão aberta", login)
	}
	o, err := m.operador(login)
	if err != nil {
		return entity.Operador{}, err
	}
	return m.uc().autenticar.Conferir(m.ctx, o.ID, carimbo)
}

func (m *Mundo) ultimoRegistro(acao, login, tipo, alvo, resultado string) error {
	regs := m.auditoria.Todos()
	if len(regs) == 0 {
		return fmt.Errorf("auditoria vazia")
	}
	return m.conferirRegistro(regs[len(regs)-1], acao, login, tipo, alvo, resultado)
}

func (m *Mundo) conferirRegistro(r entity.RegistroAuditoria, acao, login, tipo, alvo, resultado string) error {
	alvoID := id(alvo)
	if tipo == "operador" {
		o, err := m.operador(alvo)
		if err != nil {
			return err
		}
		alvoID = o.ID
	}
	if string(r.Acao) != acao || r.Operador != login || string(r.AlvoTipo) != tipo || r.AlvoID != alvoID ||
		string(r.Resultado) != resultado {
		return fmt.Errorf("último registro: %+v", r)
	}
	return nil
}
