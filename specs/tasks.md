# Tarefas (T-xx)

Perfila · Equipe Omega · Plano SDD v1.0 · atualizado em 12/11/2026 (cronograma: escopo fecha 19/11, evento 24/11)

Legenda de status: ✅ concluída · 🟡 parcial · ⬜ a fazer. "Prova" é o que precisa estar verde para a tarefa ser aceita (Definição de pronto no `specs/plan.md`, seção 3).

## Core (27/10–08/11) — 50%

| ID | Tarefa | Depende de | Prova | Status |
|---|---|---|---|---|
| T-01 | Bootstrap: `go.mod` (`perfila`), layout da Spec 3.8, `Makefile` com `verify`, CI (`.github/workflows/ci.yml`) | — | CI verde | ✅ |
| T-02 | Teste de arquitetura (C-01, C-05, C-07, C-08) | T-01 | `internal/arquitetura_test.go` fica vermelho ao violar | ✅ |
| T-03 | Harness godog (relógio fixo, fakes, mundo por cenário) | T-01 | `features/bdd_test.go` + `features/steps/` | ✅ |
| T-04 | Decisões D-01 a D-21 e Spec v1.0; UML1..UML8 em texto | — | `specs/spec.md` v1.0; `docs/uml/` | ✅ D-01 a D-23 adotadas como finais e Q-01 a Q-06 decididas (07/11); imagens corrigidas em `docs/uml/img/v1/` |
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

| T-28 | [nova] UC-09 Cadastrar Cliente: `NovoCliente`, `CadastrarClienteUseCase` (D-22) | T-13 | `cliente.feature` | ✅ |
| T-29 | [nova] Diagramas corrigidos renderizados do Mermaid (`docs/uml/img/v1/`) | T-04 | 9 imagens | ✅ |

**Portão do Core:** 79 cenários `@dominio` + `@aplicacao` verdes, teste de arquitetura verde, cobertura de domínio + aplicação 92,2% (meta ≥ 80%).

## Support — semana 1 (09–13/11) — 25%

| ID | Tarefa | Depende de | Prova | Status |
|---|---|---|---|---|
| T-20 | Gateways (`ConsentimentoGateway` local, `IntegracaoGateway` stub) e leitura de `config.env` | T-13 | testes unitários | ✅ |
| T-21 | DTOs, handlers, rotas (`net/http` 1.22) e mapeamento erro → HTTP (Spec 3.5 e 3.7) | T-14 a T-18 | `api.feature` | ✅ |
| T-22 | Middlewares de autenticação (Bearer) e LGPD | T-21 | `api.feature` (401 e 403) | ✅ |
| T-23 | CMD `omega-server`: raiz de composição; subcomandos `serve` e `expirar-recomendacoes` | T-20 a T-22, T-27 | `cmd/omega-server/main_test.go` + `make serve` / `make demo-api` | ✅ |

**Portão do Support:** `api.feature` com 20 cenários verdes (`@api` ativo no filtro); total 99 cenários / 668 passos; build e testes conferidos também com Go 1.22 (versão do CI). Concluído em 07/11, adiantado.

## Generic — semana 2 (16–19/11) — 25%

| ID | Tarefa | Depende de | Prova | Status |
|---|---|---|---|---|
| T-19 | Migrações SQL, sete `Pg*Repository`, suíte de contrato rodando contra fake e PostgreSQL | T-13 | contrato verde nos dois backends | ✅ 11/11: migrações 001 (esquema + `padrao`) e 002 (dados sintéticos, Q-08) aplicadas pelo servidor com trava (Q-09); contrato com 7 repositórios verde em memória e PostgreSQL 16; fluxo ponta a ponta também sobre o banco; CI com serviço PostgreSQL |
| T-24 | Dockerfile multiestágio, `docker-compose.yml` (Nginx + 2 réplicas + PostgreSQL), `nginx.conf` | T-19, T-23 | `implantacao.feature` (`@slow`) | ✅ 11/11: Nginx 1.28 com TLS autoassinado (Q-12) e `X-Omega-Replica` (Q-11), `app1`/`app2`, PostgreSQL em rede interna; segredos gerados (Q-14); 3 cenários `@implantacao` verdes; job `implantacao` no CI (Q-13). Sem corte: duas réplicas no MVP |
| T-25 | Checklist LGPD, auditoria de logs sem dados pessoais, README final, roteiro de demonstração | T-24 | checklist concluído | ✅ 11/11: `docs/CHECKLIST_LGPD.md` (L-01..L-20, com evidências e pendências pós-MVP); auditoria de logs encontrou e corrigiu `time.Now` no log de acesso da API (C-07, teste de arquitetura estendido a todo `internal/`); e-mail em query string coberto pelo teste de C-09; README final; `docs/ROTEIRO_DEMONSTRACAO.md` e `make demo-implantacao` |

| T-30 | [nova] Correções do mapeamento de bugs (Q-15 a Q-19): limite de `Valor` e resposta nunca "200 vazio"; tolerância de relógio de 2 min; critério de segmento inválido e segmento inexistente como falha de configuração (500); uma recomendação ATIVA por cliente (`ListarAtivasPorCliente`, migração 003); `/saude` com o banco; corpo JSON único | T-25 | 8 cenários novos (107 no verify); contrato nos dois backends; sondagens repetidas contra a implantação | ✅ 11/11 |

Filtro padrão do BDD: `~@slow && ~@wip`; `@implantacao` roda em `make bdd-implantacao` e no job `implantacao` do CI (Q-13).
Corte planejado das réplicas: **não foi necessário** (T-24 concluída em 11/11 com duas réplicas).

## Painel web (12–19/11) — Spec v1.3

Interface do operador a partir do design system "Perfila Console" (`docs/design/`), com os ajustes da Q-27. Decisões Q-20 a Q-29.

| ID | Tarefa | Depende de | Prova | Status |
|---|---|---|---|---|
| T-31 | Spec v1.3: Q-20 a Q-29, casos de uso do painel, ports novos, rotas 3.7.1, configuração `PAINEL_*`, UML7 e UML8 em texto, `features/painel.feature` (`@web`, vermelha) | T-30 | Spec, UML e feature revisados | ✅ |
| T-32 | `ProcessarClienteUseCase` (RN-09 com estado por etapa) em `application/usecase`; a API passa a usá-lo | T-31 | `@api` verde; cenários `@aplicacao` do processamento | ✅ `processamento.feature` (4 cenários) |
| T-33 | `ClienteRepository.Listar`, `RecomendacaoRepository.ListarPorCliente`, `ProdutoRepository.ListarTodos` na memória e no PostgreSQL; migração 004 | T-31 | contrato nos dois backends | ✅ busca, paginação, histórico e catálogo no contrato (memória e PostgreSQL 16) |
| T-34 | Casos de uso de leitura: `ListarClientes`, `ConsultarCliente`, `ConsultarResultado` (RN-01), `ListarSegmentos`, `ListarCatalogo` | T-33 | cenários `@aplicacao` | ✅ `leitura.feature` (17 cenários) |
| T-35 | `internal/web` base: rotas, sessão HMAC, CSRF, limite de tentativas, cabeçalhos e CSP, Entrar, layout, static embutido, regra "web" no teste de arquitetura, `PAINEL_*`, `hash-senha`, montagem no `omega-server` | T-34 | cenários `@web` de sessão e segurança | ✅ 9 cenários `@web`; testes de sessão, limitador e formatos |
| T-36 | Telas Clientes (busca, filtro, paginação) e Cadastrar cliente | T-35 | cenários `@web` de cadastro e busca | ✅ 4 cenários `@web` |
| T-37 | Ficha do cliente: cabeçalho com consentimento, Visão geral (etapas, perfil, segmento, recomendação), Eventos e Consentimento, com diálogos; sem JavaScript (PRG) e com htmx | T-36 | cenários `@web` de consentimento, processamento e eventos | ✅ 12 cenários `@web`; fluxo A conferido no navegador. T-35 a T-37 num só commit: as telas dividem layout, partes e funções de template |
| T-38 | Segmentos, Catálogo e Operação; `@implantacao` confere `/painel/entrar` pelo Nginx; README, roteiro e checklist LGPD; acessibilidade | T-37 | `make bdd-implantacao`; fluxos A a E no navegador | ✅ 4 cenários `@implantacao` contra a implantação no Docker; contraste AA de todos os pares de cor nos dois temas (mínimo 5,3:1); README, roteiro e checklist LGPD (L-21 a L-27) |
| T-39 | [nova] Clientes de exemplo, um por situação do painel (Q-30): carga pelos casos de uso, automática no modo memória e por `carregar-exemplos` no PostgreSQL | T-38 | `TestCarregarExemplos` | ✅ |
| T-40 | [nova] Cliente inativo só para leitura (Q-31): `Precondicoes.VerificarLeitura`, ficha com aviso e sem ação; revogado diferente de não concedido e linha do tempo do consentimento | T-39 | cenários `@aplicacao` e `@web` do inativo | ✅ |
| T-41 | [nova] Demonstração ampliada (Q-32): migração 005 (segmentos e catálogo) e 14 clientes com história | T-40 | `TestMigracoes`, `TestCarregarExemplos` | ✅ |

## Histórico do consentimento — Spec v1.4

Prova de cada concessão e revogação (Q-34, RN-10, L-06). Base para os públicos e o portal do titular.

| ID | Tarefa | Depende de | Prova | Status |
|---|---|---|---|---|
| T-42 | Spec v1.4, UML2 e UML5 em texto, Q-34; cenários novos em `consentimento.feature` e `painel.feature` (`@wip` até a tarefa que os implementa) | T-41 | Spec, UML e features revisados | ✅ 12 cenários novos (`@wip`) |
| T-43 | `MovimentoConsentimento` e `NovoMovimento`; `ConsentimentoRepository.Registrar` e `ListarMovimentos` na memória e no PostgreSQL; migração 006 com reconstrução | T-42 | testes de tabela; contrato nos dois backends; `TestMigracoes` | ✅ `TestNovoMovimento`, contrato `MovimentoConsentimento` e `TestMigracaoReconstroiMovimentos` (memória e PostgreSQL 16) |
| T-44 | UC-08 grava o movimento e expõe `Historico`; origem pelo contexto (`ComOrigem`) na API; `ConsultarCliente` devolve o histórico | T-43 | cenários `@aplicacao` do histórico; teste da origem na API | ✅ 8 cenários `@aplicacao`; `TestConsentimentoPelaAPIRegistraOrigem` |
| T-45 | Painel: origem `painel` com o operador; linha do tempo pelo histórico, com a linha pelas datas para quem não tem movimentos | T-44 | cenários `@web` do histórico | ✅ 4 cenários `@web`; `TestDetalheMovimento`, `TestLinhaPeloHistorico` |
| T-46 | Exemplos (Karina: v1, revoga, v2) conferidos pelo histórico; checklist L-06; README; fluxo no navegador | T-45 | `TestCarregarExemplos`; painel conferido | ✅ Karina com v1 pela API, revogação e v2 pelo painel; L-06 atendido; fluxo conferido no navegador; cobertura 91,0% |

## Usuários, perfis de acesso e auditoria — Spec v1.5

Operadores no banco com papel (ADMINISTRADOR, ANALISTA, ATENDIMENTO), matriz de permissões e trilha de auditoria (Q-35 a Q-41, L-12, L-23). Aprovadas em 13/11.

| ID | Tarefa | Depende de | Prova | Status |
|---|---|---|---|---|
| T-47 | Spec v1.5 (atores, entidades, casos de uso, ports, rotas, configuração), UML1, UML2, UML7 e UML8 em texto, Q-35 a Q-41; `usuarios.feature` e cenários novos em `painel.feature` (`@wip`) | T-46 | Spec, UML e features revisados | ✅ 54 cenários novos (`@wip`): 30 `@aplicacao` em `usuarios.feature`, 24 `@web` em `painel.feature` |
| T-48 | `Operador`, `PapelOperador`, matriz `Pode`, `RegistroAuditoria`; ports `OperadorRepository`, `AuditoriaRepository`, `HashSenha`; memória e contrato | T-47 | testes de tabela; contrato em memória | ✅ `TestPapelPode`, `TestNovoOperador`, `TestOperadorAlteracoes`, `TestNovoRegistroAuditoria`; contrato `Operador` e `Auditoria` em memória |
| T-49 | PostgreSQL: migração 007, `PgOperadorRepository`, `PgAuditoriaRepository`; bcrypt em `infra/senha` | T-48 | contrato nos dois backends; `TestMigracoes` | ✅ `TestPapelPode`, `TestNovoOperador`, `TestOperadorAlteracoes`, `TestNovoRegistroAuditoria`; contrato `Operador` e `Auditoria` em memória |
| T-50 | Casos de uso `AutenticarOperador`, `GerenciarOperadores`, `RegistrarAuditoria`, `ListarAuditoria`; administrador inicial | T-48 | cenários `@aplicacao` | ✅ `TestPapelPode`, `TestNovoOperador`, `TestOperadorAlteracoes`, `TestNovoRegistroAuditoria`; contrato `Operador` e `Auditoria` em memória |
| T-51 | Painel: Entrar pelo caso de uso, sessão com o ID do operador conferida a cada requisição, permissão por rota (403), menu por papel, auditoria das ações e das leituras de ficha | T-50 | cenários `@web` de papéis e sessão | ✅ `usuarios.feature` verde (`@aplicacao`); cobertura de domínio + aplicação 88,2%. `TrocarSenha` devolve o operador para o painel renovar a sessão (Spec 3.5 ajustada) |
| T-52 | Telas Usuários (criar, mudar papel, desativar, ativar, redefinir senha) e Minha senha | T-51 | cenários `@web` de usuários | ✅ 11 cenários `@web` novos (menu por papel, ação sem permissão, sessão encerrada, mudança de papel, auditoria das ações e leituras, auditoria sem dado pessoal); `TestAdministradorInicial` |
| T-53 | Tela Auditoria (filtros e paginação) e subcomando `expurgar-auditoria` | T-51 | cenários `@web` da auditoria; `TestExpurgarAuditoria` | ✅ 11 cenários `@web` novos (menu por papel, ação sem permissão, sessão encerrada, mudança de papel, auditoria das ações e leituras, auditoria sem dado pessoal); `TestAdministradorInicial` |
| T-54 | Exemplos (um operador por papel, só no modo memória), README, roteiro, checklist (L-12, L-23, L-28); fluxo no navegador | T-52, T-53 | `TestCarregarExemplos`; painel conferido | ⬜ |

## Fechamento

| Data | Marco |
|---|---|
| 19/11 | Escopo fecha: depois disso, só correções |
| 20–22/11 | Correções |
| 23/11 | Ajustes finais e ensaio da apresentação |
| 24/11 | Evento final |

Trello: cards atualizados toda sexta (13/11, 20/11).
