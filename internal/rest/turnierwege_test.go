package rest_test

// **Die Abweisungen der Turnierwege** — Werkstück 5 von Phase 4.7
// (#1294).
//
// **Zwei Klassen, und beide sind Bedienfehler**: eine Kennung, die keine ist,
// und ein Rumpf, der kein JSON ist. Beide enden mit 400 — die Anfrage stimmt
// nicht, und der Bestand kommt gar nicht erst dran.

import (
	"net/http"
	"testing"
)

// **Jeder Weg weist eine unbrauchbare Kennung ab.**
func TestDieTurnierwegeWeisenKennungenAb(t *testing.T) {
	t.Parallel()

	faelle := []endpunkt{
		{name: "Turnier", methode: http.MethodGet, weg: "/turniere/x"},
		{name: "Turnier löschen", methode: http.MethodDelete, weg: "/turniere/x"},
		{name: "Rangliste", methode: http.MethodGet, weg: "/turniere/x/rangliste"},
		{name: "Vorschlag", methode: http.MethodGet, weg: "/turniere/x/vorschlag"},
		{name: "Runde", methode: http.MethodPost, weg: "/turniere/x/runden"},
		{name: "Abschluss", methode: http.MethodPost, weg: "/turniere/x/abschluss"},
		{name: "Wiedereröffnen", methode: http.MethodDelete, weg: "/turniere/x/abschluss"},
		{name: "Protokoll", methode: http.MethodGet, weg: "/turniere/x/protokoll"},
		{name: "Ausgabe", methode: http.MethodGet, weg: "/turniere/x/ausgabe"},
		{name: "Bericht", methode: http.MethodGet, weg: "/turniere/x/bericht"},
		{
			name: "Kopie", methode: http.MethodPost, weg: "/turniere/x/kopie",
			rumpf: `{"name":"Küchentisch II"}`,
		},
		{
			name: "Folgen, Paarung", methode: http.MethodPost,
			weg: "/turniere/1/paarungen/x/folgen", rumpf: `{"spieleA":2}`,
		},
		{
			name: "Folgen, Turnier", methode: http.MethodPost,
			weg: "/turniere/x/paarungen/1/folgen", rumpf: `{"spieleA":2}`,
		},
		{
			name: "Anmelden", methode: http.MethodPost, weg: "/turniere/x/teilnehmer",
			rumpf: `{"name":"Ada"}`,
		},
		{
			// **Beide Kennungen bekommen ihren Fall.** Der Weg liest zuerst
			// das Turnier und dann den Teilnehmer; ein Fall für nur eine
			// ließe den anderen Zweig unbetreten.
			name: "Aussteigen, Turnier", methode: http.MethodPut,
			weg: "/turniere/x/teilnehmer/1", rumpf: `{"ausgestiegen":true}`,
		},
		{
			name: "Aussteigen, Teilnehmer", methode: http.MethodPut,
			weg: "/turniere/1/teilnehmer/x", rumpf: `{"ausgestiegen":true}`,
		},
		{
			// **Die Paarung zuerst** — sie steht im Weg vor dem Turnier, und
			// beide Zweige brauchen ihren eigenen Fall.
			name: "Ergebnis, Paarung", methode: http.MethodPut,
			weg: "/turniere/1/paarungen/x", rumpf: `{"spieleA":2}`,
		},
		{
			name: "Ergebnis, Turnier", methode: http.MethodPut,
			weg: "/turniere/x/paarungen/1", rumpf: `{"spieleA":2}`,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t), fall, gueltig)
			if antwort.Code != http.StatusBadRequest {
				t.Errorf("endete mit %d statt 400: %s",
					antwort.Code, antwort.Body.String())
			}
		})
	}
}

// **Und jeder mit Rumpf weist einen unlesbaren ab.**
func TestDieTurnierwegeWeisenRuempfeAb(t *testing.T) {
	t.Parallel()

	faelle := []endpunkt{
		{name: "Anlegen", methode: http.MethodPost, weg: "/turniere", rumpf: `{`},
		{
			name: "Anmelden", methode: http.MethodPost,
			weg: "/turniere/1/teilnehmer", rumpf: `{`,
		},
		{
			name: "Aussteigen", methode: http.MethodPut,
			weg: "/turniere/1/teilnehmer/2", rumpf: `{`,
		},
		{
			name: "Ergebnis", methode: http.MethodPut,
			weg: "/turniere/1/paarungen/3", rumpf: `{`,
		},
		{
			name: "Kopie", methode: http.MethodPost,
			weg: "/turniere/1/kopie", rumpf: `{`,
		},
		{
			name: "Folgen", methode: http.MethodPost,
			weg: "/turniere/1/paarungen/3/folgen", rumpf: `{`,
		},
		{name: "Einlesen", methode: http.MethodPost, weg: "/turniere/einlesen", rumpf: `{`},
		{name: "Einlesen ohne Text", methode: http.MethodPost, weg: "/turniere/einlesen", rumpf: `{"inhalt":""}`},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t), fall, gueltig)
			if antwort.Code != http.StatusBadRequest {
				t.Errorf("endete mit %d statt 400: %s",
					antwort.Code, antwort.Body.String())
			}
		})
	}
}

// **Ein Ergebnis ohne ein einziges Feld nimmt es zurück.**
//
// **Und das ist etwas anderes als 0–0–0.** Ein Rumpf `{}` heißt „doch nicht
// gespielt"; drei Nullen heißen einen Spielstand, den es so nicht gibt.
func TestEinLeererErgebnisrumpfNimmtZurueck(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, neuerDienst(t), endpunkt{
		methode: http.MethodPut, weg: "/turniere/1/paarungen/3", rumpf: `{}`,
	}, gueltig)

	if antwort.Code != http.StatusNoContent {
		t.Errorf("endete mit %d statt 204: %s", antwort.Code, antwort.Body.String())
	}
}

// **Und der Weg nimmt beide Kennungen wirklich entgegen.**
//
// **Der Erfolgsfall zur Abweisung darüber**: ohne ihn bliebe die Zeile, die
// die Turnierkennung liest, unbetreten — genau die, die Befund 6 verlangt
// hat.
func TestDerAusstiegNimmtBeideKennungen(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, neuerDienst(t), endpunkt{
		methode: http.MethodPut, weg: "/turniere/1/teilnehmer/2",
		rumpf: `{"ausgestiegen":true}`,
	}, gueltig)

	if antwort.Code != http.StatusNoContent {
		t.Errorf("endete mit %d statt 204: %s", antwort.Code, antwort.Body.String())
	}
}
