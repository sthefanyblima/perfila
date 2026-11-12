# Perfila

**Sistema de Segmentação e Perfilamento do Comportamento do Cliente** · Equipe Omega · Go 1.22

**Equipe:** Alexia Rodrigues, Carla Evellyn, Davi Silva, Jhenyffer Kyria, Leila Biggi, Lara Nogueira e Sthefany Barboza.

O Perfila transforma eventos de comportamento (navegação, compra, interação) em um **perfil** do cliente, atribui um **segmento** por critérios declarativos e gera uma **recomendação** de produto, sempre condicionado ao **consentimento LGPD**.

Desenvolvido com **Spec-Driven Development** (a especificação é a fonte da verdade e o agente de IA implementa) e **Behavior-Driven Development** (cenários Gherkin executáveis provam o comportamento).

## Estado (11/11/2026): MVP concluído

| Marco | Período | Estado |
|---|---|---|
| Arquivos `.md` (CLAUDE.md + UML1..UML8) | 20/10–27/10 | ✅ |
| **Core** (domínio + aplicação), 50% | 27/10–08/11 | ✅ 79 cenários · cobertura 92,5% |
| 1ª chamada de acompanhamento | 09–11/11 | ✅ `docs/RELATORIO_PARCIAL.md` |
| **Support** (API, middlewares, CMD, gateways), 25% | 09–13/11 | ✅ 20 cenários de API |
| **Generic** (PostgreSQL, Docker/Nginx, LGPD), 25% | 16–19/11 | ✅ contrato nos dois backends · 3 cenários de implantação · checklist LGPD |
| Fechamento do escopo | 19/11 | daqui em diante, só correções |
| Ensaio · evento final | 23/11 · 24/11 | roteiro em `docs/ROTEIRO_DEMONSTRACAO.md` |

**110 cenários BDD** (107 no `make verify` + 3 de implantação), teste de arquitetura e suíte de contrato dos repositórios em memória e em PostgreSQL.

## Arquitetura

Clean Architecture (UML7), com a regra de dependência verificada por teste (`internal/arquitetura_test.go`):

```text
api ──► application ──► domain ◄── infra        shared/errs: importável por todos
        (casos de uso)  (entity, port, service)  (memoria, relogio, repository/pg, external, config)
```

Implantação (UML8): o **Nginx** (TLS, única porta de entrada) distribui para **duas réplicas** do `omega-server`, que acessam o **PostgreSQL** `omega_db` numa rede interna, inalcançável pelo dispositivo do cliente.

## Como rodar

### Desenvolvimento (Go 1.22+)

```bash
make verify   # gofmt, vet, testes com -race, 107 cenários BDD e teste de arquitetura
make demo     # fluxo do diagrama de sequência em memória, passo a passo
make serve    # API em modo memória na porta 8080 (token dev-token)
make demo-api # roteiro de curl contra o make serve (em outro terminal)
make cover    # cobertura de domínio + aplicação
make pg-teste # sobe um PostgreSQL descartável no Docker e roda o verify contra ele
make build    # binário estático bin/omega-server
```

### Implantação (Docker com Compose v2)

```bash
make implantar         # gera segredos (1ª vez) e sobe nginx + app1 + app2 + db
make demo-implantacao  # roteiro da API pelo Nginx em HTTPS, gravando no PostgreSQL
make bdd-implantacao   # 3 cenários @implantacao (inclui parar uma réplica)
make expirar           # administrador: expira recomendações vencidas
make derrubar          # derruba (os dados ficam no volume omega_db)
```

- A única porta publicada é `https://localhost:8443`, com certificado autoassinado gerado na primeira subida.
- O token da API fica em `deploy/config.env` e a senha do banco em `deploy/.env`. Os dois são gerados aleatoriamente e ficam fora do git.
- O cabeçalho `X-Omega-Replica` mostra qual réplica atendeu.
- No Windows sem `make` nem Go, os comandos equivalentes estão em `docs/ROTEIRO_DEMONSTRACAO.md`, seção 4.

## Configuração (`config.env`, Spec 3.8)

| Variável | Padrão | Uso |
|---|---|---|
| `HTTP_PORT` | 8080 | porta do servidor |
| `DATABASE_URL` | vazia = modo memória | PostgreSQL `omega_db`; o servidor aplica as migrações ao subir. A implantação sempre a define (Q-07) |
| `AUTH_TOKEN` | obrigatória | token Bearer |
| `RECOMENDACAO_VALIDADE_DIAS` | 30 | validade da recomendação (RN-08) |
| `ENGAJAMENTO_META_EVENTOS` | 30 | meta de eventos em 30 dias para o score (RN-05) |
| `LOG_LEVEL` | info | debug, info, warn ou error |

O ambiente prevalece sobre o arquivo. Exemplo em `deploy/config.env.example`.

## API (Spec 3.7)

| Rota | Caso de uso | Exige |
|---|---|---|
| `POST /v1/clientes` | UC-09 Cadastrar | token |
| `PUT /v1/clientes/{id}/consentimento` | UC-08 Conceder | token |
| `POST /v1/clientes/{id}/consentimento/revogacao` | UC-08 Revogar | token |
| `GET /v1/clientes/{id}/consentimento` | UC-08 Consultar | token |
| `POST /v1/clientes/{id}/eventos` | UC-01 Registrar evento | token + consentimento |
| `POST /v1/clientes/{id}/processamento` | UC-01 → UC-02 → UC-03 → UC-04 (RN-09) | token + consentimento |
| `GET /saude` | saúde, para o balanceador | — |

Erros: JSON `{codigo, mensagem}` com códigos estáveis (Spec 3.5): 401 sem token, 403 `consentimento_invalido`, 404 `cliente_nao_encontrado`, 409 `cliente_inativo` / `transicao_invalida` / `perfil_nao_encontrado`, 422 `dados_invalidos`.

## LGPD

Nenhum caso de uso que processa dado pessoal roda sem consentimento válido (C-03), e a revogação tem efeito imediato (RN-03). Os logs da aplicação e do Nginx não trazem nome, e-mail nem IP (C-09), e os testes usam só dados sintéticos. O checklist completo, com a evidência de cada item e o que ficou para depois do MVP, está em [`docs/CHECKLIST_LGPD.md`](docs/CHECKLIST_LGPD.md). Os requisitos são interpretação técnica da equipe, não parecer jurídico.

## Mapa do repositório

```text
CLAUDE.md                   constituição para o agente (C-01..C-12) e mapa de contexto
OPEN_QUESTIONS.md           decisões Q-01..Q-19, com o porquê
specs/                      spec.md (v1.2) · plan.md (roadmap) · tasks.md · bdd-rules.md
docs/uml/                   UML1..UML8 em texto (Mermaid) + imagens originais e corrigidas (img/v1)
docs/                       checklist LGPD · roteiro de demonstração · relatório parcial · auditoria
features/                   cenários Gherkin + passos godog
cmd/omega-server            binário do UML8 (serve, expirar-recomendacoes) — raiz de composição
cmd/perfila-demo            demonstração do core em memória
internal/domain/            entity · port · service
internal/application/       usecase
internal/api/               rotas, handlers, middlewares (autenticação, LGPD, registro)
internal/infra/             memoria · relogio · external · config · repository/pg (+ migrações) · repository/contrato
internal/shared/errs        erros de domínio com código estável
deploy/                     Dockerfile · docker-compose.yml · nginx/ · config.env.example
scripts/                    preparar-implantacao.sh · demo-api.sh
```
