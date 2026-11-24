# DataTable

Tabela para listas de registros: clientes, segmentos, catálogo, pedidos, públicos.

**O que fornecer.** Um painel (`pf-panel`) com a barra de filtros (`pf-filters`: busca `pf-search` + `SegmentedControl`) e `<div class="pf-tablewrap">` (rolagem horizontal própria) com `<table class="pf-table">`, `<th scope="col">` em todos os cabeçalhos em caixa alta e, se houver paginação, `pf-table__foot` com "1–20 de 134" e os botões Anterior/Próxima. Sem bordas entre colunas: só uma linha `line` acima de cada linha. Números e datas alinhados com números tabulares e sem quebra; valores monetários à direita (`pf-r`).

**Primeira coluna.** O nome em `pf-cell-main` (negrito, bloco próprio, quebra se for longo) e, abaixo, o `UuidTag`. É a única coluna que quebra palavras.

**Linha clicável.** `data-href` na `<tr>` leva à ficha e a linha inteira ganha `brand-tint` em hover (com cantos de `radius-field` nas pontas); a primeira célula tem também um link real (para teclado e leitor de tela).

**Ações.** Botões por linha vão em `pf-acell`, na última coluna.

**Regras.** 20 linhas por página; e-mail mascarado; ID em `UuidTag`. Vazio: troque a tabela por `EmptyState`. Carregando: esqueleto (`Skeleton`) com a forma da tabela.
