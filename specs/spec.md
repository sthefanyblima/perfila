# Especificação para IA Generativa (Spec)

**Perfila — Sistema de Segmentação e Perfilamento do Comportamento do Cliente — Go**

Equipe Omega · GoLang & CLP · Documento 1 de 3 · **Versão 1.10 (14/11/2026)** · Base: Módulos 1 a 5 (casos de uso, classes, pacotes, sequência, estados, componentes e implantação), transcritos em `docs/uml/UML1..UML8`

## Como usar este documento

Este documento transforma a modelagem UML da equipe em artefatos que um agente de IA generativa consegue consumir para gerar o código em Go (Spec-Driven Development, SDD). Premissa adotada: “IA generativa” é o agente de código que implementa o sistema; o “Agente Inteligente” do diagrama de casos de uso é tratado como ator de automação com regras determinísticas, não como um modelo de linguagem dentro do produto.

| Seção | Conteúdo | Arquivo no repositório |
|---|---|---|
| 2 | Constituição: regras invioláveis para o agente | `CLAUDE.md` (lido automaticamente pelo Claude Code; `AGENTS.md` aponta para ele) |
| 3 | Especificação: o que o sistema faz e como se comporta | `specs/spec.md` (este arquivo) |
| — | Os oito diagramas em texto (Mermaid + regras + divergências) | `docs/uml/UML1..UML8_*.md` |

Documentos companheiros: **Documento 2 — Plano SDD** (`specs/plan.md` e `specs/tasks.md`: roadmap, fases, portões e tarefas) e **Documento 3 — Regras BDD** (`specs/bdd-rules.md` e `features/*.feature`).

Marcações: **[ADIÇÃO]** indica algo que não existe nos diagramas, mas é necessário para implementar. **D-xx** indica a decisão padrão adotada para uma lacuna ou inconsistência da modelagem (tabela da seção 3.9). Enquanto a equipe não decidir outra coisa, o agente segue o padrão indicado.

## 1. Fonte da verdade e escopo do MVP

Os diagramas foram feitos em módulos diferentes e nem sempre concordam. Para o agente não escolher ao acaso, cada aspecto tem uma fonte de verdade:

| Aspecto | Fonte de verdade | Arquivo | Observação |
|---|---|---|---|
| Atores e casos de uso | Diagrama de Casos de Uso (Mód. 2) | UML1 | Ver D-04 e D-17 |
| Entidades, campos, métodos e nomes | Diagrama de Classes (Mód. 3) | UML2 | Complementado pelo Diagrama de Componentes |
| Organização em pacotes | Diagrama de Pacotes (Mód. 3) | UML3 | Componentes prevalece (D-10) |
| Ordem do fluxo de processamento | Diagrama de Sequência (Mód. 4) | UML4 | Ordem e responsabilidades; ver D-07 e D-09 |
| Ciclo de vida do consentimento | Diagrama de Estados (Mód. 4) | UML5 | Corrigido por D-02 e D-03 |
| Ciclo de vida da recomendação | Diagrama de Estados (Mód. 4) | UML6 | Corrigido por D-08 |
| Camadas, ports e direção das dependências | Diagrama de Componentes (Mód. 5) | UML7 | Prevalece sobre o de Pacotes (D-10) |
| Servidores, contêineres e configuração | Diagrama de Implantação (Mód. 5) | UML8 | Ver D-21 |
| Trechos de código dos Mód. 2 e 3 | Nenhuma | — | São didáticos e divergem do modelo (D-11) |

**Dentro do MVP:** UC-01, UC-02, UC-03, UC-04, UC-08 e UC-09 (cadastro, D-22) de ponta a ponta, como no diagrama de sequência, com consentimento obrigatório, API HTTP, PostgreSQL e implantação em contêineres; desde a v1.3, também o **painel web do operador** (`/painel`, Q-20), com as regras visuais e de texto do design system "Perfila Console" (`docs/design/`, ajustado pela Q-27); desde a v1.9, os **direitos do titular** pelo painel (correção, acesso e portabilidade, eliminação) e a **retenção** após a revogação (Q-48 a Q-53, RN-17, RN-18); desde a v1.10, o **portal do titular** (`/titular`, Q-55 a Q-60): o próprio cliente entra por um link de uso único enviado por e-mail, vê e revoga o consentimento, concede de novo, baixa os próprios dados e pede a eliminação (RN-19, RN-20).

**Fora do MVP:** UC-05, UC-06 e UC-07 (D-04); notificação real ao cliente (o portal usa uma caixa de saída simulada, Q-57); autenticação JWT e perfis de acesso; gestão de segmentos pela API (segmentos entram por migração); rotas da API JSON para os direitos do titular (o titular os exerce pelo portal, Q-53, Q-55).

## 2. Constituição

A Constituição (regras C-01 a C-12) foi movida para **`CLAUDE.md`**, na raiz do repositório, para ser carregada automaticamente pelo Claude Code em toda sessão. Os IDs não mudaram. Mudança da v1.0: C-01 agora diz explicitamente que `shared` não importa nenhuma outra camada, e C-12 protege também `docs/uml/`.

## 3. Especificação

### 3.1 Atores e casos de uso

| ID | Caso de uso | Ator | Caso de uso de aplicação | MVP |
|---|---|---|---|---|
| UC-01 | Coletar Dados Comportamentais do Cliente | Cliente | `ColetarDadosUseCase` | Sim |
| UC-02 | Gerar Perfil do Cliente (incluído por UC-03) | Agente Inteligente | `GerarPerfilUseCase` | Sim |
| UC-03 | Segmentar Clientes Automaticamente | Agente Inteligente | `SegmentarClienteUseCase` | Sim |
| UC-04 | Recomendar Produtos Personalizados («extend» de UC-03) | Agente Inteligente | `GerarRecomendacaoUseCase` | Sim |
| UC-05 | Analisar Padrões de Comportamento | Analista de Marketing | — | Não (D-04) |
| UC-06 | Criar Campanha de Marketing Direcionada | Analista de Marketing | — | Não (D-04) |
| UC-07 | Visualizar Relatórios de Segmentação | Analista de Marketing | — | Não (D-04) |
| UC-08 | Gerenciar Consentimento de Dados (LGPD) | Cliente (titular) e Administrador do Sistema | `GerenciarConsentimentoUseCase` | Sim |
| UC-09 | Cadastrar Cliente [ADIÇÃO, D-22] | Cliente / Sistema Externo | `CadastrarClienteUseCase` | Sim |

Administrador do Sistema: ligado a UC-08 no UML1 e, pelo UML8, executa os subcomandos do CMD (`serve`, `expirar-recomendacoes`, `hash-senha` (Q-28) e `carregar-exemplos` (Q-30)) por SSH. No painel web (v1.3), é a pessoa **operadora**: cadastra (UC-09), registra o consentimento (UC-08), lança eventos (UC-01), processa (UC-01 a UC-04) e expira recomendações.

Casos de uso de aplicação do painel [ADIÇÃO, Q-23, Q-26]. Não são casos de uso novos do UML1: dão ao operador a leitura do que UC-01 a UC-09 gravaram e reúnem a orquestração de RN-09.

| Caso de uso de aplicação | Exige consentimento (RN-01) | Usado por |
|---|---|---|
| `ProcessarClienteUseCase` (RN-09) | Sim, em cada etapa | API (`POST .../processamento`) e painel |
| `ListarClientesUseCase` | Não; o segmento de cada linha só aparece com consentimento válido (Q-21) | Painel |
| `ConsultarClienteUseCase` | Não: só cadastro e consentimento (Q-21), como UC-08 | Painel |
| `ConsultarResultadoUseCase` | Sim (RN-01, exceto cliente ativo: inativo é só leitura, Q-31) | Painel |
| `ListarSegmentosUseCase`, `ListarCatalogoUseCase` | Não: configuração, sem dado pessoal | Painel |

Agente Inteligente: ator de automação [ADIÇÃO] que dispara UC-03 e UC-04 (D-17).

Casos de uso de aplicação dos direitos do titular [ADIÇÃO, v1.9, Q-48 a Q-53]. Atendem pedidos do titular (art. 18 da LGPD) feitos à pessoa operadora. Não exigem consentimento, porque não usam os dados para perfilamento: corrigem, entregam ou apagam (exceção prevista em C-03).

| Caso de uso de aplicação | Exige consentimento (RN-01) | Usado por |
|---|---|---|
| `AtualizarCadastroUseCase` (correção, art. 18, III) | Não: é cadastro, como UC-09 (Q-48) | Painel |
| `ExportarDadosTitularUseCase` (acesso e portabilidade, art. 18, II e V) | Não: o próprio titular recebe o que o sistema guarda (Q-49) | Painel |
| `EliminarDadosTitularUseCase` (eliminação, art. 18, VI; RN-17) | Não; recusada com consentimento CONCEDIDO (Q-50) | Painel |
| `AplicarRetencaoUseCase` (RN-18) | Não: atinge só quem revogou (Q-51) | Subcomando `aplicar-retencao` e painel |

Titular [ADIÇÃO, v1.10, Q-55]: é o ator Cliente do UML1 quando age sozinho, pelo portal `/titular`, sobre os próprios dados (UC-08 e os direitos do titular). Entra por um link de uso único (RN-19) e só alcança o próprio cadastro.

| Caso de uso de aplicação | Exige consentimento (RN-01) | Usado por |
|---|---|---|
| `SolicitarAcessoTitularUseCase` (RN-19) | Não: só procura o cadastro pelo e-mail e envia o link | Portal |
| `AutenticarTitularUseCase` (RN-19) | Não: confere o link | Portal |
| `ConsultarMeusDadosUseCase` | Não: cadastro, consentimento e histórico, como `ConsultarCliente` (Q-21) | Portal |
| UC-08 `Conceder`, `Revogar`, `Historico` e `ExportarDadosTitularUseCase` | Não (C-03) | Portal, com a origem `portal` (RN-10) |
| `PedirEliminacaoUseCase` (RN-20) | Não; recusado com consentimento CONCEDIDO (Q-58) | Portal |
| `ListarPedidosEliminacaoUseCase`, `ListarCaixaSaidaUseCase` | Não: a operação acompanha os pedidos e a caixa de saída simulada | Painel |

### 3.2 Modelo de domínio

Pacote `internal/domain/entity`. Campos marcados com [ADIÇÃO] não existem no diagrama de classes.

```go
type TipoEvento string          // navegacao, compra, interacao
type StatusConsentimento string // NAO_CONCEDIDO, CONCEDIDO, REVOGADO
type StatusRecomendacao string  // GERADA, ATIVA, EXPIRADA

type Cliente struct {
    ID           uuid.UUID
    Nome         string
    Email        string
    DataCadastro time.Time
    Ativo        bool
    DataEliminacao *time.Time // [ADIÇÃO, v1.9, Q-50] preenchida quando o cadastro é anonimizado (RN-17)
}

type ConsentimentoLGPD struct {
    ID                uuid.UUID
    ClienteID         uuid.UUID
    Status            StatusConsentimento // [ADIÇÃO] substitui Concedido bool (D-02)
    Finalidade        string
    VersaoTermo       string
    DataConsentimento time.Time
    DataRevogacao     *time.Time          // [ADIÇÃO]
    DataAtualizacao   time.Time           // [ADIÇÃO]
}

// [ADIÇÃO, v1.4, Q-34] Prova de cada concessão e revogação (RN-10). Só inclusão.
type TipoMovimento string   // CONCESSAO, REVOGACAO
type OrigemMovimento string // api, painel, portal, sistema, migracao

type MovimentoConsentimento struct {
    ID          uuid.UUID
    ClienteID   uuid.UUID
    Tipo        TipoMovimento
    Finalidade  string          // a concedida ou a que foi revogada
    VersaoTermo string          // idem; vazia só em revogação reconstruída sem termo conhecido
    Data        time.Time       // DataAtualizacao da transição
    Origem      OrigemMovimento
    Operador    string          // login do operador; obrigatório só com Origem painel
}

type EventoComportamental struct {
    ID        uuid.UUID
    ClienteID uuid.UUID
    Tipo      TipoEvento
    Origem    string
    Timestamp time.Time
    Valor     float64
    Categoria string // [ADIÇÃO] opcional (D-05)
}

type PerfilCliente struct {
    ID                 uuid.UUID
    ClienteID          uuid.UUID
    FrequenciaCompras  int
    TicketMedio        float64
    CategoriaPreferida string
    ScoreEngajamento   float64    // entre 0 e 1
    SegmentoID         *uuid.UUID // [ADIÇÃO] (D-06)
}

type Segmento struct {
    ID         uuid.UUID
    Nome       string
    Descricao  string
    Criterio   string // gramática em RN-06
    Prioridade int    // [ADIÇÃO] menor valor = avaliado primeiro
    Ativo      bool   // [ADIÇÃO]
}

type RecomendacaoProduto struct {
    ID            uuid.UUID
    ClienteID     uuid.UUID
    ProdutoID     uuid.UUID
    Motivo        string
    Status        StatusRecomendacao // [ADIÇÃO] (D-08)
    DataGeracao   time.Time
    DataAtivacao  *time.Time         // [ADIÇÃO]
    Validade      *time.Time         // [ADIÇÃO]
    DataExpiracao *time.Time         // [ADIÇÃO]
}

type Produto struct {
    ID        uuid.UUID
    Nome      string
    Categoria string
    Preco     float64
    Ativo     bool
}

// [ADIÇÃO, v1.10, Q-56] Link de uso único do portal do titular. O token em claro só existe no link.
type TokenAcessoTitular struct {
    ID        uuid.UUID
    ClienteID uuid.UUID
    Hash      string     // SHA-256 do token, 64 dígitos hexadecimais minúsculos
    CriadoEm  time.Time
    ExpiraEm  time.Time  // CriadoEm + 30 minutos
    UsadoEm   *time.Time // preenchida no primeiro uso
}

// [ADIÇÃO, v1.10, Q-58] Pedido de eliminação feito pelo titular no portal.
type StatusPedido string // ABERTO, ATENDIDO

type PedidoEliminacao struct {
    ID              uuid.UUID
    ClienteID       uuid.UUID
    Status          StatusPedido
    DataPedido      time.Time
    DataAtendimento *time.Time
}

// [ADIÇÃO, v1.10, Q-57] Mensagem gravada pelo NotificadorSimulado na caixa de saída simulada.
type MensagemSimulada struct {
    ID        uuid.UUID
    Para      string // e-mail mascarado (c***@exemplo.test), nunca o completo
    Assunto   string
    Corpo     string // texto com o link de acesso
    EnviadaEm time.Time
}
```

| Entidade | Métodos | Observação |
|---|---|---|
| Cliente | `NovoCliente(id uuid.UUID, nome, email string, agora time.Time) (Cliente, error)` (construtor, D-22), `AtualizarDados(nome, email string) error`, `Ativar()`, `Desativar()`; [ADIÇÃO, v1.9, Q-50] `Anonimizar(agora time.Time) error`, `EstaEliminado() bool` | Nome e e-mail não podem ser vazios; nasce ativo, com DataCadastro = agora. `Anonimizar` troca Nome por "Titular eliminado" e Email por `eliminado@anonimizado.invalid`, desativa e preenche DataEliminacao = agora; num cliente já eliminado, `ErrClienteEliminado`. `AtualizarDados` num cliente eliminado: `ErrClienteEliminado` |
| ConsentimentoLGPD | `NovoConsentimento(id, clienteID uuid.UUID, agora time.Time)` (construtor, estado NAO_CONCEDIDO), `Conceder(finalidade, versaoTermo string, agora time.Time) error`, `Revogar(agora time.Time) error`, `EstaValido() bool` | Transições da seção 3.4.1 |
| MovimentoConsentimento [ADIÇÃO, Q-34] | `NovoMovimento(id uuid.UUID, c ConsentimentoLGPD, tipo TipoMovimento, origem OrigemMovimento, operador string) (MovimentoConsentimento, error)` | Construído logo depois da transição, com Finalidade, VersaoTermo e Data (= `DataAtualizacao`) do consentimento. Tipo e Origem fora dos enumerados, ou Origem painel sem operador: `ErrDadosInvalidos`. Não tem métodos de alteração |
| EventoComportamental | `Validar(agora time.Time) error` | RN-04 |
| PerfilCliente | `CalcularScore(eventos []EventoComportamental, agora time.Time, meta int)`, `AtualizarPerfil(novo PerfilCliente)` | RN-05. `AtualizarPerfil` copia as métricas e preserva ID, ClienteID e SegmentoID. `meta <= 0` resulta em score 0 |
| Segmento | `ClassificarPerfil(p PerfilCliente) (bool, error)` | RN-06. Nome corrigido na v1.0 para o do UML2 (era `ClassificaPerfil`, contrariando C-04) |
| RecomendacaoProduto | `NovaRecomendacao(...)` (equivale a `Gerar()`), `Disponibilizar(agora time.Time, validadeDias int) error`, `Expirar(agora time.Time) error` | Transições da seção 3.4.2 |
| Produto | `AtualizarPreco(preco float64) error` | Preço maior ou igual a zero |
| TokenAcessoTitular [ADIÇÃO, v1.10] | `NovoTokenAcesso(id, clienteID uuid.UUID, hash string, agora time.Time) (TokenAcessoTitular, error)`, `EstaValido(agora time.Time) bool`, `Usar(agora time.Time) error` | Hash fora do formato ou ClienteID nulo: `ErrDadosInvalidos`. ExpiraEm = agora + 30 min. Válido enquanto não usado e com agora antes de ExpiraEm. `Usar` num token inválido: `ErrAcessoInvalido`; senão, UsadoEm = agora |
| PedidoEliminacao [ADIÇÃO, v1.10] | `NovoPedidoEliminacao(id, clienteID uuid.UUID, agora time.Time) (PedidoEliminacao, error)`, `Atender(agora time.Time) error` | Nasce ABERTO, com DataPedido = agora; ClienteID nulo: `ErrDadosInvalidos`. `Atender`: ABERTO → ATENDIDO com DataAtendimento = agora; já ATENDIDO: `ErrTransicaoInvalida` |
| MensagemSimulada [ADIÇÃO, v1.10] | — | Só leitura depois de gravada |

Desvios idiomáticos em relação aos diagramas (não são lacunas): `context.Context` como primeiro parâmetro; `error` no lugar de `bool` em `Validar` para informar a causa; construtor no lugar de `Gerar()`; tempo recebido por parâmetro ou `Clock`, nunca `time.Now()` (C-07); assinaturas dos serviços ajustadas na seção 3.6.

### 3.3 Regras de negócio

- **RN-01 Pré-condições comuns (UC-01 a UC-04).** Verificar nesta ordem: cliente existe (`ErrClienteNaoEncontrado`), cliente não foi eliminado (`ErrClienteEliminado`, v1.9, RN-17), cliente está ativo (`ErrClienteInativo`), consentimento é válido (`ErrConsentimentoInvalido`).
- **RN-02 Consentimento válido.** Válido se e somente se Status é CONCEDIDO. A ausência de registro equivale a NAO_CONCEDIDO (D-02). A finalidade é registrada, mas no MVP não filtra o uso dos dados.
- **RN-03 Revogação.** Tem efeito imediato: a partir dela, UC-01 a UC-04 falham com `ErrConsentimentoInvalido`. Dados já persistidos permanecem até a retenção (RN-18) ou um pedido de eliminação (RN-17); desde a v1.9 (Q-51), a D-13 foi substituída.
- **RN-04 Validação de evento.** ClienteID não nulo; Tipo entre navegacao, compra e interacao; Origem não vazia; Timestamp não zero e menor ou igual a agora mais 2 minutos de tolerância de relógio (Q-16); Valor maior ou igual a zero e no máximo 1.000.000 (Q-15); se Tipo é compra, Valor maior que zero. Qualquer violação: `ErrDadosInvalidos`.
- **RN-05 Perfil.** Função pura dos eventos do cliente e do relógio:
    - FrequenciaCompras é o número de eventos do tipo compra.
    - TicketMedio é a média do Valor das compras (zero se não houver compras).
    - CategoriaPreferida é a categoria mais frequente entre as compras com categoria preenchida; empate resolvido por ordem alfabética; vazia se não houver.
    - ScoreEngajamento é o menor valor entre 1 e N dividido por `ENGAJAMENTO_META_EVENTOS`, onde N é o número de eventos de qualquer tipo com Timestamp maior que agora menos 30 dias e menor ou igual a agora.
    - Cliente sem eventos gera perfil zerado, sem erro.
    - Há um perfil por cliente: `Salvar` faz upsert por ClienteID e preserva o ID original.
- **RN-06 Segmentação.** Segmentos ativos são avaliados em ordem crescente de Prioridade; o primeiro cujo critério casa vence. Critério vazio casa sempre e deve ter a maior Prioridade (segmento padrao, criado por migração). Gramática do critério:
    - critério: vazio, ou condição seguida de zero ou mais condições ligadas por `&&`
    - condição: campo, operador e número
    - campo: frequenciaCompras, ticketMedio ou scoreEngajamento
    - operador: `>`, `>=`, `<`, `<=` ou `==`
    - número: decimal com ponto
    - A avaliação usa função própria, sem reflexão e sem execução dinâmica de código. Critério inválido: `ErrDadosInvalidos`. Todas as condições são validadas, mesmo depois de uma condição falsa, para que um critério mal escrito nunca passe despercebido.
    - Nenhum segmento ativo casa (só possível sem o segmento padrao): falha de configuração `ErrSemSegmentoAplicavel`, que não é erro de domínio e vira HTTP 500 (Q-02).
    - Na segmentação (UC-03), um segmento configurado com critério inválido, e na recomendação (UC-04) um perfil apontando para segmento inexistente, também são falhas de configuração (`ErrSegmentoMalConfigurado`, HTTP 500), não erro do cliente (Q-17). A avaliação isolada de um critério continua devolvendo `ErrDadosInvalidos`.
    - O resultado é persistido em `PerfilCliente.SegmentoID` (D-06).
- **RN-07 Recomendação (regra provisória, D-05).** Candidatos são os produtos ativos da CategoriaPreferida do perfil; escolhe-se o primeiro por Nome em ordem alfabética, com desempate por ID. Sem categoria ou sem candidatos: `ErrSemProdutoElegivel`, que não é falha (a API devolve recomendação nula). O Motivo segue o formato `segmento:<nome>;categoria:<categoria>`.
- **RN-08 Ciclo da recomendação.** UC-04 cria e já disponibiliza: GERADA é transitória e a recomendação é persistida uma única vez, já em ATIVA. Há no máximo uma recomendação ATIVA por cliente: antes de gravar a nova, UC-04 aplica `Expirar` às ATIVA anteriores do cliente (Q-18). Validade é DataAtivacao mais `RECOMENDACAO_VALIDADE_DIAS` (padrão 30). Expira quando Validade é menor ou igual a agora, pelo comando `expirar-recomendacoes` (D-08).
- **RN-09 Processamento.** `ProcessarClienteUseCase` orquestra, nesta ordem, UC-01 (Executar), UC-02, UC-03 e UC-04 (v1.3, Q-26; antes ficava no handler da API) e devolve o resultado consolidado (sequência, passo 23) com o estado de cada etapa: concluída, atenção, falhou ou pendente. Erro em qualquer etapa interrompe o fluxo: a etapa fica como falhou, as seguintes como pendentes, e o que já foi persistido permanece. `ErrSemProdutoElegivel` não interrompe: a etapa 4 fica em atenção, sem recomendação. A API e o painel usam o mesmo caso de uso.
- **RN-10 Histórico do consentimento (v1.4, Q-34).** Toda transição de 3.4.1 aplicada por UC-08 gera um `MovimentoConsentimento`, gravado junto com o registro do consentimento numa única operação atômica (`ConsentimentoRepository.Registrar`): ou os dois ficam gravados, ou nenhum. Movimentos são só de inclusão: nunca alterados nem apagados. O histórico é a lista dos movimentos do cliente em ordem cronológica (desempate pela ordem de gravação). O registro de `consentimentos_lgpd` continua sendo a fonte do estado (RN-02); o histórico é prova e não decide a validade. Transição inválida ou guarda violada não gera movimento. A origem vem do adaptador que chamou o caso de uso (API: `api`; painel: `painel` com o login do operador; portal do titular, fase 2: `portal`); sem origem informada (subcomandos, exemplos, testes): `sistema`. Movimentos criados pela migração 006 a partir das datas existentes têm origem `migracao`.
- **RN-17 Eliminação a pedido do titular (v1.9, Q-50).** `EliminarDadosTitularUseCase` atende o art. 18, VI, nesta ordem:
    1. Cliente existe (`ErrClienteNaoEncontrado`) e ainda não foi eliminado (`ErrClienteEliminado`).
    2. A confirmação é o nome atual do cliente, comparado sem diferenciar maiúsculas, acentos e espaços nas pontas. Diferente: `ErrDadosInvalidos`, e nada é apagado.
    3. Consentimento CONCEDIDO: `ErrTransicaoInvalida`. É preciso revogar antes, pelo UC-08, o que deixa o movimento REVOGACAO como prova.
    4. Apaga os eventos, o perfil e as recomendações do cliente e, desde a v1.10, os tokens de acesso ao portal (Q-58).
    5. Desde a v1.10 (RN-20), marca como ATENDIDO o pedido de eliminação aberto do cliente, se houver.
    6. Por último, `Anonimizar` e grava o cadastro. Sendo o último passo, uma falha antes dele deixa o cliente sem eliminar, e repetir conclui o pedido.
    - O consentimento e os movimentos ficam como prova do tratamento (art. 16, I): não têm nome nem e-mail.
    - A eliminação é irreversível. Repetir depois de uma falha conclui o pedido, porque as remoções não falham sem dados.
    - Depois dela, nenhuma escrita vale para o cliente: a correção, UC-08 Conceder e Revogar e as pré-condições RN-01 respondem `ErrClienteEliminado`. A busca pelo nome antigo não o encontra, porque o nome não existe mais. A exportação entrega o que restou.
- **RN-18 Retenção após a revogação (v1.9, Q-51).** O comportamento (eventos, perfil e recomendações) de quem está com o consentimento REVOGADO e `DataRevogacao` há `RETENCAO_DIAS_REVOGADO` dias ou mais (menor ou igual a agora menos o prazo) é apagado por `AplicarRetencaoUseCase`. O cadastro, o consentimento e o histórico ficam. Não atinge quem concedeu de novo (CONCEDIDO), quem nunca concedeu nem quem está dentro do prazo. Devolve quantos clientes tiveram algum dado apagado; executar de novo sem nada vencido devolve 0. Executada pelo subcomando `aplicar-retencao` e pelo painel (Operação).
- **RN-19 Acesso do titular ao portal (v1.10, Q-56).**
    1. `SolicitarAcessoTitularUseCase` recebe um e-mail (sem espaços nas pontas; vazio ou sem `@`: `ErrDadosInvalidos`) e procura os cadastros com esse e-mail, sem diferenciar maiúsculas. Para cada cadastro não eliminado com menos de 3 links criados na última hora, gera um token de 32 bytes aleatórios, grava só o hash SHA-256 (`TokenAcessoTitular`, 30 minutos) e envia pelo `Notificador` o link `PORTAL_URL_BASE + "/titular/acesso?t=" + token`.
    2. O resultado é o mesmo para e-mail desconhecido, cadastro eliminado e limite atingido: nada é enviado e nenhum erro é devolvido. Quem chama nunca sabe se o e-mail existe.
    3. `AutenticarTitularUseCase` recebe o token, calcula o hash e o consome no repositório (`Consumir`, atômico): inexistente, expirado ou já usado é `ErrAcessoInvalido`, e um segundo uso do mesmo link falha mesmo em outra réplica. Cliente inexistente ou eliminado também é `ErrAcessoInvalido`. Sucesso devolve o ID do cliente, que vira a sessão do portal.
    4. O token e o link nunca vão para o log (C-09).
- **RN-20 Pedido de eliminação pelo portal (v1.10, Q-58).** `PedirEliminacaoUseCase` confere, nesta ordem: cliente existe (`ErrClienteNaoEncontrado`), não foi eliminado (`ErrClienteEliminado`) e o consentimento não está CONCEDIDO (`ErrTransicaoInvalida`: é preciso revogar antes, como na RN-17). Se já houver um pedido ABERTO do cliente, devolve esse mesmo pedido; senão, cria um pedido ABERTO. Quem atende é a operação, pela eliminação da RN-17, que marca o pedido como ATENDIDO. Com pedido aberto, o portal não oferece conceder de novo.

### 3.4 Máquinas de estado

#### 3.4.1 ConsentimentoLGPD

| De | Evento | Guarda | Para | Efeito |
|---|---|---|---|---|
| NAO_CONCEDIDO | Conceder | Finalidade e VersaoTermo não vazias | CONCEDIDO | DataConsentimento = agora |
| CONCEDIDO | Revogar | — | REVOGADO | DataRevogacao = agora; interrompe o processamento (RN-03) |
| REVOGADO | Conceder | Finalidade e VersaoTermo não vazias | CONCEDIDO | DataConsentimento = agora |

DataAtualizacao = agora em toda transição; cada transição aplicada por UC-08 gera um movimento (RN-10: Conceder gera CONCESSAO, Revogar gera REVOGACAO). Qualquer outra combinação de estado e evento: `ErrTransicaoInvalida`. Guarda violada: `ErrDadosInvalidos`. Não existe estado final (D-03) e `EstaValido()` é uma consulta, não uma transição.

#### 3.4.2 RecomendacaoProduto

| De | Evento | Guarda | Para | Efeito |
|---|---|---|---|---|
| (início) | Gerar | — | GERADA | DataGeracao = agora |
| GERADA | Disponibilizar | — | ATIVA | DataAtivacao = agora; Validade = agora mais N dias; registra log `recomendacao_ativada` (notificação real fora do MVP) |
| ATIVA | Expirar | — | EXPIRADA | DataExpiracao = agora; deixa de ser exibida |
| EXPIRADA | — | — | (final) | — |

Qualquer outra combinação: `ErrTransicaoInvalida`. O comando `expirar-recomendacoes` seleciona recomendações ATIVA com Validade menor ou igual a agora e aplica `Expirar`.

### 3.5 Contratos dos casos de uso e catálogo de erros

| Erro (sentinela) | Quando | HTTP | Código no corpo |
|---|---|---|---|
| `ErrClienteNaoEncontrado` | Cliente inexistente | 404 | cliente_nao_encontrado |
| `ErrClienteInativo` | Cliente desativado | 409 | cliente_inativo |
| `ErrClienteEliminado` [v1.9] | Cliente com os dados eliminados (RN-17) | 409 | cliente_eliminado |
| `ErrConsentimentoInvalido` | Consentimento diferente de CONCEDIDO | 403 | consentimento_invalido |
| `ErrTransicaoInvalida` | Transição fora das tabelas da seção 3.4 | 409 | transicao_invalida |
| `ErrDadosInvalidos` | Argumento, evento ou critério inválido | 422 | dados_invalidos |
| `ErrPerfilNaoEncontrado` | Perfil ausente ou ainda sem segmento | 409 | perfil_nao_encontrado |
| `ErrSemProdutoElegivel` | RN-07 | Não é erro HTTP: 200 com recomendação nula | sem_produto_elegivel |
| `ErrAcessoInvalido` [v1.10] | Link do portal inexistente, expirado ou já usado; cliente do link eliminado (RN-19) | 401 | acesso_invalido |
| `ErrConsentimentoNaoEncontrado` | Uso interno do repositório; o caso de uso traduz em NAO_CONCEDIDO | — | — |
| `ErrPedidoNaoEncontrado` [v1.10] | Uso interno: cliente sem pedido de eliminação aberto | — | — |

**UC-09 `CadastrarClienteUseCase`** [ADIÇÃO, D-22] (não exige consentimento: é o ponto de entrada do cliente)

- `Executar(ctx, nome, email string) (Cliente, error)`: `NovoCliente` (nome e e-mail obrigatórios, `ErrDadosInvalidos`), ID novo, Ativo, DataCadastro = agora; persiste com `ClienteRepository.Salvar`.
- O cliente nasce sem registro de consentimento (equivale a NAO_CONCEDIDO, D-02): nada além de nome e e-mail é tratado até UC-08 Conceder.

**UC-08 `GerenciarConsentimentoUseCase`** (não exige consentimento válido, pois é quem o gerencia)

- `Conceder(ctx, clienteID, finalidade, versaoTermo string) error`: cria o registro se não existir, aplica a transição 3.4.1, monta o movimento CONCESSAO (`NovoMovimento`) e persiste os dois com `ConsentimentoRepository.Registrar` (v1.4, RN-10; antes `Salvar`).
- `Revogar(ctx, clienteID) error`: idem, com o movimento REVOGACAO.
- `Consultar(ctx, clienteID) (ConsentimentoLGPD, error)` [ADIÇÃO]: sem registro, devolve um consentimento sintético em NAO_CONCEDIDO.
- `Historico(ctx, clienteID) ([]MovimentoConsentimento, error)` [ADIÇÃO, v1.4]: os movimentos em ordem cronológica (RN-10); vazio se não houver.
- Pré-condição: cliente existe (`ErrClienteNaoEncontrado`); em `Conceder` e `Revogar`, cliente não eliminado (`ErrClienteEliminado`, v1.9).
- **Origem (v1.4, Q-34):** as assinaturas não mudam. O adaptador põe a origem no `context.Context` com `usecase.ComOrigem(ctx, origem, operador)`; o caso de uso a lê ao montar o movimento. Sem valor no contexto: `sistema`. Origem `painel` sem operador: `ErrDadosInvalidos`, e nada é gravado.

**UC-01 `ColetarDadosUseCase`**

- `Executar(ctx, clienteID) ([]EventoComportamental, error)`: RN-01 e `EventoRepository.BuscarPorCliente` (sequência, passos 3 e 4).
- `Registrar(ctx, e EventoComportamental) (EventoComportamental, error)` [ADIÇÃO, D-12]: RN-01, `Validar` (RN-04), atribui ID se vazio e `EventoRepository.Salvar`; devolve o evento gravado (v1.1: a API responde 201 com ele).

**UC-02 `GerarPerfilUseCase`**

- `Executar(ctx, clienteID) (PerfilCliente, error)`: RN-01; obtém os eventos reutilizando `ColetarDadosUseCase.Executar` (D-07); chama `PerfilamentoService.GerarPerfil`; salva com `PerfilRepository.Salvar` (upsert) e devolve o perfil salvo.

**UC-03 `SegmentarClienteUseCase`**

- `Executar(ctx, clienteID) (Segmento, error)`: RN-01; `PerfilRepository.BuscarPorCliente` (`ErrPerfilNaoEncontrado`); `SegmentacaoService.Segmentar` (lê `SegmentoRepository.ListarAtivos`); atribui `SegmentoID` e salva o perfil.

**UC-04 `GerarRecomendacaoUseCase`**

- `Executar(ctx, clienteID) (RecomendacaoProduto, error)`: RN-01; carrega perfil e segmento (`PerfilRepository` e `SegmentoRepository.BuscarPorID`); perfil sem segmento gera `ErrPerfilNaoEncontrado`; chama `RecomendacaoService.Recomendar`, depois `Disponibilizar`, e salva uma única vez com `RecomendacaoRepository.Salvar`.

**Pré-condições RN-01** ficam em `Precondicoes.Verificar(ctx, clienteID) (Cliente, error)`, compartilhado por UC-01 a UC-04.

**`ExpirarRecomendacoesUseCase`** [ADIÇÃO, D-08]

- `Executar(ctx) (int, error)`: lê `RecomendacaoRepository.ListarAtivasVencidas(agora)`, aplica `Expirar` e salva cada uma; devolve quantas expirou. É o que o subcomando `expirar-recomendacoes` chama.

**`ProcessarClienteUseCase`** [ADIÇÃO, Q-26]

- `Executar(ctx, clienteID) (ResultadoProcessamento, error)`: RN-09. `ResultadoProcessamento` traz `Eventos` (quantos UC-01 coletou), `Perfil`, `Segmento`, `Recomendacao *RecomendacaoProduto` (nula sem produto elegível) e `Etapas`, as quatro na ordem (`coletar`, `perfil`, `segmento`, `recomendacao`), cada uma com `Estado` (`EstadoEtapa`: `concluida`, `atencao`, `falhou`, `pendente`) e o erro que a parou. O `error` devolvido é o da etapa que falhou (o mesmo de antes para a API).

**Casos de uso de leitura do painel** [ADIÇÃO, Q-21, Q-23]

- `ListarClientesUseCase.Executar(ctx, FiltroListagem) (PaginaClientes, error)`: `FiltroListagem` tem `Busca` (nome ou ID), `Consentimento` (`*StatusConsentimento`, opcional) e `Pagina` (a partir de 1; 20 por página). `PaginaClientes` traz `Linhas` (cliente, estado do consentimento e nome do segmento, este só com cliente ativo e consentimento válido), `Total` e `Pagina`. O filtro de consentimento é aplicado pelo caso de uso sobre a busca do repositório (volume do MVP).
- `ConsultarClienteUseCase.Executar(ctx, clienteID) (FichaCliente, error)`: cadastro, consentimento (sem registro: sintético em NAO_CONCEDIDO, como UC-08) e, desde a v1.4, `Historico` (os movimentos de UC-08 `Historico`, RN-10). Não exige consentimento (Q-21); `ErrClienteNaoEncontrado`.
- `ConsultarResultadoUseCase.Executar(ctx, clienteID) (ResultadoCliente, error)`: cliente existe e consentimento válido primeiro (`Precondicoes.VerificarLeitura`: RN-01 sem a exigência de cliente ativo, Q-31; cliente inativo é só leitura); depois eventos (mais recente primeiro), perfil e segmento (nulos se ainda não processado), recomendações (`ListarPorCliente`, mais recente primeiro) e os produtos que elas citam. Nada de comportamento é lido sem consentimento válido (C-03).
- `ListarSegmentosUseCase.Executar(ctx) ([]Segmento, error)`: segmentos ativos em ordem de Prioridade.
- `ListarCatalogoUseCase.Executar(ctx) ([]Produto, error)`: todos os produtos, por Categoria e depois por Nome.

**Casos de uso dos direitos do titular** [ADIÇÃO, v1.9, Q-48 a Q-52]

- `AtualizarCadastroUseCase.Executar(ctx, clienteID, nome, email string) (Cliente, error)`: cliente existe (`ErrClienteNaoEncontrado`); `AtualizarDados` (nome e e-mail obrigatórios, `ErrDadosInvalidos`; eliminado, `ErrClienteEliminado`); `ClienteRepository.Salvar`. Não exige consentimento nem cliente ativo (Q-48).
- `ExportarDadosTitularUseCase.Executar(ctx, clienteID) (DadosTitular, error)`: cliente existe; devolve `DadosTitular` com `GeradoEm` (agora), `Cliente`, `Consentimento` (sintético em NAO_CONCEDIDO sem registro), `Historico`, `Eventos` (Timestamp crescente), `Perfil *PerfilCliente` e `Segmento *Segmento` (nulos sem perfil ou sem segmento), `Recomendacoes` (DataGeracao decrescente) e `Produtos` (os citados pelas recomendações). Lê tudo com qualquer estado do consentimento, inclusive do cliente eliminado (Q-49).
- `EliminarDadosTitularUseCase.Executar(ctx, clienteID, confirmacao string) error`: RN-17.
- `AplicarRetencaoUseCase.Executar(ctx) (int, error)`: RN-18. `ConsentimentoRepository.ListarRevogadosAte(ctx, agora menos o prazo)` e, para cada cliente, as remoções do passo 4 da RN-17 (eventos, perfil e recomendações); devolve quantos clientes tiveram algum dado apagado.
- Desde a v1.10, `NovoEliminarDadosTitularUseCase` recebe também `TokenAcessoRepository` e `PedidoEliminacaoRepository` (passos 4 e 5 da RN-17).

**Casos de uso do portal do titular** [ADIÇÃO, v1.10, Q-55 a Q-59]

- `SolicitarAcessoTitularUseCase.Executar(ctx, email string) error`: RN-19, passos 1 e 2. Recebe no construtor o endereço base do link (`PORTAL_URL_BASE`). Só devolve erro com e-mail vazio ou sem `@` (`ErrDadosInvalidos`) ou falha de infraestrutura.
- `AutenticarTitularUseCase.Executar(ctx, token string) (uuid.UUID, error)`: RN-19, passo 3.
- `ConsultarMeusDadosUseCase.Executar(ctx, clienteID) (MeusDados, error)`: cliente existe (`ErrClienteNaoEncontrado`) e não eliminado (`ErrClienteEliminado`); `MeusDados` traz `Cliente`, `Consentimento` (sintético em NAO_CONCEDIDO sem registro), `Historico` (RN-10) e `PedidoAberto *PedidoEliminacao` (nulo sem pedido aberto). Não exige consentimento (Q-21, C-03).
- `PedirEliminacaoUseCase.Executar(ctx, clienteID) (PedidoEliminacao, error)`: RN-20.
- `ListarPedidosEliminacaoUseCase.Executar(ctx) ([]LinhaPedido, error)`: os pedidos, abertos primeiro e depois por DataPedido decrescente, cada um com o cliente e o estado atual do consentimento (`LinhaPedido{Pedido, Cliente, Consentimento}`), para a operação abrir a ficha.
- `ListarCaixaSaidaUseCase.Executar(ctx) ([]MensagemSimulada, error)`: as mensagens das últimas 24 horas, da mais recente para a mais antiga.
- Conceder, revogar e baixar os dados no portal usam UC-08 e `ExportarDadosTitularUseCase`, com `usecase.ComOrigem(ctx, portal, "")` (RN-10).

Serviços de domínio calculam e só leem ports quando precisam (segmentos e produtos). Quem persiste resultados são os casos de uso (D-09).

### 3.6 Ports e serviços (assinaturas)

```go
// internal/domain/port — todo método de I/O recebe ctx como primeiro parâmetro
type ClienteRepository interface {
    Salvar(ctx context.Context, c entity.Cliente) error
    BuscarPorID(ctx context.Context, id uuid.UUID) (entity.Cliente, error) // ErrClienteNaoEncontrado
    Listar(ctx context.Context, f FiltroClientes) ([]entity.Cliente, int, error) // [ADIÇÃO] Q-23: página e total
    BuscarPorEmail(ctx context.Context, email string) ([]entity.Cliente, error) // [ADIÇÃO] v1.10 Q-59: sem diferenciar maiúsculas e espaços nas pontas; cadastro mais recente primeiro; vazio sem erro
}
type FiltroClientes struct { // [ADIÇÃO] Q-23
    Busca        string // vazio = todos; UUID = ID exato; senão, nome contém (sem diferenciar maiúsculas e acentos)
    Deslocamento int
    Limite       int    // 0 = sem limite
} // ordem: DataCadastro decrescente, desempate por ID
type EventoRepository interface {
    Salvar(ctx context.Context, e entity.EventoComportamental) error
    BuscarPorCliente(ctx context.Context, clienteID uuid.UUID) ([]entity.EventoComportamental, error)
    RemoverPorCliente(ctx context.Context, clienteID uuid.UUID) (int, error) // [ADIÇÃO] v1.9 Q-52: quantos removeu; 0 sem erro
}
type PerfilRepository interface {
    Salvar(ctx context.Context, p entity.PerfilCliente) (entity.PerfilCliente, error) // upsert por ClienteID
    BuscarPorCliente(ctx context.Context, clienteID uuid.UUID) (entity.PerfilCliente, error) // ErrPerfilNaoEncontrado
    RemoverPorCliente(ctx context.Context, clienteID uuid.UUID) (int, error) // [ADIÇÃO] v1.9 Q-52: 0 ou 1
}
type SegmentoRepository interface {
    ListarAtivos(ctx context.Context) ([]entity.Segmento, error)            // ordem crescente de Prioridade
    BuscarPorID(ctx context.Context, id uuid.UUID) (entity.Segmento, error) // [ADIÇÃO] ErrDadosInvalidos se inexistente
}
type RecomendacaoRepository interface {
    Salvar(ctx context.Context, r entity.RecomendacaoProduto) error // insere ou atualiza por ID
    ListarAtivasVencidas(ctx context.Context, agora time.Time) ([]entity.RecomendacaoProduto, error) // [ADIÇÃO]
    ListarAtivasPorCliente(ctx context.Context, clienteID uuid.UUID) ([]entity.RecomendacaoProduto, error) // [ADIÇÃO] Q-18
    ListarPorCliente(ctx context.Context, clienteID uuid.UUID) ([]entity.RecomendacaoProduto, error) // [ADIÇÃO] Q-23: DataGeracao decrescente
    RemoverPorCliente(ctx context.Context, clienteID uuid.UUID) (int, error) // [ADIÇÃO] v1.9 Q-52: quantas removeu
}
type ConsentimentoRepository interface {
    Salvar(ctx context.Context, c entity.ConsentimentoLGPD) error // upsert por ClienteID, sem movimento (cargas e testes)
    BuscarPorCliente(ctx context.Context, clienteID uuid.UUID) (entity.ConsentimentoLGPD, error) // ErrConsentimentoNaoEncontrado
    Registrar(ctx context.Context, c entity.ConsentimentoLGPD, m entity.MovimentoConsentimento) error // [ADIÇÃO] v1.4 RN-10: upsert + inclusão do movimento, atômico; ID de movimento repetido falha sem gravar nada
    ListarMovimentos(ctx context.Context, clienteID uuid.UUID) ([]entity.MovimentoConsentimento, error) // [ADIÇÃO] v1.4: Data crescente, desempate pela ordem de gravação
    ListarRevogadosAte(ctx context.Context, limite time.Time) ([]uuid.UUID, error) // [ADIÇÃO] v1.9 Q-51: clientes REVOGADO com DataRevogacao <= limite, da revogação mais antiga para a mais recente
}
type ProdutoRepository interface { // [ADIÇÃO] (D-05)
    ListarAtivosPorCategoria(ctx context.Context, categoria string) ([]entity.Produto, error)
    ListarTodos(ctx context.Context) ([]entity.Produto, error) // [ADIÇÃO] Q-23: por Categoria e Nome
}
type ConsentimentoGateway interface { // MVP: implementação local sobre ConsentimentoRepository
    Validar(ctx context.Context, clienteID uuid.UUID) (bool, error)
}
type IntegracaoGateway interface { // e-commerce/CRM; MVP: stub que devolve lista vazia (D-12)
    BuscarEventos(ctx context.Context, clienteID uuid.UUID, desde time.Time) ([]entity.EventoComportamental, error)
}
type Clock interface{ Agora() time.Time }

// [ADIÇÃO] v1.10 — portal do titular (Q-56 a Q-59)
type TokenAcessoRepository interface {
    Salvar(ctx context.Context, t entity.TokenAcessoTitular) error
    Consumir(ctx context.Context, hash string, agora time.Time) (entity.TokenAcessoTitular, error) // marca UsadoEm de forma atômica; inexistente, expirado ou usado: ErrAcessoInvalido
    ContarDesde(ctx context.Context, clienteID uuid.UUID, desde time.Time) (int, error)            // tokens do cliente com CriadoEm >= desde
    RemoverPorCliente(ctx context.Context, clienteID uuid.UUID) (int, error)                       // RN-17
}
type PedidoEliminacaoRepository interface {
    Salvar(ctx context.Context, p entity.PedidoEliminacao) error // insere ou atualiza por ID; um ABERTO por cliente
    BuscarAbertoPorCliente(ctx context.Context, clienteID uuid.UUID) (entity.PedidoEliminacao, error) // ErrPedidoNaoEncontrado
    Listar(ctx context.Context) ([]entity.PedidoEliminacao, error) // ABERTO primeiro, depois DataPedido decrescente
}
type Notificador interface { // MVP: NotificadorSimulado (caixa de saída simulada, Q-57)
    EnviarLinkAcesso(ctx context.Context, para, link string, expiraEm, dataCadastro time.Time) error
}
type CaixaSaidaSimulada interface {
    Guardar(ctx context.Context, m entity.MensagemSimulada) error
    Listar(ctx context.Context, desde time.Time) ([]entity.MensagemSimulada, error) // EnviadaEm >= desde, mais recente primeiro
    RemoverAte(ctx context.Context, limite time.Time) (int, error)                   // EnviadaEm < limite
}

// internal/domain/service
func (s *PerfilamentoService) GerarPerfil(clienteID uuid.UUID, eventos []entity.EventoComportamental) entity.PerfilCliente
func (s *SegmentacaoService) Segmentar(ctx context.Context, p entity.PerfilCliente) (entity.Segmento, error)
func (s *RecomendacaoService) Recomendar(ctx context.Context, c entity.Cliente, p entity.PerfilCliente, seg entity.Segmento) (entity.RecomendacaoProduto, error)
func (v *ValidadorLGPD) ConsentimentoValido(ctx context.Context, clienteID uuid.UUID) (bool, error)
```

`ClienteRepository.Salvar` passa a gravar `DataEliminacao` (v1.9); `BuscarPorID` e `Listar` a devolvem. Os movimentos do consentimento continuam só de inclusão: nenhuma remoção os atinge (RN-10, RN-17).

`Recomendar` recebe também o perfil `p`, porque RN-07 precisa da categoria preferida; o diagrama de classes traz apenas cliente e segmento. `PerfilamentoService` é puro (usa só o `Clock` injetado no construtor), por isso não recebe `ctx` nem devolve `error`.

### 3.7 API HTTP

Proposta da equipe: os diagramas citam handlers, rotas e DTOs, mas não as definem. Roteamento com `net/http` (padrões com método e parâmetros de caminho, disponíveis a partir do Go 1.22).

| Método e rota | Caso de uso | Entrada | Sucesso | Middlewares |
|---|---|---|---|---|
| `POST /v1/clientes` | UC-09 Cadastrar | nome e email | 201 com id, nome, email, ativo e dataCadastro | Autenticação |
| `POST /v1/clientes/{clienteID}/processamento` | UC-01 a UC-04 (RN-09) | — | 200 com perfil, segmento e recomendacao (nula se `ErrSemProdutoElegivel`) | Autenticação e LGPD |
| `POST /v1/clientes/{clienteID}/eventos` | UC-01 Registrar | Evento: tipo, origem, timestamp, valor, categoria | 201 | Autenticação e LGPD |
| `PUT /v1/clientes/{clienteID}/consentimento` | UC-08 Conceder | finalidade e versaoTermo | 200 | Autenticação |
| `POST /v1/clientes/{clienteID}/consentimento/revogacao` | UC-08 Revogar | — | 200 | Autenticação |
| `GET /v1/clientes/{clienteID}/consentimento` | UC-08 Consultar | — | 200 com status e valido | Autenticação |
| `GET /saude` | — | — | 200 com `status: ok`; 503 com `status: indisponivel` se o banco não responder (Q-19) | Nenhum (usado pelo balanceador) |

- Autenticação por cabeçalho `Authorization: Bearer` com o valor de `AUTH_TOKEN` (D-16); ausente ou inválido responde 401.
- Middleware LGPD: extrai o clienteID da rota e consulta `ValidadorLGPD`; sem consentimento válido responde 403 antes do handler. Os casos de uso repetem a checagem (C-03).
- Conceder e revogar pela API gravam o movimento com origem `api` (RN-10); as respostas não mudam.
- Erros: corpo JSON com as chaves `codigo` e `mensagem`, conforme a tabela da seção 3.5. Erros inesperados viram 500 sem detalhes internos.
- Corpo de requisição: um único objeto JSON, sem campos desconhecidos e sem conteúdo depois dele; caso contrário, `dados_invalidos` (Q-19). Uma resposta que não puder ser serializada vira 500, nunca 200 com corpo vazio (Q-15).
- Campos JSON em camelCase, com os mesmos nomes dos campos Go.

### 3.7.1 Painel web do operador [ADIÇÃO, v1.3]

Adaptador de apresentação HTML em `internal/web` (irmão de `internal/api`, Q-20), montado no mesmo `omega-server` sob `/painel`. HTML gerado no servidor com `html/template`; htmx 2.0.3, fontes IBM Plex e CSS servidos do próprio binário. Telas, fluxos e textos: `docs/design/02-telas-e-fluxos.md` e `03-textos.md`.

| Método e rota | Tela ou ação | Caso de uso |
|---|---|---|
| `GET /painel/entrar`, `POST /painel/entrar` | Entrar | — (Q-22) |
| `POST /painel/sair` | Sair | — |
| `GET /painel/clientes?q=&consentimento=&pagina=` | Lista e busca | `ListarClientes` |
| `GET /painel/clientes/novo`, `POST /painel/clientes` | Cadastrar cliente | UC-09 |
| `GET /painel/clientes/{id}` | Ficha: visão geral | `ConsultarCliente` e, com consentimento válido, `ConsultarResultado` |
| `POST /painel/clientes/{id}/processamento` | Processar agora | `ProcessarCliente` |
| `GET /painel/clientes/{id}/eventos`, `POST /painel/clientes/{id}/eventos` | Ficha: eventos | `ConsultarResultado`, UC-01 Registrar |
| `GET /painel/clientes/{id}/consentimento` | Ficha: consentimento | `ConsultarCliente` |
| `GET` e `POST /painel/clientes/{id}/consentimento/conceder` | Confirmar e conceder | UC-08 Conceder |
| `GET` e `POST /painel/clientes/{id}/consentimento/revogar` | Confirmar e revogar | UC-08 Revogar |
| `GET /painel/clientes/{id}/direitos` | Ficha: direitos do titular [v1.9] | `ConsultarCliente` |
| `POST /painel/clientes/{id}/cadastro` | Corrigir cadastro [v1.9] | `AtualizarCadastro` |
| `GET /painel/clientes/{id}/dados` | Baixar dados (JSON) [v1.9] | `ExportarDadosTitular` |
| `GET` e `POST /painel/clientes/{id}/eliminar` | Confirmar e eliminar [v1.9] | `EliminarDadosTitular` |
| `GET /painel/segmentos`, `GET /painel/catalogo` | Configuração (somente leitura) | `ListarSegmentos`, `ListarCatalogo` |
| `GET /painel/operacao`, `GET` e `POST /painel/operacao/expirar` | Saúde e expiração | `ExpirarRecomendacoes` |
| `GET` e `POST /painel/operacao/retencao` | Aplicar retenção [v1.9] | `AplicarRetencao` |
| `GET /painel/pedidos` | Pedidos do titular [v1.10] | `ListarPedidosEliminacao` |
| `GET /painel/operacao/caixa-saida` | Caixa de saída (simulada) [v1.10] | `ListarCaixaSaida` |
| `GET /painel/static/{arquivo}` | CSS, JS, fontes | — |

- **Sessão (Q-22):** cookie `perfila_sessao` com `operador|expiraEm|csrf` assinado por HMAC-SHA256 (`PAINEL_CHAVE_SESSAO`), `HttpOnly`, `Secure`, `SameSite=Strict`, `Path=/painel`, 8 h renovadas a cada uso; sem estado no servidor. Sem sessão válida, toda rota (menos Entrar e estáticos) leva a Entrar com o destino guardado; sessão vencida mostra "Sua sessão expirou." Em requisição htmx, a resposta é `HX-Redirect`.
- **Entrar:** senha conferida com bcrypt; erro genérico "Usuário ou senha incorretos."; 5 tentativas por IP em 5 minutos (em memória, por réplica), depois "Muitas tentativas. Aguarde 5 minutos." (429).
- **CSRF:** todo POST confere o token da sessão, vindo do campo `_csrf` ou do cabeçalho `X-CSRF-Token`; ausente ou diferente responde 403.
- **Cabeçalhos em toda resposta:** `Content-Security-Policy: default-src 'self'; img-src 'self' data:; style-src 'self'; script-src 'self'; frame-ancestors 'none'; form-action 'self'; base-uri 'none'`, `X-Content-Type-Options: nosniff`, `Referrer-Policy: same-origin`, `X-Frame-Options: DENY`; `Cache-Control: no-store` nas páginas; estáticos com `max-age` de um ano e versão na URL.
- **Consentimento (C-03, Q-21):** sem consentimento válido, a ficha mostra só cadastro e consentimento, com o aviso de bloqueio; nenhum caso de uso de comportamento é chamado para montar a tela.
- **Histórico do consentimento (v1.4, Q-34):** conceder e revogar pelo painel gravam o movimento com origem `painel` e o login da sessão. A linha do tempo da aba Consentimento mostra o cadastro e cada movimento, do mais recente para o mais antigo, com finalidade, versão do termo e origem ("pela API", "pelo painel, por {operador}", "pelo portal do titular", "pelo sistema", "reconstruído das datas"). Cliente sem nenhum movimento continua com a linha montada pelas datas do registro, com o aviso de que ela não é o histórico completo.
- **Direitos do titular (v1.9, Q-48 a Q-53):** aba "Direitos do titular" na ficha, com três painéis.
    - *Corrigir cadastro*: nome e e-mail, com qualquer estado do consentimento.
    - *Acesso e portabilidade*: "Baixar dados (JSON)" entrega `perfila-dados-{id}.json` como anexo (`Content-Disposition: attachment`, `application/json`, `Cache-Control: no-store`), com `versaoFormato` "1", `geradoEm` e as seções de `DadosTitular` em camelCase.
    - *Eliminação*: com consentimento CONCEDIDO, só o aviso "Revogue o consentimento antes de eliminar os dados." (a ação de consentimento fica na aba Consentimento, Q-27). Senão, "Eliminar dados" leva à confirmação, que explica o efeito, pede o nome do cliente digitado e tem o foco inicial em Cancelar.
    - Cliente eliminado: a ficha mostra o selo "Eliminado", o aviso com a data e só a linha do tempo do consentimento; nenhuma ação é oferecida.
    - A Operação ganha "Aplicar retenção", com confirmação que cita o prazo.
    - Cada pedido atendido gera o log `direito_atendido` (Q-53).
- **Pedidos do titular e caixa de saída (v1.10, Q-57, Q-58):** "Pedidos do titular" (grupo Clientes) lista os pedidos de eliminação feitos pelo portal, abertos primeiro, com cliente, data, situação e o consentimento atual; cada linha leva à aba Direitos da ficha, que mostra o aviso "O titular pediu a eliminação pelo portal em {data}." enquanto o pedido estiver aberto. A eliminação é a mesma da v1.9 e marca o pedido como atendido. "Caixa de saída (simulada)" (grupo Operação) mostra as mensagens das últimas 24 horas (destinatário mascarado, assunto, texto com o link, data). Só aparece com o portal ligado; hoje é da pessoa operadora única (Q-22) e passa a exigir ADMINISTRADOR quando os papéis da v1.5 entrarem na main.
- **Sem JavaScript:** formulários comuns com redirecionamento 303 no sucesso (PRG) e confirmação em página própria. Com JavaScript, o htmx (`hx-boost`) faz as mesmas requisições e troca só a área do painel (`#app`), sem recarregar a página; os diálogos usam `<dialog>` nativo. Respostas de erro (4xx e 5xx) também são exibidas.
- **Erros:** o código estável (`errs.Codigo`) escolhe a mensagem de `docs/design/03-textos.md`; o texto técnico nunca aparece. Falha inesperada mostra o código da requisição (`X-Request-ID` do Nginx).
- **Logs (C-09, Q-24):** método, caminho (só UUIDs) e status; nas ações, operador, ação e `clienteId`. Nunca nome, e-mail, senha nem o termo de busca. Desde a v1.9, `direito_atendido` com operador, pedido (`correcao`, `exportacao`, `eliminacao`, `retencao`) e `clienteId` (ou a quantidade, na retenção). É a auditoria dos pedidos até a v1.5 (`RegistroAuditoria`) entrar na main; aí passam a ser gravados nela e a eliminação fica restrita a ADMINISTRADOR (Q-53).

### 3.7.2 Portal do titular [ADIÇÃO, v1.10]

Segundo adaptador HTML de `internal/web` (Q-55), montado no mesmo `omega-server` sob `/titular`, com roteador, sessão e templates próprios; reaproveita o CSS, as fontes e o htmx do painel, servidos também em `/titular/static/`. Nenhuma rota do portal recebe ID de cliente: o cliente é sempre o da sessão. Telas e textos: `docs/design/02-telas-e-fluxos.md` e `03-textos.md` (seção Portal do titular).

| Método e rota | Tela ou ação | Caso de uso |
|---|---|---|
| `GET /titular` | Leva a Meus dados (com sessão) ou a Entrar | — |
| `GET /titular/entrar`, `POST /titular/entrar` | Pedir o link de acesso | `SolicitarAcessoTitular` |
| `GET /titular/acesso?t=`, `POST /titular/acesso` | Abrir o link (não consome) e entrar (consome) | `AutenticarTitular` |
| `POST /titular/sair` | Sair | — |
| `GET /titular/meus-dados` | Cadastro, autorização e histórico | `ConsultarMeusDados` |
| `GET` e `POST /titular/revogar` | Confirmar e retirar a autorização | UC-08 Revogar |
| `GET` e `POST /titular/conceder` | Confirmar e autorizar com o termo vigente | UC-08 Conceder |
| `GET /titular/dados` | Baixar meus dados (JSON) | `ExportarDadosTitular` |
| `GET` e `POST /titular/eliminacao` | Confirmar e pedir a eliminação | `PedirEliminacao` |
| `GET /titular/static/{arquivo}` | CSS, JS, fontes | — |

- **Sessão (Q-55):** cookie `perfila_titular` com `clienteId|expiraEm|csrf` assinado por HMAC-SHA256 (`PORTAL_CHAVE_SESSAO`), `HttpOnly`, `Secure`, `SameSite=Strict`, `Path=/titular`, 30 minutos renovados a cada uso; sem estado no servidor. Cada requisição com sessão chama `ConsultarMeusDados`: cliente eliminado ou inexistente encerra a sessão e leva a Entrar. Sem sessão, as rotas de Meus dados levam a Entrar; sessão vencida mostra "Sua sessão terminou. Peça um novo link para continuar."
- **Entrar (RN-19):** formulário com o e-mail. Depois do POST, sempre o mesmo destino (`/titular/entrar?enviado=1`, PRG) e o mesmo texto: "Se o e-mail estiver cadastrado, enviamos um link de acesso. Ele vale por 30 minutos e só pode ser usado uma vez." E-mail vazio ou sem `@`: erro ao lado do campo (422). 10 pedidos por IP em 15 minutos (memória, por réplica): 429 "Muitas tentativas. Aguarde 15 minutos."
- **Link (RN-19, Q-56):** `GET /titular/acesso?t=` só mostra o botão "Entrar no portal" (sem consumir); o POST consome e, com sucesso, grava a sessão e leva a Meus dados (303). Link inválido: 401 com "Este link não vale mais. Peça um novo." e o caminho para Entrar.
- **CSRF:** os POST com sessão conferem o token da sessão (`_csrf` ou `X-CSRF-Token`); os POST de Entrar e do link conferem o token de um cookie próprio, `perfila_titular_entrar`. Ausente ou diferente: 403.
- **Cabeçalhos:** os mesmos do painel (3.7.1), com `Cache-Control: no-store` nas páginas e `Referrer-Policy: same-origin`, que não deixa o token sair na navegação.
- **Meus dados:** cadastro (nome e e-mail completos: é o próprio titular), o estado da autorização com finalidade, termo e datas, e a linha do tempo do consentimento (como a do painel, com a origem "por você, no portal" no lugar de "pelo portal do titular"). Uma ação de consentimento por tela (Q-27): "Retirar autorização" com CONCEDIDO; "Autorizar de novo" com NAO_CONCEDIDO ou REVOGADO, se não houver pedido de eliminação aberto. "Baixar meus dados (JSON)" sempre; "Pedir eliminação" sem consentimento CONCEDIDO e sem pedido aberto; com CONCEDIDO, o aviso "Para pedir a eliminação, retire a autorização antes."; com pedido aberto, "Você pediu a eliminação em {data}. A equipe vai atender o pedido."
- **Revogar e conceder:** confirmação em página própria (sem JavaScript) ou em `<dialog>`, com o texto de `03-textos.md`; revogar com o foco inicial em Cancelar. Movimento com origem `portal` (RN-10). Conceder usa `PAINEL_FINALIDADE` e `PAINEL_VERSAO_TERMO`, que o titular não edita.
- **Baixar meus dados:** o mesmo arquivo da v1.9 (`perfila-dados-{id}.json`, anexo, `versaoFormato` "1").
- **Logs (C-09):** `portal_http` com método, caminho e status (nunca a query string, que leva o token) e `portal_acao` com a ação (`link_pedido`, `entrada`, `saida`, `revogacao`, `concessao`, `exportacao`, `eliminacao_pedida`) e o `clienteId` quando houver sessão. `link_pedido` não traz `clienteId` nem e-mail. Nunca nome, e-mail, token ou link.

### 3.8 Configuração, estrutura de pastas e persistência

| Variável | Padrão | Uso |
|---|---|---|
| `HTTP_PORT` | 8080 | Porta do servidor; o balanceador distribui para ela |
| `DATABASE_URL` | vazia = modo memória (D-23) | Conexão com o PostgreSQL (porta 5432, banco omega_db). Preenchida: o servidor aplica as migrações ao subir e usa os `Pg*Repository`. A implantação (UML8) sempre a define; vazia só em desenvolvimento, BDD e demo local (Q-07) |
| `AUTH_TOKEN` | obrigatória | Token Bearer (D-16) |
| `RECOMENDACAO_VALIDADE_DIAS` | 30 | RN-08 |
| `ENGAJAMENTO_META_EVENTOS` | 30 | RN-05 |
| `LOG_LEVEL` | info | Nível de log |
| `PAINEL_USUARIO` | operador | Login do operador do painel (Q-22) |
| `PAINEL_SENHA_HASH` | vazia = painel desligado (Q-29) | Hash bcrypt da senha do operador (`omega-server hash-senha`, Q-28) |
| `PAINEL_CHAVE_SESSAO` | vazia = painel desligado (Q-29) | Chave HMAC da sessão: 64 dígitos hexadecimais (32 bytes) |
| `PAINEL_FINALIDADE` | segmentacao | Finalidade sugerida ao conceder (Q-25) |
| `PAINEL_VERSAO_TERMO` | v1 | Versão vigente do termo (Q-25) |
| `RETENCAO_DIAS_REVOGADO` | 180 | Prazo da RN-18 em dias, inteiro maior que zero (Q-51) |
| `PORTAL_URL_BASE` | vazia = portal desligado (Q-60) | Endereço público do portal usado no link (`https://…`, sem barra no fim); nunca o cabeçalho `Host` |
| `PORTAL_CHAVE_SESSAO` | vazia = portal desligado (Q-60) | Chave HMAC da sessão do titular: 64 dígitos hexadecimais, diferente da do painel |

As variáveis são lidas de `config.env` ao subir o contêiner; variável obrigatória ausente aborta a inicialização com mensagem clara.

```text
perfila/                          módulo Go `perfila`
  CLAUDE.md                       constituição (C-01..C-12) e mapa de contexto
  cmd/omega-server/main.go        [Support] raiz de composição; subcomandos serve, expirar-recomendacoes, hash-senha, carregar-exemplos e aplicar-retencao (v1.9)
  cmd/perfila-demo/main.go        [Core] demonstração do fluxo em memória
  internal/
    api/                          [Support] handlers, rotas, middlewares, DTOs
    web/                          [v1.3] painel do operador: rotas /painel, sessão, CSRF, templates e static embutidos; [v1.10] portal do titular (/titular), com sessão própria
    application/usecase/          [Core] os cinco casos de uso + ExpirarRecomendacoes + Precondicoes; [v1.3] ProcessarCliente e leituras do painel
    domain/entity/                [Core] entidades e enums (importa só uuid e shared/errs)
    domain/port/                  [Core] repositórios, gateways e Clock
    domain/service/               [Core] PerfilamentoService, SegmentacaoService, RecomendacaoService, ValidadorLGPD
    infra/memoria/                [Core] fakes em memória dos ports (BDD, contrato, demo)
    infra/relogio/                [Core] Clock real e Clock fixo (único lugar com time.Now)
    infra/repository/pg/          [Generic] Pg*Repository e migrações
    infra/external/               [Support] gateways (consentimento local, integração stub)
    infra/config/                 [Support] leitura de config.env
    shared/errs/                  [Core] erros de domínio (sentinelas com código)
    arquitetura_test.go           [Core] teste de arquitetura (C-01, C-05, C-07, C-08)
  features/                       arquivos .feature, steps/ e runner godog
  specs/                          spec, plano, tarefas, regras BDD
  docs/uml/                       UML1..UML8 em texto + imagens originais
  deploy/                         [Generic] Dockerfile, docker-compose.yml, nginx.conf
```

Persistência e implantação:

- PostgreSQL via `database/sql` com o driver `pgx/v5/stdlib` (Q-10); migrações em SQL versionado, embutidas no binário e aplicadas pelo `omega-server` ao subir, sob `pg_advisory_lock` e registradas em `schema_migracoes` (Q-09); uma tabela por entidade (clientes, eventos_comportamentais, perfis_cliente, segmentos, recomendacoes_produto, consentimentos_lgpd, produtos e, desde a v1.4, movimentos_consentimento), em snake_case.
- Chaves UUID; timestamps em `timestamptz` (UTC); `UNIQUE(cliente_id)` em perfis_cliente e consentimentos_lgpd; `CHECK` nos enums de status; chaves estrangeiras para clientes. Reforços de integridade que só repetem regras já existentes: chaves estrangeiras `perfis_cliente.segmento_id → segmentos` (D-06) e `recomendacoes_produto.produto_id → produtos` (D-05); `CHECK` de valor ≥ 0 nos eventos (RN-04), score entre 0 e 1 (RN-05) e preço ≥ 0 (`Produto.AtualizarPreco`).
- Clientes de exemplo (Q-30, Q-32): quinze clientes sintéticos com história (o 15º, desde a v1.9, revogou há mais de 180 dias para demonstrar a retenção, RN-18) (datas de cadastro, consentimento, revogação, eventos e recomendações ao longo de meses), criados pelos casos de uso com um relógio de carga; no modo memória ao subir, no PostgreSQL por `carregar-exemplos`, e só com o banco sem clientes.
- A migração 010 (v1.10, Q-59) cria `tokens_acesso_titular` (hash único, chave estrangeira para clientes), `pedidos_eliminacao` (`CHECK` no status, um ABERTO por cliente por índice único parcial) e `caixa_saida_simulada`, e um índice em `lower(clientes.email)` para `BuscarPorEmail`. A suíte de contrato cobre os três ports novos e `BuscarPorEmail`.
- A migração 009 (v1.9, Q-50) acrescenta `clientes.data_eliminacao` (nula). Recebe o número 009 porque 007 e 008 estão nas branches abertas da v1.5 e da v1.6; o migrador aplica cada versão pelo nome, seja qual for a ordem de chegada (Q-09).
- Subcomando `aplicar-retencao` (v1.9, RN-18): o administrador o executa como `expirar-recomendacoes`; o log `retencao_aplicada` traz só a quantidade.
- A migração 006 (v1.4, Q-34) cria `movimentos_consentimento` (chave UUID, coluna de ordem de gravação, `CHECK` em tipo e origem, operador obrigatório com origem painel, chave estrangeira para clientes) com um gatilho que recusa `UPDATE` e `DELETE` (só inclusão, RN-10), e reconstrói os movimentos dos clientes antigos a partir das datas que o registro prova: a última concessão (com finalidade e termo) e a última revogação (com finalidade e termo só se o estado ainda for REVOGADO; senão, em branco), com origem `migracao`.
- A migração 005 (Q-32) amplia a demonstração: segmentos vip, fiel, ocasional e explorador e catálogo de 7 categorias, iguais aos do modo memória.
- A migração 004 (v1.3) cria os índices da listagem de clientes: `data_cadastro DESC` e o nome normalizado (minúsculas e sem acentos), Q-23.
- A migração inicial cria o segmento padrao (critério vazio, Prioridade 99). Uma segunda migração carrega os segmentos e o catálogo sintéticos de demonstração, com os mesmos IDs do modo memória (Q-08). Dados de teste são sintéticos.
- Uma suíte de contrato (`internal/infra/repository/contrato`) define o comportamento observável dos sete repositórios e roda contra os fakes em memória e contra o PostgreSQL (`PERFILA_TEST_DATABASE_URL`; sem ela, a parte PostgreSQL é pulada).
- Binário estático `omega-server` em contêiner Docker com `config.env`; duas réplicas atrás de Nginx (TLS e proxy para a porta 8080); PostgreSQL em servidor próprio, nunca acessível do dispositivo do cliente.
- Implantação (T-24): `deploy/docker-compose.yml` com `nginx`, `app1`, `app2` e `db`. Redes `aplicacao` (Nginx e réplicas) e `dados` (réplicas e banco, interna, sem porta publicada). Só o Nginx publica porta, a HTTPS (Q-12); ele identifica a réplica no cabeçalho `X-Omega-Replica` (Q-11), tira réplicas paradas do balanceamento e registra logs sem IP nem query string (Q-14). O administrador executa os subcomandos com `docker compose exec` (equivalente local do SSH do UML8). Segredos gerados por script, fora do git (Q-14). Cenários `@implantacao` rodam fora do `make verify` (Q-13).

### 3.9 Lacunas e inconsistências da modelagem (decisões)

Prioridade Alta: impede gerar código correto. Média: gera comportamento ambíguo. Baixa: ajuste de limpeza. **Status:** todas as decisões abaixo foram adotadas como finais em 07/11/2026 (T-04); as perguntas Q-01 a Q-06 também foram decididas (`OPEN_QUESTIONS.md`). Mudança futura entra primeiro aqui.

| ID | Onde | Problema | Decisão padrão | Prior. |
|---|---|---|---|---|
| D-01 | Sequência | Nenhum passo verifica consentimento, embora o projeto declare que sem consentimento nada é processado e exista middleware LGPD | Verificação obrigatória em UC-01 a UC-04 (RN-01, C-03) e no middleware | Alta |
| D-02 | Classe ConsentimentoLGPD | `concedido: bool` não distingue NAO_CONCEDIDO de REVOGADO, que o diagrama de estados exige | Campo `Status` enumerado; ausência de registro equivale a NAO_CONCEDIDO | Alta |
| D-03 | Estados do consentimento | Transição de REVOGADO para o estado final rotulada `EstaValido()` com retorno true, mas `EstaValido()` é consulta e REVOGADO volta a CONCEDIDO | Sem estado final; `EstaValido()` é Status igual a CONCEDIDO | Alta |
| D-04 | Casos de uso | UC-05, UC-06 e UC-07 não aparecem em classes, componentes nem sequência | Fora do MVP; modelar antes de implementar | Alta |
| D-05 | Classes e componentes | Não há catálogo de produtos (nenhum `ProdutoRepository`) e `EventoComportamental` não tem categoria; categoriaPreferida e recomendação ficam inviáveis | [ADIÇÃO] `ProdutoRepository` e `Evento.Categoria`; seleção provisória RN-07 | Alta |
| D-06 | Classes | Atribuição de segmento sem persistência; relação Segmento e PerfilCliente desenhada 1 para 1 (deveria ser N para 1); critério é texto livre; faltam prioridade e flag ativo apesar de `ListarAtivos` | `PerfilCliente.SegmentoID`; relação N para 1; `Prioridade` e `Ativo`; gramática da RN-06 | Média |
| D-07 | Sequência | A sequência mostra Coletar chamando GerarPerfil (passo 5), mas o «include» aponta do caso base para o incluído e o perfil depende da coleta. **Correção v1.0:** o diagrama de casos de uso (UML1) está certo (Gerar Perfil «include» Coletar; Segmentar «include» Gerar Perfil); só a sequência inverte. Segmentar não carrega o perfil na sequência | GerarPerfil reutiliza Coletar; Segmentar carrega o perfil; o comportamento observável é o da sequência | Média |
| D-08 | Estados da recomendação | Atributos do diagrama de estados (status, ativação, validade, expiração) e a transição Disponibilizar não existem na classe; ninguém dispara `Expirar()`; `notificarCliente()` não tem port | Atributos e métodos adicionados; expiração por comando; notificação reduzida a log | Média |
| D-09 | Sequência | No desenho, as ativações de SegmentacaoService e RecomendacaoService aparecem sobre a linha de vida do PerfilamentoService; os serviços aparecem salvando (passos 8 e 20); os quatro repositórios parecem dividir duas linhas de vida | Nomes e responsabilidades vêm de classes e componentes; casos de uso persistem, serviços calculam | Média |
| D-10 | Pacotes e classes | Setas de domain para service e de domain para repository invertem a regra de dependência e, em Go, gerariam ciclo de importação. O Diagrama de Classes põe os serviços na camada "Aplicação" e as interfaces de repositório na "Infraestrutura"; o de Pacotes lista "interfaces" em application e em domain | O Diagrama de Componentes prevalece: serviços em `domain/service`, ports em `domain/port`, infra implementa os ports | Média |
| D-11 | Snippets dos Mód. 2 e 3 | `Cliente` com Historico e ConsentimentoLGPD bool, ID string e retorno de texto fixo divergem do modelo de classes | Ignorar os snippets; a spec manda | Baixa |
| D-12 | Componentes e sequência | Nenhum diagrama mostra como os eventos entram; `IntegracaoGateway` não tem métodos | `ColetarDadosUseCase.Registrar` e `POST .../eventos`; gateway como stub | Média |
| D-13 | LGPD | Retenção e eliminação de dados após revogação não estão modeladas | Fora do MVP; registrar como requisito futuro. **Substituída na v1.9** por RN-17 e RN-18 (Q-50, Q-51) | Baixa |
| D-14 | Classes | Valores monetários em float64 (preco, ticketMedio, valor) | Manter float64 no MVP; avaliar centavos inteiros depois | Baixa |
| D-15 | Classes e atores | A relação “EventoComportamental recebe RecomendacaoProduto” provavelmente é Cliente 1 para 0..*; o ator Administrador (PDF antigo e implantação) convive com o Agente Inteligente | Usar ClienteID; quatro atores (seção 3.1) | Baixa |
| D-16 | Pacotes e componentes | Autenticação citada sem mecanismo | Token Bearer estático em `AUTH_TOKEN` | Baixa |
| D-17 | Casos de uso | O UML1 liga o **Administrador do Sistema** a UC-08 e não tem "Agente Inteligente"; UC-02 a UC-04 ficam sem ator. A Spec v0 usava Cliente para UC-08 e Agente Inteligente para UC-02 a UC-04 | UC-08: Cliente (titular, via API) e Administrador. Agente Inteligente entra como ator de automação [ADIÇÃO]; acrescentar ao UML1 até a apresentação final | Média |
| D-18 | Classes | Métodos nomeados de forma diferente entre Spec v0 e UML2 (`ClassificaPerfil` × `ClassificarPerfil`) | Vale o nome do diagrama (C-04): `ClassificarPerfil` | Baixa |
| D-19 | Classes | Associação "ConsentimentoLGPD 1 possui 0..1 PerfilCliente": o perfil pertence ao cliente, não ao consentimento | `Cliente 1 → 0..1 PerfilCliente` via `PerfilCliente.ClienteID` | Baixa |
| D-20 | Estados × classes | O UML5 chama o atributo de `dataConcessao`; o UML2, de `dataConsentimento` | Vale o nome da classe: `DataConsentimento` | Baixa |
| D-22 | Casos de uso, classes, sequência | Nenhum diagrama diz como um Cliente entra no sistema; todos os UCs exigem cliente existente | [ADIÇÃO] UC-09 `CadastrarClienteUseCase`, `NovoCliente` e `POST /v1/clientes`; acrescentar ao UML1 até 19/11 | Alta |
| D-23 | Configuração | Até a fase Generic não há PostgreSQL, mas a API precisa rodar para demonstração | `DATABASE_URL` vazia = repositórios em memória com segmentos e catálogo sintéticos; com valor, PostgreSQL (desde a T-19, Q-07) | Baixa |
| D-21 | Implantação | O produto foi nomeado **Perfila**, mas o UML8 usa `omega-server` e `omega_db` | Artefatos mantêm os nomes do diagrama; módulo Go e repositório usam `perfila` | Baixa |

## 4. Histórico de versões

| Versão | Data | Mudanças |
|---|---|---|
| v0 | 27/10/2026 | Documento único (PDF) e depois dividido em três `.md` |
| v1.10 | 14/11/2026 | Portal do titular (Q-55 a Q-60): ator Titular e casos de uso `SolicitarAcessoTitular`, `AutenticarTitular`, `ConsultarMeusDados`, `PedirEliminacao`, `ListarPedidosEliminacao` e `ListarCaixaSaida` (3.1, 3.5); entidades `TokenAcessoTitular`, `PedidoEliminacao` e `MensagemSimulada` (3.2); RN-19 (acesso por link de uso único) e RN-20 (pedido de eliminação), RN-17 com os tokens e o pedido atendido (3.3); `ErrAcessoInvalido` (3.5); `ClienteRepository.BuscarPorEmail`, `TokenAcessoRepository`, `PedidoEliminacaoRepository`, `Notificador` e `CaixaSaidaSimulada` (3.6); pedidos e caixa de saída no painel (3.7.1); seção 3.7.2; `PORTAL_URL_BASE`, `PORTAL_CHAVE_SESSAO` e migração 010 (3.8). Empilhada sobre a v1.9; os números seguem os das branches abertas |
| v1.9 | 13/11/2026 | Direitos do titular e retenção (Q-48 a Q-53): `Cliente.DataEliminacao`, `Anonimizar` e `EstaEliminado` (3.2); RN-01 com o cliente eliminado, RN-03 sem D-13, RN-17 (eliminação) e RN-18 (retenção) (3.3); `ErrClienteEliminado` e os casos de uso `AtualizarCadastro`, `ExportarDadosTitular`, `EliminarDadosTitular` e `AplicarRetencao` (3.1, 3.5); `RemoverPorCliente` em eventos, perfis e recomendações e `ConsentimentoRepository.ListarRevogadosAte` (3.6); aba Direitos do titular e retenção na Operação, log `direito_atendido` (3.7.1); `RETENCAO_DIAS_REVOGADO`, subcomando `aplicar-retencao` e migração 009 (3.8); exceção dos direitos do titular em C-03. As versões v1.5 a v1.8 estão nas branches abertas (usuários e auditoria, segmentos e catálogo, relatórios, importação de eventos) |
| v1.4 | 12/11/2026 | Histórico completo do consentimento (Q-34): entidade `MovimentoConsentimento` (3.2), RN-10, UC-08 `Historico` e origem pelo contexto (3.5), `FichaCliente.Historico`, `ConsentimentoRepository.Registrar` e `ListarMovimentos` (3.6), origem nas rotas da API e do painel e linha do tempo pelo histórico (3.7, 3.7.1), migração 006 (3.8) |
| v1.3 | 12/11/2026 | Painel web do operador (Q-20 a Q-29): seção 3.7.1, casos de uso de aplicação do painel (3.1 e 3.5), `ProcessarClienteUseCase` com o estado por etapa (RN-09, Q-26), `ClienteRepository.Listar`, `RecomendacaoRepository.ListarPorCliente` e `ProdutoRepository.ListarTodos` (3.6), configuração `PAINEL_*` (3.8), migração 004, subcomando `hash-senha` (Q-28); camada BDD `@web` |
| v1.2 | 11/11/2026 | Fase Generic (T-19): `DATABASE_URL` vazia mantém o modo memória fora da implantação (Q-07); dados de demonstração por migração (Q-08); migrações aplicadas pelo servidor com trava (Q-09); driver pgx (Q-10); suíte de contrato dos repositórios; reforços de integridade no esquema (FKs para segmentos e produtos, CHECKs de faixa); erro de `SegmentoRepository.BuscarPorID` explicitado. T-24: implantação com Nginx, duas réplicas e banco isolado (Q-11 a Q-14). Correções do mapeamento de bugs (Q-15 a Q-19): limite de Valor, tolerância de relógio, falhas de configuração como 500, uma recomendação ATIVA por cliente, `/saude` com o banco e corpo JSON único |
| v1.1 | 07/11/2026 | UC-09 Cadastrar Cliente (D-22), `NovoCliente` e rota `POST /v1/clientes`; Support implementado (API, middlewares, CMD, gateways, configuração); D-01 a D-23 adotadas como finais; Q-01 a Q-06 decididas (Q-02 muda o erro de "nenhum segmento" para falha de configuração) |
| **v1.0** | 03/11/2026 | Nome do produto (Perfila); Constituição movida para `CLAUDE.md`; diagramas transcritos em `docs/uml/UML1..UML8`; `ClassificaPerfil` → `ClassificarPerfil` (D-18); `NovoConsentimento`, `Precondicoes` e `ExpirarRecomendacoesUseCase` explicitados; atores de UC-08 corrigidos (D-17); D-07 e D-10 corrigidos; D-17 a D-21 novos; estrutura de pastas marcada por fase (Core, Support, Generic) |
