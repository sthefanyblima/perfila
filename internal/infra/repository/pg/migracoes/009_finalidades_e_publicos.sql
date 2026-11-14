-- 009: consentimento por finalidade e públicos para campanha (Spec v1.9, Q-48 a Q-51).
-- 007 e 008 ficam reservadas para as branches de operadores e de configuração pelo painel.

-- Q-48: os registros existentes passam a valer só para "segmentacao"; um registro por
-- (cliente, finalidade), dentro do enumerado. Os movimentos (só inclusão, RN-10) não mudam.
UPDATE consentimentos_lgpd SET finalidade = 'segmentacao';

ALTER TABLE consentimentos_lgpd DROP CONSTRAINT consentimentos_lgpd_cliente_id_key;
ALTER TABLE consentimentos_lgpd
    ADD CONSTRAINT consentimentos_lgpd_cliente_finalidade UNIQUE (cliente_id, finalidade),
    ADD CONSTRAINT consentimentos_lgpd_finalidade CHECK (finalidade IN ('segmentacao', 'publicidade_terceiros'));

-- RN-17: registro de cada arquivo gerado, sem dado pessoal. O conteúdo fica só até
-- arquivo_expira_em (Q-51) e depois é apagado; o registro fica como prova.
CREATE TABLE exportacoes_publico (
    id                uuid        PRIMARY KEY,
    segmento_id       uuid        NOT NULL REFERENCES segmentos (id),
    segmento_nome     text        NOT NULL,
    tipo              text        NOT NULL CHECK (tipo IN ('INCLUSAO', 'REMOCAO')),
    destino           text        NOT NULL CHECK (destino IN ('arquivo_csv')),
    operador          text        NOT NULL,
    data              timestamptz NOT NULL,
    quantidade        integer     NOT NULL CHECK (quantidade > 0),
    hash_arquivo      text        NOT NULL,
    nome_arquivo      text        NOT NULL,
    arquivo_expira_em timestamptz NOT NULL,
    arquivo           bytea
);

CREATE INDEX exportacoes_publico_segmento ON exportacoes_publico (segmento_id, data DESC);

-- RN-17, 3.4.3: cada cliente de um arquivo de inclusão, pelo hash do e-mail (pseudônimo).
CREATE TABLE membros_publico (
    id                    uuid        PRIMARY KEY,
    exportacao_id         uuid        NOT NULL REFERENCES exportacoes_publico (id),
    segmento_id           uuid        NOT NULL REFERENCES segmentos (id),
    cliente_id            uuid        NOT NULL REFERENCES clientes (id),
    hash_email            text        NOT NULL CHECK (hash_email <> ''),
    status                text        NOT NULL CHECK (status IN ('EXPORTADO', 'REMOCAO_PENDENTE', 'REMOVIDO')),
    data_inclusao         timestamptz NOT NULL,
    data_remocao_pendente timestamptz,
    remocao_id            uuid        REFERENCES exportacoes_publico (id),
    data_remocao          timestamptz,
    CHECK (status <> 'REMOVIDO' OR remocao_id IS NOT NULL)
);

CREATE INDEX membros_publico_segmento ON membros_publico (segmento_id, status, hash_email);
CREATE INDEX membros_publico_cliente ON membros_publico (cliente_id, status);
