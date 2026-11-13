# Checklist LGPD e auditoria de logs (T-25)

Perfila · Equipe Omega · 13/11/2026 · Spec v1.5 (com o painel web, seção 4b, o histórico do consentimento, Q-34, e os perfis de acesso e a auditoria do painel, Q-35 a Q-41)

Interpretação técnica da equipe sobre a Lei 13.709/2018, **não é parecer jurídico** (CLAUDE.md, seção 4). Cada item aponta a evidência que o prova: um cenário BDD, um teste automático ou um arquivo de configuração. ✅ atendido no MVP · ⚠️ atendido com limitação · ⬜ fora do MVP (requisito futuro).

## 1. Consentimento (UC-08, UML5)

| # | Requisito | Como o Perfila atende | Evidência | |
|---|---|---|---|---|
| L-01 | Nenhum dado pessoal é processado sem consentimento válido | `Precondicoes.Verificar` em UC-01 a UC-04 e, em reforço, middleware LGPD na API (C-03, RN-01) | cenários `@lgpd` em `evento`, `perfilamento`, `segmentacao`, `recomendacao` e `api.feature`; `TestFluxoPontaAPonta` (403 antes do consentimento) | ✅ |
| L-02 | Consentimento é livre e explícito: ausência de registro não vale como consentimento | Sem registro equivale a NAO_CONCEDIDO (D-02); cadastro (UC-09) não concede nada | `consentimento.feature` ("sem registro"), `cliente.feature` | ✅ |
| L-03 | Revogação a qualquer momento, com efeito imediato | `Revogar` muda o estado; UC-01 a UC-04 passam a falhar na hora (RN-03) | `consentimento.feature`, `api.feature`; `TestFluxoPontaAPonta` (403 após revogar) | ✅ |
| L-04 | Finalidade e versão do termo registradas | `Finalidade` e `VersaoTermo` obrigatórias em `Conceder` (guarda da 3.4.1) | `consentimento.feature` (guarda violada) | ⚠️ a finalidade é registrada, mas não filtra o uso dos dados (RN-02) |
| L-05 | Prova de quando o titular consentiu e revogou | `DataConsentimento`, `DataRevogacao` (mantida após nova concessão, Q-01) e `DataAtualizacao` em UTC | `consentimento.feature`; contrato `Consentimento` nos dois backends | ✅ |
| L-06 | Troca de termo deixa histórico explícito | Conceder sobre CONCEDIDO é `transicao_invalida`; é preciso revogar e conceder de novo (Q-06). Desde a v1.4 (Q-34, RN-10), cada concessão e revogação grava um `MovimentoConsentimento` com finalidade, versão do termo, data e origem (API, painel com o login do operador, portal), na mesma transação do registro; a tabela é só de inclusão (o banco recusa `UPDATE` e `DELETE`). A migração 006 reconstruiu, com origem `migracao`, só o que as datas antigas provavam | `consentimento.feature` e `painel.feature` (`@RN-10`), `api.feature` (409); contrato `MovimentoConsentimento` nos dois backends; `TestMigracaoReconstroiMovimentos`; `TestConsentimentoPelaAPIRegistraOrigem` | ✅ Clientes anteriores à 006: a versão do termo de concessões antigas, já sobrescritas, não pôde ser recuperada |

## 2. Minimização e qualidade dos dados

| # | Requisito | Como o Perfila atende | Evidência | |
|---|---|---|---|---|
| L-07 | Coletar só o necessário | Dado pessoal direto: só nome e e-mail do cliente. Eventos, perfil, segmento e recomendação usam apenas o `ClienteID` (UUID aleatório) | Spec 3.2; `001_inicial.sql` | ✅ |
| L-08 | Dados de teste e demonstração sintéticos | IDs `c-001`, e-mails `@exemplo.test`, catálogo inventado (B-08) | varredura do repositório sem e-mails reais (11/11); `seed.go`, `002_dados_demonstracao.sql` | ✅ |
| L-09 | Perfilamento explicável | Regras determinísticas e testáveis: RN-05 (perfil), RN-06 (critério declarativo), RN-07 (recomendação); o motivo vai na resposta (`segmento:<nome>;categoria:<categoria>`) | `perfilamento.feature`, `segmentacao.feature`, `recomendacao.feature` | ✅ |

## 3. Segurança (art. 46)

| # | Requisito | Como o Perfila atende | Evidência | |
|---|---|---|---|---|
| L-10 | Tráfego do titular cifrado | Nginx com TLS 1.2/1.3 e HSTS; a única porta publicada é a HTTPS (Q-12) | `deploy/nginx/nginx.conf`; cenário "todo o tráfego entra pelo balanceador" (confere TLS) | ⚠️ certificado autoassinado na demonstração; em produção, montar um certificado real |
| L-11 | Banco inacessível de fora | `db` só na rede interna `dados`, sem porta publicada | `deploy/docker-compose.yml`; cenário "o banco não é acessível pelo dispositivo do cliente" | ✅ |
| L-12 | Acesso à API autenticado | Bearer `AUTH_TOKEN` comparado em tempo constante; 401 sem token (D-16) | `api.feature` (401) | ⚠️ token estático único para os sistemas que chamam a API; JWT fora do MVP. Pessoas não usam a API: operam pelo painel, que tem usuários e papéis desde a v1.5 (L-23) |
| L-13 | Segredos fora do código | Senha do banco e `AUTH_TOKEN` gerados aleatoriamente em `deploy/.env` e `deploy/config.env`, ignorados pelo git (Q-14); a URL do banco nunca vai para log ou erro | `scripts/preparar-implantacao.sh`; `.gitignore`; `TestArgumentosEConfig` (senha ausente do erro) | ✅ |
| L-14 | Menor privilégio | `omega-server` roda como usuário sem privilégio (uid 10001), binário estático | `deploy/Dockerfile` | ✅ |
| L-15 | Integridade dos dados | Chaves estrangeiras, `UNIQUE(cliente_id)` e `CHECK` nos estados e faixas (Spec 3.8) | `001_inicial.sql`; suíte de contrato | ✅ |
| L-16 | Cifra em repouso e entre balanceador e réplicas | — | — | ⬜ volume do PostgreSQL e tráfego interno Nginx → réplicas sem cifra (rede Docker isolada) |

## 4. Direitos do titular (art. 18)

| # | Direito | Situação | |
|---|---|---|---|
| L-17 | Confirmação e consulta do consentimento | `GET /v1/clientes/{id}/consentimento` | ✅ |
| L-18 | Revogação | `POST /v1/clientes/{id}/consentimento/revogacao` | ✅ |
| L-19 | Acesso, correção, portabilidade e eliminação dos dados | `Cliente.AtualizarDados` existe no domínio, mas não há rota; não há exportação nem eliminação | ⬜ |
| L-20 | Retenção e eliminação após revogação | Dados já persistidos permanecem (RN-03); política fora do MVP (D-13) | ⬜ |

## 4b. Painel web do operador (Spec 3.7.1, v1.3)

| # | Requisito | Como o Perfila atende | Evidência | |
|---|---|---|---|---|
| L-21 | Sem consentimento, o painel não mostra nem lê dados de comportamento | A ficha mostra só cadastro e consentimento; `ConsultarResultadoUseCase` aplica RN-01 antes de ler eventos, perfil, segmento e recomendação (C-03, Q-21) | `painel.feature` (`@lgpd`: ficha e registro de evento sem consentimento), `leitura.feature` | ✅ |
| L-22 | Mostrar o mínimo de dado pessoal | Lista com e-mail mascarado (`c***@exemplo.test`) e ID truncado; e-mail completo só na ficha | `TestFormatos`; tela Clientes | ✅ |
| L-23 | Acesso ao painel autenticado e por usuário | Desde a v1.5, cada pessoa é um operador no banco, com login próprio, senha em bcrypt (12 caracteres no mínimo) e papel (Q-35, Q-40). A senha é conferida por `AutenticarOperador` com o mesmo erro para login inexistente, senha errada e operador inativo (Q-36). Sessão assinada (HMAC-SHA256) com 8 h, conferida no banco a cada página: desativar alguém ou redefinir a senha derruba a sessão na hora (Q-38). Senha provisória obriga a troca no primeiro acesso. 5 tentativas por IP em 5 minutos (Q-22); a senha do administrador inicial e a chave ficam fora do git (Q-28) | `painel.feature` (sem sessão, senha errada, tentativas, sessão vencida, operador desativado, senha provisória); `usuarios.feature` (credenciais, sessão recusada, troca de senha); `TestSessao`, `TestLimitador`, `TestAdministradorInicial` | ⚠️ limite de tentativas por réplica (até 10 com o balanceamento); sem segundo fator |
| L-24 | Logs do painel sem dado pessoal | Registra método, caminho (só UUIDs), status e, nas ações, operador, ação e `clienteId`; nunca nome, e-mail, senha nem o termo de busca (Q-24). O IP do `X-Real-IP` só alimenta o limite de tentativas, em memória | cenário "Os logs do painel não contêm nome, e-mail nem o termo de busca" | ✅ |
| L-25 | Nenhuma requisição a terceiros com o IP de quem opera | Fontes IBM Plex, CSS e htmx servidos pelo próprio binário; CSP `default-src 'self'` | cenário `@web` e `@implantacao` dos cabeçalhos de segurança | ✅ |
| L-26 | Proteção contra envio forjado e enquadramento | CSRF em todo POST, cookies `HttpOnly`, `Secure`, `SameSite=Strict`; `frame-ancestors 'none'` e `X-Frame-Options: DENY` | cenário "Envio sem o token CSRF é recusado" | ✅ |
| L-27 | Atos com efeito jurídico deliberados | Conceder e revogar passam por confirmação que diz o efeito (diálogo ou página própria); em revogar, o foco inicial fica em Cancelar | cenário "Revogar exige confirmação" | ✅ |
| L-28 | Registro das operações com dado pessoal (art. 37) | Trilha de auditoria do painel (RN-13, Q-39): entradas, saídas, sessões encerradas, acessos negados, cada aba de ficha aberta e toda ação que muda dados, com data, login do operador, ação, tipo e ID do alvo e resultado. Sem nome, e-mail, senha nem termo de busca. Só inclusão (o banco recusa `UPDATE`); a ficha não aparece se a leitura não puder ser registrada. Retenção de 180 dias, expurgo por `omega-server expurgar-auditoria`. Tela Auditoria só para o administrador | `painel.feature` (`@RN-13`: ações e leituras na auditoria, auditoria sem dado pessoal, filtro); `usuarios.feature` (registro, só administrador lê, expurgo com 179, 180 e 181 dias); `TestAuditoriaSoInclusao`, `TestExpurgarAuditoria`; contrato `Auditoria` nos dois backends | ✅ só o painel é auditado; a API (sistemas, token único) fica com o log de acesso e os movimentos do consentimento |
| L-29 | Acesso só ao necessário para a função (art. 6º, III e VII) | Matriz de permissões no domínio (RN-11, Q-37): Atendimento opera clientes, consentimento, eventos e processamento; Analista vê só Segmentos e Catálogo, sem nenhum dado de cliente; Administrador cuida também de Usuários, Auditoria e Operação. O painel confere a permissão em cada rota (403 e registro `acesso_negado`) e não mostra no menu o que o papel não pode; os casos de uso de gestão conferem de novo | `TestPapelPode`; `painel.feature` ("Cada papel só abre o que pode", "O menu mostra só as telas do papel", "Sem permissão, a ação é recusada"); `usuarios.feature` (só o administrador gerencia operadores) | ✅ relatórios para o Analista (UC-05 a UC-07) e eliminação de dados ficam reservados na matriz, sem tela |
| L-30 | Configuração de segmentos sem expor titulares (art. 6º, III) | A prévia de uma mudança de segmento (Q-42) começa pelos consentimentos concedidos, só lê o perfil salvo de clientes ativos com consentimento válido e devolve apenas contagens por segmento: nenhum nome, e-mail ou ID chega à tela do Analista. "Testar critério" usa só valores digitados. A auditoria das alterações guarda os nomes dos campos de uma lista fechada, nunca valores (Q-44) | cenários "A prévia conta só clientes ativos com consentimento válido, pelo perfil salvo" (`@aplicacao`) e "Salvar passa pela prévia, que mostra só contagens" (`@web`); contrato `AuditoriaCampos` e `CHECK` da coluna `campos` (migração 008) | ✅ |

## 5. Auditoria de logs (C-09)

Todos os eventos de log do sistema, conferidos no código em 11/11/2026 (`grep` por chamadas de `slog` em `cmd/` e `internal/`):

| Evento | Origem | Campos | Dado pessoal? |
|---|---|---|---|
| `http` | `api.Registro` | método, caminho (só UUIDs), status, duração | Não. A query string não é registrada |
| `erro_interno` | `api.escreverErro` (só erros 500) | caminho, mensagem do erro | Não. Os erros de infraestrutura citam só IDs; os do driver trazem a mensagem do PostgreSQL, sem os valores |
| `recomendacao_ativada` | UC-04 (notificação reduzida a log, D-08) | IDs da recomendação, do cliente e do produto | Não |
| `servidor_no_ar`, `encerrando`, `modo_memoria` | `cmd/omega-server` | porta, aviso fixo | Não |
| `postgres_conectado` | `cmd/omega-server` | versões de migração aplicadas | Não. A URL do banco não é registrada |
| `recomendacoes_expiradas` | subcomando `expirar-recomendacoes` | quantidade | Não |
| `painel_http` | `web.registro` | método, caminho (só UUIDs), status | Não. Sem query string (termo de busca) |
| `painel_acao` | handlers do painel | operador, ação, `clienteId` | Não. O operador é o login da equipe, não o titular |
| `painel_erro_interno`, `painel_template` | painel (só falhas 500) | caminho, mensagem do erro | Não |
| `painel_desligado` | `cmd/omega-server` | aviso fixo | Não |
| `administrador_inicial_criado`, `operadores_exemplo_carregados` | `cmd/omega-server` (v1.5) | aviso fixo, quantidade | Não. Nem login nem senha |
| `auditoria_expurgada` | subcomando `expurgar-auditoria` | quantidade, retenção | Não |
| `auditoria_falhou` | painel (falha ao gravar a auditoria) | ação, mensagem do erro | Não. Os erros citam só IDs |
| acesso do Nginx | `deploy/nginx/nginx.conf` (`log_format omega`) | horário, método, caminho sem query string, status, tempo, réplica, id da requisição | Não. **Sem IP do cliente** (IP é dado pessoal, Q-14) |
| PostgreSQL | configuração padrão da imagem | erros, sem texto de parâmetros | Não |

**Prova automática:** `TestFluxoPontaAPonta` (memória e PostgreSQL) cadastra um cliente com nome e e-mail sintéticos, envia o e-mail também numa query string e falha se qualquer um deles aparecer no log.

**Observação:** o `clienteId` (UUID) aparece nos logs. É um identificador pseudônimo, permitido por C-09 ("somente IDs"); sozinho, ele não identifica a pessoa sem acesso ao banco.

## 6. Requisitos futuros (pós-MVP)

1. Rotas de acesso, correção, portabilidade e eliminação (L-19), e uma política de retenção com expurgo após revogação (L-20, D-13).
2. Filtrar o uso dos dados pela finalidade consentida (L-04). O histórico de versões do termo (L-06) entrou na v1.4; falta expô-lo ao titular (portal, fase 2).
3. Autenticação por sistema na API (JWT/OIDC) no lugar do token único (L-12). No painel, usuários e papéis entraram na v1.5 (L-23, L-29); falta segundo fator e limite de tentativas compartilhado entre as réplicas.
4. Certificado emitido por uma autoridade, cifra do volume do banco e TLS interno (L-10, L-16).
5. Prazo de retenção dos logs definido junto com a política de dados. A auditoria do painel já tem prazo (180 dias, L-28).
6. Reprocessar todos os clientes depois de mudar um critério (Q-45): trata dado de todos os titulares de uma vez e precisa de confirmação, execução em segundo plano e auditoria próprias.
