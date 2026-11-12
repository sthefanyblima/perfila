# language: pt
@api
Funcionalidade: Processamento do cliente pela API
  Os casos de uso do core expostos por HTTP, com autenticação e checagem LGPD (Spec 3.7)

  Contexto:
    Dado que a data atual é "2026-03-10T12:00:00Z"
    E que a API está no ar e exige o token "token-de-teste"
    E que existe o cliente "c-001" ativo com eventos de compra na categoria "livros"
    E que o catálogo possui produtos ativos na categoria "livros"
    E que as requisições que exigem corpo usam a finalidade "segmentacao" e o termo "v1"

  @UC-01 @UC-02 @UC-03 @UC-04 @RN-09 @smoke
  Cenário: Processar um cliente com consentimento válido
    Dado que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    Quando "POST /v1/clientes/c-001/processamento" é chamado com o token "token-de-teste"
    Então a resposta tem status 200
    E o corpo contém "perfil", "segmento" e "recomendacao"

  @lgpd @RN-01
  Esquema do Cenário: Sem consentimento válido nada é processado
    Dado que o consentimento do cliente "c-001" está no estado "<estado>"
    Quando "POST /v1/clientes/c-001/processamento" é chamado com o token "token-de-teste"
    Então a resposta tem status 403
    E o corpo tem o código "consentimento_invalido"
    E nenhum perfil é gravado para o cliente "c-001"

    Exemplos:
      | estado        |
      | NAO_CONCEDIDO |
      | REVOGADO      |

  Esquema do Cenário: A autenticação é obrigatória
    Dado que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    Quando "POST /v1/clientes/c-001/processamento" é chamado com o token "<token>"
    Então a resposta tem status 401

    Exemplos:
      | token  |
      |        |
      | errado |

  @RN-01
  Cenário: Cliente inexistente
    Quando "POST /v1/clientes/c-999/processamento" é chamado com o token "token-de-teste"
    Então a resposta tem status 404
    E o corpo tem o código "cliente_nao_encontrado"

  @RN-01
  Cenário: Cliente inativo
    Dado que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    E que o cliente "c-001" está inativo
    Quando "POST /v1/clientes/c-001/processamento" é chamado com o token "token-de-teste"
    Então a resposta tem status 409
    E o corpo tem o código "cliente_inativo"

  @RN-07
  Cenário: Sem produto elegível a recomendação é nula
    Dado que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    E que todos os produtos da categoria "livros" estão inativos
    Quando "POST /v1/clientes/c-001/processamento" é chamado com o token "token-de-teste"
    Então a resposta tem status 200
    E a recomendação no corpo é nula

  @UC-08
  Esquema do Cenário: As rotas de consentimento seguem a máquina de estados sem exigir consentimento prévio
    Dado que o consentimento do cliente "c-001" está no estado "<estado>"
    Quando "<requisicao>" é chamado com o token "token-de-teste"
    Então a resposta tem status <status>

    Exemplos:
      | estado        | requisicao                                      | status |
      | NAO_CONCEDIDO | PUT /v1/clientes/c-001/consentimento            | 200    |
      | CONCEDIDO     | PUT /v1/clientes/c-001/consentimento            | 409    |
      | CONCEDIDO     | POST /v1/clientes/c-001/consentimento/revogacao | 200    |
      | NAO_CONCEDIDO | POST /v1/clientes/c-001/consentimento/revogacao | 409    |
      | REVOGADO      | GET /v1/clientes/c-001/consentimento            | 200    |

  @UC-09
  Cenário: Cadastrar um cliente pela API
    Quando "POST /v1/clientes" é chamado com o token "token-de-teste" e o corpo:
      """
      {"nome": "Cliente Sintético", "email": "c-010@exemplo.test"}
      """
    Então a resposta tem status 201
    E o corpo contém "id", "dataCadastro" e "ativo"

  @UC-09
  Esquema do Cenário: Corpo inválido no cadastro é rejeitado
    Quando "POST /v1/clientes" é chamado com o token "token-de-teste" e o corpo:
      """
      <corpo>
      """
    Então a resposta tem status 422
    E o corpo tem o código "dados_invalidos"

    Exemplos:
      | corpo                                   |
      | {"nome": "", "email": "x@exemplo.test"} |
      | {"nome": "Ana"                          |
      | {"nome": "Ana", "cpf": "000"}           |
      | {"nome": "Ana", "email": "a@exemplo.test"} {"extra": 1} |

  @UC-01 @RN-04
  Cenário: Registrar um evento pela API
    Dado que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    Quando "POST /v1/clientes/c-001/eventos" é chamado com o token "token-de-teste" e o corpo:
      """
      {"tipo": "navegacao", "origem": "web", "valor": 0, "timestamp": "2026-03-09T10:00:00Z"}
      """
    Então a resposta tem status 201
    E o corpo contém "id", "tipo" e "timestamp"

  @UC-01 @RN-04
  Cenário: Evento inválido pela API
    Dado que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    Quando "POST /v1/clientes/c-001/eventos" é chamado com o token "token-de-teste" e o corpo:
      """
      {"tipo": "compra", "origem": "web", "valor": 0, "timestamp": "2026-03-09T10:00:00Z"}
      """
    Então a resposta tem status 422
    E o corpo tem o código "dados_invalidos"

  Cenário: A rota de saúde não exige token
    Quando "GET /saude" é chamado com o token ""
    Então a resposta tem status 200
