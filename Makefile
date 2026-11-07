# Perfila — alvos de verificação (Plano SDD, seção 3).
# Fase Support: só @implantacao fica fora do filtro até a fase Generic (ver features/bdd_test.go).

GO ?= go

.PHONY: verify fmt vet test bdd arch cover demo serve demo-api build

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

cover: ## portão F2: domínio + aplicação >= 80%
	$(GO) test -count=1 -coverpkg=./internal/domain/...,./internal/application/... -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out | tail -1

demo: ## fluxo do diagrama de sequência em memória
	$(GO) run ./cmd/perfila-demo

build: ## binário estático do UML8
	CGO_ENABLED=0 $(GO) build -o bin/omega-server ./cmd/omega-server

serve: ## API em modo memória na porta 8080 (token: dev-token)
	AUTH_TOKEN=$${AUTH_TOKEN:-dev-token} ENGAJAMENTO_META_EVENTOS=$${ENGAJAMENTO_META_EVENTOS:-10} $(GO) run ./cmd/omega-server serve

demo-api: ## roda o roteiro de curl contra um servidor já no ar (make serve em outro terminal)
	./scripts/demo-api.sh
