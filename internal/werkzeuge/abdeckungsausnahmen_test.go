// Wächter über die Abdeckungsausnahmen (AGENTS.md §3).
//
// Sie stehen an zwei Orten, und das lässt sich nicht vermeiden: `.covignore`
// führt sie samt Begründung, `frontend/vite.config.ts` setzt sie für Vitest
// durch — ein Shellskript kann eine Vitest-Schwelle nicht beeinflussen.
//
// Zwei Orte laufen auseinander, sobald niemand nachzählt. Der Zweitblick am
// Ende von Phase 1 hat genau das gefunden: `scripts/ci-abdeckung.sh` filtert
// die Frontend-Zeilen aus `.covignore` gegen das **Go**-Profil, wo sie nie
// vorkommen. Wer dort einen Eintrag entfernte, änderte an der wirklichen
// Ausnahme nichts, und wer einen hinzufügte, bekam keine.
//
// Also zählt hier jemand nach.
package werkzeuge

import (
	"regexp"
	"slices"
	"strings"
	"testing"
)

// vitestAusnahme findet die Einträge im exclude-Feld der Abdeckung.
var vitestAusnahme = regexp.MustCompile(`(?s)exclude:\s*\[(.*?)\]`)

var eintrag = regexp.MustCompile(`'([^']+)'`)

// Der Abgleich läuft in **beide** Richtungen.
//
// Der Nachlauf des Zweitblicks hat gezeigt, dass eine Richtung die halbe Zusage
// ist: wer `frontend/src/main.tsx` aus `.covignore` entfernte, ließ die
// Vitest-Ausnahme unberührt — die Datei wäre weiter ausgenommen gewesen, ohne
// dass irgendwo eine Begründung dafür stünde.
func TestFrontendausnahmenStehenInBeidenDateien(t *testing.T) {
	ausCovignore := frontendausnahmen(t)
	ausVite := viteausnahmen(t)

	for _, pfad := range ausCovignore {
		t.Run(".covignore führt "+pfad, func(t *testing.T) {
			if !slices.Contains(ausVite, alsVitemuster(pfad)) {
				t.Errorf(".covignore nimmt %q aus, frontend/vite.config.ts kennt dafür keinen "+
					"Ausschluss %q — die Ausnahme wirkt also nicht", pfad, alsVitemuster(pfad))
			}
		})
	}

	erwartet := make([]string, 0, len(ausCovignore))
	for _, pfad := range ausCovignore {
		erwartet = append(erwartet, alsVitemuster(pfad))
	}

	for _, muster := range ausVite {
		if istTestmuster(muster) {
			continue
		}

		t.Run("vite.config.ts führt "+muster, func(t *testing.T) {
			if !slices.Contains(erwartet, muster) {
				t.Errorf("frontend/vite.config.ts nimmt %q aus, .covignore nennt dafür keinen "+
					"Eintrag — die Ausnahme hat damit keine Begründung (AGENTS.md §3)", muster)
			}
		})
	}
}

// alsVitemuster bildet einen .covignore-Pfad auf die Schreibweise ab, in der
// Vitest ihn erwartet.
//
// Eindeutig und nicht als Präfixvergleich: `src/main` und `src/main.tsx` sind
// zwei verschiedene Dinge, und ein Präfixtreffer hätte sie gleich behandelt.
func alsVitemuster(pfad string) string {
	ohneFrontend := strings.TrimPrefix(pfad, "frontend/")

	if strings.HasSuffix(ohneFrontend, "/") {
		return ohneFrontend + "**"
	}

	return ohneFrontend
}

// istTestmuster erkennt die Ausschlüsse, die zu den Tests selbst gehören.
//
// Sie stehen nicht in `.covignore`, weil sie keine Ausnahme von der Regel sind:
// eine Testdatei ist kein Produktivcode, und sie zu messen ergäbe eine Zahl
// über die Tests statt über die Anwendung.
func istTestmuster(muster string) bool {
	return strings.Contains(muster, ".test.")
}

// frontendausnahmen liest die Zeilen aus .covignore, die das Frontend betreffen.
func frontendausnahmen(t *testing.T) []string {
	t.Helper()

	var pfade []string

	for _, zeile := range strings.Split(string(projektdatei(t, ".covignore")), "\n") {
		gestutzt := strings.TrimSpace(zeile)
		if gestutzt == "" || strings.HasPrefix(gestutzt, "#") {
			continue
		}

		pfad := strings.TrimSpace(strings.SplitN(gestutzt, "#", 2)[0])
		if strings.HasPrefix(pfad, "frontend/") {
			pfade = append(pfade, pfad)
		}
	}

	if len(pfade) == 0 {
		t.Fatal(".covignore führt keine Frontend-Ausnahme — hat sich der Aufbau geändert?")
	}

	return pfade
}

// viteausnahmen liest das exclude-Feld aus frontend/vite.config.ts.
func viteausnahmen(t *testing.T) []string {
	t.Helper()

	treffer := vitestAusnahme.FindSubmatch(projektdatei(t, "frontend/vite.config.ts"))
	if treffer == nil {
		t.Fatal("frontend/vite.config.ts nennt kein exclude in der Abdeckung")
	}

	funde := eintrag.FindAllStringSubmatch(string(treffer[1]), -1)

	muster := make([]string, 0, len(funde))
	for _, fund := range funde {
		muster = append(muster, fund[1])
	}

	if len(muster) == 0 {
		t.Fatal("das exclude-Feld ist leer — hat sich die Schreibweise geändert?")
	}

	return muster
}
