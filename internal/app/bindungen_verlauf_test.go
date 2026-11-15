// Die Verlaufsbindungen der Oberfläche — ⌘Z und ⇧⌘Z.
//
// **Eigene Datei wie Bilder und Bogen.** `bindungen_test.go` führt einen
// gemeinsamen Aufbau für ein Dutzend Bindungen; wer jede neue dort anhängt,
// bekommt eine Datei, in der jede Änderung stattfindet und niemand mehr sieht,
// welche Bindung ungeprüft ist (AGENTS.md §2).
package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDieVerlaufsbindungenReichenDurch(t *testing.T) {
	for name, rufen := range map[string]func(*app.Bindungen) app.BewegungAntwort{
		"Zuruecknehmen":    func(b *app.Bindungen) app.BewegungAntwort { return b.Zuruecknehmen(t.Context(), "1", "") },
		"Wiederherstellen": func(b *app.Bindungen) app.BewegungAntwort { return b.Wiederherstellen(t.Context(), "1", "") },
	} {
		t.Run(name, func(t *testing.T) {
			probe := &verlaufsprobe{bewegung: service.Bewegung{
				Bewegt: true, KarteID: 7, Feld: "name", Urheber: domain.UrheberFenster,
			}}

			anwendung := app.NeueApp("")
			anwendung.MitVerlauf(probe)

			antwort := rufen(app.NeueBindungen(anwendung))
			if antwort.Fehler != nil {
				t.Fatalf("%+v", antwort.Fehler)
			}

			if !antwort.Bewegt || antwort.KarteID != "7" || antwort.Feld != "name" {
				t.Errorf("%+v", antwort)
			}

			if antwort.Urheber != string(domain.UrheberFenster) {
				t.Errorf("Urheber %q", antwort.Urheber)
			}

			if probe.gefragt != 1 {
				t.Errorf("gefragt nach Set %d", probe.gefragt)
			}
		})
	}
}
