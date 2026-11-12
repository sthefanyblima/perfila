# Direitos do titular e retenção

Branch: `feat/direitos-titular`

Contexto: o checklist LGPD marca como pendentes os direitos de acesso, correção, portabilidade e
eliminação (L-19) e a política de retenção após a revogação (L-20, D-13). Cliente.AtualizarDados
existe no domínio, mas não há rota. Depende de "Histórico completo do consentimento" (a prova de
consentimento não pode ser apagada) e, se já estiver na main, de "Usuários, perfis e auditoria"
(só ADMINISTRADOR elimina).

Leia também: docs/CHECKLIST_LGPD.md (seções 4 e 6), RN-03, D-13, entity/cliente.go, os ports de
repositório e as migrações.

Decisões a propor e registrar:
- Correção do cadastro (nome e e-mail) pela ficha, sem exigir consentimento (é cadastro).
- Portabilidade e acesso (art. 18, II e V): exportar em JSON tudo o que o sistema guarda do
  titular (cadastro, consentimento e histórico, eventos, perfil, segmento, recomendações), mesmo
  com consentimento revogado, porque é o próprio titular exercendo o direito.
- Eliminação (art. 18, VI): anonimizar o cadastro e apagar eventos, perfil e recomendações; manter
  o histórico de consentimento como prova (art. 16, I), sem nome nem e-mail. Campo novo no
  Cliente para marcar a anonimização (ADIÇÃO no UML2). Irreversível: confirmação digitando o
  nome do cliente.
- Retenção: dados de comportamento de quem revogou há mais de N dias (configuração, proposta 180)
  são eliminados por um subcomando do administrador e por um botão na tela Operação.
- Métodos novos de remoção nos ports de evento, perfil e recomendação, na suíte de contrato.
- Auditoria de cada pedido atendido.

Tarefas:
1. Spec (RN nova de retenção, casos de uso, ports, rotas, configuração, subcomando), UML2 e UML8,
   decisões Q-xx e direitos.feature: correção vale sem consentimento; exportação traz tudo;
   eliminação apaga o comportamento e preserva a prova de consentimento; retenção só atinge
   revogados vencidos; eliminado não aparece na busca pelo nome antigo.
2. Ports e repositórios (remoções), migração, contrato.
3. Casos de uso: AtualizarCadastro, ExportarDadosTitular, EliminarDadosTitular, AplicarRetencao.
4. Painel: ações na ficha (corrigir, baixar dados, eliminar com confirmação forte) e retenção na
   Operação.
5. Exemplos (um cliente revogado há mais de 180 dias), checklist (L-19 e L-20), README e roteiro.

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
