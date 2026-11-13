# language: pt
@UC-05 @UC-07 @RN-15 @Q-46
Funcionalidade: Relatórios agregados de segmentação e comportamento
  O analista vê números agregados sobre os clientes, sem identificar ninguém: o comportamento
  só entra de quem tem consentimento válido, e grupos de 1 a 4 clientes aparecem como
  "menos de 5" (Spec RN-15, 3.5 GerarRelatorioUseCase, 3.7.1)

  Contexto:
    Dado que a data atual é "2026-03-10T12:00:00Z"
    E que o catálogo possui produtos ativos na categoria "livros"

  @dominio
  Esquema do Cenário: Uma contagem com 1 a 4 clientes aparece como "menos de 5"
    Quando uma contagem de <valor> com <clientes> clientes distintos é protegida
    Então o valor exibido é "<exibido>"

    Exemplos:
      | valor | clientes | exibido    |
      | 0     | 0        | 0          |
      | 1     | 1        | menos de 5 |
      | 12    | 4        | menos de 5 |
      | 5     | 5        | 5          |
      | 30    | 6        | 30         |

  @dominio
  Esquema do Cenário: A média de um grupo pequeno fica oculta
    Quando a média de uma soma de <soma> com <clientes> clientes é protegida
    Então o valor exibido é "<exibido>"

    Exemplos:
      | soma  | clientes | exibido   |
      | 0     | 0        | sem dados |
      | 400.0 | 4        | oculta    |
      | 500.0 | 5        | 100.00    |

  @aplicacao
  Cenário: Os números batem com os dados de cada indicador
    Dado que existe o segmento ativo "padrao" com prioridade 99 e critério ""
    E que existem 6 clientes com consentimento concedido e compras na categoria "livros"
    E que existem 5 clientes que compraram na categoria "livros" e revogaram o consentimento
    Quando o relatório dos últimos 6 meses é gerado
    Então o relatório mostra:
      | indicador                        | valor |
      | clientes                         | 11    |
      | ativos                           | 11    |
      | consentimento CONCEDIDO          | 6     |
      | consentimento NAO_CONCEDIDO      | 0     |
      | consentimento REVOGADO           | 5     |
      | taxa de concessão                | 54.5  |
      | segmento padrao: clientes        | 6     |
      | segmento padrao: ticket médio    | 100.00 |
      | segmento padrao: score médio     | 0.03  |
      | recomendações ativas             | 6     |
      | recomendações expiradas          | 0     |
      | categoria livros: compras        | 6     |
      | cadastros em 2026-03             | 11    |
      | eventos em 2026-03               | 6     |
      | revogações em 2026-03            | 5     |

  @aplicacao @lgpd
  Cenário: Revogados não entram nas métricas de comportamento
    Dado que existe o segmento ativo "padrao" com prioridade 99 e critério ""
    E que existem 5 clientes que compraram na categoria "livros" e revogaram o consentimento
    Quando o relatório dos últimos 6 meses é gerado
    Então o relatório mostra:
      | indicador                     | valor     |
      | consentimento REVOGADO        | 5         |
      | segmento padrao: clientes     | 0         |
      | segmento padrao: ticket médio | sem dados |
      | recomendações ativas          | 0         |
      | eventos em 2026-03            | 0         |
    E o relatório não mostra a categoria "livros"

  @aplicacao @lgpd
  Cenário: Clientes sem consentimento contam no cadastro e não no comportamento
    Dado que existe o segmento ativo "padrao" com prioridade 99 e critério ""
    E que existem 5 clientes com consentimento concedido e compras na categoria "livros"
    E que existem 5 clientes sem consentimento
    Quando o relatório dos últimos 6 meses é gerado
    Então o relatório mostra:
      | indicador                   | valor |
      | clientes                    | 10    |
      | consentimento NAO_CONCEDIDO | 5     |
      | taxa de concessão           | 50.0  |
      | segmento padrao: clientes   | 5     |
      | cadastros em 2026-03        | 10    |

  @aplicacao
  Cenário: Grupos pequenos ficam ocultos e a taxa de concessão também
    Dado que existe o segmento ativo "padrao" com prioridade 99 e critério ""
    E que existem 3 clientes com consentimento concedido e compras na categoria "livros"
    E que existem 6 clientes sem consentimento
    Quando o relatório dos últimos 6 meses é gerado
    Então o relatório mostra:
      | indicador                     | valor      |
      | clientes                      | 9          |
      | consentimento CONCEDIDO       | menos de 5 |
      | consentimento NAO_CONCEDIDO   | 6          |
      | taxa de concessão             | oculta     |
      | segmento padrao: clientes     | menos de 5 |
      | segmento padrao: ticket médio | oculta     |
      | recomendações ativas          | menos de 5 |

  @aplicacao
  Cenário: Categorias com menos de 5 compradores entram em "Outras categorias", sem nome
    Dado que existe o segmento ativo "padrao" com prioridade 99 e critério ""
    E que existem 5 clientes com consentimento concedido e compras na categoria "livros"
    E que existem 3 clientes com consentimento concedido e compras na categoria "games"
    Quando o relatório dos últimos 6 meses é gerado
    Então o relatório mostra:
      | indicador                       | valor      |
      | categoria livros: compras       | 5          |
      | outras categorias: categorias   | 1          |
      | outras categorias: compras      | menos de 5 |
      | segmento padrao: clientes       | 8          |
    E o relatório não mostra a categoria "games"

  @aplicacao @Q-31
  Cenário: Cliente inativo com consentimento válido entra no comportamento e não nos ativos
    Dado que existe o segmento ativo "padrao" com prioridade 99 e critério ""
    E que existem 3 clientes com consentimento concedido e compras na categoria "livros"
    E que existem 2 clientes inativos com consentimento concedido e compras na categoria "livros"
    Quando o relatório dos últimos 6 meses é gerado
    Então o relatório mostra:
      | indicador                 | valor      |
      | clientes                  | 5          |
      | ativos                    | menos de 5 |
      | segmento padrao: clientes | 5          |
      | categoria livros: compras | 5          |

  @aplicacao
  Esquema do Cenário: As compras contam só dentro do período escolhido
    Dado que existe o segmento ativo "padrao" com prioridade 99 e critério ""
    E que existem 5 clientes com consentimento concedido e compras na categoria "livros" há 8 meses
    Quando o relatório dos últimos <periodo> meses é gerado
    Então as compras da categoria "livros" no relatório são "<compras>"

    Exemplos:
      | periodo | compras  |
      | 6       | nenhuma  |
      | 12      | 5        |

  @aplicacao
  Esquema do Cenário: A evolução mensal cobre os meses do período, contando o mês atual
    Quando o relatório dos últimos <periodo> meses é gerado
    Então o relatório tem <periodo> meses, de "<primeiro>" a "2026-03"

    Exemplos:
      | periodo | primeiro |
      | 6       | 2025-10  |
      | 12      | 2025-04  |

  @aplicacao
  Esquema do Cenário: Só os períodos de 6 e 12 meses são aceitos
    Quando o relatório dos últimos <periodo> meses é gerado
    Então a operação falha com o erro "dados_invalidos"

    Exemplos:
      | periodo |
      | 0       |
      | 3       |
      | 24      |

  @web
  Cenário: A tela Relatórios mostra indicadores, gráficos e tabelas sem identificar ninguém
    Dado que o painel está no ar com o operador "operador" e a senha "s3nha-forte"
    E que o operador entrou no painel
    E que existem 5 clientes com consentimento concedido e compras na categoria "livros"
    Quando a página "/painel/relatorios" é aberta
    Então a página tem o gráfico e a tabela "segmentos"
    E a página tem o gráfico e a tabela "meses"
    E a página não mostra "@exemplo.test"

  @web
  Cenário: Um grupo pequeno aparece por extenso na tela
    Dado que o painel está no ar com o operador "operador" e a senha "s3nha-forte"
    E que o operador entrou no painel
    E que existem 3 clientes com consentimento concedido e compras na categoria "livros"
    Quando a página "/painel/relatorios?periodo=12" é aberta
    Então a página mostra "menos de 5"
    E a página mostra "Últimos 12 meses"

  @web
  Cenário: O relatório agregado é exportado em CSV, sem identificadores
    Dado que o painel está no ar com o operador "operador" e a senha "s3nha-forte"
    E que o operador entrou no painel
    E que existem 5 clientes com consentimento concedido e compras na categoria "livros"
    Quando a página "/painel/relatorios.csv?periodo=6" é aberta
    Então a resposta é um arquivo CSV
    E o arquivo tem a linha "segmentos;padrao;clientes;5"
    E a página não mostra "@exemplo.test"

  @web
  Cenário: Um período fora de 6 e 12 meses é recusado
    Dado que o painel está no ar com o operador "operador" e a senha "s3nha-forte"
    E que o operador entrou no painel
    Quando a página "/painel/relatorios?periodo=3" é aberta
    Então a resposta tem status 422
    E a página mostra "Escolha o período de 6 ou 12 meses."
