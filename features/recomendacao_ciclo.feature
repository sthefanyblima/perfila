# language: pt
@estado @RN-08
Funcionalidade: Ciclo de vida da recomendação de produto

  Contexto:
    Dado que a data atual é "2026-03-10T12:00:00Z"
    E que a validade padrão é de 30 dias

  @dominio
  Cenário: Disponibilizar uma recomendação gerada
    Dado que existe uma recomendação no estado "GERADA"
    Quando a recomendação é disponibilizada
    Então a recomendação está no estado "ATIVA"
    E a data de ativação é "2026-03-10T12:00:00Z"
    E a validade é "2026-04-09T12:00:00Z"

  @dominio
  Cenário: Expirar uma recomendação ativa
    Dado que existe uma recomendação no estado "ATIVA"
    Quando a recomendação é expirada
    Então a recomendação está no estado "EXPIRADA"
    E a data de expiração é "2026-03-10T12:00:00Z"

  @dominio
  Esquema do Cenário: Transições inválidas são rejeitadas
    Dado que existe uma recomendação no estado "<estado>"
    Quando tenta-se "<acao>" a recomendação
    Então a operação falha com o erro "transicao_invalida"
    E a recomendação está no estado "<estado>"

    Exemplos:
      | estado   | acao           |
      | GERADA   | expirar        |
      | ATIVA    | disponibilizar |
      | EXPIRADA | disponibilizar |
      | EXPIRADA | expirar        |

  @aplicacao
  Esquema do Cenário: O comando de expiração respeita a validade
    Dado que existe uma recomendação ativada em "2026-03-10T12:00:00Z" com validade de 30 dias
    Quando passam <dias> dias e o comando "expirar-recomendacoes" é executado
    Então a recomendação está no estado "<estado>"

    Exemplos:
      | dias | estado   |
      | 29   | ATIVA    |
      | 30   | EXPIRADA |
      | 31   | EXPIRADA |
