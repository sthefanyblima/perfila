# language: pt
@Q-48 @Q-49 @Q-50 @Q-51
Funcionalidade: Direitos do titular e retenção após a revogação
  O titular pede a correção, uma cópia ou a eliminação dos seus dados (art. 18 da LGPD), e o
  comportamento de quem revogou é apagado depois do prazo de retenção. A prova do consentimento
  fica sempre guardada, sem nome nem e-mail (Spec v1.9, RN-17, RN-18)

  Contexto:
    Dado que a data atual é "2026-11-13T12:00:00Z"
    E que existe o cliente "c-001" ativo com o nome "Ana Sintética"
    E que o catálogo possui produtos ativos na categoria "livros"

  # ---- Correção do cadastro (aplicação) ----

  @aplicacao @UC-09 @Q-48
  Esquema do Cenário: A correção do cadastro vale com qualquer estado do consentimento
    Dado que o consentimento do cliente "c-001" está no estado "<estado>"
    Quando o cadastro do cliente "c-001" é corrigido para o nome "Ana Souza Sintética" e o e-mail "ana.souza@exemplo.test"
    Então o cadastro do cliente "c-001" tem o nome "Ana Souza Sintética"
    E o cadastro do cliente "c-001" tem o e-mail "ana.souza@exemplo.test"
    E o consentimento do cliente "c-001" permanece no estado "<estado>"

    Exemplos:
      | estado        |
      | NAO_CONCEDIDO |
      | CONCEDIDO     |
      | REVOGADO      |

  @aplicacao @UC-09 @Q-48
  Cenário: A correção vale para cadastro inativo
    Dado que o cliente "c-001" está inativo
    Quando o cadastro do cliente "c-001" é corrigido para o nome "Ana Souza Sintética" e o e-mail "ana.souza@exemplo.test"
    Então o cadastro do cliente "c-001" tem o nome "Ana Souza Sintética"

  @aplicacao @UC-09 @Q-48
  Esquema do Cenário: A correção exige nome e e-mail
    Quando o cadastro do cliente "c-001" é corrigido para o nome "<nome>" e o e-mail "<email>"
    Então a operação falha com o erro "dados_invalidos"
    E o cadastro do cliente "c-001" tem o nome "Ana Sintética"

    Exemplos:
      | nome                | email                  |
      |                     | ana.souza@exemplo.test |
      | Ana Souza Sintética |                        |

  @aplicacao @UC-09 @Q-48
  Cenário: Corrigir o cadastro de um cliente inexistente
    Quando o cadastro do cliente "c-999" é corrigido para o nome "Ana Souza Sintética" e o e-mail "ana.souza@exemplo.test"
    Então a operação falha com o erro "cliente_nao_encontrado"

  # ---- Acesso e portabilidade (aplicação) ----

  @aplicacao @lgpd @RN-01 @Q-49
  Cenário: A exportação traz tudo o que o sistema guarda do titular
    Dado que o cliente "c-001" tem dados de comportamento desde "2026-11-01T10:00:00Z"
    Quando os dados do cliente "c-001" são exportados
    Então a exportação traz o cadastro com o nome "Ana Sintética"
    E a exportação traz o consentimento "CONCEDIDO" com 1 movimento
    E a exportação traz 2 eventos, o perfil no segmento "padrao" e 1 recomendação

  @aplicacao @lgpd @RN-01 @Q-49
  Cenário: A exportação vale depois da revogação
    Dado que o cliente "c-001" tem dados de comportamento desde "2026-11-01T10:00:00Z"
    E que o cliente "c-001" revogou o consentimento em "2026-11-10T10:00:00Z"
    Quando os dados do cliente "c-001" são exportados
    Então a exportação traz o consentimento "REVOGADO" com 2 movimentos
    E a exportação traz 2 eventos, o perfil no segmento "padrao" e 1 recomendação

  @aplicacao @lgpd @RN-01 @Q-49
  Cenário: A exportação de quem nunca concedeu traz só o cadastro e o consentimento
    Quando os dados do cliente "c-001" são exportados
    Então a exportação traz o cadastro com o nome "Ana Sintética"
    E a exportação traz o consentimento "NAO_CONCEDIDO" com 0 movimentos
    E a exportação traz 0 eventos, sem perfil e sem recomendações

  @aplicacao @Q-49
  Cenário: Exportar os dados de um cliente inexistente
    Quando os dados do cliente "c-999" são exportados
    Então a operação falha com o erro "cliente_nao_encontrado"

  # ---- Eliminação (aplicação) ----

  @aplicacao @lgpd @RN-17
  Cenário: A eliminação apaga o comportamento e preserva a prova do consentimento
    Dado que o cliente "c-001" tem dados de comportamento desde "2026-11-01T10:00:00Z"
    E que o cliente "c-001" revogou o consentimento em "2026-11-10T10:00:00Z"
    Quando os dados do cliente "c-001" são eliminados com a confirmação "Ana Sintética"
    Então o cliente "c-001" não tem eventos, perfil nem recomendações
    E o histórico do consentimento do cliente "c-001" tem 2 movimentos
    E o consentimento do cliente "c-001" está no estado "REVOGADO"

  @aplicacao @lgpd @RN-17
  Cenário: A eliminação anonimiza o cadastro
    Quando os dados do cliente "c-001" são eliminados com a confirmação "Ana Sintética"
    Então o cadastro do cliente "c-001" está anonimizado, inativo e eliminado em "2026-11-13T12:00:00Z"

  @aplicacao @lgpd @RN-17
  Cenário: O cliente eliminado não aparece na busca pelo nome antigo
    Dado que os dados do cliente "c-001" foram eliminados
    Quando os clientes são buscados por "Ana"
    Então a lista tem 0 clientes

  @aplicacao @RN-17
  Esquema do Cenário: A confirmação é o nome do cliente, sem diferenciar maiúsculas, acentos e espaços
    Dado que o cliente "c-001" tem dados de comportamento desde "2026-11-01T10:00:00Z"
    E que o cliente "c-001" revogou o consentimento em "2026-11-10T10:00:00Z"
    Quando os dados do cliente "c-001" são eliminados com a confirmação "<confirmacao>"
    Então <resultado>

    Exemplos:
      | confirmacao   | resultado                                                   |
      | Ana Sintética | o cliente "c-001" não tem eventos, perfil nem recomendações |
      | ANA SINTETICA | o cliente "c-001" não tem eventos, perfil nem recomendações |
      | Ana           | o cliente "c-001" continua com os dados de comportamento    |
      | Outra Pessoa  | o cliente "c-001" continua com os dados de comportamento    |
      |               | o cliente "c-001" continua com os dados de comportamento    |

  @aplicacao @RN-17
  Cenário: Confirmação diferente do nome é recusada
    Quando os dados do cliente "c-001" são eliminados com a confirmação "Outra Pessoa"
    Então a operação falha com o erro "dados_invalidos"
    E o cadastro do cliente "c-001" tem o nome "Ana Sintética"

  @aplicacao @lgpd @RN-17
  Cenário: Com consentimento concedido, é preciso revogar antes de eliminar
    Dado que o cliente "c-001" tem dados de comportamento desde "2026-11-01T10:00:00Z"
    Quando os dados do cliente "c-001" são eliminados com a confirmação "Ana Sintética"
    Então a operação falha com o erro "transicao_invalida"
    E o cliente "c-001" continua com os dados de comportamento
    E o cadastro do cliente "c-001" tem o nome "Ana Sintética"

  @aplicacao @RN-17
  Cenário: Eliminar os dados de um cliente inexistente
    Quando os dados do cliente "c-999" são eliminados com a confirmação "Ana Sintética"
    Então a operação falha com o erro "cliente_nao_encontrado"

  @aplicacao @lgpd @RN-17 @RN-01
  Esquema do Cenário: Depois da eliminação, nenhuma escrita vale para o cliente
    Dado que os dados do cliente "c-001" foram eliminados
    Quando a ação "<acao>" é tentada para o cliente "c-001"
    Então a operação falha com o erro "cliente_eliminado"

    Exemplos:
      | acao                     |
      | corrigir o cadastro      |
      | conceder o consentimento |
      | revogar o consentimento  |
      | registrar um evento      |
      | processar                |
      | eliminar os dados        |

  @aplicacao @lgpd @RN-17 @Q-49
  Cenário: A exportação do cliente eliminado entrega o que restou
    Dado que o cliente "c-001" tem dados de comportamento desde "2026-11-01T10:00:00Z"
    E que o cliente "c-001" revogou o consentimento em "2026-11-10T10:00:00Z"
    E que os dados do cliente "c-001" foram eliminados
    Quando os dados do cliente "c-001" são exportados
    Então a exportação traz o cadastro com o nome "Titular eliminado"
    E a exportação traz o consentimento "REVOGADO" com 2 movimentos
    E a exportação traz 0 eventos, sem perfil e sem recomendações

  # ---- Retenção após a revogação (aplicação) ----

  @aplicacao @lgpd @RN-18
  Esquema do Cenário: A retenção só atinge quem revogou há 180 dias ou mais
    Dado que o prazo de retenção é de 180 dias
    E que o cliente "c-001" tem dados de comportamento desde "2026-01-10T10:00:00Z"
    E que o cliente "c-001" revogou o consentimento em "<revogacao>"
    Quando a retenção é aplicada
    Então a retenção atinge <atingidos> clientes
    E <resultado>

    Exemplos:
      | revogacao            | atingidos | resultado                                                   |
      | 2026-05-18T12:00:00Z | 0         | o cliente "c-001" continua com os dados de comportamento    |
      | 2026-05-17T12:00:00Z | 1         | o cliente "c-001" não tem eventos, perfil nem recomendações |
      | 2026-05-16T12:00:00Z | 1         | o cliente "c-001" não tem eventos, perfil nem recomendações |

  @aplicacao @lgpd @RN-18
  Cenário: A retenção preserva o cadastro e a prova do consentimento
    Dado que o prazo de retenção é de 180 dias
    E que o cliente "c-001" tem dados de comportamento desde "2026-01-10T10:00:00Z"
    E que o cliente "c-001" revogou o consentimento em "2026-03-01T10:00:00Z"
    Quando a retenção é aplicada
    Então o cadastro do cliente "c-001" tem o nome "Ana Sintética"
    E o histórico do consentimento do cliente "c-001" tem 2 movimentos
    E o consentimento do cliente "c-001" está no estado "REVOGADO"

  @aplicacao @lgpd @RN-18
  Cenário: A retenção não atinge quem concedeu de novo
    Dado que o prazo de retenção é de 180 dias
    E que o cliente "c-001" tem dados de comportamento desde "2026-01-10T10:00:00Z"
    E que o cliente "c-001" revogou o consentimento em "2026-03-01T10:00:00Z"
    E que o cliente "c-001" tem dados de comportamento desde "2026-04-01T10:00:00Z"
    Quando a retenção é aplicada
    Então a retenção atinge 0 clientes
    E o cliente "c-001" continua com os dados de comportamento

  @aplicacao @lgpd @RN-18
  Cenário: A retenção não atinge quem nunca concedeu nem quem está concedido
    Dado que o prazo de retenção é de 180 dias
    E que existe o cliente "c-002" ativo com o nome "Bruno Sintético"
    E que o cliente "c-002" tem dados de comportamento desde "2026-01-10T10:00:00Z"
    Quando a retenção é aplicada
    Então a retenção atinge 0 clientes
    E o cliente "c-002" continua com os dados de comportamento

  @aplicacao @RN-18
  Cenário: Aplicar a retenção de novo não atinge ninguém
    Dado que o prazo de retenção é de 180 dias
    E que o cliente "c-001" tem dados de comportamento desde "2026-01-10T10:00:00Z"
    E que o cliente "c-001" revogou o consentimento em "2026-03-01T10:00:00Z"
    E que a retenção foi aplicada
    Quando a retenção é aplicada
    Então a retenção atinge 0 clientes

  # ---- Painel (web) ----

  @wip @web @UC-09 @Q-48
  Cenário: Corrigir o cadastro pela ficha, sem consentimento
    Dado que o operador entrou no painel
    Quando o operador corrige o cadastro do cliente "c-001" para o nome "Ana Souza Sintética" e o e-mail "ana.souza@exemplo.test"
    Então a resposta leva a "/painel/clientes/c-001/direitos"
    E a página de destino mostra "Ana Souza Sintética"
    E o cadastro do cliente "c-001" tem o e-mail "ana.souza@exemplo.test"

  @wip @web @UC-09 @Q-48
  Cenário: Correção sem nome mostra o erro ao lado do campo
    Dado que o operador entrou no painel
    Quando o operador corrige o cadastro do cliente "c-001" para o nome "" e o e-mail "ana.souza@exemplo.test"
    Então a resposta tem status 422
    E a página mostra "Informe o nome do cliente."

  @wip @web @lgpd @Q-49
  Cenário: Baixar os dados entrega um arquivo JSON, mesmo depois da revogação
    Dado que o operador entrou no painel
    E que o cliente "c-001" tem dados de comportamento desde "2026-11-01T10:00:00Z"
    E que o cliente "c-001" revogou o consentimento em "2026-11-10T10:00:00Z"
    Quando o operador baixa os dados do cliente "c-001"
    Então a resposta tem status 200
    E a resposta é um arquivo JSON anexo com o cadastro de "Ana Sintética" e 2 eventos

  @wip @web @RN-17
  Cenário: Eliminar exige a confirmação com o nome do cliente
    Dado que o operador entrou no painel
    Quando a página "/painel/clientes/c-001/eliminar" é aberta
    Então a resposta tem status 200
    E a página pede a confirmação "eliminar"
    E o cadastro do cliente "c-001" tem o nome "Ana Sintética"

  @wip @web @lgpd @RN-17
  Cenário: Eliminar pela ficha com o nome digitado
    Dado que o operador entrou no painel
    E que o cliente "c-001" tem dados de comportamento desde "2026-11-01T10:00:00Z"
    E que o cliente "c-001" revogou o consentimento em "2026-11-10T10:00:00Z"
    Quando o operador confirma a eliminação dos dados do cliente "c-001" digitando "ana sintetica"
    Então a resposta leva a "/painel/clientes/c-001"
    E a página de destino mostra o selo "Eliminado"
    E o cliente "c-001" não tem eventos, perfil nem recomendações

  @wip @web @RN-17
  Cenário: Nome errado na confirmação mostra o erro e não elimina nada
    Dado que o operador entrou no painel
    Quando o operador confirma a eliminação dos dados do cliente "c-001" digitando "Outra Pessoa"
    Então a resposta tem status 422
    E a página mostra "O nome digitado não confere com o cadastro."
    E o cadastro do cliente "c-001" tem o nome "Ana Sintética"

  @wip @web @lgpd @RN-17 @Q-27
  Cenário: Com consentimento concedido, a ficha pede a revogação antes de eliminar
    Dado que o operador entrou no painel
    E que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    Quando a página "/painel/clientes/c-001/direitos" é aberta
    Então a página mostra o aviso "Revogue o consentimento antes de eliminar os dados."
    E a página de destino não oferece "Eliminar dados"

  @wip @web @lgpd @RN-17
  Cenário: A ficha do cliente eliminado mostra só a prova do consentimento
    Dado que o operador entrou no painel
    E que os dados do cliente "c-001" foram eliminados
    Quando a página "/painel/clientes/c-001" é aberta
    Então a página mostra o selo "Eliminado"
    E a página de destino não oferece "Processar agora"
    E a página de destino não oferece "Conceder consentimento"

  @wip @web @lgpd @RN-18
  Cenário: Aplicar a retenção pela Operação
    Dado que o operador entrou no painel
    E que o cliente "c-001" tem dados de comportamento desde "2026-01-10T10:00:00Z"
    E que o cliente "c-001" revogou o consentimento em "2026-03-01T10:00:00Z"
    Quando o operador confirma a aplicação da retenção
    Então a resposta leva a "/painel/operacao"
    E a página de destino mostra "Retenção aplicada a 1 cliente"
    E o cliente "c-001" não tem eventos, perfil nem recomendações

  @wip @web @lgpd @Q-53
  Cenário: Cada pedido atendido fica no log, sem nome nem e-mail
    Dado que o operador entrou no painel
    Quando o operador corrige o cadastro do cliente "c-001" para o nome "Bia Sintética" e o e-mail "bia@exemplo.test"
    Então o log registra o pedido "correcao" atendido para o cliente "c-001"
    E o log do painel não contém "Ana Sintética", "Bia Sintética" nem "bia@exemplo.test"
