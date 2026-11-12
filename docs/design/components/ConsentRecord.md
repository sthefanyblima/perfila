# ConsentRecord

O selo do consentimento LGPD: estado, finalidade, termo, datas e a ação possível a partir dele (UML5).

**Duas formas.** `--compact` no cabeçalho da ficha, sempre visível: selo, estado, uma linha de contexto e a ação. Completa na aba Consentimento: acrescente uma `KeyValue` com Finalidade, Versão do termo, Data do consentimento, Data da revogação (mantida após nova concessão, Q-01) e Última atualização.

**Ações por estado (máquina de estados 3.4.1).** Não concedido → "Conceder consentimento" (primário). Concedido → "Revogar consentimento" (`danger`). Revogado → "Conceder novamente" (primário). Nunca ofereça uma transição que o domínio recusa.

**Selo.** Concedido: círculo cheio com visto (`ok`). Revogado: barra diagonal (`critical`). Não concedido: círculo tracejado com traço (`neutral`). Ambas as ações abrem o `ConfirmDialog` com o efeito escrito.
