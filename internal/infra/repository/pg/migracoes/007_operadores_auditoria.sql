-- 007: operadores do painel com papel e trilha de auditoria (Spec v1.5, RN-11 a RN-13, Q-35, Q-39).

CREATE TABLE operadores (
    id                  uuid        PRIMARY KEY,
    login               text        NOT NULL UNIQUE CHECK (login = lower(login) AND length(login) BETWEEN 3 AND 32),
    nome                text        NOT NULL CHECK (nome <> ''),
    papel               text        NOT NULL CHECK (papel IN ('ADMINISTRADOR', 'ATENDIMENTO', 'ANALISTA')),
    senha_hash          text        NOT NULL CHECK (senha_hash <> ''),
    ativo               boolean     NOT NULL,
    criado_em           timestamptz NOT NULL,
    trocar_senha        boolean     NOT NULL,
    senha_atualizada_em timestamptz NOT NULL
);

-- Sem dado pessoal: só login, ação, tipo e ID do alvo e resultado (C-09). Sem chave estrangeira
-- em operador: o login fica em branco numa entrada com login desconhecido.
CREATE TABLE registros_auditoria (
    id         uuid        PRIMARY KEY,
    ordem      bigint      GENERATED ALWAYS AS IDENTITY UNIQUE, -- desempate pela ordem de gravação
    data       timestamptz NOT NULL,
    operador   text        NOT NULL DEFAULT '',
    acao       text        NOT NULL CHECK (acao IN ('entrar', 'sair', 'sessao_encerrada', 'acesso_negado',
        'ler_ficha', 'cadastrar_cliente', 'conceder_consentimento', 'revogar_consentimento', 'registrar_evento',
        'processar', 'expirar_recomendacoes', 'criar_operador', 'mudar_papel', 'desativar_operador',
        'ativar_operador', 'redefinir_senha', 'trocar_senha')),
    alvo_tipo  text        NOT NULL DEFAULT '' CHECK (alvo_tipo IN ('', 'cliente', 'segmento', 'produto', 'operador')),
    alvo_id    uuid,
    resultado  text        NOT NULL CHECK (resultado IN ('SUCESSO', 'NEGADO', 'FALHA')),
    CHECK ((alvo_tipo = '') = (alvo_id IS NULL))
);

CREATE INDEX registros_auditoria_data ON registros_auditoria (data DESC, ordem DESC);
CREATE INDEX registros_auditoria_operador ON registros_auditoria (operador, data DESC);
CREATE INDEX registros_auditoria_acao ON registros_auditoria (acao, data DESC);

-- Só inclusão: alterar um registro é recusado pelo banco. O DELETE fica para o expurgo (RN-13).
CREATE FUNCTION registros_auditoria_so_inclusao() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'registros_auditoria é só de inclusão (RN-13)';
END;
$$;

CREATE TRIGGER registros_auditoria_so_inclusao
    BEFORE UPDATE ON registros_auditoria
    FOR EACH ROW EXECUTE FUNCTION registros_auditoria_so_inclusao();
