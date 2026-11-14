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
Configuração (somente leitura no MVP)
  /painel/segmentos                    Segmentos e critérios (RN-06)
  /painel/catalogo                     Produtos por categoria (RN-07)
Operação
  /painel/operacao                     Saúde do sistema e expiração de recomendações (RN-08)
  /painel/operacao/caixa-saida         Caixa de saída (simulada), com o portal ligado (v1.10)
```

No grupo Clientes, desde a v1.10: `/painel/pedidos`, os pedidos de eliminação feitos pelo portal do titular.

O portal do titular (`/titular`, v1.10) tem mapa e moldura próprios (seção Portal do titular, abaixo). Relatórios de segmento (UC-05 a UC-07) continuam fora deste escopo.

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
- **Abas (`Tabs`):** Visão geral · Eventos · Consentimento · Direitos do titular (v1.9).
- **Visão geral, com consentimento Concedido:**
  - Painel **Processamento**: `PipelineTrace` com as 4 etapas, resultado de cada uma e o botão "Processar agora" (primário da página). Antes do primeiro processamento: "Este cliente ainda não foi processado."
  - Painel **Perfil** (`KeyValue`): frequência de compras, ticket médio (R$), categoria preferida, engajamento (`ScoreMeter` com "4 de 10 eventos em 30 dias").
  - Painel **Segmento**: nome, descrição e o critério em `CriterionChips`, com cada condição marcada como atendida ou não pelo perfil atual (é o que responde "por que este segmento?").
  - Painel **Recomendação**: produto, motivo legível ("Segmento premium, categoria livros"), `StatusTag` Ativa, validade, mais o histórico recolhido das expiradas.
- **Visão geral, sem consentimento:** os quatro painéis são trocados por um único `Alert` crítico: "Sem consentimento válido. Os dados de comportamento deste cliente não podem ser usados. [Conceder consentimento]". Nenhum dado de comportamento é buscado no servidor (C-03).
- **Eventos:** formulário em linha para registrar (Tipo, Valor em R$, Categoria, Origem, Data e hora, que vem preenchida com agora) + `EventTimeline` do mais recente para o mais antigo. Sem consentimento: o formulário fica desabilitado, com o motivo escrito ao lado.
- **Consentimento:** o `ConsentRecord` completo (estado, finalidade, versão do termo, datas de consentimento, de revogação e de atualização) com as ações. Fica disponível sempre, inclusive sem consentimento: é o UC-08.
- **Direitos do titular (v1.9, Q-48 a Q-53):** disponível com qualquer estado do consentimento. Três painéis, um por pedido do titular:
  - **Corrigir cadastro**: formulário com Nome e E-mail preenchidos e o botão "Salvar correção". Toast "Cadastro corrigido".
  - **Acesso e portabilidade**: o texto diz o que vai no arquivo; o botão "Baixar dados (JSON)" baixa `perfila-dados-{id}.json`.
  - **Eliminação**: o texto diz o efeito. Com consentimento Concedido, só o `Alert` de atenção "Revogue o consentimento antes de eliminar os dados." (a ação fica na aba Consentimento: uma ação de consentimento por tela). Senão, "Eliminar dados" (danger) abre a confirmação forte: o efeito, o nome do cliente a digitar, **Eliminar dados** (danger) e **Cancelar** (foco inicial). Nome diferente: erro ao lado do campo.
- **Cliente eliminado:** cabeçalho com "Titular eliminado" e o selo **Eliminado**; no lugar das abas de dados, o `Alert` "Dados eliminados em {data}. O cadastro foi anonimizado; resta o histórico do consentimento, guardado como prova." e a linha do tempo do consentimento. Nenhuma ação é oferecida.

### Segmentos

`DataTable` com Prioridade, Nome, Descrição e Critério (`CriterionChips`), ordenada por prioridade. Uma nota no topo: "Segmentos entram por migração; para alterar, fale com a equipe técnica." O segmento `padrao` aparece com a etiqueta "Critério vazio: casa sempre".

### Catálogo

Tabela agrupada por categoria, com Nome, Preço (R$) e Situação. A nota explica a regra RN-07: "A recomendação escolhe o primeiro produto ativo da categoria preferida, por ordem alfabética."

### Operação

- Painel **Saúde**: estado do banco (verde ou vermelho, vindo de `/saude`) e versão das migrações.
- Painel **Expirar recomendações vencidas**: o texto explica a regra ("Recomendações ativas com validade vencida passam a Expirada."); o botão "Expirar agora" abre um diálogo de confirmação; o resultado aparece em toast: "3 recomendações expiradas" ou "Nenhuma recomendação vencida".
- Painel **Retenção após a revogação** (v1.9, RN-18): o texto explica a regra com o prazo configurado; o botão "Aplicar retenção" (danger) abre a confirmação; o resultado aparece em toast: "Retenção aplicada a 2 clientes" ou "Nenhum cliente com a retenção vencida".

### Pedidos do titular (v1.10)

`DataTable` com **Cliente** (nome e `UuidTag`), **Pedido em**, **Situação** (`StatusTag`: Aberto em warn, Atendido em ok) e **Consentimento** (estado atual). Abertos primeiro. A linha leva à aba Direitos do titular da ficha, onde fica o aviso "O titular pediu a eliminação pelo portal em {data}." e a eliminação de sempre. Vazio: "Nenhum pedido do titular."

### Caixa de saída (simulada) (v1.10)

Nota no topo: "Simulação para a demonstração: nenhum e-mail sai do sistema. Mensagens com mais de 24 horas são apagadas." Lista das mensagens, da mais recente para a mais antiga: **Para** (mascarado), **Assunto**, **Enviada em** e o texto com o link. Vazio: "Nenhuma mensagem nas últimas 24 horas."

## Portal do titular (v1.10, Q-55 a Q-58)

Para a pessoa titular dos dados, não para a operação. Sem barra lateral: uma coluna única (`size-measure`), cabeçalho com a marca **Perfila** em destaque e, com sessão, "Sair" à direita. Linguagem do dia a dia: "autorização" no lugar de "consentimento" nas ações, sem siglas nem números de regra.

```text
/titular/entrar        Pedir o link de acesso (e-mail)
/titular/acesso?t=     Abrir o link: botão "Entrar no portal"
/titular/meus-dados    Cadastro, autorização, histórico e pedidos
/titular/revogar       Confirmar: retirar a autorização
/titular/conceder      Confirmar: autorizar de novo com o termo vigente
/titular/dados         Baixar meus dados (JSON)
/titular/eliminacao    Confirmar: pedir a eliminação
```

- **Entrar:** título "Acesse seus dados", o campo **E-mail** e o botão "Enviar link de acesso". Depois de enviar, sempre o mesmo `Alert` info: "Se o e-mail estiver cadastrado, enviamos um link de acesso. Ele vale por 30 minutos e só pode ser usado uma vez." Nada indica se o e-mail existe.
- **Abrir o link:** "Entrar no portal" (primário). Link inválido: `Alert` crítico "Este link não vale mais. Peça um novo." com "Pedir novo link".
- **Meus dados:** três painéis.
  - **Seus dados de cadastro**: nome, e-mail e data do cadastro.
  - **Sua autorização**: `StatusTag` do consentimento, finalidade em palavras ("segmentação: entender seu comportamento de compra para recomendar produtos"), termo e datas; uma só ação: "Retirar autorização" (danger) com Concedido, ou "Autorizar de novo" (primário) com Não concedido ou Revogado. Abaixo, a linha do tempo, igual à da ficha.
  - **Seus direitos**: "Baixar meus dados (JSON)" e "Pedir eliminação" (danger). Com Concedido, no lugar do botão, o `Alert` atenção "Para pedir a eliminação, retire a autorização antes."; com pedido aberto, o `Alert` info "Você pediu a eliminação em {data}. A equipe vai atender o pedido." e nenhuma ação de autorizar.
- **Confirmações:** página própria (ou `<dialog>` com JavaScript), com o efeito escrito; em retirar e em pedir eliminação, o foco inicial fica em Cancelar.

## Fluxos principais

### A. Primeiro atendimento (cadastro → consentimento → eventos → processamento)

1. Clientes → **Cadastrar cliente** → nome e e-mail → **Cadastrar cliente**.
2. A ficha abre com o selo **Não concedido** e o aviso de bloqueio. Ação: **Conceder consentimento**.
3. O diálogo pede **Finalidade** (padrão "segmentacao") e **Versão do termo** (padrão: a vigente, configurável), e diz o efeito: "A partir de agora, os eventos deste cliente podem ser usados para perfil, segmento e recomendação." Ação: **Conceder**.
4. O selo vira **Concedido**; o toast diz "Consentimento concedido"; os painéis aparecem vazios ("ainda não processado").
5. Aba **Eventos** → registrar 3 compras e 1 navegação; cada uma entra no topo da linha do tempo, com o toast "Evento registrado".
6. Visão geral → **Processar agora** → o `PipelineTrace` marca as 4 etapas como concluídas e os painéis mostram perfil, segmento e recomendação.

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

### F. Pedido de eliminação (v1.9)

1. Se o selo diz **Concedido**, aba **Consentimento** → **Revogar consentimento** (fluxo B).
2. Aba **Direitos do titular** → **Baixar dados (JSON)**, se o titular também pediu uma cópia antes de eliminar.
3. **Eliminar dados** → a confirmação explica o efeito e pede o nome do cliente → digitar o nome → **Eliminar dados**.
4. A ficha mostra o selo **Eliminado** e o aviso com a data; o toast diz "Dados eliminados". A busca pelo nome antigo não encontra mais o cliente.

### G. Titular retira a autorização sozinho (v1.10)

1. `/titular/entrar` → e-mail → "Enviar link de acesso" → a mensagem de sempre.
2. Na demonstração, a operação abre **Operação › Caixa de saída (simulada)** e copia o link; na vida real, o titular abre o e-mail.
3. O link mostra "Entrar no portal" → **Meus dados**.
4. **Retirar autorização** → a confirmação diz o efeito → **Retirar autorização**. O selo vira **Revogado**; a linha do tempo mostra "Revogação, por você, no portal". No painel, a ficha mostra o mesmo movimento "pelo portal do titular".
5. **Pedir eliminação** → confirmação → o pedido aparece em **Clientes › Pedidos do titular**; a operação segue o fluxo F.

## Estados obrigatórios em toda tela

| Estado | Tratamento |
| --- | --- |
| Carregando | Esqueleto com a forma do conteúdo final (linhas de tabela, blocos de painel), nunca spinner de página inteira |
| Vazio | `EmptyState`: o que vai aparecer aqui e a ação para criar o primeiro |
| Erro de rede ou 503 | `Alert` crítico no topo: "Não foi possível falar com o servidor. Tente de novo em instantes." + "Tentar de novo" |
| Sem permissão (sessão) | Volta para Entrar |
| Sem consentimento | `Alert` crítico no lugar dos dados de comportamento |
