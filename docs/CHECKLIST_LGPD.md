# Checklist LGPD e auditoria de logs (T-25)

Perfila · Equipe Omega · 14/11/2026 · Spec v1.10 (com o painel web, seção 4b, o histórico do consentimento, Q-34, os perfis de acesso e a auditoria do painel, Q-35 a Q-41, os relatórios agregados, seção 4c, a importação de eventos, seção 4d, os direitos do titular e a retenção, Q-48 a Q-53, e o portal do titular, seção 4a)

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
| L-17 | Confirmação e consulta do consentimento | `GET /v1/clientes/{id}/consentimento`; desde a v1.10, o próprio titular vê o estado e o histórico em **Meus dados**, no portal (Q-55) | ✅ |
| L-18 | Revogação | `POST /v1/clientes/{id}/consentimento/revogacao`; desde a v1.10, o próprio titular retira a autorização no portal, com efeito imediato e movimento com origem `portal` (`portal.feature`, "Retirar a autorização pelo portal tem efeito imediato e entra no histórico") | ✅ |
| L-19 | Acesso, correção, portabilidade e eliminação dos dados | Desde a v1.9 (Q-48 a Q-50), na aba Direitos do titular do painel, com qualquer estado do consentimento: correção de nome e e-mail; exportação em JSON (`versaoFormato` "1") de cadastro, consentimento, histórico, eventos, perfil, segmento e recomendações; eliminação (RN-17) que apaga o comportamento, anonimiza o cadastro (`DataEliminacao`, migração 009) e mantém o consentimento e os movimentos como prova (art. 16, I). Com consentimento concedido, a eliminação pede a revogação antes; a confirmação é o nome digitado. Evidência: `direitos.feature` (`@Q-48`, `@Q-49`, `@RN-17`), contrato `DireitosDoTitular`, `TestClienteAnonimizar`, `TestExportacaoFormato`, `TestMigracaoEliminacaoExigeAnonimizado` | ⚠️ só pelo painel, por meio da pessoa operadora; rotas da API e o portal do titular ficam para a fase 2 (Q-53). Restrição da eliminação ao papel ADMINISTRADOR quando a v1.5 entrar na main |
| L-20 | Retenção e eliminação após revogação | RN-18 (Q-51): o comportamento de quem está REVOGADO há `RETENCAO_DIAS_REVOGADO` dias ou mais (padrão 180) é apagado pelo subcomando `aplicar-retencao` e pelo botão da Operação; cadastro, consentimento e histórico ficam; quem concedeu de novo não é atingido. Evidência: `direitos.feature` (`@RN-18`, limites de 179, 180 e 181 dias), contrato `RevogadosAte`, `TestCarregarExemplos` (Olga atingida, Fábio não) | ⚠️ a retenção não roda sozinha: o administrador a agenda (cron do servidor) ou aciona pelo painel |

### 4a. Portal do titular (Spec 3.7.2, v1.10)

| # | Requisito | Como o Perfila atende | Evidência | |
|---|---|---|---|---|
| L-40 | O titular exerce os direitos sem intermediário | Portal `/titular`: ver cadastro, estado e histórico da autorização; retirar e autorizar de novo com o termo vigente; baixar os próprios dados (o mesmo JSON da portabilidade); pedir a eliminação, que a operação atende no painel com a confirmação forte da RN-17 (Q-55, Q-58, RN-20) | `portal.feature` (`@web` e `@aplicacao`); pedido no painel em "Pedidos do titular" | ✅ |
| L-41 | Identificar o titular sem revelar quem é cliente | Link de uso único por e-mail, 30 minutos, só o SHA-256 guardado, consumido por POST (abrir o link não o gasta); a resposta ao pedido é sempre a mesma; 3 links por cadastro por hora (no banco) e 10 pedidos por IP em 15 minutos; endereço do link vindo da configuração, nunca do `Host` (Q-56, RN-19) | `portal.feature` (resposta igual, limites 59/60/61 minutos, uso único, validade 29/30/31 minutos, link adulterado); contrato `TokenAcesso` nos dois backends | ⚠️ o pedido com e-mail cadastrado grava o token e leva um pouco mais de tempo; o limite por IP é por réplica |
| L-42 | Sessão do titular separada da operação | Cookie `perfila_titular` com caminho, chave (`PORTAL_CHAVE_SESSAO`) e validade (30 minutos) próprios; nenhuma rota recebe ID de cliente; cliente eliminado perde a sessão na requisição seguinte (Q-55) | `portal.feature` ("A sessão do portal não abre o painel", "A sessão do painel não abre o portal", "Pelo link, o titular vê só os próprios dados", sessão de 30 minutos); `TestCarregarPortal` (chave diferente da do painel) | ✅ |
| L-43 | O link de acesso não vaza | Nunca vai para o log (o portal não registra a query string); a caixa de saída simulada guarda o destinatário mascarado, apaga mensagens de mais de 24 horas e só aparece no painel (Q-57) | "Os logs do portal não contêm nome, e-mail nem o link"; "A caixa de saída simulada mostra o destinatário mascarado"; `TestNotificadorSimulado` | ⚠️ não há envio real de e-mail no MVP; a caixa simulada é vista pela pessoa operadora única até os papéis da v1.5 a restringirem a ADMINISTRADOR |

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

## 4c. Relatórios agregados (Spec RN-15, v1.7, Q-46)

| # | Requisito | Como o Perfila atende | Evidência | |
|---|---|---|---|---|
| L-31 | Relatório não identifica ninguém | Só contagens, somas e médias: o `RelatorioRepository` nunca devolve linha de cliente, e a tela e o CSV não têm nome, e-mail nem ID | contrato `Relatorio` (memória e PostgreSQL); cenários "A tela Relatórios mostra… sem identificar ninguém" e "…exportado em CSV, sem identificadores" | ✅ |
| L-32 | Comportamento só de quem consente | Segmento, médias, recomendações, categorias e eventos contam só clientes com consentimento CONCEDIDO no momento do relatório; revogados saem, inclusive o que gravaram antes (RN-03) | cenários "Revogados não entram nas métricas de comportamento" e "Clientes sem consentimento contam no cadastro e não no comportamento" | ✅ |
| L-33 | Grupos pequenos não reidentificam | Grupo de 1 a 4 clientes aparece como "menos de 5"; médias de grupo pequeno ocultas; taxa de concessão oculta se depender de um grupo oculto; categorias com menos de 5 compradores somadas em "Outras", sem nome | `TestProteger*`; cenários `@dominio` e "Grupos pequenos ficam ocultos…"; `TestRelatorioDosExemplos` | ⚠️ supressão só primária: a diferença entre o total e as partes visíveis pode revelar o tamanho de um grupo oculto (nunca um atributo de comportamento, porque as médias também ficam ocultas) |
| L-34 | Exportação sem vazamento | O CSV traz os mesmos valores protegidos da tela; célula que começa com `=`, `+`, `-` ou `@` ganha apóstrofo; o log registra só operador, ação e período | `TestCelulaCSVSemFormula`; cenário do CSV | ✅ |
| L-35 | Quem pode ver | Só ANALISTA e ADMINISTRADOR veem os relatórios (permissão `relatorios`, RN-11); o ATENDIMENTO não tem o menu e recebe 403 na rota | Q-46 (8); `TestPapelPode` | ✅ |

## 4d. Importação de eventos em lote (Spec v1.8, RN-16, Q-47)

| # | Requisito | Como o Perfila atende | Evidência | |
|---|---|---|---|---|
| L-36 | Lote não contorna o consentimento | Cada linha passa por RN-01 (cliente existe, ativo, consentimento válido) antes de gravar; a recusa é só daquela linha. Na API, sem cliente na rota, o caso de uso confere cada linha (C-03) | `importacao.feature` (`@lgpd`: não concedido, revogado e revogado entre a prévia e a confirmação); `TestArquivoDeExemploImporta` | ✅ |
| L-37 | O arquivo enviado não fica guardado | A prévia não grava nada; o conteúdo volta ao navegador num campo oculto de página `no-store` e só os eventos aceitos são gravados | cenário "Enviar o arquivo pelo painel mostra a prévia sem gravar"; Spec RN-16 | ✅ |
| L-38 | Auditoria do lote sem dado pessoal | Log `painel_acao` com operador, `importar_eventos` e as quantidades; na API, `importacao_eventos` só com as quantidades. Nenhum ID de cliente nem conteúdo do arquivo | cenário "A auditoria do lote registra o operador e as quantidades, sem dado pessoal" | ✅ (tabela de auditoria quando o PR de usuários entrar) |
| L-39 | Reenvio não duplica dados | Duplicado (cliente, tipo, instante e valor) é ignorado e contado à parte | cenários "Reenviar o mesmo arquivo não grava nada de novo" e de duplicados | ✅ |

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
| `retencao_aplicada` (v1.9) | subcomando `aplicar-retencao` | quantidade, prazo em dias | Não |
| `direito_atendido` (v1.9) | painel e subcomando `aplicar-retencao` | operador, pedido (`correcao`, `exportacao`, `eliminacao`, `retencao`), `clienteId` ou quantidade | Não. Auditoria dos pedidos até a v1.5 (`RegistroAuditoria`) entrar na main (Q-53); cenário "Cada pedido atendido fica no log, sem nome nem e-mail" |
| `painel_http` | `web.registro` | método, caminho (só UUIDs), status | Não. Sem query string (termo de busca) |
| `painel_acao` | handlers do painel | operador, ação, `clienteId`; nos relatórios (`ver_relatorio`, `exportar_relatorio`), operador, ação e período, sem nenhum número; em `importar_eventos`, só as quantidades (sem `clienteId`) | Não. O operador é o login da equipe, não o titular |
| `importacao_eventos` | `api.importarLote` | gravado, aceitas, duplicadas, rejeitadas | Não |
| `painel_erro_interno`, `web_template` | painel e portal (só falhas 500) | caminho, mensagem do erro | Não |
| `portal_http` | `web.portal.registro` (v1.10) | método, caminho, status | Não. Sem query string: o token do link nunca é registrado |
| `portal_acao` | handlers do portal (v1.10) | ação e `clienteId` da sessão; `link_pedido` sem nenhum dos dois | Não. Nunca nome, e-mail, token nem link |
| `portal_erro_interno` | portal (só falhas 500) | caminho, mensagem do erro | Não |
| `portal_desligado` | `cmd/omega-server` | aviso fixo | Não |
| `painel_desligado` | `cmd/omega-server` | aviso fixo | Não |
| `administrador_inicial_criado`, `operadores_exemplo_carregados` | `cmd/omega-server` (v1.5) | aviso fixo, quantidade | Não. Nem login nem senha |
| `auditoria_expurgada` | subcomando `expurgar-auditoria` | quantidade, retenção | Não |
| `auditoria_falhou` | painel (falha ao gravar a auditoria) | ação, mensagem do erro | Não. Os erros citam só IDs |
| acesso do Nginx | `deploy/nginx/nginx.conf` (`log_format omega`) | horário, método, caminho sem query string, status, tempo, réplica, id da requisição | Não. **Sem IP do cliente** (IP é dado pessoal, Q-14) |
| PostgreSQL | configuração padrão da imagem | erros, sem texto de parâmetros | Não |

**Prova automática:** `TestFluxoPontaAPonta` (memória e PostgreSQL) cadastra um cliente com nome e e-mail sintéticos, envia o e-mail também numa query string e falha se qualquer um deles aparecer no log.

**Observação:** o `clienteId` (UUID) aparece nos logs. É um identificador pseudônimo, permitido por C-09 ("somente IDs"); sozinho, ele não identifica a pessoa sem acesso ao banco.

## 6. Requisitos futuros (pós-MVP)

1. Direitos do titular pela API (o portal da v1.10 já atende acesso, portabilidade, revogação e o pedido de eliminação; a correção pelo titular continua só no painel). Agendar a retenção (L-20) no servidor.
2. Filtrar o uso dos dados pela finalidade consentida (L-04). O histórico de versões do termo (L-06) entrou na v1.4 e, desde a v1.10, o titular o vê no portal.
3. Autenticação por sistema na API (JWT/OIDC) no lugar do token único (L-12). No painel, usuários e papéis entraram na v1.5 (L-23, L-29); falta segundo fator e limite de tentativas compartilhado entre as réplicas.
4. Certificado emitido por uma autoridade, cifra do volume do banco e TLS interno (L-10, L-16).
5. Prazo de retenção dos logs definido junto com a política de dados. A auditoria do painel já tem prazo (180 dias, L-28).
6. Reprocessar todos os clientes depois de mudar um critério (Q-45): trata dado de todos os titulares de uma vez e precisa de confirmação, execução em segundo plano e auditoria próprias.
7. Envio real do link de acesso (SMTP ou serviço de e-mail) como outra implementação do port `Notificador`. A caixa de saída simulada já ficou restrita a ADMINISTRADOR (permissão `operacao`, L-43).
