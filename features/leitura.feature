# language: pt
@aplicacao
Funcionalidade: Leitura dos dados para o painel do operador
  O painel lê o que os casos de uso gravaram sem quebrar a regra do consentimento: o cadastro
  aparece sempre; o comportamento só com consentimento válido (Spec 3.5, Q-21, Q-23)

  Contexto:
    Dado que a data atual é "2026-03-10T12:00:00Z"
    E que existem os segmentos ativos:
      | prioridade | nome    | criterio                                     |
      | 1          | premium | ticketMedio >= 500 && frequenciaCompras >= 3 |
      | 99         | padrao  |                                              |
    E que o catálogo possui produtos ativos na categoria "livros"

  @Q-21 @Q-23
  Cenário: A lista mostra o consentimento e só mostra o segmento com consentimento válido
    Dado que existe o cliente "c-001" ativo com eventos de compra na categoria "livros"
    E que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    E que o cliente "c-001" foi processado
    E que existe o cliente "c-002" ativo com eventos de compra na categoria "livros"
    E que o consentimento do cliente "c-002" está no estado "CONCEDIDO"
    E que o cliente "c-002" foi processado
    E que o consentimento do cliente "c-002" está no estado "REVOGADO"
    Quando os clientes são listados
    Então a lista tem 2 clientes
    E a linha do cliente "c-001" mostra o consentimento "CONCEDIDO" e o segmento "padrao"
    E a linha do cliente "c-002" mostra o consentimento "REVOGADO" e o segmento ""

  @Q-23
  Esquema do Cenário: A lista pode ser filtrada pelo estado do consentimento
    Dado que existe o cliente "c-001" ativo
    E que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    E que existe o cliente "c-002" ativo
    E que existe o cliente "c-003" ativo
    E que o consentimento do cliente "c-003" está no estado "REVOGADO"
    Quando os clientes com consentimento "<estado>" são listados
    Então a lista tem <total> clientes

    Exemplos:
      | estado        | total |
      | CONCEDIDO     | 1     |
      | NAO_CONCEDIDO | 1     |
      | REVOGADO      | 1     |
      | TODOS         | 3     |

  @Q-23
  Esquema do Cenário: A lista é paginada de 20 em 20
    Dado que existem 25 clientes ativos cadastrados
    Quando a página <pagina> de clientes é listada
    Então a página tem <linhas> clientes de um total de 25

    Exemplos:
      | pagina | linhas |
      | 1      | 20     |
      | 2      | 5      |
      | 3      | 0      |

  @Q-23
  Cenário: A busca por nome não diferencia maiúsculas nem acentos
    Dado que existe o cliente "c-001" ativo com o nome "Joana Sintética"
    E que existe o cliente "c-002" ativo com o nome "Pedro Sintético"
    Quando os clientes são buscados por "JOANA SINTETICA"
    Então a lista tem 1 cliente

  @Q-21 @UC-08
  Cenário: A ficha mostra o cadastro mesmo sem consentimento
    Dado que existe o cliente "c-001" ativo
    Quando a ficha do cliente "c-001" é consultada
    Então a consulta não falha
    E a ficha mostra o consentimento "NAO_CONCEDIDO"

  @Q-21
  Cenário: Ficha de cliente inexistente
    Quando a ficha do cliente "c-999" é consultada
    Então a operação falha com o erro "cliente_nao_encontrado"

  @lgpd @RN-01
  Esquema do Cenário: O resultado do cliente exige consentimento válido
    Dado que existe o cliente "c-001" ativo com eventos de compra na categoria "livros"
    E que o consentimento do cliente "c-001" está no estado "<estado>"
    Quando o resultado do cliente "c-001" é consultado
    Então a operação falha com o erro "consentimento_invalido"

    Exemplos:
      | estado        |
      | NAO_CONCEDIDO |
      | REVOGADO      |

  @RN-01
  Cenário: O resultado de um cliente ainda não processado traz só os eventos
    Dado que existe o cliente "c-001" ativo com eventos de compra na categoria "livros"
    E que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    Quando o resultado do cliente "c-001" é consultado
    Então a consulta não falha
    E o resultado traz 3 eventos, sem perfil e sem recomendações

  @UC-02 @UC-03 @UC-04
  Cenário: O resultado de um cliente processado traz perfil, segmento e histórico de recomendações
    Dado que existe o cliente "c-001" ativo com eventos de compra na categoria "livros"
    E que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    E que o cliente "c-001" foi processado
    E que o cliente "c-001" foi processado de novo 1 hora depois
    Quando o resultado do cliente "c-001" é consultado
    Então o resultado traz o segmento "padrao" e 2 recomendações, a mais recente "ATIVA"

  @RN-06
  Cenário: Os segmentos são listados por prioridade
    Quando os segmentos são listados
    Então os segmentos listados são "premium" e "padrao"

  @RN-07
  Cenário: O catálogo lista todos os produtos por categoria e nome
    Dado que o catálogo possui os produtos:
      | nome   | categoria | preco | ativo |
      | Jogo X | games     | 200   | não   |
    Quando o catálogo é listado
    Então o catálogo listado é "games/Jogo X", "livros/Item 1" e "livros/Item 2"
