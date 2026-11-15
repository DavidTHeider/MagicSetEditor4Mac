// **Die Sammelanfrage** — `POST /cards/collection`
// (#1348).
//
// **Gegen eine eigene Gegenstelle**, nicht gegen Scryfall (§16): was hier
// läuft, geht an einen `httptest.Server` im selben Prozess.
package scryfall_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// **Eine Sammelanfrage stellt sie so, wie Scryfall sie verlangt.**
//
// Nachgelesen am 2026-08-31: POST, `Content-Type: application/json`, der Rumpf
// ist `{"identifiers": […]}`. Wer eines davon anders macht, bekommt keinen
// Fehler dieses Programms, sondern einen von Scryfall — und der stünde dann
// in der Oberfläche.
func TestEineSammelanfrageSiehtAusWieVerlangt(t *testing.T) {
	t.Parallel()

	var (
		methode    string
		inhaltstyp string
		gesehen    struct {
			Kennungen []scryfall.Kennung `json:"identifiers"`
		}
	)

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			methode = anfrage.Method
			inhaltstyp = anfrage.Header.Get("Content-Type")

			roh, lesefehler := io.ReadAll(anfrage.Body)
			if lesefehler != nil {
				t.Errorf("den Rumpf lesen: %v", lesefehler)
			}
			if err := json.Unmarshal(roh, &gesehen); err != nil {
				t.Errorf("der Rumpf ist kein JSON: %v", err)
			}

			schreibeAntwort(t, schreiben,
				`{"object":"list","data":[`+
					`{"object":"card","id":"a","name":"Lightning Bolt",`+
					`"set":"m11","collector_number":"149"}],"not_found":[]}`)
		}))
	t.Cleanup(gegenstelle.Close)

	klient := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

	antwort, _, _, fehler := klient.Sammlung(t.Context(), []scryfall.Kennung{
		{Name: "Lightning Bolt", Setcode: "m11", Nummer: "149"},
	})
	if fehler != nil {
		t.Fatalf("Sammlung: %v", fehler)
	}

	if methode != http.MethodPost {
		t.Errorf("Methode %q, erwartet war POST", methode)
	}

	if !strings.HasPrefix(inhaltstyp, "application/json") {
		t.Errorf("Content-Type %q", inhaltstyp)
	}

	if len(gesehen.Kennungen) != 1 ||
		gesehen.Kennungen[0].Name != "Lightning Bolt" ||
		gesehen.Kennungen[0].Setcode != "m11" ||
		gesehen.Kennungen[0].Nummer != "149" {
		t.Errorf("die Kennungen kamen als %+v an", gesehen.Kennungen)
	}

	if len(antwort.Karten) != 1 || antwort.Karten[0].Name != "Lightning Bolt" {
		t.Errorf("die Antwort lautet %+v", antwort.Karten)
	}
}

// **Was nicht gefunden wird, ist kein Fehler — es steht in `not_found`.**
//
// §8: eine Liste mit einem Tippfehler soll die übrigen Zeilen trotzdem
// bekommen. Und §16: nichts wird still verschluckt.
func TestNichtGefundenesStehtDaUndBrichtNichtAb(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreibeAntwort(t, schreiben,
				`{"object":"list","data":[`+
					`{"object":"card","id":"a","name":"Lightning Bolt"}],`+
					`"not_found":[{"name":"Gibtesnichtxyz"}]}`)
		}))
	t.Cleanup(gegenstelle.Close)

	klient := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

	antwort, _, _, fehler := klient.Sammlung(t.Context(), []scryfall.Kennung{
		{Name: "Lightning Bolt"}, {Name: "Gibtesnichtxyz"},
	})
	if fehler != nil {
		t.Fatalf("Sammlung: %v", fehler)
	}

	if len(antwort.Karten) != 1 {
		t.Errorf("%d Karten, erwartet war eine", len(antwort.Karten))
	}

	if len(antwort.NichtGefunden) != 1 ||
		antwort.NichtGefunden[0].Name != "Gibtesnichtxyz" {
		t.Errorf("nicht gefunden: %+v", antwort.NichtGefunden)
	}
}

// **Ohne `not_found` ist die Liste leer und nicht `nil`.**
//
// Der Aufrufer zählt sie; `nil` und `[]` sind für ihn nicht dasselbe.
func TestOhneNotFoundIstDieListeLeer(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreibeAntwort(t, schreiben, `{"object":"list","data":[]}`)
		}))
	t.Cleanup(gegenstelle.Close)

	klient := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

	antwort, _, _, fehler := klient.Sammlung(t.Context(),
		[]scryfall.Kennung{{Name: "Lightning Bolt"}})
	if fehler != nil {
		t.Fatalf("Sammlung: %v", fehler)
	}

	if antwort.NichtGefunden == nil {
		t.Error("die Liste ist nil")
	}
}

// **Mehr als die Grenze wird gar nicht erst geschickt.**
//
// „A maximum of 75 card references may be submitted per request." Wer mehr
// schickt, bekommt einen Fehler von Scryfall — und hat eine Anfrage
// verbraucht. Der Fehler entsteht hier, ohne Netz.
func TestUeberDerSammelgrenzeGehtNichtsHinaus(t *testing.T) {
	t.Parallel()

	gerufen := false

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			gerufen = true

			schreibeAntwort(t, schreiben, `{"object":"list","data":[]}`)
		}))
	t.Cleanup(gegenstelle.Close)

	klient := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

	zuviele := make([]scryfall.Kennung, scryfall.Sammelgrenze+1)
	for i := range zuviele {
		zuviele[i] = scryfall.Kennung{Name: "Lightning Bolt"}
	}

	_, _, _, fehler := klient.Sammlung(t.Context(), zuviele)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("Fehler %v, erwartet war ErrValidation", fehler)
	}

	if gerufen {
		t.Error("die Anfrage ging trotzdem hinaus")
	}
}

// **Und ohne Kennungen wird auch nichts gefragt.**
func TestOhneKennungenGehtNichtsHinaus(t *testing.T) {
	t.Parallel()

	gerufen := false

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			gerufen = true

			schreibeAntwort(t, schreiben, `{"object":"list","data":[]}`)
		}))
	t.Cleanup(gegenstelle.Close)

	klient := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

	antwort, _, _, fehler := klient.Sammlung(t.Context(), nil)
	if fehler != nil {
		t.Fatalf("Sammlung: %v", fehler)
	}

	if gerufen {
		t.Error("eine leere Sammelanfrage ging hinaus")
	}

	if len(antwort.Karten) != 0 || antwort.NichtGefunden == nil {
		t.Errorf("die Antwort lautet %+v", antwort)
	}
}

// **Zwei verschiedene Sammelanfragen bekommen nicht dieselbe Antwort.**
//
// # Der Fehler, gegen den dieser Lauf steht
//
// Die Adresse ist für **jede** Sammelanfrage dieselbe — die Frage steht im
// Rumpf. Ein Zwischenspeicher, der über die Adresse ginge, gäbe der zweiten
// Anfrage die Antwort der ersten, und niemand sähe es: beide sind gültige
// Listen.
func TestZweiSammelanfragenLiegenNichtUnterDemselbenNamen(t *testing.T) {
	t.Parallel()

	rufe := 0

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			rufe++

			roh, lesefehler := io.ReadAll(anfrage.Body)
			if lesefehler != nil {
				t.Errorf("den Rumpf lesen: %v", lesefehler)
			}

			name := "Lightning Bolt"
			if strings.Contains(string(roh), "Counterspell") {
				name = "Counterspell"
			}

			schreibeAntwort(t, schreiben,
				`{"object":"list","data":[{"object":"card","id":"a","name":"`+
					name+`"}],"not_found":[]}`)
		}))
	t.Cleanup(gegenstelle.Close)

	klient := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

	erste, _, _, fehler := klient.Sammlung(t.Context(),
		[]scryfall.Kennung{{Name: "Lightning Bolt"}})
	if fehler != nil {
		t.Fatalf("erste Sammlung: %v", fehler)
	}

	zweite, _, _, fehler := klient.Sammlung(t.Context(),
		[]scryfall.Kennung{{Name: "Counterspell"}})
	if fehler != nil {
		t.Fatalf("zweite Sammlung: %v", fehler)
	}

	if erste.Karten[0].Name == zweite.Karten[0].Name {
		t.Errorf("beide gaben %q — die zweite bekam die Antwort der ersten",
			erste.Karten[0].Name)
	}

	// **Und dieselbe Frage wird nur einmal gestellt.** Die Gegenprobe zur
	// Zeile darüber: ein Schlüssel, der jede Anfrage unterscheidet, dürfte
	// nicht so weit gehen, dass er auch dieselbe unterscheidet.
	if _, _, _, fehler := klient.Sammlung(t.Context(),
		[]scryfall.Kennung{{Name: "Lightning Bolt"}}); fehler != nil {
		t.Fatalf("dritte Sammlung: %v", fehler)
	}

	if rufe != 2 {
		t.Errorf("%d Anfragen, erwartet waren zwei", rufe)
	}
}

// **Eine Antwort ohne `object` ist keine Liste.**
//
// §16: Pflichtfelder streng. Wer sie hinnähme, verarbeitete eine Fehlerantwort
// von Scryfall als leeres Ergebnis.
func TestEineAntwortOhneObjektWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"kein object": `{"data":[]}`,
		"kein JSON":   `dies ist keine Antwort`,
	}

	for name, rumpf := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gegenstelle := httptest.NewServer(http.HandlerFunc(
				func(schreiben http.ResponseWriter, _ *http.Request) {
					schreibeAntwort(t, schreiben, rumpf)
				}))
			t.Cleanup(gegenstelle.Close)

			klient := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

			if _, _, _, fehler := klient.Sammlung(t.Context(),
				[]scryfall.Kennung{{Name: "Lightning Bolt"}}); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Errorf("Fehler %v, erwartet war ErrValidation", fehler)
			}
		})
	}
}

// **Und ein Status, der nicht 200 ist, ebenso.**
func TestEinSchlechterStatusWirdGemeldet(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(http.StatusBadRequest)
		}))
	t.Cleanup(gegenstelle.Close)

	klient := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

	_, _, _, fehler := klient.Sammlung(t.Context(),
		[]scryfall.Kennung{{Name: "Lightning Bolt"}})
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("Fehler %v, erwartet war ErrValidation", fehler)
	}

	if !strings.Contains(fehler.Error(), "400") {
		t.Errorf("die Meldung nennt den Status nicht: %v", fehler)
	}
}

// schreibeAntwort schreibt einen Rumpf — und meldet, wenn es nicht geht.
//
// **Ein `_, _ =` verschwiege einen Schreibfehler**, und der Lauf endete dann
// mit „die Antwort ist kein JSON" statt mit dem, was wirklich war.
func schreibeAntwort(t *testing.T, schreiben http.ResponseWriter, rumpf string) {
	t.Helper()

	if _, fehler := schreiben.Write([]byte(rumpf)); fehler != nil {
		t.Errorf("die Antwort schreiben: %v", fehler)
	}
}

// **Eine Gegenstelle, die gar nicht antwortet, kommt als Fehler durch.**
//
// Der Weg durch `netz.SendeJSON` ohne Antwort — nicht derselbe wie ein Status,
// den niemand mag: hier gibt es keinen.
func TestEineTauteGegenstelleKommtAlsFehlerDurch(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(http.ResponseWriter, *http.Request) {}))

	klient := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

	// **Erst danach schließen** — der Klient soll ins Leere greifen.
	gegenstelle.Close()

	if _, _, _, fehler := klient.Sammlung(t.Context(),
		[]scryfall.Kennung{{Name: "Lightning Bolt"}}); fehler == nil {
		t.Error("die tote Gegenstelle kam als Erfolg zurück")
	}
}

// **Eine Karte, die keine ist, bricht die Antwort ab.**
//
// §16: Pflichtfelder streng. Ein `data`-Eintrag ohne `object` ist keine Karte,
// und ihn als eine zu zählen hieße, eine halbe Antwort auszugeben.
func TestEineUnlesbareKarteInDerSammelantwort(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreibeAntwort(t, schreiben,
				`{"object":"list","data":[{"object":"nichtkarte"}],"not_found":[]}`)
		}))
	t.Cleanup(gegenstelle.Close)

	klient := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

	if _, _, _, fehler := klient.Sammlung(t.Context(),
		[]scryfall.Kennung{{Name: "Lightning Bolt"}}); fehler == nil {
		t.Error("die Nichtkarte kam als Erfolg zurück")
	}
}
