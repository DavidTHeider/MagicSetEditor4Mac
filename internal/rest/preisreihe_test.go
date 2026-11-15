package rest_test

// **Eine halbe Preisreihe wird abgewiesen** — Werkstück A3 von Phase 5.3
// (#1442).
//
// Der Weg hieß `GET /drucke/{printing}/preise/{waehrung}/{finish}`, und die
// drei fehlenden Teile ergänzte die Naht mit „Scryfall, Papier, Verkauf".
// Solange es eine Quelle gab, war das richtig geraten; mit einer zweiten ist
// es eine **Preisstrategie**, und A3 verbietet sie: „Quellenwahl explizit".

import (
	"net/http"
	"strings"
	"testing"
)

func TestEineHalbePreisreiheWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	dienst := neuerDienst(t)

	faelle := map[string]struct {
		abfrage string
		nennt   string
	}{
		"gar nichts": {"", "art, finish, markt, quelle, waehrung"},
		"nur Währung und Ausführung": {
			"?waehrung=EUR&finish=nonfoil", "art, markt, quelle",
		},
		"ohne Preisart": {
			"?quelle=scryfall&markt=paper&waehrung=EUR&finish=nonfoil", "art",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, dienst, endpunkt{
				methode: http.MethodGet,
				weg:     "/drucke/probe-druck/preisverlauf" + fall.abfrage,
			}, gueltig)

			if antwort.Code != http.StatusBadRequest {
				t.Fatalf("Status %d, erwartet 400", antwort.Code)
			}

			if !strings.Contains(antwort.Body.String(), fall.nennt) {
				t.Errorf("die Meldung nennt %q nicht: %s",
					fall.nennt, antwort.Body.String())
			}
		})
	}

	// **Die Gegenprobe**: vollständig geht es durch.
	//
	// **Zustand und Sprache fehlen dabei**, und das ist kein Mangel: leer ist
	// dort eine Angabe („unbekannt"), und heute füllt sie keine Quelle.
	antwort := ruf(t, dienst, endpunkt{
		methode: http.MethodGet,
		weg: "/drucke/probe-druck/preisverlauf?quelle=scryfall&markt=paper" +
			"&art=retail&waehrung=EUR&finish=nonfoil",
	}, gueltig)

	if antwort.Code != http.StatusOK {
		t.Errorf("die vollständige Reihe: %d (%s)",
			antwort.Code, antwort.Body.String())
	}
}
