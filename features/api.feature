# language: pt
# Fase: Support (10/11–17/11): API HTTP, middlewares e CMD. Fora do filtro do CI até a fase começar (ver Makefile, BDD_TAGS).
@api @smoke
Funcionalidade: Processamento do cliente pela API

  Contexto:
    Dado que a API está no ar e exige o token "token-de-teste"
    E que existe o cliente "c-001" ativo com eventos de compra na categoria "livros"
    E que o catálogo possui produtos ativos na categoria "livros"
    E que as requisições que exigem corpo usam a finalidade "segmentacao" e o termo "v1"

  @UC-01 @UC-02 @UC-03 @UC-04 @RN-09
  Cenário: Processar um cliente com consentimento válido
    Dado que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    Quando "POST /v1/clientes/c-001/processamento" é chamado com o token "token-de-teste"
    Então a resposta tem status 200
    E o corpo contém "perfil", "segmento" e "recomendacao"

  @lgpd
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

  Cenário: Cliente inexistente
    Quando "POST /v1/clientes/c-999/processamento" é chamado com o token "token-de-teste"
    Então a resposta tem status 404
    E o corpo tem o código "cliente_nao_encontrado"

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
