# SideNav

Navegação principal à esquerda, numa caixa branca arredondada que flutua com 16px de margem, com grupos por área.

**Grupos.** Clientes (Clientes, Importar eventos, Pedidos do titular), Análise (Relatórios), Campanhas (Públicos), Configuração (Segmentos, Catálogo) e Administração (Operação, Caixa de saída, Operadores, Auditoria). Cada grupo só aparece para o papel que pode abri-lo (RN-11): Atendimento vê Clientes e Configuração; Analista, Análise, Campanhas e Configuração; Administrador, todos.

**O que fornecer.** `<nav class="pf-nav" aria-label="Navegação principal">` com `pf-nav__box`, a marca (`pf-nav__brand`: quadrado violeta inclinado com "p" + "perfila"), grupos com `pf-nav__heading` e links `pf-nav__link` com ícone de traço de 16px. O link da página atual leva `aria-current="page"` e fica preenchido em violeta, com sombra. Em hover, o link ganha `brand-tint` e desliza 3px.

**Rodapé.** `pf-nav__foot` com o operador logado (`pf-avatar` em laranja com a inicial centralizada, nome e papel, que terminam em reticências se longos) e os links `pf-link`: "Minha senha", "Sair". Todos os itens do rodapé começam na mesma posição horizontal.

**Regras.** Largura `size-sidebar`. Abaixo de 960px vira uma gaveta deslizante aberta por um botão "Menu" na barra do topo (`pf-menu-bar`), com `scrim` atrás.
