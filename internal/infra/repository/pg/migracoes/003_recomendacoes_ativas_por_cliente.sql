-- Spec: RN-08, Q-18. UC-04 expira as recomendações ATIVA anteriores do cliente antes de gravar
-- a nova; este índice atende ListarAtivasPorCliente.
CREATE INDEX recomendacoes_produto_ativas_cliente_idx ON recomendacoes_produto (cliente_id) WHERE status = 'ATIVA';
