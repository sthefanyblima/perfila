# ScoreMeter

Mostra o engajamento do cliente como uma fileira de 10 marcas.

**O que fornecer.** O rótulo "Engajamento" em negrito, `pf-pips` com 10 `<i>` que acendem (`is-on`, `brand`) conforme os eventos dos últimos 30 dias (máximo 10), e a nota em `pf-why`: "4 de 10 eventos em 30 dias", com a meta de `ENGAJAMENTO_META_EVENTOS`.

**Regras.** A frase com o número sempre aparece; as marcas só reforçam. Dentro do painel Perfil, abaixo do `KeyValue`. `role="meter"` com `aria-valuenow`, `aria-valuemin="0"` e `aria-valuemax="10"` no contêiner.
