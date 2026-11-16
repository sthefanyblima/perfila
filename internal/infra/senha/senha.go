// Package senha implementa o port HashSenha com bcrypt (custo 12). Único lugar, além da raiz
// de composição, que importa golang.org/x/crypto/bcrypt.
//
// Spec: 3.6, Q-36. Fonte: UML7 (componente Senha, v1.5).
package senha

import (
	"golang.org/x/crypto/bcrypt"

	"perfila/internal/domain/port"
)

var _ port.HashSenha = Bcrypt{} // C-02

// Custo do bcrypt: o mesmo do subcomando hash-senha desde a Q-28.
const Custo = 12

// Bcrypt é o HashSenha da implantação. O valor zero está pronto para uso.
type Bcrypt struct{ custo int }

// NovoBcrypt devolve o hash com o custo indicado; custo <= 0 usa Custo. Testes usam o mínimo
// do bcrypt para não ficarem lentos.
func NovoBcrypt(custo int) Bcrypt { return Bcrypt{custo: custo} }

func (b Bcrypt) Gerar(senha string) (string, error) {
	c := b.custo
	if c <= 0 {
		c = Custo
	}
	h, err := bcrypt.GenerateFromPassword([]byte(senha), c)
	return string(h), err
}

func (b Bcrypt) Conferir(hash, senha string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(senha)) == nil
}
