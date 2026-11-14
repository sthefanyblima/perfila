# Arquitetura técnica da interface

Como o painel entra no código do Perfila sem quebrar a Constituição (`CLAUDE.md`, C-01 a C-12) nem os diagramas.

## Decisão de stack

**HTML renderizado no servidor em Go (`html/template`), CSS próprio com os tokens deste sistema e htmx para atualizações parciais.** Sem SPA, sem Node, sem framework de CSS.

| Alternativa | Por que não |
| --- | --- |
| SPA (React/Vue) chamando a API JSON | O navegador precisaria do `AUTH_TOKEN`, e quem tem o token age sobre qualquer cliente (checklist LGPD, L-12). Exigiria um BFF de qualquer jeito, mais uma cadeia de build (Node) que a equipe de Go não usa |
| Tailwind / biblioteca de componentes | Traz a estética genérica que este sistema evita e uma segunda fonte de verdade para cores e espaços |
| `html/template` + htmx | Mesma linguagem e mesmo binário (`omega-server`, UML8); o token nunca sai do servidor; funciona sem JavaScript (formulário + redirecionamento) e fica melhor com ele |

## Onde fica no código

```text
internal/web/                 adaptador de apresentação HTML (irmão de internal/api)
  router.go                   rotas /painel/...; middlewares sessão, CSRF, cabeçalhos de segurança
  sessao.go                   cookie de sessão assinado (HMAC-SHA256), CSRF, limite de tentativas
  handlers_*.go               um arquivo por área: clientes, ficha, eventos, consentimento, operação
  formato.go                  datas (UTC → America/Sao_Paulo), R$, decimal com vírgula, máscara de e-mail, ID truncado
  mensagens.go                código de erro → mensagem (seção Textos)
  templates/*.html            layout, telas e fragmentos (embed)
  static/                     perfila.css (tokens + componentes), htmx.min.js fixado, fontes IBM Plex woff2, ícones SVG (embed)
internal/application/usecase/ casos de uso novos de leitura e a orquestração do processamento
cmd/omega-server/main.go      monta web.NovoRouter junto da API (raiz de composição, C-08)
```

- **Regra de dependência (C-01):** `web` importa `application`, `domain/entity` e `shared/errs`, como `api`; nunca `infra` nem `database/sql`. Acrescente `{"web", …}` às regras de `internal/arquitetura_test.go` com as mesmas proibições de `api`. Dependências externas permitidas em `web`: `github.com/google/uuid` e `golang.org/x/crypto/bcrypt` (já presente no `go.sum` via pgx).
- **Tempo (C-07):** expiração de sessão e "há 3 minutos" usam o `Clock` injetado; a regra de `TestRegraDoTempo` já cobre `internal/web`.
- **Consentimento (C-03):** a interface não decide nada sozinha. Toda leitura de dado de comportamento passa por um caso de uso que chama `Precondicoes.Verificar`; a tela só reflete o erro `consentimento_invalido`.
- **Logs (C-09):** cada ação registra `operador`, `clienteId` e a ação; nunca nome, e-mail nem o termo de busca.

## Orquestração do processamento (RN-09)

Hoje o encadeamento UC-01 → UC-04 vive no handler `processar` da API. Mova-o para `application/usecase/processar_cliente.go` (`ProcessarClienteUseCase.Executar(ctx, clienteID) (ResultadoProcessamento, error)`), devolvendo o resultado e o estado de **cada etapa** (concluída, atenção, falhou, pendente). A API JSON e o painel passam a chamar o mesmo caso de uso, e o `PipelineTrace` é desenhado direto desse resultado. Os cenários `@api` de processamento precisam continuar verdes.

## O que falta no domínio e na aplicação

Vira **Spec v1.3** antes do código (ordem Spec → features → código; registrar como Q-20 em diante no `OPEN_QUESTIONS.md`):

| Item | Proposta |
| --- | --- |
| Interface no escopo | O painel do operador entra no MVP como tarefas T-31 em diante, antes do fechamento de 19/11; o portal do titular fica para a fase 2 |
| Base legal para mostrar nome e e-mail sem consentimento | Dados cadastrais (UC-09) aparecem ao operador autenticado sob a base de cadastro; dados de comportamento (eventos, perfil, segmento, recomendação) só com CONCEDIDO, como hoje (C-03) |
| Autenticação do operador | Operadores em configuração; senha com hash bcrypt; sessão assinada sem estado no servidor (funciona com as duas réplicas). Perfis de acesso fora do MVP |
| `ClienteRepository.Listar(ctx, FiltroClientes) ([]Cliente, int, error)` | Busca por nome (contém, sem diferenciar maiúsculas e acentos) ou ID exato, paginação de 20, ordenada por data de cadastro decrescente; memória, PostgreSQL e suíte de contrato |
| `RecomendacaoRepository.ListarPorCliente(ctx, clienteID)` | Todas as recomendações do cliente, a mais recente primeiro (histórico na ficha) |
| `ProdutoRepository.ListarTodos(ctx)` | Catálogo completo para a tela de catálogo |
| Casos de uso de leitura | `ListarClientesUseCase` (com estado do consentimento por linha; o segmento só aparece com CONCEDIDO), `ConsultarClienteUseCase` (cadastro + consentimento, sem exigir consentimento, como UC-08), `ConsultarResultadoUseCase` (perfil, segmento e recomendações, com RN-01), `ListarSegmentosUseCase`, `ListarCatalogoUseCase` |
| Índice de busca | Migração `004`: índice em `clientes (data_cadastro DESC)` e `lower(nome)` |
| Configuração nova | `PAINEL_USUARIO`, `PAINEL_SENHA_HASH`, `PAINEL_CHAVE_SESSAO` (32 bytes em hex), `PAINEL_VERSAO_TERMO` (padrão `v1`) (desde a v1.9, `PAINEL_FINALIDADE` não existe mais, Q-48, e entra `PUBLICOS_ARQUIVO_HORAS`, Q-51); gerados por `scripts/preparar-implantacao.sh` |

## Sessão, CSRF e cabeçalhos

- **Cookie `perfila_sessao`:** `operador|expiraEm|csrf`, assinado com HMAC-SHA256 (`PAINEL_CHAVE_SESSAO`), com `HttpOnly`, `Secure`, `SameSite=Strict`, `Path=/painel`, validade de 8h renovada a cada uso. Sem estado no servidor, então as duas réplicas aceitam a mesma sessão.
- **CSRF:** o token vem dentro da sessão; vai em campo oculto `_csrf` nos formulários e no cabeçalho `X-CSRF-Token` nas requisições htmx (`hx-headers` no `<body>`). Todo POST confere.
- **Limite de tentativas no login:** 5 por IP em 5 minutos, em memória por réplica (aceitável no MVP; anotar).
- **Cabeçalhos em toda resposta do painel:** `Content-Security-Policy: default-src 'self'; img-src 'self' data:; style-src 'self'; script-src 'self'; frame-ancestors 'none'; form-action 'self'`, `X-Content-Type-Options: nosniff`, `Referrer-Policy: same-origin`, `Cache-Control: no-store` nas páginas com dado pessoal. Estáticos com `Cache-Control: public, max-age=31536000, immutable` e nome com hash.
- **Fontes e scripts locais:** nada de CDN nem do Google Fonts (CSP `'self'` e nenhuma requisição a terceiros com o IP de quem opera). Fontes IBM Plex (licença OFL) e `htmx.min.js` com versão fixada ficam em `static/`.

## Interação

- **Funciona sem JavaScript:** todo formulário é um `<form method="post">` comum que, no sucesso, redireciona (PRG, 303). O htmx só melhora: `hx-post` troca o fragmento da ficha sem recarregar a página.
- **Fragmentos:** a ficha tem regiões nomeadas (`#selo-consentimento`, `#painel-processamento`, `#linha-do-tempo`); a resposta htmx devolve a região alterada e, com `hx-swap-oob`, o selo de consentimento quando ele muda.
- **Diálogos:** `<dialog>` nativo com `showModal()` (foco preso e Esc já resolvidos pelo navegador); sem JS, a confirmação vira uma página própria (`/painel/clientes/{id}/consentimento/revogar`).
- **Toasts:** cabeçalho `HX-Trigger` com a mensagem, ou um parâmetro na URL de redirecionamento quando não há JS.

## Como provar (C-10)

- **BDD `@web`** em `features/painel.feature`, rodando no `make verify` com `httptest` e fakes em memória, sem navegador. Os passos conferem status, redirecionamentos e marcadores `data-teste` no HTML:
  - sem sessão, qualquer página do painel leva a Entrar;
  - senha errada mostra a mensagem genérica;
  - POST sem CSRF responde 403;
  - sem consentimento, a ficha não mostra perfil, segmento nem eventos e mostra o aviso;
  - revogar pela interface muda o selo e bloqueia o processamento;
  - processamento com falha de configuração mostra a etapa que falhou;
  - os logs do painel não contêm nome nem e-mail.
- **Testes unitários** de `formato.go` (datas, dinheiro, máscara) e `sessao.go` (assinatura, expiração, adulteração), orientados a tabela.
- **Acessibilidade:** navegação completa por teclado nos fluxos A e B; leitor de tela (NVDA) na ficha; contraste já conferido nos tokens.
- **Implantação:** o cenário `@implantacao` passa a abrir `/painel/entrar` pelo Nginx e conferir 200 com os cabeçalhos de segurança.
