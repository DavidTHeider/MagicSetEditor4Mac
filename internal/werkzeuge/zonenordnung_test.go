// **Geordnet heißt im Fenster dasselbe wie im Kern** — Befund **F04**
// (#1748).
//
// # Warum das Fenster die Frage überhaupt stellt
//
// Weil eine Stelle in einer ungeordneten Zone ein Irrtum ist und kein Wunsch:
// `service.gemeinteStelle` weist sie ab. Richtig ist das — falsch war, dass
// das Formular sie überhaupt anbot. Seit F04 fragt es die Stelle nur, wenn die
// **Zielzone** geordnet ist, und dafür muss es wissen, welche das sind.
//
// # Warum diese Antwort gewacht gehört
//
// **Weil es die zweite Stelle ist, an der sie steht** — und zwei Antworten auf
// dieselbe Frage sind schon einmal auseinandergelaufen: bis zum 2026-09-17
// galt der Friedhof in `internal/domain` als ungeordnet, während
// `frontend/src/tisch/zonenfeld.ts` daneben CR 400.5 richtig wiedergab. Der
// Widerspruch stand ein Vierteljahr da und kostete „oben auf den Friedhof",
// also den Abwurf im Aufräumschritt (CR 514.1).
//
// # Was er hält
//
// `frontend/src/spiel/zonenordnung.ts` führt genau die Zonen, für die
// `domain.Zone.Geordnet()` wahr sagt — in **beide** Richtungen.
//
// # Was er **nicht** hält
//
//   - **Nicht, dass die Antwort stimmt.** Ob Bibliothek, Friedhof und Stapel
//     die richtigen drei sind, sagt CR 400.5 und kein Wächter; dass sie
//     nachgelesen ist, steht als Regelnummer am Code (§11).
//   - **Nicht `tisch/zonenfeld.ts`.** Dessen `geordnet` beschreibt ein
//     gezeichnetes Feld und führt die Zonen eines **Sitzes**; der gemeinsame
//     Stapel kommt dort gar nicht vor. Es gegen dieselbe Liste zu halten
//     verlangte von ihm eine Vollständigkeit, die es nicht hat — und der
//     Wächter wäre rot, ohne dass etwas falsch wäre.
package werkzeuge

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// geordneteImFenster liest die Liste aus `spiel/zonenordnung.ts`.
var geordneteImFenster = regexp.MustCompile(
	`(?s)export const geordneteZonen = \[(.*?)\] as const`)

// TestDasFensterKenntDieselbenGeordnetenZonen hält F04s zweite Hälfte.
func TestDasFensterKenntDieselbenGeordnetenZonen(t *testing.T) {
	t.Parallel()

	imKern := make([]string, 0, len(domain.Zonen()))

	for _, zone := range domain.Zonen() {
		if zone.Geordnet() {
			imKern = append(imKern, string(zone))
		}
	}

	if len(imKern) == 0 {
		t.Fatal("der Kern nennt keine geordnete Zone — der Leser greift daneben")
	}

	imFenster := geordneteZonenLesen(t)

	sort.Strings(imKern)
	sort.Strings(imFenster)

	if strings.Join(imKern, ",") != strings.Join(imFenster, ",") {
		t.Errorf("der Kern nennt %v geordnet, das Fenster %v — zwei Antworten "+
			"auf dieselbe Frage, und die Stellenwahl hängt daran (CR 400.5)",
			imKern, imFenster)
	}
}

// geordneteZonenLesen gibt die Zonen aus der Liste des Fensters.
func geordneteZonenLesen(t *testing.T) []string {
	t.Helper()

	pfad := filepath.Join(wurzel, "frontend", "src", "spiel", "zonenordnung.ts")

	roh, fehler := os.ReadFile(pfad) // #nosec G304 -- fester Pfad
	if fehler != nil {
		t.Fatalf("%s lesen: %v", pfad, fehler)
	}

	treffer := geordneteImFenster.FindSubmatch(roh)
	if treffer == nil {
		t.Fatalf("%s führt kein `export const geordneteZonen = […] as const` "+
			"mehr — der Wächter greift daneben, nicht der Code", pfad)
	}

	var gefunden []string

	for _, teil := range strings.Split(string(treffer[1]), ",") {
		name := strings.Trim(strings.TrimSpace(teil), "'\"")
		if name != "" {
			gefunden = append(gefunden, name)
		}
	}

	return gefunden
}
