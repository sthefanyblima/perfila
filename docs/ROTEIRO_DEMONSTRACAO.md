# Roteiro de demonstração — evento final (24/11)

Perfila · Equipe Omega · ensaio em 23/11 · duração alvo: **15 minutos**

Cada passo traz o comando, o que aparece na tela e o que falar. Os comandos usam `make`; no Windows sem `make` nem Go, use os equivalentes da seção 4.

## 1. Preparação (antes de entrar na sala)

| Item | Como conferir |
|---|---|
| Docker Desktop no ar, com Compose v2 | `docker compose version` |
| Porta 8443 livre | `docker ps --format '{{.Ports}}'` não mostra 8443 de outro projeto |
| Imagens já construídas (evita esperar o download ao vivo) | `make implantar` uma vez, na véspera |
| Implantação no ar e saudável | `docker compose -f deploy/docker-compose.yml ps` com `app1` e `app2` em `healthy` |
| Terminal com fonte grande e `jq` instalado (opcional, deixa o JSON legível) | — |

Para recomeçar do zero (banco vazio): `docker compose -f deploy/docker-compose.yml down -v && make implantar`.

## 2. Roteiro

| # | Tempo | Comando | O que aparece | O que falar |
|---|---|---|---|---|
| 1 | 1 min | — (mostrar `docs/uml/img/v1/UML4_sequencia.png` e `UML8_implantacao.png`) | Diagramas corrigidos | Problema: perfilar e segmentar clientes **só com consentimento LGPD**. Os 8 diagramas viraram texto (`docs/uml/`) e Spec (`specs/spec.md`), que guiou o agente de IA (SDD) |
| 2 | 2 min | `make verify` | gofmt, vet, testes, **513 cenários BDD verdes** e teste de arquitetura | O BDD é o critério de aceite executável. O teste de arquitetura falha se alguém quebrar a Clean Architecture (C-01, C-07) |
| 3 | 1 min | `make demo` | Fluxo da sequência em memória, passo a passo | UC-01 → UC-02 → UC-03 → UC-04, como no UML4; sem consentimento, nada roda |
| 4 | 1 min | `docker compose -f deploy/docker-compose.yml ps` | `nginx`, `app1`, `app2`, `db`; **só o Nginx publica porta** (8443) | É o UML8: balanceador com TLS, duas réplicas do `omega-server`, PostgreSQL `omega_db` numa rede interna |
| 5 | 3 min | `make demo-implantacao` | Cadastro → **403 sem consentimento** → concessão → eventos → processamento (segmento `premium`, recomendação `ATIVA`) → revogação → **403** | Tudo passa pelo Nginx em HTTPS e grava no PostgreSQL. A revogação tem efeito imediato (RN-03) |
| 5b | 3 min | Abrir `https://localhost:8443/painel` (usuário `operador`; senha: `grep PAINEL_SENHA deploy/.env`) | Fluxo A: Cadastrar cliente → selo **Não concedido** e aviso de bloqueio → Conceder (diálogo) → Eventos: 3 compras em livros → Processar agora → 4 etapas concluídas, perfil, segmento com critérios e recomendação. Fluxo B: Revogar (foco em Cancelar) → selo **Revogado** e painéis bloqueados | A mesma regra da API, agora para quem opera: o consentimento aparece antes de tudo e a interface nem oferece o que o domínio recusaria. Tudo pelo Nginx, nas duas réplicas (sessão assinada) |
| 5c | 2 min | No `make serve` (operadores de exemplo, Q-41): entrar como `analista` / `dev-senha-analista` → o menu só tem Segmentos e Catálogo; abrir `/painel/clientes` → **Sem permissão**. Entrar como `admin` / `dev-senha-admin` → Usuários: desativar `atendimento` (com o painel do atendimento aberto noutra janela, a próxima página leva a Entrar com "Sua sessão foi encerrada") → Auditoria filtrada por "Abrir ficha" | Perfis de acesso (Q-35 a Q-41): cada papel só alcança o que precisa (minimização), a mudança vale na hora e cada ação fica na trilha, só com IDs | Atendimento, analista e administrador são os atores do UML1; a auditoria guarda 6 meses e é expurgada por subcomando (L-12, L-23, L-28) |
| 5d | 2 min | No `make serve`, como `analista` / `dev-senha-analista`: Segmentos → **Editar** `engajado` → trocar 0,7 por 0,5 → **Ver prévia** (só contagens: engajado ganha, explorador perde) → **Salvar alterações**. Tentar desativar o `padrao` pela URL `/painel/segmentos/{id}/desativar` → recusado. Catálogo → **Novo produto** "Atlas" em "Livros" → gravado como `livros`. Entrar como `admin` → Auditoria filtrada por "Editar segmento" mostra o campo `criterio` | Configuração sem migração (Q-33): a regra continua no domínio, a prévia não expõe nenhum cliente e a mudança vale no próximo processamento | `ValidarSegmentos` impede configuração que quebre o processamento (Q-02, Q-17); a auditoria guarda só os nomes dos campos (Q-44) |
| 5e | 1 min | No painel, como `analista` ou `admin`, **Relatórios** (grupo Análise); trocar o período para 12 meses; **Exportar CSV** | Indicadores, consentimento por estado, clientes por segmento, categorias (com "Outras categorias"), cadastros, revogações e eventos por mês; grupos pequenos como "menos de 5" e barras tracejadas | UC-05 e UC-07 em versão parcial: só números agregados, sem nenhum nome nem ID. Quem revogou sai das métricas de comportamento, e um grupo de 1 a 4 clientes nunca aparece como número (RN-15, Q-46). O CSV leva exatamente o que a tela mostra |
| 5f | 1 min | No painel, como `admin` ou `atendimento`, **Importar eventos** → enviar `docs/exemplos/eventos-lote.csv` → **Conferir arquivo** → **Importar 4 eventos** | Prévia: 4 prontas, 1 duplicada, 5 com erro (Elisa e Fábio sem consentimento, Gabi inativa, compra com valor zero, data sem fuso); relatório com "Importação concluída". Enviar de novo: 0 importados, 5 duplicados | Lojas mandam lotes, mas **cada linha passa pelo consentimento** do cliente (RN-16): o lote não para e não deixa nada passar. O arquivo não fica guardado e o reenvio é seguro |
| 5g | 2 min | Abrir `https://localhost:8443/titular`, pedir o link com o e-mail de um cliente de exemplo (`c-006@exemplo.test`); no painel, **Operação › Caixa de saída (simulada)**, copiar o link | "Se o e-mail estiver cadastrado, enviamos um link" (igual para qualquer e-mail) → "Entrar no portal" → Meus dados com a autorização e o histórico → Retirar ou autorizar de novo → no painel, a ficha mostra o movimento "pelo portal do titular"; **Pedidos do titular** lista o pedido de eliminação | O próprio titular exerce os direitos, sem telefonar. O link vale uma vez e por 30 minutos, e a resposta nunca revela quem é cliente. A caixa de saída é simulada: nenhum e-mail sai do sistema (Q-56, Q-57) |
| 5h | 2 min | No painel, como `admin`: ficha de **Nina** → Consentimento → "Conceder publicidade" (confirmação com o efeito) → **Públicos** → segmento `padrao` → "Gerar arquivo" → confirmar → "Baixar arquivo" (abrir o CSV) → voltar à ficha → "Revogar publicidade" → Públicos mostra **1 remoção pendente** → "Gerar arquivo de remoção". Depois, **Relatório agregado** | Duas finalidades, dois consentimentos: a segmentação é uso interno; mandar o público para anúncio exige outra autorização (art. 7º §5º). O CSV só tem o e-mail em SHA-256 e o nome do segmento, e o histórico guarda quem exportou e o hash do arquivo, sem dado pessoal. Revogar tira o cliente dos públicos já entregues (art. 18 §6º). O relatório agregado esconde grupos com menos de 5 |
| 6 | 1 min | `for i in 1 2 3 4; do curl -s --cacert bin/omega.crt -o /dev/null -D - https://localhost:8443/saude \| grep -i x-omega-replica; done` | Endereços das duas réplicas se alternando | O balanceador distribui a carga; o cabeçalho mostra quem atendeu (Q-11) |
| 7 | 2 min | `make bdd-implantacao` | 5 cenários (inclui o painel e o portal do titular com os cabeçalhos de segurança): tráfego pelo balanceador, **queda de uma réplica sem interrupção**, banco inacessível | O cenário 2 para a `app2` de verdade e confere 20 respostas 200; depois a religa |
| 8 | 30 s | `make expirar` | `recomendacoes_expiradas` com a quantidade | Subcomando do administrador do UML8 (RN-08) |
| 8b | 2 min | Painel: ficha do Fábio (revogado) → aba **Direitos do titular** → Baixar dados (JSON) → Eliminar dados → digitar "Fábio Sintético" → Eliminar. Depois, Operação → **Aplicar retenção** (ou `make retencao`) → **Auditoria**, filtro "Eliminar dados do titular" | O arquivo JSON com cadastro, histórico e eventos; a ficha com o selo **Eliminado** e só a linha do tempo do consentimento; a busca por "Fábio" vazia; "Retenção aplicada a 1 cliente" (Olga, revogada há mais de 180 dias); na Auditoria, "Baixar dados do titular", "Eliminar dados do titular" e "Aplicar retenção" com o login, sem nome nem e-mail | Direitos do art. 18 da LGPD: acesso, portabilidade, correção e eliminação. A prova do consentimento fica, sem nome nem e-mail (art. 16, I). A retenção fecha a D-13 com prazo configurável. Desde a v1.12, cada pedido fica na trilha de auditoria (art. 37, Q-68) |
| 9 | 30 s | `docker compose -f deploy/docker-compose.yml logs --tail 5 nginx app1` | Logs só com IDs, sem nome, e-mail nem IP | C-09 e `docs/CHECKLIST_LGPD.md`: cada requisito LGPD com sua evidência e o que ficou para depois |

Encerramento (se houver tempo): `specs/tasks.md` mostra todas as tarefas T-01 a T-54 concluídas; `OPEN_QUESTIONS.md` mostra as decisões tomadas e por quê.

## 3. Plano B

| Falha | Alternativa |
|---|---|
| Docker não sobe ou porta ocupada | `PERFILA_PORTA_HTTPS=9443 make implantar` e trocar 8443 por 9443 nos passos 5 e 6 |
| Sem Docker | `make serve` num terminal e `make demo-api` noutro: a mesma API, em memória (D-23) |
| Sem rede para baixar imagens | Construir na véspera (seção 1); as imagens ficam em cache |
| `make bdd-implantacao` falha | Mostrar a gravação do ensaio (23/11) e seguir com o passo 8 |

## 4. Windows sem `make` nem Go

No Git Bash, a partir da raiz do repositório. O `MSYS_NO_PATHCONV=1` impede que o Git Bash converta os caminhos `/src` e `/etc/...`.

| Alvo | Equivalente |
|---|---|
| `make verify` | `MSYS_NO_PATHCONV=1 docker run --rm -v "$(pwd -W)":/src -w /src golang:1.22 sh -c 'git config --global --add safe.directory /src; make verify'` |
| `make demo` | mesmo comando, trocando `make verify` por `make demo` |
| `make implantar` | `sh scripts/preparar-implantacao.sh && docker compose -f deploy/docker-compose.yml up -d --build --wait` |
| `make demo-implantacao` | `mkdir -p bin && docker compose -f deploy/docker-compose.yml cp nginx:/etc/nginx/certs/omega.crt bin/omega.crt && PERFILA_CA=bin/omega.crt bash scripts/demo-api.sh https://localhost:8443 "$(sed -n 's/^AUTH_TOKEN=//p' deploy/config.env)"` |
| `make bdd-implantacao` | compile o binário de testes para Windows no contêiner: `MSYS_NO_PATHCONV=1 docker run --rm -v "$(pwd -W)":/src -w /src -e GOOS=windows -e CGO_ENABLED=0 golang:1.22 go test -c -o bin/features.test.exe ./features` e rode `cd features && BDD_TAGS=@implantacao ../bin/features.test.exe -test.count=1` |
| `make serve` (painel em memória) | `MSYS_NO_PATHCONV=1 docker run --rm -p 8080:8080 -v "$(pwd -W)":/src -w /src golang:1.22 make serve` e abra `http://localhost:8080/painel` (admin / dev-senha-admin; também atendimento / dev-senha-atendimento e analista / dev-senha-analista) |
| `make expirar` | `MSYS_NO_PATHCONV=1 docker compose -f deploy/docker-compose.yml exec app1 omega-server -config /etc/omega/config.env expirar-recomendacoes` |
| `make retencao` | `MSYS_NO_PATHCONV=1 docker compose -f deploy/docker-compose.yml exec app1 omega-server -config /etc/omega/config.env aplicar-retencao` |
| `make exemplos` | `MSYS_NO_PATHCONV=1 docker compose -f deploy/docker-compose.yml exec app1 omega-server -config /etc/omega/config.env carregar-exemplos` |
| `make derrubar` | `docker compose -f deploy/docker-compose.yml down` |
