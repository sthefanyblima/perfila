# language: pt
@UC-02 @RN-05 @aplicacao
Funcionalidade: Perfilamento do comportamento do cliente

  Contexto:
    Dado que a data atual é "2026-03-10T12:00:00Z"
    E que existe o cliente "c-001" ativo
    E que o consentimento do cliente "c-001" está no estado "CONCEDIDO"

  Cenário: Calcular o perfil a partir de compras e navegação
    Dado que a meta de engajamento é de 10 eventos
    E que o cliente "c-001" possui os eventos:
      | tipo      | valor | categoria | quando               |
      | compra    | 100   | livros    | 2026-03-01T10:00:00Z |
      | compra    | 200   | livros    | 2026-03-05T10:00:00Z |
      | compra    | 300   | games     | 2026-03-08T10:00:00Z |
      | navegacao | 0     | games     | 2026-03-09T10:00:00Z |
    Quando o perfil do cliente "c-001" é gerado
    Então a frequência de compras é 3
    E o ticket médio é 200
    E a categoria preferida é "livros"
    E o score de engajamento é 0.4

  Esquema do Cenário: Só eventos dos últimos 30 dias contam para o engajamento
    Dado que a meta de engajamento é de 10 eventos
    E que o cliente "c-001" possui um evento "navegacao" ocorrido há <dias> dias
    Quando o perfil do cliente "c-001" é gerado
    Então o score de engajamento é <score>

    Exemplos:
      | dias | score |
      | 29   | 0.1   |
      | 30   | 0     |
      | 31   | 0     |

  Cenário: O score de engajamento nunca passa de 1
    Dado que a meta de engajamento é de 10 eventos
    E que o cliente "c-001" possui 12 eventos "navegacao" nos últimos 5 dias
    Quando o perfil do cliente "c-001" é gerado
    Então o score de engajamento é 1

  Cenário: Cliente sem eventos recebe perfil zerado
    Dado que o cliente "c-001" não possui eventos
    Quando o perfil do cliente "c-001" é gerado
    Então a frequência de compras é 0
    E o ticket médio é 0
    E a categoria preferida é vazia
    E o score de engajamento é 0

  Cenário: Empate de categoria é resolvido por ordem alfabética
    Dado que o cliente "c-001" possui os eventos:
      | tipo   | valor | categoria | quando               |
      | compra | 50    | livros    | 2026-03-02T10:00:00Z |
      | compra | 70    | games     | 2026-03-03T10:00:00Z |
    Quando o perfil do cliente "c-001" é gerado
    Então a categoria preferida é "games"

  Cenário: Gerar o perfil de novo atualiza o existente
    Dado que o cliente "c-001" já possui um perfil com frequência de compras 1
    E que o cliente "c-001" possui os eventos:
      | tipo   | valor | categoria | quando               |
      | compra | 50    | livros    | 2026-03-02T10:00:00Z |
      | compra | 60    | livros    | 2026-03-03T10:00:00Z |
      | compra | 70    | livros    | 2026-03-04T10:00:00Z |
    Quando o perfil do cliente "c-001" é gerado
    Então existe um único perfil para o cliente "c-001"
    E o identificador do perfil é preservado
    E a frequência de compras é 3

  @lgpd @RN-01
  Cenário: Sem consentimento válido o perfil não é gerado
    Dado que o consentimento do cliente "c-001" está no estado "REVOGADO"
    Quando o perfil do cliente "c-001" é gerado
    Então a operação falha com o erro "consentimento_invalido"
    E nenhum perfil é gravado para o cliente "c-001"
