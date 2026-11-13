# language: pt
@UC-01 @RN-16 @Q-47
Funcionalidade: Importação de eventos em lote
  Lojas mandam eventos em lote. Cada linha passa pelas regras de UC-01 (RN-01 e RN-04):
  as aceitas entram, as outras ficam no relatório com o motivo, sem parar o lote.
  Duplicados são ignorados, para o reenvio ser seguro. A prévia não grava nada.

  Contexto:
    Dado que a data atual é "2026-11-13T12:00:00Z"
    E que existe o cliente "c-001" ativo
    E que o consentimento do cliente "c-001" está no estado "CONCEDIDO"

  # ---- Caso de uso (aplicação) ----

  @aplicacao
  Cenário: Linhas válidas entram e o relatório conta cada uma
    Dado que o arquivo de eventos é:
      """
      cliente_id,tipo,origem,timestamp,valor,categoria
      c-001,compra,loja on-line,2026-11-10T14:30:00-03:00,150.90,livros
      c-001,navegacao,app,2026-11-11T09:00:00Z,0,
      c-001,interacao,app,2026-11-12T10:00:00Z,,
      """
    Quando o arquivo é importado
    Então o resultado tem 3 aceitas, 0 duplicadas e 0 rejeitadas
    E a coleta do cliente "c-001" retorna 3 eventos

  @aplicacao
  Cenário: A prévia confere tudo e não grava nada
    Dado que o arquivo de eventos é:
      """
      cliente_id,tipo,origem,timestamp,valor
      c-001,compra,loja on-line,2026-11-10T14:30:00Z,150.90
      c-001,compra,loja on-line,2026-11-10T15:30:00Z,0
      """
    Quando a prévia do arquivo é feita
    Então o resultado tem 1 aceita, 0 duplicadas e 1 rejeitada
    E nenhum evento é gravado

  @aplicacao
  Cenário: As colunas valem pelo nome do cabeçalho, em qualquer ordem
    Dado que o arquivo de eventos é:
      """
      valor,timestamp,tipo,cliente_id,origem
      89.90,2026-11-10T14:30:00Z,compra,c-001,loja on-line
      """
    Quando o arquivo é importado
    Então o resultado tem 1 aceita, 0 duplicadas e 0 rejeitadas

  @aplicacao @RN-04
  Esquema do Cenário: Linha inválida aparece com o motivo e as outras entram
    Dado que o arquivo de eventos é:
      """
      cliente_id,tipo,origem,timestamp,valor,categoria
      c-001,compra,loja on-line,2026-11-10T14:30:00Z,150.90,livros
      <linha>
      """
    Quando o arquivo é importado
    Então o resultado tem 1 aceita, 0 duplicadas e 1 rejeitada
    E a linha 3 é rejeitada com o erro "dados_invalidos" e o motivo "<motivo>"

    Exemplos:
      | linha                                                          | motivo           |
      | 12345,compra,loja on-line,2026-11-10T14:30:00Z,10,livros       | cliente_id       |
      | c-001,devolucao,loja on-line,2026-11-10T14:30:00Z,10,livros    | tipo             |
      | c-001,compra,,2026-11-10T14:30:00Z,10,livros                   | origem           |
      | c-001,compra,loja on-line,10/11/2026 14:30,10,livros           | timestamp        |
      | c-001,compra,loja on-line,2026-11-10T14:30:00,10,livros        | timestamp        |
      | c-001,navegacao,app,2026-11-13T12:02:01Z,0,                    | timestamp_futuro |
      | c-001,compra,loja on-line,2026-11-10T14:30:00Z,"150,90",livros | valor            |
      | c-001,compra,loja on-line,2026-11-10T14:30:00Z,1000000.01,     | valor            |
      | c-001,navegacao,app,2026-11-10T14:30:00Z,-1,                   | valor            |
      | c-001,compra,loja on-line,2026-11-10T14:30:00Z,0,livros        | valor_compra     |
      | c-001,compra,loja on-line,2026-11-10T14:30:00Z                 | colunas          |

  @aplicacao @RN-04
  Cenário: Os limites de valor e de relógio de RN-04 valem em cada linha
    Dado que o arquivo de eventos é:
      """
      cliente_id,tipo,origem,timestamp,valor
      c-001,compra,loja on-line,2026-11-10T14:30:00Z,1000000
      c-001,compra,loja on-line,2026-11-10T14:31:00Z,0.01
      c-001,navegacao,app,2026-11-13T12:02:00Z,0
      """
    Quando o arquivo é importado
    Então o resultado tem 3 aceitas, 0 duplicadas e 0 rejeitadas

  @aplicacao @lgpd @RN-01
  Esquema do Cenário: Cliente sem consentimento, inativo ou inexistente é recusado só na sua linha
    Dado que existe o cliente "c-002" ativo
    E que o consentimento do cliente "c-002" está no estado "<estado>"
    E que o arquivo de eventos é:
      """
      cliente_id,tipo,origem,timestamp,valor,categoria
      c-001,compra,loja on-line,2026-11-10T14:30:00Z,150.90,livros
      <cliente>,compra,loja on-line,2026-11-10T14:30:00Z,99.90,livros
      """
    Quando o arquivo é importado
    Então o resultado tem 1 aceita, 0 duplicadas e 1 rejeitada
    E a linha 3 é rejeitada com o erro "<erro>"
    E o cliente "c-002" tem 0 eventos gravados

    Exemplos:
      | estado        | cliente | erro                   |
      | NAO_CONCEDIDO | c-002   | consentimento_invalido |
      | REVOGADO      | c-002   | consentimento_invalido |
      | CONCEDIDO     | c-999   | cliente_nao_encontrado |

  @aplicacao @RN-01
  Cenário: Cliente inativo é recusado na sua linha
    Dado que existe o cliente "c-002" ativo
    E que o consentimento do cliente "c-002" está no estado "CONCEDIDO"
    E que o cliente "c-002" está inativo
    E que o arquivo de eventos é:
      """
      cliente_id,tipo,origem,timestamp,valor
      c-002,navegacao,app,2026-11-10T14:30:00Z,0
      """
    Quando o arquivo é importado
    Então o resultado tem 0 aceitas, 0 duplicadas e 1 rejeitada
    E a linha 2 é rejeitada com o erro "cliente_inativo"

  @aplicacao
  Cenário: Evento já gravado é ignorado como duplicado
    Dado que o cliente "c-001" possui os eventos:
      | tipo   | valor | categoria | quando               |
      | compra | 150.9 | livros    | 2026-11-10T17:30:00Z |
    E que o arquivo de eventos é:
      """
      cliente_id,tipo,origem,timestamp,valor,categoria
      c-001,compra,outra loja,2026-11-10T14:30:00-03:00,150.90,moda
      c-001,compra,loja on-line,2026-11-10T14:30:00-03:00,150.91,livros
      """
    Quando o arquivo é importado
    Então o resultado tem 1 aceita, 1 duplicada e 0 rejeitadas
    E a linha 2 é duplicada
    E a coleta do cliente "c-001" retorna 2 eventos

  @aplicacao
  Cenário: Linha repetida no mesmo arquivo entra uma vez só
    Dado que o arquivo de eventos é:
      """
      cliente_id,tipo,origem,timestamp,valor
      c-001,compra,loja on-line,2026-11-10T14:30:00Z,150.90
      c-001,compra,loja on-line,2026-11-10T14:30:00Z,150.90
      """
    Quando o arquivo é importado
    Então o resultado tem 1 aceita, 1 duplicada e 0 rejeitadas
    E a linha 3 é duplicada

  @aplicacao
  Cenário: Reenviar o mesmo arquivo não grava nada de novo
    Dado que o arquivo de eventos é:
      """
      cliente_id,tipo,origem,timestamp,valor
      c-001,compra,loja on-line,2026-11-10T14:30:00Z,150.90
      c-001,navegacao,app,2026-11-11T09:00:00Z,0
      """
    E que o arquivo já foi importado
    Quando o arquivo é importado
    Então o resultado tem 0 aceitas, 2 duplicadas e 0 rejeitadas
    E a coleta do cliente "c-001" retorna 2 eventos

  @aplicacao @lgpd @RN-03
  Cenário: Consentimento revogado depois da prévia vale na confirmação
    Dado que o arquivo de eventos é:
      """
      cliente_id,tipo,origem,timestamp,valor
      c-001,compra,loja on-line,2026-11-10T14:30:00Z,150.90
      """
    E que a prévia do arquivo mostrou 1 linha aceita
    E que o cliente "c-001" revoga o consentimento
    Quando o arquivo é importado
    Então o resultado tem 0 aceitas, 0 duplicadas e 1 rejeitada
    E a linha 2 é rejeitada com o erro "consentimento_invalido"
    E nenhum evento é gravado

  @aplicacao
  Esquema do Cenário: Arquivo com até 1.000 linhas é aceito, com mais é recusado inteiro
    Dado que o arquivo tem <linhas> linhas de navegação válidas do cliente "c-001"
    Quando a prévia do arquivo é feita
    Então <resultado>

    Exemplos:
      | linhas | resultado                                               |
      | 1000   | o resultado tem 1000 aceitas, 0 duplicadas e 0 rejeitadas |
      | 1001   | o arquivo é recusado com o motivo "linhas_demais"       |

  @aplicacao
  Esquema do Cenário: Arquivo inválido é recusado inteiro e nada é gravado
    Dado que o arquivo de eventos é <arquivo>
    Quando o arquivo é importado
    Então o arquivo é recusado com o motivo "<motivo>"
    E nenhum evento é gravado

    Exemplos:
      | arquivo                                   | motivo               |
      | maior que 256 KiB                         | arquivo_grande       |
      | vazio                                     | arquivo_vazio        |
      | só com o cabeçalho                        | arquivo_vazio        |
      | em Latin-1                                | codificacao_invalida |
      | com o cabeçalho separado por ponto e vírgula | cabecalho_invalido |
      | sem a coluna "valor"                      | cabecalho_invalido   |
      | com a coluna desconhecida "produto"       | cabecalho_invalido   |
      | com aspas mal fechadas                    | formato_invalido     |

  @aplicacao
  Cenário: O BOM do Excel no início do arquivo é ignorado
    Dado que o arquivo de eventos começa com o BOM e é:
      """
      cliente_id,tipo,origem,timestamp,valor
      c-001,navegacao,app,2026-11-11T09:00:00Z,0
      """
    Quando o arquivo é importado
    Então o resultado tem 1 aceita, 0 duplicadas e 0 rejeitadas

  # ---- API (JSON) ----

  @api @wip
  Cenário: Lote em JSON pela API usa as mesmas regras
    Dado que a API está no ar e exige o token "t-123"
    E que existe o cliente "c-002" ativo
    Quando o lote de eventos é enviado à API com o token "t-123":
      """
      {"eventos": [
        {"clienteId": "c-001", "tipo": "compra", "origem": "loja on-line", "timestamp": "2026-11-10T14:30:00Z", "valor": 150.9, "categoria": "livros"},
        {"clienteId": "c-002", "tipo": "compra", "origem": "loja on-line", "timestamp": "2026-11-10T14:30:00Z", "valor": 99.9}
      ]}
      """
    Então a resposta tem status 200
    E o corpo informa 1 aceita, 0 duplicadas e 1 rejeitada
    E o corpo informa a linha 2 com o código "consentimento_invalido"
    E a coleta do cliente "c-001" retorna 1 evento

  @api @wip
  Cenário: A prévia pela API não grava nada
    Dado que a API está no ar e exige o token "t-123"
    Quando o lote de eventos é enviado à API com o token "t-123" só para conferir:
      """
      {"eventos": [{"clienteId": "c-001", "tipo": "navegacao", "origem": "app", "timestamp": "2026-11-10T14:30:00Z", "valor": 0}]}
      """
    Então a resposta tem status 200
    E o corpo informa 1 aceita, 0 duplicadas e 0 rejeitadas
    E nenhum evento é gravado

  @api @wip
  Esquema do Cenário: Lote vazio ou sem token é recusado
    Dado que a API está no ar e exige o token "t-123"
    Quando o lote de eventos é enviado à API com o token "<token>":
      """
      {"eventos": []}
      """
    Então a resposta tem status <status>

    Exemplos:
      | token  | status |
      | t-123  | 422    |
      | errado | 401    |

  # ---- Painel ----

  @web @wip
  Cenário: Enviar o arquivo pelo painel mostra a prévia sem gravar
    Dado que o operador entrou no painel
    E que o arquivo de eventos é:
      """
      cliente_id,tipo,origem,timestamp,valor,categoria
      c-001,compra,loja on-line,2026-11-10T14:30:00Z,150.90,livros
      c-001,compra,loja on-line,2026-11-10T14:30:00Z,150.90,livros
      c-001,compra,loja on-line,2026-11-10T15:30:00Z,0,livros
      """
    Quando o operador envia o arquivo de eventos pelo painel
    Então a resposta tem status 200
    E a página mostra 1 pronta, 1 duplicada e 1 com erro
    E a página mostra "Uma compra precisa ter valor maior que zero."
    E nenhum evento é gravado

  @web @wip
  Cenário: Confirmar a importação grava as linhas aceitas e mostra o relatório
    Dado que o operador entrou no painel
    E que o arquivo de eventos é:
      """
      cliente_id,tipo,origem,timestamp,valor,categoria
      c-001,compra,loja on-line,2026-11-10T14:30:00Z,150.90,livros
      c-001,navegacao,app,2026-11-11T09:00:00Z,0,
      """
    E que o operador enviou o arquivo de eventos pelo painel
    Quando o operador confirma a importação
    Então a página mostra o aviso "importacao-concluida"
    E a página mostra 2 prontas, 0 duplicadas e 0 com erro
    E a coleta do cliente "c-001" retorna 2 eventos

  @web @lgpd @wip
  Cenário: Sem consentimento, a linha aparece recusada na prévia
    Dado que o operador entrou no painel
    E que o consentimento do cliente "c-001" está no estado "REVOGADO"
    E que o arquivo de eventos é:
      """
      cliente_id,tipo,origem,timestamp,valor
      c-001,navegacao,app,2026-11-11T09:00:00Z,0
      """
    Quando o operador envia o arquivo de eventos pelo painel
    Então a página mostra 0 prontas, 0 duplicadas e 1 com erro
    E a página mostra "Sem consentimento válido. O evento não pode ser usado."
    E a página não oferece "Importar"

  @web @wip
  Esquema do Cenário: Arquivo grande é recusado com o motivo
    Dado que o operador entrou no painel
    E que o arquivo de eventos é maior que <tamanho> KiB
    Quando o operador envia o arquivo de eventos pelo painel
    Então a página de destino mostra "O arquivo passa de 256 KiB. Divida em arquivos menores."
    E nenhum evento é gravado

    Exemplos:
      | tamanho |
      | 256     |
      | 600     |

  @web @wip
  Cenário: O modelo do arquivo pode ser baixado
    Dado que o operador entrou no painel
    Quando a página "/painel/eventos/modelo.csv" é aberta
    Então a resposta tem status 200
    E a resposta é um CSV que começa com "cliente_id,tipo,origem,timestamp,valor,categoria"

  @web @C-09 @wip
  Cenário: A auditoria do lote registra o operador e as quantidades, sem dado pessoal
    Dado que o operador entrou no painel
    E que o arquivo de eventos é:
      """
      cliente_id,tipo,origem,timestamp,valor
      c-001,navegacao,app,2026-11-11T09:00:00Z,0
      c-001,compra,app,2026-11-11T10:00:00Z,0
      """
    E que o operador enviou o arquivo de eventos pelo painel
    Quando o operador confirma a importação
    Então o log do painel registra a importação com 1 importado, 0 duplicados e 1 rejeitado
    E o log do painel não contém o ID do cliente "c-001"
