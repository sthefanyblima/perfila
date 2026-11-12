# Field

Rótulo, controle, dica e erro de um campo de formulário, sempre nessa ordem.

**O que fornecer.** `<div class="pf-field">` com `<label class="pf-field__label" for>`, o controle (`<input class="pf-input">` ou `<select class="pf-select">`) com `id`, a dica opcional em `pf-field__hint` e, no erro, `pf-field__error` com `id` ligado ao controle por `aria-describedby`, mais `aria-invalid="true"` no controle. Use `pf-input--mono` para IDs.

**Regras.** Rótulo sempre visível (placeholder não substitui rótulo). A dica diz o formato esperado ("Em reais. Compras precisam de valor maior que zero."). A mensagem de erro diz como corrigir; os textos estão na seção Textos. Valide no navegador (`required`, `min`, `max`, `step="0.01"`) e de novo no servidor.

**Não faça.** Rótulo flutuante, campo sem borda, erro só em vermelho sem texto.
