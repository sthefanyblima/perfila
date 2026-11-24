# UuidTag

ID de cliente, evento ou recomendação, em mono e com cópia ao clique.

**Formato.** Pílula `pf-uuid` em `neutral-tint` com o ID e o ícone de copiar; em hover vira `brand-tint` com texto `brand`. IDs longos (UUID): 8 primeiros caracteres + "…" + 4 últimos (`6200ae2f…a4c9`), com o ID completo no `title` e no `aria-label`. IDs curtos de demonstração (`c-001`) aparecem inteiros. Ao copiar, toast "ID copiado".

**O que fornecer.** `<button class="pf-uuid" data-copiar="<id completo>">` com o texto e o ícone. Em célula de tabela, vem abaixo do nome, com 4px de folga. Na busca de clientes, aceite o ID completo colado.

**Não faça.** Mostrar o UUID inteiro em tabela. Esconder o ID: ele é o elo com os logs (C-09 só registra IDs).
