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

  @RN-08
  Cenário: Uma nova recomendação substitui a ativa anterior
    Dado que uma recomendação do cliente "c-001" já foi gerada
    Quando a recomendação do cliente "c-001" é gerada
    Então o cliente "c-001" tem 1 recomendação no estado "ATIVA"
    E a recomendação anterior está no estado "EXPIRADA"

  Cenário: Perfil segmentado num segmento inexistente é falha de configuração
    Dado que o perfil do cliente "c-001" aponta para um segmento inexistente
    Quando a recomendação do cliente "c-001" é gerada
    Então a operação falha por erro de configuração
    E nenhuma recomendação é gravada

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

  # ---- Edição do catálogo pelo painel (v1.6, Q-43, RN-14) ----

  @RN-07 @RN-14 @Q-43
  Cenário: Produto desativado deixa de ser recomendado
    Dado que o operador "caio" com o papel "ANALISTA" desativou o produto "Livro A"
    Quando a recomendação do cliente "c-001" é gerada
    Então o produto recomendado é "Livro B"

  @RN-07 @RN-14 @Q-43
  Cenário: Produto criado com a categoria normalizada entra na recomendação
    Dado que o operador "caio" com o papel "ANALISTA" criou o produto "Atlas" na categoria "  Livros " com preço 90
    Quando a recomendação do cliente "c-001" é gerada
    Então o produto recomendado é "Atlas"

  @RN-14 @Q-43
  Esquema do Cenário: Produto com dados fora dos limites é recusado
    Quando o operador "caio" com o papel "ANALISTA" cria o produto "<nome>" na categoria "<categoria>" com preço <preco>
    Então a operação falha com o erro "dados_invalidos" no campo "<campo>"

    Exemplos:
      | nome   | categoria | preco      | campo     |
      | Atlas  | livros    | -1         | preco     |
      | Atlas  | livros    | 1000000.01 | preco     |
      |        | livros    | 10         | nome      |
      | Atlas  |           | 10         | categoria |

  @RN-14 @Q-43
  Cenário: O preço editado vale para o produto gravado
    Quando o operador "caio" com o papel "ANALISTA" muda o preço do produto "Livro B" para 1000000
    Então o produto "Livro B" está gravado com o preço 1000000

  @RN-11 @Q-43
  Cenário: Quem só atende não edita o catálogo
    Quando o operador "rita" com o papel "ATENDIMENTO" muda o preço do produto "Livro B" para 10
    Então a operação falha com o erro "sem_permissao"
