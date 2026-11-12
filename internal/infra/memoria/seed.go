package memoria

import (
	"github.com/google/uuid"

	"perfila/internal/domain/entity"
)

// CarregarDemonstracao povoa segmentos e catálogo com dados sintéticos para o modo memória
// (D-23), iguais às migrações 002 e 005 (Q-08, Q-32). No PostgreSQL, entram por migração (Spec 3.8).
func CarregarDemonstracao(segmentos *SegmentoRepository, produtos *ProdutoRepository) {
	id := func(s string) uuid.UUID { return uuid.NewSHA1(uuid.NameSpaceOID, []byte("perfila:demo:"+s)) }
	for _, s := range []entity.Segmento{
		{ID: id("seg:premium"), Nome: "premium", Descricao: "alto valor e recorrência",
			Criterio: "ticketMedio >= 500 && frequenciaCompras >= 3", Prioridade: 2, Ativo: true},
		{ID: id("seg:engajado"), Nome: "engajado", Descricao: "interage com frequência",
			Criterio: "scoreEngajamento >= 0.7", Prioridade: 4, Ativo: true},
		// Q-32 (migração 005)
		{ID: id("seg:vip"), Nome: "vip", Descricao: "compras altas e frequentes",
			Criterio: "ticketMedio >= 1000 && frequenciaCompras >= 4", Prioridade: 1, Ativo: true},
		{ID: id("seg:fiel"), Nome: "fiel", Descricao: "compra com frequência",
			Criterio: "frequenciaCompras >= 5", Prioridade: 3, Ativo: true},
		{ID: id("seg:explorador"), Nome: "explorador", Descricao: "navega bastante e ainda não comprou",
			Criterio: "scoreEngajamento >= 0.4 && frequenciaCompras == 0", Prioridade: 5, Ativo: true},
		{ID: id("seg:ocasional"), Nome: "ocasional", Descricao: "comprou uma ou duas vezes",
			Criterio: "frequenciaCompras >= 1 && frequenciaCompras <= 2", Prioridade: 6, Ativo: true},
		{ID: id("seg:padrao"), Nome: "padrao", Descricao: "segmento padrão", Criterio: "", Prioridade: 99, Ativo: true},
	} {
		segmentos.Adicionar(s)
	}
	for _, p := range []entity.Produto{
		{ID: id("prod:livro-a"), Nome: "Livro A", Categoria: "livros", Preco: 80, Ativo: true},
		{ID: id("prod:livro-b"), Nome: "Livro B", Categoria: "livros", Preco: 50, Ativo: true},
		{ID: id("prod:jogo-x"), Nome: "Jogo X", Categoria: "games", Preco: 200, Ativo: true},
		{ID: id("prod:fone-z"), Nome: "Fone Z", Categoria: "eletronicos", Preco: 350, Ativo: true},
		// Q-32 (migração 005)
		{ID: id("prod:livro-c"), Nome: "Livro C", Categoria: "livros", Preco: 35, Ativo: true},
		{ID: id("prod:livro-esgotado"), Nome: "Livro Esgotado", Categoria: "livros", Preco: 60, Ativo: false},
		{ID: id("prod:jogo-y"), Nome: "Jogo Y", Categoria: "games", Preco: 150, Ativo: true},
		{ID: id("prod:controle-pro"), Nome: "Controle Pro", Categoria: "games", Preco: 320, Ativo: true},
		{ID: id("prod:caixa-de-som"), Nome: "Caixa de Som", Categoria: "eletronicos", Preco: 260, Ativo: true},
		{ID: id("prod:smartwatch-z"), Nome: "Smartwatch Z", Categoria: "eletronicos", Preco: 900, Ativo: true},
		{ID: id("prod:camiseta-basica"), Nome: "Camiseta Básica", Categoria: "moda", Preco: 60, Ativo: true},
		{ID: id("prod:tenis-corrida"), Nome: "Tênis Corrida", Categoria: "moda", Preco: 420, Ativo: true},
		{ID: id("prod:jaqueta-couro"), Nome: "Jaqueta Couro", Categoria: "moda", Preco: 380, Ativo: false},
		{ID: id("prod:cafeteira"), Nome: "Cafeteira", Categoria: "casa", Preco: 290, Ativo: true},
		{ID: id("prod:luminaria"), Nome: "Luminária", Categoria: "casa", Preco: 140, Ativo: true},
		{ID: id("prod:kit-skincare"), Nome: "Kit Skincare", Categoria: "beleza", Preco: 180, Ativo: true},
		{ID: id("prod:perfume-floral"), Nome: "Perfume Floral", Categoria: "beleza", Preco: 240, Ativo: true},
		{ID: id("prod:bola-oficial"), Nome: "Bola Oficial", Categoria: "esportes", Preco: 120, Ativo: true},
		{ID: id("prod:garrafa-termica"), Nome: "Garrafa Térmica", Categoria: "esportes", Preco: 70, Ativo: true},
	} {
		produtos.Adicionar(p)
	}
}
