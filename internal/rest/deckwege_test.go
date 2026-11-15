// **Die Abweisungen der Deckwege** — Werkstück 7 von Phase 4.6
// (#1293).
//
// **Neben den vier Klassen und nicht statt ihnen.** Jene prüfen, dass der Weg
// bis zum Anwendungsfall trägt; hier stehen die Fälle, die nur diese Wege
// haben — eine unbrauchbare Kennung im Pfad und ein unlesbarer Rumpf.
package rest_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// **Jeder Deckweg mit Kennung weist eine unbrauchbare ab.**
func TestDieDeckwegeWeisenKennungenAb(t *testing.T) {
	t.Parallel()

	faelle := []endpunkt{
		{name: "Deck", methode: http.MethodGet, weg: "/decks/x"},
		{
			name: "Deck umbenennen", methode: http.MethodPut, weg: "/decks/x",
			rumpf: `{"fassung":1,"name":"A"}`,
		},
		{name: "Deck löschen", methode: http.MethodDelete, weg: "/decks/x"},
		{name: "Deck kopieren", methode: http.MethodPost, weg: "/decks/x/kopie", rumpf: `{"name":"A"}`},
		{name: "Fehlkarten wünschen", methode: http.MethodPost, weg: "/decks/x/fehlkarten/wunschliste"},
		{
			name: "Platz legen", methode: http.MethodPut, weg: "/decks/x/plaetze",
			rumpf: `{"menge":1,"bereich":"hauptdeck"}`,
		},
		{name: "Platz nehmen", methode: http.MethodDelete, weg: "/decks/x/plaetze/1"},
		{name: "Platz nehmen — Platz", methode: http.MethodDelete, weg: "/decks/1/plaetze/x"},
		{
			name: "Platz verschieben", methode: http.MethodPut,
			weg: "/decks/x/plaetze/1/bereich", rumpf: `{"bereich":"sideboard"}`,
		},
		{
			name: "Platz verschieben — Platz", methode: http.MethodPut,
			weg: "/decks/1/plaetze/x/bereich", rumpf: `{"bereich":"sideboard"}`,
		},
		{
			name: "Format wählen", methode: http.MethodPut, weg: "/decks/x/format",
			rumpf: `{"formatID":5}`,
		},
		{name: "Prüfung", methode: http.MethodGet, weg: "/decks/x/pruefung"},
		{name: "Bestand", methode: http.MethodGet, weg: "/decks/x/bestand"},

		// **Die Zuordnung** — Werkstück U9 von Phase 5.1 (#1420).
		{name: "Zuordnungen", methode: http.MethodGet, weg: "/decks/x/zuordnungen"},
		{name: "Fehlkarten", methode: http.MethodGet, weg: "/decks/x/fehlkarten"},
		{name: "Deckliste", methode: http.MethodGet, weg: "/decks/x/liste"},
		{name: "Decks eines Exemplars", methode: http.MethodGet, weg: "/sammlung/exemplare/x/decks"},
		{
			name: "Zuordnen", methode: http.MethodPut, weg: "/decks/x/zuordnungen",
			rumpf: `{"platzID":1,"exemplarID":1,"menge":1}`,
		},
		{
			name: "Zuordnen — kaputter Rumpf", methode: http.MethodPut,
			weg: "/decks/1/zuordnungen", rumpf: `{"platzID":`,
		},
		{name: "Zuordnung lösen", methode: http.MethodDelete, weg: "/decks/x/zuordnungen/1"},
		{
			name: "Zuordnung lösen — Zuordnung", methode: http.MethodDelete,
			weg: "/decks/1/zuordnungen/x",
		},
		{
			name: "Format setzen", methode: http.MethodPut, weg: "/formate/x",
			rumpf: `{"fassung":1}`,
		},
		{name: "Format löschen", methode: http.MethodDelete, weg: "/formate/x"},

		// **Die Decklisteneinfuhr** — Werkstück T8 von Phase 5.6 (#1473).
		{
			name: "Deckliste deuten", methode: http.MethodPost,
			weg: "/decks/x/liste/vorschau", rumpf: `{"text":"4 Bolt"}`,
		},
		{
			name: "Deckliste übernehmen", methode: http.MethodPost,
			weg: "/decks/x/liste", rumpf: `{"stand":"a","zeilen":[]}`,
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

// **Und jeder mit Rumpf weist einen unlesbaren ab.**
func TestDieDeckwegeWeisenRuempfeAb(t *testing.T) {
	t.Parallel()

	faelle := []endpunkt{
		{name: "Deck anlegen", methode: http.MethodPost, weg: "/decks", rumpf: `{`},
		{name: "Deck umbenennen", methode: http.MethodPut, weg: "/decks/1", rumpf: `{`},
		{name: "Deck kopieren", methode: http.MethodPost, weg: "/decks/1/kopie", rumpf: `{`},
		{name: "Deck aus einem Set", methode: http.MethodPost, weg: "/decks/aus-set", rumpf: `{`},
		{name: "Platz legen", methode: http.MethodPut, weg: "/decks/1/plaetze", rumpf: `{`},
		{
			name: "Platz verschieben", methode: http.MethodPut,
			weg: "/decks/1/plaetze/2/bereich", rumpf: `{`,
		},
		{name: "Format wählen", methode: http.MethodPut, weg: "/decks/1/format", rumpf: `{`},
		{name: "Format anlegen", methode: http.MethodPost, weg: "/formate", rumpf: `{`},
		{name: "Format setzen", methode: http.MethodPut, weg: "/formate/5", rumpf: `{`},

		// **Die Decklisteneinfuhr** — Werkstück T8 von Phase 5.6 (#1473).
		{
			name: "Deckliste deuten", methode: http.MethodPost,
			weg: "/decks/1/liste/vorschau", rumpf: `{`,
		},
		{
			name: "Deckliste übernehmen", methode: http.MethodPost,
			weg: "/decks/1/liste", rumpf: `{`,
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

// **Ein Format ohne Bauart geht den anderen Weg.**
//
// `POST /formate` hat **zwei** Zweige: mit `bauart` kommen die nachgelesenen
// Zahlen mit, ohne sie stehen die eigenen im Rumpf. Der Ablauf der vier
// Klassen fährt nur den ersten.
func TestEinFormatOhneBauartGehtDenAnderenWeg(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, neuerDienst(t), endpunkt{
		methode: http.MethodPost, weg: "/formate",
		rumpf: `{"name":"Eigen","spiel":"magic","mindestgroesse":40}`,
	}, gueltig)

	if antwort.Code != http.StatusCreated {
		t.Errorf("endete mit %d statt 201: %s", antwort.Code, antwort.Body.String())
	}
}

// **Und ein leerer Rumpf stürzt nicht ab.**
//
// **Der Fall, den `TestJederEndpunktReichtSeinenFehlerDurch` gefunden hat**:
// `json.Unmarshal` in eine `map[string]any` gibt für `null` eine **nil**-Map,
// und ein Schreibzugriff darauf stürzt ab.
func TestEinLeererRumpfStuerztNichtAb(t *testing.T) {
	t.Parallel()

	for _, fall := range []endpunkt{
		{name: "Format anlegen", methode: http.MethodPost, weg: "/formate", rumpf: `null`},
		{name: "Format setzen", methode: http.MethodPut, weg: "/formate/5", rumpf: `null`},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if antwort := ruf(t, neuerDienst(t), fall, gueltig); antwort.Code >= 500 {
				t.Errorf("endete mit %d: %s", antwort.Code, antwort.Body.String())
			}
		})
	}
}

// **Und ein Format mit Bauart geht den ersten.**
//
// **Beide Zweige von `POST /formate`** — ohne diesen Lauf stünde der eine da
// und niemand ginge ihn (§3).
func TestEinFormatMitBauartGehtDenErstenWeg(t *testing.T) {
	t.Parallel()

	for _, fall := range []endpunkt{
		{
			name:  "vollständig",
			rumpf: `{"bauart":"commander","name":"C","spiel":"magic","scryfallname":"commander"}`,
		},
		{
			// **Ohne Namen und Spiel** — `text` gibt dann die leere
			// Zeichenkette, und der Dienst weist ab statt zu raten.
			name: "nur die Bauart", rumpf: `{"bauart":"commander"}`,
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t), endpunkt{
				methode: http.MethodPost, weg: "/formate", rumpf: fall.rumpf,
			}, gueltig)

			if antwort.Code != http.StatusCreated {
				t.Errorf("endete mit %d statt 201: %s", antwort.Code, antwort.Body.String())
			}
		})
	}
}

// **Und ein Fehler auf dem Bauartweg kommt heraus.**
//
// **Ein eigener Lauf**, weil `TestJederEndpunktReichtSeinenFehlerDurch` den
// **anderen** Zweig von `POST /formate` fährt — den ohne Bauart. Ein
// Fehlerweg, den niemand geht, ist keiner (§3).
func TestEinFehlerAufDemBauartwegKommtHeraus(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.September, 1, 12, 0, 0, 0, time.UTC)

	dienst := rest.NeuerDienst(probemarke, func() time.Time { return jetzt },
		rest.Naehte{
			Decks: unwilliger{fehler: errors.New("die Naht schweigt")},
		})

	antwort := ruf(t, dienst, endpunkt{
		methode: http.MethodPost, weg: "/formate",
		rumpf: `{"bauart":"commander","name":"C","spiel":"magic"}`,
	}, gueltig)

	if antwort.Code < 400 {
		t.Errorf("endete mit %d: %s", antwort.Code, antwort.Body.String())
	}
}
