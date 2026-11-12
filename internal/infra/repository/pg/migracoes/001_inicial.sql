-- Spec: 3.8 (persistência), T-19. Uma tabela por entidade do UML2, em snake_case.
-- Chaves UUID, timestamptz (UTC), UNIQUE(cliente_id) em perfis e consentimentos,
-- CHECK nos enums de status e chaves estrangeiras para clientes.

CREATE TABLE clientes (
    id            uuid        PRIMARY KEY,
    nome          text        NOT NULL,
    email         text        NOT NULL,
    data_cadastro timestamptz NOT NULL,
    ativo         boolean     NOT NULL
);

CREATE TABLE consentimentos_lgpd (
    id                 uuid        PRIMARY KEY,
    cliente_id         uuid        NOT NULL UNIQUE REFERENCES clientes (id),
    status             text        NOT NULL CHECK (status IN ('NAO_CONCEDIDO', 'CONCEDIDO', 'REVOGADO')),
    finalidade         text        NOT NULL,
    versao_termo       text        NOT NULL,
    data_consentimento timestamptz,          -- nula enquanto NAO_CONCEDIDO
    data_revogacao     timestamptz,          -- mantida após nova concessão (Q-01)
    data_atualizacao   timestamptz NOT NULL
);

CREATE TABLE eventos_comportamentais (
    id          uuid             PRIMARY KEY,
    cliente_id  uuid             NOT NULL REFERENCES clientes (id),
    tipo        text             NOT NULL CHECK (tipo IN ('navegacao', 'compra', 'interacao')),
    origem      text             NOT NULL,
    "timestamp" timestamptz      NOT NULL,
    valor       double precision NOT NULL CHECK (valor >= 0),  -- D-14
    categoria   text             NOT NULL DEFAULT ''           -- D-05, opcional
);
CREATE INDEX eventos_comportamentais_cliente_idx ON eventos_comportamentais (cliente_id, "timestamp");

CREATE TABLE segmentos (
    id         uuid    PRIMARY KEY,
    nome       text    NOT NULL,
    descricao  text    NOT NULL,
    criterio   text    NOT NULL,  -- gramática da RN-06
    prioridade integer NOT NULL,  -- menor valor = avaliado primeiro
    ativo      boolean NOT NULL
);

CREATE TABLE perfis_cliente (
    id                  uuid             PRIMARY KEY,
    cliente_id          uuid             NOT NULL UNIQUE REFERENCES clientes (id),
    frequencia_compras  integer          NOT NULL CHECK (frequencia_compras >= 0),
    ticket_medio        double precision NOT NULL CHECK (ticket_medio >= 0),
    categoria_preferida text             NOT NULL,
    score_engajamento   double precision NOT NULL CHECK (score_engajamento BETWEEN 0 AND 1),
    segmento_id         uuid             REFERENCES segmentos (id)  -- D-06
);

CREATE TABLE produtos (
    id        uuid             PRIMARY KEY,
    nome      text             NOT NULL,
    categoria text             NOT NULL,
    preco     double precision NOT NULL CHECK (preco >= 0),
    ativo     boolean          NOT NULL
);
CREATE INDEX produtos_categoria_idx ON produtos (categoria) WHERE ativo;

CREATE TABLE recomendacoes_produto (
    id             uuid        PRIMARY KEY,
    cliente_id     uuid        NOT NULL REFERENCES clientes (id),
    produto_id     uuid        NOT NULL REFERENCES produtos (id),
    motivo         text        NOT NULL,
    status         text        NOT NULL CHECK (status IN ('GERADA', 'ATIVA', 'EXPIRADA')),
    data_geracao   timestamptz NOT NULL,
    data_ativacao  timestamptz,
    validade       timestamptz,
    data_expiracao timestamptz
);
-- Seleção do comando expirar-recomendacoes (RN-08).
CREATE INDEX recomendacoes_produto_vencidas_idx ON recomendacoes_produto (validade) WHERE status = 'ATIVA';

-- Segmento padrao: critério vazio casa sempre e é avaliado por último (RN-06, Q-02).
-- Mesmo ID do modo memória (internal/infra/memoria/seed.go).
INSERT INTO segmentos (id, nome, descricao, criterio, prioridade, ativo)
VALUES ('633e7f9c-7b37-5b47-a1b8-e82a4c9c8c5a', 'padrao', 'segmento padrão', '', 99, true);
