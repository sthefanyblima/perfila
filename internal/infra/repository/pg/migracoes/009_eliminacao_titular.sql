-- 009: eliminação a pedido do titular (Spec v1.9, RN-17, Q-50).
-- O cadastro eliminado é anonimizado e marcado com a data; consentimento e movimentos ficam como
-- prova. Número 009 porque 007 e 008 estão nas branches abertas da v1.5 e da v1.6; o migrador
-- aplica cada versão pelo nome (Q-09).

ALTER TABLE clientes ADD COLUMN data_eliminacao timestamptz; -- nula enquanto não eliminado

-- Cadastro eliminado é sempre o anonimizado e inativo (Cliente.Anonimizar).
ALTER TABLE clientes ADD CONSTRAINT clientes_eliminado_anonimizado CHECK (
    data_eliminacao IS NULL
    OR (nome = 'Titular eliminado' AND email = 'eliminado@anonimizado.invalid' AND NOT ativo)
);

-- Seleção da retenção (RN-18): revogados pela data da revogação.
CREATE INDEX consentimentos_lgpd_revogados_idx ON consentimentos_lgpd (data_revogacao) WHERE status = 'REVOGADO';
