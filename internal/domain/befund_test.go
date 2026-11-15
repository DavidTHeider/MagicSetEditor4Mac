package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Ein Befund nennt seine Stelle, und zwar in der Ausdrucksweise seines Formats.
//
// Eine `.mse-set` wird zeilenweise gelesen — die Stelle ist eine Zeilennummer.
// MTG JSON hat keine Zeilen, die etwas bedeuten: dort ist die Stelle ein Pfad
// durch die Datei. Ein gemeinsamer Typ mit einer festen Zeilennummer zwänge das
// eine Format in die Form des anderen und schriebe für JSON überall `:0:`.
func TestBefundNenntSeineStelle(t *testing.T) {
	faelle := map[string]struct {
		befund   domain.Befund
		erwartet string
	}{
		"Datei und Ort": {
			befund:   domain.Befund{Datei: "probe.mse-set/set", Ort: "12", Meldung: "eine Karte ohne einen einzigen Feldwert"},
			erwartet: "probe.mse-set/set:12: eine Karte ohne einen einzigen Feldwert",
		},
		"ein Pfad statt einer Zeile": {
			befund:   domain.Befund{Datei: "PRB.json", Ort: "Karte 2 → uuid", Meldung: "die Zuordnung kennt das Feld nicht"},
			erwartet: "PRB.json:Karte 2 → uuid: die Zuordnung kennt das Feld nicht",
		},
		"ohne Ort": {
			befund:   domain.Befund{Datei: "PRB.json", Meldung: "kein einziges Set"},
			erwartet: "PRB.json: kein einziges Set",
		},
		"ohne Datei": {
			befund:   domain.Befund{Meldung: "die Quelle nennt keinen Pfad"},
			erwartet: "die Quelle nennt keinen Pfad",
		},

		// Beim Schreiben gibt es noch keine Datei — der Ort ist trotzdem die
		// Auskunft, auf die es ankommt.
		"ein Ort ohne Datei": {
			befund:   domain.Befund{Ort: "PRB → super type", Meldung: "MSE führt beides in einem Feld"},
			erwartet: "PRB → super type: MSE führt beides in einem Feld",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			if gefunden := fall.befund.String(); gefunden != fall.erwartet {
				t.Errorf("String: %q, erwartet %q", gefunden, fall.erwartet)
			}
		})
	}
}
