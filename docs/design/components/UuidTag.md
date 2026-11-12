# UuidTag

ID de cliente, evento ou recomendação, truncado em mono e com botão de copiar.

**Formato.** 8 primeiros caracteres + "…" + 4 últimos (`6200ae2f…a4c9`); o ID completo vai no `title` e no `aria-label` do botão. Ao copiar, toast "ID copiado".

**O que fornecer.** `<span class="pf-uuid">` com o texto truncado e `<button class="pf-uuid__copy" data-copiar="<id completo>">`. Na busca de clientes, aceite o ID completo colado.

**Não faça.** Mostrar o UUID inteiro em tabela. Esconder o ID: ele é o elo com os logs (C-09 só registra IDs).
