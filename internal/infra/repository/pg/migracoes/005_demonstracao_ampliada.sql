-- Spec: 3.8, Q-32. Demonstração ampliada: mais segmentos (vip, fiel, explorador, ocasional) e um
-- catálogo de 7 categorias com produtos inativos, iguais aos do modo memória (memoria/seed.go,
-- IDs uuid v5 de "perfila:demo:<chave>"). premium e engajado mudam de prioridade.

UPDATE segmentos SET prioridade = 2 WHERE id = '44f8b08a-d52f-560b-8040-5ee5d9bfe81e';
UPDATE segmentos SET prioridade = 4 WHERE id = '8d4d4e41-45b4-5c4a-8c10-c69c0dca1bab';

INSERT INTO segmentos (id, nome, descricao, criterio, prioridade, ativo) VALUES
    ('ea657da6-44da-5862-a67a-6200d4c59ae7', 'vip', 'compras altas e frequentes',
     'ticketMedio >= 1000 && frequenciaCompras >= 4', 1, true),
    ('2e51ad28-d39a-5685-9c6a-5b188cbf6868', 'fiel', 'compra com frequência',
     'frequenciaCompras >= 5', 3, true),
    ('4761750e-bfe9-5cb8-9a52-05a08740ed06', 'explorador', 'navega bastante e ainda não comprou',
     'scoreEngajamento >= 0.4 && frequenciaCompras == 0', 5, true),
    ('a3fff71d-da00-5a9a-83a7-ebd4245d6731', 'ocasional', 'comprou uma ou duas vezes',
     'frequenciaCompras >= 1 && frequenciaCompras <= 2', 6, true);

INSERT INTO produtos (id, nome, categoria, preco, ativo) VALUES
    ('b6f04d0b-0ba8-5c8c-a83c-cdc10701c4eb', 'Livro C', 'livros', 35, true),
    ('387fbd07-2f4f-5f43-a1ed-e524273f92d7', 'Livro Esgotado', 'livros', 60, false),
    ('4fda16a1-efe6-5335-87cb-0fef24111155', 'Jogo Y', 'games', 150, true),
    ('170fbf90-7532-5d13-9efb-dc46e1cd57e8', 'Controle Pro', 'games', 320, true),
    ('70addd8a-5e50-5056-9a73-9a294041ba80', 'Caixa de Som', 'eletronicos', 260, true),
    ('fe023ff8-9494-5f29-89c5-bf61b7f442de', 'Smartwatch Z', 'eletronicos', 900, true),
    ('d9d2725c-1c17-5817-989d-96fbc6c74713', 'Camiseta Básica', 'moda', 60, true),
    ('0485d22d-e312-5343-a6f8-ccf094eea64f', 'Tênis Corrida', 'moda', 420, true),
    ('81680aa8-d432-55d7-9ac3-2b70e79a9980', 'Jaqueta Couro', 'moda', 380, false),
    ('50eff468-9395-5493-8922-e09eff049c5b', 'Cafeteira', 'casa', 290, true),
    ('5bdbc70f-fbb8-5404-b83f-40d9fb671c99', 'Luminária', 'casa', 140, true),
    ('6443f59a-c1d6-575a-b111-d6375bb2170a', 'Kit Skincare', 'beleza', 180, true),
    ('e56c654b-155b-585d-9581-bf314326462e', 'Perfume Floral', 'beleza', 240, true),
    ('d0579e05-f875-5f58-9d92-4bba153fca65', 'Bola Oficial', 'esportes', 120, true),
    ('d87ecd40-3c87-5418-9ffb-8b4e3a9ef6d3', 'Garrafa Térmica', 'esportes', 70, true);
