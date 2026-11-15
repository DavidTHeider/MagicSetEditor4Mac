// **Der Endpunkt, der eine Paketdatei prüft** — Werkstück B von Phase 4.9
// (#1324).
//
// # Was `endpunktliste_test.go` nicht prüft
//
// Dort steht dieser Weg mit seinen vier Klassen — Marke, Herkunft, gültiger
// Aufruf, Drossel. Was der gültige Aufruf **antwortet**, prüft er nicht: er
// zählt, dass der Anwendungsfall gerufen wurde, und die Zählnaht meldet
// keinen Befund. Hier steht die andere Hälfte.
package rest_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// **Ein Befund kommt mit Zeile, Spalte und Satz an.**
func TestDiePaketpruefungReichtIhrenBefundDurch(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)

	// `williger.Pruefen` meldet Zeile 3, Spalte 7 — die Naht selbst prüft
	// nichts, und das ist hier richtig: geprüft wird der **Weg**, nicht der
	// Zerteiler. Der steht in `internal/service`.
	dienst := rest.NeuerDienst(probemarke, func() time.Time { return jetzt },
		rest.Naehte{Werkstatt: williger{}})

	antwort := ruf(t, dienst, endpunkt{
		methode: http.MethodPost, weg: "/paketdateien/pruefung",
		rumpf: `{"inhalt":"script: f(1, 2"}`, erfolg: http.StatusOK,
	}, gueltig)

	if antwort.Code != http.StatusOK {
		t.Fatalf("Status %d (%s)", antwort.Code, antwort.Body.String())
	}

	for _, erwartet := range []string{
		`"gefunden":true`, `"zeile":3`, `"spalte":7`, `hier fehlt ein Wert`,
	} {
		if !strings.Contains(antwort.Body.String(), erwartet) {
			t.Errorf("die Antwort enthält %s nicht: %s", erwartet, antwort.Body.String())
		}
	}
}

// **Und ein unlesbarer Rumpf wird abgewiesen, nicht geprüft.**
//
// Die Gegenprobe: ohne sie wäre der Weg auch dann grün, wenn er jeden Rumpf
// als leeren Text durchreichte — und dann meldete er „kein Fehler" für eine
// Anfrage, die nie angekommen ist.
func TestEinUnlesbarerRumpfWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)

	dienst := rest.NeuerDienst(probemarke, func() time.Time { return jetzt },
		rest.Naehte{Werkstatt: williger{}})

	antwort := ruf(t, dienst, endpunkt{
		methode: http.MethodPost, weg: "/paketdateien/pruefung",
		rumpf: `{"inhalt":`, erfolg: http.StatusOK,
	}, gueltig)

	if antwort.Code != http.StatusBadRequest {
		t.Errorf("Status %d, erwartet 400: %s", antwort.Code, antwort.Body.String())
	}
}
