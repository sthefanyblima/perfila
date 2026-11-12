# UML8 — Diagrama de Implantação

> **Fonte:** Módulo 5 · imagem original: [`img/UML8_implantacao.png`](img/UML8_implantacao.png)
> **Imagem corrigida (v1, gerada do Mermaid abaixo):** [`img/v1/UML8_implantacao.png`](img/v1/UML8_implantacao.png)
> **Fonte da verdade para:** servidores, contêineres e configuração (Spec 3.8).
> **Versão:** v1.3 · **Fase de implementação:** Generic (16–19/11) e painel web (T-31 a T-38) · **Decisões relacionadas:** D-16, D-21, Q-22, Q-28, Q-29.

## Diagrama (v1.3)

```mermaid
flowchart LR
  subgraph DEV["«device» Dispositivo do Cliente"]
    NAV["«artifact» Navegador / App<br/>cliente HTTP · painel /painel «v1.3»"]
    CLI["«artifact» Terminal (CLI)<br/>administrador"]
  end
  subgraph LB["«device» Balanceador de Carga"]
    NGX["«executionEnvironment» Nginx<br/>«artifact» nginx.conf · TLS · única porta de entrada"]
  end
  subgraph APPS["«device» ×2 Servidores de Aplicação (Ubuntu Server)"]
    subgraph DOCKER["«executionEnvironment» Container Docker — Go runtime"]
      BIN["«artifact» omega-server (binário Go)<br/>manifest: CMD · API · Web «v1.3» · Casos de Uso · Domínio · Repositórios · Shared"]
      ENV["«artifact» config.env"]
    end
  end
  subgraph DB["«device» Servidor de Banco de Dados"]
    PG["«executionEnvironment» PostgreSQL<br/>«artifact» omega_db"]
  end
  subgraph EXTN["«device» «external» Serviços Externos"]
    CONS["Serviço de Consentimento (LGPD)"]
    INT["Integrações (e-commerce/CRM)"]
  end

  NAV -- HTTPS --> NGX
  CLI -- SSH --> BIN
  NGX -- "HTTP :8080 (distribui)" --> BIN
  BIN -- "TCP/IP SQL :5432" --> PG
  BIN -- HTTPS --> CONS
  BIN -- HTTPS --> INT
```

## Configuração (`config.env`)

| Variável | Padrão | Uso |
|---|---|---|
| `HTTP_PORT` | 8080 | porta do servidor |
| `DATABASE_URL` | obrigatória na implantação | PostgreSQL :5432, banco `omega_db`; vazia = modo memória, só fora da implantação (D-23, Q-07) |
| `AUTH_TOKEN` | obrigatória | token Bearer (D-16) |
| `RECOMENDACAO_VALIDADE_DIAS` | 30 | RN-08 |
| `ENGAJAMENTO_META_EVENTOS` | 30 | RN-05 (deve ser > 0) |
| `LOG_LEVEL` | info | nível de log |
| `PAINEL_USUARIO` | operador | login do operador do painel (Q-22) |
| `PAINEL_SENHA_HASH` | vazia = painel desligado | hash bcrypt (`omega-server hash-senha`, Q-28) |
| `PAINEL_CHAVE_SESSAO` | vazia = painel desligado | chave HMAC da sessão, 64 dígitos hexadecimais |
| `PAINEL_FINALIDADE` | segmentacao | finalidade sugerida ao conceder (Q-25) |
| `PAINEL_VERSAO_TERMO` | v1 | versão vigente do termo (Q-25) |

## Regras para o agente

1. Binário estático `omega-server` com subcomandos `serve`, `expirar-recomendacoes`, `hash-senha` (Q-28) e `carregar-exemplos` (Q-30); o administrador os executa por SSH.
4. O painel (`/painel`) é servido pelo mesmo binário, atrás do mesmo Nginx; a sessão é assinada e sem estado, então qualquer réplica atende (Q-22). Fontes, CSS e htmx saem do próprio binário: nenhuma requisição a terceiros.
2. O cliente nunca acessa o banco: só o balanceador expõe porta pública (cenário em `implantacao.feature`).
3. `GET /saude` sem autenticação, para o balanceador.

## Divergências registradas

| ID | Divergência | Decisão |
|---|---|---|
| D-21 | O produto passou a se chamar **Perfila**, mas os artefatos do diagrama usam o nome da equipe (`omega-server`, `omega_db`). | Mantidos como no diagrama (são nomes de artefato da Equipe Omega); módulo Go e repositório usam `perfila`. |

## Rastreabilidade

Spec 3.8 · Feature: `implantacao.feature` (`@slow`, fora do filtro na fase Core) · Tarefas T-23, T-24
