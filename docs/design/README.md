O Perfila Console é a interface de operação do Perfila: onde a equipe cadastra clientes, registra o consentimento LGPD, lança eventos de comportamento e vê o resultado do processamento (perfil, segmento e recomendação). É uma ferramenta de trabalho para usar o dia inteiro, não uma vitrine. Tudo aqui serve para responder rápido, e sem margem de erro, a duas perguntas: **posso usar os dados deste cliente?** e **o que o sistema concluiu sobre ele?**

As seções seguintes cobrem o processo de design thinking que gerou estas regras, a arquitetura de informação, as telas e fluxos, os textos da interface e a arquitetura técnica. Leia esta página primeiro.

## Princípios

1. **O consentimento vem antes de tudo.** O estado do consentimento aparece em todas as telas de um cliente, sempre no mesmo lugar (cabeçalho da ficha, use `ConsentRecord` resumido). Sem CONCEDIDO, eventos, perfil, segmento e recomendação nem são carregados: no lugar deles vai um `Alert` crítico que explica o bloqueio e oferece a ação de conceder.
2. **Mostrar o que o sistema fez, na ordem em que fez.** O processamento segue o diagrama de sequência (UML4): coletar → perfil → segmento → recomendação. A interface mostra essa ordem com `PipelineTrace` e diz em qual etapa parou, quando parar.
3. **Estado tem forma, palavra e cor, nessa ordem de importância.** Toda etiqueta de estado tem ícone de forma própria, a palavra do domínio e a cor semântica. Nunca só cor.
4. **Densidade com respiro.** Tabelas e listas chave-valor densas, de 40px por linha (`size-row`), mas com espaçamento constante da escala de 4px. Sem cartões soltos flutuando: painéis alinhados à mesma grade.
5. **Superfícies sólidas.** Nada de vidro, desfoque, transparência, gradiente ou brilho. Separação por borda `line`; sombra (`shadow-overlay`) só no que sobrepõe a página.
6. **O mínimo de dado pessoal na tela.** Listas mostram nome e ID; o e-mail aparece mascarado nas listas e completo só na ficha. IDs aparecem truncados em `UuidTag`, com cópia.

## Conteúdo e voz

- Português do Brasil, frases curtas, voz ativa. Fale com quem opera ("Registre o consentimento antes de processar."), sem jargão de implementação: diga **cliente**, **evento**, **perfil**, **segmento**, **recomendação**, **consentimento**, nunca "entidade", "payload", "UC-04" ou "upsert".
- Botões começam pelo verbo e dizem exatamente o que acontece: "Cadastrar cliente", "Registrar evento", "Processar agora", "Conceder consentimento", "Revogar consentimento". O toast confirma no particípio: "Consentimento revogado".
- Caixa de frase em tudo (só a primeira letra maiúscula), exceto rótulos `label` em caixa alta e os estados do domínio, que aparecem como o sistema os nomeia em forma legível: **Concedido**, **Não concedido**, **Revogado**, **Ativa**, **Gerada**, **Expirada**.
- Erros dizem o que aconteceu e como resolver, sem pedir desculpas: "Sem consentimento válido. Registre o consentimento do titular para usar os dados de comportamento." A tabela completa de mensagens por código está na seção de textos.
- Sem emoji, sem exclamação, sem "Ops!".
- Formatos brasileiros: data `11/11/2026 22:07` no fuso de São Paulo (o servidor guarda UTC), dinheiro `R$ 1.234,56`, decimal com vírgula (`0,40`), números alinhados com `font-variant-numeric: tabular-nums`.

## Fundamentos visuais

**Cor.** O fundo da página é `surface-0`; o trabalho acontece em `surface-1`; `surface-2` marca cabeçalho de tabela, barra lateral e linha selecionada. Texto em `ink`; secundário em `ink-muted`. `brand` (anil) é só para ação primária, link, navegação ativa e foco: nunca para estado do domínio. Estados usam a família semântica com o par cor + tint: `ok`/`ok-tint` (Concedido, Ativa, etapa concluída), `warn`/`warn-tint` (Gerada, sem produto elegível, perfil desatualizado), `critical`/`critical-tint` (Revogado, falha, ação destrutiva), `neutral`/`neutral-tint` (Não concedido, Expirada, cliente inativo). Uma ação primária por tela.

**Tipografia.** IBM Plex, três famílias com papéis fixos: `display` (Plex Sans Condensed) só no título da página (`display`) e de painel (`title-1`); `sans` (Plex Sans) para todo o resto (`title-2`, `body`, `body-strong`, `small`, `label`); `mono` (Plex Mono) para IDs, critérios de segmento, códigos e horários (`mono`, `mono-small`). Texto corrido até `size-measure`. Na implementação, sirva as fontes do próprio servidor (woff2), sem chamar o Google Fonts: a requisição a terceiros expõe o IP de quem opera.

**Espaço e grade.** Escala de 4px: `space-1` a `space-12`. Padding de painel `space-4`; folga entre painéis `space-6`; entre o cabeçalho da página e o conteúdo `space-8`. Controles com `size-control` (36px) de altura; linhas de tabela com `size-row` (40px) no mínimo. A área de trabalho tem até `size-content` e uma barra lateral de `size-sidebar`.

**Forma.** Raio pequeno e por papel: `radius-0` em tabela, barra lateral e cabeçalho; `radius-1` em etiquetas (cantos de carimbo); `radius-2` em botão, campo, painel e aviso; `radius-3` só em diálogo e menu. Nada em pílula.

**Profundidade.** Painéis têm borda de 1px em `line` e nenhuma sombra. `shadow-overlay` só em diálogo, menu e toast. Diálogo escurece a página com uma camada opaca de `ink` a 40%, sem desfoque.

**Movimento.** Transições de 120ms em cor e borda; diálogo e toast entram em 160ms, sem quique. Com `prefers-reduced-motion`, nada se move: só troca de estado.

**Estados de interação.** Hover escurece um passo (`brand` → `brand-strong`; superfícies → `surface-2`). Foco: anel sólido de 2px em `focus` com 2px de folga, sempre visível no teclado (`:focus-visible`). Desabilitado: 50% de opacidade no controle inteiro, cursor padrão, e o motivo escrito ao lado ("Conceda o consentimento para registrar eventos").

## Iconografia

Ícones de traço de 1,5px em grade de 16px, desenhados como SVG inline com `currentColor`, para herdar a cor do texto. Os estados têm forma própria e não dependem de cor:

| Estado | Forma | Cor |
| --- | --- | --- |
| Concedido, Ativa, etapa concluída | círculo cheio com visto | `ok` |
| Não concedido | círculo vazado tracejado | `neutral` |
| Revogado, falha | círculo com barra diagonal | `critical` |
| Gerada, atenção | triângulo | `warn` |
| Expirada, inativo | círculo vazado | `neutral` |
| Etapa pendente | anel cinza | `ink-muted` |

Não use emoji nem ilustração. Não há logotipo: o nome "Perfila" aparece em texto `title-2` na barra lateral.

## Acessibilidade (WCAG 2.2 AA)

- Contraste: texto ≥ 4,5:1 e bordas de controle, foco e ícones com significado ≥ 3:1, conferidos nos dois temas para todos os pares acima.
- Teclado: tudo operável por Tab/Shift+Tab/Enter/Espaço/Esc; ordem de foco igual à visual; link "Pular para o conteúdo" no topo; diálogo prende o foco e devolve ao gatilho ao fechar.
- Leitores de tela: tabelas com `<th scope>`; etiquetas de estado com texto real (não só ícone); toasts em `aria-live="polite"`, erros de formulário em `aria-live="assertive"` e ligados ao campo por `aria-describedby`.
- Alvo de clique ≥ 24×24px (36px nos controles principais).
- A interface funciona a partir de 360px de largura: a barra lateral vira menu, tabelas rolam na horizontal dentro do próprio contêiner.

## Uso dos componentes

Os componentes ficam em `components/`, como classes CSS com prefixo `pf-` em `components/bundle.css`, consumidas por templates HTML no servidor (sem framework de JavaScript). Cada componente tem diretrizes e uma prévia ao vivo. `TelaClientes` e `TelaFicha` mostram as duas telas principais montadas com eles.
