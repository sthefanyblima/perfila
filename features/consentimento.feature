# language: pt
@UC-08 @lgpd @aplicacao
Funcionalidade: Ciclo de vida do consentimento LGPD
  Para tratar dados pessoais de forma legítima
  O sistema controla o consentimento do cliente como uma máquina de estados

  Contexto:
    Dado que a data atual é "2026-03-10T12:00:00Z"
    E que existe o cliente "c-001" ativo

  @RN-02
  Cenário: Conceder consentimento pela primeira vez
    Dado que o consentimento do cliente "c-001" está no estado "NAO_CONCEDIDO"
    Quando o cliente "c-001" concede consentimento para a finalidade "segmentacao" com o termo "v1"
    Então o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    E o consentimento do cliente "c-001" é válido
    E a data de consentimento é "2026-03-10T12:00:00Z"

  @RN-03
  Cenário: Revogar consentimento concedido
    Dado que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    Quando o cliente "c-001" revoga o consentimento
    Então o consentimento do cliente "c-001" está no estado "REVOGADO"
    E o consentimento do cliente "c-001" não é válido
    E a data de revogação é "2026-03-10T12:00:00Z"

  Cenário: Conceder novamente após a revogação
    Dado que o consentimento do cliente "c-001" está no estado "REVOGADO"
    Quando o cliente "c-001" concede consentimento para a finalidade "segmentacao" com o termo "v2"
    Então o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    E o consentimento do cliente "c-001" é válido

  Esquema do Cenário: Transições inválidas são rejeitadas e o estado não muda
    Dado que o consentimento do cliente "c-001" está no estado "<estado>"
    Quando o cliente "c-001" tenta "<acao>" o consentimento
    Então a operação falha com o erro "transicao_invalida"
    E o consentimento do cliente "c-001" permanece no estado "<estado>"

    Exemplos:
      | estado        | acao     |
      | CONCEDIDO     | conceder |
      | NAO_CONCEDIDO | revogar  |
      | REVOGADO      | revogar  |

  @RN-02
  Esquema do Cenário: A validade depende somente do estado
    Dado que o consentimento do cliente "c-001" está no estado "<estado>"
    Quando o consentimento do cliente "c-001" é consultado
    Então a validade retornada é "<valido>"

    Exemplos:
      | estado        | valido     |
      | NAO_CONCEDIDO | falso      |
      | CONCEDIDO     | verdadeiro |
      | REVOGADO      | falso      |

  Esquema do Cenário: Conceder exige finalidade e versão do termo
    Dado que o consentimento do cliente "c-001" está no estado "NAO_CONCEDIDO"
    Quando o cliente "c-001" concede consentimento para a finalidade "<finalidade>" com o termo "<termo>"
    Então a operação falha com o erro "dados_invalidos"
    E o consentimento do cliente "c-001" permanece no estado "NAO_CONCEDIDO"

    Exemplos:
      | finalidade  | termo |
      |             | v1    |
      | segmentacao |       |

  Cenário: Cliente sem registro de consentimento é tratado como não concedido
    Quando o consentimento do cliente "c-001" é consultado
    Então a validade retornada é "falso"
    E o consentimento do cliente "c-001" está no estado "NAO_CONCEDIDO"

  Cenário: Cliente inexistente não gerencia consentimento
    Quando o cliente "c-999" concede consentimento para a finalidade "segmentacao" com o termo "v1"
    Então a operação falha com o erro "cliente_nao_encontrado"
