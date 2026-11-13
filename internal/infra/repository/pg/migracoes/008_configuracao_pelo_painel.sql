-- Spec: 3.8, RN-14, Q-33, Q-44. Edição de segmentos e catálogo pelo painel (v1.6).

-- Repete no banco as unicidades da RN-14, porque duas réplicas podem gravar ao mesmo tempo.
CREATE UNIQUE INDEX segmentos_prioridade_ativos ON segmentos (prioridade) WHERE ativo;
CREATE UNIQUE INDEX segmentos_nome_unico ON segmentos (lower(nome));

-- Auditoria: nomes dos campos alterados (lista fechada da RN-13, separados por vírgula; nunca valores).
ALTER TABLE registros_auditoria ADD COLUMN campos text NOT NULL DEFAULT ''
    CHECK (campos ~ '^((nome|descricao|criterio|prioridade|ativo|categoria|preco)(,(nome|descricao|criterio|prioridade|ativo|categoria|preco))*)?$');

ALTER TABLE registros_auditoria DROP CONSTRAINT registros_auditoria_acao_check;
ALTER TABLE registros_auditoria ADD CONSTRAINT registros_auditoria_acao_check CHECK (acao IN ('entrar', 'sair',
    'sessao_encerrada', 'acesso_negado', 'ler_ficha', 'cadastrar_cliente', 'conceder_consentimento',
    'revogar_consentimento', 'registrar_evento', 'processar', 'expirar_recomendacoes', 'criar_operador',
    'mudar_papel', 'desativar_operador', 'ativar_operador', 'redefinir_senha', 'trocar_senha',
    'criar_segmento', 'editar_segmento', 'ativar_segmento', 'desativar_segmento', 'prever_segmentacao',
    'criar_produto', 'editar_produto', 'ativar_produto', 'desativar_produto'));
