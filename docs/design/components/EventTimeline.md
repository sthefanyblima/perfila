# EventTimeline

Lista cronológica dos eventos de comportamento de um cliente, do mais recente para o mais antigo.

**O que fornecer.** `<ol class="pf-timeline">` com itens `pf-timeline__item`: data e hora (mono, fuso de São Paulo), marcador por tipo (`--compra` cheio, navegação vazado, `--interacao` quadrado: forma, não cor), tipo + categoria + origem e o valor à direita (só em compras).

**Regras.** Só aparece com consentimento Concedido. Mais de 50 eventos: paginação "Carregar mais 50" no fim.
