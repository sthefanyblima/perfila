# DataTable

Tabela densa para listas de registros: clientes, segmentos, catálogo.

**O que fornecer.** `<div class="pf-tablewrap">` (rolagem horizontal própria) com `<table class="pf-table">`, `<th scope="col">` em todos os cabeçalhos e, se houver paginação, `pf-table__foot` com "1–20 de 134" e os botões Anterior/Próxima. Números e datas alinhados com números tabulares; valores monetários à direita (`pf-r`).

**Linha clicável.** `data-href` na `<tr>` leva à ficha, mas a primeira célula tem também um link real (para teclado e leitor de tela).

**Regras.** Linha mínima de `size-row`; 20 linhas por página; e-mail mascarado; ID em `UuidTag`. Vazio: troque a tabela por `EmptyState`.
