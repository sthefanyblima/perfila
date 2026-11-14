# language: pt
@UC-06 @RN-17 @Q-48 @Q-53 @web
Funcionalidade: Públicos e consentimento por finalidade no painel
  A pessoa operadora concede e revoga cada finalidade na ficha, gera o público de um
  segmento com confirmação, baixa o arquivo e vê as remoções pendentes (Spec 3.7.1, v1.9)

  Contexto:
    Dado que a data atual é "2026-11-13T12:00:00Z"
    E que o painel está no ar com o operador "analista" e a senha "senha-de-teste"
    E que existe o segmento ativo "vip" com prioridade 1 e critério "ticketMedio >= 500"
    E que o operador entrou no painel

  @lgpd @RN-02
  Cenário: A ficha mostra um selo por finalidade
    Dado que o cliente "c-001" com o e-mail "c-001@exemplo.test" está no segmento "vip" com as finalidades "segmentacao"
    Quando a página "/painel/clientes/c-001/consentimento" é aberta
    Então a página mostra o selo "CONCEDIDO" da finalidade "segmentacao"
    E a página mostra o selo "NAO_CONCEDIDO" da finalidade "publicidade_terceiros"

  @lgpd @UC-08
  Cenário: Conceder a publicidade pelo painel pede confirmação com o efeito
    Dado que o cliente "c-001" com o e-mail "c-001@exemplo.test" está no segmento "vip" com as finalidades "segmentacao"
    Quando a página "/painel/clientes/c-001/consentimento/conceder?finalidade=publicidade_terceiros" é aberta
    Então a página pede a confirmação "conceder"
    E a página mostra "pode entrar nos públicos enviados a plataformas de anúncio"

  @lgpd @UC-08 @RN-10
  Cenário: Conceder a publicidade pelo painel não muda a segmentação
    Dado que o cliente "c-001" com o e-mail "c-001@exemplo.test" está no segmento "vip" com as finalidades "segmentacao"
    Quando o operador confirma "conceder" a finalidade "publicidade_terceiros" do cliente "c-001"
    Então o consentimento do cliente "c-001" para "publicidade_terceiros" está no estado "CONCEDIDO"
    E o consentimento do cliente "c-001" para "segmentacao" está no estado "CONCEDIDO"
    E o último movimento do cliente "c-001" é "CONCESSAO" com a origem "painel" e o operador "analista"

  Cenário: A tela Públicos mostra quantos entram e quantos ficam de fora
    Dado que o cliente "c-001" com o e-mail "c-001@exemplo.test" está no segmento "vip" com as finalidades "segmentacao,publicidade_terceiros"
    E que o cliente "c-002" com o e-mail "c-002@exemplo.test" está no segmento "vip" com as finalidades "segmentacao"
    Quando o operador abre o público do segmento "vip"
    Então a página mostra que entram 1 e ficam de fora 1
    E a página não mostra "c-001@exemplo.test"

  @Q-53
  Cenário: Gerar o arquivo pede confirmação com o efeito escrito
    Dado que o cliente "c-001" com o e-mail "c-001@exemplo.test" está no segmento "vip" com as finalidades "segmentacao,publicidade_terceiros"
    Quando o operador abre a confirmação do público do segmento "vip"
    Então a página pede a confirmação "exportar"
    E a página mostra "continua sendo dado pessoal"

  @lgpd
  Cenário: Gerar e baixar o arquivo do público
    Dado que o cliente "c-001" com o e-mail "c-001@exemplo.test" está no segmento "vip" com as finalidades "segmentacao,publicidade_terceiros"
    Quando o operador confirma gerar o arquivo do público do segmento "vip"
    E o operador baixa o arquivo gerado
    Então o download é um CSV com o cabeçalho "email_sha256,segmento" e 1 linha
    E o download não contém "c-001@exemplo.test"

  @lgpd @Q-51
  Cenário: O arquivo vencido não é mais baixado
    Dado que o cliente "c-001" com o e-mail "c-001@exemplo.test" está no segmento "vip" com as finalidades "segmentacao,publicidade_terceiros"
    E que o operador gerou o arquivo do público do segmento "vip"
    E que se passaram 24 horas
    Quando o operador baixa o arquivo gerado
    Então a resposta tem status 410
    E a página mostra "Este arquivo não está mais disponível"

  @lgpd @Q-50
  Cenário: Revogar a publicidade pelo painel deixa a remoção pendente
    Dado que o cliente "c-001" com o e-mail "c-001@exemplo.test" está no segmento "vip" com as finalidades "segmentacao,publicidade_terceiros"
    E que o operador gerou o arquivo do público do segmento "vip"
    E que o operador revogou a finalidade "publicidade_terceiros" do cliente "c-001"
    Quando o operador abre o público do segmento "vip"
    Então a página mostra 1 remoção pendente
    E a página oferece "gerar-remocao"

  @lgpd @RN-18
  Cenário: O relatório agregado não detalha grupos pequenos
    Dado que 4 clientes estão no segmento "vip" com ticket médio 600 e a categoria preferida "livros"
    Quando a página "/painel/publicos/relatorio" é aberta
    Então a resposta tem status 200
    E a página mostra "menos de 5"

  @lgpd
  Cenário: Os logs do painel ao exportar não contêm dado pessoal
    Dado que o cliente "c-001" com o e-mail "c-001@exemplo.test" está no segmento "vip" com as finalidades "segmentacao,publicidade_terceiros"
    Quando o operador confirma gerar o arquivo do público do segmento "vip"
    Então o log do painel não contém "c-001@exemplo.test", "Cliente c-001" nem "9c4cd21419f4d9f66ea7"
