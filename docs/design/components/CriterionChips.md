# CriterionChips

Traduz o critério de um segmento (gramática da RN-06) em condições legíveis e mostra quais o perfil atende.

**O que fornecer.** `pf-crit` com uma `pf-crit__cond` por condição, em pílula mono: nome legível do campo (ticket médio, compras, engajamento), operador (>=, <=, =) e o número, com vírgula e R$ onde for dinheiro. Na ficha, acrescente `--met` (fundo `ok-tint`, ícone de visto) ou `--unmet` (fundo `warn-tint`, ícone "x") conforme o perfil atual, e a legenda "Cada condição mostra se o perfil atual a atende."; na tela de segmentos, sem modificador. Critério vazio: `pf-crit__always` "Critério vazio: casa sempre". No formulário do segmento, as condições são linhas com campo, operador e valor, ligadas por uma etiqueta "E".

**Por quê.** Responde "por que este cliente está neste segmento?" sem obrigar ninguém a ler `ticketMedio >= 500 && frequenciaCompras >= 3`. O texto original fica no `title`.
