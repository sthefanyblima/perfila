O Perfila Console é a interface de operação do Perfila: onde a equipe cadastra clientes, registra o consentimento LGPD, lança eventos de comportamento e vê o resultado do processamento (perfil, segmento e recomendação). É uma ferramenta de trabalho para usar o dia inteiro, não uma vitrine. Tudo aqui serve para responder rápido, e sem margem de erro, a duas perguntas: **posso usar os dados deste cliente?** e **o que o sistema concluiu sobre ele?**

A identidade visual chama-se **Manhã**: papel quente, violeta como cor de ação, laranja como acento e formas macias. A fonte da verdade visual é o protótipo navegável [`prototipo-manha.html`](prototipo-manha.html), que mostra todas as telas do painel e do portal do titular. Os tokens (`tokens.json`) e os componentes (`bundle.css`) transcrevem os valores do protótipo sem alterá-los.

As seções seguintes cobrem o processo de design thinking que gerou estas regras, a arquitetura de informação, as telas e fluxos, os textos da interface e a arquitetura técnica. Leia esta página primeiro.

## Princípios

1. **O consentimento vem antes de tudo.** O estado do consentimento aparece em todas as telas de um cliente, sempre no mesmo lugar: o **carimbo** à direita do nome, no cabeçalho da ficha (`ConsentRecord`). Sem CONCEDIDO, eventos, perfil, segmento e recomendação nem são carregados: no lugar deles vai um `Alert` crítico que explica o bloqueio e oferece a ação de conceder.
2. **Mostrar o que o sistema fez, na ordem em que fez.** O processamento segue o diagrama de sequência (UML4): coletar → perfil → segmento → recomendação. A interface mostra essa ordem na faixa de processamento (`PipelineTrace`), acende a régua etapa por etapa e diz em qual etapa parou, quando parar. Logo abaixo, o **laudo** conta o resultado em uma frase.
3. **Estado tem forma, palavra e cor, nessa ordem de importância.** Toda etiqueta de estado tem ícone de forma própria, a palavra do domínio e a cor semântica. Nunca só cor.
4. **Vivo, mas sem distrair.** O movimento existe para mostrar que algo mudou (o carimbo bate, a régua acende, a recomendação vira ingresso) ou para responder ao toque (botão que sobe, item que desliza). Nunca decora, nunca atrasa uma ação.
5. **Sólido onde se trabalha, cor no entorno.** Painéis, tabelas e campos são superfícies brancas com borda `line`. A cor viva fica nas formas de fundo, na faixa de processamento, no ingresso da recomendação e no selo de segmento. Sombra só na ação primária, no toast e no diálogo.
6. **O mínimo de dado pessoal na tela.** Listas mostram nome e ID; o e-mail aparece mascarado nas listas e completo só na ficha. IDs aparecem em `UuidTag`, com cópia.
7. **Nenhum nome quebra a tela.** Nomes, e-mails, segmentos e produtos longos quebram linha dentro do próprio contêiner; nada estoura nem se sobrepõe (seção Nomes longos).

## Conteúdo e voz

- Português do Brasil, frases curtas, voz ativa. Fale com quem opera ("Registre o consentimento antes de processar."), sem jargão de implementação: diga **cliente**, **evento**, **perfil**, **segmento**, **recomendação**, **consentimento**, nunca "entidade", "payload", "UC-04" ou "upsert".
- Botões começam pelo verbo e dizem exatamente o que acontece: "Cadastrar cliente", "Registrar evento", "Processar agora", "Conceder consentimento", "Revogar consentimento". O toast confirma no particípio: "Consentimento revogado".
- Caixa de frase em tudo (só a primeira letra maiúscula), exceto rótulos `label` em caixa alta e os estados do domínio, que aparecem como o sistema os nomeia em forma legível: **Concedido**, **Não concedido**, **Revogado**, **Ativa**, **Gerada**, **Expirada**.
- O **laudo** fala do cliente pelo primeiro nome e destaca o essencial com marca (`mark`) e cor (`em`): "Marina está no segmento Padrão: as compras somam R$ 89,90, abaixo do mínimo de R$ 300,00."
- Erros dizem o que aconteceu e como resolver, sem pedir desculpas: "Sem consentimento válido. Registre o consentimento do titular para usar os dados de comportamento." A tabela completa de mensagens por código está na seção de textos.
- Sem emoji, sem exclamação, sem "Ops!".
- Formatos brasileiros: data `11/11/2026 22:07` no fuso de São Paulo (o servidor guarda UTC), dinheiro `R$ 1.234,56`, decimal com vírgula (`0,40`), números alinhados com `font-variant-numeric: tabular-nums`.

## Fundamentos visuais

**Cor.** O fundo da página é `surface-0` (papel quente) com duas formas circulares (`blob-1`, `blob-2`) que se movem de leve com o ponteiro. O trabalho acontece em `surface-1`. Texto em `ink`; secundário em `ink-muted`. `brand` (violeta) é só para ação primária, navegação ativa, foco e os blocos de destaque (faixa de processamento, ingresso); nunca para estado do domínio. `accent` (laranja) marca o selo de segmento e o avatar; o texto sobre ele é `on-accent`, escuro. Estados usam a família semântica com o par cor + tint: `ok`/`ok-tint` (Concedido, Ativa, etapa concluída), `warn`/`warn-tint` (Gerada, sem produto elegível, perfil desatualizado), `critical`/`critical-tint` (Revogado, falha, ação destrutiva), `neutral`/`neutral-tint` (Não concedido, Expirada, cliente inativo ou eliminado). Uma ação primária por tela. Há só o tema claro; o tema escuro da versão 1 foi retirado (Pendências).

**Tipografia.** Três famílias com papéis fixos: `display` (Bricolage Grotesque, peso 800, espaçamento apertado) para o nome do cliente, o título da página, o título de painel, do aviso e do diálogo, o nome da etapa, o laudo (peso 600) e os números grandes; `sans` (DM Sans) para todo o resto; `mono` (JetBrains Mono) para IDs, trilha de navegação, datas de metadados, critérios de segmento e códigos. Escala completa em `tokens.json` (`name`, `display`, `title-1`, `verdict`, `stat`, `body`, `label`…). Texto corrido até `size-measure` em formulários. Na implementação, sirva as fontes do próprio servidor (woff2, licença OFL), sem chamar o Google Fonts: a requisição a terceiros expõe o IP de quem opera.

**Espaço e grade.** Folga de 22px (`space-5`) entre painéis, avisos e blocos empilhados, mesmo quando se encostam (a regra `pf-panel + pf-panel` garante). Padding de painel `space-panel`; margem lateral `space-page`. Dentro de grades (`pf-g2`, `pf-g3`, `pf-stack`) quem manda é o `gap`. Botões e campos têm `size-control` (46px) de altura; o botão pequeno, 36px. Linhas de tabela têm 14px de respiro vertical (`size-row`). A área de trabalho tem até `size-content`.

**Forma.** Quanto maior a superfície, maior o raio: painel 28px, diálogo e barra lateral 30px, faixa de processamento 36px, aviso 20px, botão 15px, campo e item de navegação 14px. Pílula (`radius-pill`) para o que é pequeno e informativo: etiqueta de estado, ID, chip de metadado, aba, critério, selo de segmento e toast. O carimbo é a exceção: cantos de 8px, borda dupla e inclinação.

**Profundidade.** Painéis têm borda de 1px em `line` e nenhuma sombra. A ação primária e o item de navegação ativo levam uma sombra colorida do próprio `brand` (`shadow-primary`). Diálogo e toast usam `shadow-overlay` e `shadow-toast`. O diálogo escurece a página com `scrim`, sem desfoque.

**Movimento.** Curva única `ease`. Botão sobe 2px e a navegação desliza 3px em 250ms; a página entra em 350ms; o carimbo bate em 340ms; cada etapa do processamento dura 750ms e a agulha da régua a percorre em linha reta; ao concluir o processamento ou ativar a recomendação, sai um pequeno estouro de confete do ingresso. As marcações da régua perto do ponteiro crescem. Com `prefers-reduced-motion`, nada se move: só troca de estado.

**Estados de interação.** Hover: botão sobe e a borda vira `ink`; linha de tabela e item de navegação ganham `brand-tint`; ID vira `brand`. Foco: anel sólido de 3px em `focus` com 3px de folga, sempre visível no teclado (`:focus-visible`); em campo, a borda vira `brand` e surge um halo de 4px em `brand-tint`. Desabilitado: 45% de opacidade, sem interação, e o motivo escrito ao lado ("Conceda o consentimento para registrar eventos e processar.").

**Carimbo.** O consentimento é um carimbo, não uma etiqueta: borda dupla na cor do estado, texto em caixa alta, inclinação de -4° e acabamento de tinta irregular. A irregularidade vem de um filtro SVG que a página inclui uma vez, e o componente o referencia por `filter: url(#pf-rough)`:

```html
<svg width="0" height="0" style="position:absolute" aria-hidden="true">
  <filter id="pf-rough" x="-5%" y="-5%" width="110%" height="110%">
    <feTurbulence type="fractalNoise" baseFrequency=".8" numOctaves="2" seed="4" result="n"/>
    <feDisplacementMap in="SourceGraphic" in2="n" scale="2.4" result="d"/>
    <feTurbulence type="fractalNoise" baseFrequency=".55" numOctaves="2" seed="9" result="t"/>
    <feColorMatrix in="t" type="matrix" values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 -7 4.8" result="m"/>
    <feComposite in="d" in2="m" operator="in"/>
  </filter>
</svg>
```

Concedido: verde, borda dupla. Não concedido: neutro, tracejado, "nenhum registro". Revogado: crítico, texto riscado. Eliminado: neutro. O carimbo tem margem própria (10px, 16px embaixo; a versão pequena, 8px e 22px) porque a inclinação o faz ocupar mais que a caixa de layout: sem ela, ele invade o texto vizinho. Texto e estado são sempre reais (`role="status"`), nunca imagem.

## Nomes longos

Todo contêiner tolera nomes longos sem estourar ou se sobrepor:

- Nome do cliente, título, células da primeira coluna, valores de `KeyValue`, avisos, diálogos e selo de segmento quebram com `overflow-wrap: anywhere`.
- O nome grande da ficha reduz a fonte acima de 20 caracteres (`pf-ficha__nome--lg`) e de 40 (`--xl`, com sobrenome corrido na mesma linha).
- Em tabelas, datas, números, etiquetas e e-mail mascarado não quebram (`white-space: nowrap`); a tabela rola na horizontal dentro do próprio contêiner. O nome ocupa uma linha própria e o ID vem abaixo.
- Nome do operador no rodapé do menu termina em reticências. Botões de ação de uma célula ficam sempre lado a lado, à direita (`pf-acell`).
- Grades e colunas de página usam `min-width: 0` nos filhos, para o conteúdo não alargar a página inteira no celular.

## Iconografia

Ícones de traço de 1,7px em grade de 16px, desenhados como SVG inline com `currentColor`, para herdar a cor do texto. Os estados têm forma própria e não dependem de cor:

| Estado | Forma | Cor |
| --- | --- | --- |
| Concedido, Ativa, etapa concluída | círculo com visto | `ok` |
| Não concedido | círculo tracejado | `neutral` |
| Revogado, falha, bloqueada | círculo com barra diagonal (na etapa, "x") | `critical` |
| Gerada, atenção | triângulo | `warn` |
| Expirada, inativo | círculo tracejado | `neutral` |
| Etapa pendente | ponto | `ink-muted` |
| Etapa em andamento | arco girando | `on-stage` |

Tipos de evento: navegação (seta), compra (sacola), interação (envelope). A marca "p" em quadrado de canto vincado (`pf-nav__mark`), girada -6°, é o único logotipo; o nome "Perfila" vem ao lado em texto. Não use emoji nem ilustração além da forma decorativa do `EmptyState`.

## Acessibilidade (WCAG 2.2 AA)

- Teclado: tudo operável por Tab/Shift+Tab/Enter/Espaço/Esc; ordem de foco igual à visual; link "Pular para o conteúdo" no topo; diálogo prende o foco e devolve ao gatilho ao fechar.
- Leitores de tela: tabelas com `<th scope>`; etiquetas e carimbo com texto real (não só ícone); toasts em `aria-live="polite"`, erros de formulário em `aria-live="assertive"` e ligados ao campo por `aria-describedby`; a trilha de processamento também diz o estado em texto.
- Alvo de clique ≥ 24×24px (46px nos controles principais, 36px nos pequenos).
- A interface funciona a partir de 360px de largura: abaixo de 960px a barra lateral vira um menu em gaveta, abaixo de 900px as grades viram uma coluna e a faixa de processamento vira 2×2, e as tabelas rolam na horizontal dentro do próprio contêiner.
- Contraste: ver a seção Contraste abaixo; todos os pares de texto e de campo passam.

## Nomes de classe

`bundle.css` e `internal/web/static/perfila.css` usam os mesmos nomes onde o componente existe nos dois: `pf-ficha`, `pf-stamp`, `pf-stage`, `pf-pipe__step--concluida`, `--atencao` e `--falhou`, `pf-laudo`, `pf-segseal`, `pf-ticket`, `pf-tag--accent`, `pf-search`, `pf-segctl` (no painel, em links com `aria-current="true"`; no design system também em botões com `aria-pressed`), `pf-stat`, `pf-pips`, `pf-stage--rodando` e `pf-confete`. Só existem no design system, sem uso no painel: `pf-switch`, `pf-steps`, `pf-skel`, `pf-pipe__step--run` e `--blocked`. O painel acrescenta `pf-kpi` e `pf-score` (cabeçalho e nota do engajamento, que agora usa `pf-pips`).

## Contraste (WCAG 2.2 AA)

Conferido sobre os valores finais dos tokens (razão de contraste, texto de 12px ou mais pequeno exige 4,5:1; borda de campo e marcação gráfica, 3:1):

| Par | Razão | Onde aparece |
| --- | --- | --- |
| `ok` sobre `ok-tint` | 4,64:1 | Etiqueta Concedido/Ativa, aviso ok |
| `critical` sobre `critical-tint` | 4,78:1 | Etiqueta Revogado, aviso crítico |
| `warn` sobre `warn-tint` | 4,93:1 | Etiqueta Gerada, aviso de atenção |
| `ok` e `critical` sobre `surface-0` | 4,98:1 e 5,44:1 | Texto de estado direto sobre o fundo |
| `stage-muted` sobre `stage` | 5,0:1 | Resultado de cada etapa |
| `stage-lit`, `stage-warn`, `stage-critical` sobre `stage` | 4,77:1, 4,75:1 e 4,59:1 | Número e nome da etapa concluída, com atenção e com falha |
| `line-strong` sobre `surface-1` | 3,57:1 | Borda de campo de formulário |
| `ink` | 15,6:1 a 16,8:1 | Texto principal |
| `ink-muted` | 4,6:1 a 5,5:1 | Texto secundário |
| `brand` e `on-brand` sobre `brand` | 6,1:1 | Ação primária, aba ativa |
| `on-accent` sobre `accent` | 7,1:1 | Selo de segmento e avatar |

Os valores de `ok`, `warn`, `critical`, `stage-muted`, `stage-lit` e `stage-critical` ficaram um pouco mais escuros ou claros que na primeira versão do protótipo para passar; `line-strong` foi acrescentado para a borda dos campos. A borda `line` continua só decorativa (painéis, divisórias). Texto e estado nunca dependem só de cor.

## Limites conhecidos

O protótipo não tem o link "Pular para o conteúdo"; a interface implementada tem.

Fora isso, ainda não existe o **tema escuro** (removido porque o protótipo não o define). A interface implementada (`internal/web`) já usa a identidade Manhã, só no tema claro.
