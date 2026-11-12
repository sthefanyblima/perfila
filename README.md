# Perfila

**Sistema de Segmentação e Perfilamento do Comportamento do Cliente** · Equipe Omega · Go

**Equipe:** Alexia Rodrigues, Carla Evellyn, Davi Silva, Jhenyffer Kyria, Leila Biggi, Lara Nogueira e Sthefany Barboza.

O Perfila transforma eventos de comportamento (navegação, compra, interação) em um **perfil** do cliente, atribui um **segmento** por critérios declarativos e gera uma **recomendação** de produto — sempre condicionado ao **consentimento LGPD**.

Desenvolvido com **Spec-Driven Development** (a especificação é a fonte da verdade e o agente de IA implementa) e **Behavior-Driven Development** (cenários Gherkin executáveis provam o comportamento).

## Estado (11/11/2026) — Core e Support concluídos; Generic em andamento

| Marco | Período | Estado |
|---|---|---|
| Arquivos `.md` (CLAUDE.md + UML1..UML8) | 20/10–27/10 | ✅ |
| **Core** (domínio + aplicação) — 50% | 27/10–08/11 | ✅ 79 cenários verdes · cobertura 92,2% |
| 1ª chamada de acompanhamento (relatório parcial) | 09–11/11 | `docs/RELATORIO_PARCIAL.md` |
| Support (API, middlewares, CMD) — 25% | 09–13/11 | ✅ 20 cenários de API verdes · `make serve` |
| Generic (PostgreSQL, Docker/Nginx, LGPD) — 25% | 16–19/11 | 🟡 T-19 PostgreSQL ✅ · T-24 e T-25 ⬜ |
| Fechamento do escopo | 19/11 | ⬜ |
| Ensaio · evento final | 23/11 · 24/11 | ⬜ |

## Como rodar

Requisitos: Go 1.22 ou superior.

```bash
make verify   # formatação, vet, testes com -race, BDD e teste de arquitetura
make demo     # fluxo do diagrama de sequência em memória, passo a passo
make serve    # API HTTP em modo memória na porta 8080 (token dev-token)
make demo-api # roteiro de curl contra o servidor (rode make serve em outro terminal)
make build    # binário estático bin/omega-server
make cover    # cobertura de domínio + aplicação
make contrato # suíte de contrato dos repositórios (memória; PostgreSQL se houver banco)
make pg-teste # sobe um PostgreSQL descartável no Docker e roda make verify contra ele
```

Com `DATABASE_URL` preenchida, o `omega-server` usa o PostgreSQL (`omega_db`) e aplica as migrações de `internal/infra/repository/pg/migracoes` ao subir. Vazia, roda em memória com dados sintéticos — só para desenvolvimento e demonstração local (Q-07).

Saída da demo (resumida):

```text
Sem consentimento → consentimento_invalido
{ "perfil": { "frequenciaCompras": "3", "ticketMedio": "536.67", "categoriaPreferida": "livros", ... },
  "segmento": { "nome": "premium", ... },
  "recomendacao": { "status": "ATIVA", "motivo": "segmento:premium;categoria:livros", ... } }
Após revogar → consentimento_invalido
```

## Mapa do repositório

```text
CLAUDE.md                 constituição para o agente (C-01..C-12) e mapa de contexto
specs/                    spec.md (v1.2) · plan.md (roadmap) · tasks.md · bdd-rules.md
docs/uml/                 UML1..UML8 em texto (Mermaid) + imagens originais em img/
docs/                     relatório parcial e auditoria de coerência
features/                 cenários Gherkin + passos godog
internal/domain/          entity · port · service           (Core)
internal/application/     usecase                           (Core)
internal/infra/           memoria · relogio (Core) · external · config (Support)
internal/infra/repository pg (Pg*Repository + migrações) · contrato (suíte comum) (Generic)
internal/shared/errs      erros de domínio com código estável
cmd/perfila-demo          demonstração do core
OPEN_QUESTIONS.md         dúvidas com decisão padrão e dono
```

## Arquitetura

Clean Architecture com a regra de dependência verificada por teste (`internal/arquitetura_test.go`): API → aplicação → domínio; infraestrutura implementa os ports do domínio. Detalhes em `docs/uml/UML7_componentes.md`.

## LGPD

Nenhum caso de uso que processa dado pessoal roda sem consentimento válido (C-03); a revogação tem efeito imediato (RN-03); logs só carregam IDs (C-09); testes usam dados sintéticos. Requisitos interpretados pela equipe, não parecer jurídico.
