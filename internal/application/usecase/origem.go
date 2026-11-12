package usecase

import (
	"context"

	"perfila/internal/domain/entity"
)

// chaveOrigem é a chave do contexto para a origem do pedido; tipo próprio, sem estado global (C-08).
type chaveOrigem struct{}

type origemPedido struct {
	origem   entity.OrigemMovimento
	operador string
}

// ComOrigem devolve um contexto que leva a origem do pedido até UC-08 sem mudar as assinaturas da
// Spec 3.5. Quem define é o adaptador: a API (`api`), o painel (`painel` e o login da sessão) e,
// na fase 2, o portal do titular. Spec: RN-10, Q-34.
func ComOrigem(ctx context.Context, origem entity.OrigemMovimento, operador string) context.Context {
	return context.WithValue(ctx, chaveOrigem{}, origemPedido{origem: origem, operador: operador})
}

// origemDe lê a origem do contexto; sem valor, `sistema` (subcomandos, exemplos, testes). Spec: RN-10.
func origemDe(ctx context.Context) (entity.OrigemMovimento, string) {
	if o, ok := ctx.Value(chaveOrigem{}).(origemPedido); ok {
		return o.origem, o.operador
	}
	return entity.OrigemSistema, ""
}
