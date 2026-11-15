// **Welche Frage ist welche** — der Schlüssel der Sammelanfrage.
//
// # Warum eine eigene Datei
//
// Weil es eine andere Frage ist als in `sammlung_test.go`: dort steht, ob die
// Anfrage aussieht wie verlangt und die Antwort richtig gedeutet wird — hier,
// ob **zwei verschiedene Fragen** auch zwei verschiedene Antworten bekommen.
// Der Schnitt folgt der Frage und nicht der Zeilenzahl (§2).
package scryfall_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// **Eine Kennung darf eine Druck-UUID sein** — und heißt dann `id`.
//
// # Warum das nachgetragen wurde
//
// [Klient.Sammlung] kannte bis zum 2026-09-01 drei Kennungsformen: Name, Set
// und Sammlernummer. Das genügte dem Zeilendeuter, der eine Deckliste liest —
// aber nicht der Formatprüfung: die hat den Druck schon als UUID vor sich und
// fragte deshalb je Platz **einzeln** nach. Befund 6 des Zweitblicks über
// Phase 4.6.
//
// **Nachgemessen am 2026-09-01**, nicht erinnert (§16, Regel 1): ein `POST
// /cards/collection` mit `{"identifiers":[{"id":"6a0b230b-…"}]}` antwortet mit
// `200` und einer Karte; eine UUID, die es nicht gibt, kommt unter `not_found`
// als `{"id":"…"}` zurück.
func TestEineKennungDarfEineDruckkennungSein(t *testing.T) {
	t.Parallel()

	var gesehen struct {
		Kennungen []map[string]string `json:"identifiers"`
	}

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			roh, lesefehler := io.ReadAll(anfrage.Body)
			if lesefehler != nil {
				t.Errorf("den Rumpf lesen: %v", lesefehler)
			}
			if err := json.Unmarshal(roh, &gesehen); err != nil {
				t.Errorf("der Rumpf ist kein JSON: %v", err)
			}

			schreibeAntwort(t, schreiben,
				`{"object":"list","data":[`+
					`{"object":"card","id":"6a0b230b","name":"Llanowar Elves"}],`+
					`"not_found":[{"id":"00000000"}]}`)
		}))
	t.Cleanup(gegenstelle.Close)

	klient := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

	antwort, _, _, fehler := klient.Sammlung(t.Context(), []scryfall.Kennung{
		{ID: "6a0b230b"}, {ID: "00000000"},
	})
	if fehler != nil {
		t.Fatalf("die Sammelanfrage: %v", fehler)
	}

	// **Das Feld heißt `id`.** Ein `ID` mit großem Anfangsbuchstaben käme bei
	// Scryfall als unbekanntes Feld an, und die Antwort wäre leer statt falsch
	// — was schlimmer ist, weil es wie ein leeres Deck aussieht.
	if len(gesehen.Kennungen) != 2 || gesehen.Kennungen[0]["id"] != "6a0b230b" {
		t.Errorf("gesendet wurde %v", gesehen.Kennungen)
	}

	// **Und die Kennung darf nichts anderes mitschicken.** `omitempty` steht
	// auf jedem Feld; ohne es stünde `"name":""` im Rumpf, und Scryfall
	// beantwortet eine Kennung mit zwei Feldern anders als eine mit einem.
	if _, hat := gesehen.Kennungen[0]["name"]; hat {
		t.Errorf("die Kennung trägt ein leeres Feld mit: %v", gesehen.Kennungen[0])
	}

	if len(antwort.NichtGefunden) != 1 || antwort.NichtGefunden[0].ID != "00000000" {
		t.Errorf("nicht gefunden: %v", antwort.NichtGefunden)
	}
}

// **Zwei Schübe verschiedener Drucke sind zwei Fragen** — und dürfen nicht
// dieselbe Antwort bekommen.
//
// # Der Fehler, den dieser Test fängt
//
// `sammelschluessel` baute den Namen aus Name, Setcode und Nummer. Für
// Kennungen, die **nur** eine UUID tragen, ist das dreimal die leere
// Zeichenkette — jeder Schub bekäme denselben Schlüssel, und der zweite
// bekäme die Karten des ersten aus dem Zwischenspeicher. Still, mit einem
// plausiblen Ergebnis: ein Deck voller falscher Karten.
func TestZweiSchuebeVerschiedenerDruckeKollidierenNicht(t *testing.T) {
	t.Parallel()

	var gefragt int

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			gefragt++

			roh, lesefehler := io.ReadAll(anfrage.Body)
			if lesefehler != nil {
				t.Errorf("den Rumpf lesen: %v", lesefehler)
			}
			name := "Erster"
			if strings.Contains(string(roh), "zweiter") {
				name = "Zweiter"
			}

			schreibeAntwort(t, schreiben,
				`{"object":"list","data":[{"object":"card","id":"x","name":"`+name+
					`"}],"not_found":[]}`)
		}))
	t.Cleanup(gegenstelle.Close)

	klient := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

	erste, _, _, fehler := klient.Sammlung(t.Context(), []scryfall.Kennung{{ID: "erster"}})
	if fehler != nil {
		t.Fatalf("der erste Schub: %v", fehler)
	}

	zweite, _, _, fehler := klient.Sammlung(t.Context(), []scryfall.Kennung{{ID: "zweiter"}})
	if fehler != nil {
		t.Fatalf("der zweite Schub: %v", fehler)
	}

	if erste.Karten[0].Name == zweite.Karten[0].Name {
		t.Errorf("beide Schübe gaben %q — der zweite kam aus dem Speicher des ersten",
			erste.Karten[0].Name)
	}

	if gefragt != 2 {
		t.Errorf("die Gegenstelle wurde %dmal gefragt, erwartet 2", gefragt)
	}
}

// **Und dieselbe Klasse eine Ebene tiefer**: die alte Zusammensetzung war
// **mehrdeutig**.
//
// Der Schlüssel entstand aus `Name|Setcode|Nummer` je Kennung, mit `\n`
// dazwischen. Zwei Kennungen `a` und `b` ergaben damit `"a||\nb||\n"` — und
// **eine** Kennung mit dem Namen `"a||\nb"` ergab dieselbe Zeichenkette.
//
// **Ein konstruierter Fall, und das steht hier so**: einen Kartennamen mit
// einem Zeilenumbruch darin gibt es bei Scryfall nicht. Er belegt nicht, dass
// jemand darüber stolpert, sondern dass die Zusammensetzung die Frage nicht
// eindeutig abbildet — und die UUID-Kollision darüber ist derselbe Fehler an
// einem Fall, den es wirklich gibt. Behoben wird die Klasse und nicht der
// Anlass (§10).
func TestEineMehrdeutigeZusammensetzungKollidiertNicht(t *testing.T) {
	t.Parallel()

	var gefragt int

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			gefragt++

			schreibeAntwort(t, schreiben, `{"object":"list","data":[],"not_found":[]}`)
		}))
	t.Cleanup(gegenstelle.Close)

	klient := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

	if _, _, _, fehler := klient.Sammlung(t.Context(),
		[]scryfall.Kennung{{Name: "a"}, {Name: "b"}}); fehler != nil {
		t.Fatalf("der erste Schub: %v", fehler)
	}

	if _, _, _, fehler := klient.Sammlung(t.Context(),
		[]scryfall.Kennung{{Name: "a||\nb"}}); fehler != nil {
		t.Fatalf("der zweite Schub: %v", fehler)
	}

	if gefragt != 2 {
		t.Errorf("die Gegenstelle wurde %dmal gefragt, erwartet 2 — "+
			"beide Fragen bekamen denselben Schlüssel", gefragt)
	}
}
