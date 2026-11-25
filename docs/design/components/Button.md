# Button

Dispara uma ação; o rótulo começa pelo verbo e diz exatamente o que acontece.

**Variantes.** `pf-btn--primary` para a ação que é o motivo da tela (no máximo uma por tela: "Processar agora", "Cadastrar cliente"), em violeta com sombra colorida. Sem modificador, o botão secundário: fundo branco e borda `line`. `pf-btn--danger` para ações com efeito jurídico ou destrutivo ("Revogar consentimento"), só contorno e texto em `critical`; `pf-btn--danger-solid` só como confirmação dentro do `ConfirmDialog`. `pf-btn--sm` (36px) em tabelas, avisos e linhas de ação. Links discretos ("Sair", "Minha senha") usam `pf-link`.

**Movimento.** Em hover o botão sobe 2px e a borda vira `ink`; ao pressionar, encolhe 2%. O primário aumenta a sombra.

**O que fornecer.** Um `<button type="button|submit">` ou `<a>` com a classe; ícone opcional (SVG 16px com `pf-ico`) antes do texto. Desabilitado: atributo `disabled` (ou `aria-disabled="true"` em links) **e** o motivo escrito ao lado ou abaixo, em `pf-why`.

**Ações em célula.** Vários botões numa célula de tabela vão dentro de `pf-acell`: ficam lado a lado, alinhados à direita, sem quebrar.

**Não faça.** Dois primários na mesma tela. Botão só com ícone sem `aria-label`. "OK", "Enviar", "Clique aqui".
