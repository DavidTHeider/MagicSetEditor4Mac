package werkzeuge_test

// **Die Zeitzonen, die das Fenster anbietet, nimmt Go an** — UXR-F1
// (#2007),
// R7 aus `docs/UIREVIEW.md`.
//
// # Wofür es ihn gibt
//
// Die Einstellung »zeitzone« war ein Freitext, und R7 verlangt ab fünfzehn
// Werten eine Auswahl mit Suche. Die Auswahl bietet an, was in
// `frontend/src/sprache/zeitzonen.json` steht — und **jede Zeile darin muss
// eine sein, die `domain.PruefeZeitzone` durchlässt**. Sonst wählte jemand
// einen Namen aus der Liste, und das Schreiben wiese ihn ab: das
// Bedienelement versprochen, die Wirkung nicht.
//
// # Was er hält
//
//   - jede Zeile besteht die Prüfung von Go **auf diesem Rechner** — dieselbe
//     Zonendatenbank, aus der auch die Einstellung liest;
//   - die Liste ist **geordnet und ohne Doppel**, sonst stünde dieselbe
//     Zone zweimal in der Auswahl;
//   - »Local« steht nicht darin (Gos Wort für das System; die Einstellung
//     weist es ab), und die Liste hat genug Werte, dass eine Suche dazugehört.
//
// # Was er nicht hält
//
// **Dass die Liste vollständig ist.** Sie stammt aus `Intl.supportedValuesOf`
// (Node 24.15.0, ICU 78.2, tzdata 2026a) und der Zone `UTC`, am 2026-10-03
// gegen Go 1.27.1 auf macOS geprüft. Ein Name, den die Zonendatenbank eines
// anderen Rechners nicht kennt, würde hier rot — und auf diesem Rechner nicht.
// Was die Liste nicht kennt, aber Go annimmt (`US/Pacific`), bleibt als Wert
// sichtbar gewählt (`Suchauswahl`), statt zu verschwinden.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func zeitzonenAusDemFenster(t *testing.T) []string {
	t.Helper()

	ort := filepath.Join("..", "..", "frontend", "src", "sprache", "zeitzonen.json")

	roh, fehler := os.ReadFile(ort) //nolint:gosec // ein fester Ort im Baum
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	var namen []string
	if fehler := json.Unmarshal(roh, &namen); fehler != nil {
		t.Fatalf("%s ist keine Liste von Namen: %v", ort, fehler)
	}

	return namen
}

func TestJedeAngeboteneZeitzoneNimmtGoAn(t *testing.T) {
	namen := zeitzonenAusDemFenster(t)

	abgelehnt := []string{}

	for _, name := range namen {
		if fehler := domain.PruefeZeitzone(name); fehler != nil {
			abgelehnt = append(abgelehnt, name)
		}
	}

	if len(abgelehnt) > 0 {
		t.Errorf("Go weist %d der angebotenen Zeitzonen ab: %v", len(abgelehnt), abgelehnt)
	}
}

// Die Gegenprobe: ohne sie wäre der Lauf darüber grün, auch wenn die Prüfung
// nichts abwiese.
func TestDieZeitzonenpruefungWeistEinenFremdenNamenAb(t *testing.T) {
	for _, name := range []string{"Europe/Nirgendwo", "Local", "berlin"} {
		if domain.PruefeZeitzone(name) == nil {
			t.Errorf("%q kam durch die Prüfung — dann beweist die Liste nichts", name)
		}
	}
}

func TestDieAngeboteneZeitzonenlisteIstGeordnetUndOhneDoppel(t *testing.T) {
	namen := zeitzonenAusDemFenster(t)

	for stelle := 1; stelle < len(namen); stelle++ {
		if namen[stelle-1] >= namen[stelle] {
			t.Errorf("%q steht vor %q — die Liste ist nicht streng geordnet",
				namen[stelle-1], namen[stelle])
		}
	}
}

func TestDieAngeboteneZeitzonenlisteHatGenugWerteFuerEineSuche(t *testing.T) {
	// R7: ab fünfzehn Werten gehört zur Auswahl eine Suche.
	const suchschwelle = 15

	namen := zeitzonenAusDemFenster(t)

	if len(namen) < suchschwelle {
		t.Errorf("%d Zeitzonen — unter %d brauchte es keine Suche", len(namen), suchschwelle)
	}

	for _, name := range namen {
		if name == "Local" || name == "" {
			t.Errorf("%q gehört nicht in die Liste: leer heißt »wie das System«, und das ist eine eigene Zeile", name)
		}
	}
}
