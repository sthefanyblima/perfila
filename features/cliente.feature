# language: pt
@UC-09 @aplicacao
Funcionalidade: Cadastro de cliente
  Para que um cliente possa ser perfilado
  Ele primeiro entra no sistema com dados mínimos e sem consentimento

  Contexto:
    Dado que a data atual é "2026-03-10T12:00:00Z"

  Cenário: Cadastrar um cliente com nome e e-mail
    Quando um cliente é cadastrado com nome "Cliente Sintético" e e-mail "c-010@exemplo.test"
    Então o cliente cadastrado está ativo
    E a data de cadastro é "2026-03-10T12:00:00Z"

  Esquema do Cenário: Nome e e-mail são obrigatórios
    Quando um cliente é cadastrado com nome "<nome>" e e-mail "<email>"
    Então a operação falha com o erro "dados_invalidos"

    Exemplos:
      | nome              | email              |
      |                   | c-010@exemplo.test |
      | Cliente Sintético |                    |

  @lgpd @RN-01
  Cenário: Cliente recém-cadastrado começa sem consentimento
    Quando um cliente é cadastrado com nome "Cliente Sintético" e e-mail "c-010@exemplo.test"
    Então o consentimento do cliente cadastrado está no estado "NAO_CONCEDIDO"

  @lgpd @RN-01
  Cenário: Cliente recém-cadastrado não é perfilado antes de consentir
    Dado que um cliente foi cadastrado com nome "Cliente Sintético" e e-mail "c-010@exemplo.test"
    Quando o perfil do cliente cadastrado é gerado
    Então a operação falha com o erro "consentimento_invalido"

  @lgpd
  Cenário: Depois de consentir, o cliente cadastrado é perfilado
    Dado que um cliente foi cadastrado com nome "Cliente Sintético" e e-mail "c-010@exemplo.test"
    E que o cliente cadastrado concedeu consentimento para a finalidade "segmentacao" com o termo "v1"
    Quando o perfil do cliente cadastrado é gerado
    Então a frequência de compras é 0
