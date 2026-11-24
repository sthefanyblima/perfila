# ConsentRecord

O consentimento LGPD de um cliente: estado, finalidade, termo, datas e a ação possível a partir dele (UML5). Seu rosto é o **carimbo** (`pf-stamp`).

**Duas formas.** No cabeçalho da ficha, o carimbo grande fica à direita do nome, com a data embaixo, sempre visível; entre os chips acima do nome, "Publicidade: Concedido/Não concedido/Revogado" mostra a segunda finalidade. Na aba Consentimento há um painel por finalidade (Segmentação e Publicidade com terceiros), cada um com o carimbo pequeno (`pf-stamp--sm`), uma `KeyValue` (Finalidade, Versão do termo, Consentimento em, Revogação em, Atualizado em), a ação e a linha do tempo.

**Ações por estado (máquina de estados 3.4.1).** Não concedido → "Conceder consentimento" (primário). Concedido → "Revogar consentimento" (`pf-btn--danger`). Revogado → "Conceder consentimento" (primário). Nunca ofereça uma transição que o domínio recusa. As duas ações abrem o `ConfirmDialog` com o efeito escrito; conceder pede a versão do termo.

**Carimbo.** Concedido: verde, borda dupla. Não concedido: neutro, tracejado, "nenhum registro". Revogado: crítico, texto riscado. Eliminado: neutro. Bate (`is-new`) quando o estado muda. Detalhes e filtro SVG: README, seção Carimbo.

**Linha do tempo.** Concessões e revogações em `pf-timeline` com "por quem": "por Rita (painel)", "por você, no portal". Revogação leva o marcador em `critical`.
