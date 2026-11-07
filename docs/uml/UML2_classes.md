# UML2 — Diagrama de Classes

> **Fonte:** Módulo 3 · imagem original: [`img/UML2_classes.jpg`](img/UML2_classes.jpg)
> **Imagem corrigida (v1, gerada do Mermaid abaixo):** [`img/v1/UML2_classes_dominio.png`](img/v1/UML2_classes_dominio.png) · [`img/v1/UML2_classes_aplicacao.png`](img/v1/UML2_classes_aplicacao.png)
> **Fonte da verdade para:** entidades, campos, métodos e nomes (Spec 1, C-04).
> **Versão:** v1.0 (Core) · **Decisões relacionadas:** D-02, D-05, D-06, D-08, D-10, D-14, D-15, D-19.

O modelo abaixo é o **implementado** em `internal/domain/entity`. Campos e métodos marcados `«ADIÇÃO»` não existem na imagem original e foram acrescentados pela Spec.

## Domínio (entidades)

```mermaid
classDiagram
  direction LR
  class Cliente {
    -ID uuid
    -Nome string
    -Email string
    -DataCadastro time
    -Ativo bool
    +NovoCliente(id, nome, email, agora) «ADIÇÃO D-22»
    +AtualizarDados(nome, email) error
    +Ativar()
    +Desativar()
  }
  class ConsentimentoLGPD {
    -ID uuid
    -ClienteID uuid
    -Status StatusConsentimento «ADIÇÃO D-02»
    -Finalidade string
    -VersaoTermo string
    -DataConsentimento time
    -DataRevogacao *time «ADIÇÃO»
    -DataAtualizacao time «ADIÇÃO»
    +Conceder(finalidade, versaoTermo, agora) error
    +Revogar(agora) error
    +EstaValido() bool
  }
  class EventoComportamental {
    -ID uuid
    -ClienteID uuid
    -Tipo TipoEvento
    -Origem string
    -Timestamp time
    -Valor float64
    -Categoria string «ADIÇÃO D-05»
    +Validar(agora) error
  }
  class PerfilCliente {
    -ID uuid
    -ClienteID uuid
    -FrequenciaCompras int
    -TicketMedio float64
    -CategoriaPreferida string
    -ScoreEngajamento float64
    -SegmentoID *uuid «ADIÇÃO D-06»
    +CalcularScore(eventos, agora, meta)
    +AtualizarPerfil(novo)
  }
  class Segmento {
    -ID uuid
    -Nome string
    -Descricao string
    -Criterio string
    -Prioridade int «ADIÇÃO»
    -Ativo bool «ADIÇÃO»
    +ClassificarPerfil(perfil) (bool, error)
  }
  class RecomendacaoProduto {
    -ID uuid
    -ClienteID uuid
    -ProdutoID uuid
    -Motivo string
    -Status StatusRecomendacao «ADIÇÃO D-08»
    -DataGeracao time
    -DataAtivacao *time «ADIÇÃO»
    -Validade *time «ADIÇÃO»
    -DataExpiracao *time «ADIÇÃO»
    +NovaRecomendacao(...) «equivale a Gerar()»
    +Disponibilizar(agora, validadeDias) error «ADIÇÃO»
    +Expirar(agora) error
  }
  class Produto {
    -ID uuid
    -Nome string
    -Categoria string
    -Preco float64
    -Ativo bool
    +AtualizarPreco(preco) error
  }

  Cliente "1" --> "0..1" ConsentimentoLGPD : possui
  Cliente "1" --> "0..1" PerfilCliente : possui (D-19)
  Cliente "1" --> "0..*" EventoComportamental : gera
  Cliente "1" --> "0..*" RecomendacaoProduto : recebe (D-15)
  PerfilCliente "0..*" --> "0..1" Segmento : pertence a (D-06)
  RecomendacaoProduto "0..*" --> "1" Produto : refere-se a
```

Enumerações (C-05): `TipoEvento` = navegacao | compra | interacao · `StatusConsentimento` = NAO_CONCEDIDO | CONCEDIDO | REVOGADO · `StatusRecomendacao` = GERADA | ATIVA | EXPIRADA.

## Aplicação e contratos

```mermaid
classDiagram
  direction TB
  class ColetarDadosUseCase {
    +Executar(ctx, clienteID) []Evento, error
    +Registrar(ctx, evento) error «ADIÇÃO D-12»
  }
  class GerarPerfilUseCase {
    +Executar(ctx, clienteID) PerfilCliente, error
  }
  class SegmentarClienteUseCase {
    +Executar(ctx, clienteID) Segmento, error
  }
  class GerarRecomendacaoUseCase {
    +Executar(ctx, clienteID) RecomendacaoProduto, error
  }
  class GerenciarConsentimentoUseCase {
    +Conceder(ctx, clienteID, finalidade, versao) error
    +Revogar(ctx, clienteID) error
    +Consultar(ctx, clienteID) «ADIÇÃO»
  }
  class CadastrarClienteUseCase {
    +Executar(ctx, nome, email) Cliente, error «ADIÇÃO D-22»
  }
  class ExpirarRecomendacoesUseCase {
    +Executar(ctx) int, error «ADIÇÃO D-08»
  }
  class PerfilamentoService {
    +GerarPerfil(clienteID, eventos) PerfilCliente
  }
  class SegmentacaoService {
    +Segmentar(ctx, perfil) Segmento, error
  }
  class RecomendacaoService {
    +Recomendar(ctx, cliente, perfil, segmento) RecomendacaoProduto, error
  }
  class ValidadorLGPD {
    +ConsentimentoValido(ctx, clienteID) bool, error
  }

  GerarPerfilUseCase ..> ColetarDadosUseCase : «include»
  GerarPerfilUseCase ..> PerfilamentoService
  SegmentarClienteUseCase ..> SegmentacaoService
  GerarRecomendacaoUseCase ..> RecomendacaoService
  ColetarDadosUseCase ..> ValidadorLGPD : RN-01
```

Ports (interfaces) — declarados no **domínio** (`internal/domain/port`, C-02): `ClienteRepository`, `EventoRepository`, `PerfilRepository`, `SegmentoRepository`, `RecomendacaoRepository`, `ConsentimentoRepository`, `ProdutoRepository` «ADIÇÃO D-05», `ConsentimentoGateway`, `IntegracaoGateway`, `Clock`. Assinaturas completas na Spec 3.6.

## Diferenças em relação à imagem original

| Original (Módulo 3) | v1.0 | Motivo |
|---|---|---|
| `concedido: bool` | `Status StatusConsentimento` | D-02: bool não distingue NAO_CONCEDIDO de REVOGADO |
| `Segmento.ClassificarPerfil(perfil): bool` | `ClassificarPerfil(p) (bool, error)` | Critério inválido precisa informar `ErrDadosInvalidos` (RN-06) |
| `EventoComportamental.Validar(): bool` | `Validar(agora) error` | Informar a causa; tempo por parâmetro (C-07) |
| `RecomendacaoProduto.Gerar()` | construtor `NovaRecomendacao(...)` + `Disponibilizar` | D-08 |
| `Recomendar(cliente, segmento)` | `Recomendar(ctx, cliente, perfil, segmento)` | RN-07 precisa da categoria preferida |
| Serviços na camada "Aplicação" | Serviços em `domain/service` | D-10: o Diagrama de Componentes prevalece |
| Interfaces de repositório na "Infraestrutura" | Ports em `domain/port` | C-02, D-10 |
| `ConsentimentoLGPD 1 → 0..1 PerfilCliente (possui)` | `Cliente 1 → 0..1 PerfilCliente` | D-19 |
| `EventoComportamental 1 → 0..* RecomendacaoProduto (recebe)` | `Cliente 1 → 0..* RecomendacaoProduto` | D-15 |
| `Segmento 1 — 1 PerfilCliente` | `PerfilCliente 0..* → 0..1 Segmento` | D-06 |
| `ListarAtivos(): [[Segmento` | `ListarAtivos(ctx) ([]Segmento, error)` | erro de digitação no diagrama |
| — (sem catálogo) | `ProdutoRepository` | D-05 |

## Rastreabilidade

Spec 3.2, 3.3, 3.6 · Features: `evento`, `perfilamento`, `segmentacao`, `recomendacao` · Código: `internal/domain/entity`, `internal/domain/service`, `internal/domain/port`
