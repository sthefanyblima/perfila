# PipelineTrace

As quatro etapas do processamento (UML4, RN-09), na ordem, com o resultado ou o motivo da parada em cada uma.

**Etapas fixas.** 1 Coletar eventos · 2 Gerar perfil · 3 Segmentar · 4 Recomendar. Estados: `--ok` (concluída), `--warn` (concluída com atenção, ex.: sem produto elegível), `--fail` (falhou), `--pending` (não executada porque uma anterior falhou, ou ainda não processado).

**O que fornecer.** O resultado de `ProcessarClienteUseCase` por etapa: um resumo curto ("4 eventos", "Ticket médio R$ 610,00", "premium", "Livro A, válida até 12/12") ou a mensagem da parada. O estado também vai em texto para o leitor de tela.

**Regras.** Sempre as quatro etapas, mesmo pendentes. Falha em uma etapa: as seguintes ficam pendentes e um `Alert` crítico acima explica o que ficou salvo. Abaixo de 720px, vira 2×2.
