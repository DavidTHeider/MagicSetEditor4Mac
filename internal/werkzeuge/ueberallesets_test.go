package werkzeuge_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// **Wer „über alle Sets" liest, liest über die Sets des Profils** —
// ENG-05 (#1730).
//
// # Warum der Wächter darüber nicht reichte
//
// `TestJedeWurzelabfrageKlemmtAufEinProfil` sieht Abfragen auf die zehn
// Wurzeltafeln. Die Suche las `card_suche` und `cards` — Abkömmlinge — und
// ließ mit einer Setkennung 0 („über alle Sets") jede Bedingung fallen; sie
// nannte keine Wurzel und fiel deshalb durch sein Muster. Gemessen am
// 2026-09-23: `mse suche Feuer` fand im eigenen Profil die Karte eines
// anderen Nutzers.
//
// # Was er hält
//
// **Jede Abfrage mit der Aufweitung `(? = 0 OR …)` nennt `profil_id`.** Eine
// Setkennung 0 heißt dann „über die Sets dieses Profils" und nicht „über den
// ganzen Bestand".
//
// # Was er nicht hält
//
// **Eine Kennung, die ein Aufrufer nennt.** Eine Abfrage mit
// `WHERE set_id = ?` ist nur so sicher wie die Kennung, die hineingeht —
// das ist die Abkömmlingsregel aus 0066, und sie steht nicht unter diesem
// Wächter.
func TestUeberAlleSetsHeisstUeberDieSetsDesProfils(t *testing.T) {
	t.Parallel()

	aufweitung := regexp.MustCompile(`\(\?\s*=\s*0\s+OR\b`)

	dateien, err := filepath.Glob(filepath.Join("..", "storage", "*.go"))
	if err != nil {
		t.Fatalf("die Dateien finden: %v", err)
	}

	gesehen := 0

	for _, pfad := range dateien {
		name := filepath.Base(pfad)
		if strings.HasSuffix(name, "_test.go") || begruendet(name) {
			continue
		}

		inhalt, lesefehler := os.ReadFile(pfad) //nolint:gosec // aus Glob
		if lesefehler != nil {
			t.Fatalf("%s lesen: %v", pfad, lesefehler)
		}

		for _, abfrage := range abfragen(string(inhalt)) {
			if !aufweitung.MatchString(abfrage) {
				continue
			}

			gesehen++

			if !strings.Contains(abfrage, "profil_id") {
				t.Errorf("%s: eine Abfrage geht ohne Set über den ganzen Bestand:\n%s",
					name, gekuerzt(abfrage))
			}
		}
	}

	// **Der Wächter hat etwas gesehen**: die drei Suchabfragen in
	// `suche.go` und `markensuche.go` weiten auf.
	if gesehen < 3 {
		t.Errorf("nur %d aufweitende Abfragen gesehen — das Muster greift daneben", gesehen)
	}
}
