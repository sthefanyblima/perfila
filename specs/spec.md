# Especificação para IA Generativa (Spec)

**Perfila — Sistema de Segmentação e Perfilamento do Comportamento do Cliente — Go**

Equipe Omega · GoLang & CLP · Documento 1 de 3 · **Versão 1.11 (16/11/2026)** · Base: Módulos 1 a 5 (casos de uso, classes, pacotes, sequência, estados, componentes e implantação), transcritos em `docs/uml/UML1..UML8`

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

**Dentro do MVP:** UC-01, UC-02, UC-03, UC-04, UC-08 e UC-09 (cadastro, D-22) de ponta a ponta, como no diagrama de sequência, com consentimento obrigatório, API HTTP, PostgreSQL e implantação em contêineres; desde a v1.3, também o **painel web do operador** (`/painel`, Q-20), com as regras visuais e de texto do design system "Perfila Console" (`docs/design/`, ajustado pela Q-27); desde a v1.5, **operadores com papel, matriz de permissões e trilha de auditoria** no painel (Q-35 a Q-41); desde a v1.6, **edição de segmentos e do catálogo pelo painel**, com prévia e auditoria (Q-33, Q-42 a Q-44); desde a v1.7, os **relatórios agregados** de UC-05 e UC-07 em versão parcial (RN-15, Q-46); desde a v1.8, a **importação de eventos em lote** (CSV pelo painel e JSON pela API), linha a linha e com o consentimento conferido em cada uma (RN-16, Q-47); desde a v1.9, os **direitos do titular** pelo painel (correção, acesso e portabilidade, eliminação) e a **retenção** após a revogação (Q-48 a Q-53, RN-17, RN-18); desde a v1.10, o **portal do titular** (`/titular`, Q-55 a Q-60): o próprio cliente entra por um link de uso único enviado por e-mail, vê e revoga o consentimento, concede de novo, baixa os próprios dados e pede a eliminação (RN-19, RN-20); desde a v1.11, o **consentimento por finalidade** (`segmentacao` e `publicidade_terceiros`, Q-61) e **UC-06 em versão parcial, "Criar público para campanha"**: exportação de públicos em arquivo CSV com o e-mail em hash, remoção propagada pela revogação e relatório agregado por segmento (RN-21, RN-22, Q-62 a Q-67).

**Fora do MVP:** o restante de UC-05 a UC-07 (D-04; a v1.7 entrega só os indicadores agregados da RN-15, sem recorte por segmento, comparação entre períodos nem relatórios salvos, e a v1.11 só o público em arquivo de UC-06, sem campanha, orçamento, criativos nem integração por API com plataformas de anúncio); notificação real ao cliente (o portal usa uma caixa de saída simulada, Q-57); autenticação JWT ou OIDC e perfis de acesso na API (o `AUTH_TOKEN` continua único); gestão de segmentos e do catálogo pela API (desde a v1.6, só pelo painel; o `padrao` continua nascendo na migração inicial); reprocessamento de todos os clientes depois de mudar um critério (Q-45); rotas da API JSON para os direitos do titular (o titular os exerce pelo portal, Q-53, Q-55).

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
| UC-05 | Analisar Padrões de Comportamento | Analista de Marketing | `GerarRelatorioUseCase` | Parcial (v1.7, Q-46): categorias compradas e evolução mensal da RN-15 |
| UC-06 | Criar Campanha de Marketing Direcionada (v1.11, parcial: "Criar público para campanha", Q-62) | Analista de Marketing | `GerarPublicoUseCase`, `ExportarPublicoUseCase`, `PropagarRevogacaoUseCase`, `RelatorioAgregadoSegmentosUseCase` | Parcial (RN-21, RN-22) |
| UC-07 | Visualizar Relatórios de Segmentação | Analista de Marketing | `GerarRelatorioUseCase` | Parcial (v1.7, Q-46): clientes, consentimento, segmentos e recomendações da RN-15 |
| UC-08 | Gerenciar Consentimento de Dados (LGPD) | Cliente (titular) e Administrador do Sistema | `GerenciarConsentimentoUseCase` | Sim |
| UC-09 | Cadastrar Cliente [ADIÇÃO, D-22] | Cliente / Sistema Externo | `CadastrarClienteUseCase` | Sim |

Administrador do Sistema: ligado a UC-08 no UML1 e, pelo UML8, executa os subcomandos do CMD (`serve`, `expirar-recomendacoes`, `hash-senha` (Q-28), `carregar-exemplos` (Q-30) e `expurgar-auditoria` (Q-39)) por SSH.

**Pessoas que operam o painel (v1.5, Q-35, Q-37).** Cada uma é um `Operador` com um papel; o que cada papel alcança está na matriz da RN-11.

| Papel (`PapelOperador`) | Ator do UML1 | No painel |
|---|---|---|
| `ADMINISTRADOR` | Administrador do Sistema | Tudo o que o atendimento faz, mais Relatórios (v1.7), a eliminação de dados do titular (v1.9), Usuários, Auditoria e Operação (saúde, expiração e, desde a v1.9, retenção) |
| `ATENDIMENTO` | Atendimento [ADIÇÃO, Q-35] (persona Rita, `docs/design/01-design-thinking.md`), ligado a UC-01, UC-08 e UC-09 | Cadastra (UC-09), registra o consentimento (UC-08), lança eventos (UC-01) e processa (UC-01 a UC-04); desde a v1.8, importa eventos; desde a v1.9, na aba Direitos, corrige o cadastro e baixa os dados do titular; lê Segmentos e Catálogo |
| `ANALISTA` | Analista de Marketing | Lê e edita Segmentos e Catálogo e vê os Relatórios agregados (v1.7, RN-15), sem nenhum dado de cliente; UC-06 e o restante de UC-05 e UC-07 continuam fora do MVP (D-04) |

Casos de uso de aplicação do painel [ADIÇÃO, Q-23, Q-26]. Não são casos de uso novos do UML1: dão ao operador a leitura do que UC-01 a UC-09 gravaram e reúnem a orquestração de RN-09.

| Caso de uso de aplicação | Exige consentimento (RN-01) | Usado por |
|---|---|---|
| `ProcessarClienteUseCase` (RN-09) | Sim, em cada etapa | API (`POST .../processamento`) e painel |
| `ListarClientesUseCase` | Não; o segmento de cada linha só aparece com consentimento válido (Q-21) | Painel |
| `ConsultarClienteUseCase` | Não: só cadastro e consentimento (Q-21), como UC-08 | Painel |
| `ConsultarResultadoUseCase` | Sim (RN-01, exceto cliente ativo: inativo é só leitura, Q-31) | Painel |
| `ListarSegmentosUseCase`, `ListarCatalogoUseCase` | Não: configuração, sem dado pessoal | Painel |
| `SalvarSegmentoUseCase`, `SalvarProdutoUseCase` [v1.6, Q-33, Q-43] | Não: configuração, sem dado pessoal | Painel |
| `PreverSegmentacaoUseCase` [v1.6, Q-42] | Sim: só conta clientes ativos com consentimento válido, e só devolve números | Painel |
| `AutenticarOperadorUseCase`, `GerenciarOperadoresUseCase` [v1.5, Q-35, Q-36] | Não: dados de quem opera, não do titular | Painel e CMD (administrador inicial) |
| `RegistrarAuditoriaUseCase`, `ListarAuditoriaUseCase`, `ExpurgarAuditoriaUseCase` [v1.5, Q-39] | Não: a auditoria guarda só IDs | Painel e CMD (`expurgar-auditoria`) |
| `GerarRelatorioUseCase` (v1.7, RN-15) | Não por cliente: só números agregados, e o comportamento entra só de quem tem consentimento válido | Painel (tela Relatórios e CSV) |
| `ImportarEventosUseCase` (v1.8, RN-16) | Sim, em cada linha: as regras de UC-01 `Registrar` (RN-01 e RN-04) | Painel (tela Importar eventos) e API (`POST /v1/eventos/lote`) |

| `GerarPublicoUseCase`, `ExportarPublicoUseCase` (UC-06 parcial, v1.11) | Sim, por cliente: `segmentacao` **e** `publicidade_terceiros` (RN-21); quem não tem fica de fora, sem erro | Painel |
| `RelatorioAgregadoSegmentosUseCase` (v1.11) | Só conta quem tem `segmentacao` válida; não expõe nenhum identificador (RN-22) | Painel |

Analista de Marketing (UC-06, v1.11): gera e exporta públicos com o Administrador (permissão `publicos`, RN-11, Q-66).

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
type Finalidade string          // [ADIÇÃO, v1.11, Q-61] segmentacao, publicidade_terceiros

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
    Finalidade        Finalidade          // v1.11 (Q-61): enumerada e parte da chave; um registro por (cliente, finalidade)
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

// [ADIÇÃO, v1.5, Q-35] Quem opera o painel.
type PapelOperador string // ADMINISTRADOR, ANALISTA, ATENDIMENTO
type Permissao string     // clientes, consentimento, eventos, processamento, configuracao,
                          // operacao, usuarios, auditoria, relatorios, eliminacao, publicos (RN-11)

type Operador struct {
    ID                uuid.UUID
    Login             string        // único e imutável; minúsculas (RN-12)
    Nome              string        // para a tela Usuários; nunca vai para log nem auditoria (C-09)
    Papel             PapelOperador
    SenhaHash         string        // gerado pelo port HashSenha; nunca a senha
    Ativo             bool
    CriadoEm          time.Time
    TrocarSenha       bool          // senha provisória: só Minha senha até trocar (Q-40)
    SenhaAtualizadaEm time.Time     // carimbo da sessão: mudou, as sessões abertas caem (Q-38)
}

// [ADIÇÃO, v1.5, Q-39] Trilha de auditoria do painel. Só inclusão; sem dado pessoal.
type AcaoAuditoria string      // lista na RN-13
type TipoAlvo string           // cliente, segmento, produto, operador; vazio = sem alvo
type ResultadoAuditoria string // SUCESSO, NEGADO, FALHA

type RegistroAuditoria struct {
    ID        uuid.UUID
    Data      time.Time
    Operador  string    // login; vazio em entrada com login desconhecido
    Acao      AcaoAuditoria
    AlvoTipo  TipoAlvo
    AlvoID    uuid.UUID // uuid.Nil quando AlvoTipo é vazio
    Resultado ResultadoAuditoria
    Campos    []string  // [v1.6, Q-44] nomes dos campos alterados (lista fechada da RN-13); nunca valores
}

// [ADIÇÃO, v1.7, Q-46] Valores protegidos dos relatórios (RN-15). Não são persistidos.
const MinimoGrupo = 5

type Contagem struct {
    Valor  int
    Oculta bool // 1 a 4 clientes por trás do número: aparece "menos de 5"
}

type Media struct {
    Valor  float64
    Oculta bool // grupo de 1 a 4 clientes
    Vazia  bool // grupo sem clientes: não há média
}

type Percentual struct {
    Valor  float64 // de 0 a 100
    Oculta bool    // parte ou total oculto, ou total zero
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

// [ADIÇÃO, v1.11, Q-62 a Q-64] UC-06 parcial: públicos para campanha (RN-21).
type TipoExportacao string // INCLUSAO, REMOCAO
type StatusMembro string   // EXPORTADO, REMOCAO_PENDENTE, REMOVIDO
type TipoDestino string    // arquivo_csv (único no MVP)

type ExportacaoPublico struct { // registro de cada arquivo gerado; sem dado pessoal
    ID              uuid.UUID
    SegmentoID      uuid.UUID
    SegmentoNome    string
    Tipo            TipoExportacao
    Destino         TipoDestino
    Operador        string    // login de quem exportou (origem do contexto, como RN-10); "sistema" sem origem
    Data            time.Time
    Quantidade      int       // linhas do arquivo
    HashArquivo     string    // SHA-256 do conteúdo, em hexadecimal
    NomeArquivo     string
    ArquivoExpiraEm time.Time // Data + PUBLICOS_ARQUIVO_HORAS (Q-64)
}

type MembroPublico struct { // um cliente dentro de um arquivo de inclusão: pseudônimo, ainda dado pessoal
    ID                  uuid.UUID
    ExportacaoID        uuid.UUID  // a exportação de inclusão que o levou
    SegmentoID          uuid.UUID
    ClienteID           uuid.UUID
    HashEmail           string     // HashEmail(email) no momento da inclusão
    Status              StatusMembro
    DataInclusao        time.Time
    DataRemocaoPendente *time.Time
    RemocaoID           *uuid.UUID // a exportação de remoção que o tirou
    DataRemocao         *time.Time
}

type LotePublico struct { // o que vai para o destino: só hashes e o nome do segmento
    Tipo         TipoExportacao
    SegmentoNome string
    Data         time.Time
    Hashes       []string // em ordem crescente, sem repetição
}

type ReciboEnvio struct { // o que o destino devolve
    Destino     TipoDestino
    NomeArquivo string
    Conteudo    []byte // o arquivo gerado (destino arquivo_csv); vazio num destino por API
    Hash        string // SHA-256 do Conteudo, em hexadecimal
}

// [ADIÇÃO, v1.11, Q-65] Relatório agregado (RN-22): nenhum identificador.
type LinhaAgregada struct {
    SegmentoNome       string
    Quantidade         int     // 0 quando Oculta
    TicketMedio        float64 // média dos TicketMedio dos perfis; 0 quando Oculta
    CategoriaMaisComum string  // vazia quando Oculta
    Oculta             bool    // grupo com menos de TamanhoMinimoGrupo (5) clientes
}
```

| Entidade | Métodos | Observação |
|---|---|---|
| Cliente | `NovoCliente(id uuid.UUID, nome, email string, agora time.Time) (Cliente, error)` (construtor, D-22), `AtualizarDados(nome, email string) error`, `Ativar()`, `Desativar()`; [ADIÇÃO, v1.9, Q-50] `Anonimizar(agora time.Time) error`, `EstaEliminado() bool` | Nome e e-mail não podem ser vazios; nasce ativo, com DataCadastro = agora. `Anonimizar` troca Nome por "Titular eliminado" e Email por `eliminado@anonimizado.invalid`, desativa e preenche DataEliminacao = agora; num cliente já eliminado, `ErrClienteEliminado`. `AtualizarDados` num cliente eliminado: `ErrClienteEliminado` |
| ConsentimentoLGPD | `NovoConsentimento(id, clienteID uuid.UUID, finalidade Finalidade, agora time.Time)` (construtor, estado NAO_CONCEDIDO; a v1.11 acrescenta a finalidade, Q-61), `Conceder(finalidade Finalidade, versaoTermo string, agora time.Time) error`, `Revogar(agora time.Time) error`, `EstaValido() bool` | Transições da seção 3.4.1. Desde a v1.11 o registro é de uma finalidade só: `Conceder` exige a mesma finalidade do registro |
| Finalidade [ADIÇÃO, v1.11] | `Valida() bool`; `Finalidades() []Finalidade` (as duas, nesta ordem: `segmentacao`, `publicidade_terceiros`) | Fora do enumerado: `ErrDadosInvalidos` em quem a recebe |
| MembroPublico [ADIÇÃO, v1.11] | `NovoMembro(id, exportacaoID, segmentoID, clienteID uuid.UUID, hashEmail string, agora time.Time) (MembroPublico, error)` (estado EXPORTADO), `MarcarRemocao(agora time.Time) error`, `Remover(remocaoID uuid.UUID, agora time.Time) error` | Transições da seção 3.4.3 |
| (função) [ADIÇÃO, v1.11] | `HashEmail(email string) string` | SHA-256, em hexadecimal minúsculo, do e-mail normalizado: sem nenhum espaço e em minúsculas (RN-21) |
| MovimentoConsentimento [ADIÇÃO, Q-34] | `NovoMovimento(id uuid.UUID, c ConsentimentoLGPD, tipo TipoMovimento, origem OrigemMovimento, operador string) (MovimentoConsentimento, error)` | Construído logo depois da transição, com Finalidade, VersaoTermo e Data (= `DataAtualizacao`) do consentimento. Tipo e Origem fora dos enumerados, ou Origem painel sem operador: `ErrDadosInvalidos`. Não tem métodos de alteração. Desde a v1.11, `FinalidadeEfetiva() Finalidade`: a finalidade do texto, ou `segmentacao` se o texto estiver fora do enumerado (movimentos antigos, Q-61) |
| EventoComportamental | `Validar(agora time.Time) error` | RN-04 |
| PerfilCliente | `CalcularScore(eventos []EventoComportamental, agora time.Time, meta int)`, `AtualizarPerfil(novo PerfilCliente)` | RN-05. `AtualizarPerfil` copia as métricas e preserva ID, ClienteID e SegmentoID. `meta <= 0` resulta em score 0 |
| Segmento | `ClassificarPerfil(p PerfilCliente) (bool, error)`; [v1.6] `Validar() error`, `EhPadrao() bool` | RN-06. Nome corrigido na v1.0 para o do UML2 (era `ClassificaPerfil`, contrariando C-04). `Validar` aplica as regras de um segmento da RN-14 (nome, descrição, prioridade e critério pela gramática, com `ValidarCriterio`); `EhPadrao` é critério vazio |
| Configuração de segmentos [ADIÇÃO, v1.6, Q-33] | `ValidarCriterio(criterio string) error`, `ValidarSegmentos(lista []Segmento) error` | Funções puras da RN-14: a primeira só confere a gramática da RN-06; a segunda, as regras do conjunto. Erro: `errs.ErroCampo` (embrulha `ErrDadosInvalidos`) |
| RecomendacaoProduto | `NovaRecomendacao(...)` (equivale a `Gerar()`), `Disponibilizar(agora time.Time, validadeDias int) error`, `Expirar(agora time.Time) error` | Transições da seção 3.4.2 |
| Produto | `AtualizarPreco(preco float64) error`; [v1.6] `Validar() error`, `NormalizarCategoria(categoria string) string` | Preço de 0 a 1.000.000 (v1.6, Q-43; antes só maior ou igual a zero). `Validar` e `NormalizarCategoria`: RN-14 |
| Operador [ADIÇÃO, Q-35] | `NovoOperador(id uuid.UUID, login, nome string, papel PapelOperador, senhaHash string, agora time.Time) (Operador, error)`, `MudarPapel(p PapelOperador) error`, `Desativar()`, `Ativar()`, `DefinirSenha(senhaHash string, trocar bool, agora time.Time) error` | RN-12. Nasce ativo, com CriadoEm e SenhaAtualizadaEm = agora e `TrocarSenha` verdadeiro. Login, nome, papel ou hash inválidos: `ErrDadosInvalidos` |
| PapelOperador [ADIÇÃO, Q-37] | `Pode(p Permissao) bool` | Matriz da RN-11; papel fora do enumerado não pode nada |
| RegistroAuditoria [ADIÇÃO, Q-39] | `NovoRegistroAuditoria(id uuid.UUID, data time.Time, operador string, acao AcaoAuditoria, alvoTipo TipoAlvo, alvoID uuid.UUID, resultado ResultadoAuditoria, campos ...string) (RegistroAuditoria, error)` | RN-13. Enumerados fora da lista, alvo sem tipo ou tipo sem alvo, campo fora da lista fechada (v1.6, Q-44): `ErrDadosInvalidos`. Campos sem repetição, na ordem da lista. Sem métodos de alteração |
| Contagem, Media, Percentual [ADIÇÃO, Q-46] | `ProtegerContagem(valor, clientes int) Contagem`, `ProtegerMedia(soma float64, clientes int) Media`, `ProtegerPercentual(parte, total Contagem) Percentual` | RN-15. Funções puras; `clientes` é o número de clientes distintos por trás do valor |
| TokenAcessoTitular [ADIÇÃO, v1.10] | `NovoTokenAcesso(id, clienteID uuid.UUID, hash string, agora time.Time) (TokenAcessoTitular, error)`, `EstaValido(agora time.Time) bool`, `Usar(agora time.Time) error` | Hash fora do formato ou ClienteID nulo: `ErrDadosInvalidos`. ExpiraEm = agora + 30 min. Válido enquanto não usado e com agora antes de ExpiraEm. `Usar` num token inválido: `ErrAcessoInvalido`; senão, UsadoEm = agora |
| PedidoEliminacao [ADIÇÃO, v1.10] | `NovoPedidoEliminacao(id, clienteID uuid.UUID, agora time.Time) (PedidoEliminacao, error)`, `Atender(agora time.Time) error` | Nasce ABERTO, com DataPedido = agora; ClienteID nulo: `ErrDadosInvalidos`. `Atender`: ABERTO → ATENDIDO com DataAtendimento = agora; já ATENDIDO: `ErrTransicaoInvalida` |
| MensagemSimulada [ADIÇÃO, v1.10] | — | Só leitura depois de gravada |

Desvios idiomáticos em relação aos diagramas (não são lacunas): `context.Context` como primeiro parâmetro; `error` no lugar de `bool` em `Validar` para informar a causa; construtor no lugar de `Gerar()`; tempo recebido por parâmetro ou `Clock`, nunca `time.Now()` (C-07); assinaturas dos serviços ajustadas na seção 3.6.

### 3.3 Regras de negócio

- **RN-01 Pré-condições comuns (UC-01 a UC-04).** Verificar nesta ordem: cliente existe (`ErrClienteNaoEncontrado`), cliente não foi eliminado (`ErrClienteEliminado`, v1.9, RN-17), cliente está ativo (`ErrClienteInativo`), consentimento é válido **para a finalidade `segmentacao`** (`ErrConsentimentoInvalido`; v1.11, Q-61). `publicidade_terceiros` sozinha não libera UC-01 a UC-04.
- **RN-02 Consentimento válido.** Válido para uma finalidade se e somente se o registro do cliente **naquela finalidade** tem Status CONCEDIDO. A ausência de registro equivale a NAO_CONCEDIDO (D-02). Desde a v1.11 (Q-61) a finalidade filtra o uso dos dados: há um registro por (cliente, finalidade), cada um com a máquina de estados da 3.4.1, e cada uso pede a sua (RN-01: `segmentacao`; RN-21: `segmentacao` e `publicidade_terceiros`). Finalidades: `segmentacao` (uso interno: perfil, segmento e recomendação) e `publicidade_terceiros` (compartilhar o público com plataformas de anúncio, LGPD art. 7º §5º e art. 8º §4º). Os registros anteriores à v1.11 passam a valer só para `segmentacao` (migração 011).
- **RN-03 Revogação.** Tem efeito imediato e vale só para a finalidade revogada: revogar `segmentacao` faz UC-01 a UC-04 falharem com `ErrConsentimentoInvalido`; revogar `publicidade_terceiros` não muda a `segmentacao`. Nos dois casos, o cliente sai dos públicos já exportados (RN-21, remoção propagada). Dados já persistidos permanecem até a retenção (RN-18, que olha a `segmentacao`) ou um pedido de eliminação (RN-17); desde a v1.9 (Q-51), a D-13 foi substituída.
- **RN-04 Validação de evento.** ClienteID não nulo; Tipo entre navegacao, compra e interacao; Origem não vazia; Timestamp não zero e menor ou igual a agora mais 2 minutos de tolerância de relógio (Q-16); Valor maior ou igual a zero e no máximo 1.000.000 (Q-15); se Tipo é compra, Valor maior que zero. Qualquer violação: `ErrDadosInvalidos`.
- **RN-05 Perfil.** Função pura dos eventos do cliente e do relógio:
    - FrequenciaCompras é o número de eventos do tipo compra.
    - TicketMedio é a média do Valor das compras (zero se não houver compras).
    - CategoriaPreferida é a categoria mais frequente entre as compras com categoria preenchida; empate resolvido por ordem alfabética; vazia se não houver.
    - ScoreEngajamento é o menor valor entre 1 e N dividido por `ENGAJAMENTO_META_EVENTOS`, onde N é o número de eventos de qualquer tipo com Timestamp maior que agora menos 30 dias e menor ou igual a agora.
    - Cliente sem eventos gera perfil zerado, sem erro.
    - Há um perfil por cliente: `Salvar` faz upsert por ClienteID e preserva o ID original.
- **RN-06 Segmentação.** Segmentos ativos são avaliados em ordem crescente de Prioridade; o primeiro cujo critério casa vence. Critério vazio casa sempre e deve ter a maior Prioridade (segmento padrao, criado por migração e protegido pela RN-14). Segmentos inativos não são avaliados. Gramática do critério:
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
- **RN-10 Histórico do consentimento (v1.4, Q-34).** Toda transição de 3.4.1 aplicada por UC-08 gera um `MovimentoConsentimento`, gravado junto com o registro do consentimento numa única operação atômica (`ConsentimentoRepository.Registrar`): ou os dois ficam gravados, ou nenhum. Movimentos são só de inclusão: nunca alterados nem apagados. O histórico é a lista dos movimentos do cliente em ordem cronológica (desempate pela ordem de gravação). O registro de `consentimentos_lgpd` continua sendo a fonte do estado (RN-02); o histórico é prova e não decide a validade. Transição inválida ou guarda violada não gera movimento. A origem vem do adaptador que chamou o caso de uso (API: `api`; painel: `painel` com o login do operador; portal do titular, fase 2: `portal`); sem origem informada (subcomandos, exemplos, testes): `sistema`. Movimentos criados pela migração 006 a partir das datas existentes têm origem `migracao`. Desde a v1.11, cada movimento pertence à finalidade que registra; movimentos antigos com uma finalidade fora do enumerado contam como `segmentacao` (Q-61), e o texto original é mantido.
- **RN-17 Eliminação a pedido do titular (v1.9, Q-50).** `EliminarDadosTitularUseCase` atende o art. 18, VI, nesta ordem:
    1. Cliente existe (`ErrClienteNaoEncontrado`) e ainda não foi eliminado (`ErrClienteEliminado`).
    2. A confirmação é o nome atual do cliente, comparado sem diferenciar maiúsculas, acentos e espaços nas pontas. Diferente: `ErrDadosInvalidos`, e nada é apagado.
    3. Consentimento CONCEDIDO em qualquer finalidade (v1.11): `ErrTransicaoInvalida`. É preciso revogar antes, pelo UC-08, o que deixa o movimento REVOGACAO como prova e, na `publicidade_terceiros`, propaga a remoção aos públicos (RN-21).
    4. Apaga os eventos, o perfil e as recomendações do cliente e, desde a v1.10, os tokens de acesso ao portal (Q-58).
    5. Desde a v1.10 (RN-20), marca como ATENDIDO o pedido de eliminação aberto do cliente, se houver.
    6. Por último, `Anonimizar` e grava o cadastro. Sendo o último passo, uma falha antes dele deixa o cliente sem eliminar, e repetir conclui o pedido.
    - O consentimento e os movimentos ficam como prova do tratamento (art. 16, I): não têm nome nem e-mail.
    - A eliminação é irreversível. Repetir depois de uma falha conclui o pedido, porque as remoções não falham sem dados.
    - Depois dela, nenhuma escrita vale para o cliente: a correção, UC-08 Conceder e Revogar e as pré-condições RN-01 respondem `ErrClienteEliminado`. A busca pelo nome antigo não o encontra, porque o nome não existe mais. A exportação entrega o que restou.
- **RN-18 Retenção após a revogação (v1.9, Q-51).** O comportamento (eventos, perfil e recomendações) de quem está com o consentimento de `segmentacao` (v1.11) REVOGADO e `DataRevogacao` há `RETENCAO_DIAS_REVOGADO` dias ou mais (menor ou igual a agora menos o prazo) é apagado por `AplicarRetencaoUseCase`. O cadastro, o consentimento e o histórico ficam. Não atinge quem concedeu de novo (CONCEDIDO), quem nunca concedeu nem quem está dentro do prazo. Devolve quantos clientes tiveram algum dado apagado; executar de novo sem nada vencido devolve 0. Executada pelo subcomando `aplicar-retencao` e pelo painel (Operação).
- **RN-19 Acesso do titular ao portal (v1.10, Q-56).**
    1. `SolicitarAcessoTitularUseCase` recebe um e-mail (sem espaços nas pontas; vazio ou sem `@`: `ErrDadosInvalidos`) e procura os cadastros com esse e-mail, sem diferenciar maiúsculas. Para cada cadastro não eliminado com menos de 3 links criados na última hora, gera um token de 32 bytes aleatórios, grava só o hash SHA-256 (`TokenAcessoTitular`, 30 minutos) e envia pelo `Notificador` o link `PORTAL_URL_BASE + "/titular/acesso?t=" + token`.
    2. O resultado é o mesmo para e-mail desconhecido, cadastro eliminado e limite atingido: nada é enviado e nenhum erro é devolvido. Quem chama nunca sabe se o e-mail existe.
    3. `AutenticarTitularUseCase` recebe o token, calcula o hash e o consome no repositório (`Consumir`, atômico): inexistente, expirado ou já usado é `ErrAcessoInvalido`, e um segundo uso do mesmo link falha mesmo em outra réplica. Cliente inexistente ou eliminado também é `ErrAcessoInvalido`. Sucesso devolve o ID do cliente, que vira a sessão do portal.
    4. O token e o link nunca vão para o log (C-09).
- **RN-20 Pedido de eliminação pelo portal (v1.10, Q-58).** `PedirEliminacaoUseCase` confere, nesta ordem: cliente existe (`ErrClienteNaoEncontrado`), não foi eliminado (`ErrClienteEliminado`) e o consentimento não está CONCEDIDO (`ErrTransicaoInvalida`: é preciso revogar antes, como na RN-17). Se já houver um pedido ABERTO do cliente, devolve esse mesmo pedido; senão, cria um pedido ABERTO. Quem atende é a operação, pela eliminação da RN-17, que marca o pedido como ATENDIDO. Com pedido aberto, o portal não oferece conceder de novo.

- **RN-11 Papéis e permissões (v1.5, Q-37).** O que cada papel pode é uma função pura do domínio (`PapelOperador.Pode`):

    | Permissão | O que libera no painel | ADMINISTRADOR | ATENDIMENTO | ANALISTA |
    |---|---|---|---|---|
    | `clientes` | Lista, busca, cadastro e ficha (todas as abas) | Sim | Sim | Não |
    | `consentimento` | Conceder e revogar | Sim | Sim | Não |
    | `eventos` | Registrar evento; desde a v1.8, Importar eventos (RN-16) | Sim | Sim | Não |
    | `processamento` | Processar agora | Sim | Sim | Não |
    | `configuracao` | Segmentos e Catálogo (leitura) | Sim | Sim | Sim |
    | `editar_configuracao` [v1.6, Q-33] | Segmentos e Catálogo: criar, editar, ativar, desativar, prévia e testar critério | Sim | Não | Sim |
    | `operacao` | Operação: saúde, expiração de recomendações, retenção (v1.9) e Caixa de saída simulada (v1.10) | Sim | Não | Não |
    | `usuarios` | Usuários: criar, mudar papel, desativar, ativar, redefinir senha | Sim | Não | Não |
    | `auditoria` | Auditoria | Sim | Não | Não |
    | `relatorios` [v1.7, Q-46] | Relatórios agregados e o CSV (RN-15) | Sim | Não | Sim |
    | `eliminacao` [v1.9, Q-53] | Eliminar dados do titular (RN-17); Pedidos do titular (v1.10) | Sim | Não | Não |
    | `publicos` [v1.11, Q-66] | Públicos para campanha e relatório por segmento (RN-21, RN-22) | Sim | Não | Sim |

    Todo operador, de qualquer papel, troca a própria senha ("Minha senha") e sai. Sem a permissão: `ErrSemPermissao`.
- **RN-12 Operadores (v1.5, Q-35, Q-36, Q-40).**
    - Login: de 3 a 32 caracteres entre letras minúsculas sem acento, dígitos, `.`, `-` e `_`, começando por letra; a entrada é convertida para minúsculas antes de validar. Único (`ErrLoginEmUso`), imutável e nunca reutilizado: operador não é excluído, é desativado.
    - Nome: obrigatório, até 100 caracteres. Papel: um dos três.
    - Senha: de 12 caracteres a 72 bytes (limite do bcrypt), validada no caso de uso antes do hash (`ErrDadosInvalidos`). O domínio só guarda o hash.
    - Operador criado ou com senha redefinida pelo administrador fica com `TrocarSenha`: até trocar, só "Minha senha" e Sair. A troca pelo próprio operador exige a senha atual (`ErrCredenciaisInvalidas`) e apaga `TrocarSenha`. Toda definição de senha atualiza `SenhaAtualizadaEm`, o que encerra as sessões abertas (Q-38).
    - O administrador não muda o próprio papel nem desativa a si mesmo (`ErrPropriaConta`). Como só um administrador ativo gerencia operadores, sempre resta pelo menos um.
    - Administrador inicial: sem nenhum operador gravado, `PAINEL_USUARIO` e `PAINEL_SENHA_HASH` criam um `ADMINISTRADOR` (nome "Administrador", sem `TrocarSenha`, porque a senha já foi gerada fora do sistema, Q-28). A criação é atômica (`OperadorRepository.CriarSeVazio`): com duas réplicas subindo juntas, só uma grava. Com algum operador gravado, as duas variáveis são ignoradas.
    - Entrar: login inexistente, senha errada e operador inativo dão o mesmo `ErrCredenciaisInvalidas`; login inexistente também passa pela conferência de hash, para o tempo não revelar quem existe.
- **RN-13 Auditoria (v1.5, Q-39).** Toda ação do painel que muda dados, toda leitura de ficha e todo acesso negado geram um `RegistroAuditoria` antes da resposta. Ações (`AcaoAuditoria`) e alvo:

    | Ação | Alvo | Quando |
    |---|---|---|
    | `entrar` | — | Entrada: SUCESSO, ou NEGADO com o login só se ele existir |
    | `sair`, `sessao_encerrada` | — | Saída; sessão recusada por operador inativo, inexistente ou com senha trocada |
    | `acesso_negado` | cliente, se a rota tiver um | Rota sem a permissão do papel (403) |
    | `ler_ficha` | cliente | Cada aba da ficha aberta: visão geral, eventos, consentimento |
    | `cadastrar_cliente`, `conceder_consentimento`, `revogar_consentimento`, `registrar_evento`, `processar` | cliente | UC-09, UC-08, UC-01, RN-09 |
    | `expirar_recomendacoes` | — | Expiração pela tela Operação |
    | `criar_operador`, `mudar_papel`, `desativar_operador`, `ativar_operador`, `redefinir_senha` | operador | Tela Usuários |
    | `trocar_senha` | operador | Minha senha |
    | `criar_segmento`, `editar_segmento`, `ativar_segmento`, `desativar_segmento` [v1.6] | segmento | Tela Segmentos, com os campos alterados (RN-14) |
    | `prever_segmentacao` [v1.6] | segmento (vazio para um segmento novo) | Prévia antes de salvar (Q-42) |
    | `criar_produto`, `editar_produto`, `ativar_produto`, `desativar_produto` [v1.6] | produto | Tela Catálogo, com os campos alterados |

    - Resultado: `SUCESSO`; `NEGADO` (sem permissão ou credenciais); `FALHA` (erro de domínio ou interno; o código do erro vai só para o log).
    - Nunca guarda nome, e-mail, senha, termo de busca nem conteúdo de formulário: só login do operador, tipo e ID do alvo (C-09) e, desde a v1.6 (Q-44), os **nomes** dos campos alterados de segmento ou produto, de uma lista fechada: `nome`, `descricao`, `criterio`, `prioridade`, `ativo`, `categoria`, `preco`. A lista e a busca de clientes não são registradas (Q-24).
    - Leitura sem registro não é mostrada: se a gravação da auditoria falhar, a ficha responde erro interno. Numa ação que muda dados, a ação já aconteceu: a resposta segue e a falha vai para o log (`auditoria_falhou`).
    - Só inclusão, como os movimentos (RN-10): nunca alterados. Retenção de `AUDITORIA_RETENCAO_DIAS` (padrão 180, seis meses): o subcomando `expurgar-auditoria` apaga os registros com Data anterior a agora menos esse prazo; um registro com exatamente o prazo permanece.
    - Só o painel é auditado: a API não tem operador (D-16) e já deixa o log de acesso e os movimentos do consentimento.
- **RN-14 Configuração de segmentos e catálogo (v1.6, Q-33, Q-42, Q-43).**
    - **Segmento** (`Segmento.Validar`): nome de 1 a 40 caracteres depois de tirar os espaços das pontas, sem `;` nem `:` (o nome entra no `Motivo`, RN-07); descrição até 200; prioridade de 1 a 999; critério pela gramática da RN-06 (`ValidarCriterio`), com a condição errada na mensagem.
    - **Conjunto** (`ValidarSegmentos`, aplicada à configuração inteira já com a mudança): exatamente um segmento de critério vazio, o `padrao`; ele está ativo e tem a maior prioridade entre os ativos; prioridade única entre os ativos; nome único sem diferenciar maiúsculas. Assim, nenhum segmento novo ou editado fica com critério vazio, o `padrao` não ganha critério nem é desativado, e a configuração gravada nunca leva o processamento a `ErrSemSegmentoAplicavel` (Q-02) ou a `ErrSegmentoMalConfigurado` (Q-17).
    - Segmentos e produtos **nunca são excluídos**, só desativados (chaves estrangeiras de `perfis_cliente.segmento_id` e `recomendacoes_produto.produto_id`). ID e nome do `padrao` não mudam pela edição de outro segmento.
    - **Perfis já segmentados** mantêm o `SegmentoID` até o próximo processamento do cliente (UC-03), mesmo que o segmento mude ou seja desativado; UC-04 continua lendo o segmento por `BuscarPorID`. Reprocessar todos fica fora (Q-45).
    - **Prévia** (Q-42): obrigatória antes de salvar critério, prioridade ou situação de um segmento. Considera só os clientes ativos com consentimento válido e o perfil salvo de cada um (nada é recalculado nem gravado). Para cada segmento ativo da nova configuração, diz quantos estão nele hoje e quantos ficariam com a nova configuração; e quantos ainda não têm perfil. Só contagens, nunca nome, e-mail ou ID.
    - **Testar critério**: avalia um critério com frequência de compras, ticket médio e score digitados; não lê nem grava dado de cliente.
    - **Produto** (`Produto.Validar`): nome de 1 a 100 caracteres; categoria de 1 a 40, gravada por `NormalizarCategoria` (sem espaços nas pontas e em minúsculas), porque precisa casar exatamente com a dos eventos (RN-05); preço de 0 a 1.000.000 (`AtualizarPreco`, Q-43). Produto desativado sai da RN-07 a partir do próximo processamento; recomendações ATIVA dele seguem até expirar ou até o próximo processamento do cliente.
    - Erros de validação são `errs.ErroCampo` (campo e mensagem em português), que embrulham `ErrDadosInvalidos`: o painel mostra a mensagem ao lado do campo.
- **RN-15 Relatórios agregados (v1.7, Q-46).** Os relatórios de UC-05 e UC-07 mostram só números agregados, nunca nome, e-mail, ID ou qualquer linha de cliente.
    - **Base de cada indicador.** Contagens de cadastro (total, ativos, cadastros por mês) e de consentimento (clientes por estado, taxa de concessão, revogações por mês) incluem todos os clientes; cliente sem registro de consentimento conta como NAO_CONCEDIDO (D-02). Métricas de comportamento (segmento, ticket médio, score médio, recomendações, categorias compradas, eventos por mês) contam só clientes com consentimento válido no momento do relatório (RN-02), ativos ou inativos (Q-31); revogados e não concedidos ficam de fora, inclusive o que gravaram antes.
    - **Grupos pequenos.** Uma contagem com 1 a 4 clientes distintos por trás aparece como "menos de 5" (`MinimoGrupo` = 5); zero aparece como 0. Contagens que não são de clientes (eventos, compras, recomendações) seguem o número de clientes distintos que as geraram. Médias (ticket médio, score médio) de grupo com 1 a 4 clientes ficam ocultas; sem clientes, não há média. A taxa de concessão (concedidos ÷ total, em %) fica oculta se uma das duas contagens estiver oculta ou se o total for zero.
    - **Categorias.** Só compras com categoria preenchida, dentro do período, por número de compras (decrescente) e nome. Categorias com 1 a 4 clientes compradores não aparecem pelo nome: são somadas na linha "Outras categorias", cuja contagem de compras segue os clientes distintos do conjunto.
    - **Segmentos.** Uma linha por segmento ativo, em ordem de Prioridade, com clientes, ticket médio e score médio dos perfis; segmento inativo só aparece se ainda tiver perfis; clientes com consentimento válido ainda sem segmento ficam na linha "Ainda não segmentados".
    - **Recomendações.** ATIVA e EXPIRADA, contadas por recomendação e protegidas pelos clientes distintos de cada estado.
    - **Período.** Últimos 6 (padrão) ou 12 meses-calendário, contando o mês atual, no fuso America/Sao_Paulo; outro valor: `ErrDadosInvalidos`. Categorias e evolução mensal (cadastros, eventos e revogações, estas pelos movimentos de RN-10) respeitam o período; clientes, consentimento, segmentos e recomendações são a situação no momento do relatório.
    - **Exportação.** O mesmo relatório em CSV, com os mesmos valores protegidos ("menos de 5"), sem nenhuma outra coluna.
- **RN-16 Importação de eventos em lote (v1.8, Q-47).** Um lote é uma lista de linhas de evento (CSV pelo painel, JSON pela API) tratada **linha a linha**: as linhas aceitas são gravadas e as outras ficam no relatório, sem interromper o lote.
    - **Arquivo CSV.** UTF-8 (BOM inicial aceito e ignorado), vírgula como separador, aspas duplas do RFC 4180. A primeira linha é o cabeçalho, com as colunas por nome, em qualquer ordem: `cliente_id`, `tipo`, `origem`, `timestamp` e `valor` obrigatórias, `categoria` opcional; coluna repetida, desconhecida ou faltando torna o arquivo inválido. Linhas em branco são ignoradas. Cada linha é identificada pelo número da linha no arquivo (o cabeçalho é a linha 1).
    - **Limites.** Até 256 KiB por arquivo e até 1.000 linhas de evento por lote (CSV ou JSON). Arquivo vazio ou só com o cabeçalho, maior que o limite, com mais linhas, fora de UTF-8, com cabeçalho inválido ou com aspas mal formadas é recusado inteiro, sem gravar nada (`ErrDadosInvalidos`, com o motivo do arquivo).
    - **Formato de cada linha.** `cliente_id` é um UUID; `timestamp` segue o ISO 8601 com data, hora e fuso (RFC 3339, por exemplo `2026-11-10T14:30:00-03:00`), guardado em UTC; `valor` é decimal com ponto (vazio vale 0); `tipo` é `compra`, `navegacao` ou `interacao`. Linha com número de colunas diferente do cabeçalho, ou com campo fora do formato, é recusada com `dados_invalidos` e o motivo.
    - **Regras de cada linha, na ordem de UC-01 `Registrar`:** formato; RN-01 (cliente existe, está ativo e tem consentimento válido: `cliente_nao_encontrado`, `cliente_inativo`, `consentimento_invalido`); RN-04 (`dados_invalidos`, com o motivo do campo); duplicado.
    - **Duplicado.** Mesmo cliente, tipo, instante (UTC, em microssegundos) e valor de um evento já gravado ou de uma linha anterior aceita no mesmo lote. É ignorado, sem erro, e contado à parte; categoria e origem não entram na comparação. Assim, reenviar o mesmo lote não grava nada de novo.
    - **Prévia e confirmação.** A prévia aplica todas as regras e não grava nada. A confirmação aplica tudo de novo (o consentimento ou os eventos podem ter mudado depois da prévia) e grava cada linha aceita por UC-01 `Registrar`. O relatório traz, para cada linha não aceita, o número, a situação (duplicada ou rejeitada), o ID do cliente quando legível, o código do erro e o motivo; e os totais de aceitas, duplicadas e rejeitadas.
    - **Falha de infraestrutura** ao gravar interrompe o lote com o erro (500): as linhas já gravadas permanecem, e o reenvio as reconhece como duplicadas.
    - **Sem guardar o arquivo.** Nem o arquivo nem o lote ficam guardados no servidor; só os eventos aceitos. O registro do lote é a auditoria: operador, quantidades aceitas, duplicadas e rejeitadas, sem ID de cliente nem dado pessoal (C-09).
- **RN-21 Públicos para campanha (v1.11, UC-06 parcial, Q-62 a Q-64).**
    - **Quem entra:** o público de um segmento ativo é formado pelos clientes cujo perfil está hoje nesse segmento (`PerfilCliente.SegmentoID`), que estão ativos e que têm consentimento CONCEDIDO para `segmentacao` **e** para `publicidade_terceiros`. Quem não cumpre fica de fora, sem erro; a contagem de quem ficou de fora não identifica ninguém.
    - **Mínimo de dados:** o arquivo leva só o `HashEmail` de cada cliente (SHA-256 do e-mail normalizado: sem espaços e em minúsculas) e o nome do segmento. Nunca nome, e-mail em claro, ID, eventos, valores nem o perfil. O hash pseudonimiza, não anonimiza: continua dado pessoal (LGPD art. 12 e art. 13 §4º).
    - **Formato do destino `arquivo_csv`:** UTF-8, separador vírgula, cabeçalho `email_sha256,segmento`, uma linha por hash, em ordem crescente, sem repetição; nome `publico-<segmento>-<inclusao|remocao>-<aaaammddThhmmssZ>.csv`. O destino fica atrás do port `DestinoPublicitarioGateway`, para trocar por uma API de plataforma depois sem mudar os casos de uso.
    - **Exportar:** público vazio não gera arquivo (`ErrDadosInvalidos`). Cada exportação grava um `ExportacaoPublico` (quando, operador, segmento, destino, quantidade, hash e nome do arquivo) e um `MembroPublico` EXPORTADO por cliente, numa única operação atômica. Antes de exportar, os membros EXPORTADO do mesmo segmento que não entram mais (revogaram, ficaram inativos ou mudaram de segmento) passam a REMOCAO_PENDENTE.
    - **Remoção propagada (LGPD art. 18 §6º):** revogar `publicidade_terceiros` ou `segmentacao` passa todos os membros EXPORTADO do cliente a REMOCAO_PENDENTE (`PropagarRevogacaoUseCase`, chamado por UC-08 `Revogar` logo depois de gravar a revogação). O arquivo de remoção de um segmento leva os hashes REMOCAO_PENDENTE desse segmento, no mesmo formato, e passa esses membros a REMOVIDO. Sem remoção pendente, não há arquivo (`ErrDadosInvalidos`). A propagação não fica na mesma transação da revogação: se falhar, a revogação vale do mesmo jeito, o erro vai para o log, e a próxima exportação do segmento marca a remoção (regra acima).
    - **Retenção do arquivo (Q-64):** o conteúdo fica guardado por `PUBLICOS_ARQUIVO_HORAS` (padrão 24) para ser baixado; depois disso não é mais entregue (`ErrArquivoIndisponivel`) e é apagado na próxima exportação ou tentativa de download. O registro da exportação e os membros ficam como prova (eliminação segue D-13).
    - **Logs (C-09):** `publico_exportado` com operador, tipo, `segmentoId`, `exportacaoId`, quantidade e hash do arquivo; `remocao_propagada` com `clienteId` e quantidade; nunca e-mail, hash de e-mail nem nome.
- **RN-22 Relatório agregado por segmento (v1.11, Q-65).** Para cada segmento ativo, em ordem de Prioridade: quantos clientes ativos com `segmentacao` válida estão nele, a média dos seus `TicketMedio` e a `CategoriaPreferida` mais comum (empate por ordem alfabética; vazias não contam). Não leva nenhum identificador, por isso pode ser exportado sem `publicidade_terceiros`. Grupo com menos de **5** clientes (`TamanhoMinimoGrupo`) aparece só como "menos de 5", sem quantidade exata, ticket nem categoria, para não reidentificar ninguém; não há linha de total.

### 3.4 Máquinas de estado

#### 3.4.1 ConsentimentoLGPD

| De | Evento | Guarda | Para | Efeito |
|---|---|---|---|---|
| NAO_CONCEDIDO | Conceder | Finalidade e VersaoTermo não vazias | CONCEDIDO | DataConsentimento = agora |
| CONCEDIDO | Revogar | — | REVOGADO | DataRevogacao = agora; interrompe o processamento (RN-03) |
| REVOGADO | Conceder | Finalidade e VersaoTermo não vazias | CONCEDIDO | DataConsentimento = agora |

DataAtualizacao = agora em toda transição; cada transição aplicada por UC-08 gera um movimento (RN-10: Conceder gera CONCESSAO, Revogar gera REVOGACAO). Qualquer outra combinação de estado e evento: `ErrTransicaoInvalida`. Guarda violada: `ErrDadosInvalidos`. Não existe estado final (D-03) e `EstaValido()` é uma consulta, não uma transição.

Desde a v1.11 (Q-61), a máquina vale **por finalidade**: cada registro nasce com a sua (`NovoConsentimento`), e a guarda de Conceder passa a ser "finalidade igual à do registro e dentro do enumerado; versão do termo não vazia". As duas finalidades de um cliente mudam de estado de forma independente.

#### 3.4.3 MembroPublico [ADIÇÃO, v1.11]

| De | Evento | Guarda | Para | Efeito |
|---|---|---|---|---|
| (início) | Incluir (`NovoMembro`) | hash não vazio | EXPORTADO | DataInclusao = agora |
| EXPORTADO | MarcarRemocao | — | REMOCAO_PENDENTE | DataRemocaoPendente = agora |
| REMOCAO_PENDENTE | Remover | exportação de remoção informada | REMOVIDO | RemocaoID e DataRemocao = agora |
| REMOVIDO | — | — | (final) | — |

Qualquer outra combinação: `ErrTransicaoInvalida`. Guarda violada: `ErrDadosInvalidos`. Um cliente que volta a se qualificar entra de novo como membro novo, na próxima exportação de inclusão.

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
| `ErrCredenciaisInvalidas` [v1.5] | Login inexistente, senha errada ou operador inativo (RN-12); senha atual errada em Minha senha | 401 (painel) | nao_autenticado |
| `ErrSemPermissao` [v1.5] | Papel sem a permissão (RN-11) | 403 (painel) | sem_permissao |
| `ErrOperadorNaoEncontrado` [v1.5] | Operador inexistente | 404 (painel) | operador_nao_encontrado |
| `ErrLoginEmUso` [v1.5] | Login já usado por outro operador (RN-12) | 409 (painel) | login_em_uso |
| `ErrPropriaConta` [v1.5] | Administrador muda o próprio papel ou desativa a si mesmo (RN-12) | 409 (painel) | propria_conta |
| `ErrSegmentoNaoEncontrado` [v1.6] | Segmento inexistente na edição (o processamento continua com `ErrSegmentoMalConfigurado`, Q-17) | 404 (painel) | segmento_nao_encontrado |
| `ErrProdutoNaoEncontrado` [v1.6] | Produto inexistente | 404 (painel) | produto_nao_encontrado |

| `ErrPedidoNaoEncontrado` [v1.10] | Uso interno: cliente sem pedido de eliminação aberto | — | — |
| `ErrArquivoIndisponivel` [ADIÇÃO, v1.11] | Arquivo de público inexistente, já apagado ou com o prazo de `PUBLICOS_ARQUIVO_HORAS` vencido (Q-64) | 410 | arquivo_indisponivel |
`errs.ErroCampo` [v1.6, RN-14] é um erro de validação com `Campo` e `Mensagem` (em português, para a tela); embrulha `ErrDadosInvalidos`, então o código continua `dados_invalidos` (422).

**UC-09 `CadastrarClienteUseCase`** [ADIÇÃO, D-22] (não exige consentimento: é o ponto de entrada do cliente)

- `Executar(ctx, nome, email string) (Cliente, error)`: `NovoCliente` (nome e e-mail obrigatórios, `ErrDadosInvalidos`), ID novo, Ativo, DataCadastro = agora; persiste com `ClienteRepository.Salvar`.
- O cliente nasce sem registro de consentimento (equivale a NAO_CONCEDIDO, D-02): nada além de nome e e-mail é tratado até UC-08 Conceder.

**UC-08 `GerenciarConsentimentoUseCase`** (não exige consentimento válido, pois é quem o gerencia)

- `Conceder(ctx, clienteID, finalidade entity.Finalidade, versaoTermo string) error`: finalidade fora do enumerado: `ErrDadosInvalidos`; cria o registro daquela finalidade se não existir, aplica a transição 3.4.1, monta o movimento CONCESSAO (`NovoMovimento`) e persiste os dois com `ConsentimentoRepository.Registrar` (v1.4, RN-10; antes `Salvar`).
- `Revogar(ctx, clienteID, finalidade entity.Finalidade) error` (a v1.11 acrescenta a finalidade, Q-61): idem, com o movimento REVOGACAO; depois de gravar, chama `PropagarRevogacaoUseCase.Executar` (RN-21). Falha da propagação vai para o log e não desfaz a revogação.
- `Consultar(ctx, clienteID, finalidade entity.Finalidade) (ConsentimentoLGPD, error)` [ADIÇÃO]: sem registro, devolve um consentimento sintético em NAO_CONCEDIDO daquela finalidade.
- `ConsultarFinalidades(ctx, clienteID) ([]ConsentimentoLGPD, error)` [ADIÇÃO, v1.11]: um registro por finalidade, na ordem de `Finalidades()`, sintético onde não houver.
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

**Pré-condições RN-01** ficam em `Precondicoes.Verificar(ctx, clienteID) (Cliente, error)`, compartilhado por UC-01 a UC-04; desde a v1.11 conferem a finalidade `segmentacao` (fecha L-04).

**UC-06 parcial — públicos para campanha** [ADIÇÃO, v1.11, RN-21, RN-22]. A origem e o operador vêm do contexto (`usecase.ComOrigem`, como na RN-10).

- `GerarPublicoUseCase.Executar(ctx, segmentoID) (PublicoGerado, error)`: segmento inexistente ou inativo: `ErrDadosInvalidos`. Aplica a seleção da RN-21 sobre `PerfilRepository.ListarPorSegmento` e devolve `PublicoGerado` com `Segmento`, `Membros` (`[]CandidatoPublico`: `ClienteID` e `HashEmail`, em ordem de hash), `Fora` (perfis do segmento que ficaram de fora) e `RemocoesPendentes` (membros REMOCAO_PENDENTE do segmento). Não grava nada: é a prévia.
- `ExportarPublicoUseCase.Exportar(ctx, segmentoID) (ExportacaoPublico, error)`: gera o público (acima), marca a remoção de quem não entra mais, monta o `LotePublico` de inclusão, envia ao `DestinoPublicitarioGateway` e grava exportação, conteúdo e membros com `PublicoRepository.RegistrarExportacao`. Público vazio: `ErrDadosInvalidos`.
- `ExportarPublicoUseCase.ExportarRemocoes(ctx, segmentoID) (ExportacaoPublico, error)`: lote de remoção com os hashes REMOCAO_PENDENTE do segmento; grava a exportação e passa os membros a REMOVIDO na mesma operação. Sem pendências: `ErrDadosInvalidos`.
- `ExportarPublicoUseCase.Baixar(ctx, exportacaoID) (ReciboEnvio, error)`: nome e conteúdo do arquivo dentro do prazo; senão `ErrArquivoIndisponivel`.
- `ExportarPublicoUseCase.Historico(ctx, segmentoID) ([]ExportacaoPublico, error)`: as exportações do segmento, da mais recente para a mais antiga.
- `PropagarRevogacaoUseCase.Executar(ctx, clienteID) (int, error)`: aplica `MarcarRemocao` a todos os membros EXPORTADO do cliente e grava; devolve quantos marcou (zero não é erro).
- `RelatorioAgregadoSegmentosUseCase.Executar(ctx) ([]LinhaAgregada, error)`: RN-22. `CSV(linhas) []byte` monta o arquivo com o cabeçalho `segmento,clientes,ticket_medio,categoria_mais_comum` (grupo oculto: `menos de 5` e os outros campos vazios).

**`ExpirarRecomendacoesUseCase`** [ADIÇÃO, D-08]

- `Executar(ctx) (int, error)`: lê `RecomendacaoRepository.ListarAtivasVencidas(agora)`, aplica `Expirar` e salva cada uma; devolve quantas expirou. É o que o subcomando `expirar-recomendacoes` chama.

**`ProcessarClienteUseCase`** [ADIÇÃO, Q-26]

- `Executar(ctx, clienteID) (ResultadoProcessamento, error)`: RN-09. `ResultadoProcessamento` traz `Eventos` (quantos UC-01 coletou), `Perfil`, `Segmento`, `Recomendacao *RecomendacaoProduto` (nula sem produto elegível) e `Etapas`, as quatro na ordem (`coletar`, `perfil`, `segmento`, `recomendacao`), cada uma com `Estado` (`EstadoEtapa`: `concluida`, `atencao`, `falhou`, `pendente`) e o erro que a parou. O `error` devolvido é o da etapa que falhou (o mesmo de antes para a API).

**Casos de uso de leitura do painel** [ADIÇÃO, Q-21, Q-23]

- `ListarClientesUseCase.Executar(ctx, FiltroListagem) (PaginaClientes, error)`: `FiltroListagem` tem `Busca` (nome ou ID), `Consentimento` (`*StatusConsentimento`, opcional) e `Pagina` (a partir de 1; 20 por página). `PaginaClientes` traz `Linhas` (cliente, estado do consentimento e nome do segmento, este só com cliente ativo e consentimento válido), `Total` e `Pagina`. O filtro de consentimento é aplicado pelo caso de uso sobre a busca do repositório (volume do MVP).
- `ConsultarClienteUseCase.Executar(ctx, clienteID) (FichaCliente, error)`: cadastro, consentimento (sem registro: sintético em NAO_CONCEDIDO, como UC-08) e, desde a v1.4, `Historico` (os movimentos de UC-08 `Historico`, RN-10). Desde a v1.11, `Consentimento` é o de `segmentacao` e `Finalidades` traz os dois (UC-08 `ConsultarFinalidades`). Não exige consentimento (Q-21); `ErrClienteNaoEncontrado`.
- `ConsultarResultadoUseCase.Executar(ctx, clienteID) (ResultadoCliente, error)`: cliente existe e consentimento válido primeiro (`Precondicoes.VerificarLeitura`: RN-01 sem a exigência de cliente ativo, Q-31; cliente inativo é só leitura); depois eventos (mais recente primeiro), perfil e segmento (nulos se ainda não processado), recomendações (`ListarPorCliente`, mais recente primeiro) e os produtos que elas citam. Nada de comportamento é lido sem consentimento válido (C-03).
- `ListarSegmentosUseCase.Executar(ctx) ([]Segmento, error)`: segmentos ativos em ordem de Prioridade.
- `ListarCatalogoUseCase.Executar(ctx) ([]Produto, error)`: todos os produtos, por Categoria e depois por Nome.
- [v1.6] `ListarSegmentosUseCase.Todos(ctx) ([]Segmento, error)`: todos, os ativos primeiro em ordem de Prioridade, depois os inativos por Nome; `ListarSegmentosUseCase.Consultar(ctx, id) (Segmento, error)`: `ErrSegmentoNaoEncontrado`. `ListarCatalogoUseCase.Consultar(ctx, id) (Produto, error)`: `ErrProdutoNaoEncontrado`; `ListarCatalogoUseCase.Categorias(ctx) ([]string, error)`: as categorias distintas do catálogo, em ordem alfabética.

**Configuração de segmentos e catálogo** [ADIÇÃO, v1.6, Q-33, Q-42, Q-43]. Exigem do `ator` a permissão `editar_configuracao` (RN-11, `ErrSemPermissao`) e aplicam a RN-14.

- `SalvarSegmentoUseCase.Executar(ctx, ator, s Segmento) (AlteracaoSegmento, error)`: `s.ID` nulo cria (ID novo, ativo); senão edita nome, descrição, critério e prioridade do segmento existente (`ErrSegmentoNaoEncontrado`), mantendo a situação. Tira os espaços das pontas do nome, do critério e da descrição, aplica `Validar` e `ValidarSegmentos` sobre `ListarTodos` com a mudança e grava com `SegmentoRepository.Salvar`. `AlteracaoSegmento` traz `Segmento`, `Novo` e `Campos` (nomes dos campos que mudaram, na ordem da RN-13); sem nenhuma mudança, nada é gravado e `Campos` vem vazio.
- `SalvarSegmentoUseCase.Ativar(ctx, ator, id) (AlteracaoSegmento, error)` e `Desativar(ctx, ator, id) (AlteracaoSegmento, error)`: mudam só a situação, com as mesmas validações (o `padrao` não é desativado; ativar exige prioridade livre).
- `PreverSegmentacaoUseCase.Executar(ctx, ator, s Segmento) (PreviaSegmentacao, error)`: valida como `SalvarSegmento.Executar` (o mesmo erro, nada gravado) e calcula a prévia da RN-14 com `ConsentimentoRepository.ListarConcedidos`, `EstaValido`, `ClienteRepository.BuscarPorID` (só ativos), `PerfilRepository.BuscarPorCliente` e `EscolherSegmento`. `PreviaSegmentacao` traz `Linhas` (cada segmento ativo antes ou depois da mudança, na ordem da nova configuração, com `Segmento`, `Atual` e `Previsto`), `SemPerfil` e `Total` (clientes considerados). `s` com o campo `Ativo` diferente do gravado prevê a ativação ou a desativação.
- `PreverSegmentacaoUseCase.TestarCriterio(criterio string, frequenciaCompras int, ticketMedio, scoreEngajamento float64) (bool, error)`: RN-14; sem `ctx`, não faz I/O; critério inválido: `errs.ErroCampo`. Não exige `ator`: o painel confere a permissão na rota.
- `SalvarProdutoUseCase.Executar(ctx, ator, p Produto) (AlteracaoProduto, error)`: `p.ID` nulo cria (ID novo, ativo); senão edita nome, categoria e preço (`AtualizarPreco`) do produto existente (`ErrProdutoNaoEncontrado`); `NormalizarCategoria` e `Validar`; grava com `ProdutoRepository.Salvar`. `AlteracaoProduto` traz `Produto`, `Novo` e `Campos`. `Ativar` e `Desativar(ctx, ator, id) (AlteracaoProduto, error)` mudam só a situação.

**Operadores e auditoria** [ADIÇÃO, v1.5, Q-35 a Q-40]. O `ator` é o operador da sessão, já conferido; os casos de uso de gestão conferem de novo a permissão dele (RN-11, `ErrSemPermissao`).

- `AutenticarOperadorUseCase.Executar(ctx, login, senha string) (Operador, error)`: RN-12 (Entrar); `ErrCredenciaisInvalidas`. Com o login existente e a senha errada ou o operador inativo, o operador vem junto com o erro, só para a auditoria registrar o login (RN-13); nunca vale como entrada.
- `AutenticarOperadorUseCase.Conferir(ctx, id uuid.UUID, carimbo time.Time) (Operador, error)`: o operador da sessão existe, está ativo e tem `SenhaAtualizadaEm` igual ao carimbo (comparado em microssegundos, a precisão do PostgreSQL); senão, `ErrCredenciaisInvalidas`. Devolve o operador com o papel atual (Q-38).
- `GerenciarOperadoresUseCase` (permissão `usuarios`, menos `TrocarSenha` e `CriarAdministradorInicial`):
    - `Listar(ctx, ator) ([]Operador, error)` e `Consultar(ctx, ator, id) (Operador, error)` (`ErrOperadorNaoEncontrado`).
    - `Criar(ctx, ator, login, nome string, papel PapelOperador, senha string) (Operador, error)`: RN-12; nasce com `TrocarSenha`; `ErrLoginEmUso`.
    - `MudarPapel(ctx, ator, id, papel) error`, `Desativar(ctx, ator, id) error`, `Ativar(ctx, ator, id) error`: `ErrPropriaConta` quando `id` é o do ator (menos em `Ativar`, que nunca se aplica a si mesmo).
    - `RedefinirSenha(ctx, ator, id, senha string) error`: senha provisória, `TrocarSenha` e novo `SenhaAtualizadaEm`.
    - `TrocarSenha(ctx, ator, atual, nova string) (Operador, error)`: qualquer papel, só a própria; `ErrCredenciaisInvalidas` se `atual` não confere. Devolve o operador com o novo carimbo, para o painel renovar a sessão de quem trocou.
    - `CriarAdministradorInicial(ctx, login, senhaHash string) (bool, error)`: RN-12; devolve se criou.
- `RegistrarAuditoriaUseCase.Executar(ctx, operador string, acao AcaoAuditoria, alvoTipo TipoAlvo, alvoID uuid.UUID, resultado ResultadoAuditoria, campos ...string) error`: RN-13; ID novo e Data do `Clock`. `campos` desde a v1.6 (Q-44).
- `ListarAuditoriaUseCase.Executar(ctx, ator, FiltroListagemAuditoria) (PaginaAuditoria, error)`: permissão `auditoria`. O filtro tem `Operador`, `Acao`, `De` e `Ate` (instantes; vazios = sem limite; `De` ≤ Data < `Ate`) e `Pagina` (a partir de 1; 50 por página). `PaginaAuditoria` traz `Registros` (mais recente primeiro), `Total` e `Pagina`.
- `ExpurgarAuditoriaUseCase.Executar(ctx) (int, error)`: RN-13; apaga os registros com Data anterior a agora menos `AUDITORIA_RETENCAO_DIAS` e devolve quantos. É o que o subcomando `expurgar-auditoria` chama.

**`GerarRelatorioUseCase`** [ADIÇÃO, v1.7, Q-46]

- `Executar(ctx, periodoMeses int) (Relatorio, error)`: RN-15. Período 6 ou 12 (senão `ErrDadosInvalidos`); calcula os limites dos meses no fuso injetado pelo construtor (America/Sao_Paulo) a partir do `Clock`; lê os agregados de `RelatorioRepository` e os nomes de `SegmentoRepository` (`ListarAtivos` e, para segmento inativo com perfis, `BuscarPorID`); aplica `ProtegerContagem`, `ProtegerMedia` e `ProtegerPercentual` e junta as categorias pequenas.
- `Relatorio` traz `GeradoEm`, `PeriodoMeses`, `Desde` e `Ate` (início do primeiro mês e o instante do relatório); `Clientes` e `Ativos`; `Consentimento` (uma linha por estado, na ordem Concedido, Não concedido, Revogado) e `TaxaConcessao`; `Segmentos` (nome, clientes, ticket médio e score médio; nome vazio = ainda não segmentados); `RecomendacoesAtivas` e `RecomendacoesExpiradas`; `Categorias` (nome, compras e clientes; `Outras` com o número de categorias somadas); `Meses` (início do mês, cadastros, eventos e revogações). Todos os números são `Contagem`, `Media` ou `Percentual` já protegidos: nenhum número cru sai do caso de uso.
- Não verifica consentimento por cliente (não lê nenhum cliente): a base de cada indicador está no contrato de `RelatorioRepository`.

**`ImportarEventosUseCase`** [ADIÇÃO, v1.8, Q-47]

- `LerCSVEventos(conteudo []byte) ([]EntradaEvento, error)` (função do pacote `usecase`): lê o arquivo de RN-16 e devolve uma `EntradaEvento` por linha de dados, com `Numero` (linha no arquivo) e os campos em texto (`ClienteID`, `Tipo`, `Origem`, `Timestamp`, `Valor`, `Categoria`); a linha com número de colunas errado vem com `ColunasErradas`. Arquivo recusado: `*ErroArquivo` com `Motivo` (`MotivoArquivo`: `arquivo_vazio`, `arquivo_grande`, `linhas_demais`, `codificacao_invalida`, `cabecalho_invalido`, `formato_invalido`), que embrulha `ErrDadosInvalidos`.
- `Previa(ctx, entradas []EntradaEvento) (ResultadoImportacao, error)`: RN-16 sem gravar. `Importar(ctx, entradas []EntradaEvento) (ResultadoImportacao, error)`: RN-16, gravando cada linha aceita com UC-01 `Registrar`. Lote vazio ou com mais de 1.000 entradas: `*ErroArquivo` (`arquivo_vazio` ou `linhas_demais`).
- `ResultadoImportacao` traz `Linhas` (só as não aceitas, na ordem do lote: `LinhaImportacao` com `Numero`, `ClienteID` (nulo se ilegível), `Situacao` (`SituacaoLinha`: `duplicada` ou `rejeitada`), `Erro` (a sentinela, para `errs.Codigo`) e `Motivo` (`cliente_id`, `tipo`, `origem`, `timestamp`, `timestamp_futuro`, `valor`, `valor_compra`, `colunas`; vazio fora de `dados_invalidos`)), `Aceitas`, `Duplicadas`, `Rejeitadas` e `Gravado` (verdadeiro em `Importar`).
- O motivo de `dados_invalidos` em RN-04 só detalha a mensagem: quem decide é `Validar`.

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
    ListarPorSegmento(ctx context.Context, segmentoID uuid.UUID) ([]entity.PerfilCliente, error) // [ADIÇÃO] v1.11 RN-21, RN-22: ordem por ClienteID
}
type SegmentoRepository interface {
    ListarAtivos(ctx context.Context) ([]entity.Segmento, error)            // ordem crescente de Prioridade
    BuscarPorID(ctx context.Context, id uuid.UUID) (entity.Segmento, error) // [ADIÇÃO] ErrDadosInvalidos se inexistente
    ListarTodos(ctx context.Context) ([]entity.Segmento, error)             // [ADIÇÃO] v1.6: ativos e inativos, Prioridade crescente, desempate por ID
    Salvar(ctx context.Context, s entity.Segmento) error                    // [ADIÇÃO] v1.6: insere ou atualiza por ID; nome repetido (sem diferenciar maiúsculas) ou prioridade repetida entre ativos: errs.ErroCampo
}
type RecomendacaoRepository interface {
    Salvar(ctx context.Context, r entity.RecomendacaoProduto) error // insere ou atualiza por ID
    ListarAtivasVencidas(ctx context.Context, agora time.Time) ([]entity.RecomendacaoProduto, error) // [ADIÇÃO]
    ListarAtivasPorCliente(ctx context.Context, clienteID uuid.UUID) ([]entity.RecomendacaoProduto, error) // [ADIÇÃO] Q-18
    ListarPorCliente(ctx context.Context, clienteID uuid.UUID) ([]entity.RecomendacaoProduto, error) // [ADIÇÃO] Q-23: DataGeracao decrescente
    RemoverPorCliente(ctx context.Context, clienteID uuid.UUID) (int, error) // [ADIÇÃO] v1.9 Q-52: quantas removeu
}
type ConsentimentoRepository interface { // v1.11 (Q-61): um registro por (ClienteID, Finalidade)
    Salvar(ctx context.Context, c entity.ConsentimentoLGPD) error // upsert por (ClienteID, Finalidade), sem movimento (cargas e testes)
    BuscarPorCliente(ctx context.Context, clienteID uuid.UUID, finalidade entity.Finalidade) (entity.ConsentimentoLGPD, error) // ErrConsentimentoNaoEncontrado
    ListarPorCliente(ctx context.Context, clienteID uuid.UUID) ([]entity.ConsentimentoLGPD, error) // [ADIÇÃO] v1.11: só os que existem, por Finalidade
    Registrar(ctx context.Context, c entity.ConsentimentoLGPD, m entity.MovimentoConsentimento) error // [ADIÇÃO] v1.4 RN-10: upsert + inclusão do movimento, atômico; ID de movimento repetido falha sem gravar nada
    ListarMovimentos(ctx context.Context, clienteID uuid.UUID) ([]entity.MovimentoConsentimento, error) // [ADIÇÃO] v1.4: Data crescente, desempate pela ordem de gravação
    ListarConcedidos(ctx context.Context) ([]entity.ConsentimentoLGPD, error) // [ADIÇÃO] v1.6 Q-42: registros com Status CONCEDIDO, por ClienteID
    ListarRevogadosAte(ctx context.Context, limite time.Time) ([]uuid.UUID, error) // [ADIÇÃO] v1.9 Q-51: clientes REVOGADO com DataRevogacao <= limite, da revogação mais antiga para a mais recente
}
type ProdutoRepository interface { // [ADIÇÃO] (D-05)
    ListarAtivosPorCategoria(ctx context.Context, categoria string) ([]entity.Produto, error)
    ListarTodos(ctx context.Context) ([]entity.Produto, error) // [ADIÇÃO] Q-23: por Categoria e Nome
    BuscarPorID(ctx context.Context, id uuid.UUID) (entity.Produto, error) // [ADIÇÃO] v1.6: ErrProdutoNaoEncontrado
    Salvar(ctx context.Context, p entity.Produto) error                    // [ADIÇÃO] v1.6: insere ou atualiza por ID
}
type RelatorioRepository interface { // [ADIÇÃO] v1.7, Q-46, RN-15: só agregados, nunca linhas de cliente
    ContarClientes(ctx context.Context) (ContagemClientes, error)
    ResumirSegmentos(ctx context.Context) ([]ResumoSegmento, error)          // consentimento válido; um item por SegmentoID (nil = sem segmento)
    ContarRecomendacoes(ctx context.Context) (ContagemRecomendacoes, error)  // consentimento válido
    ContarCompras(ctx context.Context, desde, ate time.Time) ([]ComprasCategoria, error)                       // consentimento válido; categoria preenchida; [desde, ate)
    ContarCompradores(ctx context.Context, desde, ate time.Time, categorias []string) (int, error)             // clientes distintos que compraram em alguma das categorias
    SerieMensal(ctx context.Context, limites []time.Time) ([]ContagemMes, error)                                // um item por intervalo [limites[i], limites[i+1])
}
type ContagemClientes struct{ Total, Ativos, Concedido, NaoConcedido, Revogado int } // todos os clientes; sem registro = NaoConcedido
type ResumoSegmento struct {
    SegmentoID           *uuid.UUID
    Clientes             int
    SomaTicket, SomaScore float64
}
type ContagemRecomendacoes struct{ Ativas, ClientesAtivas, Expiradas, ClientesExpiradas int }
type ComprasCategoria struct {
    Categoria         string
    Compras, Clientes int
}
type ContagemMes struct {
    Inicio                                     time.Time
    Cadastros, Revogacoes, Eventos, ClientesEventos int // cadastros e revogações (movimentos REVOGACAO) de todos; eventos só com consentimento válido
}
type ConsentimentoGateway interface { // MVP: implementação local sobre ConsentimentoRepository; finalidade segmentacao (RN-01)
    Validar(ctx context.Context, clienteID uuid.UUID) (bool, error)
}
type PublicoRepository interface { // [ADIÇÃO] v1.11 RN-21
    RegistrarExportacao(ctx context.Context, e entity.ExportacaoPublico, conteudo []byte, membros []entity.MembroPublico) error // atômico: insere a exportação e o conteúdo, insere os membros novos e atualiza os existentes
    ListarExportacoes(ctx context.Context, segmentoID uuid.UUID) ([]entity.ExportacaoPublico, error) // Data decrescente; sem o conteúdo
    BuscarArquivo(ctx context.Context, exportacaoID uuid.UUID) (entity.ExportacaoPublico, []byte, error) // ErrArquivoIndisponivel se não houver ou se o conteúdo foi apagado
    ApagarArquivosVencidos(ctx context.Context, agora time.Time) (int, error) // apaga o conteúdo com ArquivoExpiraEm <= agora; o registro fica
    ListarMembrosPorSegmento(ctx context.Context, segmentoID uuid.UUID, status entity.StatusMembro) ([]entity.MembroPublico, error) // por HashEmail
    ListarMembrosPorCliente(ctx context.Context, clienteID uuid.UUID, status entity.StatusMembro) ([]entity.MembroPublico, error)
    AtualizarMembros(ctx context.Context, ms []entity.MembroPublico) error // só o estado e as datas; ID inexistente falha sem gravar nada
}
type DestinoPublicitarioGateway interface { // [ADIÇÃO] v1.11 RN-21; MVP: arquivo CSV (infra/external)
    Enviar(ctx context.Context, lote entity.LotePublico) (entity.ReciboEnvio, error)
}
type IntegracaoGateway interface { // e-commerce/CRM; MVP: stub que devolve lista vazia (D-12)
    BuscarEventos(ctx context.Context, clienteID uuid.UUID, desde time.Time) ([]entity.EventoComportamental, error)
}
type Clock interface{ Agora() time.Time }

// [ADIÇÃO, v1.5, Q-35, Q-36, Q-39]
type OperadorRepository interface {
    Salvar(ctx context.Context, o entity.Operador) error                         // insere ou atualiza por ID; Login de outro ID: ErrLoginEmUso
    BuscarPorID(ctx context.Context, id uuid.UUID) (entity.Operador, error)      // ErrOperadorNaoEncontrado
    BuscarPorLogin(ctx context.Context, login string) (entity.Operador, error)   // ErrOperadorNaoEncontrado
    Listar(ctx context.Context) ([]entity.Operador, error)                       // por Login
    CriarSeVazio(ctx context.Context, o entity.Operador) (bool, error)           // grava só se não houver nenhum operador; atômico
}
type AuditoriaRepository interface {
    Registrar(ctx context.Context, r entity.RegistroAuditoria) error                       // só inclusão; ID repetido falha
    Listar(ctx context.Context, f FiltroAuditoria) ([]entity.RegistroAuditoria, int, error) // Data decrescente, desempate pela gravação mais recente; página e total
    ExpurgarAntesDe(ctx context.Context, limite time.Time) (int, error)                     // apaga Data < limite
}
type FiltroAuditoria struct {
    Operador     string               // vazio = todos
    Acao         entity.AcaoAuditoria // vazio = todas
    De, Ate      time.Time            // zero = sem limite; De <= Data < Ate
    Deslocamento int
    Limite       int                  // 0 = sem limite
}
type HashSenha interface { // sem ctx: não faz I/O (C-07)
    Gerar(senha string) (string, error)
    Conferir(hash, senha string) bool
}
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
func EscolherSegmento(lista []entity.Segmento, p entity.PerfilCliente) (entity.Segmento, error) // [v1.6] RN-06 pura sobre uma lista; Segmentar e a prévia (Q-42) usam a mesma regra
func (s *RecomendacaoService) Recomendar(ctx context.Context, c entity.Cliente, p entity.PerfilCliente, seg entity.Segmento) (entity.RecomendacaoProduto, error)
func (v *ValidadorLGPD) ConsentimentoValido(ctx context.Context, clienteID uuid.UUID) (bool, error) // finalidade segmentacao (RN-01)
func (v *ValidadorLGPD) ConsentimentoValidoPara(ctx context.Context, clienteID uuid.UUID, f entity.Finalidade) (bool, error) // [ADIÇÃO] v1.11
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
| `PUT /v1/clientes/{clienteID}/consentimento` | UC-08 Conceder | finalidade (`segmentacao` ou `publicidade_terceiros`, v1.11) e versaoTermo | 200 | Autenticação |
| `POST /v1/clientes/{clienteID}/consentimento/revogacao` | UC-08 Revogar | sem corpo (= `segmentacao`) ou `{"finalidade": ...}` (v1.11) | 200 | Autenticação |
| `GET /v1/clientes/{clienteID}/consentimento` | UC-08 Consultar | — | 200 com status e valido (da `segmentacao`, como antes) e, desde a v1.11, `finalidades`: um item por finalidade com finalidade, status, valido, versaoTermo e as datas | Autenticação |
| `POST /v1/eventos/lote` [v1.8, Q-47] | `ImportarEventos` (RN-16) | `{"eventos": [{clienteId, tipo, origem, timestamp, valor, categoria}]}`, até 1.000; `?previa=true` só confere | 200 com `gravado`, `aceitas`, `duplicadas`, `rejeitadas` e `linhas` (`numero` a partir de 1, `clienteId`, `situacao`, `codigo`, `motivo`) | Autenticação (o caso de uso aplica RN-01 em cada linha) |
| `GET /saude` | — | — | 200 com `status: ok`; 503 com `status: indisponivel` se o banco não responder (Q-19) | Nenhum (usado pelo balanceador) |

- Autenticação por cabeçalho `Authorization: Bearer` com o valor de `AUTH_TOKEN` (D-16); ausente ou inválido responde 401.
- Middleware LGPD: extrai o clienteID da rota e consulta `ValidadorLGPD`; sem consentimento válido responde 403 antes do handler. Os casos de uso repetem a checagem (C-03).
- Conceder e revogar pela API gravam o movimento com origem `api` (RN-10); as respostas não mudam.
- Lote de eventos (v1.8, RN-16): no JSON, `timestamp` é texto RFC 3339 e `valor` é número; lote vazio ou com mais de 1.000 eventos responde 422 `dados_invalidos`. A resposta é 200 mesmo com linhas rejeitadas: o resultado de cada linha está no corpo. O log `importacao_eventos` registra só as quantidades.
- Finalidade (v1.11, Q-61): fora do enumerado, `dados_invalidos` (422). O middleware LGPD confere `segmentacao` (RN-01). Revogação sem corpo continua revogando `segmentacao`, para quem já usa a API.
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
| `GET` e `POST /painel/clientes/{id}/consentimento/conceder?finalidade=` | Confirmar e conceder (v1.11: a finalidade vem do botão; sem ela, `segmentacao`) | UC-08 Conceder |
| `GET` e `POST /painel/clientes/{id}/consentimento/revogar?finalidade=` | Confirmar e revogar (idem) | UC-08 Revogar |
| `GET /painel/clientes/{id}/direitos` | Ficha: direitos do titular [v1.9] | `ConsultarCliente` |
| `POST /painel/clientes/{id}/cadastro` | Corrigir cadastro [v1.9] | `AtualizarCadastro` |
| `GET /painel/clientes/{id}/dados` | Baixar dados (JSON) [v1.9] | `ExportarDadosTitular` |
| `GET` e `POST /painel/clientes/{id}/eliminar` | Confirmar e eliminar [v1.9] | `EliminarDadosTitular` |
| `GET /painel/segmentos`, `GET /painel/catalogo` | Configuração: listas (desde a v1.6, com os inativos e, para quem edita, as ações) | `ListarSegmentos.Todos`, `ListarCatalogo` |
| `GET /painel/segmentos/novo`, `GET /painel/segmentos/{id}` [v1.6] | Formulário do segmento | `ListarSegmentos.Consultar` |
| `POST /painel/segmentos/previa` [v1.6] | Prévia e confirmação (ou o formulário de volta com o erro ou com mais uma condição) | `PreverSegmentacao` |
| `POST /painel/segmentos` [v1.6] | Salvar (botão da prévia) | `SalvarSegmento.Executar` |
| `GET` e `POST /painel/segmentos/{id}/ativar`, `GET` e `POST /painel/segmentos/{id}/desativar` [v1.6] | Prévia e confirmação; ativar ou desativar | `PreverSegmentacao`, `SalvarSegmento.Ativar` e `Desativar` |
| `POST /painel/segmentos/testar` [v1.6] | Testar critério | `PreverSegmentacao.TestarCriterio` |
| `GET /painel/catalogo/novo`, `GET /painel/catalogo/{id}` [v1.6] | Formulário do produto | `ListarCatalogo.Consultar`, `Categorias` |
| `POST /painel/catalogo` [v1.6] | Salvar produto | `SalvarProduto.Executar` |
| `POST /painel/catalogo/{id}/ativar`, `GET` e `POST /painel/catalogo/{id}/desativar` [v1.6] | Ativar; confirmar e desativar | `SalvarProduto.Ativar` e `Desativar` |
| `GET /painel/operacao`, `GET` e `POST /painel/operacao/expirar` | Saúde e expiração | `ExpirarRecomendacoes` |
| `GET` e `POST /painel/operacao/retencao` | Aplicar retenção [v1.9] | `AplicarRetencao` |
| `GET /painel/relatorios?periodo=6\|12` | Relatórios (v1.7): indicadores, gráficos e tabelas | `GerarRelatorio` |
| `GET /painel/relatorios.csv?periodo=6\|12` | Relatório agregado em CSV (v1.7) | `GerarRelatorio` |
| `GET /painel/eventos/importar` | Importar eventos: enviar o arquivo (v1.8) | — |
| `POST /painel/eventos/importar` | Prévia: totais e linhas não aceitas, sem gravar (campo `arquivo`, `multipart/form-data`) | `ImportarEventos` `Previa` |
| `POST /painel/eventos/importar/confirmar` | Grava e mostra o relatório (campo oculto `conteudo`) | `ImportarEventos` `Importar` |
| `GET /painel/eventos/modelo.csv` | Modelo do arquivo para baixar | — |
| `GET /painel/pedidos` | Pedidos do titular [v1.10] | `ListarPedidosEliminacao` |
| `GET /painel/operacao/caixa-saida` | Caixa de saída (simulada) [v1.10] | `ListarCaixaSaida` |
| `GET /painel/publicos?segmento=` | Públicos (v1.11): escolher o segmento, ver quantos entram e quantos ficam de fora, remoções pendentes e histórico de exportações | `ListarSegmentos`, `GerarPublico`, `ExportarPublico.Historico` |
| `GET` e `POST /painel/publicos/exportar?segmento=` | Confirmar e gerar o arquivo de inclusão | `ExportarPublico.Exportar` |
| `GET` e `POST /painel/publicos/remocoes?segmento=` | Confirmar e gerar o arquivo de remoção | `ExportarPublico.ExportarRemocoes` |
| `GET /painel/publicos/exportacoes/{id}/arquivo` | Baixar o arquivo (CSV, `Content-Disposition: attachment`) | `ExportarPublico.Baixar` |
| `GET /painel/publicos/relatorio`, `GET /painel/publicos/relatorio.csv` | Relatório agregado por segmento, na tela e em CSV | `RelatorioAgregadoSegmentos` |
| `GET /painel/static/{arquivo}` | CSS, JS, fontes | — |
| `GET` e `POST /painel/minha-senha` [v1.5] | Minha senha | `GerenciarOperadores.TrocarSenha` |
| `GET /painel/usuarios` [v1.5] | Usuários | `GerenciarOperadores.Listar` |
| `GET /painel/usuarios/novo`, `POST /painel/usuarios` [v1.5] | Criar operador | `GerenciarOperadores.Criar` |
| `GET /painel/usuarios/{id}` [v1.5] | Operador: papel, situação e senha | `GerenciarOperadores.Consultar` |
| `POST /painel/usuarios/{id}/papel` [v1.5] | Mudar papel | `GerenciarOperadores.MudarPapel` |
| `GET` e `POST /painel/usuarios/{id}/desativar` [v1.5] | Confirmar e desativar | `GerenciarOperadores.Desativar` |
| `POST /painel/usuarios/{id}/ativar` [v1.5] | Ativar | `GerenciarOperadores.Ativar` |
| `POST /painel/usuarios/{id}/senha` [v1.5] | Redefinir senha | `GerenciarOperadores.RedefinirSenha` |
| `GET /painel/auditoria?operador=&acao=&de=&ate=&pagina=` [v1.5] | Auditoria | `ListarAuditoria` |

- **Permissão por rota (v1.5, RN-11, Q-37):** `clientes` para a lista, o cadastro e as abas da ficha (GET); `processamento`, `eventos` (POST) e `consentimento` (GET e POST de conceder e revogar) para as ações; `configuracao` para Segmentos e Catálogo; `editar_configuracao` (v1.6) para os formulários, a prévia, o teste de critério e as ações de segmento e produto; `operacao`, `usuarios` e `auditoria` para as telas de mesmo nome. Entrar, Sair, Minha senha e estáticos não exigem permissão. Sem a permissão: 403 com a tela de erro "Sem permissão" e o registro `acesso_negado` (RN-13); nada é lido nem gravado.
- **Menu e botões por papel (v1.5):** o menu mostra só as telas que o papel pode abrir; o rodapé mostra o login e o papel, "Minha senha" e "Sair". Ações sem permissão não aparecem. Depois de Entrar, ou em `/painel/`, quem tem `clientes` vai para Clientes e o ANALISTA vai para Segmentos.

- **Sessão (Q-22; v1.5, Q-38):** cookie `perfila_sessao` com `operadorID|login|carimbo|expiraEm|csrf` (o login vai para logs e auditoria) (o carimbo é `SenhaAtualizadaEm` em microssegundos) assinado por HMAC-SHA256 (`PAINEL_CHAVE_SESSAO`), `HttpOnly`, `Secure`, `SameSite=Strict`, `Path=/painel`, 8 h renovadas a cada uso; sem estado no servidor. A cada requisição, o painel confere o operador com `AutenticarOperador.Conferir`: inexistente, inativo ou com outra senha encerra a sessão, registra `sessao_encerrada` e leva a Entrar com "Sua sessão foi encerrada. Entre de novo."; o papel vem sempre do banco. Sem sessão válida, toda rota (menos Entrar e estáticos) leva a Entrar com o destino guardado; sessão vencida mostra "Sua sessão expirou." Em requisição htmx, a resposta é `HX-Redirect`. Com `TrocarSenha`, toda rota menos Minha senha e Sair leva a Minha senha.
- **Entrar:** `AutenticarOperador.Executar` (RN-12); erro genérico "Usuário ou senha incorretos."; 5 tentativas por IP em 5 minutos (em memória, por réplica), depois "Muitas tentativas. Aguarde 5 minutos." (429). Senha atual errada em Minha senha conta como tentativa.
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
- **Consentimento por finalidade (v1.11, Q-61):** a aba Consentimento mostra um registro por finalidade ("Segmentação" e "Publicidade com terceiros"), cada um com o selo, o termo, as datas, a ação (Conceder ou Revogar) e a sua linha do tempo. O cabeçalho da ficha mostra o selo de `segmentacao`, que é o que libera a ficha (RN-01), e uma etiqueta discreta para a publicidade. Os diálogos dizem o efeito de cada finalidade (`docs/design/03-textos.md`).
- **Públicos (v1.11, RN-21, Q-66):** gerar um arquivo de inclusão ou de remoção passa por confirmação com o efeito escrito (segmento, quantos clientes, que só vai o e-mail em hash, que continua sendo dado pessoal, e por quanto tempo o arquivo fica disponível). Depois de gerar, a tela volta com o link de download; o histórico mostra data, operador, tipo, quantidade, hash do arquivo e se o arquivo ainda está disponível. O download responde `Content-Type: text/csv; charset=utf-8` e `Cache-Control: no-store`. Os logs seguem a RN-21.
- **Sem JavaScript:** formulários comuns com redirecionamento 303 no sucesso (PRG) e confirmação em página própria. Com JavaScript, o htmx (`hx-boost`) faz as mesmas requisições e troca só a área do painel (`#app`), sem recarregar a página; os diálogos usam `<dialog>` nativo. Respostas de erro (4xx e 5xx) também são exibidas.
- **Importar eventos (v1.8, RN-16, Q-47):** item "Importar eventos" no grupo Clientes da navegação. O operador envia o CSV; a prévia mostra os totais (aceitas, duplicadas, rejeitadas) e a tabela das linhas não aceitas (linha, ID do cliente truncado, situação e a mensagem de `03-textos.md` pelo código e motivo), sem gravar; "Importar {n} eventos" confirma, com o conteúdo do arquivo num campo oculto (a página é `no-store` e o arquivo não fica no servidor). A confirmação mostra o relatório direto (sem PRG: reenviar só gera duplicados). Sem nenhuma linha aceita, não há botão de confirmar. Pedido acima de 512 KiB é recusado antes de ler o corpo, com redirecionamento para a tela com o aviso de arquivo grande. O log `painel_acao` com a ação `importar_eventos` leva operador e quantidades, nunca IDs de cliente nem conteúdo do arquivo. Funciona sem JavaScript; com htmx, o envio do arquivo usa o mesmo formulário.
- **Erros:** o código estável (`errs.Codigo`) escolhe a mensagem de `docs/design/03-textos.md`; o texto técnico nunca aparece. Falha inesperada mostra o código da requisição (`X-Request-ID` do Nginx).
- **Relatórios (v1.7, RN-15, Q-46):** grupo "Análise" na navegação. Indicadores no topo; gráficos de barras desenhados no servidor em SVG inline (sem biblioteca de JavaScript, CSP inalterada), cada um com `role="img"`, título e a tabela de dados logo abaixo, que é a fonte para leitores de tela. Cores: acento ameixa para dados derivados (segmentos, categorias, eventos), cores de estado do design system no consentimento e nas recomendações, neutro para cadastros. "Menos de 5" aparece por extenso na tabela, e a barra correspondente fica tracejada, sem altura que revele o número. Filtro de período (6 ou 12 meses) por formulário GET, funcionando sem JavaScript. O CSV usa `;` como separador, vírgula decimal, UTF-8 com BOM, `Content-Disposition: attachment` e colunas `secao;item;indicador;valor`; célula que começa com `=`, `+`, `-` ou `@` ganha um apóstrofo na frente. Ver ou exportar registra a ação `ver_relatorio` ou `exportar_relatorio` com o operador e o período, sem nenhum número.
- **Logs (C-09, Q-24):** método, caminho (só UUIDs) e status; nas ações, operador, ação e `clienteId`. Nunca nome, e-mail, senha nem o termo de busca. O nome do operador também não vai para o log nem para a auditoria. Desde a v1.9, `direito_atendido` com operador, pedido (`correcao`, `exportacao`, `eliminacao`, `retencao`) e `clienteId` (ou a quantidade, na retenção). É a auditoria dos pedidos até a v1.5 (`RegistroAuditoria`) entrar na main; aí passam a ser gravados nela e a eliminação fica restrita a ADMINISTRADOR (Q-53).
- **Segmentos e catálogo (v1.6, Q-33, Q-42, Q-43):** o formulário do segmento monta o critério por linhas de campo, operador e valor (uma condição por linha, ligadas por `&&`), com o texto do critério gravado visível acima; "Adicionar condição" devolve o formulário com mais uma linha (sem JavaScript, pelo `POST .../previa`; com JavaScript, na hora). Linha toda vazia é ignorada; linha incompleta é erro naquela linha. O botão principal leva à prévia; só a página da prévia tem "Salvar". A lista e a prévia avisam que os clientes já segmentados mudam só no próximo processamento. O catálogo lembra que a recomendação escolhe o primeiro produto ativo da categoria por nome (RN-07). Erros de validação aparecem ao lado do campo, com a mensagem do `errs.ErroCampo`.
- **Auditoria (v1.5, RN-13):** os handlers chamam `RegistrarAuditoria` depois de cada ação e de cada aba da ficha, com o resultado. A tela Auditoria mostra data, operador, ação, alvo (ID truncado, com link para a ficha quando é cliente) e resultado, com filtros por operador, ação e período (datas no fuso de Brasília).

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
| `PAINEL_USUARIO` | operador | Login do administrador inicial, usado só sem nenhum operador gravado (Q-35) |
| `PAINEL_SENHA_HASH` | vazia | Hash bcrypt do administrador inicial (`omega-server hash-senha`, Q-28, Q-35). Exigida só enquanto não houver operador: com a chave de sessão, sem operador e sem ela, o servidor não sobe e diz por quê. Sem a chave de sessão, é configuração inválida |
| `PAINEL_CHAVE_SESSAO` | vazia = painel desligado (Q-29, Q-35) | Chave HMAC da sessão: 64 dígitos hexadecimais (32 bytes). Liga o painel |
| `AUDITORIA_RETENCAO_DIAS` | 180 | Retenção da auditoria do painel (RN-13, Q-39) |
| `PAINEL_VERSAO_TERMO` | v1 | Versão vigente do termo (Q-25), sugerida nas duas finalidades. `PAINEL_FINALIDADE` deixou de existir na v1.11: a finalidade vem do botão (Q-61); se ainda estiver no `config.env`, é ignorada |
| `PUBLICOS_ARQUIVO_HORAS` | 24 | Por quanto tempo o arquivo de público pode ser baixado (Q-64); inteiro de 1 a 168 |
| `RETENCAO_DIAS_REVOGADO` | 180 | Prazo da RN-18 em dias, inteiro maior que zero (Q-51) |
| `PORTAL_URL_BASE` | vazia = portal desligado (Q-60) | Endereço público do portal usado no link (`https://…`, sem barra no fim); nunca o cabeçalho `Host` |
| `PORTAL_CHAVE_SESSAO` | vazia = portal desligado (Q-60) | Chave HMAC da sessão do titular: 64 dígitos hexadecimais, diferente da do painel |

As variáveis são lidas de `config.env` ao subir o contêiner; variável obrigatória ausente aborta a inicialização com mensagem clara.

```text
perfila/                          módulo Go `perfila`
  CLAUDE.md                       constituição (C-01..C-12) e mapa de contexto
  cmd/omega-server/main.go        [Support] raiz de composição; subcomandos serve, expirar-recomendacoes, hash-senha, carregar-exemplos, expurgar-auditoria e aplicar-retencao (v1.9)
  cmd/perfila-demo/main.go        [Core] demonstração do fluxo em memória
  internal/
    api/                          [Support] handlers, rotas, middlewares, DTOs
    web/                          [v1.3] painel do operador: rotas /painel, sessão, CSRF, templates e static embutidos; [v1.10] portal do titular (/titular), com sessão própria
    application/usecase/          [Core] os cinco casos de uso + ExpirarRecomendacoes + Precondicoes; [v1.3] ProcessarCliente e leituras do painel; [v1.7] GerarRelatorio
    domain/entity/                [Core] entidades e enums (importa só uuid e shared/errs)
    domain/port/                  [Core] repositórios, gateways e Clock
    domain/service/               [Core] PerfilamentoService, SegmentacaoService, RecomendacaoService, ValidadorLGPD
    infra/memoria/                [Core] fakes em memória dos ports (BDD, contrato, demo)
    infra/relogio/                [Core] Clock real e Clock fixo (único lugar com time.Now)
    infra/repository/pg/          [Generic] Pg*Repository e migrações
    infra/external/               [Support] gateways (consentimento local, integração stub)
    infra/config/                 [Support] leitura de config.env
    infra/senha/                  [v1.5] HashSenha com bcrypt (Q-36)
    shared/errs/                  [Core] erros de domínio (sentinelas com código)
    arquitetura_test.go           [Core] teste de arquitetura (C-01, C-05, C-07, C-08)
  features/                       arquivos .feature, steps/ e runner godog
  specs/                          spec, plano, tarefas, regras BDD
  docs/uml/                       UML1..UML8 em texto + imagens originais
  deploy/                         [Generic] Dockerfile, docker-compose.yml, nginx.conf
```

Persistência e implantação:

- PostgreSQL via `database/sql` com o driver `pgx/v5/stdlib` (Q-10); migrações em SQL versionado, embutidas no binário e aplicadas pelo `omega-server` ao subir, sob `pg_advisory_lock` e registradas em `schema_migracoes` (Q-09); uma tabela por entidade (clientes, eventos_comportamentais, perfis_cliente, segmentos, recomendacoes_produto, consentimentos_lgpd, produtos e, desde a v1.4, movimentos_consentimento; desde a v1.5, operadores e registros_auditoria), em snake_case.
- Chaves UUID; timestamps em `timestamptz` (UTC); `UNIQUE(cliente_id)` em perfis_cliente e, desde a v1.11, `UNIQUE(cliente_id, finalidade)` em consentimentos_lgpd; `CHECK` nos enums de status; chaves estrangeiras para clientes. Reforços de integridade que só repetem regras já existentes: chaves estrangeiras `perfis_cliente.segmento_id → segmentos` (D-06) e `recomendacoes_produto.produto_id → produtos` (D-05); `CHECK` de valor ≥ 0 nos eventos (RN-04), score entre 0 e 1 (RN-05) e preço ≥ 0 (`Produto.AtualizarPreco`).
- Clientes de exemplo (Q-30, Q-32): quinze clientes sintéticos com história (o 15º, desde a v1.9, revogou há mais de 180 dias para demonstrar a retenção, RN-18); desde a v1.8 (Q-47), cada um com ID fixo derivado do e-mail sintético (UUID v5 de `perfila:exemplo:<e-mail>`), usado pelo arquivo de exemplo `docs/exemplos/eventos-lote.csv`; (datas de cadastro, consentimento, revogação, eventos e recomendações ao longo de meses), criados pelos casos de uso com um relógio de carga; no modo memória ao subir, no PostgreSQL por `carregar-exemplos`, e só com o banco sem clientes. Desde a v1.11, Ana, Bruno, Heitor e Isabela também concederam `publicidade_terceiros`, e Carla concedeu e revogou.
- Relatórios (v1.7, Q-46): `PgRelatorioRepository` agrega no banco (`COUNT`, `SUM`, `COUNT(DISTINCT cliente_id)`, junção com `consentimentos_lgpd` em `finalidade = 'segmentacao' AND status = 'CONCEDIDO'`, desde a v1.11) e devolve só números; o fake em memória lê os mesmos repositórios em memória. Não há tabela nem migração nova. A suíte de contrato cobre os dois.
- Clientes de volume (v1.7, Q-46): além dos quinze com história, a carga de exemplos cria 60 clientes sintéticos (`c-101` a `c-160`, `@exemplo.test`) com história gerada de forma determinística pelos casos de uso, para os relatórios terem grupos acima de 5.
- A migração 009 (v1.9, Q-50) acrescenta `clientes.data_eliminacao` (nula). Vem depois da 007 (v1.5) e da 008 (v1.6); o migrador aplica cada versão pelo nome (Q-09).
- A migração 010 (v1.10, Q-59) cria `tokens_acesso_titular` (hash único, chave estrangeira para clientes), `pedidos_eliminacao` (`CHECK` no status, um ABERTO por cliente por índice único parcial) e `caixa_saida_simulada`, e um índice em `lower(clientes.email)` para `BuscarPorEmail`. A suíte de contrato cobre os três ports novos e `BuscarPorEmail`.
- A migração 011 (v1.11, Q-61 a Q-64; vem depois da 010 do portal) põe `segmentacao` na finalidade de todos os registros de consentimento existentes, troca `UNIQUE(cliente_id)` por `UNIQUE(cliente_id, finalidade)` com `CHECK` no enumerado, e cria `exportacoes_publico` (com o conteúdo do arquivo em `bytea` anulável e `CHECK` em tipo e destino) e `membros_publico` (chaves estrangeiras para a exportação, o segmento e o cliente; `CHECK` no estado). Os movimentos (só inclusão) não mudam.
- Subcomando `aplicar-retencao` (v1.9, RN-18): o administrador o executa como `expirar-recomendacoes`; o log `retencao_aplicada` traz só a quantidade.
- A migração 006 (v1.4, Q-34) cria `movimentos_consentimento` (chave UUID, coluna de ordem de gravação, `CHECK` em tipo e origem, operador obrigatório com origem painel, chave estrangeira para clientes) com um gatilho que recusa `UPDATE` e `DELETE` (só inclusão, RN-10), e reconstrói os movimentos dos clientes antigos a partir das datas que o registro prova: a última concessão (com finalidade e termo) e a última revogação (com finalidade e termo só se o estado ainda for REVOGADO; senão, em branco), com origem `migracao`.
- A migração 008 (v1.6, Q-33, Q-44) cria os índices únicos parciais de `segmentos` (prioridade entre os ativos; nome em minúsculas) e acrescenta `campos` (texto, vazio por padrão) a `registros_auditoria`, com o `CHECK` das ações atualizado.
- A migração 007 (v1.5, Q-35, Q-39) cria `operadores` (chave UUID, `login` único, `CHECK` no papel, hash, ativo, criado em, troca obrigatória e data da senha) e `registros_auditoria` (chave UUID, coluna de ordem de gravação, `CHECK` em ação, tipo do alvo e resultado, alvo com tipo e ID juntos, índices por data, operador e ação), com um gatilho que recusa `UPDATE` (só inclusão; o `DELETE` fica para o expurgo, RN-13). Sem chave estrangeira de `registros_auditoria.operador`: o login pode estar em branco (entrada com login desconhecido). `CriarSeVazio` grava sob `pg_advisory_xact_lock`, para as duas réplicas não criarem dois administradores iniciais.
- Operadores de exemplo (Q-41): `admin` (pelas variáveis de desenvolvimento do `make serve`), `atendimento` e `analista`, criados só no modo memória; `carregar-exemplos` no PostgreSQL nunca cria operadores.
- A migração 005 (Q-32) amplia a demonstração: segmentos vip, fiel, ocasional e explorador e catálogo de 7 categorias, iguais aos do modo memória.
- A migração 004 (v1.3) cria os índices da listagem de clientes: `data_cadastro DESC` e o nome normalizado (minúsculas e sem acentos), Q-23.
- A migração inicial cria o segmento padrao (critério vazio, Prioridade 99). Uma segunda migração carrega os segmentos e o catálogo sintéticos de demonstração, com os mesmos IDs do modo memória (Q-08). Dados de teste são sintéticos.
- Uma suíte de contrato (`internal/infra/repository/contrato`) define o comportamento observável dos nove repositórios (e, desde a v1.11, do `PublicoRepository`) (sete desde a T-19; operadores e auditoria desde a v1.5; desde a v1.6, a gravação de segmentos e produtos e `ListarConcedidos`) e roda contra os fakes em memória e contra o PostgreSQL (`PERFILA_TEST_DATABASE_URL`; sem ela, a parte PostgreSQL é pulada).
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
| v1.11 | 13/11/2026 | Consentimento por finalidade e públicos para campanha (Q-61 a Q-67): `Finalidade` enumerada e um `ConsentimentoLGPD` por (cliente, finalidade) (3.2, 3.4.1, RN-01 a RN-03; fecha L-04); UC-06 parcial com `ExportacaoPublico`, `MembroPublico` e a máquina 3.4.3, RN-21 (público, mínimo de dados, remoção propagada, retenção, logs) e RN-22 (agregado com grupo mínimo de 5); casos de uso de públicos e UC-08 com a finalidade em `Revogar` e `Consultar` (3.5); `ErrArquivoIndisponivel`; `PerfilRepository.ListarPorSegmento`, `ConsentimentoRepository` por finalidade, `PublicoRepository` e `DestinoPublicitarioGateway` (3.6); finalidade na API (3.7); tela Públicos e consentimento por finalidade no painel (3.7.1); `PUBLICOS_ARQUIVO_HORAS` e fim de `PAINEL_FINALIDADE` (3.8); migração 011. Numerada depois das branches abertas (v1.5 a v1.8, Q-35 a Q-47) para não colidir |
| v1.10 | 14/11/2026 | Portal do titular (Q-55 a Q-60): ator Titular e casos de uso `SolicitarAcessoTitular`, `AutenticarTitular`, `ConsultarMeusDados`, `PedirEliminacao`, `ListarPedidosEliminacao` e `ListarCaixaSaida` (3.1, 3.5); entidades `TokenAcessoTitular`, `PedidoEliminacao` e `MensagemSimulada` (3.2); RN-19 (acesso por link de uso único) e RN-20 (pedido de eliminação), RN-17 com os tokens e o pedido atendido (3.3); `ErrAcessoInvalido` (3.5); `ClienteRepository.BuscarPorEmail`, `TokenAcessoRepository`, `PedidoEliminacaoRepository`, `Notificador` e `CaixaSaidaSimulada` (3.6); pedidos e caixa de saída no painel (3.7.1); seção 3.7.2; `PORTAL_URL_BASE`, `PORTAL_CHAVE_SESSAO` e migração 010 (3.8). Empilhada sobre a v1.9; os números seguem os das branches abertas |
| v1.9 | 13/11/2026 | Direitos do titular e retenção (Q-48 a Q-53): `Cliente.DataEliminacao`, `Anonimizar` e `EstaEliminado` (3.2); RN-01 com o cliente eliminado, RN-03 sem D-13, RN-17 (eliminação) e RN-18 (retenção) (3.3); `ErrClienteEliminado` e os casos de uso `AtualizarCadastro`, `ExportarDadosTitular`, `EliminarDadosTitular` e `AplicarRetencao` (3.1, 3.5); `RemoverPorCliente` em eventos, perfis e recomendações e `ConsentimentoRepository.ListarRevogadosAte` (3.6); aba Direitos do titular e retenção na Operação, log `direito_atendido` (3.7.1); `RETENCAO_DIAS_REVOGADO`, subcomando `aplicar-retencao` e migração 009 (3.8); exceção dos direitos do titular em C-03. As versões v1.5 a v1.8 estão nas branches abertas (usuários e auditoria, segmentos e catálogo, relatórios, importação de eventos) |
| v1.8 | 13/11/2026 | Importação de eventos em lote (Q-47): RN-16 (CSV com cabeçalho, até 1.000 linhas e 256 KiB, linha a linha com RN-01 e RN-04, duplicados ignorados, prévia sem gravar); `ImportarEventosUseCase` e `LerCSVEventos` (3.1, 3.5); rota `POST /v1/eventos/lote` (3.7); tela Importar eventos e modelo CSV (3.7.1); IDs fixos nos clientes de exemplo (3.8). Numeração a partir de Q-47, T-65 e RN-16 para não colidir com as branches abertas (v1.5 a v1.7) |
| v1.7 | 13/11/2026 | Relatórios agregados, UC-05 e UC-07 parciais (Q-46): RN-15 (só agregados, comportamento só com consentimento válido, grupos de 1 a 4 clientes como "menos de 5", categorias pequenas em "Outras"); `Contagem`, `Media`, `Percentual` e as funções de proteção (3.2); `GerarRelatorioUseCase` (3.1, 3.5); `RelatorioRepository` (3.6); rotas `/painel/relatorios` e o CSV (3.7.1); clientes de volume nos exemplos (3.8). Numeração a partir de Q-46, T-60 e RN-15 para não colidir com as branches abertas (v1.5 e v1.6) |
| v1.6 | 13/11/2026 | Edição de segmentos e do catálogo pelo painel (Q-33, Q-42 a Q-44; Q-45 aberta): RN-14; permissão `editar_configuracao` (RN-11); `Segmento.Validar`, `ValidarCriterio`, `ValidarSegmentos`, `Produto.Validar`, `NormalizarCategoria` e teto de `AtualizarPreco` (3.2); `RegistroAuditoria.Campos` e ações novas (RN-13); `SalvarSegmento`, `PreverSegmentacao`, `SalvarProduto`, `ErrSegmentoNaoEncontrado`, `ErrProdutoNaoEncontrado` e `errs.ErroCampo` (3.5); `SegmentoRepository.ListarTodos` e `Salvar`, `ProdutoRepository.BuscarPorID` e `Salvar`, `ConsentimentoRepository.ListarConcedidos`, `EscolherSegmento` (3.6); rotas (3.7.1); migração 008 (3.8) |
| v1.5 | 13/11/2026 | Usuários, perfis de acesso e auditoria do painel (Q-35 a Q-41): papéis e o ator Atendimento (3.1); entidades `Operador` e `RegistroAuditoria` (3.2); RN-11 (matriz de permissões), RN-12 (operadores e senhas), RN-13 (auditoria e retenção de 180 dias); erros e casos de uso `AutenticarOperador`, `GerenciarOperadores`, `RegistrarAuditoria`, `ListarAuditoria`, `ExpurgarAuditoria` (3.5); `OperadorRepository`, `AuditoriaRepository`, `HashSenha` (3.6); rotas, permissão por rota, sessão conferida a cada requisição e menu por papel (3.7.1); `PAINEL_*` e `AUDITORIA_RETENCAO_DIAS`, subcomando `expurgar-auditoria`, `infra/senha`, migração 007 e operadores de exemplo (3.8) |
| v1.4 | 12/11/2026 | Histórico completo do consentimento (Q-34): entidade `MovimentoConsentimento` (3.2), RN-10, UC-08 `Historico` e origem pelo contexto (3.5), `FichaCliente.Historico`, `ConsentimentoRepository.Registrar` e `ListarMovimentos` (3.6), origem nas rotas da API e do painel e linha do tempo pelo histórico (3.7, 3.7.1), migração 006 (3.8) |
| v1.3 | 12/11/2026 | Painel web do operador (Q-20 a Q-29): seção 3.7.1, casos de uso de aplicação do painel (3.1 e 3.5), `ProcessarClienteUseCase` com o estado por etapa (RN-09, Q-26), `ClienteRepository.Listar`, `RecomendacaoRepository.ListarPorCliente` e `ProdutoRepository.ListarTodos` (3.6), configuração `PAINEL_*` (3.8), migração 004, subcomando `hash-senha` (Q-28); camada BDD `@web` |
| v1.2 | 11/11/2026 | Fase Generic (T-19): `DATABASE_URL` vazia mantém o modo memória fora da implantação (Q-07); dados de demonstração por migração (Q-08); migrações aplicadas pelo servidor com trava (Q-09); driver pgx (Q-10); suíte de contrato dos repositórios; reforços de integridade no esquema (FKs para segmentos e produtos, CHECKs de faixa); erro de `SegmentoRepository.BuscarPorID` explicitado. T-24: implantação com Nginx, duas réplicas e banco isolado (Q-11 a Q-14). Correções do mapeamento de bugs (Q-15 a Q-19): limite de Valor, tolerância de relógio, falhas de configuração como 500, uma recomendação ATIVA por cliente, `/saude` com o banco e corpo JSON único |
| v1.1 | 07/11/2026 | UC-09 Cadastrar Cliente (D-22), `NovoCliente` e rota `POST /v1/clientes`; Support implementado (API, middlewares, CMD, gateways, configuração); D-01 a D-23 adotadas como finais; Q-01 a Q-06 decididas (Q-02 muda o erro de "nenhum segmento" para falha de configuração) |
| **v1.0** | 03/11/2026 | Nome do produto (Perfila); Constituição movida para `CLAUDE.md`; diagramas transcritos em `docs/uml/UML1..UML8`; `ClassificaPerfil` → `ClassificarPerfil` (D-18); `NovoConsentimento`, `Precondicoes` e `ExpirarRecomendacoesUseCase` explicitados; atores de UC-08 corrigidos (D-17); D-07 e D-10 corrigidos; D-17 a D-21 novos; estrutura de pastas marcada por fase (Core, Support, Generic) |
