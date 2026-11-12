-- Spec: 3.8, Q-08. Segmentos e catálogo sintéticos de demonstração, iguais aos do modo
-- memória (internal/infra/memoria/seed.go, IDs uuid v5 de "perfila:demo:<chave>").
-- Gestão de segmentos e produtos pela API está fora do MVP: entram por migração.

INSERT INTO segmentos (id, nome, descricao, criterio, prioridade, ativo) VALUES
    ('44f8b08a-d52f-560b-8040-5ee5d9bfe81e', 'premium',  'alto valor e recorrência',
     'ticketMedio >= 500 && frequenciaCompras >= 3', 1, true),
    ('8d4d4e41-45b4-5c4a-8c10-c69c0dca1bab', 'engajado', 'interage com frequência',
     'scoreEngajamento >= 0.7', 2, true);

INSERT INTO produtos (id, nome, categoria, preco, ativo) VALUES
    ('5ccc691c-1a12-5a7a-b287-88d051b988dd', 'Livro A', 'livros',      80,  true),
    ('af8b0dc9-f50e-50a6-859e-86e00617c770', 'Livro B', 'livros',      50,  true),
    ('5a6c8780-5e12-55e4-a6ce-5a972e14e376', 'Jogo X',  'games',       200, true),
    ('1f01a68b-a3dc-58b5-b6cd-371110231b59', 'Fone Z',  'eletronicos', 350, true);
