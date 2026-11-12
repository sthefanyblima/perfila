# language: pt
@aplicacao
Funcionalidade: Processamento do cliente etapa por etapa
  O processamento encadeia coleta, perfil, segmento e recomendação (UML4) e informa o estado de
  cada etapa, para a API e o painel mostrarem onde parou (RN-09, Q-26)

  Contexto:
    Dado que a data atual é "2026-03-10T12:00:00Z"
    E que existe o cliente "c-001" ativo com eventos de compra na categoria "livros"
    E que o catálogo possui produtos ativos na categoria "livros"
    E que existem os segmentos ativos:
      | prioridade | nome   | criterio |
      | 99         | padrao |          |

  @UC-01 @UC-02 @UC-03 @UC-04 @RN-09
  Cenário: Todas as etapas concluídas
    Dado que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    Quando o cliente "c-001" é processado
    Então as etapas do processamento são "concluida", "concluida", "concluida" e "concluida"
    E o processamento coletou 3 eventos
    E a recomendação gravada está no estado "ATIVA"

  @RN-07 @RN-09
  Cenário: Sem produto elegível a recomendação fica em atenção e não interrompe
    Dado que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    E que todos os produtos da categoria "livros" estão inativos
    Quando o cliente "c-001" é processado
    Então as etapas do processamento são "concluida", "concluida", "concluida" e "atencao"
    E o processamento não falha

  @RN-06 @RN-09 @Q-17
  Cenário: Falha na segmentação deixa as etapas seguintes pendentes e mantém o perfil
    Dado que o consentimento do cliente "c-001" está no estado "CONCEDIDO"
    E que existem os segmentos ativos:
      | prioridade | nome     | criterio          |
      | 1          | quebrado | ticketMedio >>> 1 |
    Quando o cliente "c-001" é processado
    Então as etapas do processamento são "concluida", "concluida", "falhou" e "pendente"
    E existe um único perfil para o cliente "c-001"
    E nenhuma recomendação é gravada

  @lgpd @RN-01 @RN-09
  Cenário: Sem consentimento a primeira etapa falha e nada é gravado
    Dado que o consentimento do cliente "c-001" está no estado "REVOGADO"
    Quando o cliente "c-001" é processado
    Então a operação falha com o erro "consentimento_invalido"
    E as etapas do processamento são "falhou", "pendente", "pendente" e "pendente"
    E nenhum perfil é gravado para o cliente "c-001"
