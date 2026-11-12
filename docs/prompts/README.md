# Prompts das funcionalidades além do MVP

Cada arquivo é um prompt completo para colar numa sessão nova do agente de código. Rode um por vez,
nesta ordem, e só comece o seguinte depois que o PR do anterior estiver na main: os prompts usam
os próximos números livres de Q-xx e T-xx e algumas funcionalidades dependem de outras.

| Ordem | Prompt | Depende de | Por quê nesta posição |
|---|---|---|---|
| 1 | [Histórico completo do consentimento](01-historico-consentimento.md) | — | Base de prova (L-06) para direitos do titular, públicos e portal |
| 2 | [Usuários, perfis e auditoria](02-usuarios-perfis-auditoria.md) | — | As telas seguintes já nascem com permissões por papel e auditoria |
| 3 | [Edição de segmentos e catálogo](03-editar-segmentos-catalogo.md) | 2 (permissões) | Responde a pergunta aberta sobre ajustar segmentos pelo painel |
| 4 | [Painel de relatórios](04-relatorios.md) | 2 (permissões) | UC-05 e UC-07; só números agregados |
| 5 | [Importação de eventos em lote](05-importar-eventos.md) | — | Alimenta o sistema com volume real |
| 6 | [Direitos do titular e retenção](06-direitos-do-titular.md) | 1, 2 | Fecha L-19 e L-20 do checklist LGPD |
| 7 | [Públicos para campanhas](07-publicos-campanhas.md) | 1, 2 | UC-06; mexe no consentimento por finalidade |
| 8 | [Portal do titular](08-portal-do-titular.md) | 1, 6 (e 7, se pronto) | Fase 2; reaproveita histórico, portabilidade e eliminação |

Lembrete de cronograma: o escopo fecha em 19/11 e o evento final é em 24/11. O que não couber até
lá entra como roadmap pós-MVP na apresentação.
