package rest_test

// **Der Weg des Batches** — Werkstück A13.1 von Phase 5.3
// (#1452).
//
// **Er steht nicht unter einer Karte**, denn er meint mehrere: `POST
// /ki/batch` und nicht `POST /karten/{id}/…`.

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

// **Der Batch kommt am Dienst an.**
func TestDerKIBatchKommtAmDienstAn(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, kidienst(t, williger{}), endpunkt{
		methode: http.MethodPost, weg: "/ki/batch",
		rumpf: `{"karteIds":[4,7],"alias":"beispiel:3b",` +
			`"auftrag":"Schreibe."}`,
	}, gueltig)

	if antwort.Code != http.StatusOK {
		t.Fatalf("der Status ist %d: %s", antwort.Code, antwort.Body.String())
	}

	for _, teil := range []string{`"karteId":4`, `"karteId":7`} {
		if !strings.Contains(antwort.Body.String(), teil) {
			t.Errorf("die Antwort nennt %s nicht: %s", teil, antwort.Body.String())
		}
	}
}

// **Und ein Fehler der Naht kommt als Fehler an.**
func TestEinStolpernderBatchWirdAmWegGemeldet(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, kidienst(t, unwilliger{fehler: errors.New("die Naht schweigt")}),
		endpunkt{
			methode: http.MethodPost, weg: "/ki/batch",
			rumpf: `{"karteIds":[4],"alias":"beispiel:3b",` +
				`"auftrag":"Schreibe."}`,
		}, gueltig)

	if antwort.Code == http.StatusOK {
		t.Fatalf("der Status ist %d", antwort.Code)
	}
}

// **Ein unlesbarer Rumpf wird am Weg abgewiesen.**
func TestEinUnlesbarerBatchrumpfWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, kidienst(t, williger{}), endpunkt{
		methode: http.MethodPost, weg: "/ki/batch", rumpf: `{`,
	}, gueltig)

	if antwort.Code != http.StatusBadRequest {
		t.Fatalf("der Status ist %d", antwort.Code)
	}
}
