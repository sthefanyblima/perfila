# TelaFicha

A ficha do cliente com consentimento Concedido, na aba Visão geral: nome grande, carimbo de consentimento, faixa de processamento, laudo, perfil, segmento e recomendação.

Composição, de cima para baixo: cabeçalho da ficha (chips + nome + carimbo, ver `PageHeader`) → `Tabs` → `PipelineTrace` (faixa violeta com a régua) → `Verdict` (frase do laudo à esquerda, "Processar agora" à direita) → três `Panel` em `pf-g3`: **Perfil** (`KeyValue` + `ScoreMeter`), **Segmento** (`SegmentSeal` + descrição + `CriterionChips`) e **Recomendação** (`RecommendationTicket`). Sem consentimento, tudo abaixo das abas vira um único `Alert` crítico (seção Telas e fluxos). Se a segmentação falha, um `Alert` crítico antecede a faixa. Protótipo: `prototipo-manha.html`, rota `#/clientes/c-001`.
