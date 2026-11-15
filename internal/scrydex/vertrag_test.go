package scrydex_test

// **Der Vertrag mit Scrydex** — Werkstück 6 von Phase 4.8
// (#1321).
//
// # Was er hält und was nicht
//
// §16: **Pflichtfelder streng, Zusatzfelder tolerant.** Ob der Dienst diese
// Gestalt schickt, hält er **nicht** — es gibt hier kein Konto, und anders
// als bei MTGJSON gibt es auch kein Schema zum Nachsehen: eine Anfrage ohne
// Zugang endet mit **401** (gemessen am 2026-09-02).
//
// **Damit ist Scrydex der ungewisseste der vier Dienste.** Was hier steht,
// ist aus der Dokumentation gelesen und an einem selbst geschriebenen
// Beispiel geprüft — mehr geht nicht, und mehr wird nicht behauptet.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scrydex"
)

// **Ein fehlendes Pflichtfeld bricht laut.**
//
// **Zwei sind es**: `id` und `name`. Ohne die Kennung lässt sich die Karte
// nicht wiederfinden, ohne den Namen nicht anzeigen.
func TestEinFehlendesPflichtfeldBrichtLaut(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name  string
		roh   string
		nennt string
	}{
		{"ohne id", `{"data":[{"name":"X"}]}`, "id"},
		{"ohne name", `{"data":[{"id":"a"}]}`, "name"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			_, err := scrydex.LiesKarten([]byte(fall.roh), 200)

			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("Fehler ist %v, erwartet eine Validierung", err)
			}

			if !strings.Contains(err.Error(), fall.nennt) {
				t.Errorf("die Meldung %q nennt %q nicht", err, fall.nennt)
			}
		})
	}
}

// **Ein unbekanntes Feld stört nicht.**
func TestEinUnbekanntesFeldStoertNicht(t *testing.T) {
	t.Parallel()

	karten, err := scrydex.LiesKarten(
		[]byte(`{"data":[{"id":"a","name":"X","brandneu":{"tief":1}}]}`), 200)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(karten) != 1 || karten[0].Name != "X" {
		t.Errorf("gelesen: %+v", karten)
	}
}

// **Und eine vollständige Karte kommt durch.**
func TestEineVollstaendigeKarteKommtDurch(t *testing.T) {
	t.Parallel()

	karten, err := scrydex.LiesKarten(beispiel(t, "karten.json"), 200)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(karten) != 1 {
		t.Fatalf("%d Karten, erwartet 1", len(karten))
	}
}
