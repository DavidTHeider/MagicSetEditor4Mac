// Der Wächter über das, was der Entwicklungsserver **nicht** beobachtet —
// Befund aus der Sichtprüfung zu M5 am 2026-09-17.
//
// # Wofür es ihn gibt
//
// **Weil eine Sichtprüfung nach einem Abdeckungslauf unbrauchbar war.** Der
// Riegel (`make vorpruefung`) fährt `vitest --coverage`, und der schreibt
// `frontend/coverage/lcov-report/` — **rund 900 HTML-Dateien**, eine je
// Quelltextdatei. Vite beobachtet sein Wurzelverzeichnis und hielt jede davon
// für eine Änderung am Programm: das Fenster lud in Folge hunderte Male neu,
// und in der Zeit ließ sich nichts ansehen.
//
// **Es ist kein Fehler des Riegels und keiner von Vite.** Beide tun, was sie
// sollen; falsch ist nur, dass der eine im Hof des anderen ablegt. Weil beide
// bleiben, muss die Ausnahme in der Konfiguration stehen — und dort steht sie
// still, bis jemand sie beim Aufräumen entfernt.
//
// # Was er nicht prüft
//
// **Ob Vite die Ausnahme befolgt.** Er sieht, dass sie in der Konfiguration
// steht, nicht dass der Beobachter sie liest — das wäre nur mit einem
// laufenden Server zu messen, und die Sichtprüfung hat es getan (§3). Und er
// prüft **nicht**, ob die Liste vollständig ist: ein drittes Werkzeug, das
// künftig in `frontend/` ablegt, kennt er nicht.
package werkzeuge

import (
	"strings"
	"testing"
)

// **Was ein Werkzeug in `frontend/` ablegt und kein Mensch schreibt.**
//
// `dist/` steht dabei, obwohl der Bau nur beim Freigeben läuft: er legt
// dieselbe Sorte Datei ab, und eine Ausnahme, die erst nach dem Schaden
// dazukommt, ist eine, die jemand einmal erlebt hat.
var unbeobachtet = []string{"coverage", "dist"}

func TestDerEntwicklungsserverBeobachtetKeineWerkzeugablage(t *testing.T) {
	t.Parallel()

	inhalt := strings.ReplaceAll(
		string(projektdatei(t, "frontend/vite.config.ts")), `"`, "'")

	for _, ablage := range unbeobachtet {
		if !strings.Contains(inhalt, "'**/"+ablage+"/**'") {
			t.Errorf("vite.config.ts nimmt %q nicht vom Beobachten aus — "+
				"ein Lauf, der dort ablegt, lädt das Fenster in Folge neu "+
				"und macht jede Sichtprüfung unbrauchbar", ablage)
		}
	}
}
