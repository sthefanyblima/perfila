# CriterionChips

Traduz o critério de um segmento (gramática da RN-06) em condições legíveis e mostra quais o perfil atende.

**O que fornecer.** `pf-crit` com uma `pf-crit__cond` por condição, ligadas por `pf-crit__and` ("E"). Nome legível do campo (Ticket médio, Compras, Engajamento), operador matemático (≥, ≤, >, <, =) e o número com vírgula. Na ficha, acrescente `--met` (com visto) ou `--unmet` (tracejado) conforme o perfil atual; na tela de segmentos, sem modificador. Critério vazio: `pf-crit__always` "Critério vazio: casa sempre".

**Por quê.** Responde "por que este cliente está neste segmento?" sem obrigar ninguém a ler `ticketMedio >= 500 && frequenciaCompras >= 3`. O texto original fica no `title`.
