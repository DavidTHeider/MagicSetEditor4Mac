// **Eine fremde Archivfassung kommt übersetzt an** — §8, und der Schlüssel
// zählt, nicht der Satz.
//
// # Warum dieser Lauf hier neu steht
//
// Er ist die Hälfte, die beim Umzug des Rückspielwegs nach `internal/service`
// zurückblieb (#1175).
// Vorher fuhr ein Test in `app` den ganzen Weg und sah am Ende ein `*Fehler`
// mit dem richtigen Schlüssel; heute prüft der Dienst, dass
// [domain.FremdeArchivfassung] **entsteht**, und hier steht, dass die Fassade
// daraus den Schlüssel macht.
//
// **Beide Hälften werden gebraucht.** Der erste Anlauf am 2026-08-26 umschloss
// den Fehler mit `domain.ErrConflict`; im Fenster stand daraufhin „Der Eintrag
// wurde zwischenzeitlich geändert. Bitte neu laden.", und der Go-Test blieb
// grün, weil er die **Meldung** las. Das Fenster bevorzugt den Katalog (§3,
// „zwischen den Prüfebenen").
package app_test

import (
	"fmt"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineFremdeArchivfassungWirdUebersetzt(t *testing.T) {
	for name, fall := range map[string]struct {
		roh        error
		dort, hier string
	}{
		"aus der Zukunft": {
			roh:  &domain.FremdeArchivfassung{Dort: "99.0.0", Hier: "0.3.0"},
			dort: "99.0.0",
			hier: "0.3.0",
		},

		// **Ohne Angabe.** Eine Sicherung aus 0.1.0 oder 0.2.0 nennt keine
		// Fassung — und ist damit gerade **keine** von uns, denn jede, die
		// diese Anwendung je geschrieben hat, nennt ihre.
		"ohne Angabe": {
			roh:  &domain.FremdeArchivfassung{Hier: "0.3.0"},
			dort: "",
			hier: "0.3.0",
		},

		// **Und umschlossen.** `errors.As` greift durch — ein Aufrufer, der
		// den Grund mit Zusammenhang versieht, verliert die Übersetzung nicht.
		"umschlossen": {
			roh: fmt.Errorf("die Sicherung auspacken: %w",
				&domain.FremdeArchivfassung{Dort: "5.0.0", Hier: "0.3.0"}),
			dort: "5.0.0",
			hier: "0.3.0",
		},
	} {
		t.Run(name, func(t *testing.T) {
			fehler := app.Uebersetze(t.Context(), fall.roh)

			if fehler.Schluessel != domain.SchluesselSicherungFremdeFassung {
				t.Errorf("der Schlüssel ist %q — das Fenster zeigte damit den falschen Satz",
					fehler.Schluessel)
			}

			if fehler.Code != app.CodeValidierung {
				t.Errorf("der Code ist %q", fehler.Code)
			}

			if fehler.Werte["dort"] != fall.dort || fehler.Werte["hier"] != fall.hier {
				t.Errorf("die Werte sind %v", fehler.Werte)
			}
		})
	}
}
