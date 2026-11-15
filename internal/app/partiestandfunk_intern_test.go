package app

// **Ein Standwechsel am Wirt erreicht die Gäste** —
// #2506.
// Bis hierher bekam nur das eigene Fenster die neue Sicht; ein Gast sah nach
// einer Pause weiter eine laufende Partie, bis sein nächster Wunsch mit einer
// Absage zurückkam.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einTischZumWechseln legt eine Partie an und setzt einen Gast an ihren Tisch.
func einTischZumWechseln(t *testing.T) (*Bindungen, string, *angenommenendoppel) {
	t.Helper()

	db := bestandFuerDenRiegel(t)
	anwendung := MitPartienleitung(NeueApp(""), db)
	anwendung.db = db
	anwendung.profil.Store(1)

	gebunden := NeueBindungen(anwendung)

	angelegt := gebunden.PartieAnlegen(t.Context(), einEntwurf(einBaerImBestand(t, db)))
	if angelegt.Fehler != nil {
		t.Fatalf("PartieAnlegen: %+v", angelegt.Fehler)
	}

	griff := eineAngenommene(t, "partie")
	anwendung.partie.runde.Setze(einGastplatz(griff))

	return gebunden, angelegt.Partie.ID, griff
}

func TestEinStandwechselErreichtDenGast(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		wechsel func(*Bindungen, *testing.T, string) PartieAntwort
		stand   domain.Partiestand
	}{
		"Pausieren": {
			wechsel: func(b *Bindungen, t *testing.T, id string) PartieAntwort {
				t.Helper()

				return b.PartiePausieren(t.Context(), id, "sitz-1")
			},
			stand: domain.StandPausiert,
		},
		"Beenden": {
			wechsel: func(b *Bindungen, t *testing.T, id string) PartieAntwort {
				t.Helper()

				return b.PartieBeenden(t.Context(), id, "sitz-1")
			},
			stand: domain.StandBeendet,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gebunden, id, griff := einTischZumWechseln(t)

			if antwort := fall.wechsel(gebunden, t, id); antwort.Fehler != nil {
				t.Fatalf("der Wechsel: %+v", antwort.Fehler)
			}

			bilder := griff.tischbilder(t)
			if len(bilder) != 1 || bilder[0].Sicht.Stand != fall.stand {
				t.Errorf("beim Gast kam %+v an", bilder)
			}
		})
	}
}

// **Ein gescheiterter Wechsel verkündet nichts.**
func TestEinGescheiterterStandwechselVerkuendetNichts(t *testing.T) {
	t.Parallel()

	gebunden, id, griff := einTischZumWechseln(t)

	if antwort := gebunden.PartieBeenden(t.Context(), id, "sitz-1"); antwort.Fehler != nil {
		t.Fatalf("Beenden: %+v", antwort.Fehler)
	}

	vorher := len(griff.tischbilder(t))

	if antwort := gebunden.PartieFortsetzen(t.Context(), id, "sitz-1"); antwort.Fehler == nil {
		t.Fatal("eine beendete Partie ging weiter")
	}

	if nachher := len(griff.tischbilder(t)); nachher != vorher {
		t.Errorf("nach dem gescheiterten Wechsel gingen %d Bilder hinaus", nachher-vorher)
	}
}

// **Ein toter Draht nimmt dem Wechsel nicht seine Antwort** (§7, Regel 3).
func TestEinToterDrahtNimmtDemWechselNichtSeineAntwort(t *testing.T) {
	t.Parallel()

	gebunden, id, griff := einTischZumWechseln(t)
	griff.sendefehler = errors.New("die Leitung ist weg")

	antwort := gebunden.PartiePausieren(t.Context(), id, "sitz-1")

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("der tote Draht hat den Wechsel gekippt: %+v", antwort.Fehler)
	case antwort.Partie.Stand != string(domain.StandPausiert):
		t.Errorf("die Antwort zeigt %q", antwort.Partie.Stand)
	}
}
