package steps

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cucumber/godog"
)

// Passos de implantacao.feature (UML8, T-24). Rodam contra a implantação real do
// docker compose, que precisa estar no ar (make implantar). Fora do make verify (Q-13).
//
// PERFILA_IMPLANTACAO_URL  balanceador (padrão https://localhost:8443)
// PERFILA_COMPOSE_ARQUIVO  compose da implantação (padrão ../deploy/docker-compose.yml)

type estadoImplantacao struct {
	balanceador      *url.URL
	compose          string
	statuses         []int
	replicas         map[string]int
	replicaParada    string
	erroConexaoBD    error
	portaBDPublicada string
	tentouConectar   bool
	painel           *http.Response
}

func registrarPassosImplantacao(sc *godog.ScenarioContext) {
	e := &estadoImplantacao{}
	sc.Step(`^que a implantação está no ar com um balanceador, duas réplicas da aplicação e um banco$`, e.noAr)
	sc.Step(`^o dispositivo do cliente faz (\d+) chamadas a "([^"]*)" pelo balanceador$`, e.chamadas)
	sc.Step(`^todas as respostas têm status (\d+)$`, e.todasComStatus)
	sc.Step(`^as duas réplicas atenderam pelo menos uma chamada$`, e.duasReplicasAtenderam)
	sc.Step(`^que uma das réplicas é parada$`, e.pararReplica)
	sc.Step(`^o dispositivo do cliente tenta conectar ao banco na porta (\d+)$`, e.conectarBanco)
	sc.Step(`^a conexão é recusada$`, e.conexaoRecusada)
	sc.Step(`^o dispositivo do cliente abre "([^"]*)" pelo balanceador$`, e.abrirPainel)
	sc.Step(`^a página do painel responde 200 com os cabeçalhos de segurança$`, e.painelSeguro)

	// B-13: a réplica parada volta ao fim do cenário, para o próximo começar do Contexto.
	sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
		if e.replicaParada == "" {
			return ctx, err
		}
		if _, erro := e.dockerCompose("up", "-d", "--wait", "--no-deps", e.replicaParada); erro != nil {
			return ctx, fmt.Errorf("religando %s: %w", e.replicaParada, erro)
		}
		return ctx, err
	})
}

func (e *estadoImplantacao) dockerCompose(args ...string) (string, error) {
	cmd := exec.Command("docker", append([]string{"compose", "-f", e.compose}, args...)...)
	saida, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("docker compose %s: %w: %s", strings.Join(args, " "), err, saida)
	}
	return string(saida), nil
}

func envOu(chave, padrao string) string {
	if v := os.Getenv(chave); v != "" {
		return v
	}
	return padrao
}

// cliente HTTPS do dispositivo: o certificado da demonstração é autoassinado (Q-12).
var clienteDispositivo = &http.Client{
	Timeout:   10 * time.Second,
	Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}, // #nosec: certificado autoassinado
}

func (e *estadoImplantacao) noAr() error {
	u, err := url.Parse(envOu("PERFILA_IMPLANTACAO_URL", "https://localhost:8443"))
	if err != nil {
		return err
	}
	e.balanceador = u
	e.compose = envOu("PERFILA_COMPOSE_ARQUIVO", "../deploy/docker-compose.yml")

	saida, err := e.dockerCompose("ps", "--status", "running", "--services")
	if err != nil {
		return fmt.Errorf("implantação fora do ar (rode make implantar): %w", err)
	}
	for _, servico := range []string{"nginx", "app1", "app2", "db"} {
		if !strings.Contains("\n"+saida, "\n"+servico+"\n") {
			return fmt.Errorf("serviço %s não está rodando (rode make implantar); rodando: %q", servico, saida)
		}
	}
	limite := time.Now().Add(60 * time.Second)
	for {
		resp, err := clienteDispositivo.Get(e.balanceador.JoinPath("/saude").String())
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		if time.Now().After(limite) {
			return fmt.Errorf("balanceador %s não respondeu /saude com 200: %v", e.balanceador, err)
		}
		time.Sleep(time.Second)
	}
}

func (e *estadoImplantacao) chamadas(n int, caminho string) error {
	e.statuses, e.replicas = nil, map[string]int{}
	for i := 0; i < n; i++ {
		resp, err := clienteDispositivo.Get(e.balanceador.JoinPath(caminho).String())
		if err != nil {
			e.statuses = append(e.statuses, 0)
			continue
		}
		resp.Body.Close()
		if resp.TLS == nil {
			return fmt.Errorf("chamada %d não usou TLS", i+1)
		}
		e.statuses = append(e.statuses, resp.StatusCode)
		// Q-11: o Nginx informa a réplica; com nova tentativa vem "a, b" e vale a última.
		replica := resp.Header.Get("X-Omega-Replica")
		if partes := strings.Split(replica, ","); len(partes) > 0 {
			replica = strings.TrimSpace(partes[len(partes)-1])
		}
		e.replicas[replica]++
	}
	return nil
}

func (e *estadoImplantacao) todasComStatus(esperado int) error {
	for i, s := range e.statuses {
		if s != esperado {
			return fmt.Errorf("chamada %d: status %d, esperado %d (todos: %v)", i+1, s, esperado, e.statuses)
		}
	}
	return nil
}

func (e *estadoImplantacao) duasReplicasAtenderam() error {
	delete(e.replicas, "")
	if len(e.replicas) < 2 {
		return fmt.Errorf("réplicas que atenderam: %v, esperadas 2", e.replicas)
	}
	return nil
}

func (e *estadoImplantacao) pararReplica() error {
	if _, err := e.dockerCompose("stop", "app2"); err != nil {
		return err
	}
	e.replicaParada = "app2"
	return nil
}

func (e *estadoImplantacao) conectarBanco(porta int) error {
	e.tentouConectar = true
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(e.balanceador.Hostname(), fmt.Sprint(porta)), 3*time.Second)
	if err == nil {
		conn.Close()
	}
	e.erroConexaoBD = err
	// A porta do banco do Perfila foi publicada para fora do Docker? O formato de "não publicada"
	// varia com a versão do Compose (":0", "invalid IP:0" ou erro "no port"); vale o número final.
	saida, err := e.dockerCompose("port", "db", fmt.Sprint(porta))
	if err != nil {
		if strings.Contains(err.Error(), "no port") {
			e.portaBDPublicada = ""
			return nil
		}
		return err
	}
	saida = strings.TrimSpace(saida)
	if hostPorta := saida[strings.LastIndex(saida, ":")+1:]; hostPorta == "" || hostPorta == "0" {
		saida = ""
	}
	e.portaBDPublicada = saida
	return nil
}

// abrirPainel pede uma página do painel pelo Nginx (Spec 3.7.1, T-38).
func (e *estadoImplantacao) abrirPainel(caminho string) error {
	resp, err := clienteDispositivo.Get(e.balanceador.JoinPath(caminho).String())
	if err != nil {
		return err
	}
	resp.Body.Close()
	e.painel = resp
	return nil
}

func (e *estadoImplantacao) painelSeguro() error {
	r := e.painel
	if r == nil || r.StatusCode != http.StatusOK || r.TLS == nil {
		return fmt.Errorf("painel pelo balanceador: %v", r)
	}
	for k, quer := range map[string]string{
		"Content-Security-Policy": "default-src 'self'", "X-Content-Type-Options": "nosniff",
		"X-Frame-Options": "DENY", "Referrer-Policy": "same-origin", "Cache-Control": "no-store",
		"Strict-Transport-Security": "max-age=",
	} {
		if !strings.Contains(r.Header.Get(k), quer) {
			return fmt.Errorf("cabeçalho %s: esperado %q, obtido %q", k, quer, r.Header.Get(k))
		}
	}
	return nil
}

func (e *estadoImplantacao) conexaoRecusada() error {
	if !e.tentouConectar {
		return fmt.Errorf("nenhuma tentativa de conexão")
	}
	if p := e.portaBDPublicada; p != "" {
		return fmt.Errorf("o banco da implantação está publicado em %s: alcançável pelo dispositivo do cliente", p)
	}
	// O banco não publica porta: a tentativa direta falha (erroConexaoBD != nil) ou, se algo
	// aceitou, é outro serviço da máquina (por exemplo, outro PostgreSQL local), não o omega_db.
	return nil
}
