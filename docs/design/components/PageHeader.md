# PageHeader

Trilha, título da página e ações: o topo de toda tela, exceto a ficha do cliente (que tem o cabeçalho de nome grande).

**O que fornecer.** `pf-crumbs` (mono, links até a página atual, que não é link), `h1.pf-header__title` (estilo `display`, uma vez por tela) e, à direita, `pf-header__actions` com no máximo uma ação primária. Em telas estreitas as ações descem para baixo do título.

**Ficha do cliente.** O título é o nome do cliente em `pf-ficha__nome`: primeiro nome em `ink` e o resto em `brand`, em tamanho enorme. Acima dele, os chips de metadados (ID, e-mail, cadastro, situação, Publicidade) em `pf-ficha__meta`; à direita, o carimbo do consentimento. Nomes longos reduzem a fonte (`--lg`, `--xl`). Cliente eliminado: título "Titular eliminado" e carimbo Eliminado.

**Regras.** O título é o nome do objeto ("Marina Costa") ou da lista ("Clientes"), nunca "Detalhes" ou "Página do cliente".
