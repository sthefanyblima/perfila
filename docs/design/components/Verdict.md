# Verdict

O laudo: a frase que resume o que o sistema concluiu sobre o cliente, logo abaixo da faixa de processamento.

**O que fornecer.** `pf-laudo` com `pf-laudo__texto` (display, peso 600, até 36px) e, à direita, `pf-laudo__acoes` com a ação principal ("Processar agora", primária quando o perfil não existe, está desatualizado ou falhou). A frase fala do cliente pelo primeiro nome e destaca com `<mark>` o dado principal (segmento, quantidade de eventos) e com `<em>` o nome da recomendação.

**Estados.** Em andamento: "Processando os N eventos de Marina, na ordem do fluxo…". Nunca processado: "Este cliente ainda não foi processado. Há N eventos esperando para virar um perfil." Concluído: "Marina está no segmento Padrão: as compras somam R$ 89,90, abaixo do mínimo de R$ 300,00. Recomendamos Mais vendidos de ficção." Eventos novos: acrescenta "Chegaram eventos novos desde o último cálculo." Falha: "O perfil foi atualizado, mas a segmentação falhou…" com `mark[data-k="blocked"]`. O texto de cada caso fica na seção Textos.
