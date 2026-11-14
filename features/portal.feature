# language: pt
@wip @Q-55 @Q-56 @Q-57 @Q-58
Funcionalidade: Portal do titular
  O próprio cliente entra no portal por um link de uso único enviado por e-mail, vê e retira a
  autorização, autoriza de novo, baixa os próprios dados e pede a eliminação, sem depender da
  operação. A resposta ao pedido de link nunca revela quem é cliente (Spec v1.10, RN-19, RN-20)

  Contexto:
    Dado que a data atual é "2026-11-14T12:00:00Z"
    E que existe o cliente "c-001" ativo com o nome "Ana Sintética"
    E que o painel está no ar com o operador "operador" e a senha "senha-de-teste"
    E que o portal do titular está no ar no endereço "https://perfila.exemplo.test"

  # ---- Pedido do link (aplicação) ----

  @aplicacao @RN-19 @lgpd
  Cenário: O titular cadastrado recebe um link de uso único que vale por 30 minutos
    Quando o titular pede acesso com o e-mail "c-001@exemplo.test"
    Então o pedido de acesso é aceito sem erro
    E a caixa de saída tem 1 mensagem para "c***@exemplo.test"
    E o link enviado começa com "https://perfila.exemplo.test/titular/acesso?t="
    E o link enviado vale até "2026-11-14T12:30:00Z"
    E o sistema guarda só o hash do token do link enviado

  @aplicacao @RN-19 @lgpd
  Esquema do Cenário: A resposta ao pedido de link é a mesma com ou sem cadastro
    Dado que os dados do cliente "c-001" <situacao>
    Quando o titular pede acesso com o e-mail "<email>"
    Então o pedido de acesso é aceito sem erro
    E a caixa de saída tem <mensagens> mensagens

    Exemplos:
      | situacao          | email                       | mensagens |
      | estão cadastrados | c-001@exemplo.test          | 1         |
      | estão cadastrados | ninguem@exemplo.test        | 0         |
      | foram eliminados  | c-001@exemplo.test          | 0         |
      | foram eliminados  | eliminado@anonimizado.invalid | 0       |

  @aplicacao @RN-19
  Cenário: O e-mail é procurado sem diferenciar maiúsculas e espaços nas pontas
    Quando o titular pede acesso com o e-mail "  C-001@Exemplo.TEST "
    Então a caixa de saída tem 1 mensagem para "c***@exemplo.test"

  @aplicacao @RN-19
  Esquema do Cenário: O pedido de link exige um e-mail
    Quando o titular pede acesso com o e-mail "<email>"
    Então a operação falha com o erro "dados_invalidos"
    E a caixa de saída tem 0 mensagens

    Exemplos:
      | email           |
      |                 |
      | c-001.exemplo   |

  @aplicacao @RN-19
  Esquema do Cenário: Cada cadastro recebe no máximo 3 links por hora, sem mudar a resposta
    Dado que o titular pediu 3 links com o e-mail "c-001@exemplo.test"
    E que se passaram <minutos> minutos
    Quando o titular pede acesso com o e-mail "c-001@exemplo.test"
    Então o pedido de acesso é aceito sem erro
    E a caixa de saída tem <mensagens> mensagens

    Exemplos:
      | minutos | mensagens |
      | 59      | 3         |
      | 60      | 3         |
      | 61      | 4         |

  @aplicacao @RN-19
  Cenário: Cada cadastro com o mesmo e-mail recebe o próprio link
    Dado que existe o cliente "c-002" ativo com o nome "Ana Sintética" e o e-mail "c-001@exemplo.test"
    Quando o titular pede acesso com o e-mail "c-001@exemplo.test"
    Então a caixa de saída tem 2 mensagens
    E cada link enviado entra num cadastro diferente

  # ---- Uso do link (aplicação) ----

  @aplicacao @RN-19
  Cenário: O link enviado identifica o titular uma única vez
    Dado que o titular pediu um link com o e-mail "c-001@exemplo.test"
    Quando o titular usa o link enviado
    Então o titular entra como o cliente "c-001"
    E usar o mesmo link de novo falha com o erro "acesso_invalido"

  @aplicacao @RN-19
  Esquema do Cenário: O link vale por 30 minutos
    Dado que o titular pediu um link com o e-mail "c-001@exemplo.test"
    E que se passaram <minutos> minutos
    Quando o titular usa o link enviado
    Então o resultado do acesso é "<resultado>"

    Exemplos:
      | minutos | resultado        |
      | 29      | entrou           |
      | 30      | acesso_invalido  |
      | 31      | acesso_invalido  |

  @aplicacao @RN-19
  Cenário: Um link adulterado não identifica ninguém
    Quando o titular usa o link com o token "0000000000000000000000000000000000000000000000000000000000000000"
    Então o resultado do acesso é "acesso_invalido"

  @aplicacao @RN-19 @RN-17
  Cenário: O link de um titular eliminado depois do pedido não vale mais
    Dado que o titular pediu um link com o e-mail "c-001@exemplo.test"
    E que os dados do cliente "c-001" foram eliminados
    Quando o titular usa o link enviado
    Então o resultado do acesso é "acesso_invalido"

  # ---- Meus dados e pedido de eliminação (aplicação) ----

  @aplicacao @UC-08 @RN-10
  Cenário: Meus dados trazem o cadastro, a autorização e o histórico, com qualquer estado
    Dado que o cliente "c-001" revogou o consentimento em "2026-11-10T10:00:00Z"
    Quando o titular do cliente "c-001" consulta os próprios dados
    Então os dados consultados trazem o nome "Ana Sintética" e o consentimento "REVOGADO" com 2 movimentos
    E os dados consultados não trazem pedido de eliminação aberto

  @aplicacao @RN-20 @lgpd
  Esquema do Cenário: O pedido de eliminação exige que a autorização não esteja concedida
    Dado que o consentimento do cliente "c-001" está no estado "<estado>"
    Quando o titular do cliente "c-001" pede a eliminação dos dados
    Então o resultado do pedido de eliminação é "<resultado>"

    Exemplos:
      | estado        | resultado           |
      | NAO_CONCEDIDO | ABERTO              |
      | REVOGADO      | ABERTO              |
      | CONCEDIDO     | transicao_invalida  |

  @aplicacao @RN-20
  Cenário: Pedir de novo a eliminação devolve o mesmo pedido aberto
    Dado que o cliente "c-001" revogou o consentimento em "2026-11-10T10:00:00Z"
    E que o titular do cliente "c-001" pediu a eliminação dos dados
    Quando o titular do cliente "c-001" pede a eliminação dos dados
    Então o cliente "c-001" tem 1 pedido de eliminação "ABERTO"

  @aplicacao @RN-20
  Esquema do Cenário: Pedido de eliminação de cliente eliminado ou inexistente
    Dado que os dados do cliente "c-001" foram eliminados
    Quando o titular do cliente "<cliente>" pede a eliminação dos dados
    Então a operação falha com o erro "<erro>"

    Exemplos:
      | cliente | erro                   |
      | c-001   | cliente_eliminado      |
      | c-999   | cliente_nao_encontrado |

  @aplicacao @RN-17 @RN-20
  Cenário: A eliminação pela operação atende o pedido e apaga os links do titular
    Dado que o cliente "c-001" revogou o consentimento em "2026-11-10T10:00:00Z"
    E que o titular do cliente "c-001" pediu a eliminação dos dados
    E que o titular pediu um link com o e-mail "c-001@exemplo.test"
    Quando os dados do cliente "c-001" são eliminados com a confirmação "Ana Sintética"
    Então o cliente "c-001" tem 1 pedido de eliminação "ATENDIDO"
    E o cliente "c-001" não tem links de acesso guardados

  @aplicacao @RN-20
  Cenário: A operação vê os pedidos abertos antes dos atendidos
    Dado que existe o cliente "c-002" ativo com o nome "Bruno Sintético"
    E que os clientes "c-001" e "c-002" pediram a eliminação dos dados
    E que os dados do cliente "c-001" foram eliminados
    Quando a operação lista os pedidos de eliminação
    Então a lista de pedidos é "c-002 ABERTO, c-001 ATENDIDO"

  # ---- Estados do pedido (domínio) ----

  @dominio @RN-20
  Cenário: Atender um pedido aberto
    Dado que existe um pedido de eliminação "ABERTO"
    Quando o pedido é atendido
    Então o pedido está no estado "ATENDIDO" com a data de atendimento "2026-11-14T12:00:00Z"

  @dominio @RN-20
  Cenário: Atender de novo um pedido já atendido é transição inválida
    Dado que existe um pedido de eliminação "ATENDIDO"
    Quando o pedido é atendido
    Então a operação falha com o erro "transicao_invalida"

  # ---- Portal (web) ----

  @web @RN-19 @lgpd
  Esquema do Cenário: Pelo portal, a resposta ao pedido de link é sempre a mesma
    Quando o titular pede o link pelo portal com o e-mail "<email>"
    Então a resposta leva a "/titular/entrar"
    E a página de destino mostra "Se o e-mail estiver cadastrado, enviamos um link de acesso."

    Exemplos:
      | email                |
      | c-001@exemplo.test   |
      | ninguem@exemplo.test |

  @web @RN-19
  Cenário: Muitos pedidos de link do mesmo endereço são recusados
    Dado que o titular pediu o link pelo portal 10 vezes
    Quando o titular pede o link pelo portal com o e-mail "c-001@exemplo.test"
    Então a resposta tem status 429
    E a página mostra "Muitas tentativas. Aguarde 15 minutos."

  @web @RN-19
  Cenário: Abrir o link não o consome
    Dado que o titular pediu um link com o e-mail "c-001@exemplo.test"
    Quando o titular abre o link enviado
    Então a resposta tem status 200
    E a página pede a confirmação "entrar-portal"
    E o link enviado ainda vale

  @web @RN-19 @lgpd
  Cenário: Pelo link, o titular vê só os próprios dados
    Dado que existe o cliente "c-002" ativo com o nome "Bruno Sintético"
    E que o titular pediu um link com o e-mail "c-001@exemplo.test"
    Quando o titular entra pelo link enviado
    Então a resposta leva a "/titular/meus-dados"
    E a página de destino mostra "Ana Sintética"
    E a página de destino não mostra "Bruno Sintético"

  @web @RN-19
  Cenário: O link usado não entra de novo
    Dado que o titular entrou no portal pelo link enviado
    Quando o titular entra pelo link enviado
    Então a resposta tem status 401
    E a página mostra "Este link não vale mais. Peça um novo."

  @web @RN-19
  Cenário: O link expirado não entra
    Dado que o titular pediu um link com o e-mail "c-001@exemplo.test"
    E que se passaram 31 minutos
    Quando o titular entra pelo link enviado
    Então a resposta tem status 401

  @web @Q-55
  Cenário: Sem sessão, Meus dados leva a Entrar
    Quando a página "/titular/meus-dados" é aberta sem sessão
    Então a resposta leva a "/titular/entrar"

  @web @Q-55
  Cenário: A sessão do portal termina depois de 30 minutos sem uso
    Dado que o titular entrou no portal pelo link enviado
    E que se passaram 31 minutos
    Quando a página "/titular/meus-dados" é aberta
    Então a resposta leva a "/titular/entrar"

  @web @Q-55 @lgpd
  Cenário: A sessão do portal não abre o painel
    Dado que o titular entrou no portal pelo link enviado
    Quando a página "/painel/clientes" é aberta
    Então a resposta leva a "/painel/entrar"

  @web @Q-55 @lgpd
  Cenário: A sessão do painel não abre o portal
    Dado que o operador entrou no painel
    Quando a página "/titular/meus-dados" é aberta
    Então a resposta leva a "/titular/entrar"

  @web @lgpd @UC-08 @RN-03 @RN-10
  Cenário: Retirar a autorização pelo portal tem efeito imediato e entra no histórico
    Dado que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    E que o titular entrou no portal pelo link enviado
    Quando o titular confirma "revogar" a autorização pelo portal
    Então a resposta leva a "/titular/meus-dados"
    E o consentimento do cliente "c-001" está no estado "REVOGADO"
    E o último movimento do cliente "c-001" é "REVOGACAO" com a origem "portal" e o operador ""
    E a página de destino mostra "por você, no portal"

  @web @UC-08 @RN-10
  Cenário: Autorizar de novo pelo portal usa o termo vigente
    Dado que o cliente "c-001" revogou o consentimento em "2026-11-10T10:00:00Z"
    E que o titular entrou no portal pelo link enviado
    Quando o titular confirma "conceder" a autorização pelo portal
    Então o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    E o último movimento do cliente "c-001" é "CONCESSAO" com a origem "portal" e o operador ""

  @web @UC-08
  Cenário: Retirar a autorização exige confirmação
    Dado que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    E que o titular entrou no portal pelo link enviado
    Quando a página "/titular/revogar" é aberta
    Então a página pede a confirmação "revogar"
    E o consentimento do cliente "c-001" está no estado "CONCEDIDO"

  @web @Q-49 @lgpd
  Cenário: Baixar meus dados entrega só o arquivo do próprio titular
    Dado que o cliente "c-001" tem dados de comportamento desde "2026-11-01T10:00:00Z"
    E que o titular entrou no portal pelo link enviado
    Quando a página "/titular/dados" é aberta
    Então a resposta é um arquivo JSON anexo com o cadastro de "Ana Sintética" e 2 eventos

  @web @RN-20 @Q-27
  Cenário: Com a autorização concedida, o portal pede para retirá-la antes de pedir a eliminação
    Dado que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    E que o titular entrou no portal pelo link enviado
    Quando a página "/titular/meus-dados" é aberta
    Então a página mostra "Para pedir a eliminação, retire a autorização antes."
    E a página de destino não oferece "pedir-eliminacao"

  @web @RN-20
  Cenário: O pedido de eliminação feito no portal aparece no painel
    Dado que o cliente "c-001" revogou o consentimento em "2026-11-10T10:00:00Z"
    E que o titular entrou no portal pelo link enviado
    E que o titular confirmou o pedido de eliminação pelo portal
    E que o operador entrou no painel
    Quando a página "/painel/pedidos" é aberta
    Então a página mostra "Ana Sintética"
    E a página mostra o selo "ABERTO"

  @web @RN-20
  Cenário: Com o pedido aberto, o portal mostra o pedido e não oferece autorizar de novo
    Dado que o cliente "c-001" revogou o consentimento em "2026-11-10T10:00:00Z"
    E que o titular entrou no portal pelo link enviado
    Quando o titular confirma o pedido de eliminação pelo portal
    Então a resposta leva a "/titular/meus-dados"
    E a página de destino mostra "Você pediu a eliminação em 14/11/2026"
    E a página de destino não oferece "conceder"

  @web @Q-57 @lgpd
  Cenário: A caixa de saída simulada mostra o destinatário mascarado
    Dado que o titular pediu um link com o e-mail "c-001@exemplo.test"
    E que o operador entrou no painel
    Quando a página "/painel/operacao/caixa-saida" é aberta
    Então a página mostra "c***@exemplo.test"
    E a página de destino não mostra "c-001@exemplo.test"

  @web @Q-55
  Cenário: O portal responde com os cabeçalhos de segurança
    Quando a página "/titular/entrar" é aberta sem sessão
    Então a resposta tem os cabeçalhos de segurança do painel

  @web @Q-55
  Cenário: Envio pelo portal sem o token CSRF é recusado
    Dado que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    E que o titular entrou no portal pelo link enviado
    Quando o titular envia "/titular/revogar" sem o token CSRF
    Então a resposta tem status 403
    E o consentimento do cliente "c-001" está no estado "CONCEDIDO"

  @web @lgpd @Q-56
  Cenário: Os logs do portal não contêm nome, e-mail nem o link
    Dado que o titular entrou no portal pelo link enviado
    Quando o titular pede o link pelo portal com o e-mail "c-001@exemplo.test"
    Então o log do portal não contém o nome "Ana Sintética", o e-mail "c-001@exemplo.test" nem o link enviado
