# CLAUDE.md — Perfila

> Constituição do projeto para o agente de código (Claude Code). Lida automaticamente no início de cada sessão.
> **Sistema de Segmentação e Perfilamento do Comportamento do Cliente** · Equipe Omega · Go 1.22+
> Versão da Spec: **v1.2** · MVP concluído (Core, Support e Generic; T-01 a T-30) · daqui até 19/11, só correções · escopo fecha em **19/11** · evento final **24/11**

## 1. O que é o projeto

O Perfila coleta eventos comportamentais de clientes (navegação, compra, interação), calcula um **perfil**, atribui um **segmento** por critérios declarativos e gera uma **recomendação** de produto, sempre condicionado ao **consentimento LGPD** do cliente. Arquitetura limpa (Clean Architecture) em Go, desenvolvida com Spec-Driven Development (SDD) e validada por cenários BDD (Gherkin + godog).

"IA generativa" é você, o agente que implementa o código. O "Agente Inteligente" do diagrama de casos de uso é um ator de automação com **regras determinísticas**; não existe modelo de linguagem dentro do produto.

## 2. Mapa de contexto (leia nesta ordem, só o necessário para a tarefa)

| Preciso de… | Arquivo |
|---|---|
| Regras invioláveis | este arquivo, seção 4 |
| O que o sistema faz (regras RN, estados, contratos, erros) | `specs/spec.md` |
| Fases, tarefas T-xx, portões e roadmap | `specs/plan.md` e `specs/tasks.md` |
| Como escrever cenários | `specs/bdd-rules.md` |
| Critérios de aceite executáveis | `features/*.feature` |
| Modelo de cada diagrama, em texto | `docs/uml/UML1..UML8_*.md` |
| Dúvidas em aberto | `OPEN_QUESTIONS.md` |

| UML | Diagrama | Fonte da verdade para |
|---|---|---|
| UML1 | Casos de Uso | atores e casos de uso |
| UML2 | Classes | entidades, campos, métodos e **nomes** |
| UML3 | Pacotes | organização em pacotes (UML7 prevalece) |
| UML4 | Sequência | ordem do fluxo de processamento |
| UML5 | Estados — ConsentimentoLGPD | ciclo do consentimento |
| UML6 | Estados — RecomendacaoProduto | ciclo da recomendação |
| UML7 | Componentes | camadas, ports, direção das dependências |
| UML8 | Implantação | servidores, contêineres, configuração |

Trechos de código dos Módulos 2 e 3 **não** são fonte de verdade (D-11).

## 3. Comandos

```bash
make verify   # gofmt + go vet + testes (-race) + BDD + teste de arquitetura — portão de toda tarefa
make bdd      # só os cenários (filtro da fase em features/bdd_test.go; BDD_TAGS sobrescreve)
make cover    # cobertura de domínio + aplicação (portão F2: >= 80%)
make demo     # fluxo do diagrama de sequência em memória
make implantar && make bdd-implantacao   # UML8 no Docker (Nginx + 2 réplicas + PostgreSQL) e cenários @implantacao
```

## 4. Constituição (regras invioláveis)

Cada regra é verificável em revisão ou por teste automático (`internal/arquitetura_test.go` cobre C-01, C-05, C-07 e C-08).

- **C-01 Regra de dependência.** Importações apontam só para dentro: api → application → domain; infra → domain; shared é importável por todos e não importa ninguém. O pacote `domain` não importa `api`, `application`, `infra`, `net/http`, `database/sql` nem frameworks. Única dependência externa permitida no domínio e na aplicação: `github.com/google/uuid`.
- **C-02 Ports no domínio.** Interfaces de repositório e de gateway ficam em `internal/domain/port`. Implementações ficam em `internal/infra` e provam conformidade em tempo de compilação: `var _ port.PerfilRepository = (*PgPerfilRepository)(nil)`.
- **C-03 Consentimento primeiro.** Nenhum caso de uso lê, grava ou processa dado pessoal sem consentimento válido, exceto UC-08. A checagem acontece no caso de uso (`Precondicoes.Verificar`) e, em reforço, no middleware da API. Sem consentimento: `ErrConsentimentoInvalido`.
- **C-04 Nomes fiéis ao modelo.** Usar exatamente os nomes dos diagramas (UML2). Exportados em PascalCase, internos em camelCase. Não inventar campos, métodos, rotas ou tabelas; o que faltar vira pergunta (C-12).
- **C-05 Tipos explícitos.** Proibido `any` e `interface{}` no domínio e na aplicação. Estados são tipos enumerados (`type StatusConsentimento string` + constantes), nunca strings soltas. Valores monetários em `float64` como no modelo (D-14).
- **C-06 Erros por valor.** Funções que podem falhar retornam `error` como último valor. Erros de domínio são as sentinelas de `internal/shared/errs`, embrulhadas com `%w`. Proibido `panic` em regra de negócio.
- **C-07 Contexto e tempo.** Todo método de caso de uso e de port com I/O recebe `context.Context` como primeiro parâmetro. Proibido `time.Now()` fora de `internal/infra/relogio`; o tempo entra pelo port `Clock`. Datas em UTC.
- **C-08 Sem estado global.** Nenhuma variável de pacote mutável no domínio e na aplicação. Dependências são injetadas por construtor e montadas somente em `cmd/` (raiz de composição).
- **C-09 Logs sem dados pessoais.** Nunca registrar nome ou e-mail; somente IDs.
- **C-10 Testes como prova.** Toda regra de negócio e toda transição de estado tem cenário BDD; regras puras também têm testes orientados a tabela. Testes unitários não usam rede nem banco.
- **C-11 Rastreabilidade.** Tipos, funções públicas e testes citam o ID da spec em comentário: `// Spec: UC-03 RN-06`.
- **C-12 Comportamento do agente.** Não alterar `specs/`, `docs/uml/` nem arquivos `.feature`, exceto em tarefa explícita de atualização de spec. Diante de ambiguidade ou conflito: parar, registrar a dúvida em `OPEN_QUESTIONS.md` e seguir a decisão padrão D-xx, se existir. Uma tarefa (T-xx) por vez, com mudanças de até ~300 linhas.

Os requisitos de LGPD são interpretação técnica da equipe, não parecer jurídico. Dados de teste são sempre sintéticos (`c-001`, `@exemplo.test`).

## 5. Ciclo de uma tarefa

1. Escolha uma tarefa T-xx em `specs/tasks.md` cujas dependências estejam concluídas.
2. Leia só: este arquivo, as seções da Spec citadas na tarefa, os UML citados e as features com as tags da tarefa.
3. Escreva/ajuste passos e testes primeiro (vermelho), depois o código (verde), depois refatore.
4. Rode `make verify`. Até 3 ciclos de correção; depois, pare e peça triagem humana.
5. Resuma: arquivos alterados, cenários cobertos, perguntas abertas.

Modelo de pedido:

```text
Contexto: leia CLAUDE.md, specs/spec.md (seções: <lista>), docs/uml/<UMLx>.md e as features com as tags <lista>.
Tarefa: <T-xx — descrição>.
Restrições: não altere specs/, docs/uml/ nem .feature; não crie campos, métodos, rotas ou tabelas fora da spec;
se algo estiver ambíguo, pare e registre em OPEN_QUESTIONS.md.
Aceite: cenários <tags> verdes; make verify verde; mudança de até ~300 linhas.
```

## 6. Convenções rápidas

- Camadas: `internal/domain/{entity,port,service}`, `internal/application/usecase`, `internal/infra/{memoria,relogio,repository/pg,repository/contrato,external,config}`, `internal/api`, `internal/shared/errs`, `cmd/`.
- Erro → código: `errs.Codigo(err)` devolve o código estável (`consentimento_invalido`, …) usado pela API e pelos cenários.
- Fakes em memória (`internal/infra/memoria`) são a infraestrutura dos cenários `@dominio` e `@aplicacao` e da demo.
- `@implantacao` (`@slow`) fica fora do `make verify` e roda com `make implantar` + `make bdd-implantacao` (Q-13); `@api` roda no verify.
- Sem `DATABASE_URL`, o `omega-server` usa repositórios em memória com dados sintéticos (D-23); com ela, PostgreSQL com migrações aplicadas ao subir (Q-07, Q-09). A implantação sempre a define.
- Repositório novo ou alterado: a suíte `internal/infra/repository/contrato` precisa passar nos dois backends (`make pg-teste`).
