# Tabs

Divide a ficha do cliente em Visão geral, Eventos e Consentimento.

**O que fornecer.** Cada aba é um link para a sua rota (`/painel/clientes/{id}/eventos`), com `aria-current="page"` na atual, então funciona sem JavaScript e pode ser favoritada. A contagem opcional vai em `pf-tab__count` ("Eventos 4").

**Regras.** No máximo 5 abas; o rótulo é substantivo. Sem consentimento, a aba Eventos continua visível e explica o bloqueio, em vez de sumir.
