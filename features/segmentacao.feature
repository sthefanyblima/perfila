# language: pt
@UC-03 @RN-06
Funcionalidade: Segmentação automática de clientes

  Contexto:
    Dado que a data atual é "2026-03-10T12:00:00Z"
    E que existe o cliente "c-001" ativo
    E que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    E que existem os segmentos ativos:
      | prioridade | nome     | criterio                                     |
      | 1          | premium  | ticketMedio >= 500 && frequenciaCompras >= 3 |
      | 2          | engajado | scoreEngajamento >= 0.7                      |
      | 99         | padrao   |                                              |

  @aplicacao
  Esquema do Cenário: O perfil recebe o primeiro segmento cujo critério casa
    Dado que o perfil do cliente "c-001" tem ticket médio <ticket>, frequência de compras <frequencia> e score de engajamento <score>
    Quando o cliente "c-001" é segmentado
    Então o segmento atribuído é "<segmento>"

    Exemplos:
      | ticket | frequencia | score | segmento |
      | 500    | 3          | 0.1   | premium  |
      | 499.99 | 3          | 0.1   | padrao   |
      | 500    | 2          | 0.1   | padrao   |
      | 500    | 3          | 0.9   | premium  |
      | 100    | 1          | 0.7   | engajado |
      | 100    | 1          | 0.69  | padrao   |
      | 0      | 0          | 0     | padrao   |

  @aplicacao
  Cenário: A atribuição fica registrada no perfil
    Dado que o perfil do cliente "c-001" tem ticket médio 600, frequência de compras 4 e score de engajamento 0.2
    Quando o cliente "c-001" é segmentado
    Então o perfil do cliente "c-001" está associado ao segmento "premium"

  @aplicacao
  Cenário: Cliente sem perfil não pode ser segmentado
    Dado que o cliente "c-001" não possui perfil
    Quando o cliente "c-001" é segmentado
    Então a operação falha com o erro "perfil_nao_encontrado"

  @aplicacao
  Cenário: Segmento configurado com critério inválido é falha de configuração
    Dado que o perfil do cliente "c-001" tem ticket médio 600, frequência de compras 4 e score de engajamento 0.2
    E que existe o segmento ativo "quebrado" com prioridade 0 e critério "idade > 18"
    Quando o cliente "c-001" é segmentado
    Então a operação falha por erro de configuração
    E o perfil do cliente "c-001" continua sem segmento

  @aplicacao @lgpd @RN-01
  Cenário: Sem consentimento válido o cliente não é segmentado
    Dado que o perfil do cliente "c-001" tem ticket médio 600, frequência de compras 4 e score de engajamento 0.2
    E que o consentimento do cliente "c-001" está no estado "REVOGADO"
    Quando o cliente "c-001" é segmentado
    Então a operação falha com o erro "consentimento_invalido"
    E o perfil do cliente "c-001" continua sem segmento

  @dominio
  Esquema do Cenário: Critérios dentro da gramática são aceitos
    Quando o critério "<criterio>" é avaliado
    Então a avaliação do critério não falha

    Exemplos:
      | criterio                                             |
      | ticketMedio > 1                                      |
      | frequenciaCompras==2                                 |
      | scoreEngajamento <= 0.5 && ticketMedio < 10.25       |

  @dominio
  Esquema do Cenário: Critérios fora da gramática são rejeitados
    Quando o critério "<criterio>" é avaliado
    Então a operação falha com o erro "dados_invalidos"

    Exemplos:
      | criterio           |
      | idade > 18         |
      | ticketMedio >> 1   |
      | ticketMedio > abc  |
      | ticketMedio >      |
      | ticketMedio > 1 && |
      | ticketMedio != 1   |
