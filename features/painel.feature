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
