// Der Wächter über die Naht, die #206 gefehlt hat.
//
// **Er prüft nicht, dass der Weg gebaut ist — sondern dass er benutzt wird.**
// Das ist der Unterschied, an dem drei Werkstücke der Phase 6 unsichtbar
// blieben: `Stilplan` stand seit #120, `Bildquelle` seit #173, die
// Stilumgebung seit #186 — und die Anwendung rief den Behelfsplan.
package app

import (
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Composition Root richtet dem Planbauer eine Stilquelle ein.**
//
// # Die Gegenprobe, die dahintersteht
//
// Wer `MitStilen(…)` aus `app.go` entfernt, macht diesen Lauf rot. Ohne den
// Wächter fiele die Anwendung stillschweigend auf den Behelfsplan zurück, jeder
// andere Test bliebe grün — und genau so war der Stand vor #206.
func TestDerStartRichtetEineStilquelleEin(t *testing.T) {
	anwendung := gestartet(t)

	plan, befunde, err := anwendung.planbauer.Kartenplan(t.Context(),
		&domain.CardRow{ID: 1, SetID: 1, Values: []domain.Value{{Field: "name", Text: "Probe"}}})
	if err != nil {
		t.Fatalf("bauen: %v", err)
	}

	if plan == nil {
		t.Fatal("kein Plan")
	}

	// **Über den Befund und nicht über die Maße.** Auf einem Rechner ohne
	// bezogenes Paket kommt hier der Behelfsplan — richtig so. Was der Wächter
	// hält, ist, dass er *begründet* kommt: „kein Stilverzeichnis eingerichtet"
	// ist der Satz eines Planbauers **ohne** Stilquelle, und genau der darf hier
	// nicht stehen.
	for _, befund := range befunde {
		if befund.Meldung == "für die Vorschau ist kein Stilverzeichnis eingerichtet" {
			t.Error("der Planbauer hat keine Stilquelle — app.go verdrahtet sie nicht")
		}
	}
}

// gestartet baut eine Anwendung auf einer frischen Datenbank.
func gestartet(t *testing.T) *App {
	t.Helper()

	anwendung := NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		if err := anwendung.Beenden(); err != nil {
			t.Errorf("beenden: %v", err)
		}
	})

	return anwendung
}
