package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Abbruch eines laufenden Ausdrucks** — Werkstück U13 von Phase 5.1
// (#1424).
//
// Der Auswerter prüft seinen Kontext bei jedem Schritt
// (`msescript.auswerter.werte`); was fehlte, war jemand, der ihn beendet.
// Der Riegel hält je Karte den Abbruch des laufenden Ausdrucks — **je
// Karte**, weil die Karte das Bezugsobjekt der Konsole ist (§7, Regel 4).

// skriptlaeufe hält, welche Karte gerade rechnet.
type skriptlaeufe struct {
	skriptriegel    sync.Mutex
	skriptAbbrueche map[int64]context.CancelFunc
}

// SkriptAbbrechen beendet den laufenden Ausdruck an dieser Karte. Ohne Lauf
// ist es nichts und kein Fehler; die Antwort des Laufs sagt `abgebrochen`.
// **Ohne Profil**, wie [App.BefehlAbbrechen]: der Abbruch liest und schreibt
// nichts im Bestand, er beendet einen Lauf, den dieselbe Sitzung begonnen hat.
func (a *App) SkriptAbbrechen(_ context.Context, karteID string) SchreibAntwort {
	kennung, befund := kennungAus(karteID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	a.skriptriegel.Lock()
	defer a.skriptriegel.Unlock()

	if abbrechen, laeuft := a.skriptAbbrueche[kennung]; laeuft {
		abbrechen()
	}

	return SchreibAntwort{}
}

// skriptBeginnen trägt den Lauf ein — oder weist ihn ab, wenn an dieser
// Karte schon einer läuft.
func (a *App) skriptBeginnen(ctx context.Context, kennung int64) (context.Context, func(), *Fehler) {
	a.skriptriegel.Lock()
	defer a.skriptriegel.Unlock()

	if _, laeuft := a.skriptAbbrueche[kennung]; laeuft {
		return nil, nil, Uebersetze(ctx, fmt.Errorf(
			"%w: an dieser Karte läuft noch ein Ausdruck — erst abwarten oder abbrechen", domain.ErrConflict))
	}

	if a.skriptAbbrueche == nil {
		a.skriptAbbrueche = map[int64]context.CancelFunc{}
	}

	lauf, abbrechen := context.WithCancel(ctx)
	a.skriptAbbrueche[kennung] = abbrechen

	return lauf, func() {
		a.skriptriegel.Lock()
		defer a.skriptriegel.Unlock()

		abbrechen()
		delete(a.skriptAbbrueche, kennung)
	}, nil
}

// SkriptAbbrechen beendet den laufenden Ausdruck an einer Karte (#1424).
func (b *Bindungen) SkriptAbbrechen(ctx context.Context, karteID string) SchreibAntwort {
	return b.app(ctx).SkriptAbbrechen(b.ruf(ctx), karteID)
}
