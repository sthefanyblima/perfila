# ScoreMeter

Mostra o score de engajamento (0 a 1) como barra com a explicação do cálculo.

**O que fornecer.** `pf-score` com o valor ("0,40"), o rótulo, a trilha (`pf-score__fill` com `width` igual ao score em %) e a nota "4 de 10 eventos em 30 dias", com a meta de `ENGAJAMENTO_META_EVENTOS`. Opcional: `pf-score__mark` na posição do limiar de um segmento ("engajado ≥ 0,70"), com a legenda na nota.

**Regras.** O número sempre aparece; a barra só reforça. `role="meter"` com `aria-valuenow`, `aria-valuemin="0"`, `aria-valuemax="1"`.
