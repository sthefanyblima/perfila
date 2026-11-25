# Toast

Confirmação breve de uma ação concluída, no canto inferior direito.

**O que fornecer.** Uma região `pf-toasts` com `aria-live="polite"` e cada `pf-toast` com a mensagem no particípio ("Consentimento revogado"): pílula em `ink` com texto da cor do fundo, que sobe e aparece em 250ms e some sozinha em cerca de 2,5s.

**Regras.** Uma ação, um toast. Erros não usam toast: usam `Alert` ou o erro do campo, que não somem.
