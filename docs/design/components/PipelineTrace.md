# PipelineTrace

As quatro etapas do processamento (UML4, RN-09), na ordem, numa faixa violeta com uma régua de marcações embaixo.

**Etapas fixas.** 01 Coletar eventos · 02 Calcular perfil · 03 Atribuir segmento · 04 Gerar recomendação. Estados: `--ok` (concluída: número em círculo cheio claro), `--run` (em andamento: arco girando), `--warn` (concluída com atenção, ex.: sem produto elegível), `--fail` (falhou), `--blocked` (bloqueada por falta de consentimento) e sem modificador (pendente).

**A régua.** `pf-rule` com 101 marcações (`pf-ticks`; a cada 25 uma maior, a cada 5 uma média) e a agulha "agora" (`pf-needle`). Cada etapa ocupa um quarto: ao processar, as marcações acendem uma a uma em `stage-lit` e a agulha percorre a etapa em 750ms, em linha reta. Falha na etapa 3 para a agulha no meio dela. Sem consentimento, a agulha some e nada acende. As marcações perto do ponteiro crescem (efeito de lupa; desligado com movimento reduzido).

**O que fornecer.** O resultado de `ProcessarClienteUseCase` por etapa em `pf-pipe__res`: um resumo curto ("9 eventos lidos", "5 navegação · 1 compra · 3 interação", "Segmento Padrão", "Coleção de ficção, edição especial") ou a mensagem da parada. O estado também vai em texto para o leitor de tela.

**Regras.** Sempre as quatro etapas, mesmo pendentes. Falha em uma etapa: as seguintes ficam pendentes e um `Alert` crítico acima explica o que ficou salvo, com o código da requisição em mono. Abaixo de 900px vira 2×2. Logo abaixo da faixa vem o `Verdict`.
