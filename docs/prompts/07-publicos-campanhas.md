# Públicos para campanhas (UC-06)

Branch: `feat/publicos`

Depende de: "Histórico completo do consentimento" (01) e, se já estiver na main, "Usuários, perfis e auditoria" (02: só ANALISTA e ADMINISTRADOR geram públicos).

Contexto: você vai implementar no Perfila a exportação de públicos para campanhas de marketing
(tráfego pago e anúncios em redes sociais), a partir dos segmentos que o sistema já calcula.
Hoje isso não é permitido: o consentimento registra uma única finalidade ("segmentacao", uso
interno) e a finalidade não filtra o uso dos dados (RN-02, pendência L-04 do checklist LGPD).
Compartilhar perfil com terceiros é outra finalidade e exige consentimento específico
(LGPD art. 7º §5º e art. 8º §4º). Interpretação técnica da equipe, não parecer jurídico.

Antes de qualquer código:
1. Leia CLAUDE.md (Constituição C-01 a C-12), specs/spec.md (v1.3), specs/tasks.md,
   specs/bdd-rules.md, OPEN_QUESTIONS.md (Q-01 a Q-32), docs/CHECKLIST_LGPD.md,
   docs/uml/UML1, UML2, UML5 e UML7, e docs/design/ (o painel segue o design system
   "Perfila Console" com os ajustes da Q-27).
2. Leia o código de consentimento (internal/domain/entity/consentimento.go,
   GerenciarConsentimentoUseCase, Precondicoes, ValidadorLGPD), dos segmentos e do painel
   (internal/web), e a carga de exemplos (cmd/omega-server/exemplos.go).

Regras do trabalho:
- Crie a branch feat/publicos a partir da main atualizada. Uma tarefa T-xx por vez, mudanças de
  até ~300 linhas, commits pequenos. Rode make verify com PostgreSQL antes de cada commit e faça
  push ao fim de cada tarefa. O Go não está instalado no Windows: use o contêiner golang:1.22
  (docs/ROTEIRO_DEMONSTRACAO.md, seção 4) e um postgres:16-alpine com PERFILA_TEST_DATABASE_URL.
- Commits e PR em português, só com a autoria git configurada. NUNCA adicione "Co-Authored-By",
  "Generated with" ou qualquer menção a Claude ou IA em commits, PRs ou arquivos.
- Ordem sempre Spec → features → código (C-12, B-14). Clean Architecture: o gateway de destino
  é um port no domínio com implementação em internal/infra; internal/web não importa infra.
- Dados de teste e de exemplo sempre sintéticos (c-001, @exemplo.test).
- Não há integração real com Meta, Google ou TikTok neste escopo (sem credenciais nem contrato):
  o destino do MVP é um arquivo CSV gerado pelo próprio sistema, no formato de público
  personalizado (identificador com hash), por trás de um port que permita trocar por API depois.

Decisões a registrar em OPEN_QUESTIONS.md (próximos números Q-xx livres) e propor numa nova versão da Spec. Se alguma
ficar ambígua, PARE, registre e me pergunte antes de codar:
- Modelo de consentimento por finalidade. Opção A: um ConsentimentoLGPD por (cliente,
  finalidade), com a máquina de estados atual por finalidade (muda UNIQUE(cliente_id) para
  (cliente_id, finalidade), UML2 e UML5). Opção B: entidade nova de autorização de
  compartilhamento. Recomende uma, com o impacto em RN-01, RN-02, RN-03 e nas telas.
- Finalidades como tipo enumerado (C-05): "segmentacao" (atual) e "publicidade_terceiros".
  Dados antigos continuam valendo só para "segmentacao".
- RN nova (RN-10, Exportação de públicos): entra no público só quem está ativo, tem consentimento
  CONCEDIDO para "publicidade_terceiros" e está no segmento escolhido. Mínimo de dados: só
  SHA-256 do e-mail normalizado (minúsculas, sem espaços) e o nome do segmento; nunca nome,
  eventos, valores nem o perfil. O hash pseudonimiza, não anonimiza: continua dado pessoal.
- Revogação que se propaga (art. 18 §6º): revogar "publicidade_terceiros" tira o cliente dos
  públicos já gerados (marca a remoção pendente e entra no próximo arquivo de remoção) sem
  afetar a "segmentacao".
- Registro de cada exportação (quando, operador, segmento, destino, quantidade, hash do arquivo),
  sem dado pessoal no log (C-09). Retenção do arquivo gerado: proponha um prazo.
- Transferência internacional (art. 33) e contrato com a plataforma: fica como pendência
  documentada no checklist, não bloqueia o MVP de arquivo.
- Relatório agregado por segmento (contagem, ticket médio, categoria mais comum) sem nenhum
  identificador: pode ser exportado sem a finalidade nova, porque não é dado pessoal. Defina um
  tamanho mínimo de grupo (por exemplo, 5) para não reidentificar ninguém.
- Quem opera: o Analista de Marketing do UML1 (UC-06). No MVP do painel há um único operador
  (Q-22); proponha se a exportação exige confirmação extra (diálogo com o efeito escrito).

Tarefas (registre com os próximos números T-xx livres em specs/tasks.md):
1. Spec: Q-33 em diante, RN-10, UC-06 parcial ("Criar público para campanha"),
     contratos dos casos de uso novos, port do destino, rotas do painel, configuração, UML1, UML2,
     UML5 e UML7 em texto, e features/publicos.feature (@aplicacao, @web e @lgpd) com os
     cenários: só entra quem tem a finalidade; inativo e revogado ficam de fora; revogação
     propaga a remoção; arquivo sem nome nem e-mail em claro; agregado esconde grupos pequenos;
     log sem dado pessoal. Deixe os cenários vermelhos.
2. Domínio: consentimento por finalidade conforme a opção escolhida, filtro por finalidade
     nas pré-condições (fecha L-04) e migração dos dados existentes para "segmentacao".
3. Ports e infraestrutura: repositório de públicos e exportações (memória, PostgreSQL e
     suíte de contrato), DestinoPublicitarioGateway com a implementação em CSV, migração nova.
4. Casos de uso: GerarPublico, ExportarPublico, PropagarRevogacao (chamado pela revogação da
     finalidade) e RelatorioAgregadoSegmentos.
5. Painel: conceder e revogar cada finalidade na ficha (selo e linha do tempo por finalidade),
     tela "Públicos" (escolher segmento, ver quantos entram, gerar e baixar o arquivo, histórico
     de exportações e remoções pendentes) e relatório agregado, seguindo o design system.
6. Exemplos e documentação: alguns clientes de exemplo com a finalidade nova (e um que a
     revogou), checklist LGPD (itens novos e o que segue pendente), README e roteiro.

Aceite: make verify verde com PostgreSQL; cenários @aplicacao, @web, @lgpd e o contrato verdes;
cobertura de domínio + aplicação >= 80%; nenhum dado pessoal em claro no arquivo nem nos logs;
fluxo no navegador: conceder a finalidade a um cliente, gerar o público de um segmento, baixar o
arquivo, revogar e ver a remoção pendente. Ao terminar, abra um PR para a main.
