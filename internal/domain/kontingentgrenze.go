package domain

import (
	"errors"
	"fmt"
	"math"
	"time"
)

// ErrKontingent meldet, dass ein Aufrufer über **seinem eigenen** Budget an
// dieser Anwendung liegt — anders als [ErrGedrosselt], bei dem eine fremde
// Gegenstelle zu viele Anfragen meldet. Der Dienst nennt es seit #1429
// `drossel`; MCP trägt seit dem Zweitblick über Phase 1 von 0.9.0 (Befund 4
// zu F03, #2342) dieselbe Klasse.
var ErrKontingent = errors.New("zu viele Rufe")

// Kontingentgrenze ist [ErrKontingent] mit der Wartezeit, bis wieder ein Ruf
// geht.
type Kontingentgrenze struct {
	Warten time.Duration
}

// NeueKontingentgrenze meldet, dass ein Ruf in `warten` wieder geht.
func NeueKontingentgrenze(warten time.Duration) error {
	return &Kontingentgrenze{Warten: warten}
}

// Sekunden ist die Wartezeit, **aufgerundet und nie null** — `in 0 s` hieße
// „sofort", und das stimmte nicht.
func (k *Kontingentgrenze) Sekunden() int {
	return max(int(math.Ceil(k.Warten.Seconds())), 1)
}

func (k *Kontingentgrenze) Error() string {
	return fmt.Sprintf("%s — in %d s wieder", ErrKontingent, k.Sekunden())
}

// Unwrap gibt die Klasse.
func (k *Kontingentgrenze) Unwrap() error { return ErrKontingent }
