# SegmentedControl

Escolha de um valor entre poucos, para filtrar uma lista (Todos, Concedido, Não concedido, Revogado).

**O que fornecer.** `pf-segctl` com um `<button aria-pressed>` por opção. O ativo fica preenchido em `ink` com texto da cor do fundo. Quebra em duas linhas em telas estreitas, sem passar da borda.

**Variante em links (a que o painel usa).** Quando cada opção é uma página ou uma consulta (filtro de consentimento da lista de clientes, período dos relatórios), use `<nav class="pf-segctl">` com um `<a>` por opção e `aria-current="true"` na atual. O link mantém os outros parâmetros da consulta (por exemplo, o texto da busca).
