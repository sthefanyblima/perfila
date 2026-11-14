# Checklist LGPD e auditoria de logs (T-25)

Perfila · Equipe Omega · 14/11/2026 · Spec v1.9 (com o painel web, seção 4b, o histórico do consentimento, Q-34, e os públicos para campanha com consentimento por finalidade, seção 4c, Q-48 a Q-54)

Interpretação técnica da equipe sobre a Lei 13.709/2018, **não é parecer jurídico** (CLAUDE.md, seção 4). Cada item aponta a evidência que o prova: um cenário BDD, um teste automático ou um arquivo de configuração. ✅ atendido no MVP · ⚠️ atendido com limitação · ⬜ fora do MVP (requisito futuro).

## 1. Consentimento (UC-08, UML5)

| # | Requisito | Como o Perfila atende | Evidência | |
|---|---|---|---|---|
| L-01 | Nenhum dado pessoal é processado sem consentimento válido | `Precondicoes.Verificar` em UC-01 a UC-04 e, em reforço, middleware LGPD na API (C-03, RN-01) | cenários `@lgpd` em `evento`, `perfilamento`, `segmentacao`, `recomendacao` e `api.feature`; `TestFluxoPontaAPonta` (403 antes do consentimento) | ✅ |
| L-02 | Consentimento é livre e explícito: ausência de registro não vale como consentimento | Sem registro equivale a NAO_CONCEDIDO (D-02); cadastro (UC-09) não concede nada | `consentimento.feature` ("sem registro"), `cliente.feature` | ✅ |
| L-03 | Revogação a qualquer momento, com efeito imediato | `Revogar` muda o estado; UC-01 a UC-04 passam a falhar na hora (RN-03) | `consentimento.feature`, `api.feature`; `TestFluxoPontaAPonta` (403 após revogar) | ✅ |
| L-04 | Finalidade e versão do termo registradas, e a finalidade filtra o uso | Desde a v1.9 (Q-48), um consentimento por (cliente, finalidade), com `Finalidade` enumerada (`segmentacao`, `publicidade_terceiros`) e `VersaoTermo` obrigatória. UC-01 a UC-04 exigem `segmentacao` (RN-01); o público de campanha exige as duas (RN-17). Registros anteriores passaram a valer só para `segmentacao` (migração 009) | `publicos.feature` ("Só a publicidade não libera o uso dos dados de comportamento", "Cada finalidade tem o seu próprio estado", finalidade fora do enumerado); `consentimento.feature`; `TestMigracaoFinalidades` | ✅ |
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
| L-12 | Acesso à API autenticado | Bearer `AUTH_TOKEN` comparado em tempo constante; 401 sem token (D-16) | `api.feature` (401) | ⚠️ token estático único; JWT e perfis de acesso fora do MVP |
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
| L-23 | Acesso ao painel autenticado | Operador por configuração com senha em bcrypt, sessão assinada (HMAC-SHA256) com 8 h, 5 tentativas por IP em 5 minutos (Q-22); senha e chave geradas fora do git (Q-28) | `painel.feature` (sem sessão, senha errada, tentativas, sessão vencida); `TestSessao`, `TestLimitador` | ⚠️ operador único e limite por réplica; perfis de acesso fora do MVP |
| L-24 | Logs do painel sem dado pessoal | Registra método, caminho (só UUIDs), status e, nas ações, operador, ação e `clienteId`; nunca nome, e-mail, senha nem o termo de busca (Q-24). O IP do `X-Real-IP` só alimenta o limite de tentativas, em memória | cenário "Os logs do painel não contêm nome, e-mail nem o termo de busca" | ✅ |
| L-25 | Nenhuma requisição a terceiros com o IP de quem opera | Fontes IBM Plex, CSS e htmx servidos pelo próprio binário; CSP `default-src 'self'` | cenário `@web` e `@implantacao` dos cabeçalhos de segurança | ✅ |
| L-26 | Proteção contra envio forjado e enquadramento | CSRF em todo POST, cookies `HttpOnly`, `Secure`, `SameSite=Strict`; `frame-ancestors 'none'` e `X-Frame-Options: DENY` | cenário "Envio sem o token CSRF é recusado" | ✅ |
| L-27 | Atos com efeito jurídico deliberados | Conceder e revogar passam por confirmação que diz o efeito (diálogo ou página própria); em revogar, o foco inicial fica em Cancelar | cenário "Revogar exige confirmação" | ✅ |

## 4c. Públicos para campanha (Spec v1.9, RN-17, RN-18, Q-48 a Q-54)

| # | Requisito | Como o Perfila atende | Evidência | |
|---|---|---|---|---|
| L-33 | Compartilhar com terceiros exige consentimento específico (art. 7º §5º, art. 8º §4º) | Finalidade própria `publicidade_terceiros`, concedida e revogada à parte, com confirmação que diz o efeito; só entra no público quem tem ela **e** a `segmentacao`, está ativo e está no segmento | `publicos.feature` ("Só entra no público quem concedeu as duas finalidades", inativo e revogado ficam de fora); `publicos_painel.feature` (conceder pelo painel) | ✅ |
| L-34 | Mínimo de dados no que sai do sistema (art. 6º, III) | O arquivo leva só o SHA-256 do e-mail normalizado e o nome do segmento; nunca nome, e-mail em claro, ID, eventos, valores nem perfil. O hash pseudonimiza, não anonimiza: o arquivo continua sendo dado pessoal e é tratado como tal | cenário "O arquivo leva só o e-mail em hash e o nome do segmento"; `TestHashEmail`; `TestDestinoCSV` | ✅ |
| L-35 | A revogação chega a quem recebeu os dados (art. 18 §6º) | Revogar `publicidade_terceiros` ou `segmentacao` deixa o cliente com remoção pendente em todos os públicos exportados; o arquivo de remoção do segmento tira ele. Quem deixa de se qualificar também fica pendente na exportação seguinte | cenários "Revogar qualquer finalidade tira o cliente do público exportado", "O arquivo de remoção leva quem revogou", "Quem mudou de segmento fica pendente"; `@web` "Revogar a publicidade pelo painel deixa a remoção pendente" | ⚠️ a propagação não fica na mesma transação da revogação; uma falha vai para o log (`remocao_propagada_falhou`) e é corrigida pela exportação seguinte do segmento. A remoção na plataforma depende de a equipe subir o arquivo |
| L-36 | Registro de cada compartilhamento, sem dado pessoal | `ExportacaoPublico` com data, operador, segmento, tipo, destino, quantidade, nome e SHA-256 do arquivo; log `publico_exportado` só com IDs e números | cenário "A exportação fica registrada sem dado pessoal"; `@web` "Os logs do painel ao exportar não contêm dado pessoal" | ✅ |
| L-37 | Retenção do arquivo gerado | O conteúdo fica 24 h no banco (`PUBLICOS_ARQUIVO_HORAS`) para ser baixado; depois não é entregue (410) e é apagado na próxima exportação ou download. Registro e membros ficam como prova | cenários "Depois do prazo o arquivo não é mais entregue" (`@aplicacao` e `@web`); contrato `Publico` (retenção) | ⚠️ sem rotina agendada de expurgo: o apagamento acontece na próxima exportação ou download; a eliminação dos membros segue a política de retenção ainda fora do MVP (D-13, L-20). Quem baixa o arquivo deve apagá-lo depois de subir na plataforma |
| L-38 | Relatório sem reidentificação | Agregado por segmento sem identificador, contando só quem tem `segmentacao`; grupo com menos de 5 aparece como "menos de 5", sem ticket nem categoria, e não há linha de total | cenários "O agregado esconde grupos com menos de 5 clientes", "O CSV do agregado não tem identificador"; `@web` do relatório | ✅ |
| L-39 | Transferência internacional (art. 33) e contrato com a plataforma (art. 39) | O sistema não envia nada a terceiros: o destino do MVP é um arquivo que a equipe sobe. Antes de uma integração por API (o port `DestinoPublicitarioGateway` já permite), é preciso o contrato com a plataforma (papéis de controlador e operador) e a base para a transferência internacional (cláusulas-padrão da ANPD) | Q-54 | ⬜ pendência documentada; não bloqueia o MVP de arquivo |
| L-40 | Quem pode exportar | Operador único do painel (Q-22), com o login no registro; quando os papéis da Q-35 entrarem, só `ANALISTA` e `ADMINISTRADOR` | Q-53 | ⚠️ sem perfis de acesso nesta entrega |

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
| `publico_exportado` (v1.9) | `ExportarPublicoUseCase` | operador, tipo, `segmentoId`, `exportacaoId`, quantidade, SHA-256 do arquivo | Não. Nem e-mail nem hash de e-mail |
| `remocao_propagada`, `remocao_propagada_falhou` (v1.9) | `PropagarRevogacaoUseCase` | `clienteId`, quantidade ou mensagem do erro | Não |
| acesso do Nginx | `deploy/nginx/nginx.conf` (`log_format omega`) | horário, método, caminho sem query string, status, tempo, réplica, id da requisição | Não. **Sem IP do cliente** (IP é dado pessoal, Q-14) |
| PostgreSQL | configuração padrão da imagem | erros, sem texto de parâmetros | Não |

**Prova automática:** `TestFluxoPontaAPonta` (memória e PostgreSQL) cadastra um cliente com nome e e-mail sintéticos, envia o e-mail também numa query string e falha se qualquer um deles aparecer no log.

**Observação:** o `clienteId` (UUID) aparece nos logs. É um identificador pseudônimo, permitido por C-09 ("somente IDs"); sozinho, ele não identifica a pessoa sem acesso ao banco.

## 6. Requisitos futuros (pós-MVP)

1. Rotas de acesso, correção, portabilidade e eliminação (L-19), e uma política de retenção com expurgo após revogação (L-20, D-13).
2. O filtro por finalidade (L-04) entrou na v1.9. O histórico de versões do termo (L-06) entrou na v1.4; falta expô-lo ao titular (portal, fase 2), com as duas finalidades.
6. Públicos (L-35, L-37, L-39): expurgo agendado do arquivo e dos membros removidos, contrato com a plataforma e base de transferência internacional antes de qualquer integração por API.
3. Autenticação por usuário (JWT/OIDC) com perfis Cliente, Administrador e Analista (L-12).
4. Certificado emitido por uma autoridade, cifra do volume do banco e TLS interno (L-10, L-16).
5. Prazo de retenção dos logs definido junto com a política de dados.
