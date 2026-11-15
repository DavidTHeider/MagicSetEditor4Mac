package app

// **Ein Ereignis des eigenen Tischs trägt die Epoche des Bestands** — #2307,
// Folge aus dem Nachlauf über J05 (#2269), Befund 4. Nach dem Zurückspielen
// einer älteren Sicherung derselben Partie passen Kennung und Sicht weiter;
// erst die Epoche sagt dem Fenster, dass ein verspätetes Bild aus dem alten
// Bestand stammt.

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// gemeldet ist ein Ereignis, wie es ans Fenster ginge.
type gemeldet struct {
	name  string
	werte []any
}

func meldungenVon(a *App) *[]gemeldet {
	alle := &[]gemeldet{}
	a.MitMeldung(func(_ context.Context, name string, werte ...any) {
		*alle = append(*alle, gemeldet{name: name, werte: werte})
	})

	return alle
}

func TestDerEigeneTischMeldetDieEpoche(t *testing.T) {
	a := gestartet(t)
	alle := meldungenVon(a)
	melder := a.eigenerTischmelder(t.Context(), "p1")

	a.lauf.epoche.Store(3)
	_ = melder.Tischbild(t.Context(), domain.Tischbild{}) //nolint:errcheck // der Melder scheitert nicht
	_ = melder.Ereignis(t.Context(), domain.Ereignis{})   //nolint:errcheck // dito

	if len(*alle) != 2 {
		t.Fatalf("gemeldet: %+v", *alle)
	}

	for _, meldung := range *alle {
		if len(meldung.werte) != 3 || meldung.werte[0] != "p1" {
			t.Errorf("%s meldet %v — erwartet Partie, Inhalt und Epoche", meldung.name, meldung.werte)

			continue
		}

		if epoche, ist := meldung.werte[2].(uint64); !ist || epoche != 3 {
			t.Errorf("%s meldet die Epoche %#v, erwartet 3", meldung.name, meldung.werte[2])
		}
	}
}

// **Ein fremder Tisch trägt keine** — er liegt im Bestand des Wirts, und ein
// Tausch hier ändert an ihm nichts. Die Gegenprobe zum Lauf darüber.
func TestEinFremderTischMeldetKeineEpoche(t *testing.T) {
	a := gestartet(t)
	alle := meldungenVon(a)
	melder := a.tischmelder(t.Context(), "p1")

	_ = melder.Tischbild(t.Context(), domain.Tischbild{}) //nolint:errcheck // dito

	if len(*alle) != 1 || len((*alle)[0].werte) != 2 {
		t.Errorf("gemeldet: %+v", *alle)
	}
}
