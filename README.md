# Perfila

**Sistema de Segmentação e Perfilamento do Comportamento do Cliente** · Equipe Omega · Go 1.22

**Equipe:** Alexia Rodrigues, Carla Evellyn, Davi Silva, Jhenyffer Kyria, Leila Biggi, Lara Nogueira e Sthefany Barboza.

O Perfila transforma eventos de comportamento (navegação, compra, interação) em um **perfil** do cliente, atribui um **segmento** por critérios declarativos e gera uma **recomendação** de produto, sempre condicionado ao **consentimento LGPD**.

Desenvolvido com **Spec-Driven Development** (a especificação é a fonte da verdade e o agente de IA implementa) e **Behavior-Driven Development** (cenários Gherkin executáveis provam o comportamento).

## Estado (19/11/2026): MVP concluído + painel, portal do titular, públicos e auditoria; escopo fechado, só correções

| Marco | Período | Estado |
|---|---|---|
| Arquivos `.md` (CLAUDE.md + UML1..UML8) | 20/10–27/10 | ✅ |
| **Core** (domínio + aplicação), 50% | 27/10–08/11 | ✅ 79 cenários · cobertura 92,5% |
| 1ª chamada de acompanhamento | 09–11/11 | ✅ `docs/RELATORIO_PARCIAL.md` |
| **Support** (API, middlewares, CMD, gateways), 25% | 09–13/11 | ✅ 20 cenários de API |
| **Generic** (PostgreSQL, Docker/Nginx, LGPD), 25% | 16–19/11 | ✅ contrato nos dois backends · 3 cenários de implantação · checklist LGPD |
| **Painel web do operador** (Spec v1.3, T-31 a T-38) | 12/11 | ✅ 26 cenários `@web` · 21 de aplicação novos · 1 de implantação |
| **Histórico do consentimento** (Spec v1.4, T-42 a T-46, Q-34) | 12/11 | ✅ 8 cenários de aplicação · 4 `@web` · contrato e migração 006 nos dois backends |
| **Edição de segmentos e catálogo** (Spec v1.6, T-55 a T-59, Q-33, Q-42 a Q-44) | 13/11 | ✅ 15 cenários de aplicação e domínio · 11 `@web` · contrato e migração 008 nos dois backends |
| **Importação de eventos em lote** (Spec v1.8, T-65 a T-68, Q-47) | 13/11 | ✅ 34 cenários de aplicação · 4 `@api` · 7 `@web` |
| **Direitos do titular e retenção** (Spec v1.9, T-69 a T-73, Q-48 a Q-53) | 13–14/11 | ✅ 36 cenários de aplicação · 10 `@web` · contrato e migração 009 nos dois backends |
| **Portal do titular** (Spec v1.10, T-75 a T-79, Q-55 a Q-60) | 14/11 | ✅ 51 cenários (29 de aplicação e domínio, 21 `@web`) · 1 de implantação · contrato e migração 010 nos dois backends |
| **Públicos para campanhas** (Spec v1.11, T-80 a T-85, Q-61 a Q-67) | 13–14/11 | ✅ consentimento por finalidade · 36 cenários de aplicação, domínio e API · 10 `@web` · contrato e migração 011 nos dois backends |
| **Auditoria dos pedidos do titular, da importação e dos públicos** (Spec v1.12, T-86 a T-89, Q-68, Q-69) | 16/11 | ✅ 8 ações novas na trilha (RN-13) · 26 cenários novos · migração 012 · contrato nos dois backends |
| **Revisão de 19/11** (PR #14 e #15, Q-70) | 19/11 | ✅ categoria dos eventos normalizada · e-mail validado no cadastro · identidade visual do painel · menu do celular como gaveta e página de erro com menu |
| Fechamento do escopo | 19/11 | daqui em diante, só correções |
| Ensaio · evento final | 23/11 · 24/11 | roteiro em `docs/ROTEIRO_DEMONSTRACAO.md` |

**513 cenários BDD** (508 no `make verify` + 5 de implantação), teste de arquitetura e suíte de contrato dos repositórios em memória e em PostgreSQL.

## Arquitetura

Clean Architecture (UML7), com a regra de dependência verificada por teste (`internal/arquitetura_test.go`):

```text
api ─┐
     ├─► application ──► domain ◄── infra        shared/errs: importável por todos
web ─┘   (casos de uso)  (entity, port, service)  (memoria, relogio, repository/pg, external, config)
```

Implantação (UML8): o **Nginx** (TLS, única porta de entrada) distribui para **duas réplicas** do `omega-server`, que acessam o **PostgreSQL** `omega_db` numa rede interna, inalcançável pelo dispositivo do cliente.

## Como rodar

### Desenvolvimento (Go 1.22+)

```bash
make verify   # gofmt, vet, testes com -race, 508 cenários BDD e teste de arquitetura
make demo     # fluxo do diagrama de sequência em memória, passo a passo
make serve    # API e painel em modo memória: http://localhost:8080/painel (admin / dev-senha-admin)
make demo-api # roteiro de curl contra o make serve (em outro terminal)
make cover    # cobertura de domínio + aplicação
make pg-teste # sobe um PostgreSQL descartável no Docker e roda o verify contra ele
make build    # binário estático bin/omega-server
```

### Implantação (Docker com Compose v2)

```bash
make implantar         # gera segredos (1ª vez) e sobe nginx + app1 + app2 + db
make demo-implantacao  # roteiro da API pelo Nginx em HTTPS, gravando no PostgreSQL
make bdd-implantacao   # 5 cenários @implantacao (inclui parar uma réplica, o painel e o portal do titular pelo Nginx)
make expirar           # administrador: expira recomendações vencidas
make retencao          # administrador: apaga o comportamento de quem revogou há 180 dias ou mais (RN-18)
make exemplos          # carrega os 75 clientes de exemplo, 15 com história e 60 de volume (só com o banco sem clientes; nunca operadores)
make derrubar          # derruba (os dados ficam no volume omega_db)
```

- A única porta publicada é `https://localhost:8443`, com certificado autoassinado gerado na primeira subida.
- O token da API fica em `deploy/config.env` e a senha do banco em `deploy/.env`. Os dois são gerados aleatoriamente e ficam fora do git.
- O painel fica em `https://localhost:8443/painel`. Na primeira subida, sem nenhum operador no banco, o usuário `operador` vira o administrador inicial (Q-35); a senha aleatória está em `deploy/.env` (`grep PAINEL_SENHA deploy/.env`) e só o hash bcrypt vai para o `config.env` (Q-28). Os demais operadores são criados por ele na tela Usuários.
- A auditoria do painel guarda 6 meses (`AUDITORIA_RETENCAO_DIAS`); agende `omega-server expurgar-auditoria` como a expiração de recomendações (Q-39).
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
| `PAINEL_CHAVE_SESSAO` | vazia = painel desligado | chave HMAC da sessão, 64 dígitos hexadecimais; liga o painel (Q-22, Q-35) |
| `PAINEL_USUARIO`, `PAINEL_SENHA_HASH` | operador, vazia | administrador inicial, criado só com o banco sem operadores; hash bcrypt de `omega-server hash-senha` (12 caracteres no mínimo). Com operadores gravados, são ignorados (Q-35) |
| `AUDITORIA_RETENCAO_DIAS` | 180 | retenção da auditoria do painel (RN-13, Q-39) |
| `PAINEL_VERSAO_TERMO` | v1 | versão do termo sugerida ao conceder (Q-25); no portal, é o termo vigente que o titular aceita. `PAINEL_FINALIDADE` deixou de existir na v1.11 (a finalidade vem do botão, Q-61) |
| `PUBLICOS_ARQUIVO_HORAS` | 24 | por quanto tempo o arquivo de público pode ser baixado, de 1 a 168 (Q-64) |
| `RETENCAO_DIAS_REVOGADO` | 180 | prazo da retenção após a revogação, em dias (RN-18, Q-51) |
| `PORTAL_URL_BASE`, `PORTAL_CHAVE_SESSAO` | vazias = portal desligado | endereço público usado no link de acesso (o do Nginx) e chave HMAC da sessão do titular, diferente da do painel (Q-60) |

O ambiente prevalece sobre o arquivo. Exemplo em `deploy/config.env.example`.

## API (Spec 3.7)

| Rota | Caso de uso | Exige |
|---|---|---|
| `POST /v1/clientes` | UC-09 Cadastrar | token |
| `PUT /v1/clientes/{id}/consentimento` | UC-08 Conceder (`finalidade`: `segmentacao` ou `publicidade_terceiros`) | token |
| `POST /v1/clientes/{id}/consentimento/revogacao` | UC-08 Revogar (sem corpo: `segmentacao`; ou `{"finalidade": ...}`) | token |
| `GET /v1/clientes/{id}/consentimento` | UC-08 Consultar (a `segmentacao` no topo e a lista `finalidades`) | token |
| `POST /v1/clientes/{id}/eventos` | UC-01 Registrar evento | token + consentimento |
| `POST /v1/clientes/{id}/processamento` | UC-01 → UC-02 → UC-03 → UC-04 (RN-09) | token + consentimento |
| `POST /v1/eventos/lote` (`?previa=true` só confere) | lote de até 1.000 eventos em JSON, linha a linha (RN-16) | token; consentimento conferido em cada linha |
| `GET /saude` | saúde, para o balanceador | — |

Erros: JSON `{codigo, mensagem}` com códigos estáveis (Spec 3.5): 401 sem token, 403 `consentimento_invalido`, 404 `cliente_nao_encontrado`, 409 `cliente_inativo` / `cliente_eliminado` / `transicao_invalida` / `perfil_nao_encontrado`, 422 `dados_invalidos`.

## Painel do operador (Spec 3.7.1)

Interface web em `/painel`, servida pelo próprio `omega-server`, seguindo o design system "Perfila Console" (`docs/design/`) com cantos suaves, neutros cinza, primária petróleo claro, acento ameixa e verde positivo próprio (Q-27). HTML gerado no servidor (`html/template`), CSS próprio a partir dos tokens, htmx 2.0.3 e fontes IBM Plex embutidos no binário: nenhuma requisição a terceiros.

| Tela | O que faz |
|---|---|
| Clientes | busca por nome (sem acentos) ou ID, filtro de consentimento, 20 por página, e-mail mascarado |
| Cadastrar cliente | UC-09; abre a ficha já com o selo "Não concedido" |
| Ficha | selo de consentimento sempre visível; Visão geral (etapas do processamento, perfil, segmento com os critérios atendidos, recomendação), Eventos e Consentimento. Sem consentimento, nenhum dado de comportamento é lido |
| Consentimento | registro atual e linha do tempo com cada concessão e revogação: finalidade, versão do termo e origem (API, painel com o operador, portal). Só inclusão; cliente sem movimentos mostra a linha pelas datas (Q-34) |
| Direitos do titular | aba da ficha, com qualquer estado do consentimento: corrigir nome e e-mail, baixar tudo o que o sistema guarda do titular em JSON e eliminar os dados (pede a revogação antes, se concedido, e o nome digitado). A eliminação apaga eventos, perfil e recomendações, anonimiza o cadastro e guarda o histórico do consentimento como prova (RN-17) |
| Importar eventos | CSV com cabeçalho (`cliente_id,tipo,origem,timestamp,valor,categoria`), até 1.000 linhas e 256 KiB; prévia sem gravar, confirmação e relatório por linha. Cada linha passa por RN-01 e RN-04; duplicados (cliente, tipo, instante e valor) são ignorados, então reenviar é seguro. O arquivo não fica guardado (RN-16, Q-47). Modelo em "Baixar modelo"; exemplo pronto para a demonstração em `docs/exemplos/eventos-lote.csv` |
| Relatórios | UC-05 e UC-07 parciais (Q-46): clientes, consentimento por estado e taxa de concessão, distribuição por segmento (clientes, ticket médio, score médio), recomendações ativas e expiradas, categorias mais compradas e evolução mensal de cadastros, eventos e revogações, nos últimos 6 ou 12 meses. Gráficos SVG desenhados no servidor, cada um com a tabela de dados, e exportação em CSV. Só números agregados: comportamento conta só quem tem consentimento válido, e grupos com menos de 5 clientes aparecem como "menos de 5" |
| Consentimento por finalidade (v1.11) | um registro por finalidade: **Segmentação** (perfil, segmento e recomendação) e **Publicidade com terceiros** (entrar nos públicos de anúncio), cada um com selo, ações e linha do tempo (Q-61) |
| Públicos (v1.11) | escolher o segmento, ver quantos entram e quantos ficam de fora, gerar com confirmação e baixar o CSV (só o e-mail em SHA-256 e o nome do segmento), remoções pendentes depois de uma revogação e histórico de exportações (RN-21) |
| Relatório agregado (v1.11) | por segmento, quantidade, ticket médio e categoria mais comum, sem identificador; grupos com menos de 5 aparecem como "menos de 5" (RN-22) |
| Segmentos | ativos por prioridade e inativos. Quem edita (Administrador e Analista) cria, edita, ativa e desativa: o critério é montado por linhas de campo, operador e valor, com o texto gravado visível; "Testar critério" usa três valores digitados. Salvar passa por uma **prévia** que só conta quantos clientes ativos com consentimento ficariam em cada segmento, pelo perfil salvo. O `padrao` não é desativado nem ganha critério; prioridade e nome não se repetem. Quem já foi segmentado muda no próximo processamento (Q-33, Q-42) |
| Catálogo | produtos por categoria. Quem edita cria produtos e muda nome, categoria (gravada em minúsculas, com as existentes sugeridas) e preço (de 0 a 1.000.000); ativa e desativa com confirmação. Nada é excluído (Q-43) |
| Operação | saúde do banco, expiração de recomendações vencidas e retenção após a revogação (RN-18) |
| Usuários | operadores e papéis: criar com senha provisória, mudar papel, desativar, ativar e redefinir senha. Ninguém é excluído; o administrador não muda a própria conta (Q-35) |
| Auditoria | quem fez o quê: entradas, acessos negados, leituras de ficha e toda ação que muda dados, sem nome nem e-mail; desde a v1.12, também os pedidos do titular (correção, dados baixados, eliminação), a retenção, a importação confirmada e os arquivos de público gerados e baixados (Q-68); filtros por operador, ação e período (Q-39) |
| Minha senha | troca da própria senha (12 caracteres no mínimo); obrigatória depois de uma senha provisória (Q-40) |

**Papéis (RN-11, Q-37).** Cada operador tem um papel, conferido no banco a cada página: desativar alguém ou mudar o papel vale na hora (Q-38).

| Papel | Alcança |
|---|---|
| Atendimento | clientes, consentimento, eventos (inclusive a importação, v1.8) e processamento; na aba Direitos, corrigir o cadastro e baixar os dados (v1.9); Segmentos e Catálogo só para leitura |
| Analista | Segmentos e Catálogo, com edição (v1.6), os Relatórios agregados (v1.7, RN-15) e os Públicos para campanha (v1.11, Q-66); fora o arquivo de público (e-mail em hash), nenhum dado de cliente |
| Administrador | tudo isso, mais Relatórios, a eliminação de dados do titular (v1.9), Usuários, Auditoria e Operação (com a retenção) |

Cada alteração de segmento ou produto e cada prévia ficam na Auditoria, com os nomes dos campos alterados, nunca os valores (Q-44). Os pedidos do titular, a importação e os públicos também (v1.12, Q-68): a trilha guarda só login, ação, alvo (cliente ou segmento; nenhum na retenção e na importação) e resultado, e as quantidades continuam só no log. Baixar os dados do titular ou o arquivo de um público sem conseguir registrar na trilha dá erro, e nenhum arquivo sai.

Operadores de exemplo, **só no `make serve`** (modo memória; senhas de desenvolvimento, nunca na implantação, Q-41): `admin` / `dev-senha-admin` (administrador), `atendimento` / `dev-senha-atendimento` e `analista` / `dev-senha-analista`.

Demonstração (Q-30, Q-32): 7 segmentos (vip, premium, fiel, engajado, explorador, ocasional, padrao), catálogo de 7 categorias e 15 clientes com história ao longo de meses: Ana (premium, com recomendações expiradas), Bruno (engajado), Carla e Davi (ocasional; Davi sem produto elegível), Elisa (nunca concedeu), Fábio (revogou), Gabi (inativa, histórico só leitura), Heitor (vip), Isabela (fiel), João (explorador), Karina (concedeu v1 pela API, revogou e concedeu v2 pelo painel: histórico completo), Lucas (ainda não processado), Marina (só recomendação expirada), Nina (padrão) e Olga (revogou há mais de 180 dias: a retenção apaga o comportamento dela). Para os relatórios terem grupos acima de 5, a carga soma 60 clientes de volume (`Cliente Volume 101` a `160`, `c-101@exemplo.test` em diante) com história gerada de forma determinística: compras em várias categorias, 6 sem consentimento e 6 que revogaram (Q-46). Desde a v1.11, Ana, Bruno, Heitor e Isabela também concederam **Publicidade com terceiros**, e Carla concedeu e revogou (a segmentação dela continua valendo). No `make serve` todos entram sozinhos; na implantação, com `make exemplos`. Desde a v1.8, cada um dos 15 clientes com história tem ID fixo derivado do e-mail sintético, para o arquivo de exemplo da importação funcionar em qualquer instalação.

Sessão assinada sem estado (funciona com as duas réplicas) e conferida no banco a cada página, CSRF em todo envio, 5 tentativas de entrada por IP em 5 minutos, CSP `'self'`. Funciona sem JavaScript (formulários com redirecionamento e páginas de confirmação); com JavaScript, troca só a área do painel e usa diálogos nativos.

## Portal do titular (Spec 3.7.2)

Em `/titular`, no mesmo `omega-server`, para o próprio cliente, sem depender da operação (Q-55). Mesmo design system, com a marca em destaque e linguagem sem jargão.

| Tela | O que faz |
|---|---|
| Entrar | o titular informa o e-mail; a resposta é **sempre a mesma**, com ou sem cadastro, para não revelar quem é cliente |
| Link de acesso | uso único, 30 minutos, só o hash guardado; abrir o link mostra "Entrar no portal" e só o envio consome (Q-56) |
| Meus dados | cadastro, estado e histórico da autorização; retirar (efeito imediato, movimento com origem `portal`) ou autorizar de novo com o termo vigente; baixar os próprios dados (JSON); pedir a eliminação |

Sessão própria (cookie `perfila_titular`, caminho `/titular`, 30 minutos), sem acesso a nada do painel. Limites: 3 links por cadastro por hora (no banco) e 10 pedidos por IP em 15 minutos. O envio de e-mail fica atrás do port `Notificador`; no MVP, a **caixa de saída simulada** guarda a mensagem com o destinatário mascarado por 24 horas e aparece só no painel, em **Operação › Caixa de saída (simulada)** (Q-57). O pedido de eliminação aparece em **Clientes › Pedidos do titular** e é atendido na aba Direitos do titular da ficha, que o marca como atendido (Q-58).

Experimente com `make serve`: abra `http://localhost:8080/titular`, peça o link com `c-006@exemplo.test` (Fábio, que revogou pelo portal e tem um pedido aberto) e copie o link da caixa de saída do painel.

## LGPD

Nenhum caso de uso que processa dado pessoal roda sem consentimento válido para a sua finalidade (C-03, RN-02): `segmentacao` para perfil, segmento e recomendação, e também `publicidade_terceiros` para entrar num público de campanha. A revogação tem efeito imediato (RN-03) e, nos públicos, deixa o cliente pendente de remoção nos arquivos já gerados (RN-21, art. 18 §6º). O arquivo de público leva só o e-mail em hash, que continua sendo dado pessoal, e fica disponível por 24 horas. Não há integração real com plataformas de anúncio: o arquivo é subido pela equipe. O titular pode pedir a correção, uma cópia em JSON ou a eliminação dos dados; quem revogou há mais de 180 dias tem o comportamento apagado pela retenção, e a prova do consentimento fica sempre guardada (Spec v1.9). Cada pedido atendido gera o log `direito_atendido`, só com IDs. Os logs da aplicação e do Nginx não trazem nome, e-mail nem IP (C-09), e os testes usam só dados sintéticos. Os relatórios (RN-15) mostram só números agregados, sem identificar ninguém, e escondem grupos com menos de 5 clientes. O checklist completo, com a evidência de cada item e o que ficou para depois do MVP, está em [`docs/CHECKLIST_LGPD.md`](docs/CHECKLIST_LGPD.md). Os requisitos são interpretação técnica da equipe, não parecer jurídico.

## Mapa do repositório

```text
CLAUDE.md                   constituição para o agente (C-01..C-12) e mapa de contexto
OPEN_QUESTIONS.md           decisões Q-01..Q-46, com o porquê
specs/                      spec.md (v1.7) · plan.md (roadmap) · tasks.md · bdd-rules.md
docs/uml/                   UML1..UML8 em texto (Mermaid) + imagens originais e corrigidas (img/v1)
docs/design/                design system Perfila Console (princípios, telas, textos, tokens, componentes)
docs/                       checklist LGPD · roteiro de demonstração · relatório parcial · auditoria
docs/pdf/                   os documentos em PDF para entrega (Spec, Plano, Tarefas e BDD v1.2, decisões, checklist, roteiro)
features/                   cenários Gherkin + passos godog
cmd/omega-server            binário do UML8 (serve, expirar-recomendacoes, aplicar-retencao, hash-senha, carregar-exemplos) — raiz de composição
cmd/perfila-demo            demonstração do core em memória
internal/domain/            entity · port · service
internal/application/       usecase
internal/api/               rotas, handlers, middlewares (autenticação, LGPD, registro)
internal/web/               painel do operador: sessão, CSRF, telas, templates e estáticos embutidos
internal/infra/             memoria · relogio · external · config · repository/pg (+ migrações) · repository/contrato
internal/shared/errs        erros de domínio com código estável
deploy/                     Dockerfile · docker-compose.yml · nginx/ · config.env.example
scripts/                    preparar-implantacao.sh · demo-api.sh
```
