# Telas e fluxos

## Mapa do painel

Todas as telas ficam sob `/painel` no próprio `omega-server`, atrás do Nginx. A navegação lateral (`SideNav`) tem três grupos:

```text
/painel/entrar                         Entrar (sem sessão)
Clientes
  /painel/clientes                     Lista e busca
  /painel/clientes/novo                Cadastrar cliente (UC-09)
  /painel/clientes/{id}                Ficha: visão geral
  /painel/clientes/{id}/eventos        Ficha: eventos (UC-01)
  /painel/clientes/{id}/consentimento  Ficha: consentimento (UC-08)
Campanhas (v1.9, UC-06 parcial)
  /painel/publicos                     Públicos: segmento, contagem, gerar e baixar, histórico, remoções (RN-17)
  /painel/publicos/relatorio           Relatório agregado por segmento (RN-18)
Configuração (somente leitura no MVP)
  /painel/segmentos                    Segmentos e critérios (RN-06)
  /painel/catalogo                     Produtos por categoria (RN-07)
Operação
  /painel/operacao                     Saúde do sistema e expiração de recomendações (RN-08)
```

Fase 2, fora deste escopo: `/titular` (portal do titular para ver e revogar o próprio consentimento, que exige autenticação do titular) e relatórios de segmento (UC-05 a UC-07).

## Moldura de toda tela

- `SideNav` à esquerda (`size-sidebar`) em `surface-2`; nome "Perfila" no topo; operador logado e "Sair" no rodapé.
- Área de trabalho em `surface-0`, até `size-content`, com margem lateral `space-6`.
- `PageHeader`: trilha (Clientes › Nome), título `display`, ações da página à direita (no máximo uma primária).
- Toasts no canto inferior direito, acima de tudo.
- Abaixo de 960px, a barra lateral vira um botão "Menu" no cabeçalho.

## Telas

### Entrar

Formulário único centralizado em `surface-1`, com usuário, senha e o botão "Entrar". Erro genérico, para não revelar qual campo errou: "Usuário ou senha incorretos." Depois de 5 tentativas em 5 minutos: "Muitas tentativas. Aguarde 5 minutos." Ao entrar, segue para a página pedida ou para `/painel/clientes`.

### Clientes (lista)

- Ação primária: **Cadastrar cliente**.
- Barra de filtros: busca por nome ou ID (campo único, a partir de 2 caracteres) e filtro de consentimento (Todos, Concedido, Não concedido, Revogado).
- `DataTable` com as colunas **Cliente** (nome; `UuidTag` embaixo), **E-mail** (mascarado: `c***@exemplo.test`), **Consentimento** (`StatusTag`), **Segmento** (nome, ou "—"), **Cadastro** (data). Linha inteira clicável, levando à ficha; 20 por página; paginação no rodapé com "1–20 de 134".
- Vazio sem filtro: `EmptyState` "Nenhum cliente cadastrado", com a ação "Cadastrar cliente". Vazio com filtro: "Nenhum cliente encontrado para 'ana'", com "Limpar busca".

### Cadastrar cliente

Formulário de uma coluna (`size-measure`): **Nome completo** e **E-mail**. Rodapé com "Cadastrar cliente" (primário) e "Cancelar". Erro de servidor `dados_invalidos` fica ao lado do campo. Ao concluir, vai para a ficha do cliente, já com o selo "Não concedido" e a ação "Conceder consentimento" em destaque.

### Ficha do cliente

O centro do produto.

- **Cabeçalho:** nome (`display`), `UuidTag`, e-mail completo, data de cadastro, situação (Ativo/Inativo) e, à direita, o `ConsentRecord` compacto com o estado e a ação contextual (Conceder ou Revogar).
- **Abas (`Tabs`):** Visão geral · Eventos · Consentimento.
- **Visão geral, com consentimento Concedido:**
  - Painel **Processamento**: `PipelineTrace` com as 4 etapas, resultado de cada uma e o botão "Processar agora" (primário da página). Antes do primeiro processamento: "Este cliente ainda não foi processado."
  - Painel **Perfil** (`KeyValue`): frequência de compras, ticket médio (R$), categoria preferida, engajamento (`ScoreMeter` com "4 de 10 eventos em 30 dias").
  - Painel **Segmento**: nome, descrição e o critério em `CriterionChips`, com cada condição marcada como atendida ou não pelo perfil atual (é o que responde "por que este segmento?").
  - Painel **Recomendação**: produto, motivo legível ("Segmento premium, categoria livros"), `StatusTag` Ativa, validade, mais o histórico recolhido das expiradas.
- **Visão geral, sem consentimento:** os quatro painéis são trocados por um único `Alert` crítico: "Sem consentimento válido. Os dados de comportamento deste cliente não podem ser usados. [Conceder consentimento]". Nenhum dado de comportamento é buscado no servidor (C-03).
- **Eventos:** formulário em linha para registrar (Tipo, Valor em R$, Categoria, Origem, Data e hora, que vem preenchida com agora) + `EventTimeline` do mais recente para o mais antigo. Sem consentimento: o formulário fica desabilitado, com o motivo escrito ao lado.
- **Consentimento:** o `ConsentRecord` completo (estado, finalidade, versão do termo, datas de consentimento, de revogação e de atualização) com as ações. Fica disponível sempre, inclusive sem consentimento: é o UC-08.
- **Consentimento por finalidade (v1.9, Q-48):** a aba mostra dois `ConsentRecord`, um embaixo do outro: **Segmentação** ("perfil, segmento e recomendação, uso interno") e **Publicidade com terceiros** ("entrar nos públicos enviados a plataformas de anúncio"). Cada um tem o seu `StatusTag`, termo, datas, a ação (Conceder ou Revogar, que leva `?finalidade=`) e a sua linha do tempo. No cabeçalho da ficha, o `ConsentRecord` compacto continua sendo o da Segmentação; ao lado, uma etiqueta `neutral` pequena "Publicidade: Concedido/Não concedido/Revogado". Conceder a publicidade sem a segmentação é permitido, mas o registro mostra a nota "Sem a Segmentação, este cliente não entra em nenhum público."

### Públicos (v1.9, RN-17)

- `PageHeader` "Públicos", com a trilha Campanhas › Públicos e, à direita, o link "Relatório agregado".
- Nota no topo (`Alert` info): "Entra no público quem está no segmento, está ativo e concedeu Segmentação e Publicidade com terceiros. O arquivo leva só o e-mail em hash (SHA-256), que ainda é dado pessoal."
- Seletor de segmento (lista dos ativos, por prioridade) com "Ver público". Com um segmento escolhido:
  - Painel **Público** (`KeyValue`): "Entram" (número grande), "Ficam de fora" (com a explicação "sem a finalidade, inativos ou sem consentimento") e o botão primário **Gerar arquivo** (desabilitado com zero, com o motivo "Nenhum cliente entra neste público").
  - Painel **Remoções pendentes**: quantos clientes saíram (revogaram ou deixaram de se qualificar) e ainda estão no público da plataforma, com o botão **Gerar arquivo de remoção** (secundário; desabilitado com zero).
  - Painel **Histórico** (`DataTable`): Data, Tipo (Inclusão/Remoção), Operador, Clientes, Hash do arquivo (12 primeiros caracteres, completo no `title`) e Arquivo ("Baixar" enquanto disponível; "Expirado" depois do prazo).
- Logo depois de gerar, a tela volta com um `Alert` ok: "Arquivo gerado com {n} clientes. Disponível para baixar até {data}." e o botão "Baixar arquivo".
- Vazio sem segmentos com clientes: `EmptyState` "Nenhum cliente processado ainda".

### Relatório agregado (v1.9, RN-18)

`DataTable` por segmento: Segmento, Clientes, Ticket médio (R$), Categoria mais comum. Grupo com menos de 5: a linha mostra "menos de 5" em Clientes e "—" nos outros campos, com a nota "Grupos com menos de 5 clientes não são detalhados, para ninguém ser identificado." Ação "Baixar CSV". Nenhum nome, e-mail ou ID aparece.

### Segmentos

`DataTable` com Prioridade, Nome, Descrição e Critério (`CriterionChips`), ordenada por prioridade. Uma nota no topo: "Segmentos entram por migração; para alterar, fale com a equipe técnica." O segmento `padrao` aparece com a etiqueta "Critério vazio: casa sempre".

### Catálogo

Tabela agrupada por categoria, com Nome, Preço (R$) e Situação. A nota explica a regra RN-07: "A recomendação escolhe o primeiro produto ativo da categoria preferida, por ordem alfabética."

### Operação

- Painel **Saúde**: estado do banco (verde ou vermelho, vindo de `/saude`) e versão das migrações.
- Painel **Expirar recomendações vencidas**: o texto explica a regra ("Recomendações ativas com validade vencida passam a Expirada."); o botão "Expirar agora" abre um diálogo de confirmação; o resultado aparece em toast: "3 recomendações expiradas" ou "Nenhuma recomendação vencida".

## Fluxos principais

### A. Primeiro atendimento (cadastro → consentimento → eventos → processamento)

1. Clientes → **Cadastrar cliente** → nome e e-mail → **Cadastrar cliente**.
2. A ficha abre com o selo **Não concedido** e o aviso de bloqueio. Ação: **Conceder consentimento**.
3. O diálogo pede **Finalidade** (padrão "segmentacao") e **Versão do termo** (padrão: a vigente, configurável), e diz o efeito: "A partir de agora, os eventos deste cliente podem ser usados para perfil, segmento e recomendação." Ação: **Conceder**.
4. O selo vira **Concedido**; o toast diz "Consentimento concedido"; os painéis aparecem vazios ("ainda não processado").
5. Aba **Eventos** → registrar 3 compras e 1 navegação; cada uma entra no topo da linha do tempo, com o toast "Evento registrado".
6. Visão geral → **Processar agora** → o `PipelineTrace` marca as 4 etapas como concluídas e os painéis mostram perfil, segmento e recomendação.

### F. Público para campanha (v1.9)

Ficha → Consentimento → Conceder em "Publicidade com terceiros" → Públicos → escolher o segmento → Gerar arquivo → confirmar → Baixar arquivo. Se o titular revogar depois: Ficha → Revogar a publicidade → Públicos mostra "1 remoção pendente" → Gerar arquivo de remoção → Baixar.

### B. Revogação

1. Na ficha, **Revogar consentimento** (botão `danger`, secundário).
2. O `ConfirmDialog` diz o efeito ("O processamento dos dados deste cliente para imediatamente. Os dados já registrados continuam guardados. Você pode conceder de novo depois.") e oferece **Revogar consentimento** (danger) ou **Cancelar** (com o foco inicial em Cancelar).
3. O selo vira **Revogado**, com a data; os painéis de comportamento dão lugar ao `Alert` crítico; o toast diz "Consentimento revogado".

### C. Processamento que para no meio

1. **Processar agora** → por exemplo, um segmento mal configurado (erro 500 de configuração, Q-17).
2. `PipelineTrace`: Coletar ✓, Perfil ✓, Segmento ✕ ("Falha de configuração dos segmentos. A equipe técnica foi avisada pelo log."), Recomendação, que fica pendente.
3. Um `Alert` crítico acima diz: "O perfil foi atualizado, mas a segmentação falhou. Nada mais foi alterado. Código da requisição: 4f2a…". O ID vem do cabeçalho `X-Request-ID` do Nginx, para a equipe técnica achar no log.

### D. Sem produto elegível

O processamento conclui as 3 primeiras etapas; a etapa Recomendação fica em **atenção** (warn): "Nenhum produto ativo na categoria preferida (games)." Não é erro: o painel Recomendação mostra o estado vazio explicado.

### E. Sessão expirada

Qualquer ação com a sessão vencida volta para **Entrar**, com o aviso "Sua sessão expirou. Entre de novo para continuar."; depois de entrar, a pessoa volta para a página onde estava.

## Estados obrigatórios em toda tela

| Estado | Tratamento |
| --- | --- |
| Carregando | Esqueleto com a forma do conteúdo final (linhas de tabela, blocos de painel), nunca spinner de página inteira |
| Vazio | `EmptyState`: o que vai aparecer aqui e a ação para criar o primeiro |
| Erro de rede ou 503 | `Alert` crítico no topo: "Não foi possível falar com o servidor. Tente de novo em instantes." + "Tentar de novo" |
| Sem permissão (sessão) | Volta para Entrar |
| Sem consentimento | `Alert` crítico no lugar dos dados de comportamento |
