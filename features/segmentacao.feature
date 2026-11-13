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

  # ---- Edição de segmentos pelo painel (v1.6, Q-33, Q-42, RN-14) ----

  @aplicacao @RN-14 @Q-33
  Cenário: Um segmento novo entra na avaliação pela prioridade
    Dado que o perfil do cliente "c-001" tem ticket médio 100, frequência de compras 1 e score de engajamento 0.2
    E que o operador "caio" com o papel "ANALISTA" criou o segmento "ocasional" com prioridade 3 e critério "frequenciaCompras >= 1"
    Quando o cliente "c-001" é segmentado
    Então o segmento atribuído é "ocasional"

  @aplicacao @RN-14 @Q-33
  Esquema do Cenário: Critério fora da gramática é recusado ao salvar, com a condição errada
    Quando o operador "caio" com o papel "ANALISTA" muda o critério do segmento "engajado" para "<criterio>"
    Então a operação falha com o erro "dados_invalidos" no campo "criterio"
    E a mensagem do erro cita "<condicao>"
    E o segmento "engajado" continua com o critério "scoreEngajamento >= 0.7"

    Exemplos:
      | criterio                                | condicao        |
      | idade > 18                              | idade > 18      |
      | scoreEngajamento >= 0.5 && ticket > 1   | ticket > 1      |
      | scoreEngajamento >= 0.5 &&              | &&              |

  @aplicacao @RN-14 @Q-33
  Esquema do Cenário: O segmento padrao é protegido
    Quando o operador "caio" com o papel "ANALISTA" <acao>
    Então a operação falha com o erro "dados_invalidos" no campo "<campo>"
    E os segmentos ativos são "premium", "engajado" e "padrao"

    Exemplos:
      | acao                                                       | campo      |
      | desativa o segmento "padrao"                               | ativo      |
      | muda o critério do segmento "padrao" para "ticketMedio > 0" | criterio   |
      | muda o critério do segmento "engajado" para ""              | criterio   |
      | muda a prioridade do segmento "engajado" para 100           | prioridade |

  @aplicacao @RN-14 @Q-33
  Esquema do Cenário: Nome e prioridade não se repetem entre os segmentos ativos
    Quando o operador "caio" com o papel "ANALISTA" cria o segmento "<nome>" com prioridade <prioridade> e critério "frequenciaCompras >= 1"
    Então a operação falha com o erro "dados_invalidos" no campo "<campo>"

    Exemplos:
      | nome      | prioridade | campo      |
      | ocasional | 2          | prioridade |
      | Premium   | 5          | nome       |
      | vip;novo  | 5          | nome       |
      | ocasional | 0          | prioridade |

  @aplicacao @RN-14 @Q-33
  Cenário: Desativar tira o segmento da avaliação
    Dado que o perfil do cliente "c-001" tem ticket médio 100, frequência de compras 1 e score de engajamento 0.9
    E que o operador "caio" com o papel "ANALISTA" desativou o segmento "engajado"
    Quando o cliente "c-001" é segmentado
    Então o segmento atribuído é "padrao"

  @aplicacao @RN-14 @Q-33
  Cenário: Perfil já segmentado fica como está até o próximo processamento
    Dado que o perfil do cliente "c-001" tem ticket médio 100, frequência de compras 1 e score de engajamento 0.9
    E que o cliente "c-001" foi segmentado
    Quando o operador "caio" com o papel "ANALISTA" desativa o segmento "engajado"
    Então o perfil do cliente "c-001" está associado ao segmento "engajado"

  @aplicacao @RN-11 @Q-33
  Cenário: Quem só atende não edita segmentos
    Quando o operador "rita" com o papel "ATENDIMENTO" muda o critério do segmento "engajado" para "scoreEngajamento >= 0.5"
    Então a operação falha com o erro "sem_permissao"
    E o segmento "engajado" continua com o critério "scoreEngajamento >= 0.7"

  @aplicacao @RN-14 @Q-42 @lgpd
  Cenário: A prévia conta só clientes ativos com consentimento válido, pelo perfil salvo
    Dado que existem os clientes com perfil salvo:
      | cliente | consentimento | ativo | ticket | frequencia | score | segmento |
      | c-001   | CONCEDIDO     | sim   | 600    | 4          | 0.2   | premium  |
      | c-002   | CONCEDIDO     | sim   | 100    | 1          | 0.6   | padrao   |
      | c-003   | CONCEDIDO     | sim   | 100    | 1          | 0.8   | engajado |
      | c-004   | REVOGADO      | sim   | 100    | 1          | 0.6   | padrao   |
      | c-005   | CONCEDIDO     | não   | 100    | 1          | 0.6   | padrao   |
    E que existe o cliente "c-006" ativo, com consentimento válido e sem perfil
    Quando o operador "caio" com o papel "ANALISTA" pede a prévia do critério "scoreEngajamento >= 0.5" para o segmento "engajado"
    Então a prévia mostra:
      | segmento | atual | previsto |
      | premium  | 1     | 1        |
      | engajado | 1     | 2        |
      | padrao   | 1     | 0        |
    E a prévia considera 4 clientes, 1 deles sem perfil
    E o segmento "engajado" continua com o critério "scoreEngajamento >= 0.7"

  @aplicacao @RN-14 @Q-42
  Cenário: A prévia de um critério inválido falha como o salvamento
    Quando o operador "caio" com o papel "ANALISTA" pede a prévia do critério "idade > 18" para o segmento "engajado"
    Então a operação falha com o erro "dados_invalidos" no campo "criterio"

  @dominio @RN-14 @Q-42
  Esquema do Cenário: Testar critério com valores digitados
    Quando o critério "<criterio>" é testado com frequência de compras <frequencia>, ticket médio <ticket> e score <score>
    Então o teste do critério dá "<resultado>"

    Exemplos:
      | criterio                                     | frequencia | ticket | score | resultado |
      | ticketMedio >= 500 && frequenciaCompras >= 3 | 3          | 500    | 0     | casa      |
      | ticketMedio >= 500 && frequenciaCompras >= 3 | 2          | 500    | 0     | não casa  |
      | scoreEngajamento > 0.7                       | 0          | 0      | 0.7   | não casa  |
      |                                              | 0          | 0      | 0     | casa      |
