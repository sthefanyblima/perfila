# StatusTag

Etiqueta de estado do domínio: forma, palavra e cor, nunca só cor.

**Mapeamento fixo.** Consentimento: Concedido (`pf-tag--ok`), Não concedido (`pf-tag--neutral`), Revogado (`pf-tag--critical`). Recomendação: Ativa (`ok`), Gerada (`warn`), Expirada (`neutral`). Cliente: Inativo, Eliminado (`neutral`). Perfil desatualizado (`warn`). Pedido do titular: Aberto (`warn`), Atendido (`ok`). Importação: Duplicada (`neutral`), Com erro (`critical`). O ícone é o da tabela de iconografia do README; o texto é a palavra legível do estado (seção Textos).

**Forma.** Pílula (`radius-pill`) com fundo `*-tint`, texto em negrito de 12px na cor do estado e ícone de 16px. Sem borda.

**O que fornecer.** `<span class="pf-tag pf-tag--…">` com o SVG do ícone e o texto. Em tabelas, a etiqueta fica sozinha na célula e não quebra. `pf-tag--brand` existe só para marcas informativas (a etiqueta "E" do formulário de critério), nunca para estado do domínio.

**Não faça.** Inventar estados que o domínio não tem. Usar `brand` para estado. Mostrar o estado só pela cor.
