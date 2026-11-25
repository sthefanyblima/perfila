# RecommendationTicket

A recomendação do cliente, em forma de ingresso violeta com picote.

**O que fornecer.** `pf-ticket` com a data e a validade em `pf-ticket__when` (mono), o produto em `pf-ticket__title` (display), o preço e o motivo legível ("Segmento premium, categoria livros"), um picote tracejado (`hr`) e, embaixo, o estado: Ativa (`StatusTag` em branco) ou, se Gerada, o botão "Ativar recomendação". As duas meias-luas laterais são recortes na cor do painel.

**Sem produto elegível.** Não é erro: no lugar do ingresso vai um `Alert` de atenção "Sem produto elegível" com o motivo ("Nenhum produto ativo na categoria preferida (games).").

**Movimento.** Ao concluir o processamento ou ativar a recomendação, o ingresso pulsa e sai um pequeno estouro de confete dele (desligado com movimento reduzido).

**Confete no painel.** Ao voltar do processamento (aviso "Processamento concluído"), 22 pontos (`pf-confete`) saem do ingresso. O painel não tem ação "Ativar recomendação": a recomendação nasce ativa, então o confete só acontece no processamento.
