// **Was drei Werkzeuge zusagen, ohne dass es jemand nachzählen kann.**
//
// Aus `werkzeugumfang_test.go` herausgelöst, als die Datei über 400 Zeilen
// wuchs (§2) — der Anlass war die Zusage des Löschens
// (#1218).
// Der Schnitt folgt der Frage: dort steht, **welche Zusage** ein Werkzeug
// trägt, hier, ob sein **Text** sie ausspricht.

package mcp_test

import (
	"strings"
	"testing"
)

// **Und keines der drei überschreibt etwas.**
//
// Der Wächter darüber kann es nicht sehen — er liest Namen, keine Wirkung.
// **Was er nicht hält, steht hier**: dass `sichern` und `zeichnen` ein
// vorhandenes Ziel wirklich abweisen, prüft `cmd/mse-mcp`, wo die Naht
// zusammengesetzt wird; dass `einfahren` kein zweites Set anlegt, prüft
// `internal/service/schonoffen_test.go`.
//
// Hier steht nur, dass jedes der drei es in seinem **Text** zusagt — denn
// danach richtet sich ein Sprachmodell, bevor es ruft.
func TestDieDreiSagenZuWasSieNichtTun(t *testing.T) {
	zusagen := map[string]string{
		"schreiben.set.sichern":    "abgewiesen",
		"schreiben.karte.zeichnen": "abgewiesen",
		"schreiben.set.einfahren":  "zweites Mal",
	}

	gesehen := 0

	for _, werkzeug := range werkzeugliste(t) {
		name := zeichenkette(t, werkzeug["name"])

		zusage, gefragt := zusagen[name]
		if !gefragt {
			continue
		}

		gesehen++

		if !strings.Contains(zeichenkette(t, werkzeug["description"]), zusage) {
			t.Errorf("%q sagt %q nicht zu: %s",
				name, zusage, zeichenkette(t, werkzeug["description"]))
		}
	}

	if gesehen != len(zusagen) {
		t.Errorf("%d der drei gesehen, erwartet %d", gesehen, len(zusagen))
	}
}
