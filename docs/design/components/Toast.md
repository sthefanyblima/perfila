# Toast

Confirmação breve de uma ação concluída, no canto inferior direito.

**O que fornecer.** Uma região `pf-toasts` com `aria-live="polite"` e cada `pf-toast` com o ícone e a mensagem no particípio ("Consentimento revogado"). Some sozinho em 5s; fica enquanto o ponteiro ou o foco estiverem nele; botão fechar com `aria-label="Fechar aviso"`.

**Regras.** Uma ação, um toast. Erros não usam toast: usam `Alert` ou o erro do campo, que não somem.
