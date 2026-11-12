# ConfirmDialog

Pede confirmação para ações com efeito jurídico ou que afetam muitos registros, dizendo o efeito antes.

**Quando usar.** Conceder e revogar consentimento; expirar recomendações. Não use para ações rotineiras (registrar evento, processar).

**O que fornecer.** `<dialog class="pf-dialog">` aberto com `showModal()`: título `pf-dialog__title` igual ao rótulo da ação, corpo com o efeito em linguagem simples (textos na seção Textos) e rodapé com Cancelar + a ação. Em ação destrutiva, o foco inicial fica em **Cancelar** e a ação usa `pf-btn--danger-solid`. Fundo da página com camada opaca de `ink` a 40%, sem desfoque.

**Sem JavaScript.** A ação leva a uma página de confirmação com o mesmo conteúdo.
