# Relatório Parcial — Perfila

**Sistema de Segmentação e Perfilamento do Comportamento do Cliente**  
Equipe Omega · GoLang & CLP · 1ª chamada de acompanhamento: **09–11/11/2026**

**Equipe:** Alexia Rodrigues, Carla Evellyn, Davi Silva, Jhenyffer Kyria, Leila Biggi, Lara Nogueira e Sthefany Barboza.

## 1. Objetivo

Construir, em Go, um sistema que coleta eventos de comportamento do cliente, calcula um perfil, atribui um segmento e recomenda um produto, respeitando a LGPD: nenhum dado pessoal é processado sem consentimento válido. O desenvolvimento usa um agente de IA generativa (Claude Code) guiado por especificação (SDD) e validado por cenários executáveis (BDD).

## 2. Roadmap e situação

| Período | Marco | Peso | Situação em 07/11 |
|---|---|---|---|
| 20/10–27/10 | Arquivos `.md`: `CLAUDE.md` + `UML1.md`…`UML8.md`, SDD e BDD | — | Concluído |
| 27/10–08/11 | **Core**: domínio + aplicação | 50% | **Concluído** |
| 09–11/11 | 1ª chamada de acompanhamento | — | Este relatório |
| 09–13/11 | Support: API HTTP, middlewares, CMD, gateways | 25% | **Concluído (adiantado)** |
| 16–19/11 | Generic: PostgreSQL, autenticação, Docker/Nginx | 25% | A iniciar |
| 19/11 | Fechamento do escopo (depois, só correções) | — | — |
| 23/11 · 24/11 | Ensaio · evento final | — | — |

Classificação DDD usada para dividir o trabalho:

- **Core** — Perfilamento, Segmentação e Recomendação (o diferencial do produto).
- **Support** — Consentimento LGPD, Coleta de eventos, Catálogo (necessários, não diferenciais). O *modelo* desses subdomínios já está pronto, porque o core depende dele; falta expô-los pela API.
- **Generic** — Autenticação, Persistência, Implantação, Logs (soluções de mercado).

## 3. Metodologia

**Spec-Driven Development.** A modelagem UML (8 diagramas, Módulos 2 a 5) foi transcrita em artefatos que o agente consome:

| Artefato | Papel |
|---|---|
| `CLAUDE.md` | Constituição: 12 regras invioláveis (dependências, consentimento primeiro, nomes fiéis, sem `time.Now()`, sem estado global, logs sem dado pessoal…) |
| `docs/uml/UML1..UML8_*.md` | Cada diagrama em Mermaid, com regras para o agente e divergências registradas |
| `specs/spec.md` | O que o sistema faz: 9 regras de negócio, 2 máquinas de estado, contratos, catálogo de erros, 21 decisões de modelagem |
| `specs/plan.md`, `specs/tasks.md` | Fases, portões, 27 tarefas pequenas, uma por vez |
| `specs/bdd-rules.md`, `features/*.feature` | 15 regras de escrita e os critérios de aceite em Gherkin (português) |

**Ciclo por tarefa:** contexto mínimo (CLAUDE.md + seções da spec + UML + features da tarefa) → testes primeiro (vermelho) → código (verde) → `make verify` → revisão com checklist. O agente não pode editar `specs/`, `docs/uml/` nem `.feature`; dúvidas vão para `OPEN_QUESTIONS.md`.

**BDD:** cenários em português (`# language: pt`) executados pelo godog; relógio fixo e repositórios em memória tornam cada cenário determinístico e independente.

## 4. O que foi entregue no Core

**Domínio** (`internal/domain`)

- 7 entidades do Diagrama de Classes, com tipos enumerados para estados.
- Máquina de estados do **ConsentimentoLGPD** (NAO_CONCEDIDO ⇄ CONCEDIDO → REVOGADO) e da **RecomendacaoProduto** (GERADA → ATIVA → EXPIRADA).
- Avaliador próprio de critérios de segmento (gramática fechada, sem reflexão nem execução dinâmica): `ticketMedio >= 500 && frequenciaCompras >= 3`.
- Serviços: Perfilamento (frequência, ticket médio, categoria preferida, score de engajamento em janela de 30 dias), Segmentação (por prioridade), Recomendação, Validador LGPD.
- 10 ports (7 repositórios, 2 gateways, Clock).

**Aplicação** (`internal/application/usecase`)

- Os 5 casos de uso do MVP (UC-01, UC-02, UC-03, UC-04, UC-08) + expiração de recomendações.
- Pré-condições comuns (RN-01): cliente existe → ativo → consentimento válido.

**Infraestrutura de teste e demonstração**

- Repositórios em memória (implementam os mesmos ports que o PostgreSQL implementará).
- `make demo`: o fluxo do Diagrama de Sequência de ponta a ponta, inclusive o bloqueio por falta de consentimento e a revogação com efeito imediato.
- **UC-09 Cadastrar Cliente** (D-22): a modelagem não dizia como um cliente entra no sistema; o cliente agora nasce ativo e sem consentimento.
- Diagramas corrigidos renderizados a partir do Mermaid (`docs/uml/img/v1/`).

**Support (adiantado, concluído em 07/11)**

- API HTTP (`net/http`): cadastro, consentimento, eventos e o processamento consolidado do passo 23 da sequência.
- Middlewares: autenticação Bearer (401) e LGPD com as pré-condições RN-01 (404, 409, 403).
- `omega-server` com `serve` e `expirar-recomendacoes`, configuração por ambiente ou `config.env`.
- Sem banco ainda: o servidor roda em memória com dados sintéticos (D-23); `make serve` + `make demo-api` mostram o fluxo por HTTP.

## 5. Resultados medidos

| Indicador | Meta (portão) | Resultado |
|---|---|---|
| Cenários BDD `@dominio` + `@aplicacao` | 100% verdes | **79/79** |
| Cenários BDD `@api` (Support) | 100% verdes | **20/20** (total 99 cenários, 668 passos) |
| Cobertura de domínio + aplicação | ≥ 80% | **92,2%** |
| Teste de arquitetura (C-01, C-05, C-07, C-08) | verde | **verde** (e comprovadamente vermelho ao violar) |
| `go vet`, `gofmt`, testes com `-race` | sem achados | **sem achados** |
| Teste de mutação manual (validade da recomendação alterada) | cenários detectam | **3 cenários falharam**, como esperado |

## 6. Coerência da modelagem

A auditoria (`docs/AUDITORIA_COERENCIA.md`) encontrou 16 pontos. Os documentos eram coerentes entre si nas regras; os ajustes foram:

- Alinhar os arquivos ao quadro (`CLAUDE.md` + `UML1..8.md`) e ao roadmap com datas.
- Fidelidade aos diagramas: nome `ClassificarPerfil`, atores de UC-08 (o diagrama traz o Administrador), direção do «include» (o diagrama de casos de uso estava certo; a sequência é que inverte).
- Cobertura BDD: cenários para cliente inexistente no consentimento, perfil sem segmento, valores-limite de evento e critérios válidos.

Principais decisões de modelagem mantidas: consentimento com 3 estados em vez de `bool` (D-02); ports no domínio, e não na infraestrutura (D-10); catálogo de produtos adicionado para viabilizar a recomendação (D-05); recomendação gravada uma única vez, já ATIVA (D-08).

## 7. Riscos e próximos passos

| Risco | Mitigação |
|---|---|
| Imagens dos diagramas ainda divergem da Spec v1.0 | Corrigir até 19/11 (lista em D-xx) |
| Perguntas Q-01 a Q-06 sem resposta | Cada uma tem padrão implementado e dono |
| PostgreSQL pode divergir dos fakes | Suíte de contrato única rodando contra os dois (T-19) |

**Support:** concluído antes do prazo; `api.feature` com 20 cenários verdes.
**Generic (16–19/11):** migrações e 7 repositórios PostgreSQL, Docker multiestágio, Nginx com 2 réplicas, checklist LGPD. Portão: `implantacao.feature` verde. Se a semana apertar, as duas réplicas saem do MVP (decisão até 17/11).

**19/11:** escopo fecha. **23/11:** ensaio. **24/11:** evento final.
