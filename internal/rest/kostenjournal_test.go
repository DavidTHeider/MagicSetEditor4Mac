package rest_test

// **Das Kostenjournal am Dienst** — Werkstück O12 von Phase 5.2
// (#1439).

import (
	"net/http"
	"strings"
	"testing"
)

// **Ein Budget ohne Währung wird abgewiesen** — ein Betrag ohne Währung ist
// keiner, und die Absage steht vor dem Fachkern.
func TestEinBudgetOhneWaehrungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	rumpf := ruf(t, neuerDienst(t), endpunkt{
		methode: http.MethodPut, weg: "/kosten/anbieter/budget",
		rumpf: `{"zeitraum":"tag","wert":5000}`, erfolg: http.StatusBadRequest,
	}, gueltig)

	if rumpf.Code != http.StatusBadRequest ||
		!strings.Contains(rumpf.Body.String(), "waehrung") {
		t.Errorf("die Antwort ist %d: %s", rumpf.Code, rumpf.Body.String())
	}
}

// **Und ein Rumpf, der keiner ist, ebenso.**
func TestEinKrummerBudgetrumpfWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	rumpf := ruf(t, neuerDienst(t), endpunkt{
		methode: http.MethodPut, weg: "/kosten/anbieter/budget",
		rumpf: `{"wert":`, erfolg: http.StatusBadRequest,
	}, gueltig)

	if rumpf.Code != http.StatusBadRequest {
		t.Errorf("ein krummer Rumpf endete mit %d", rumpf.Code)
	}
}
