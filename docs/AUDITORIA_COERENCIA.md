# Auditoria de coerência dos `.md` (v0 → v1.0)

Perfila · Equipe Omega · 03/11/2026

Escopo: os três documentos v0 (`01_Especificacao_IA_Generativa.md`, `02_Plano_SDD.md`, `03_Regras_BDD.md`), o PDF consolidado anterior, os oito diagramas e o roadmap do quadro. Cada achado foi corrigido na v1.0 ou virou pergunta em `OPEN_QUESTIONS.md`.

## Resultado

Os três documentos estavam **coerentes entre si** nas regras de negócio (RN-01 a RN-09), nos estados, no catálogo de erros e nos IDs: todos os 73 cenários do core passaram contra uma implementação feita só a partir da Spec, sem mudar nenhuma regra. As correções foram de **alinhamento com o roadmap**, **fidelidade aos diagramas** e **lacunas de cobertura BDD**.

## Achados

| # | Severidade | Onde | Problema | Correção v1.0 |
|---|---|---|---|---|
| A-01 | Alta | Spec §Como usar · Quadro | O quadro pede `CLAUDE.md` + `UML1.md`…`UML8.md` para o Claude Code; os docs geravam `AGENTS.md` e não tinham os diagramas em texto | Constituição movida para `CLAUDE.md` (lido automaticamente); `AGENTS.md` aponta para ele; criados `docs/uml/UML1..UML8_*.md` com Mermaid, regras e divergências |
| A-02 | Alta | Plano §2 | Fases F0–F7 sem datas nem relação com Core/Support/Generic (50/25/25) e com as apresentações de 10/11 e 26/11 | Seção 0 "Roadmap" e classificação DDD dos subdomínios; fases com marco e estado; `specs/tasks.md` com status |
| A-03 | Média | Spec D-07 | Afirmava que o diagrama de casos de uso desenha o «include» na direção errada. No UML1 as setas estão corretas (Gerar Perfil → Coletar; Segmentar → Gerar Perfil); só a sequência inverte | Texto de D-07 corrigido |
| A-04 | Média | Spec 3.1 × UML1 | Spec dizia que UC-08 é do Cliente e que UC-02..04 são do "Agente Inteligente"; o UML1 liga o **Administrador** a UC-08 e não tem Agente Inteligente | D-17: UC-08 com Cliente e Administrador; Agente Inteligente como ator de automação [ADIÇÃO]; Q-04 para corrigir a imagem |
| A-05 | Média | Spec 3.2 × UML2 | `Segmento.ClassificaPerfil` na Spec × `ClassificarPerfil` no diagrama — fere C-04 (nomes fiéis) | D-18: `ClassificarPerfil` |
| A-06 | Média | Spec D-10 × UML2 | Diagrama de Classes põe serviços na "Aplicação" e interfaces de repositório na "Infraestrutura"; não estava registrado | D-10 ampliado |
| A-07 | Baixa | UML2 | Associação "ConsentimentoLGPD possui PerfilCliente" | D-19: `Cliente 1 → 0..1 PerfilCliente` |
| A-08 | Baixa | UML5 × UML2 | `dataConcessao` × `dataConsentimento` | D-20: vale a classe |
| A-09 | Baixa | Nome do produto | Sem nome; artefatos do UML8 usam `omega-*` | Produto **Perfila**; D-21 mantém `omega-server`/`omega_db` |
| A-10 | Média | Spec 3.5 × 3.4.2 | O subcomando `expirar-recomendacoes` não tinha caso de uso que o executasse | `ExpirarRecomendacoesUseCase` [ADIÇÃO] (T-27) |
| A-11 | Baixa | Spec 3.2 | UC-08 "cria o registro se não existir", mas não havia construtor no modelo | `NovoConsentimento` no estado NAO_CONCEDIDO |
| A-12 | Média | BDD `consentimento.feature` | Cenário "A validade depende somente do estado" sem passo **Quando** (fere B-03); feature com duas camadas (fere B-11) | Passo de consulta adicionado; camada única `@aplicacao` |
| A-13 | Média | BDD (B-10) | Faltavam cenários: UC-08 com cliente inexistente; perfil sem segmento em UC-04 (`perfil_nao_encontrado`); valores-limite aceitos em RN-04; critérios válidos em RN-06 | Cenários adicionados (73 no total) |
| A-14 | Baixa | BDD §5 | Exemplo do runner apontava `../features` e não mostrava o filtro por fase | Exemplo real com `BDD_TAGS` e `Strict` |
| A-15 | Baixa | Plano §2 | Portão F1 citava D-01 a D-12, mas a Spec tinha até D-16 | D-01 a D-21 |
| A-16 | Info | PDF consolidado | Versão anterior (documento único) — mapeava seções 4–6 para `plan.md`, `tasks.md`, `bdd-rules.md` | Superado pela v1.0; nomes de arquivo mantidos |

## Pendências da equipe (não resolvíveis por texto)

1. Corrigir as **imagens** dos diagramas conforme D-02, D-03, D-05 a D-10, D-17, D-19 (prazo: 26/11).
2. Responder Q-01 a Q-06 em `OPEN_QUESTIONS.md`.
