# Edição de segmentos e catálogo pelo painel

Branch: `feat/editar-segmentos`

Contexto: segmentos e produtos só entram por migração (Spec 1, Q-08, Q-32); as telas Segmentos e
Catálogo do painel são somente leitura. Existe uma pergunta aberta sobre isso em
OPEN_QUESTIONS.md (procure a que trata de ajustar segmentos pelo painel): responda os pontos dela
antes de codar e mova-a para as decididas. Depende de "Usuários, perfis e auditoria" se ela já
estiver na main (só ADMINISTRADOR e ANALISTA editam); se não estiver, registre a restrição como
pendência.

Leia também: RN-06 e RN-07, entity/segmento.go (AvaliarCriterio), SegmentacaoService, os ports
SegmentoRepository e ProdutoRepository, as migrações 001, 002 e 005, internal/infra/memoria/seed.go
e as telas tela_segmentos.html e tela_catalogo.html.

Decisões a propor e registrar:
- Segmentos: criar, editar nome, descrição, critério e prioridade, ativar e desativar (nunca
  excluir, por causa de perfis_cliente.segmento_id). O padrao não pode ser desativado nem ter
  critério; prioridade única entre os ativos; critério validado pela gramática da RN-06 ao salvar
  (ErrDadosInvalidos com a condição errada), para nunca virar ErrSegmentoMalConfigurado (Q-17).
- Prévia antes de salvar: quantos clientes com consentimento válido cairiam em cada segmento com a
  nova configuração (só contagem, sem dado pessoal), e um "testar critério" com valores digitados.
- Perfis já segmentados quando o critério muda: ficam como estão até o próximo processamento, com
  o aviso na tela; opcionalmente uma ação "reprocessar todos" com confirmação.
- Produtos: criar, editar nome, categoria e preço (Produto.AtualizarPreco), ativar e desativar.
  Lembrar que RN-07 escolhe o primeiro ativo da categoria por nome.
- Auditoria de cada alteração (operador, ID, campos alterados).

Tarefas:
1. Spec (RN, casos de uso, ports, rotas), UML2 e UML7, decisões Q-xx e features
   (segmentacao.feature e painel.feature): critério inválido é recusado; padrao protegido;
   prioridade duplicada recusada; desativar tira o segmento da avaliação; prévia correta;
   produto inativo deixa de ser recomendado.
2. Ports (Salvar e ListarTodos em segmentos; Salvar e BuscarPorID em produtos), memória,
   PostgreSQL e contrato.
3. Casos de uso: SalvarSegmento, PreverSegmentacao, SalvarProduto.
4. Painel: formulários de segmento (critério montado por campo, operador e valor, com o texto
   original visível) e de produto, prévia e confirmação, nos padrões do design system.
5. Exemplos, README e roteiro.

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
