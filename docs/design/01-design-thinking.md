# Design thinking

As cinco etapas, aplicadas ao Perfila. A origem de cada decisão está nos diagramas (UML1 a UML8), na Spec v1.2 e no checklist LGPD; a interface não inventa regra de negócio.

## 1. Empatizar: quem usa e o que precisa

Os atores vêm do diagrama de casos de uso (UML1). No MVP, a interface serve a quem opera o sistema; o titular e o analista de marketing ficam para depois, com o motivo registrado.

| Persona | Papel no UML1 | Contexto | Precisa | Teme |
| --- | --- | --- | --- | --- |
| **Rita, operação de CRM** | Administrador do Sistema (UC-08, UC-09, comando de expiração) | Atende pedidos do dia: cadastrar cliente, registrar consentimento dado por telefone ou formulário, conferir por que alguém recebeu certa recomendação | Achar o cliente em segundos; saber de relance se pode usar os dados; executar e conferir sem ler JSON | Processar dado de quem revogou; apertar o botão errado num ato que vale juridicamente |
| **Caio, analista de dados** | Agente Inteligente (UC-02 a UC-04), acionado por pessoa | Quer entender por que o cliente caiu no segmento "padrao" e não no "premium" | Ver o perfil calculado, o critério de cada segmento e a etapa em que o processamento parou | Números sem explicação; "erro inesperado" sem pista |
| **Titular** (fase 2) | Cliente (UC-08) | Quer ver e retirar a própria autorização | Uma página só com o próprio consentimento | Linguagem jurídica, botão de revogar escondido |
| **Analista de marketing** (fora do MVP) | UC-05 a UC-07 | Relatórios e campanhas por segmento | Painéis agregados | — |

O que observamos ao percorrer a API e o roteiro de demonstração como se fôssemos a Rita:

- O estado do consentimento decide tudo (C-03), mas na API ele está numa rota separada; quem opera precisa vê-lo **antes** de tentar qualquer coisa.
- "Processar" dispara quatro casos de uso em cadeia (RN-09). Quando algo falha no meio, as etapas anteriores ficam salvas. Sem mostrar a cadeia, o resultado parcial parece um bug.
- IDs são UUIDs de 36 caracteres: ninguém os lê nem dita por telefone. É preciso truncar, copiar e buscar por nome.
- Valores e datas chegam em UTC e em ponto decimal; quem opera pensa em horário de Brasília e em vírgula.
- Erros chegam como códigos estáveis (`consentimento_invalido`): ótimos para máquina, inúteis sem tradução para gente.

## 2. Definir: problemas e perguntas "como poderíamos"

**Declaração do problema.** A Rita precisa operar o ciclo de dados de um cliente (cadastro, consentimento, eventos, processamento) com certeza do que é permitido, porque um erro de consentimento é uma infração da LGPD, e hoje isso só é possível com curl e leitura de JSON.

Como poderíamos…

1. …tornar o estado do consentimento impossível de não ver, em qualquer tela de um cliente?
2. …impedir, na própria interface, ações que o domínio recusaria (em vez de deixar falhar e mostrar o erro)?
3. …explicar o resultado do processamento etapa por etapa, incluindo onde parou?
4. …deixar ações de efeito jurídico (conceder, revogar) deliberadas, sem torná-las lentas?
5. …mostrar dado pessoal só quando necessário?

**Métricas de sucesso** (medidas no teste da etapa 5):

| Métrica | Meta |
| --- | --- |
| Concluir "cadastrar, conceder, registrar 3 eventos e processar" sem ajuda | ≥ 90% das pessoas |
| Tempo dessa tarefa | ≤ 3 min |
| Achar o estado do consentimento de um cliente | ≤ 5 s após abrir a ficha |
| Explicar por que o cliente caiu no segmento atual | ≥ 80% acertam |
| Tentativas de processar sem consentimento | 0 (a interface não oferece a ação) |
| SUS (System Usability Scale) | ≥ 80 |

## 3. Idear: decisões que saíram das perguntas

| Pergunta | Ideias consideradas | Escolhida | Por quê |
| --- | --- | --- | --- |
| 1. Consentimento visível | Ícone na lista; aba própria; **selo fixo no cabeçalho da ficha** | Selo fixo (`ConsentRecord` compacto) + coluna na lista | Está sempre no mesmo lugar, com a mesma forma; a aba sozinha esconderia o estado |
| 2. Impedir ações inválidas | Deixar falhar; **desabilitar com motivo**; esconder | Desabilitar com o motivo escrito ao lado; esconder só o que nem faz sentido | Esconder confunde ("onde está o botão?"); falhar ensina pelo erro |
| 3. Explicar o processamento | Log textual; **trilha de etapas**; diagrama | `PipelineTrace` com as 4 etapas do UML4 | Reflete o modelo mental da sequência e mostra onde parou |
| 4. Ações jurídicas | Botão direto; **diálogo de confirmação com consequência escrita**; digitar "REVOGAR" | Diálogo que diz o efeito ("os dados deixam de ser processados agora") | Deliberado sem burocracia; digitar a palavra seria excesso para uma ação reversível (pode conceder de novo) |
| 5. Dado pessoal | Mostrar tudo; **mascarar nas listas**; pedir senha para ver | Mascarar o e-mail em listas; completo só na ficha | Minimização (LGPD art. 6º, III) com custo baixo de uso |

Princípios que valem para as ideias futuras: estão no README (consentimento primeiro, ordem do processamento, estado com forma, densidade com respiro, superfícies sólidas, mínimo de dado pessoal).

## 4. Prototipar

Protótipos em três fidelidades, nesta ordem:

1. **Papel / quadro:** os fluxos da seção de fluxos, um por folha, com os estados de erro.
2. **Média fidelidade:** as telas `TelaClientes` e `TelaFicha` deste sistema, que já usam os tokens e componentes reais.
3. **Alta fidelidade navegável:** a própria implementação com dados sintéticos (`c-001`, `@exemplo.test`), no modo memória do servidor, que zera a cada reinício e é ótima para testes repetidos.

## 5. Testar

**Roteiro com 5 pessoas** (o suficiente para achar a maior parte dos problemas de usabilidade), 30 min cada, em pensamento em voz alta, sempre com dados sintéticos:

1. "Uma cliente ligou e autorizou o uso dos dados para segmentação, termo v1. Cadastre e registre isso." (cadastro + concessão)
2. "Registre três compras em livros e uma navegação de hoje." (eventos)
3. "Descubra em que segmento ela está e qual produto o sistema recomendaria." (processamento)
4. "Ela ligou de novo pedindo para retirar a autorização. Faça isso." (revogação)
5. "Por que o cliente c-002 está no segmento padrão?" (explicação)

Registrar: conclusão sem ajuda, tempo, erros, hesitações acima de 5 s e o SUS ao final. Um problema que aparecer em 2 ou mais pessoas vira tarefa antes do evento final.

**Testes automáticos de interface** (detalhes na arquitetura técnica): cenários BDD `@web` para as regras (sem sessão não entra; sem consentimento não mostra dado de comportamento; revogar exige confirmação), verificação de contraste e de navegação por teclado.
