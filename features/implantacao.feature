# language: pt
# Fase: Generic (17/11–24/11): Docker, Nginx e PostgreSQL. Fora do filtro do CI até a fase começar (ver Makefile, BDD_TAGS).
@implantacao @slow
Funcionalidade: Implantação com balanceador, réplicas e banco isolado

  Contexto:
    Dado que a implantação está no ar com um balanceador, duas réplicas da aplicação e um banco

  Cenário: Todo o tráfego entra pelo balanceador
    Quando o dispositivo do cliente faz 20 chamadas a "/saude" pelo balanceador
    Então todas as respostas têm status 200
    E as duas réplicas atenderam pelo menos uma chamada

  Cenário: A queda de uma réplica não interrompe o serviço
    Dado que uma das réplicas é parada
    Quando o dispositivo do cliente faz 20 chamadas a "/saude" pelo balanceador
    Então todas as respostas têm status 200

  Cenário: O banco não é acessível pelo dispositivo do cliente
    Quando o dispositivo do cliente tenta conectar ao banco na porta 5432
    Então a conexão é recusada
