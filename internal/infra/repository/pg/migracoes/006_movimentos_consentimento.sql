-- 006: histórico completo do consentimento (Spec v1.4, RN-10, Q-34).
-- Cada concessão e revogação vira um movimento, só de inclusão. O estado continua em
-- consentimentos_lgpd (RN-02); esta tabela é a prova.

CREATE TABLE movimentos_consentimento (
    id           uuid        PRIMARY KEY,
    ordem        bigint      GENERATED ALWAYS AS IDENTITY UNIQUE, -- desempate pela ordem de gravação
    cliente_id   uuid        NOT NULL REFERENCES clientes (id),
    tipo         text        NOT NULL CHECK (tipo IN ('CONCESSAO', 'REVOGACAO')),
    finalidade   text        NOT NULL,
    versao_termo text        NOT NULL,
    data         timestamptz NOT NULL,
    origem       text        NOT NULL CHECK (origem IN ('api', 'painel', 'portal', 'sistema', 'migracao')),
    operador     text        NOT NULL DEFAULT '',
    CHECK (origem <> 'painel' OR operador <> '')
);

CREATE INDEX movimentos_consentimento_cliente ON movimentos_consentimento (cliente_id, data, ordem);

-- Só inclusão: alterar ou apagar um movimento é recusado pelo próprio banco.
CREATE FUNCTION movimentos_consentimento_so_inclusao() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'movimentos_consentimento é só de inclusão (RN-10)';
END;
$$;

CREATE TRIGGER movimentos_consentimento_so_inclusao
    BEFORE UPDATE OR DELETE ON movimentos_consentimento
    FOR EACH ROW EXECUTE FUNCTION movimentos_consentimento_so_inclusao();

-- Clientes antigos: reconstrói só o que as datas do registro provam, com origem 'migracao'.
-- A última concessão tem finalidade e termo conhecidos. A última revogação só tem finalidade e
-- termo conhecidos se o estado ainda for REVOGADO; depois de uma nova concessão, ficam em branco.
INSERT INTO movimentos_consentimento (id, cliente_id, tipo, finalidade, versao_termo, data, origem)
SELECT id, cliente_id, tipo, finalidade, versao_termo, data, 'migracao'
FROM (
    SELECT gen_random_uuid() AS id, cliente_id, 'CONCESSAO' AS tipo, finalidade, versao_termo,
           data_consentimento AS data
    FROM consentimentos_lgpd WHERE data_consentimento IS NOT NULL
    UNION ALL
    SELECT gen_random_uuid(), cliente_id, 'REVOGACAO',
           CASE WHEN status = 'REVOGADO' THEN finalidade ELSE '' END,
           CASE WHEN status = 'REVOGADO' THEN versao_termo ELSE '' END,
           data_revogacao
    FROM consentimentos_lgpd WHERE data_revogacao IS NOT NULL
) AS reconstruidos
ORDER BY cliente_id, data;
