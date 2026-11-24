# EventTimeline

Lista cronológica dos eventos de comportamento de um cliente, do mais recente para o mais antigo.

**O que fornecer.** `<ul class="pf-timeline">` com um `pf-timeline__item` por evento: linha vertical contínua e marcador redondo de 14px centralizado nela e na primeira linha do texto; o tipo em negrito ("Compra", "Navegação", "Interação") seguido de categoria e origem; a data e hora em `pf-timeline__sub` (mono, fuso de São Paulo); e o valor à direita (`pf-timeline__value`), só em compras.

**Regras.** Só aparece com consentimento Concedido; sem ele, o texto "Os eventos não são carregados sem consentimento válido." Mais de 50 eventos: paginação "Carregar mais 50" no fim. O formulário de registro (Tipo, Valor, Categoria, Origem) fica no mesmo painel, acima da linha do tempo; sem consentimento, os campos ficam desabilitados com o motivo escrito.
