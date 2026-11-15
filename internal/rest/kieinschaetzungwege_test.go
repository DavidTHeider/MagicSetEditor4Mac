package rest_test

// **Die Wege der Einschätzung** — Werkstück A13.2a von Phase 5.3
// (#1452).
//
// **Eine eigene Datei**, weil `kiwege_test.go` mit ihnen über die
// Zeilengrenze aus §2 ginge.

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

// einEinschaetzungsweg ist der Ruf, den alle drei Läufe abwandeln.
func einEinschaetzungsweg(weg, rumpf string) endpunkt {
	return endpunkt{methode: http.MethodPost, weg: weg, rumpf: rumpf}
}

// derEinschaetzungsrumpf ist ein gültiger.
const derEinschaetzungsrumpf = `{"art":"regel","alias":"beispiel:3b",` +
	`"frage":"Wie wirkt sie?"}`

// **Die Einschätzung kommt an.**
func TestDieEinschaetzungKommtAmDienstAn(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, kidienst(t, williger{}), einEinschaetzungsweg(
		"/karten/4/ki-einschaetzung", derEinschaetzungsrumpf), gueltig)

	if antwort.Code != http.StatusOK {
		t.Fatalf("der Status ist %d: %s", antwort.Code, antwort.Body.String())
	}

	if !strings.Contains(antwort.Body.String(), `"belegt":true`) {
		t.Errorf("die Antwort ist %s", antwort.Body.String())
	}
}

// **Eine Kartenkennung, die keine Zahl ist, wird am Weg abgewiesen.**
func TestEineUnlesbareKennungBeiDerEinschaetzungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, kidienst(t, williger{}), einEinschaetzungsweg(
		"/karten/vier/ki-einschaetzung", derEinschaetzungsrumpf), gueltig)

	if antwort.Code != http.StatusBadRequest {
		t.Fatalf("der Status ist %d: %s", antwort.Code, antwort.Body.String())
	}
}

// **Ein unlesbarer Rumpf ebenso.**
func TestEinUnlesbarerRumpfBeiDerEinschaetzungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, kidienst(t, williger{}), einEinschaetzungsweg(
		"/karten/4/ki-einschaetzung", `{kaputt`), gueltig)

	if antwort.Code != http.StatusBadRequest {
		t.Fatalf("der Status ist %d: %s", antwort.Code, antwort.Body.String())
	}
}

// **Und ein Fehler der Naht kommt als Fehler durch.**
func TestEinFehlerDerEinschaetzungKommtDurch(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, kidienst(t, unwilliger{fehler: errors.New("die Naht schweigt")}),
		einEinschaetzungsweg("/karten/4/ki-einschaetzung",
			derEinschaetzungsrumpf), gueltig)

	if antwort.Code == http.StatusOK {
		t.Fatalf("der Fehler kam nicht durch: %s", antwort.Body.String())
	}
}
