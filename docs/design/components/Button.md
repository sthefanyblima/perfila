# Button

Dispara uma ação; o rótulo começa pelo verbo e diz exatamente o que acontece.

**Variantes.** `pf-btn--primary` para a ação que é o motivo da tela (no máximo uma por tela: "Processar agora", "Cadastrar cliente"). Sem modificador, o botão secundário. `pf-btn--quiet` para ações de baixo peso dentro de painéis ("Limpar busca", "Sair"). `pf-btn--danger` para ações com efeito jurídico ou destrutivo ("Revogar consentimento"), e `pf-btn--danger-solid` só como confirmação dentro do `ConfirmDialog`. `pf-btn--sm` em tabelas e rodapés.

**O que fornecer.** Um `<button type="button|submit">` ou `<a>` com a classe; ícone opcional (SVG 16px com `pf-ico`) antes do texto. Desabilitado: atributo `disabled` (ou `aria-disabled="true"` em links) **e** o motivo escrito ao lado.

**Não faça.** Dois primários na mesma tela. Botão só com ícone sem `aria-label`. "OK", "Enviar", "Clique aqui".
