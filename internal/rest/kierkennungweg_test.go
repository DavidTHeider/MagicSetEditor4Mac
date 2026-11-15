package rest_test

// **Der Weg der Kartenerkennung** — Werkstück A13.3 von Phase 5.3
// (#1452).
//
// **Er steht nicht unter einer Karte**: ein Foto wird gelesen, **bevor**
// irgendeine Karte feststeht — und welche es wird, entscheidet ein Mensch.
// Ein Weg unter `/karten/{id}` behauptete eine Zugehörigkeit, deren
// Abwesenheit der ganze Punkt dieses Werkstücks ist.

import (
	"net/http"
	"strings"
	"testing"
)

// **Ein kaputter Rumpf kommt nicht bis zur Werkstatt.**
//
// **Das ist die Messung**: eine Anfrage, die Geld kostet, darf nicht an einer
// Zeichenkette losgehen, die kein Pfad ist.
func TestEinKaputterErkennungsrumpfWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, kidienst(t, williger{}), endpunkt{
		methode: http.MethodPost, weg: "/ki/erkennung",
		rumpf: `{"pfad": 17}`,
	}, gueltig)

	if antwort.Code == http.StatusOK {
		t.Fatalf("der kaputte Rumpf kam durch: %s", antwort.Body.String())
	}

	if !strings.Contains(antwort.Body.String(), "validierung") {
		t.Errorf("die Antwort ist %s", antwort.Body.String())
	}
}
