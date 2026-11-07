# Tarefas (T-xx)

Perfila · Equipe Omega · Plano SDD v1.0 · atualizado em 07/11/2026 (cronograma: escopo fecha 19/11, evento 24/11)

Legenda de status: ✅ concluída · 🟡 parcial · ⬜ a fazer. "Prova" é o que precisa estar verde para a tarefa ser aceita (Definição de pronto no `specs/plan.md`, seção 3).

## Core (27/10–08/11) — 50%

| ID | Tarefa | Depende de | Prova | Status |
|---|---|---|---|---|
| T-01 | Bootstrap: `go.mod` (`perfila`), layout da Spec 3.8, `Makefile` com `verify`, CI (`.github/workflows/ci.yml`) | — | CI verde | ✅ |
| T-02 | Teste de arquitetura (C-01, C-05, C-07, C-08) | T-01 | `internal/arquitetura_test.go` fica vermelho ao violar | ✅ |
| T-03 | Harness godog (relógio fixo, fakes, mundo por cenário) | T-01 | `features/bdd_test.go` + `features/steps/` | ✅ |
| T-04 | Decisões D-01 a D-21 e Spec v1.0; UML1..UML8 em texto | — | `specs/spec.md` v1.0; `docs/uml/` | 🟡 decisões padrão aplicadas; **equipe** valida na 1ª chamada e corrige as imagens até 19/11 |
| T-05 | Erros sentinelas e tipos enumerados | T-04 | `internal/shared/errs`, `entity/tipos.go` | ✅ |
| T-06 | `Cliente` e invariantes | T-05 | `entity_test.go` | ✅ |
| T-07 | `ConsentimentoLGPD` e máquina de estados | T-05 | `consentimento.feature`, `TestConsentimentoTransicoes` | ✅ |
| T-08 | `EventoComportamental.Validar` | T-05 | `evento.feature` (RN-04), `TestEventoValidar` | ✅ |
| T-09 | `PerfilamentoService` (RN-05) | T-08 | `perfilamento.feature`, `TestCalcularScore` | ✅ |
| T-10 | Avaliador de critério, `Segmento.ClassificarPerfil`, `SegmentacaoService` | T-09 | `segmentacao.feature`, `TestAvaliarCriterio` | ✅ |
| T-11 | `RecomendacaoProduto`, estados e `RecomendacaoService` | T-10 | `recomendacao_ciclo.feature`, `TestRecomendacaoTransicoes` | ✅ |
| T-12 | `ValidadorLGPD` | T-07 | cenários `@lgpd` | ✅ |
| T-13 | Ports (Spec 3.6) e fakes em memória | T-12 | compila com `var _ port.X`; usados por todo o BDD | ✅ |
| T-14 | `GerenciarConsentimentoUseCase` | T-13 | `consentimento.feature` | ✅ |
| T-15 | `ColetarDadosUseCase` (Executar e Registrar) | T-13 | `evento.feature` | ✅ |
| T-16 | `GerarPerfilUseCase` | T-15, T-09 | `perfilamento.feature` | ✅ |
| T-17 | `SegmentarClienteUseCase` | T-16, T-10 | `segmentacao.feature` | ✅ |
| T-18 | `GerarRecomendacaoUseCase` | T-17, T-11 | `recomendacao.feature` | ✅ |
| T-26 | [nova] Demo do fluxo da sequência em memória (`cmd/perfila-demo`) | T-18 | `make demo` | ✅ |
| T-27 | [nova] `ExpirarRecomendacoesUseCase` (D-08) | T-11, T-13 | `recomendacao_ciclo.feature` (comando) | ✅ |

**Portão do Core:** 73 cenários / 476 passos verdes (`@dominio` + `@aplicacao`), teste de arquitetura verde, cobertura de domínio + aplicação 92,5% (meta ≥ 80%).

## Support — semana 1 (09–13/11) — 25%

| ID | Tarefa | Depende de | Prova | Status |
|---|---|---|---|---|
| T-20 | Gateways (`ConsentimentoGateway` local, `IntegracaoGateway` stub) e leitura de `config.env` | T-13 | testes unitários | ⬜ |
| T-21 | DTOs, handlers, rotas (`net/http` 1.22) e mapeamento erro → HTTP (Spec 3.5 e 3.7) | T-14 a T-18 | `api.feature` | ⬜ |
| T-22 | Middlewares de autenticação (Bearer) e LGPD | T-21 | `api.feature` (401 e 403) | ⬜ |
| T-23 | CMD `omega-server`: raiz de composição; subcomandos `serve` e `expirar-recomendacoes` | T-20 a T-22, T-27 | execução local ponta a ponta (fakes) | ⬜ |

Ao iniciar: remover `~@api` de `tagsPadrao` em `features/bdd_test.go`.

## Generic — semana 2 (16–19/11) — 25%

| ID | Tarefa | Depende de | Prova | Status |
|---|---|---|---|---|
| T-19 | Migrações SQL, sete `Pg*Repository`, suíte de contrato rodando contra fake e PostgreSQL | T-13 | contrato verde nos dois backends | ⬜ |
| T-24 | Dockerfile multiestágio, `docker-compose.yml` (Nginx + 2 réplicas + PostgreSQL), `nginx.conf` | T-19, T-23 | `implantacao.feature` (`@slow`) | ⬜ |
| T-25 | Checklist LGPD, auditoria de logs sem dados pessoais, README final, roteiro de demonstração | T-24 | checklist concluído | 🟡 README e demo do core prontos |

Ao iniciar: remover `~@implantacao` de `tagsPadrao` (rodar `@slow` em job separado do CI).
Corte planejado: se até 17/11 a T-24 estiver atrasada, as duas réplicas saem do MVP (uma réplica atrás do Nginx).

## Fechamento

| Data | Marco |
|---|---|
| 19/11 | Escopo fecha: depois disso, só correções |
| 20–22/11 | Correções |
| 23/11 | Ajustes finais e ensaio da apresentação |
| 24/11 | Evento final |

Trello: cards atualizados toda sexta (13/11, 20/11).
