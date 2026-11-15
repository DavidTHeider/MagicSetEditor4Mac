package msescript_test

// **Der Baumspeicher zählt** — OPS-08.AK-01
// (#1725).
//
// **Ohne `t.Parallel()`**: der Speicher gehört dem Prozess, und gemessen wird
// der Unterschied — ein Nachbar, der zugleich zerteilte, zählte mit.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

func TestDerBaumspeicherZaehltUndLeert(t *testing.T) {
	vorher := msescript.Baumzahlen()

	for range 2 {
		if _, fehler := msescript.Zerteile(`zahl := 40 + 2 # baumzahlen`); fehler != nil {
			t.Fatal(fehler)
		}
	}

	nachher := msescript.Baumzahlen()
	if nachher.Treffer-vorher.Treffer != 1 || nachher.Fehlgriffe-vorher.Fehlgriffe != 1 ||
		nachher.Eintraege < 1 || nachher.GrenzeEintraege != msescript.GrenzeBaumspeicher {
		t.Errorf("vorher %+v, nachher %+v", vorher, nachher)
	}

	msescript.BaumspeicherLeeren()

	if geleert := msescript.Baumzahlen(); geleert.Eintraege != 0 {
		t.Errorf("nach dem Leeren: %+v", geleert)
	}
}
