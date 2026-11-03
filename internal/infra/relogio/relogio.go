// Package relogio implementa port.Clock. É o único lugar onde time.Now() é permitido (C-07).
package relogio

import (
	"sync"
	"time"

	"perfila/internal/domain/port"
)

var (
	_ port.Clock = Sistema{}
	_ port.Clock = (*Fixo)(nil)
)

// Sistema devolve a hora real em UTC.
type Sistema struct{}

func (Sistema) Agora() time.Time { return time.Now().UTC() }

// Fixo devolve uma hora controlada; usado nos cenários BDD (B-07).
type Fixo struct {
	mu sync.RWMutex
	t  time.Time
}

func NovoFixo(t time.Time) *Fixo { return &Fixo{t: t.UTC()} }

func (f *Fixo) Agora() time.Time {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.t
}

// Definir troca a hora atual.
func (f *Fixo) Definir(t time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.t = t.UTC()
}

// Avancar soma d à hora atual.
func (f *Fixo) Avancar(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.t = f.t.Add(d)
}
