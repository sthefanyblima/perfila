# PipelineTrace

As quatro etapas do processamento (UML4, RN-09), na ordem, numa faixa violeta com uma régua de marcações embaixo.

**Etapas fixas.** 01 Coletar eventos · 02 Calcular perfil · 03 Atribuir segmento · 04 Gerar recomendação. Estados: `--concluida` (número em círculo cheio claro), `--atencao` (concluída com atenção, ex.: sem produto elegível), `--falhou` e sem modificador (pendente), que são os que o painel usa; `--run` (em andamento: arco girando) e `--blocked` (bloqueada por falta de consentimento) existem só no design system, porque o painel entrega o resultado já pronto e esconde a faixa sem consentimento.

**A régua.** `pf-rule` com 101 marcações em duas camadas dentro de `pf-ticks`: a de baixo, todas apagadas (`pf-ticks__camada`), e a de cima (`pf-ticks__camada--acesa`), em que as marcações já acesas aparecem em `stage-lit` e as demais levam `is-off` (a altura é `is-major` a cada 25, `is-mid` a cada 5 e `is-minor` nas outras). A agulha "agora" (`pf-needle`) fica em `--1` a `--4` (25% a 100%) ou `--falha` (62%). Cada etapa ocupa um quarto; ao abrir a página, a camada de cima se revela da esquerda para a direita e a agulha vai da origem até a posição. Falha na etapa 3 para a agulha no meio dela. Sem processamento, a agulha não aparece e nada acende. As marcações perto do ponteiro crescem (efeito de lupa; desligado com movimento reduzido).

**O que fornecer.** O resultado de `ProcessarClienteUseCase` por etapa em `pf-pipe__res`: um resumo curto ("9 eventos lidos", "5 navegação · 1 compra · 3 interação", "Segmento Padrão", "Coleção de ficção, edição especial") ou a mensagem da parada. O estado também vai em texto para o leitor de tela.

**Regras.** Sempre as quatro etapas, mesmo pendentes. Falha em uma etapa: as seguintes ficam pendentes e um `Alert` crítico acima explica o que ficou salvo, com o código da requisição em mono. Abaixo de 900px vira 2×2. Logo abaixo da faixa vem o `Verdict`.

**Movimento ao processar.** Com a classe `pf-stage--rodando` na faixa (o painel a põe ao voltar do processamento, com o aviso "Processamento concluído"), as etapas aparecem uma a uma, a cada 750ms, e a agulha anda em linha reta até a última etapa feita. Sem a classe, a faixa aparece já pronta.
