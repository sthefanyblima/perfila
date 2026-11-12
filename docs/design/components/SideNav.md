# SideNav

Navegação principal fixa à esquerda, com três grupos: Clientes, Configuração e Operação.

**O que fornecer.** `<nav class="pf-nav" aria-label="Navegação principal">` com a marca em texto, grupos com rótulo `pf-label` e links `pf-nav__link`. O link da página atual leva `aria-current="page"`. Rodapé com o operador logado e "Sair" (POST com CSRF).

**Regras.** Largura `size-sidebar`, fundo `surface-2`, sem ícones decorativos. Abaixo de 960px vira um menu aberto por um botão "Menu" no cabeçalho.
