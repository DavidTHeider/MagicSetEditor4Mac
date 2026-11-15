// **Die Markenwege des Dienstes** — Werkstück 3 von Phase 3.5
// (#1208).
package rest_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// **Eine Karte ohne Marken gibt `[]` und nicht `null`.**
//
// Ein `null` in einer Liste zwingt jeden Aufrufer, zwei Gestalten desselben
// Nichts zu unterscheiden — und die Vorrichtung gibt hier absichtlich `nil`.
func TestOhneMarkenKommtEineLeereListe(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, neuerDienst(t), endpunkt{
		methode: http.MethodGet, weg: "/sets/1/karten/2/marken",
	}, gueltig)

	if antwort.Code != http.StatusOK {
		t.Fatalf("endete mit %d: %s", antwort.Code, antwort.Body.String())
	}

	if !strings.Contains(antwort.Body.String(), `"marken":[]`) {
		t.Errorf("Rumpf: %s", antwort.Body.String())
	}
}

// **Eine leere Ortsliste kommt als leere Liste an und nicht als `null`** —
// #1291.
//
// Der Zweig, der `nil` zu `[]` macht: ohne ihn stünde im Fenster `undefined`
// statt einer leeren Liste (§3). Auf einem Rechner ohne MSE2 ist das der
// Normalfall, nicht der Ausnahmefall.
func TestEineLeereOrtslisteKommtAlsListeAn(t *testing.T) {
	t.Parallel()

	dienst := rest.NeuerDienst(probemarke,
		func() time.Time { return time.Date(2026, time.August, 30, 12, 0, 0, 0, time.UTC) },
		rest.Naehte{Pakete: ohneOrte{}})

	antwort := ruf(t, dienst, endpunkt{methode: http.MethodGet, weg: "/mse2/orte"}, gueltig)

	if antwort.Code != http.StatusOK {
		t.Fatalf("Status %d: %s", antwort.Code, antwort.Body.String())
	}

	if strings.Contains(antwort.Body.String(), "null") {
		t.Errorf("die Orte kamen als null an: %s", antwort.Body.String())
	}
}

// ohneOrte ist ein Rechner, auf dem kein MSE2 liegt.
type ohneOrte struct{ williger }

func (ohneOrte) MSE2Orte() []string { return nil }
