# language: pt
@UC-01 @aplicacao
Funcionalidade: Coleta de eventos comportamentais

  Contexto:
    Dado que a data atual é "2026-03-10T12:00:00Z"
    E que existe o cliente "c-001" ativo
    E que o consentimento do cliente "c-001" está no estado "CONCEDIDO"

  Cenário: Registrar um evento de compra válido
    Quando o sistema registra para o cliente "c-001" um evento "compra" de origem "ecommerce" com valor 150 em "2026-03-09T10:00:00Z"
    Então o evento é gravado
    E a coleta do cliente "c-001" retorna 1 evento

  @RN-04
  Esquema do Cenário: Eventos válidos nos limites são aceitos
    Quando o sistema registra para o cliente "c-001" um evento "<tipo>" de origem "web" com valor <valor> em "<quando>"
    Então o evento é gravado

    Exemplos:
      | tipo      | valor | quando               |
      | navegacao | 0     | 2026-03-10T12:00:00Z |
      | interacao | 0     | 2026-03-09T10:00:00Z |
      | compra    | 0.01  | 2026-03-09T10:00:00Z |
      | compra    | 1000000 | 2026-03-09T10:00:00Z |
      | navegacao | 0     | 2026-03-10T12:02:00Z |

  @RN-04
  Esquema do Cenário: Eventos inválidos são rejeitados
    Quando o sistema registra para o cliente "c-001" um evento "<tipo>" de origem "<origem>" com valor <valor> em "<quando>"
    Então a operação falha com o erro "dados_invalidos"
    E nenhum evento é gravado

    Exemplos:
      | tipo         | origem    | valor | quando               |
      | desconhecido | ecommerce | 0     | 2026-03-09T10:00:00Z |
      | compra       | ecommerce | 0     | 2026-03-09T10:00:00Z |
      | compra       | ecommerce | -10   | 2026-03-09T10:00:00Z |
      | navegacao    | web       | -1    | 2026-03-09T10:00:00Z |
      | navegacao    |           | 0     | 2026-03-09T10:00:00Z |
      | navegacao    | web       | 0     | 2026-03-11T00:00:00Z |
      | navegacao    | web       | 0     | 2026-03-10T12:02:01Z |
      | compra       | ecommerce | 1000000.01 | 2026-03-09T10:00:00Z |

  @lgpd @RN-01 @RN-03
  Esquema do Cenário: Sem consentimento válido nada é gravado nem lido
    Dado que o consentimento do cliente "c-001" está no estado "<estado>"
    Quando o sistema tenta "<acao>" eventos do cliente "c-001"
    Então a operação falha com o erro "consentimento_invalido"
    E nenhum evento é gravado

    Exemplos:
      | estado        | acao      |
      | NAO_CONCEDIDO | registrar |
      | NAO_CONCEDIDO | coletar   |
      | REVOGADO      | registrar |
      | REVOGADO      | coletar   |

  @RN-01
  Cenário: Cliente inativo é rejeitado
    Dado que o cliente "c-001" está inativo
    Quando o sistema registra para o cliente "c-001" um evento "navegacao" de origem "web" com valor 0 em "2026-03-09T10:00:00Z"
    Então a operação falha com o erro "cliente_inativo"

  @RN-01
  Cenário: Cliente inexistente é rejeitado
    Quando o sistema coleta os eventos do cliente "c-999"
    Então a operação falha com o erro "cliente_nao_encontrado"

  Cenário: A coleta devolve apenas os eventos do próprio cliente
    Dado que existem 2 eventos do cliente "c-001" e 3 eventos do cliente "c-002"
    Quando o sistema coleta os eventos do cliente "c-001"
    Então a coleta retorna 2 eventos
