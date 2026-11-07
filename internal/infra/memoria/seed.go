package memoria

import (
	"github.com/google/uuid"

	"perfila/internal/domain/entity"
)

// CarregarDemonstracao povoa segmentos e catálogo com dados sintéticos para o modo memória
// (D-23). No PostgreSQL (fase Generic) os segmentos entram por migração (Spec 3.8).
func CarregarDemonstracao(segmentos *SegmentoRepository, produtos *ProdutoRepository) {
	id := func(s string) uuid.UUID { return uuid.NewSHA1(uuid.NameSpaceOID, []byte("perfila:demo:"+s)) }
	for _, s := range []entity.Segmento{
		{ID: id("seg:premium"), Nome: "premium", Descricao: "alto valor e recorrência",
			Criterio: "ticketMedio >= 500 && frequenciaCompras >= 3", Prioridade: 1, Ativo: true},
		{ID: id("seg:engajado"), Nome: "engajado", Descricao: "interage com frequência",
			Criterio: "scoreEngajamento >= 0.7", Prioridade: 2, Ativo: true},
		{ID: id("seg:padrao"), Nome: "padrao", Descricao: "segmento padrão", Criterio: "", Prioridade: 99, Ativo: true},
	} {
		segmentos.Adicionar(s)
	}
	for _, p := range []entity.Produto{
		{ID: id("prod:livro-a"), Nome: "Livro A", Categoria: "livros", Preco: 80, Ativo: true},
		{ID: id("prod:livro-b"), Nome: "Livro B", Categoria: "livros", Preco: 50, Ativo: true},
		{ID: id("prod:jogo-x"), Nome: "Jogo X", Categoria: "games", Preco: 200, Ativo: true},
		{ID: id("prod:fone-z"), Nome: "Fone Z", Categoria: "eletronicos", Preco: 350, Ativo: true},
	} {
		produtos.Adicionar(p)
	}
}
