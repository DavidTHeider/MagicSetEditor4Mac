// **Die Beispiele der Konsole laufen wirklich** — Werkstück 1 von Phase 3.3
// (#1164).
//
// # Warum dieser Wächter sein muss
//
// Weil die Zusage im Namen steckt: „Beispiele, die laufen". Der erste Entwurf
// hatte sechs, und **fünf** davon gaben auf einem Rechner ohne eingebundenes
// Stilpaket einen Fehler — `card` und `set` legt die Vorlage in die Umgebung,
// und ohne sie bleibt die Bibliothek (§8). Gerade der erste Klick trifft den,
// der auf einem frischen Rechner sitzt.
//
// Gefunden hat es kein Test, sondern ein Lauf über `mse skript`. Diese Datei
// macht daraus eine Zusage, die hält: was das Fenster als „läuft überall"
// anbietet, wird hier ausgewertet.
//
// # Was er nicht hält
//
// **Die zweite Gruppe.** `mitVorlage` braucht ein Stilpaket, und welches, hängt
// am Set; sie hier zu fahren hieße, ein Paket zu wählen und damit etwas
// anderes zu prüfen. Dass sie eine Vorlage braucht, **sagt die Oberfläche** —
// und das ist der Unterschied zur ersten Gruppe.
package service_test

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// beispielzeile findet `ausdruck: '…'` in der TypeScript-Datei.
//
// **Gelesen und nicht abgeschrieben.** Eine zweite Liste hier liefe an dem Tag
// auseinander, an dem jemand drüben ein Beispiel ändert — und der Wächter
// prüfte dann Ausdrücke, die niemandem angeboten werden (§10).
var beispielzeile = regexp.MustCompile(`ausdruck: '([^']*)'`)

func TestJedesImmerLaufendeBeispielLaeuft(t *testing.T) {
	t.Parallel()

	roh, err := os.ReadFile("../../frontend/src/seiten/beispiele.ts")
	if err != nil {
		t.Fatalf("beispiele.ts lesen: %v", err)
	}

	// **Nur die erste Gruppe.** Sie endet, wo `mitVorlage` beginnt.
	quelle := string(roh)

	schnitt := strings.Index(quelle, "export const mitVorlage")
	if schnitt < 0 {
		t.Fatal("die zweite Gruppe heißt nicht mehr `mitVorlage` — greift das Muster noch?")
	}

	treffer := beispielzeile.FindAllStringSubmatch(quelle[:schnitt], -1)
	if len(treffer) < 3 {
		t.Fatalf("nur %d Beispiele gefunden — das Muster greift daneben", len(treffer))
	}

	// **Ohne Planbauer**, und das ist der Punkt: so sieht die Umgebung auf
	// einem Rechner ohne eingebundenes Stilpaket aus.
	konsole := service.NeueSkriptkonsole(
		konsolenkarten{karte: &domain.CardRow{ID: 7}}, nil)

	for _, gefunden := range treffer {
		ausdruck := gefunden[1]

		t.Run(ausdruck, func(t *testing.T) {
			t.Parallel()

			if _, fehler := konsole.Auswerten(t.Context(), 7, ausdruck); fehler != nil {
				t.Errorf("%s", fehler)
			}
		})
	}
}

// **Und die zweite Gruppe läuft ohne Vorlage nicht** — die Gegenprobe.
//
// Ohne sie sagte der Lauf darüber auch dann ja, wenn die Trennung willkürlich
// wäre: dass `card` eine Vorlage braucht, ist der ganze Grund für die zwei
// Gruppen.
func TestDieBeispieleMitVorlageBrauchenEine(t *testing.T) {
	t.Parallel()

	roh, err := os.ReadFile("../../frontend/src/seiten/beispiele.ts")
	if err != nil {
		t.Fatalf("beispiele.ts lesen: %v", err)
	}

	quelle := string(roh)

	schnitt := strings.Index(quelle, "export const mitVorlage")
	if schnitt < 0 {
		t.Fatal("die zweite Gruppe heißt nicht mehr `mitVorlage`")
	}

	treffer := beispielzeile.FindAllStringSubmatch(quelle[schnitt:], -1)
	if len(treffer) < 2 {
		t.Fatalf("nur %d Beispiele gefunden — das Muster greift daneben", len(treffer))
	}

	konsole := service.NeueSkriptkonsole(
		konsolenkarten{karte: &domain.CardRow{ID: 7}}, nil)

	for _, gefunden := range treffer {
		ausdruck := gefunden[1]

		t.Run(ausdruck, func(t *testing.T) {
			t.Parallel()

			if _, fehler := konsole.Auswerten(t.Context(), 7, ausdruck); fehler == nil {
				t.Errorf("%s lief ohne Vorlage — dann gehört es in die erste Gruppe", ausdruck)
			}
		})
	}
}
