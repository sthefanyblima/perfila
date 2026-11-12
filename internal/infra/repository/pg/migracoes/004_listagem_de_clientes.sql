-- Spec: 3.6, 3.8, Q-23. Listagem do painel: cadastro mais recente primeiro e busca por nome sem
-- diferenciar maiúsculas e acentos (a mesma expressão de PgClienteRepository.Listar).
CREATE INDEX clientes_data_cadastro_idx ON clientes (data_cadastro DESC, id);
CREATE INDEX clientes_nome_normalizado_idx ON clientes
    (translate(lower(nome), 'áàâãäéèêëíìîïóòôõöúùûüçñ', 'aaaaaeeeeiiiiooooouuuucn'));
