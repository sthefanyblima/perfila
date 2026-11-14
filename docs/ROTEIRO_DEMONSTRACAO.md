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
| 2 | 2 min | `make verify` | gofmt, vet, testes, **154 cenários BDD verdes** e teste de arquitetura | O BDD é o critério de aceite executável. O teste de arquitetura falha se alguém quebrar a Clean Architecture (C-01, C-07) |
| 3 | 1 min | `make demo` | Fluxo da sequência em memória, passo a passo | UC-01 → UC-02 → UC-03 → UC-04, como no UML4; sem consentimento, nada roda |
| 4 | 1 min | `docker compose -f deploy/docker-compose.yml ps` | `nginx`, `app1`, `app2`, `db`; **só o Nginx publica porta** (8443) | É o UML8: balanceador com TLS, duas réplicas do `omega-server`, PostgreSQL `omega_db` numa rede interna |
| 5 | 3 min | `make demo-implantacao` | Cadastro → **403 sem consentimento** → concessão → eventos → processamento (segmento `premium`, recomendação `ATIVA`) → revogação → **403** | Tudo passa pelo Nginx em HTTPS e grava no PostgreSQL. A revogação tem efeito imediato (RN-03) |
| 5b | 3 min | Abrir `https://localhost:8443/painel` (usuário `operador`; senha: `grep PAINEL_SENHA deploy/.env`) | Fluxo A: Cadastrar cliente → selo **Não concedido** e aviso de bloqueio → Conceder (diálogo) → Eventos: 3 compras em livros → Processar agora → 4 etapas concluídas, perfil, segmento com critérios e recomendação. Fluxo B: Revogar (foco em Cancelar) → selo **Revogado** e painéis bloqueados | A mesma regra da API, agora para quem opera: o consentimento aparece antes de tudo e a interface nem oferece o que o domínio recusaria. Tudo pelo Nginx, nas duas réplicas (sessão assinada) |
| 5c | 2 min | Abrir `https://localhost:8443/titular`, pedir o link com o e-mail de um cliente de exemplo (`c-006@exemplo.test`); no painel, **Operação › Caixa de saída (simulada)**, copiar o link | "Se o e-mail estiver cadastrado, enviamos um link" (igual para qualquer e-mail) → "Entrar no portal" → Meus dados com a autorização e o histórico → Retirar ou autorizar de novo → no painel, a ficha mostra o movimento "pelo portal do titular"; **Pedidos do titular** lista o pedido de eliminação | O próprio titular exerce os direitos, sem telefonar. O link vale uma vez e por 30 minutos, e a resposta nunca revela quem é cliente. A caixa de saída é simulada: nenhum e-mail sai do sistema (Q-56, Q-57) |
| 6 | 1 min | `for i in 1 2 3 4; do curl -s --cacert bin/omega.crt -o /dev/null -D - https://localhost:8443/saude \| grep -i x-omega-replica; done` | Endereços das duas réplicas se alternando | O balanceador distribui a carga; o cabeçalho mostra quem atendeu (Q-11) |
| 7 | 2 min | `make bdd-implantacao` | 5 cenários (inclui o painel e o portal do titular com os cabeçalhos de segurança): tráfego pelo balanceador, **queda de uma réplica sem interrupção**, banco inacessível | O cenário 2 para a `app2` de verdade e confere 20 respostas 200; depois a religa |
| 8 | 30 s | `make expirar` | `recomendacoes_expiradas` com a quantidade | Subcomando do administrador do UML8 (RN-08) |
| 9 | 30 s | `docker compose -f deploy/docker-compose.yml logs --tail 5 nginx app1` | Logs só com IDs, sem nome, e-mail nem IP | C-09 e `docs/CHECKLIST_LGPD.md`: cada requisito LGPD com sua evidência e o que ficou para depois |

Encerramento (se houver tempo): `specs/tasks.md` mostra todas as tarefas T-01 a T-30 concluídas; `OPEN_QUESTIONS.md` mostra as decisões tomadas e por quê.

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
| `make serve` (painel em memória) | `MSYS_NO_PATHCONV=1 docker run --rm -p 8080:8080 -v "$(pwd -W)":/src -w /src golang:1.22 make serve` e abra `http://localhost:8080/painel` (operador / dev-senha) |
| `make expirar` | `MSYS_NO_PATHCONV=1 docker compose -f deploy/docker-compose.yml exec app1 omega-server -config /etc/omega/config.env expirar-recomendacoes` |
| `make exemplos` | `MSYS_NO_PATHCONV=1 docker compose -f deploy/docker-compose.yml exec app1 omega-server -config /etc/omega/config.env carregar-exemplos` |
| `make derrubar` | `docker compose -f deploy/docker-compose.yml down` |
