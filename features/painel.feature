# language: pt
@web
Funcionalidade: Painel web do operador
  A pessoa operadora cadastra clientes, registra o consentimento, lança eventos e acompanha o
  processamento pelo painel, sempre condicionado ao consentimento LGPD (Spec 3.7.1, Q-20 a Q-29)

  Contexto:
    Dado que a data atual é "2026-03-10T12:00:00Z"
    E que o painel está no ar com o operador "operador" e a senha "senha-de-teste"
    E que existe o cliente "c-001" ativo com eventos de compra na categoria "livros"
    E que o catálogo possui produtos ativos na categoria "livros"

  @Q-22
  Esquema do Cenário: Sem sessão, o painel leva a Entrar
    Quando a página "<pagina>" é aberta sem sessão
    Então a resposta leva a "/painel/entrar"

    Exemplos:
      | pagina                               |
      | /painel/clientes                     |
      | /painel/clientes/c-001               |
      | /painel/clientes/c-001/consentimento |
      | /painel/segmentos                    |
      | /painel/operacao                     |

  @Q-22
  Cenário: Entrar leva à página que foi pedida
    Quando o operador entra com a senha "senha-de-teste" vindo de "/painel/clientes/c-001"
    Então a resposta leva a "/painel/clientes/c-001"

  @Q-22
  Cenário: Senha errada mostra a mensagem genérica
    Quando o operador entra com a senha "errada"
    Então a resposta tem status 401
    E a página mostra "Usuário ou senha incorretos."

  @Q-22
  Cenário: Muitas tentativas de entrada são bloqueadas
    Dado que o operador errou a senha 5 vezes
    Quando o operador entra com a senha "senha-de-teste"
    Então a resposta tem status 429
    E a página mostra "Muitas tentativas. Aguarde 5 minutos."

  @Q-22
  Cenário: Sessão vencida volta para Entrar com aviso
    Dado que o operador entrou no painel
    E que se passaram 9 horas
    Quando a página "/painel/clientes" é aberta
    Então a resposta leva a "/painel/entrar"
    E a página de destino mostra "Sua sessão expirou. Entre de novo para continuar."

  @Q-22
  Cenário: Envio sem o token CSRF é recusado
    Dado que o operador entrou no painel
    E que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    Quando o operador envia "POST /painel/clientes/c-001/consentimento/revogar" sem o token CSRF
    Então a resposta tem status 403
    E o consentimento do cliente "c-001" permanece no estado "CONCEDIDO"

  @Q-22
  Cenário: As páginas do painel têm os cabeçalhos de segurança
    Quando a página "/painel/entrar" é aberta sem sessão
    Então a resposta tem status 200
    E a resposta tem os cabeçalhos de segurança do painel

  @UC-09
  Cenário: Cadastrar um cliente abre a ficha sem consentimento
    Dado que o operador entrou no painel
    Quando o operador cadastra o cliente "Cliente Sintético" com o e-mail "c-020@exemplo.test"
    Então a resposta leva à ficha do cliente cadastrado
    E a página de destino mostra o selo "NAO_CONCEDIDO"

  @UC-09
  Cenário: Cadastro sem nome mostra o erro ao lado do campo
    Dado que o operador entrou no painel
    Quando o operador cadastra o cliente "" com o e-mail "c-021@exemplo.test"
    Então a resposta tem status 422
    E a página mostra "Informe o nome do cliente."

  @Q-23
  Cenário: A busca encontra o cliente pelo nome sem diferenciar acentos
    Dado que o operador entrou no painel
    E que existe o cliente "c-002" ativo com o nome "Joana Sintética"
    Quando a página "/painel/clientes?q=sintetica" é aberta
    Então a resposta tem status 200
    E a lista mostra 1 cliente

  @lgpd @RN-01 @Q-21
  Esquema do Cenário: Sem consentimento válido, a ficha não mostra dados de comportamento
    Dado que o operador entrou no painel
    E que o consentimento do cliente "c-001" está no estado "<estado>"
    Quando a página "/painel/clientes/c-001" é aberta
    Então a resposta tem status 200
    E a página mostra o selo "<estado>"
    E a página mostra o aviso "bloqueio-consentimento"
    E a página não mostra os painéis "perfil", "segmento", "recomendacao" e "eventos"

    Exemplos:
      | estado        |
      | NAO_CONCEDIDO |
      | REVOGADO      |

  @UC-08
  Cenário: Conceder pela interface libera os dados de comportamento
    Dado que o operador entrou no painel
    E que o consentimento do cliente "c-001" está no estado "NAO_CONCEDIDO"
    Quando o operador confirma "conceder" o consentimento do cliente "c-001"
    Então o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    E a página de destino mostra o selo "CONCEDIDO"
    E a página de destino oferece "processar"

  @UC-08 @RN-03 @lgpd
  Cenário: Revogar pela interface muda o selo e bloqueia o processamento
    Dado que o operador entrou no painel
    E que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    Quando o operador confirma "revogar" o consentimento do cliente "c-001"
    Então o consentimento do cliente "c-001" está no estado "REVOGADO"
    E a página de destino mostra o selo "REVOGADO"
    E a página de destino não oferece "processar"

  @UC-08
  Cenário: Revogar exige confirmação
    Dado que o operador entrou no painel
    E que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    Quando a página "/painel/clientes/c-001/consentimento/revogar" é aberta
    Então a resposta tem status 200
    E a página pede a confirmação "revogar"
    E o consentimento do cliente "c-001" permanece no estado "CONCEDIDO"

  @RN-01 @Q-31
  Cenário: Cadastro inativo mostra o histórico só para leitura
    Dado que o operador entrou no painel
    E que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    E que o cliente "c-001" está inativo
    Quando a página "/painel/clientes/c-001" é aberta
    Então a resposta tem status 200
    E a página mostra o aviso "cadastro-inativo"
    E a página de destino não oferece "processar"

  @Q-32
  Esquema do Cenário: Revogado e não concedido têm avisos diferentes
    Dado que o operador entrou no painel
    E que o consentimento do cliente "c-001" está no estado "<estado>"
    Quando a página "/painel/clientes/c-001" é aberta
    Então a página mostra "<texto>"

    Exemplos:
      | estado        | texto                                        |
      | NAO_CONCEDIDO | ainda não autorizou o uso dos dados          |
      | REVOGADO      | revogou o consentimento                      |

  @UC-01 @UC-02 @UC-03 @UC-04 @RN-09
  Cenário: Processar agora conclui as quatro etapas
    Dado que o operador entrou no painel
    E que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    Quando o operador processa o cliente "c-001" pelo painel
    Então a página mostra as etapas "concluida", "concluida", "concluida" e "concluida"

  @RN-07 @RN-09
  Cenário: Sem produto elegível, a etapa de recomendação fica em atenção
    Dado que o operador entrou no painel
    E que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    E que todos os produtos da categoria "livros" estão inativos
    Quando o operador processa o cliente "c-001" pelo painel
    Então a página mostra as etapas "concluida", "concluida", "concluida" e "atencao"

  @RN-06 @RN-09 @Q-17
  Cenário: Processamento com falha de configuração mostra a etapa que falhou
    Dado que o operador entrou no painel
    E que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    E que existem os segmentos ativos:
      | prioridade | nome     | criterio         |
      | 1          | quebrado | ticketMedio >>> 1 |
    Quando o operador processa o cliente "c-001" pelo painel
    Então a página mostra as etapas "concluida", "concluida", "falhou" e "pendente"
    E a página mostra o aviso "falha-processamento"

  @UC-01 @RN-04
  Cenário: Registrar um evento pela interface
    Dado que o operador entrou no painel
    E que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    Quando o operador registra para o cliente "c-001" uma "compra" de valor "120,50" na categoria "livros"
    Então a resposta leva a "/painel/clientes/c-001/eventos"
    E a coleta do cliente "c-001" retorna 4 eventos

  @UC-01 @lgpd
  Cenário: Sem consentimento, registrar evento pela interface é recusado
    Dado que o operador entrou no painel
    E que o consentimento do cliente "c-001" está no estado "REVOGADO"
    Quando o operador registra para o cliente "c-001" uma "compra" de valor "120,50" na categoria "livros"
    Então a resposta tem status 403
    E a página mostra o aviso "bloqueio-consentimento"

  @RN-08
  Cenário: Expirar recomendações vencidas pelo painel
    Dado que o operador entrou no painel
    E que existe uma recomendação ativada em "2026-01-01T12:00:00Z" com validade de 30 dias
    Quando o operador confirma a expiração das recomendações vencidas
    Então a resposta leva a "/painel/operacao"
    E a página de destino mostra "1 recomendação expirada"

  @C-09 @Q-24
  Cenário: Os logs do painel não contêm nome, e-mail nem o termo de busca
    Dado que o operador entrou no painel
    E que o operador cadastrou o cliente "Cliente Sintético" com o e-mail "c-020@exemplo.test"
    Quando a página "/painel/clientes?q=Sintético" é aberta
    Então o log do painel não contém "Cliente Sintético", "c-020@exemplo.test" nem "Sintético"

  # ---- Histórico do consentimento (v1.4, Q-34, RN-10) ----

  @UC-08 @RN-10 @Q-34
  Cenário: A linha do tempo mostra cada movimento com o termo e a origem
    Dado que o consentimento do cliente "c-001" teve os movimentos:
      | data                 | acao     | termo | origem | operador |
      | 2026-03-11T10:00:00Z | conceder | v1    | api    |          |
      | 2026-04-01T09:00:00Z | revogar  |       | api    |          |
      | 2026-05-01T08:00:00Z | conceder | v2    | painel | op-01    |
    E que o operador entrou no painel
    Quando a página "/painel/clientes/c-001/consentimento" é aberta
    Então a linha do tempo do consentimento mostra, do mais recente para o mais antigo:
      | titulo                            | detalhe                        |
      | Consentimento concedido novamente | termo v2 · pelo painel, por op-01 |
      | Consentimento revogado            | termo v1 · pela API            |
      | Consentimento concedido           | termo v1 · pela API            |
      | Cadastro                          |                                |
    E a página não mostra o aviso "linha-pelas-datas"

  @UC-08 @RN-10 @Q-34
  Esquema do Cenário: Conceder e revogar pelo painel registram o operador da sessão
    Dado que o operador entrou no painel
    E que o consentimento do cliente "c-001" está no estado "<estado>"
    Quando o operador confirma "<acao>" o consentimento do cliente "c-001"
    Então o último movimento do cliente "c-001" é "<tipo>" com a origem "painel" e o operador "operador"

    Exemplos:
      | estado        | acao     | tipo      |
      | NAO_CONCEDIDO | conceder | CONCESSAO |
      | CONCEDIDO     | revogar  | REVOGACAO |

  @UC-08 @RN-10 @Q-34
  Cenário: Cliente sem movimentos continua com a linha montada pelas datas
    Dado que o operador entrou no painel
    E que o consentimento do cliente "c-001" está no estado "REVOGADO"
    Quando a página "/painel/clientes/c-001/consentimento" é aberta
    Então a página mostra o aviso "linha-pelas-datas"
    E a página mostra "Consentimento revogado"

  # ---- Usuários, perfis de acesso e auditoria (v1.5, Q-35 a Q-41, RN-11 a RN-13) ----

  @RN-11 @Q-37
  Esquema do Cenário: Cada papel só abre o que pode
    Dado que o operador "rita" com o papel "<papel>" entrou no painel
    Quando a página "<pagina>" é aberta
    Então a resposta tem status <status>

    Exemplos:
      | papel         | pagina                 | status |
      | ATENDIMENTO   | /painel/clientes       | 200    |
      | ATENDIMENTO   | /painel/clientes/c-001 | 200    |
      | ATENDIMENTO   | /painel/segmentos      | 200    |
      | ATENDIMENTO   | /painel/operacao       | 403    |
      | ATENDIMENTO   | /painel/usuarios       | 403    |
      | ATENDIMENTO   | /painel/auditoria      | 403    |
      | ANALISTA      | /painel/segmentos      | 200    |
      | ANALISTA      | /painel/catalogo       | 200    |
      | ANALISTA      | /painel/clientes       | 403    |
      | ANALISTA      | /painel/clientes/c-001 | 403    |
      | ANALISTA      | /painel/minha-senha    | 200    |
      | ADMINISTRADOR | /painel/usuarios       | 200    |
      | ADMINISTRADOR | /painel/auditoria      | 200    |
      | ADMINISTRADOR | /painel/operacao       | 200    |

  @RN-11 @Q-37
  Esquema do Cenário: O menu mostra só as telas do papel
    Dado que o operador "rita" com o papel "<papel>" entrou no painel
    Quando a página "/painel/segmentos" é aberta
    Então o menu mostra só os itens "<itens>"

    Exemplos:
      | papel         | itens                                                    |
      | ANALISTA      | segmentos, catalogo                                      |
      | ATENDIMENTO   | clientes, segmentos, catalogo                            |
      | ADMINISTRADOR | clientes, segmentos, catalogo, operacao, usuarios, auditoria |

  @RN-11 @Q-37 @lgpd
  Cenário: Sem permissão, a ação é recusada e nada é processado
    Dado que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    E que o operador "caio" com o papel "ANALISTA" entrou no painel
    Quando o operador envia "POST /painel/clientes/c-001/processamento"
    Então a resposta tem status 403
    E a página mostra "Sem permissão"
    E o cliente "c-001" não tem perfil
    E o último registro da auditoria é "acesso_negado" do operador "caio" sobre o cliente "c-001" com o resultado "NEGADO"

  @RN-12 @Q-38
  Cenário: Operador desativado perde o acesso na hora
    Dado que o operador "rita" com o papel "ATENDIMENTO" entrou no painel
    E que o operador "rita" foi desativado
    Quando a página "/painel/clientes" é aberta
    Então a resposta leva a "/painel/entrar"
    E a página de destino mostra "Sua sessão foi encerrada. Entre de novo."

  @RN-12 @Q-38
  Cenário: A mudança de papel vale na próxima página
    Dado que o operador "rita" com o papel "ATENDIMENTO" entrou no painel
    E que o papel do operador "rita" mudou para "ANALISTA"
    Quando a página "/painel/clientes" é aberta
    Então a resposta tem status 403

  @RN-12 @Q-40
  Cenário: Com senha provisória, o operador só chega a Minha senha
    Dado que existe o operador "rita" com o papel "ATENDIMENTO" e a senha "senha-provisoria-1"
    E que o operador "rita" precisa trocar a senha
    Quando o operador "rita" entra com a senha "senha-provisoria-1"
    Então a resposta leva a "/painel/minha-senha"

  @RN-12 @Q-35
  Cenário: O administrador cria um operador pela tela Usuários
    Dado que o operador entrou no painel
    Quando o operador cria pelo painel o operador "rita" com o papel "ATENDIMENTO" e a senha "senha-provisoria-1"
    Então a resposta leva à página do operador "rita"
    E o operador "rita" está ativo com o papel "ATENDIMENTO"
    E o último registro da auditoria é "criar_operador" do operador "operador" sobre o operador "rita" com o resultado "SUCESSO"

  @RN-13 @Q-39
  Esquema do Cenário: Ações e leituras de ficha ficam na auditoria
    Dado que o operador entrou no painel
    E que o consentimento do cliente "c-001" está no estado "<estado>"
    Quando <acao>
    Então a auditoria tem um registro "<registro>" do operador "operador" sobre o cliente "c-001" com o resultado "SUCESSO"

    Exemplos:
      | estado        | acao                                                         | registro              |
      | CONCEDIDO     | a página "/painel/clientes/c-001" é aberta                    | ler_ficha             |
      | CONCEDIDO     | a página "/painel/clientes/c-001/eventos" é aberta            | ler_ficha             |
      | NAO_CONCEDIDO | o operador confirma "conceder" o consentimento do cliente "c-001" | conceder_consentimento |
      | CONCEDIDO     | o operador processa o cliente "c-001" pelo painel             | processar             |

  @RN-13 @Q-39 @lgpd @C-09
  Cenário: A auditoria não guarda nome nem e-mail
    Dado que o operador entrou no painel
    E que o operador cadastrou o cliente "Cliente Sintético" com o e-mail "c-020@exemplo.test"
    Quando a página da ficha do cliente cadastrado é aberta
    Então a auditoria tem registros das ações "cadastrar_cliente" e "ler_ficha"
    E nenhum registro da auditoria contém "Cliente Sintético" nem "c-020@exemplo.test"

  @RN-13 @Q-39
  Cenário: A tela Auditoria filtra por ação
    Dado que o operador entrou no painel
    E que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    E que a página "/painel/clientes/c-001" foi aberta
    Quando a página "/painel/auditoria?acao=ler_ficha" é aberta
    Então a resposta tem status 200
    E a auditoria na tela mostra 1 registro

  # ---- Edição de segmentos e catálogo (v1.6, Q-33, Q-42 a Q-44, RN-14) ----

  @RN-11 @Q-33 @wip
  Esquema do Cenário: Só quem edita a configuração abre os formulários
    Dado que o operador "rita" com o papel "<papel>" entrou no painel
    Quando a página "<pagina>" é aberta
    Então a resposta tem status <status>

    Exemplos:
      | papel         | pagina                   | status |
      | ANALISTA      | /painel/segmentos/novo   | 200    |
      | ADMINISTRADOR | /painel/segmentos/novo   | 200    |
      | ATENDIMENTO   | /painel/segmentos/novo   | 403    |
      | ANALISTA      | /painel/catalogo/novo    | 200    |
      | ATENDIMENTO   | /painel/catalogo/novo    | 403    |

  @RN-11 @Q-33 @wip
  Esquema do Cenário: A lista de segmentos só oferece edição a quem pode editar
    Dado que o operador "rita" com o papel "<papel>" entrou no painel
    Quando a página "/painel/segmentos" é aberta
    Então a página de destino <oferece> "novo-segmento"

    Exemplos:
      | papel       | oferece    |
      | ANALISTA    | oferece    |
      | ATENDIMENTO | não oferece |

  @RN-14 @Q-33 @wip
  Cenário: O formulário mostra o critério gravado e uma linha por condição
    Dado que o operador "caio" com o papel "ANALISTA" entrou no painel
    E que existe o segmento ativo "premium" com prioridade 1 e critério "ticketMedio >= 500 && frequenciaCompras >= 3"
    Quando o formulário do segmento "premium" é aberto
    Então a página mostra "ticketMedio &gt;= 500 &amp;&amp; frequenciaCompras &gt;= 3"
    E o formulário tem 2 condições preenchidas

  @RN-14 @Q-42 @lgpd @wip
  Cenário: Salvar passa pela prévia, que mostra só contagens
    Dado que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    E que o cliente "c-001" foi processado
    E que o operador "caio" com o papel "ANALISTA" entrou no painel
    E que existe o segmento ativo "engajado" com prioridade 2 e critério "scoreEngajamento >= 0.7"
    Quando o operador envia o formulário do segmento "engajado" com o critério "scoreEngajamento >= 0.5"
    Então a página mostra o aviso "previa-segmentacao"
    E a página de destino oferece "salvar-segmento"
    E a página não mostra o nome nem o e-mail do cliente "c-001"
    E o segmento "engajado" continua com o critério "scoreEngajamento >= 0.7"
    E o último registro da auditoria é "prever_segmentacao" do operador "caio" sobre o segmento "engajado" com o resultado "SUCESSO"

  @RN-14 @RN-13 @Q-44 @wip
  Cenário: Confirmar a prévia salva o segmento e registra os campos alterados
    Dado que o operador "caio" com o papel "ANALISTA" entrou no painel
    E que existe o segmento ativo "engajado" com prioridade 2 e critério "scoreEngajamento >= 0.7"
    Quando o operador confirma o segmento "engajado" com o critério "scoreEngajamento >= 0.5"
    Então a resposta leva a "/painel/segmentos"
    E o segmento "engajado" continua com o critério "scoreEngajamento >= 0.5"
    E o último registro da auditoria é "editar_segmento" do operador "caio" sobre o segmento "engajado" com o resultado "SUCESSO"
    E o último registro da auditoria tem os campos "criterio"

  @RN-14 @Q-33 @wip
  Cenário: Critério inválido volta ao formulário com a condição errada
    Dado que o operador "caio" com o papel "ANALISTA" entrou no painel
    E que existe o segmento ativo "engajado" com prioridade 2 e critério "scoreEngajamento >= 0.7"
    Quando o operador envia o formulário do segmento "engajado" com o critério "idade > 18"
    Então a resposta tem status 422
    E a página mostra "idade &gt; 18"
    E o segmento "engajado" continua com o critério "scoreEngajamento >= 0.7"

  @RN-14 @Q-33 @wip
  Cenário: O segmento padrao recusa a desativação
    Dado que o operador "caio" com o papel "ANALISTA" entrou no painel
    Quando o operador confirma a desativação do segmento "padrao"
    Então a resposta tem status 422
    E os segmentos ativos são "padrao"

  @RN-14 @Q-33 @wip
  Cenário: Desativar um segmento pede confirmação com a prévia
    Dado que o operador "caio" com o papel "ANALISTA" entrou no painel
    E que existe o segmento ativo "engajado" com prioridade 2 e critério "scoreEngajamento >= 0.7"
    Quando o operador pede para desativar o segmento "engajado"
    Então a página mostra o aviso "previa-segmentacao"
    E a página de destino oferece "desativar-segmento"
    E os segmentos ativos são "engajado" e "padrao"

  @RN-14 @Q-42 @wip
  Cenário: Testar critério pelo painel
    Dado que o operador "caio" com o papel "ANALISTA" entrou no painel
    Quando o operador testa o critério "ticketMedio >= 500" com frequência de compras 1, ticket médio 600 e score 0
    Então a página mostra "O critério casa com estes valores."

  @RN-14 @RN-13 @Q-43 @Q-44 @wip
  Cenário: Criar um produto pelo catálogo
    Dado que o operador "caio" com o papel "ANALISTA" entrou no painel
    Quando o operador cria pelo painel o produto "Atlas" na categoria "Livros" com preço "90,50"
    Então a resposta leva a "/painel/catalogo"
    E o produto "Atlas" está gravado na categoria "livros" com o preço 90.5
    E o último registro da auditoria é "criar_produto" do operador "caio" sobre o produto "Atlas" com o resultado "SUCESSO"

  @RN-14 @RN-13 @Q-43 @Q-44 @wip
  Cenário: Desativar um produto pede confirmação e fica na auditoria
    Dado que o operador "caio" com o papel "ANALISTA" entrou no painel
    Quando o operador confirma a desativação do produto "Item 1"
    Então a resposta leva a "/painel/catalogo"
    E o produto "Item 1" está inativo
    E o último registro da auditoria é "desativar_produto" do operador "caio" sobre o produto "Item 1" com o resultado "SUCESSO"
    E o último registro da auditoria tem os campos "ativo"
