// **Die Abweisungen der Stil-, Paket-, Zähl-, Verlaufs- und
// Einstellungsendpunkte**
// (#1107).
//
// **Neben den vier Klassen und nicht statt ihnen.** Jene prüfen, dass der Weg
// bis zum Anwendungsfall trägt; hier stehen die Fälle, die nur diese Wege
// haben — eine unbrauchbare Kennung im Pfad und ein unlesbarer Rumpf.
package rest_test

import (
	"net/http"
	"testing"
)

// **Jeder Weg mit Kennung weist eine unbrauchbare ab.**
func TestDieNeuenBestandswegeWeisenKennungenAb(t *testing.T) {
	t.Parallel()

	faelle := []endpunkt{
		{name: "Stile", methode: http.MethodGet, weg: "/sets/x/stile"},
		{name: "Stilfolgen", methode: http.MethodGet, weg: "/sets/x/stile/p/folgen"},
		{
			name: "Stil wechseln", methode: http.MethodPut, weg: "/sets/x/stil",
			rumpf: `{"paket":"p"}`,
		},
		{name: "Stiloptionen", methode: http.MethodGet, weg: "/sets/x/stiloptionen"},
		{
			name: "Option setzen", methode: http.MethodPut, weg: "/sets/x/stiloptionen/rand",
			rumpf: `{"wert":"a"}`,
		},
		{
			name: "Option zurücksetzen", methode: http.MethodDelete,
			weg: "/sets/x/stiloptionen/rand",
		},
		{name: "Statistik", methode: http.MethodGet, weg: "/sets/x/statistik"},
		{name: "Verteilung", methode: http.MethodGet, weg: "/sets/x/statistik/name"},
		{name: "Zurücknehmen", methode: http.MethodPost, weg: "/sets/x/zuruecknehmen"},
		{name: "Wiederherstellen", methode: http.MethodPost, weg: "/sets/x/wiederherstellen"},
		// **Der Schnellfilter** — Werkstück U5 von Phase 5.1 (#1416).
		{name: "Karten filtern", methode: http.MethodGet, weg: "/sets/x/kartenfilter?anfrage=a"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t), fall, gueltig)
			if antwort.Code != http.StatusBadRequest {
				t.Errorf("endete mit %d statt 400: %s", antwort.Code, antwort.Body.String())
			}
		})
	}
}

// **Jeder Weg mit Rumpf weist einen unlesbaren ab.**
func TestDieNeuenBestandswegeWeisenRuempfeAb(t *testing.T) {
	t.Parallel()

	faelle := []endpunkt{
		{name: "Stil wechseln", methode: http.MethodPut, weg: "/sets/1/stil", rumpf: `{`},
		{
			name: "Option setzen", methode: http.MethodPut,
			weg: "/sets/1/stiloptionen/rand", rumpf: `{`,
		},
		{name: "Paket beziehen", methode: http.MethodPost, weg: "/pakete", rumpf: `{`},
		{
			name: "Einstellung setzen", methode: http.MethodPut,
			weg: "/einstellungen/sprache", rumpf: `{`,
		},
		// **Der Einstellungsaustausch** — Werkstück O3 von Phase 5.2 (#1427).
		{
			name: "Einfuhr vorschauen", methode: http.MethodPost,
			weg: "/einstellungen/vorschau", rumpf: `{`,
		},
		{
			name: "Einstellungen einführen", methode: http.MethodPost,
			weg: "/einstellungen/einfuhr", rumpf: `{`,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t), fall, gueltig)
			if antwort.Code != http.StatusBadRequest {
				t.Errorf("endete mit %d statt 400: %s", antwort.Code, antwort.Body.String())
			}
		})
	}
}

// **Wiederherstellen geht denselben Weg wie Zurücknehmen** — und wird trotzdem
// eigens gefahren.
//
// Die vier Klassen rufen jeden Weg einmal; dass **beide** Richtungen durch
// `bewege` kommen, sagt nur dieser Fall.
func TestBeideBewegungsrichtungenAntworten(t *testing.T) {
	t.Parallel()

	for name, weg := range map[string]string{
		"zurück": "/sets/1/zuruecknehmen",
		"vor":    "/sets/1/wiederherstellen",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t),
				endpunkt{methode: http.MethodPost, weg: weg}, gueltig)
			if antwort.Code != http.StatusOK {
				t.Errorf("endete mit %d: %s", antwort.Code, antwort.Body.String())
			}
		})
	}
}
