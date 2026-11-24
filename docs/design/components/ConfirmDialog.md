# ConfirmDialog

Pede confirmação para ações com efeito jurídico ou que afetam muitos registros, dizendo o efeito antes.

**Quando usar.** Conceder e revogar consentimento; expirar recomendações; aplicar retenção; eliminar dados; gerar arquivo de público; desativar segmento ou produto. Não use para ações rotineiras (registrar evento, processar).

**O que fornecer.** `<dialog class="pf-dialog">` aberto com `showModal()`: título `pf-dialog__title` (pergunta curta: "Revogar consentimento?"), texto `pf-dialog__text` com o efeito em linguagem simples (textos na seção Textos) e rodapé `pf-dialog__foot` com Cancelar + a ação. Em ação destrutiva, o foco inicial fica em **Cancelar** e a ação usa `pf-btn--danger-solid`. Em eliminação, acrescente um campo para digitar o nome do cliente; nome diferente gera o erro ao lado do campo. Fundo da página com `scrim`, sem desfoque; clicar fora fecha.

**Forma.** Raio `radius-dialog`, padding de 28px, `shadow-overlay`, entra com leve zoom em 220ms.

**Sem JavaScript.** A ação leva a uma página de confirmação com o mesmo conteúdo.
