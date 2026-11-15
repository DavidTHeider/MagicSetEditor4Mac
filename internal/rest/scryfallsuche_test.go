// **Die Kartensuche über die Rückschleife** — Werkstück 1 von Phase 3.6
// (#1167).
package rest_test

import (
	"net/http"
	"strings"
	"testing"
)

// **Ohne Frage gibt es keine Suche — und die Meldung sagt, wohin sie gehört.**
//
// Ein Aufrufer, der sie im Weg sucht, findet sie nie: sie steht in der
// Abfrage, weil ein `GET` keinen Rumpf hat.
func TestEineSucheOhneFrageWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, neuerDienst(t), endpunkt{
		methode: http.MethodGet, weg: "/scryfall/suche",
	}, gueltig)

	if antwort.Code != http.StatusBadRequest {
		t.Errorf("Status %d, erwartet war 400: %s", antwort.Code, antwort.Body.String())
	}

	if !strings.Contains(antwort.Body.String(), "frage") {
		t.Errorf("die Meldung nennt das Feld nicht: %s", antwort.Body.String())
	}
}

// **Und eine leere Trefferliste ist `[]` und nicht `null`.**
//
// Dieselbe Falle wie überall: ein `null` bricht jeden Aufrufer, der darüber
// läuft, ohne vorher zu prüfen.
func TestEineLeereTrefferlisteIstEineListe(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, dienstMit(t, &zaehlender{}), endpunkt{
		methode: http.MethodGet, weg: "/scryfall/suche?frage=bolt",
	}, gueltig)

	if antwort.Code != http.StatusOK {
		t.Fatalf("Status %d: %s", antwort.Code, antwort.Body.String())
	}

	if !strings.Contains(antwort.Body.String(), `"treffer":[]`) {
		t.Errorf("die leere Liste ist keine: %s", antwort.Body.String())
	}
}
