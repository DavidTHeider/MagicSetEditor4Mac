package werkzeuge_test

// **Die Seiten der Domäne sind die des Fensters** — Befund G13 des
// Zweitblicks über Phase 6.3
// (#1665).
//
// # Warum die Domäne sie kennt
//
// **Weil die Ziffernbelegung über MCP, Kommandozeile und REST gesetzt wird**,
// und dort gibt es kein Fenster, das eine unbekannte Seite abweisen könnte.
// `3=kartne` kam durch und fiel im Fenster still weg. Die Namen stehen
// deshalb in `internal/domain/ziffernbelegung.go`, und dieser Wächter hält
// sie gegen `navigation.ts` — in beide Richtungen.
//
// # Was er nicht hält
//
// **Ob eine Seite eine Ziffer verdient.** Das entscheidet ein Mensch in der
// Tafel der Kürzel.

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieFensterseitenDerDomaeneSindDieDerNavigation(t *testing.T) {
	t.Parallel()

	navigation := navigationsseiten(t)
	slices.Sort(navigation)

	if domaene := domain.Fensterseiten(); !slices.Equal(domaene, navigation) {
		t.Errorf("internal/domain/ziffernbelegung.go führt\n  %v\nnavigation.ts führt\n  %v", domaene, navigation)
	}
}

// navigationsseiten gibt die Seitenschlüssel, die `navigation.ts` führt —
// Anwendung und Set zusammen.
//
// **Stand bis 1.0.0 im Wächter über den Entwurf der Zieloberfläche**; der ist
// mit dem Entwurf ins Entwicklungsarchiv gegangen (Vermerk 0205), die Liste
// der Seiten nicht.
func navigationsseiten(t *testing.T) []string {
	t.Helper()

	navigation := lies(t, "frontend/src/navigation.ts")

	listen := []string{
		"export const anwendungsschluessel = [",
		"export const setschluessel = [",
	}

	gefuehrt := make([]string, 0, len(listen))

	for _, liste := range listen {
		gefuehrt = append(gefuehrt, schluesselAus(navigation, liste)...)
	}

	// **Eine leere Liste macht diesen Wächter hohl** — dieselbe Untergrenze
	// wie nebenan in `uxarchitektur_test.go`.
	if len(gefuehrt) < 10 {
		t.Fatalf("navigation.ts führt %d Seiten — der Leser greift ins Leere",
			len(gefuehrt))
	}

	return gefuehrt
}
