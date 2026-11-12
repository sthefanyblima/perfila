# UML5 — Diagrama de Estados: ConsentimentoLGPD

> **Fonte:** Módulo 4 · imagem original: [`img/UML5_estados_consentimento.png`](img/UML5_estados_consentimento.png)
> **Imagem corrigida (v1, gerada do Mermaid abaixo):** [`img/v1/UML5_estados_consentimento.png`](img/v1/UML5_estados_consentimento.png)
> **Fonte da verdade para:** ciclo de vida do consentimento (Spec 3.4.1).
> **Versão:** v1.4 (histórico) · **Decisões relacionadas:** D-02, D-03, Q-34.

## Diagrama (v1.0)

```mermaid
stateDiagram-v2
  [*] --> NAO_CONCEDIDO : sem registro (D-02)
  NAO_CONCEDIDO --> CONCEDIDO : Conceder [finalidade e versaoTermo não vazias]
  CONCEDIDO --> REVOGADO : Revogar
  REVOGADO --> CONCEDIDO : Conceder [finalidade e versaoTermo não vazias]

  NAO_CONCEDIDO : entry / Status = NAO_CONCEDIDO
  CONCEDIDO : entry / Status = CONCEDIDO, DataConsentimento = agora, registra movimento CONCESSAO
  CONCEDIDO : do / tratamento de dados permitido
  REVOGADO : entry / Status = REVOGADO, DataRevogacao = agora, registra movimento REVOGACAO
  REVOGADO : do / interromperProcessamento (RN-03)
```

Não há estado final (D-03). `EstaValido()` é **consulta** (`Status == CONCEDIDO`), não transição.

## Tabela de transições

| De | Evento | Guarda | Para | Efeito |
|---|---|---|---|---|
| NAO_CONCEDIDO | Conceder | finalidade e versão não vazias | CONCEDIDO | DataConsentimento = agora |
| CONCEDIDO | Revogar | — | REVOGADO | DataRevogacao = agora; UC-01..04 passam a falhar |
| REVOGADO | Conceder | finalidade e versão não vazias | CONCEDIDO | DataConsentimento = agora |
| qualquer outra | — | — | (sem mudança) | `ErrTransicaoInvalida` |
| guarda violada | — | — | (sem mudança) | `ErrDadosInvalidos` |

`DataAtualizacao = agora` em toda transição. **v1.4 (Q-34, RN-10):** cada transição aplicada por UC-08 grava, na mesma operação atômica, um `MovimentoConsentimento` (CONCESSAO ou REVOGACAO) com finalidade, versão do termo, data e origem. Transição inválida ou guarda violada não gera movimento. Os movimentos nunca são alterados nem apagados; o estado continua vindo do registro. A transição é checada **antes** da guarda (CONCEDIDO + Conceder com finalidade vazia → `transicao_invalida`).

## Regras para o agente

1. Estados são o tipo `StatusConsentimento`, nunca string solta (C-05).
2. Ausência de registro = NAO_CONCEDIDO; `Consultar` devolve um consentimento sintético.
3. A finalidade é registrada, mas no MVP não filtra o uso dos dados (RN-02).
4. O movimento é montado depois da transição (`NovoMovimento`) e gravado com o registro por `ConsentimentoRepository.Registrar` (v1.4).

## Divergências registradas

| ID | Divergência no original | Decisão |
|---|---|---|
| D-02 | Classe usa `concedido: bool`; o diagrama exige três estados. | Campo `Status` enumerado. |
| D-03 | Transição REVOGADO → estado final rotulada `EstaValido() (retorna true)`. | Sem estado final; `EstaValido` é consulta. |
| — | Nomes `dataConcessao` (estado) × `dataConsentimento` (classe). | Usa-se `DataConsentimento` (classe é a fonte dos nomes). |

## Rastreabilidade

Spec 3.4.1, RN-02, RN-03 · Feature: `consentimento.feature` · Código: `internal/domain/entity/consentimento.go`, `usecase/gerenciar_consentimento.go`
