# Histórico completo do consentimento

Branch: `feat/historico-consentimento`

Contexto: hoje o Perfila guarda só o último registro de consentimento por cliente (DataConsentimento,
DataRevogacao mantida após nova concessão, Q-01). O checklist LGPD marca isso como limitação (L-06):
não há prova de cada concessão e revogação, nem de qual versão do termo valia em cada momento. A
linha do tempo da aba Consentimento é montada só com essas datas. Esta funcionalidade é a base
para as de públicos e do portal do titular.

Leia também: docs/uml/UML2 e UML5, internal/domain/entity/consentimento.go,
GerenciarConsentimentoUseCase, ConsultarClienteUseCase, a migração 001 e a aba Consentimento
(internal/web/handlers_ficha.go, LinhaConsentimento).

Decisões a propor e registrar:
- Movimento de consentimento como registro só de inclusão (nunca alterado nem apagado): cliente,
  tipo (concessão ou revogação), finalidade, versão do termo, quando e origem (API, painel com o
  login do operador, portal do titular). Nome da entidade e do port seguindo C-04.
- Como a origem chega ao caso de uso sem mudar as assinaturas da Spec 3.5 (por exemplo, valor no
  context.Context definido pelos adaptadores) ou mudando-as com a Spec atualizada.
- Gravação do movimento na mesma transação do consentimento (no PostgreSQL) e o que fazer com os
  clientes antigos sem movimentos (migração que cria os movimentos a partir das datas atuais).
- O registro atual continua sendo a fonte do estado (RN-02); o histórico é prova.

Tarefas:
1. Spec, UML2 e UML5 em texto, decisão Q-xx e features (consentimento.feature e painel.feature):
   cada conceder e revogar gera um movimento; o histórico preserva a versão do termo de cada
   concessão; a ordem é cronológica; a origem aparece; nada é apagado.
2. Domínio e ports: entidade, port e repositórios (memória, PostgreSQL, contrato) e migração.
3. Casos de uso: GerenciarConsentimento grava o movimento; ConsultarCliente devolve o histórico.
4. Painel: a linha do tempo da aba Consentimento usa o histórico completo (com origem e versão do
   termo); clientes antigos continuam mostrando a linha montada pelas datas.
5. Exemplos com várias idas e vindas (Karina: concede v1, revoga, concede v2) e checklist (L-06).

Antes de qualquer código:
1. Leia CLAUDE.md (Constituição C-01 a C-12), specs/spec.md (versão atual), specs/tasks.md,
   specs/bdd-rules.md, OPEN_QUESTIONS.md (abertas e decididas), docs/CHECKLIST_LGPD.md, os UML
   citados abaixo e docs/design/ (o painel segue o design system "Perfila Console" com os
   ajustes da Q-27: neutros cinza, primária petróleo, acento ameixa, botões e etiquetas sólidos,
   cantos de 6/10/14 px, uma ação de consentimento por tela).
2. Leia o código citado abaixo e o painel em internal/web (rotas, sessão, CSRF, templates,
   perfila.css, painel.js) antes de propor mudanças.

Regras do trabalho:
- Crie a branch indicada a partir da main atualizada. Uma tarefa T-xx por vez (use os próximos
  números livres de specs/tasks.md), mudanças de até ~300 linhas por commit. Rode make verify
  com PostgreSQL antes de cada commit e faça push ao fim de cada tarefa. O Go não está instalado
  no Windows: use o contêiner golang:1.22 (docs/ROTEIRO_DEMONSTRACAO.md, seção 4) e um
  postgres:16-alpine com PERFILA_TEST_DATABASE_URL.
- Commits e PR em português, só com a autoria git configurada. NUNCA adicione "Co-Authored-By",
  "Generated with" ou qualquer menção a Claude ou IA em commits, PRs ou arquivos.
- Ordem sempre Spec → features → código (C-12, B-14). Decisões novas entram em OPEN_QUESTIONS.md
  com o próximo número Q-xx livre e numa nova versão da Spec. Clean Architecture: ports no
  domínio, implementações em internal/infra (memória e PostgreSQL, com casos na suíte de
  contrato), casos de uso em internal/application, internal/web sem importar infra.
- Painel sem framework de JS ou CSS, sem CDN, CSP 'self', funcionando sem JavaScript (PRG) e com
  htmx; textos em português do Brasil, frases curtas, sem emoji.
- Logs sem nome, e-mail nem outro dado pessoal (C-09). Dados de teste e de exemplo sempre
  sintéticos (c-001, @exemplo.test); atualize cmd/omega-server/exemplos.go quando a
  funcionalidade precisar de exemplos.
- Se algo for ambíguo, PARE, registre em OPEN_QUESTIONS.md e me pergunte antes de codar.
- Interpretação técnica da LGPD pela equipe, não parecer jurídico.

Aceite: make verify verde com PostgreSQL; cenários novos e o contrato verdes; cobertura de
domínio + aplicação >= 80%; checklist LGPD e README atualizados; fluxo conferido no navegador
(make serve, http://localhost:8080/painel). Ao terminar, abra um PR para a main.
