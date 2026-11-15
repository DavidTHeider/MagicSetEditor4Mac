package rest_test

// **Der Weg der atomaren Gruppe** — Werkstück A13.1b von Phase 5.3
// (#1452).
//
// **Er steht nicht unter einer Karte**, denn welche Karten er trifft,
// entscheidet erst der Server aus den Aufträgen.

import (
	"net/http"
	"strings"
	"testing"
)

// **Ein kaputter Rumpf kommt nicht bis zur Werkstatt.**
//
// **Das ist die Messung**: eine Übernahme, die Karten ändert, darf nicht an
// einer Zeichenkette losgehen, die keine Wahl ist.
func TestEinKaputterGruppenrumpfWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, kidienst(t, williger{}), endpunkt{
		methode: http.MethodPost, weg: "/ki/gruppe",
		rumpf: `{"wahl": "keine Liste"}`,
	}, gueltig)

	if antwort.Code == http.StatusOK {
		t.Fatalf("der kaputte Rumpf kam durch: %s", antwort.Body.String())
	}

	if !strings.Contains(antwort.Body.String(), "validierung") {
		t.Errorf("die Antwort ist %s", antwort.Body.String())
	}
}
