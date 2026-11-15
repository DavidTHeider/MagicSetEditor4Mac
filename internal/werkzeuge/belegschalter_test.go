package werkzeuge_test

// **Jede Umgebungsvariable, die ein Messlauf liest, ist im Plan eingeordnet**
// — als Eingabe, als Schalter oder ohne Wirkung, je mit Grund. Nachlauf zum
// Zweitblick über Phase 6.0: `MSE_STICHPROBE` bestimmt, wie viele Karten der
// feste Kanonlauf misst, und stand weder im Plan noch im Beleg. Ein Lauf mit
// `alle` und eine spätere Prüfung ohne den Schalter hießen »aktuell«.
//
// # Was er hält
//
//   - **Jedes `os.Getenv("MSE…")` und `os.LookupEnv("MSE…")`** in den
//     Testdateien der Pakete, in denen der Plan misst — markiert oder nicht,
//     denn beide übersetzen in dieselbe Testbinärdatei —, steht im Plan.
//   - **Und jeder Eintrag im Plan wird gelesen** — ein Schalter, den niemand
//     mehr fragt, wäre eine Zusage über nichts.
//
// # Was er nicht hält
//
// **Ob die Einordnung stimmt**: dass `MSE_AUSGABE` nur ein Ablageort ist,
// steht im Grund daneben und nicht in einem Test. **Und eine Variable aus
// einer Veränderlichen** — `os.Getenv(name)` — sieht er nicht.

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/laufbeleg"
)

// umgebungsfrage trifft eine gelesene Umgebungsvariable dieses Projekts.
var umgebungsfrage = regexp.MustCompile(`os\.(?:Getenv|LookupEnv)\("(MSE2?_[A-Z0-9_]+)"\)`)

// planUmgebung ist, was ein Plan über die Umgebung sagt.
type planUmgebung struct {
	Tag      string `json:"tag"`
	Eingaben []struct {
		Name string `json:"name"`
	} `json:"eingaben"`
	Schalter    map[string]string `json:"schalter"`
	OhneWirkung map[string]string `json:"ohneWirkung"`
}

func TestJedeGeleseneUmgebungStehtImPlan(t *testing.T) {
	t.Parallel()

	wurzel := filepath.Join("..", "..")

	plaene, fehler := filepath.Glob(filepath.Join(wurzel, "docs", "belege", "*.plan.json"))
	if fehler != nil || len(plaene) == 0 {
		t.Fatalf("keine Pläne unter docs/belege: %v", fehler)
	}

	for _, planpfad := range plaene {
		roh, fehler := os.ReadFile(planpfad) //nolint:gosec // ein Plan aus dem Baum
		if fehler != nil {
			t.Fatal(fehler)
		}

		var plan planUmgebung
		if fehler := json.Unmarshal(roh, &plan); fehler != nil {
			t.Fatalf("%s: %v", planpfad, fehler)
		}

		eingeordnet := map[string]bool{}
		for _, eingabe := range plan.Eingaben {
			eingeordnet[eingabe.Name] = true
		}

		for _, name := range slices.Concat(slices.Collect(maps.Keys(plan.Schalter)),
			slices.Collect(maps.Keys(plan.OhneWirkung))) {
			eingeordnet[name] = true
		}

		gelesen := umgebungImLauf(t, wurzel, plan.Tag)

		for name, woher := range gelesen {
			if !eingeordnet[name] {
				t.Errorf("%s liest %s — %s ordnet es nicht ein: Eingabe, Schalter oder ohne Wirkung?",
					woher, name, filepath.Base(planpfad))
			}
		}

		for name := range eingeordnet {
			if _, da := gelesen[name]; !da {
				t.Errorf("%s nennt %s, aber kein Messlauf liest es", filepath.Base(planpfad), name)
			}
		}
	}
}

// umgebungImLauf gibt jede Umgebungsvariable, die eine Testdatei in einem
// Paket mit markierten Fällen liest, mit der Datei, die sie liest — **wenn
// der Lauf sie übersetzt** (`uebersetztMit`): eine Datei hinter einem
// fremden Tag liest in diesem Lauf nichts.
func umgebungImLauf(t *testing.T, wurzel, tag string) map[string]string {
	t.Helper()

	modul, fehler := laufbeleg.Modul(wurzel)
	if fehler != nil {
		t.Fatal(fehler)
	}

	faelle, fehler := laufbeleg.Markierte(wurzel, tag)
	if fehler != nil || len(faelle) == 0 {
		t.Fatalf("keine markierten Fälle hinter %q: %v", tag, fehler)
	}

	gelesen := map[string]string{}

	for _, paket := range kompakteListe(faelle) {
		verzeichnis := filepath.Join(wurzel, strings.TrimPrefix(strings.TrimPrefix(paket, modul), "/"))

		dateien, fehler := filepath.Glob(filepath.Join(verzeichnis, "*_test.go"))
		if fehler != nil || len(dateien) == 0 {
			t.Fatalf("keine Testdateien unter %s: %v", verzeichnis, fehler)
		}

		for _, datei := range dateien {
			roh, lesefehler := os.ReadFile(datei) //nolint:gosec // Testdateien im Baum
			if lesefehler != nil {
				t.Fatal(lesefehler)
			}

			if !uebersetztMit(string(roh), tag) {
				continue
			}

			woher, _ := filepath.Rel(wurzel, datei) //nolint:errcheck // datei liegt unter der Wurzel

			for _, treffer := range umgebungsfrage.FindAllStringSubmatch(string(roh), -1) {
				gelesen[treffer[1]] = filepath.ToSlash(woher)
			}
		}
	}

	return gelesen
}

// kompakteListe gibt die Pakete der Fälle, jedes einmal.
func kompakteListe(faelle []laufbeleg.Fall) []string {
	pakete := make([]string, 0, len(faelle))
	for _, fall := range faelle {
		pakete = append(pakete, fall.Paket)
	}

	slices.Sort(pakete)

	return slices.Compact(pakete)
}
