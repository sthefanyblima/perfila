# language: pt
@aplicacao
Funcionalidade: Operadores, papéis e auditoria
  Cada pessoa que opera o painel tem um papel que define o que ela alcança; toda ação fica
  numa trilha de auditoria sem dado pessoal (Spec 3.3 RN-11 a RN-13, Q-35 a Q-41)

  Contexto:
    Dado que a data atual é "2026-03-10T12:00:00Z"

  @RN-12 @Q-35
  Cenário: O administrador cria um operador com senha provisória
    Dado que existe o operador "admin" com o papel "ADMINISTRADOR" e a senha "senha-admin-123"
    Quando o operador "admin" cria o operador "rita" com o papel "ATENDIMENTO" e a senha "senha-provisoria-1"
    Então o operador "rita" está ativo com o papel "ATENDIMENTO"
    E o operador "rita" precisa trocar a senha

  @RN-12 @Q-35
  Cenário: Login já usado é recusado
    Dado que existe o operador "admin" com o papel "ADMINISTRADOR" e a senha "senha-admin-123"
    Quando o operador "admin" cria o operador "ADMIN" com o papel "ANALISTA" e a senha "senha-provisoria-1"
    Então a operação falha com o erro "login_em_uso"

  @RN-12 @Q-40
  Esquema do Cenário: A senha precisa ter pelo menos 12 caracteres
    Dado que existe o operador "admin" com o papel "ADMINISTRADOR" e a senha "senha-admin-123"
    Quando o operador "admin" cria o operador "rita" com o papel "ATENDIMENTO" e a senha "<senha>"
    Então o resultado da criação é "<resultado>"

    Exemplos:
      | senha         | resultado       |
      | 12345678901   | dados_invalidos |
      | 123456789012  | ok              |
      | 1234567890123 | ok              |

  @RN-11 @Q-37
  Esquema do Cenário: Só o administrador gerencia operadores
    Dado que existe o operador "admin" com o papel "ADMINISTRADOR" e a senha "senha-admin-123"
    E que existe o operador "outro" com o papel "<papel>" e a senha "senha-outro-123"
    Quando o operador "outro" cria o operador "rita" com o papel "ATENDIMENTO" e a senha "senha-provisoria-1"
    Então a operação falha com o erro "sem_permissao"

    Exemplos:
      | papel       |
      | ATENDIMENTO |
      | ANALISTA    |

  @RN-12 @Q-35
  Esquema do Cenário: O administrador não muda a própria conta
    Dado que existe o operador "admin" com o papel "ADMINISTRADOR" e a senha "senha-admin-123"
    Quando o operador "admin" <acao>
    Então a operação falha com o erro "propria_conta"
    E o operador "admin" está ativo com o papel "ADMINISTRADOR"

    Exemplos:
      | acao                                 |
      | desativa a própria conta             |
      | muda o próprio papel para "ANALISTA" |

  @RN-12 @Q-36
  Esquema do Cenário: Entrar com credenciais inválidas dá sempre o mesmo erro
    Dado que existe o operador "admin" com o papel "ADMINISTRADOR" e a senha "senha-admin-123"
    E que existe o operador "rita" com o papel "ATENDIMENTO" e a senha "senha-rita-1234"
    E que o operador "rita" <situacao>
    Quando alguém entra com o login "<login>" e a senha "<senha>"
    Então a operação falha com o erro "nao_autenticado"

    Exemplos:
      | situacao        | login    | senha            |
      | está ativo      | rita     | senha-errada-123 |
      | está ativo      | ninguem  | senha-rita-1234  |
      | foi desativado  | rita     | senha-rita-1234  |

  @RN-12 @Q-36
  Cenário: Entrar com login e senha corretos devolve o operador com o papel
    Dado que existe o operador "admin" com o papel "ADMINISTRADOR" e a senha "senha-admin-123"
    E que existe o operador "rita" com o papel "ATENDIMENTO" e a senha "senha-rita-1234"
    Quando alguém entra com o login "RITA" e a senha "senha-rita-1234"
    Então a entrada devolve o operador "rita" com o papel "ATENDIMENTO"

  @RN-12 @Q-38
  Esquema do Cenário: Desativar ou redefinir a senha encerra a sessão aberta
    Dado que existe o operador "admin" com o papel "ADMINISTRADOR" e a senha "senha-admin-123"
    E que existe o operador "rita" com o papel "ATENDIMENTO" e a senha "senha-rita-1234"
    E que o operador "rita" tem uma sessão aberta
    Quando o operador "admin" <acao> "rita"
    Então a sessão aberta de "rita" é recusada

    Exemplos:
      | acao                                  |
      | desativa                              |
      | redefine a senha provisória de        |

  @RN-12 @Q-38
  Cenário: Mudar o papel não encerra a sessão, mas vale na hora
    Dado que existe o operador "admin" com o papel "ADMINISTRADOR" e a senha "senha-admin-123"
    E que existe o operador "rita" com o papel "ATENDIMENTO" e a senha "senha-rita-1234"
    E que o operador "rita" tem uma sessão aberta
    Quando o operador "admin" muda o papel de "rita" para "ANALISTA"
    Então a sessão aberta de "rita" devolve o papel "ANALISTA"

  @RN-12 @Q-40
  Cenário: Trocar a própria senha exige a senha atual
    Dado que existe o operador "admin" com o papel "ADMINISTRADOR" e a senha "senha-admin-123"
    E que existe o operador "rita" com o papel "ATENDIMENTO" e a senha "senha-rita-1234"
    Quando o operador "rita" troca a própria senha informando a atual "senha-errada-123"
    Então a operação falha com o erro "nao_autenticado"

  @RN-12 @Q-40
  Cenário: Trocar a própria senha tira a troca obrigatória
    Dado que existe o operador "admin" com o papel "ADMINISTRADOR" e a senha "senha-admin-123"
    E que existe o operador "rita" com o papel "ATENDIMENTO" e a senha "senha-rita-1234"
    E que o operador "rita" precisa trocar a senha
    Quando o operador "rita" troca a própria senha informando a atual "senha-rita-1234"
    Então o operador "rita" não precisa trocar a senha

  @RN-12 @Q-35
  Esquema do Cenário: O administrador inicial só é criado sem nenhum operador
    Dado que <situacao>
    Quando o servidor cria o administrador inicial "inicial"
    Então <resultado>

    Exemplos:
      | situacao                       | resultado                                                   |
      | não existe nenhum operador     | o operador "inicial" está ativo com o papel "ADMINISTRADOR" |
      | já existem operadores gravados | o operador "inicial" não existe                             |

  @RN-13 @Q-39 @lgpd
  Cenário: O registro de auditoria guarda só login, ação, alvo e resultado
    Dado que existe o operador "admin" com o papel "ADMINISTRADOR" e a senha "senha-admin-123"
    E que existe o cliente "c-001" ativo
    Quando o painel registra a ação "ler_ficha" do operador "admin" sobre o cliente "c-001" com o resultado "SUCESSO"
    Então a auditoria tem 1 registro
    E o último registro da auditoria é "ler_ficha" do operador "admin" sobre o cliente "c-001" com o resultado "SUCESSO"

  @RN-13 @Q-39
  Esquema do Cenário: Só o administrador lê a auditoria
    Dado que existe o operador "admin" com o papel "ADMINISTRADOR" e a senha "senha-admin-123"
    E que existe o operador "outro" com o papel "<papel>" e a senha "senha-outro-123"
    Quando o operador "outro" lista a auditoria
    Então a operação falha com o erro "sem_permissao"

    Exemplos:
      | papel       |
      | ATENDIMENTO |
      | ANALISTA    |

  @RN-13 @Q-39
  Esquema do Cenário: O expurgo apaga só registros com mais de 180 dias
    Dado que existe o operador "admin" com o papel "ADMINISTRADOR" e a senha "senha-admin-123"
    E que existe um registro de auditoria de <dias> dias atrás
    Quando a auditoria é expurgada
    Então a auditoria tem <restantes> registro

    Exemplos:
      | dias | restantes |
      | 179  | 1         |
      | 180  | 1         |
      | 181  | 0         |
