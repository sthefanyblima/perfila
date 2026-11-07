# Perguntas em aberto

Registro exigido por C-12. Mudança aprovada entra primeiro na Spec, depois nas features, depois no código.

**Estado em 07/11/2026:** nenhuma pergunta aberta. Q-01 a Q-06 foram decididas; as decisões estão abaixo e refletidas na Spec v1.1.

## Decididas

| ID | Pergunta | Decisão | Por quê | Efeito |
|---|---|---|---|---|
| Q-01 | Ao conceder de novo após revogar, `DataRevogacao` é apagada ou mantida? | **Mantida** como registro da última revogação | Rastro de auditoria: a LGPD exige demonstrar quando o titular revogou; apagar perderia a prova. O status já diz se vale | Nenhuma mudança de código |
| Q-02 | Se nenhum segmento ativo casar (sem o `padrao`), qual erro? | **Falha de configuração** (`ErrSemSegmentoAplicavel`) → HTTP 500, não 422 | O cliente não enviou nada inválido; quem errou foi a configuração dos segmentos. 422 culparia o cliente | `SegmentacaoService` alterado; testes unitários novos. O `padrao` é garantido por migração (Generic) e pelos dados de demonstração |
| Q-03 | O que faz o `IntegracaoGateway.BuscarEventos(desde)`? | **Fora do MVP**: continua stub vazio; eventos entram só por `POST /v1/clientes/{id}/eventos` (D-12) | Não há sistema de e-commerce/CRM real para integrar; inventar o contrato agora seria especulação. A rota de eventos já cobre a coleta | Nenhuma mudança de código; requisito futuro |
| Q-04 | Atores de UC-08 e "Agente Inteligente" no UML1 | **Confirmados**: UC-08 com Cliente (titular) e Administrador; Agente Inteligente como ator de automação de UC-03/UC-04; UC-09 com Cliente | Mantém o que o diagrama original já tinha (Administrador) e torna explícita a automação | Já no Mermaid do UML1 e na imagem `docs/uml/img/v1/UML1_casos_de_uso.png` |
| Q-05 | A regra de recomendação (primeiro produto ativo da categoria, por nome) vai para a apresentação final? | **Sim, mantida no MVP** | "Mais vendido" exige saber o produto de cada compra, e o evento não tem `produtoId` no modelo; a regra atual é determinística e testável. Evolução fica pós-MVP | Nenhuma mudança de código |
| Q-06 | `PUT /consentimento` com consentimento já CONCEDIDO: 409 ou 200 idempotente? | **409 `transicao_invalida`**, como na máquina de estados | Mantém um único comportamento entre domínio e API (UML5); o cliente consulta `GET /consentimento` antes. Trocar a versão do termo exige revogar e conceder de novo, o que deixa o histórico explícito | Nenhuma mudança de código |

## Decisões de modelagem (T-04)

As decisões D-01 a D-23 da Spec (seção 3.9) foram **adotadas como finais em 07/11/2026**. Qualquer mudança futura segue o fluxo Spec → features → código.
