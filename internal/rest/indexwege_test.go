package rest_test

// **Die Ränder der Suchindexwege** — Werkstück A15 von Phase 5.3
// (#1514).
//
// **Zwei Ränder, und beide gehören dem Dienst und nicht dem Fachkern**: ein
// Rumpf, der kein JSON ist, und eine Trefferzahl, die keine Zahl ist.

import (
	"net/http"
	"strings"
	"testing"
)

// **Ein unlesbarer Rumpf wird abgewiesen, bevor irgendetwas gebaut wird.**
func TestEinUnlesbarerRumpfBautKeinenIndex(t *testing.T) {
	t.Parallel()

	gezaehlt := &zaehlender{}

	antwort := ruf(t, dienstMit(t, gezaehlt), endpunkt{
		name: "Indexaufbau", methode: http.MethodPost,
		weg: "/index/aufbau", rumpf: `{`,
	}, gueltig)

	if antwort.Code != http.StatusBadRequest {
		t.Fatalf("der kaputte Rumpf gab %d", antwort.Code)
	}

	if gezaehlt.wieOft() != 0 {
		t.Errorf("die Naht wurde %dmal gerufen", gezaehlt.wieOft())
	}
}

// **Eine Trefferzahl, die keine ist, wird zu zehn.**
//
// **Und nicht zu null.** `?grenze=viele` ist ein Tippfehler, kein Wunsch nach
// null Treffern — eine Suche, die deshalb nichts fände, sähe aus wie ein
// leerer Index.
func TestEineUnleserlicheGrenzeWirdZuZehn(t *testing.T) {
	t.Parallel()

	for name, abfrage := range map[string]string{
		"keine Zahl": "?frage=x&grenze=viele",
		"leer":       "?frage=x&grenze=",
		"gar keine":  "?frage=x",
		"negativ":    "?frage=x&grenze=-3",
		"eine Zahl":  "?frage=x&grenze=5",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, dienstMit(t, &zaehlender{}), endpunkt{
				name: "Indexsuche", methode: http.MethodGet,
				weg: "/index/suche" + abfrage,
			}, gueltig)

			if antwort.Code != http.StatusOK {
				t.Fatalf("%s gab %d", abfrage, antwort.Code)
			}

			if !strings.Contains(antwort.Body.String(), `"modus"`) {
				t.Errorf("die Antwort nennt den Modus nicht: %s",
					antwort.Body.String())
			}
		})
	}
}
