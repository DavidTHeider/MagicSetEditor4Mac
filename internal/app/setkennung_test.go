// **Eine unlesbare Setkennung wird abgewiesen, bevor irgendetwas geschieht**
// (#808).
package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Je Weg ein Fall, und die Aufzählung ist die Zusage.**
//
// Sieben Schreibwege nennen das Set: sechs kartenbezogene seit #808, und
// `KarteAnlegen`, das schon immer nur eines trug. Jeder prüft die Kennung,
// bevor er weitergeht, und jeder tut es an einer eigenen Stelle im Quelltext —
// ein Test für einen von ihnen sagte nichts über die anderen sechs.
//
// **Die Aufzählung führte sechs und der Satz sprach von sieben**, weil
// `KarteAnlegen` fehlte: es meldete sich am gemeinsamen `id` statt an `setID`
// und wäre rot geworden. Angestrichen hat es der Zweitblick über Phase O
// (#832);
// jetzt sind es sieben Fälle und sieben Wege.
func TestJederSchreibwegWeistEineUnlesbareSetkennungAb(t *testing.T) {
	anwendung := app.NeueApp("")

	faelle := map[string]func() *app.Fehler{
		"Feld setzen": func() *app.Fehler {
			return anwendung.FeldSetzen(t.Context(), "abc", "1", "name", "x").Fehler
		},
		"Felder setzen": func() *app.Fehler {
			return anwendung.FelderSetzenMitErwartung(t.Context(), "abc",
				[]app.FeldvorwertDTO{{KarteID: "1"}}, "name", "x").Fehler
		},
		"Felder vorschauen": func() *app.Fehler {
			return anwendung.FelderVorschau(t.Context(), "abc", []string{"1"}, "name", "x").Fehler
		},
		"kopieren": func() *app.Fehler {
			return anwendung.KarteKopieren(t.Context(), "abc", "1").Fehler
		},
		"löschen": func() *app.Fehler {
			return anwendung.KarteLoeschen(t.Context(), "abc", "1", false).Fehler
		},
		"tauschen": func() *app.Fehler {
			return anwendung.KartenTauschen(t.Context(), "abc", "1", "2").Fehler
		},
		"anlegen": func() *app.Fehler {
			return anwendung.KarteAnlegen(t.Context(), "abc").Fehler
		},
		"Bild einsetzen": func() *app.Fehler {
			return anwendung.BildEinsetzen(
				probekontext(t), "abc", "1", "kunst", "/tmp/egal.png", app.AuswahlDTO{}).Fehler
		},
	}

	for name, ruf := range faelle {
		t.Run(name, func(t *testing.T) {
			befund := ruf()
			if befund == nil {
				t.Fatal("die unlesbare Kennung kam durch")
			}

			// **Und die Meldung nennt das Feld** (§8). Ohne diese Zusicherung
			// bliebe offen, ob der Befund von der Set- oder der Kartenkennung
			// stammt — und der Aufrufer suchte in der falschen.
			if befund.Feld != "setID" {
				t.Errorf("Feld %q, erwartet war setID: %+v", befund.Feld, befund)
			}
		})
	}
}
