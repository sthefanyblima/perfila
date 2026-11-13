# Usuários, perfis de acesso e auditoria

Branch: `feat/usuarios-perfis`

Contexto: o painel tem um único operador definido em configuração (PAINEL_USUARIO e
PAINEL_SENHA_HASH, Q-22), com acesso a tudo. O checklist LGPD marca isso como limitação (L-12,
L-23). O UML1 tem três atores humanos: Administrador do Sistema, Analista de Marketing e quem
opera o atendimento (persona Rita em docs/design/01-design-thinking.md).

Leia também: docs/uml/UML1 e UML7, internal/web/sessao.go e handlers_entrar.go,
internal/infra/config, cmd/omega-server/main.go (hash-senha), docs/design/01-design-thinking.md.

Decisões a propor e registrar:
- Operadores no banco: login, nome, papel (ADMINISTRADOR, ANALISTA, ATENDIMENTO como tipo
  enumerado), hash da senha, ativo, criado em. O PAINEL_USUARIO e o PAINEL_SENHA_HASH viram o
  administrador inicial quando não houver nenhum operador.
- Hash de senha atrás de um port (application só pode importar uuid): implementação bcrypt em
  internal/infra; o login passa a ser um caso de uso (AutenticarOperador).
- Matriz de permissões por rota e caso de uso. Proposta: ATENDIMENTO opera clientes,
  consentimento, eventos e processamento; ANALISTA vê relatórios, segmentos, catálogo e públicos
  sem abrir fichas (minimização); ADMINISTRADOR faz tudo, incluindo usuários, auditoria,
  eliminação de dados e operação. Sem permissão: 403 com a tela de erro.
- A sessão continua assinada e sem estado (duas réplicas), mas o operador é conferido no banco a
  cada requisição, para que desativar alguém valha na hora.
- Trilha de auditoria persistida: quando, operador, ação, alvo (ID do cliente, segmento ou
  produto) e resultado, sem dado pessoal. Registrar todas as ações que mudam dados e as leituras
  de ficha. Prazo de retenção da auditoria.
- Troca de senha pelo próprio operador e redefinição pelo administrador; senha mínima de 12
  caracteres; o limite de tentativas continua.

Tarefas:
1. Spec (atores, casos de uso, rotas, configuração), UML1 e UML7, decisões Q-xx e
   painel.feature/usuarios.feature: cada papel só vê o que pode; operador desativado perde o
   acesso na hora; ações ficam na auditoria; a auditoria não tem dado pessoal.
2. Domínio, ports e repositórios (operadores e auditoria), migração, contrato.
3. Casos de uso: AutenticarOperador, GerenciarOperadores, RegistrarAuditoria, ListarAuditoria.
4. Painel: tela Usuários (criar, mudar papel, desativar, redefinir senha), tela Auditoria (filtro
   por operador, ação e período), menu e botões conforme o papel, "Minha senha".
5. Exemplos: um operador de cada papel (senhas de desenvolvimento documentadas, nunca na
   implantação), README, roteiro e checklist (L-12, L-23).

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
