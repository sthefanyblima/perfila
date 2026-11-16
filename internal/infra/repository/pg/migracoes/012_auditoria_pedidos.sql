-- 012: auditoria dos pedidos do titular, da importação e dos públicos (Spec v1.12, RN-13, Q-68).
-- Só amplia a lista de ações aceitas, como a 008 fez na v1.6; nenhuma tabela nem coluna nova.

ALTER TABLE registros_auditoria DROP CONSTRAINT registros_auditoria_acao_check;
ALTER TABLE registros_auditoria ADD CONSTRAINT registros_auditoria_acao_check CHECK (acao IN ('entrar', 'sair',
    'sessao_encerrada', 'acesso_negado', 'ler_ficha', 'cadastrar_cliente', 'conceder_consentimento',
    'revogar_consentimento', 'registrar_evento', 'processar', 'expirar_recomendacoes', 'criar_operador',
    'mudar_papel', 'desativar_operador', 'ativar_operador', 'redefinir_senha', 'trocar_senha',
    'criar_segmento', 'editar_segmento', 'ativar_segmento', 'desativar_segmento', 'prever_segmentacao',
    'criar_produto', 'editar_produto', 'ativar_produto', 'desativar_produto',
    'corrigir_cadastro', 'exportar_dados_titular', 'eliminar_dados_titular', 'aplicar_retencao',
    'importar_eventos', 'exportar_publico', 'exportar_remocoes_publico', 'baixar_arquivo_publico'));
