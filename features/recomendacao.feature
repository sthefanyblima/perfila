# language: pt
@UC-04 @aplicacao
Funcionalidade: Recomendação de produtos personalizados

  Contexto:
    Dado que a data atual é "2026-03-10T12:00:00Z"
    E que existe o cliente "c-001" ativo
    E que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    E que o perfil do cliente "c-001" tem categoria preferida "livros" e segmento "premium"
    E que o catálogo possui os produtos:
      | nome    | categoria | preco | ativo |
      | Livro B | livros    | 50    | sim   |
      | Livro A | livros    | 80    | sim   |
      | Livro C | livros    | 10    | não   |
      | Jogo X  | games     | 200   | sim   |

  @RN-07 @RN-08
  Cenário: Recomendar o primeiro produto ativo da categoria preferida
    Quando a recomendação do cliente "c-001" é gerada
    Então o produto recomendado é "Livro A"
    E a recomendação está no estado "ATIVA"
    E o motivo é "segmento:premium;categoria:livros"
    E a validade é "2026-04-09T12:00:00Z"

  @RN-07
  Cenário: Produtos inativos não são elegíveis
    Dado que todos os produtos da categoria "livros" estão inativos
    Quando a recomendação do cliente "c-001" é gerada
    Então a operação falha com o erro "sem_produto_elegivel"
    E nenhuma recomendação é gravada

  @RN-07
  Cenário: Sem categoria preferida não há recomendação
    Dado que o perfil do cliente "c-001" tem categoria preferida vazia e segmento "padrao"
    Quando a recomendação do cliente "c-001" é gerada
    Então a operação falha com o erro "sem_produto_elegivel"
    E nenhuma recomendação é gravada

  @RN-08
  Cenário: A recomendação é gravada uma única vez, já ativa
    Quando a recomendação do cliente "c-001" é gerada
    Então a recomendação é gravada 1 vez
    E a recomendação gravada está no estado "ATIVA"

  Cenário: Perfil ainda não segmentado não recebe recomendação
    Dado que o perfil do cliente "c-001" ainda não foi segmentado
    Quando a recomendação do cliente "c-001" é gerada
    Então a operação falha com o erro "perfil_nao_encontrado"

  @lgpd @RN-01
  Cenário: Sem consentimento válido nenhuma recomendação é gerada
    Dado que o consentimento do cliente "c-001" está no estado "REVOGADO"
    Quando a recomendação do cliente "c-001" é gerada
    Então a operação falha com o erro "consentimento_invalido"
    E nenhuma recomendação é gravada
