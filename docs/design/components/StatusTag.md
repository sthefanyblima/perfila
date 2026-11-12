# StatusTag

Etiqueta de estado do domínio: forma, palavra e cor, nunca só cor.

**Mapeamento fixo.** Consentimento: Concedido (`pf-tag--ok`), Não concedido (`pf-tag--neutral`, borda tracejada), Revogado (`pf-tag--critical`). Recomendação: Ativa (`ok`), Gerada (`warn`), Expirada (`neutral pf-tag--solid`). Cliente: Inativo (`neutral pf-tag--solid`). O ícone é o da tabela de iconografia do README; o texto é a palavra legível do estado (seção Textos).

**O que fornecer.** `<span class="pf-tag pf-tag--…">` com o SVG do ícone e o texto. Em tabelas, a etiqueta fica sozinha na célula.

**Não faça.** Inventar estados que o domínio não tem. Usar `brand` para estado. Pílula arredondada.
