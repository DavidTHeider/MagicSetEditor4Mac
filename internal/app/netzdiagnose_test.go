package app_test

// **Netzdiagnose über den echten Weg** — D02
// (#2233).
//
// Eine Bindung des Fensters, durch Dienst, Scryfall-Klienten und `netz` bis
// zur Gegenstelle — mit dem Transport, den das Programm baut, nicht mit einem
// Doppel. Zwei Fälle, die sofort scheitern: ein Name ohne Auflösung und ein
// Port ohne Annahme. **Die Fristen misst `internal/netz/netzklassen_test.go`**
// mit kurzen Werten; hier wären es 30 und 60 Sekunden.

import (
	"net"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

func TestDieBindungNenntDenFremdenDienst(t *testing.T) {
	stand := bindungsprobe(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv(netz.SchleifeUmgebung, "an")

	horcher, fehler := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if fehler != nil {
		t.Fatalf("horchen: %v", fehler)
	}

	geschlossen := horcher.Addr().String()
	_ = horcher.Close() //nolint:errcheck // er soll gerade nicht mehr da sein

	for _, fall := range []struct {
		name, adresse, wirt string
		schluessel          domain.Schluessel
	}{
		{"ein Name ohne Auflösung", "http://gibtsnicht.invalid", "gibtsnicht.invalid", domain.SchluesselNetzName},
		{"ein Port ohne Annahme", "http://" + geschlossen, geschlossen, domain.SchluesselNetzNichtErreichbar},
	} {
		t.Run(fall.name, func(t *testing.T) {
			vorher := app.MitScryfalladresse(fall.adresse)
			t.Cleanup(func() { app.MitScryfalladresse(vorher) })

			antwort := stand.Bindungen.ScryfallSammeldatenHolen(t.Context())
			if antwort.Fehler == nil {
				t.Fatal("ohne erreichbare Gegenstelle blieb es ohne Fehler")
			}

			if antwort.Fehler.Code != app.CodeNetz || antwort.Fehler.Schluessel != fall.schluessel ||
				antwort.Fehler.Werte["wirt"] != fall.wirt {
				t.Errorf("übersetzt als %+v", antwort.Fehler)
			}

			// **Der Vorgang führt ins Protokoll** — dort steht der Grund, und
			// das Fenster zeigt den Verweis nur, wenn er unter `werte` reist
			// (`frontend/src/meldung.ts`, `vorgangAus`).
			if antwort.Fehler.Vorgang == "" || antwort.Fehler.Werte["vorgang"] != antwort.Fehler.Vorgang {
				t.Errorf("ohne Griff ins Protokoll: %+v", antwort.Fehler)
			}
		})
	}
}
