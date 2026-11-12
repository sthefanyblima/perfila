# Painel de relatórios (UC-05 e UC-07)

Branch: `feat/relatorios`

Contexto: o UML1 tem UC-05 (Analisar Padrões de Comportamento) e UC-07 (Visualizar Relatórios de
Segmentação) para o Analista de Marketing, fora do MVP (D-04). Hoje não há nenhuma visão
agregada. Depende de "Usuários, perfis e auditoria" se ela já estiver na main (ANALISTA e
ADMINISTRADOR veem relatórios).

Leia também: docs/uml/UML1, RN-02, RN-05 a RN-08, os ports de repositório e o design system
(regras de densidade, cores de estado e acento ameixa para dados derivados).

Decisões a propor e registrar:
- Só números agregados, sem nenhum identificador. Métricas de comportamento (perfil, segmento,
  recomendação, eventos) contam apenas clientes com consentimento válido; contagens de cadastro e
  de estado do consentimento podem incluir todos. Grupos com menos de 5 clientes aparecem como
  "menos de 5", para não reidentificar ninguém.
- Indicadores: clientes (total, ativos), consentimento por estado e taxa de concessão,
  distribuição por segmento (quantidade, ticket médio, score médio), recomendações ativas e
  expiradas, categorias mais compradas, evolução mensal de cadastros, eventos e revogações nos
  últimos 6 ou 12 meses.
- Onde agregar: consultas agregadas num port de leitura próprio (com implementação SQL e em
  memória, na suíte de contrato) em vez de carregar tudo na aplicação.
- Gráficos desenhados no servidor em SVG com as cores do design system (sem biblioteca de JS,
  CSP 'self'), com a tabela de dados ao lado para acessibilidade; filtro de período.
- Exportar o relatório agregado em CSV.

Tarefas:
1. Spec (UC-05 e UC-07 parciais, casos de uso, port, rotas), UML1 e UML7, decisões Q-xx e
   relatorios.feature: grupos pequenos ficam ocultos; revogados não entram nas métricas de
   comportamento; números batem com os dados de exemplo.
2. Port de leitura agregada, memória, PostgreSQL e contrato.
3. Caso de uso GerarRelatorio com período.
4. Painel: tela Relatórios com indicadores, gráficos SVG e tabelas, e o CSV agregado.
5. README e roteiro (um passo da demonstração com os relatórios).

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
