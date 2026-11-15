package rest_test

// **Die Blätterangaben des Marktkontos** — Werkstück `T11` von Phase 5.6
// (#1476).
//
// # Was hier geprüft wird
//
// **Dass sie wahlfrei sind und trotzdem geprüft werden.** Wer nichts nennt,
// nimmt die Vorgabe des Anbieters; wer `seite=zwei` schickt, hat sich
// vertan — und still eine andere Seite zu liefern wäre die schlechtere
// Auskunft (§8).

import (
	"net/http"
	"testing"
)

func TestDieBlaetterangabenSindWahlfreiUndGeprueft(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		weg    string
		erfolg int
	}{
		"ohne Angaben": {
			"/markt/cardtrader/konto", http.StatusOK,
		},
		"mit beiden": {
			"/markt/cardtrader/konto?seite=2&menge=50", http.StatusOK,
		},
		"eine krumme Seite": {
			"/markt/cardtrader/konto?seite=zwei", http.StatusBadRequest,
		},
		"eine negative Menge": {
			"/markt/cardtrader/konto?menge=-1", http.StatusBadRequest,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t), endpunkt{
				methode: http.MethodGet, weg: fall.weg,
			}, gueltig)

			if antwort.Code != fall.erfolg {
				t.Errorf("endete mit %d statt %d: %s",
					antwort.Code, fall.erfolg, antwort.Body.String())
			}
		})
	}
}
