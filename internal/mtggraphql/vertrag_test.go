package mtggraphql_test

// **Der Vertrag mit MTGJSON** — Werkstück 6 von Phase 4.8
// (#1321).
//
// # Was ein Vertragstest hier hält
//
// §16, „Drift wird erkannt, nicht erlitten": **Pflichtfelder streng,
// Zusatzfelder tolerant.** Ein neues Feld beim Anbieter bricht nichts; ein
// **verschwundenes Pflichtfeld** bricht laut.
//
// # Und was er ausdrücklich nicht hält
//
// Ob der Dienst diese Gestalt noch schickt. Dafür gäbe es einen Lauf gegen
// das Original — und den gibt es hier **nicht**, weil es keine Patreon-Marke
// gibt (Vermerk 0056, Entscheidung 4). Die **Feldnamen** sind trotzdem am
// echten Schema geprüft: die Introspektion antwortet ohne Marke (gemessen am
// 2026-09-02).

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtggraphql"
)

// **Ein fehlendes Pflichtfeld bricht laut.**
//
// **Zwei sind es**: `uuid` und `name`. Ohne die Kennung lässt sich die Karte
// nicht wiederfinden, ohne den Namen nicht anzeigen — beides macht sie
// unbrauchbar, und eine unbrauchbare Karte still durchzulassen wäre die
// schlechteste Antwort (§8).
//
// **`setCode` ist hier kein Pflichtfeld** — anders als bei
// magicthegathering.io. Der Grund steht im Schema: MTGJSON kennt
// `AtomicCard`, eine Karte **ohne** Ausgabe. Wer `setCode` verlangte, wiese
// eine gültige Antwort ab.
func TestEinFehlendesPflichtfeldBrichtLaut(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name  string
		roh   string
		nennt string
	}{
		{"ohne uuid", `{"data":{"cards":[{"name":"X"}]}}`, "uuid"},
		{"ohne name", `{"data":{"cards":[{"uuid":"a"}]}}`, "name"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			_, err := mtggraphql.Lies([]byte(fall.roh))

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

	karten, err := mtggraphql.Lies([]byte(
		`{"data":{"cards":[{"uuid":"a","name":"X","brandneu":{"tief":1}}]}}`))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(karten) != 1 || karten[0].Name != "X" {
		t.Errorf("gelesen: %+v", karten)
	}
}

// **Und eine vollständige Karte kommt durch** — sonst prüften die Läufe oben
// nur, dass immer abgelehnt wird.
func TestEineVollstaendigeKarteKommtDurch(t *testing.T) {
	t.Parallel()

	karten, err := mtggraphql.Lies(beispiel(t, "karten.json"))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(karten) != 1 {
		t.Fatalf("%d Karten, erwartet 1", len(karten))
	}
}
