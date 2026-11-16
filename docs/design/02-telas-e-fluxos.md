# Telas e fluxos

## Mapa do painel

Todas as telas ficam sob `/painel` no próprio `omega-server`, atrás do Nginx. A navegação lateral (`SideNav`) tem quatro grupos, e cada grupo só aparece para o papel que pode abri-lo (v1.5, RN-11): Atendimento vê Clientes e Configuração; Analista, Análise (v1.7) e Configuração; Administrador, os quatro.

```text
/painel/entrar                         Entrar (sem sessão)
Clientes
  /painel/clientes                     Lista e busca
  /painel/clientes/novo                Cadastrar cliente (UC-09)
  /painel/clientes/{id}                Ficha: visão geral
  /painel/clientes/{id}/eventos        Ficha: eventos (UC-01)
  /painel/clientes/{id}/consentimento  Ficha: consentimento (UC-08)
  /painel/eventos/importar             Importar eventos em lote (UC-01, RN-16, v1.8)
  /painel/eventos/modelo.csv           Modelo do arquivo
Análise (v1.7, Q-46; Administrador e Analista)
  /painel/relatorios                   Relatórios agregados (UC-05 e UC-07 parciais, RN-15)
  /painel/relatorios.csv               O mesmo relatório em CSV
Configuração (leitura para todos; edição para Administrador e Analista desde a v1.6)
  /painel/segmentos                    Segmentos e critérios (RN-06)
  /painel/segmentos/novo, /{id}        Formulário do segmento, testar critério (RN-14)
  /painel/segmentos/previa             Prévia (só contagens) e Salvar (Q-42)
  /painel/segmentos/{id}/desativar     Prévia e confirmação; também /ativar
  /painel/catalogo                     Produtos por categoria (RN-07)
  /painel/catalogo/novo, /{id}         Formulário do produto; desativar com confirmação (Q-43)
Administração (só Administrador, v1.5)
  /painel/operacao                     Saúde do sistema e expiração de recomendações (RN-08)
  /painel/usuarios                     Operadores: criar, papel, desativar, ativar, redefinir senha (Q-35)
  /painel/usuarios/{id}                Um operador: papel, situação e senha provisória
  /painel/auditoria                    Trilha do painel, com filtros por operador, ação e período (Q-39)
Rodapé (todos os papéis)
  /painel/minha-senha                  Trocar a própria senha; obrigatória depois de senha provisória (Q-40)
```

Fase 2, fora deste escopo: `/titular` (portal do titular para ver e revogar o próprio consentimento, que exige autenticação do titular) e o restante de UC-05 a UC-07 (campanhas, recortes por segmento, comparação entre períodos).

## Moldura de toda tela

- `SideNav` à esquerda (`size-sidebar`) em `surface-2`; nome "Perfila" no topo; login e papel do operador, "Minha senha" e "Sair" no rodapé.
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

### Importar eventos (v1.8, Q-47)

Três passos na mesma tela, sem guardar o arquivo no servidor:

1. **Enviar:** painel "Arquivo de eventos" com o campo de arquivo (`.csv`), a explicação curta do formato (colunas, data com fuso, valor com ponto, até 1.000 linhas e 256 KiB), o link "Baixar modelo" e o botão **Conferir arquivo** (primário). Nada é gravado neste passo.
2. **Prévia:** três números no topo (**Prontas para importar**, **Duplicadas**, **Com erro**) e a tabela das linhas não aceitas: **Linha**, **Cliente** (`UuidTag`, ou "—" se o ID não é legível), **Situação** (`StatusTag`: Duplicada em neutral, Com erro em critical) e **Motivo** (a mensagem de `03-textos.md`). Ação primária: **Importar {n} eventos**; ao lado, "Enviar outro arquivo". Sem nenhuma linha pronta, o botão não aparece e um `Alert` explica.
3. **Relatório:** os mesmos números, agora como **Importados**, **Duplicados ignorados** e **Com erro**, e a mesma tabela; `Alert` ok "Importação concluída". Reenviar o mesmo arquivo é seguro: o que já entrou aparece como duplicado.

Arquivo recusado inteiro (vazio, grande, cabeçalho errado, codificação) volta ao passo 1 com um `Alert` crítico que diz o motivo e como corrigir.

### Segmentos

`DataTable` com Prioridade, Nome, Descrição e Critério (`CriterionChips`), ordenada por prioridade, e um painel "Inativos" abaixo. O segmento `padrao` aparece com a etiqueta "Critério vazio: casa sempre". Quem só lê vê a nota "Somente leitura".

Desde a v1.6 (Q-33, Q-42), quem edita vê "Novo segmento" no cabeçalho e, em cada linha, Editar e Desativar (ou Ativar); o `padrao` não tem Desativar. O formulário tem três painéis: dados (nome, prioridade, descrição), critério (o texto gravado visível e uma linha por condição: campo, operador e valor, ligadas por "E"; "Adicionar condição") e "Testar critério" (compras, ticket médio e engajamento digitados). O botão principal é **Ver prévia**: a página da prévia mostra, por segmento, Hoje, Com a mudança e Diferença, com a nota de que só conta clientes ativos com consentimento pelo perfil salvo e de que quem já foi segmentado muda no próximo processamento. Só ela tem **Salvar alterações** (e "Voltar ao formulário", que mantém o que foi digitado). Desativar e ativar mostram a mesma prévia antes de confirmar. Erros aparecem ao lado do campo, com a condição errada citada.

### Catálogo

Tabela agrupada por categoria, com Nome, Preço (R$) e Situação. A nota explica a regra RN-07: "A recomendação escolhe o primeiro produto ativo da categoria preferida, por ordem alfabética."

Desde a v1.6 (Q-43), quem edita vê "Novo produto" e o nome de cada produto vira link para o formulário: nome, categoria (com as existentes sugeridas; gravada em minúsculas) e preço com vírgula. A seção Situação desativa com confirmação (diálogo ou página própria) ou ativa.

### Operação

- Painel **Saúde**: estado do banco (verde ou vermelho, vindo de `/saude`) e versão das migrações.
- Painel **Expirar recomendações vencidas**: o texto explica a regra ("Recomendações ativas com validade vencida passam a Expirada."); o botão "Expirar agora" abre um diálogo de confirmação; o resultado aparece em toast: "3 recomendações expiradas" ou "Nenhuma recomendação vencida".
- Painel **Retenção após a revogação** (v1.9, RN-18): o texto explica a regra com o prazo configurado; o botão "Aplicar retenção" (danger) abre a confirmação; o resultado aparece em toast: "Retenção aplicada a 2 clientes" ou "Nenhum cliente com a retenção vencida".

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

## Estados obrigatórios em toda tela

| Estado | Tratamento |
| --- | --- |
| Carregando | Esqueleto com a forma do conteúdo final (linhas de tabela, blocos de painel), nunca spinner de página inteira |
| Vazio | `EmptyState`: o que vai aparecer aqui e a ação para criar o primeiro |
| Erro de rede ou 503 | `Alert` crítico no topo: "Não foi possível falar com o servidor. Tente de novo em instantes." + "Tentar de novo" |
| Sem permissão (sessão) | Volta para Entrar |
| Sem consentimento | `Alert` crítico no lugar dos dados de comportamento |
