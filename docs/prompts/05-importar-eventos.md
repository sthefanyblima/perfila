# Importação de eventos em lote

Branch: `feat/importar-eventos`

Contexto: eventos entram um a um (POST /v1/clientes/{id}/eventos ou o formulário da ficha). O
IntegracaoGateway (e-commerce/CRM) é um stub vazio (D-12, Q-03). Lojas mandam eventos em lote.

Leia também: RN-01, RN-04 (limites de valor e tolerância de relógio), ColetarDadosUseCase,
IntegracaoGateway, docs/uml/UML7 e o formulário de eventos da ficha.

Decisões a propor e registrar:
- Formato: CSV UTF-8 com cabeçalho (cliente_id, tipo, origem, timestamp em ISO 8601, valor com
  ponto, categoria opcional); limite de tamanho e de linhas; modelo para baixar.
- Cada linha passa por ColetarDados.Registrar (RN-01 e RN-04): cliente inexistente, inativo ou sem
  consentimento é rejeitado por linha, sem parar o lote. Decidir se o lote é tudo-ou-nada ou por
  linha (proposta: por linha, com relatório).
- Prévia antes de gravar (quantas linhas válidas e os erros por linha, com o código e a mensagem
  de docs/design/03-textos.md) e confirmação; o arquivo não fica guardado.
- Detecção de duplicados (mesmo cliente, tipo, timestamp e valor) para reenvio seguro.
- Auditoria do lote (operador, quantidade importada e rejeitada), sem dado pessoal.
- Opcional: rota da API para lote em JSON com o mesmo caso de uso.

Tarefas:
1. Spec (caso de uso, rotas, limites), UML7, decisões Q-xx e importacao.feature: linhas válidas
   entram; inválidas aparecem com o motivo; sem consentimento é recusado por linha; duplicado é
   ignorado; arquivo grande é recusado.
2. Caso de uso ImportarEventos (com prévia) e, se precisar, port para duplicados.
3. Painel: tela Importar eventos (enviar, prévia, confirmar, relatório) e modelo CSV.
4. Exemplos de arquivo sintético em docs/, README e roteiro.

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
