package rest

import (
	"context"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Das Ende des Dienstes ist auch das Ende seiner Aufträge** — BK-07b
// (#2116).
//
// Ein angenommener Auftrag überlebt seine Startantwort absichtlich: er ist vom
// Ruf gelöst (`context.WithoutCancel`). Damit gehört er niemandem, den
// `http.Server.Shutdown` kennt — der Dienst endete, der Bestand ging zu, und
// der Auftrag schrieb weiter. **Jetzt gehört er dem Buch**, und das Buch
// schließt mit dem Dienst: erst nimmt es nichts mehr an, dann bricht es ab,
// dann wartet es — **begrenzt**, denn ein Lauf, der den Abbruch nicht hört,
// darf das Ende nicht aufhalten.

// ErrAuftraegeLaufenNoch sagt, dass beim Dienstende nicht jeder Auftrag in
// seiner Frist endete — sie laufen dann noch, während der Bestand zugeht.
var ErrAuftraegeLaufenNoch = errors.New("beim Dienstende liefen noch Aufträge — sie enden mit dem Prozess")

// errBuchGeschlossen weist einen Auftrag ab, der nach dem Ende kommt.
var errBuchGeschlossen = fmt.Errorf("%w: der Dienst endet gerade — kein neuer Auftrag", domain.ErrConflict)

// Schliessen beendet das Buch: keine neuen Aufträge, die laufenden brechen ab,
// und gewartet wird, bis sie enden oder `ctx` abläuft.
//
// **Zweimal schließen ist keines** — das zweite wartet nur noch einmal.
func (b *Auftragsbuch) Schliessen(ctx context.Context) error {
	b.sperre.Lock()
	b.geschlossen = true

	for _, auftrag := range b.auftraege {
		if auftrag.stand == auftragLaeuft {
			auftrag.abbruchVerlangt = true
			auftrag.abbrechen()
		}
	}

	b.sperre.Unlock()

	fertig := make(chan struct{})

	go func() {
		b.laeufe.Wait()
		close(fertig)
	}()

	select {
	case <-fertig:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("%w: %w", ErrAuftraegeLaufenNoch, ctx.Err())
	}
}

// auftragsschluss ist, was ein Auftragsbuch zum Dienstende kann.
type auftragsschluss interface {
	Schliessen(ctx context.Context) error
}

// AuftraegeSchliessen schließt das Auftragsbuch dieses Dienstes — **bevor**
// der Bestand zugeht, und in derselben Frist wie der Server selbst.
//
// **Ein Dienst ohne Buch hat nichts zu schließen**, und eine Naht, die nicht
// schließen kann, ebenso: sie hält dann keine Läufe, die den Bestand
// überlebten.
func (d *Dienst) AuftraegeSchliessen(ctx context.Context) error {
	buch, schliessbar := d.naehte.Auftraege.(auftragsschluss)
	if !schliessbar {
		return nil
	}

	return buch.Schliessen(ctx)
}
