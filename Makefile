# Perfila — alvos de verificação (Plano SDD, seção 3).
# @slow (@implantacao) fica fora do verify e roda em bdd-implantacao (Q-13, features/bdd_test.go).
# A suíte PostgreSQL roda quando PERFILA_TEST_DATABASE_URL está definida (CI e make pg-teste).

GO ?= go

.PHONY: verify fmt vet test bdd arch cover demo serve demo-api build contrato pg-teste implantar bdd-implantacao demo-implantacao expirar exemplos derrubar

verify: fmt vet test bdd arch ## portão de toda tarefa (Definição de pronto)

fmt:
	@test -z "$$(gofmt -l .)" || (echo "gofmt: arquivos fora do padrão:"; gofmt -l .; exit 1)

vet:
	$(GO) vet ./...

test:
	$(GO) test -race -count=1 ./internal/... ./cmd/...

bdd:
	BDD_FORMAT=pretty $(GO) test -count=1 ./features/...

arch:
	$(GO) test -count=1 -run 'TestRegra' ./internal/

contrato: ## suíte de contrato dos repositórios (memória sempre; PostgreSQL com PERFILA_TEST_DATABASE_URL)
	$(GO) test -count=1 -v -run 'Contrato|Migracoes' ./internal/infra/...

PG_TESTE_URL = postgres://omega:omega@localhost:55432/omega_db?sslmode=disable

pg-teste: ## sobe um PostgreSQL descartável (Docker, porta 55432) e roda verify contra ele
	docker rm -f perfila-pg-teste >/dev/null 2>&1 || true
	docker run -d --name perfila-pg-teste -p 55432:5432 -e POSTGRES_USER=omega -e POSTGRES_PASSWORD=omega -e POSTGRES_DB=omega_db postgres:16-alpine >/dev/null
	until docker exec perfila-pg-teste pg_isready -U omega -d omega_db >/dev/null 2>&1; do sleep 1; done
	PERFILA_TEST_DATABASE_URL='$(PG_TESTE_URL)' $(MAKE) verify; s=$$?; docker rm -f perfila-pg-teste >/dev/null; exit $$s

cover: ## portão F2: domínio + aplicação >= 80%
	$(GO) test -count=1 -coverpkg=./internal/domain/...,./internal/application/... -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out | tail -1

demo: ## fluxo do diagrama de sequência em memória
	$(GO) run ./cmd/perfila-demo

build: ## binário estático do UML8
	CGO_ENABLED=0 $(GO) build -o bin/omega-server ./cmd/omega-server

# Credenciais de desenvolvimento do painel (Q-29): operador / dev-senha. Nunca use fora da máquina local.
PAINEL_DEV_HASH = $$2a$$12$$XwydlCAe/Yz4gKKJFY8Dkem4QwOLLvhPfraOmQ3mvP9/CvqoKS3Am
PAINEL_DEV_CHAVE = 00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff

serve: ## API e painel em modo memória na porta 8080 (token: dev-token; painel: operador / dev-senha)
	AUTH_TOKEN=$${AUTH_TOKEN:-dev-token} ENGAJAMENTO_META_EVENTOS=$${ENGAJAMENTO_META_EVENTOS:-10} \
	PAINEL_SENHA_HASH='$(PAINEL_DEV_HASH)' PAINEL_CHAVE_SESSAO=$(PAINEL_DEV_CHAVE) $(GO) run ./cmd/omega-server serve

COMPOSE = docker compose -f deploy/docker-compose.yml

implantar: ## UML8: Nginx (TLS, :8443) + 2 réplicas + PostgreSQL; gera segredos na primeira vez (Q-14)
	sh scripts/preparar-implantacao.sh
	$(COMPOSE) up -d --build --wait

bdd-implantacao: ## cenários @implantacao contra a implantação no ar (make implantar antes)
	BDD_TAGS=@implantacao BDD_FORMAT=pretty $(GO) test -count=1 ./features/...

demo-implantacao: ## roteiro da API pelo Nginx (HTTPS) gravando no PostgreSQL (make implantar antes)
	mkdir -p bin
	$(COMPOSE) cp nginx:/etc/nginx/certs/omega.crt bin/omega.crt
	PERFILA_CA=bin/omega.crt ./scripts/demo-api.sh https://localhost:$${PERFILA_PORTA_HTTPS:-8443} "$$(sed -n 's/^AUTH_TOKEN=//p' deploy/config.env)"

expirar: ## administrador: expira recomendações vencidas na réplica app1 (UML8, via exec)
	$(COMPOSE) exec app1 omega-server -config /etc/omega/config.env expirar-recomendacoes

exemplos: ## carrega os 74 clientes de exemplo (14 com história e 60 de volume) na implantação, se o banco não tiver clientes (Q-30, Q-32)
	$(COMPOSE) exec app1 omega-server -config /etc/omega/config.env carregar-exemplos

derrubar: ## derruba a implantação (mantém os dados; apague com: $(COMPOSE) down -v)
	$(COMPOSE) down

demo-api: ## roda o roteiro de curl contra um servidor já no ar (make serve em outro terminal)
	./scripts/demo-api.sh
