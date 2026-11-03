# Perfila — alvos de verificação (Plano SDD, seção 3).
# Fase Core: @api e @implantacao ficam fora do filtro (ver features/bdd_test.go).

GO ?= go

.PHONY: verify fmt vet test bdd arch cover demo

verify: fmt vet test bdd arch ## portão de toda tarefa (Definição de pronto)

fmt:
	@test -z "$$(gofmt -l .)" || (echo "gofmt: arquivos fora do padrão:"; gofmt -l .; exit 1)

vet:
	$(GO) vet ./...

test:
	$(GO) test -race -count=1 ./internal/...

bdd:
	BDD_FORMAT=pretty $(GO) test -count=1 ./features/...

arch:
	$(GO) test -count=1 -run 'TestRegra' ./internal/

cover: ## portão F2: domínio + aplicação >= 80%
	$(GO) test -count=1 -coverpkg=./internal/domain/...,./internal/application/... -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out | tail -1

demo: ## fluxo do diagrama de sequência em memória
	$(GO) run ./cmd/perfila-demo
