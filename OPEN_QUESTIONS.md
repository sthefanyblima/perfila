# Perguntas em aberto

Registro exigido por C-12. Cada pergunta tem uma decisão padrão já implementada; a equipe (Guardião(ã) da Spec) confirma ou troca. Mudança aprovada entra primeiro na Spec, depois nas features, depois no código.

| ID | Origem | Pergunta | Padrão implementado | Dono | Prazo |
|---|---|---|---|---|---|
| Q-01 | UML5 / 3.4.1 | Ao conceder de novo após revogar, `DataRevogacao` deve ser apagada ou mantida como histórico? | Mantida (registra a última revogação) | Responsável LGPD | 17/11 |
| Q-02 | RN-06 | Se nenhum segmento ativo casar (sem o `padrao`), qual erro? | `ErrDadosInvalidos` | Guardião(ã) da Spec | 17/11 |
| Q-03 | D-12 | Semântica de `IntegracaoGateway.BuscarEventos(desde)`: quem guarda o último "desde"? Os eventos importados passam por `Registrar`? | Stub vazio no MVP; definir antes de T-20 | Guardião(ã) da Spec | 10/11 |
| Q-04 | D-17 | Confirmar atores de UC-08 (Cliente e Administrador) e incluir o "Agente Inteligente" na imagem do UML1 | Ambos os atores; agente como ator de automação | Equipe | 26/11 |
| Q-05 | RN-07 | A regra provisória de recomendação (primeiro por nome) vai para a apresentação final ou trocamos por "mais vendido da categoria"? | Primeiro por nome (determinístico, fácil de testar) | Equipe | 17/11 |
| Q-06 | 3.7 | `PUT /consentimento` em estado CONCEDIDO responde 409 (`transicao_invalida`). Queremos idempotência (200) para o cliente reenviar? | 409, como na máquina de estados | Guardião(ã) do BDD | 10/11 |
