-- 010: portal do titular (Spec v1.10, RN-19, RN-20, Q-56 a Q-59).
-- Número 010 porque 007 e 008 estão nas branches abertas da v1.5 e da v1.6 e 009 é a da v1.9;
-- o migrador aplica cada versão pelo nome (Q-09).

-- Links de uso único: só o SHA-256 do token é guardado (RN-19).
CREATE TABLE tokens_acesso_titular (
    id          uuid        PRIMARY KEY,
    cliente_id  uuid        NOT NULL REFERENCES clientes (id),
    hash        text        NOT NULL UNIQUE CHECK (hash ~ '^[0-9a-f]{64}$'),
    criado_em   timestamptz NOT NULL,
    expira_em   timestamptz NOT NULL CHECK (expira_em > criado_em),
    usado_em    timestamptz
);
-- Limite de links por cadastro na última hora (ContarDesde).
CREATE INDEX tokens_acesso_titular_cliente_idx ON tokens_acesso_titular (cliente_id, criado_em);

-- Pedidos de eliminação feitos no portal (RN-20): no máximo um ABERTO por cliente.
CREATE TABLE pedidos_eliminacao (
    id               uuid        PRIMARY KEY,
    cliente_id       uuid        NOT NULL REFERENCES clientes (id),
    status           text        NOT NULL CHECK (status IN ('ABERTO', 'ATENDIDO')),
    data_pedido      timestamptz NOT NULL,
    data_atendimento timestamptz,
    CHECK ((status = 'ATENDIDO') = (data_atendimento IS NOT NULL))
);
CREATE UNIQUE INDEX pedidos_eliminacao_um_aberto_idx ON pedidos_eliminacao (cliente_id) WHERE status = 'ABERTO';

-- Caixa de saída simulada (Q-57): destinatário mascarado, sem ligação com o cliente.
CREATE TABLE caixa_saida_simulada (
    ordem       bigserial   PRIMARY KEY, -- desempata mensagens do mesmo instante
    id          uuid        NOT NULL UNIQUE,
    para        text        NOT NULL,
    assunto     text        NOT NULL,
    corpo       text        NOT NULL,
    enviada_em  timestamptz NOT NULL
);
CREATE INDEX caixa_saida_simulada_enviada_idx ON caixa_saida_simulada (enviada_em);

-- BuscarPorEmail compara sem maiúsculas e sem espaços nas pontas.
CREATE INDEX clientes_email_idx ON clientes (lower(btrim(email)));
