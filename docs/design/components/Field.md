# Field

Rótulo, controle, dica e erro de um campo de formulário, sempre nessa ordem.

**O que fornecer.** `<div class="pf-field">` com `<label class="pf-field__label" for>`, o controle (`<input class="pf-input">`, `<select class="pf-select">` ou `<textarea class="pf-textarea">`) com `id`, a dica opcional em `pf-field__hint` e, no erro, `pf-field__error` com `id` ligado ao controle por `aria-describedby`, mais `aria-invalid="true"` no controle (borda `critical`). Use `pf-input--mono` para IDs. Busca: `pf-search` com a lupa dentro do campo.

**Forma.** Altura `size-control`, borda de 1,5px em `line`, raio `radius-field`. Em foco, a borda vira `brand` e surge um halo de 4px em `brand-tint`. Campo somente leitura usa o fundo da página.

**Regras.** Rótulo sempre visível (placeholder não substitui rótulo). A dica diz o formato esperado ("Pelo menos 8 caracteres."). A mensagem de erro diz como corrigir; os textos estão na seção Textos. Valide no navegador (`required`, `min`, `max`, `step="0.01"`) e de novo no servidor.

**Não faça.** Rótulo flutuante, campo sem borda, erro só em vermelho sem texto.
