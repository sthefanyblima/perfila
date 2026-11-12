# TelaFicha

A ficha do cliente com consentimento Concedido, na aba Visão geral: selo de consentimento, processamento etapa por etapa, perfil, segmento e recomendação.

Composição: `PageHeader` + `ConsentRecord` compacto + `Tabs` + `Panel` com `PipelineTrace` + `Panel`s com `KeyValue`, `ScoreMeter` e `CriterionChips`. Sem consentimento, tudo abaixo das abas vira um único `Alert` crítico (seção Telas e fluxos).
