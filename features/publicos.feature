# language: pt
@UC-06 @RN-17 @Q-48 @Q-49
Funcionalidade: Públicos para campanhas e consentimento por finalidade
  Para anunciar a um segmento em plataformas de terceiros
  O sistema gera o público só com quem concedeu a finalidade "publicidade_terceiros",
  leva só o e-mail em hash e propaga a revogação aos públicos já entregues (Spec v1.9)

  Contexto:
    Dado que a data atual é "2026-11-13T12:00:00Z"
    E que existe o segmento ativo "vip" com prioridade 1 e critério "ticketMedio >= 500"
    E que existe o segmento ativo "fiel" com prioridade 2 e critério "frequenciaCompras >= 3"

  # ---- Consentimento por finalidade (Q-48) ----

  @aplicacao @lgpd @RN-02
  Cenário: Cada finalidade tem o seu próprio estado
    Dado que existe o cliente "c-001" ativo
    E que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    Quando o cliente "c-001" concede consentimento para a finalidade "publicidade_terceiros" com o termo "v1"
    Então o consentimento do cliente "c-001" para "publicidade_terceiros" está no estado "CONCEDIDO"
    E o consentimento do cliente "c-001" para "segmentacao" está no estado "CONCEDIDO"

  @aplicacao @lgpd @RN-03
  Cenário: Revogar a publicidade não muda a segmentação
    Dado que existe o cliente "c-001" ativo
    E que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    E que o consentimento do cliente "c-001" para "publicidade_terceiros" está no estado "CONCEDIDO"
    Quando o cliente "c-001" revoga o consentimento para "publicidade_terceiros"
    Então o consentimento do cliente "c-001" para "publicidade_terceiros" está no estado "REVOGADO"
    E o consentimento do cliente "c-001" para "segmentacao" está no estado "CONCEDIDO"

  @aplicacao @lgpd @RN-01
  Cenário: Só a publicidade não libera o uso dos dados de comportamento
    Dado que existe o cliente "c-001" ativo
    E que o consentimento do cliente "c-001" para "publicidade_terceiros" está no estado "CONCEDIDO"
    Quando o sistema coleta os eventos do cliente "c-001"
    Então a operação falha com o erro "consentimento_invalido"

  @aplicacao @RN-02
  Esquema do Cenário: Finalidade fora do enumerado é recusada
    Dado que existe o cliente "c-001" ativo
    Quando o cliente "c-001" concede consentimento para a finalidade "<finalidade>" com o termo "v1"
    Então a operação falha com o erro "dados_invalidos"
    E o consentimento do cliente "c-001" para "segmentacao" está no estado "NAO_CONCEDIDO"

    Exemplos:
      | finalidade  |
      | marketing   |
      | Segmentacao |
      |             |

  @aplicacao @RN-10
  Cenário: O movimento registra a finalidade concedida
    Dado que existe o cliente "c-001" ativo
    Quando o cliente "c-001" concede consentimento para a finalidade "publicidade_terceiros" com o termo "v1"
    Então o histórico do consentimento do cliente "c-001" tem 1 movimento
    E o último movimento do cliente "c-001" é da finalidade "publicidade_terceiros"

  # ---- Quem entra no público (RN-17) ----

  @aplicacao @lgpd
  Cenário: Só entra no público quem concedeu as duas finalidades
    Dado que o cliente "c-001" com o e-mail "c-001@exemplo.test" está no segmento "vip" com as finalidades "segmentacao,publicidade_terceiros"
    E que o cliente "c-002" com o e-mail "c-002@exemplo.test" está no segmento "vip" com as finalidades "segmentacao"
    Quando o público do segmento "vip" é gerado
    Então entram no público 1 cliente e ficam de fora 1
    E o público contém o cliente "c-001"

  @aplicacao @lgpd
  Esquema do Cenário: Inativo, revogado ou sem consentimento fica de fora
    Dado que o cliente "c-001" com o e-mail "c-001@exemplo.test" está no segmento "vip" com as finalidades "segmentacao,publicidade_terceiros"
    E que o cliente "c-002" com o e-mail "c-002@exemplo.test" está no segmento "vip" com as finalidades "segmentacao,publicidade_terceiros"
    E que <situacao> do cliente "c-002"
    Quando o público do segmento "vip" é gerado
    Então entram no público 1 cliente e ficam de fora 1
    E o público não contém o cliente "c-002"

    Exemplos:
      | situacao                                     |
      | o cadastro está inativo                      |
      | a finalidade "publicidade_terceiros" foi revogada |
      | a finalidade "segmentacao" foi revogada      |

  @aplicacao
  Cenário: O público é do segmento atual do perfil
    Dado que o cliente "c-001" com o e-mail "c-001@exemplo.test" está no segmento "fiel" com as finalidades "segmentacao,publicidade_terceiros"
    Quando o público do segmento "vip" é gerado
    Então entram no público 0 clientes e ficam de fora 0

  @aplicacao
  Cenário: Público vazio não gera arquivo
    Dado que o cliente "c-001" com o e-mail "c-001@exemplo.test" está no segmento "vip" com as finalidades "segmentacao"
    Quando o operador "analista" exporta o público do segmento "vip"
    Então a operação falha com o erro "dados_invalidos"
    E nenhuma exportação é registrada

  # ---- O arquivo (mínimo de dados) ----

  @aplicacao @lgpd
  Cenário: O arquivo leva só o e-mail em hash e o nome do segmento
    Dado que o cliente "c-001" com o e-mail "c-001@exemplo.test" está no segmento "vip" com as finalidades "segmentacao,publicidade_terceiros"
    Quando o operador "analista" exporta o público do segmento "vip"
    Então o arquivo exportado é:
      """
      email_sha256,segmento
      9c4cd21419f4d9f66ea792856912326980a93ed91c0f9b487dc9df386b13a7ef,vip
      """
    E o arquivo exportado não contém "c-001@exemplo.test" nem o nome do cliente "c-001"

  @dominio
  Esquema do Cenário: O e-mail é normalizado antes do hash
    Quando o hash do e-mail "<email>" é calculado
    Então o hash é "9c4cd21419f4d9f66ea792856912326980a93ed91c0f9b487dc9df386b13a7ef"

    Exemplos:
      | email                 |
      | c-001@exemplo.test    |
      | C-001 @Exemplo.TEST   |
      | c-0 01@exemplo.test   |

  @aplicacao @lgpd
  Cenário: A exportação fica registrada sem dado pessoal
    Dado que o cliente "c-001" com o e-mail "c-001@exemplo.test" está no segmento "vip" com as finalidades "segmentacao,publicidade_terceiros"
    Quando o operador "analista" exporta o público do segmento "vip"
    Então a exportação registra o operador "analista", o segmento "vip", o tipo "INCLUSAO", 1 cliente e o hash do arquivo
    E o log não contém o e-mail, o hash do e-mail nem o nome do cliente "c-001"

  @aplicacao
  Cenário: O arquivo pode ser baixado dentro do prazo
    Dado que o cliente "c-001" com o e-mail "c-001@exemplo.test" está no segmento "vip" com as finalidades "segmentacao,publicidade_terceiros"
    E que o operador "analista" exportou o público do segmento "vip"
    E que se passaram 23 horas
    Quando o arquivo da última exportação é baixado
    Então o arquivo baixado é igual ao exportado

  @aplicacao @lgpd @Q-51
  Cenário: Depois do prazo o arquivo não é mais entregue
    Dado que o cliente "c-001" com o e-mail "c-001@exemplo.test" está no segmento "vip" com as finalidades "segmentacao,publicidade_terceiros"
    E que o operador "analista" exportou o público do segmento "vip"
    E que se passaram 24 horas
    Quando o arquivo da última exportação é baixado
    Então a operação falha com o erro "arquivo_indisponivel"

  # ---- Remoção propagada (Q-50) ----

  @aplicacao @lgpd @RN-03 @Q-50
  Esquema do Cenário: Revogar qualquer finalidade tira o cliente do público exportado
    Dado que o cliente "c-001" com o e-mail "c-001@exemplo.test" está no segmento "vip" com as finalidades "segmentacao,publicidade_terceiros"
    E que o operador "analista" exportou o público do segmento "vip"
    Quando o cliente "c-001" revoga o consentimento para "<finalidade>"
    Então o segmento "vip" tem 1 remoção pendente
    E o membro do cliente "c-001" no segmento "vip" está no estado "REMOCAO_PENDENTE"

    Exemplos:
      | finalidade            |
      | publicidade_terceiros |
      | segmentacao           |

  @aplicacao @lgpd @Q-50
  Cenário: O arquivo de remoção leva quem revogou e encerra a pendência
    Dado que o cliente "c-001" com o e-mail "c-001@exemplo.test" está no segmento "vip" com as finalidades "segmentacao,publicidade_terceiros"
    E que o operador "analista" exportou o público do segmento "vip"
    E que a finalidade "publicidade_terceiros" foi revogada do cliente "c-001"
    Quando o operador "analista" exporta as remoções do segmento "vip"
    Então o arquivo exportado tem o hash do e-mail "c-001@exemplo.test"
    E o segmento "vip" tem 0 remoções pendentes
    E o membro do cliente "c-001" no segmento "vip" está no estado "REMOVIDO"

  @aplicacao @Q-50
  Cenário: Sem remoção pendente não há arquivo de remoção
    Dado que o cliente "c-001" com o e-mail "c-001@exemplo.test" está no segmento "vip" com as finalidades "segmentacao,publicidade_terceiros"
    E que o operador "analista" exportou o público do segmento "vip"
    Quando o operador "analista" exporta as remoções do segmento "vip"
    Então a operação falha com o erro "dados_invalidos"

  @aplicacao @lgpd @Q-50
  Cenário: Quem mudou de segmento fica pendente de remoção na exportação seguinte
    Dado que o cliente "c-001" com o e-mail "c-001@exemplo.test" está no segmento "vip" com as finalidades "segmentacao,publicidade_terceiros"
    E que o cliente "c-002" com o e-mail "c-002@exemplo.test" está no segmento "vip" com as finalidades "segmentacao,publicidade_terceiros"
    E que o operador "analista" exportou o público do segmento "vip"
    E que o perfil do cliente "c-001" passou para o segmento "fiel"
    Quando o operador "analista" exporta o público do segmento "vip"
    Então o membro do cliente "c-001" no segmento "vip" está no estado "REMOCAO_PENDENTE"

  @dominio @Q-50
  Esquema do Cenário: Transições inválidas do membro do público são rejeitadas
    Dado que existe um membro de público no estado "<estado>"
    Quando tenta-se "<acao>" o membro
    Então a operação falha com o erro "transicao_invalida"

    Exemplos:
      | estado           | acao             |
      | REMOCAO_PENDENTE | marcar remoção   |
      | EXPORTADO        | remover          |
      | REMOVIDO         | marcar remoção   |
      | REMOVIDO         | remover          |

  # ---- Relatório agregado (RN-18, Q-52) ----

  @aplicacao @lgpd @RN-18 @Q-52
  Cenário: O agregado esconde grupos com menos de 5 clientes
    Dado que 5 clientes estão no segmento "vip" com ticket médio 600 e a categoria preferida "livros"
    E que 4 clientes estão no segmento "fiel" com ticket médio 80 e a categoria preferida "casa"
    Quando o relatório agregado por segmento é gerado
    Então a linha do segmento "vip" mostra 5 clientes, ticket médio 600 e a categoria "livros"
    E a linha do segmento "fiel" mostra "menos de 5" sem ticket nem categoria

  @aplicacao @lgpd @RN-18
  Cenário: O agregado conta só quem tem a segmentação válida
    Dado que 5 clientes estão no segmento "vip" com ticket médio 600 e a categoria preferida "livros"
    E que 1 dos clientes do segmento "vip" revogou a segmentação
    Quando o relatório agregado por segmento é gerado
    Então a linha do segmento "vip" mostra "menos de 5" sem ticket nem categoria

  @aplicacao @lgpd @RN-18
  Cenário: O CSV do agregado não tem identificador
    Dado que 5 clientes estão no segmento "vip" com ticket médio 600 e a categoria preferida "livros"
    Quando o relatório agregado por segmento é exportado em CSV
    Então o CSV do agregado é:
      """
      segmento,clientes,ticket_medio,categoria_mais_comum
      vip,5,600.00,livros
      fiel,menos de 5,,
      """

  # ---- API (Spec 3.7, Q-48) ----

  @api @lgpd
  Cenário: A API concede a publicidade sem mexer na segmentação
    Dado que a API está no ar e exige o token "token-de-teste"
    E que existe o cliente "c-001" ativo
    Quando "PUT /v1/clientes/c-001/consentimento" é chamado com o token "token-de-teste" e o corpo:
      """
      {"finalidade": "publicidade_terceiros", "versaoTermo": "v1"}
      """
    Então a resposta tem status 200
    E o consentimento do cliente "c-001" para "publicidade_terceiros" está no estado "CONCEDIDO"
    E o consentimento do cliente "c-001" para "segmentacao" está no estado "NAO_CONCEDIDO"

  @api
  Cenário: A API recusa finalidade fora do enumerado
    Dado que a API está no ar e exige o token "token-de-teste"
    E que existe o cliente "c-001" ativo
    Quando "PUT /v1/clientes/c-001/consentimento" é chamado com o token "token-de-teste" e o corpo:
      """
      {"finalidade": "marketing", "versaoTermo": "v1"}
      """
    Então a resposta tem status 422
    E o corpo tem o código "dados_invalidos"

  @api @lgpd @RN-03
  Cenário: Revogação sem corpo continua revogando a segmentação
    Dado que a API está no ar e exige o token "token-de-teste"
    E que existe o cliente "c-001" ativo
    E que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    E que o consentimento do cliente "c-001" para "publicidade_terceiros" está no estado "CONCEDIDO"
    Quando "POST /v1/clientes/c-001/consentimento/revogacao" é chamado com o token "token-de-teste"
    Então o consentimento do cliente "c-001" para "segmentacao" está no estado "REVOGADO"
    E o consentimento do cliente "c-001" para "publicidade_terceiros" está no estado "CONCEDIDO"

  @api
  Cenário: A consulta pela API lista as duas finalidades
    Dado que a API está no ar e exige o token "token-de-teste"
    E que existe o cliente "c-001" ativo
    E que o consentimento do cliente "c-001" para "publicidade_terceiros" está no estado "CONCEDIDO"
    Quando "GET /v1/clientes/c-001/consentimento" é chamado com o token "token-de-teste"
    Então a resposta tem status 200
    E o corpo lista a finalidade "segmentacao" no estado "NAO_CONCEDIDO" e "publicidade_terceiros" no estado "CONCEDIDO"
