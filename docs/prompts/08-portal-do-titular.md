# Portal do titular

Branch: `feat/portal-titular`

Contexto: o portal do titular ficou para a fase 2 (Q-20): o próprio cliente ver e revogar o
consentimento, sem depender da operação. Depende de "Histórico completo do consentimento" e de
"Direitos do titular"; se "Públicos para campanhas" já estiver na main, o portal também mostra e
revoga a autorização de publicidade.

Leia também: docs/design/01-design-thinking.md (persona Titular), UC-08, docs/uml/UML1 e UML8,
internal/web (sessão, CSRF, cabeçalhos), o checklist LGPD (seção 4).

Decisões a propor e registrar:
- Rotas sob /titular, separadas do painel, com sessão própria (cookie diferente, escopo só do
  próprio cliente). Nada do painel acessível pelo portal.
- Acesso por link de uso único enviado por e-mail: o titular informa o e-mail e a resposta é
  sempre a mesma ("se o e-mail estiver cadastrado, enviamos um link"), para não revelar quem é
  cliente. Token aleatório guardado só como hash, validade de 30 minutos, uso único, limite de
  pedidos por e-mail e por IP.
- Envio de e-mail atrás de um port (Notificador); no MVP, uma "caixa de saída simulada" visível
  só ao ADMINISTRADOR no painel, para a demonstração. O link nunca vai para o log.
- O que o titular vê e faz: cadastro, estado e histórico do consentimento, revogar (com a mesma
  confirmação do painel), conceder de novo com o termo vigente, baixar os próprios dados (JSON da
  portabilidade) e pedir a eliminação (vira um pedido para o ADMINISTRADOR atender no painel).
- Linguagem simples, sem jargão; o mesmo design system, com a marca em destaque.

Tarefas:
1. Spec (ator Titular no UML1, rotas, casos de uso, port Notificador, configuração), UML1, UML7
   e UML8, decisões Q-xx e portal.feature: link expira e não pode ser reusado; e-mail
   desconhecido recebe a mesma resposta; titular só vê os próprios dados; revogar pelo portal tem
   efeito imediato e entra no histórico com a origem "portal"; pedido de eliminação aparece no
   painel.
2. Domínio, ports e repositórios (tokens, pedidos), Notificador simulado, migração, contrato.
3. Casos de uso: SolicitarAcessoTitular, AutenticarTitular, ConsultarMeusDados, PedirEliminacao.
4. Portal web e, no painel, a caixa de saída simulada e a lista de pedidos do titular.
5. Cenário @implantacao abrindo /titular pelo Nginx, README, roteiro e checklist (seção 4).

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
