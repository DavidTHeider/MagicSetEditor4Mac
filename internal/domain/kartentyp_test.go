package domain_test

// **Die Typzeile als Text** — Werkstück M2 der Regelmaschine
// (#1604),
// erweitert für
// #1622.
//
// **Eine eigene Datei**, weil die Läufe über das **Deuten** einer Typzeile
// nebenan stehen (`kartenmerkmale_test.go`) und hier die Frage nach dem
// **Bauen** beantwortet wird: aus zwei Feldern eines MSE-Pakets wird die
// Zeile, die CR 205.3a beschreibt.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Aus zwei Feldern wird eine Typzeile** — CR 205.3a
// (#1622).
//
// **Wofür es sie gibt**: die Höchstzahl gleicher Karten fragt nach der
// **Zeile** (CR 100.2a über [IstStandardland]), und ein MSE-Paket hält zwei
// Felder. Ohne diese Stelle baute jeder Aufrufer sie selbst — mit einem
// Bindestrich statt des Gedankenstrichs, und dann fände `Cut` nichts.
func TestAusZweiFeldernWirdEineTypzeile(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		ober, unter string
		erwartet    string
	}{
		"beides":                 {"Basic Land", "Mountain", "Basic Land — Mountain"},
		"ohne Untertyp":          {"Basic Land", "", "Basic Land"},
		"ohne Obertyp":           {"", "Mountain", "— Mountain"},
		"gar nichts":             {"", "", ""},
		"mit überflüssigem Weiß": {"  Basic Land ", " Mountain ", "Basic Land — Mountain"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if gebaut := domain.TypzeileText(
				fall.ober, fall.unter); gebaut != fall.erwartet {
				t.Errorf("gebaut wurde %q", gebaut)
			}
		})
	}
}

// **Und sie trifft, was IstStandardland liest** — die Naht zwischen beiden.
//
// **Zwei Funktionen, eine Zusage**: wer die Zeile baut, baut sie so, dass die
// Prüfung sie versteht. Ein Lauf über jede einzeln sagte darüber nichts (§10).
func TestEineGebauteTypzeileTrifftDieStandardlandpruefung(t *testing.T) {
	t.Parallel()

	if !domain.IstStandardland(
		domain.TypzeileText("Basic Land", "Mountain")) {
		t.Error("das eigene Gebirge gilt nicht als Standardland")
	}

	// **Die Gegenprobe** — CR 205.4c: der Untertyp allein macht keines.
	if domain.IstStandardland(
		domain.TypzeileText("Legendary Creature Land", "Dryad Forest")) {
		t.Error("Dryad Arbor gilt als Standardland")
	}
}
