package rest_test

// **Die abgestufte Rücksetzung am Dienst** — Werkstück O4 von Phase 5.2
// (#1428).

import (
	"net/http"
	"strings"
	"testing"
)

// **Ein Rumpf, der keiner ist, wird abgewiesen** — und der Anwendungsfall
// wird gar nicht erst gerufen.
//
// **Das ist die vierte Klasse aus `endpunkte_test.go`**, hier für den einen
// Weg, den sie nicht abdeckt: dort geht jeder Endpunkt mit **gültigem** Rumpf
// hindurch, und dieser Zweig liegt davor.
func TestZuruecksetzenWeistEinenKrummenRumpfAb(t *testing.T) {
	t.Parallel()

	rumpf := ruf(t, neuerDienst(t), endpunkt{
		methode: http.MethodPost, weg: "/ruecksetzung/layout",
		rumpf: `{"name":`, erfolg: http.StatusBadRequest,
	}, gueltig)

	if rumpf.Code != http.StatusBadRequest {
		t.Errorf("ein krummer Rumpf endete mit %d: %s", rumpf.Code, rumpf.Body.String())
	}

	if !strings.Contains(rumpf.Body.String(), "validierung") {
		t.Errorf("die Antwort nennt die Klasse nicht: %s", rumpf.Body.String())
	}
}
