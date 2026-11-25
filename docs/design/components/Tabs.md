# Tabs

Divide a ficha do cliente em Visão geral, Eventos, Consentimento e Direitos do titular.

**O que fornecer.** `pf-tabs` com um link `pf-tab` por aba, cada um para a sua rota (`/painel/clientes/{id}/eventos`), com `aria-current="page"` na atual; funciona sem JavaScript e pode ser favoritada. A aba ativa é uma pílula preenchida em `ink` com texto da cor do fundo; as demais são texto `ink-muted`, e em hover ganham fundo branco. A contagem opcional vai em `pf-tab__count`.

**Regras.** No máximo 5 abas; o rótulo é substantivo. Sem consentimento, a aba Eventos continua visível e explica o bloqueio, em vez de sumir; Consentimento e Direitos do titular ficam sempre disponíveis. Abaixo de 700px as abas rolam na horizontal, em uma linha só.
