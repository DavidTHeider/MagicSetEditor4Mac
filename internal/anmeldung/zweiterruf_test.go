package anmeldung_test

// **Die zweiten Rufe und die unbrauchbaren Adressen** — abgetrennt von
// `tauschfehler_test.go`, als auch die über die Zeilengrenze aus §2 kam.
//
// **Nach der Frage geschnitten**: hier steht, was der **zweite Ruf** hergibt —
// GitHubs flache Gestalt, Xs Gestalt in `data`, und was geschieht, wenn eine
// Adresse gar keine ist.

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/anmeldung"
)

// **X packt seine Antwort in `data` und führt `id` als Zeichenkette.**
//
// **Zwei Gestalten für denselben zweiten Ruf** — GitHub antwortet flach mit
// `id` als **Zahl**. Wer nur eine liest, hat einen Anbieter, der still eine
// leere Kennung gibt.
func TestDieXGestaltDesZweitenRufsWirdGelesen(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			schreiben.Header().Set("Content-Type", "application/json")

			if strings.HasSuffix(anfrage.URL.Path, "/token") {
				antworte(t, schreiben, `{"access_token":"abc"}`)

				return
			}

			antworte(t, schreiben,
				`{"data":{"id":"1234","name":"Wer","username":"wer"}}`)
		}))
	t.Cleanup(gegenstelle.Close)

	anbieter := anmeldung.X("kennung", "")
	anbieter.Tauschadresse = gegenstelle.URL + "/token"
	anbieter.Nutzeradresse = gegenstelle.URL + "/me"

	lauf := mussBeginnen(t, anbieter)

	kennung, fehler := lauf.Einloesen(t.Context(), anmeldung.RuferUeber(gegenstelle.Client()), "c")
	if fehler != nil {
		t.Fatalf("einlösen: %v", fehler)
	}

	if kennung.Subjekt != "1234" || kennung.Anzeigename != "Wer" {
		t.Errorf("Identität %+v", kennung)
	}
}

// **Ein Anbieter ohne Namen gibt einen leeren Anzeigenamen.**
//
// **Und das ist keine Lücke**: der Anzeigename ist Zierrat, nichts hängt daran
// (siehe [domain.Anbieterkonto]). Wer stattdessen abwiese, sperrte jemanden
// aus, weil er sein Profil beim Anbieter nicht ausgefüllt hat.
func TestEinAnbieterOhneNamenGibtEinenLeerenAnzeigenamen(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			schreiben.Header().Set("Content-Type", "application/json")

			if strings.HasSuffix(anfrage.URL.Path, "/token") {
				antworte(t, schreiben, `{"access_token":"abc"}`)

				return
			}

			antworte(t, schreiben, `{"id":9}`)
		}))
	t.Cleanup(gegenstelle.Close)

	anbieter := anmeldung.GitHub("kennung", "geheim")
	anbieter.Tauschadresse = gegenstelle.URL + "/token"
	anbieter.Nutzeradresse = gegenstelle.URL + "/user"

	lauf := mussBeginnen(t, anbieter)

	kennung, fehler := lauf.Einloesen(t.Context(), anmeldung.RuferUeber(gegenstelle.Client()), "c")
	if fehler != nil {
		t.Fatalf("einlösen: %v", fehler)
	}

	if kennung.Anzeigename != "" {
		t.Errorf("Anzeigename %q, erwartet war leer", kennung.Anzeigename)
	}
}

// **Eine unerreichbare Gegenstelle meldet sich an beiden Rufen.**
func TestEineUnerreichbareGegenstelleMeldetSich(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			antworte(t, schreiben, `{"access_token":"abc"}`)
		}))

	anbieter := anmeldung.GitHub("kennung", "geheim")
	anbieter.Tauschadresse = gegenstelle.URL + "/token"
	anbieter.Nutzeradresse = gegenstelle.URL + "/user"

	lauf := mussBeginnen(t, anbieter)

	// **Erst den zweiten Ruf**: solange die Gegenstelle steht, glückt der
	// Tausch, und dann kommt der zweite an eine Adresse, die es nicht gibt.
	zu := anbieter
	zu.Nutzeradresse = "http://127.0.0.1:1/gibtsnicht"

	zulauf := mussBeginnen(t, zu)

	if _, fehler := zulauf.Einloesen(
		t.Context(), anmeldung.RuferUeber(gegenstelle.Client()), "c"); fehler == nil {
		t.Error("der zweite Ruf ins Leere kam durch")
	}

	// **Dann den Tausch selbst.**
	gegenstelle.Close()

	if _, fehler := lauf.Einloesen(
		t.Context(), anmeldung.RuferUeber(gegenstelle.Client()), "c"); fehler == nil {
		t.Error("der Tausch gegen eine tote Gegenstelle kam durch")
	}
}

// **Eine unbrauchbare Tauschadresse wird beim Bauen abgewiesen.**
func TestEineUnbrauchbareTauschadresseWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	anbieter := anmeldung.GitHub("kennung", "geheim")
	anbieter.Tauschadresse = "://kein Schema"

	lauf := mussBeginnen(t, anbieter)

	if _, fehler := lauf.Einloesen(
		t.Context(), anmeldung.NeuerRufer(), "c"); fehler == nil {
		t.Error("die unbrauchbare Adresse kam durch")
	}
}

// **Und eine unbrauchbare Nutzeradresse ebenso.**
func TestEineUnbrauchbareNutzeradresseWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			antworte(t, schreiben, `{"access_token":"abc"}`)
		}))
	t.Cleanup(gegenstelle.Close)

	anbieter := anmeldung.GitHub("kennung", "geheim")
	anbieter.Tauschadresse = gegenstelle.URL + "/token"
	anbieter.Nutzeradresse = "://kein Schema"

	lauf := mussBeginnen(t, anbieter)

	if _, fehler := lauf.Einloesen(
		t.Context(), anmeldung.RuferUeber(gegenstelle.Client()), "c"); fehler == nil {
		t.Error("die unbrauchbare Adresse kam durch")
	}
}

// **Eine abgeschnittene Antwort meldet sich — an beiden Rufen.**
//
// # Warum eine gekaperte Verbindung
//
// **Weil `io.ReadAll` nur scheitert, wenn der Strom mitten im Satz abreißt.**
// Ein Server, der ordentlich antwortet, gibt einen vollständigen Rumpf; eine
// Antwort mit angekündigter Länge und weniger Inhalt gibt es nur, wenn man die
// Verbindung selbst schreibt.
//
// **Und der Fall kommt vor**: ein Netz, das mitten in der Antwort ausfällt,
// sieht genau so aus.
func TestEineAbgeschnitteneAntwortMeldetSich(t *testing.T) {
	t.Parallel()

	abgeschnitten := func(schreiben http.ResponseWriter) {
		kaperer, kann := schreiben.(http.Hijacker)
		if !kann {
			t.Error("dieser Testserver lässt sich nicht kapern")

			return
		}

		verbindung, _, kaperfehler := kaperer.Hijack()
		if kaperfehler != nil {
			t.Errorf("kapern: %v", kaperfehler)

			return
		}

		// **Mehr angekündigt als geschrieben, dann zu.** Genau so sieht ein
		// Netz aus, das mitten in der Antwort ausfällt — und nur so scheitert
		// `io.ReadAll`.
		if _, schreibfehler := verbindung.Write([]byte(
			"HTTP/1.1 200 OK\r\nContent-Length: 100\r\n\r\n{\"access_",
		)); schreibfehler != nil {
			t.Errorf("die halbe Antwort schreiben: %v", schreibfehler)
		}

		if schliessfehler := verbindung.Close(); schliessfehler != nil {
			t.Errorf("die Verbindung schließen: %v", schliessfehler)
		}
	}

	t.Run("beim Tausch", func(t *testing.T) {
		t.Parallel()

		gegenstelle := httptest.NewServer(http.HandlerFunc(
			func(schreiben http.ResponseWriter, _ *http.Request) {
				abgeschnitten(schreiben)
			}))
		t.Cleanup(gegenstelle.Close)

		anbieter := anmeldung.GitHub("kennung", "geheim")
		anbieter.Tauschadresse = gegenstelle.URL + "/token"

		lauf := mussBeginnen(t, anbieter)

		if _, fehler := lauf.Einloesen(
			t.Context(), anmeldung.RuferUeber(gegenstelle.Client()), "c"); fehler == nil {
			t.Error("die abgeschnittene Antwort kam durch")
		}
	})

	t.Run("beim zweiten Ruf", func(t *testing.T) {
		t.Parallel()

		gegenstelle := httptest.NewServer(http.HandlerFunc(
			func(schreiben http.ResponseWriter, anfrage *http.Request) {
				if strings.HasSuffix(anfrage.URL.Path, "/token") {
					antworte(t, schreiben, `{"access_token":"abc"}`)

					return
				}

				abgeschnitten(schreiben)
			}))
		t.Cleanup(gegenstelle.Close)

		anbieter := anmeldung.GitHub("kennung", "geheim")
		anbieter.Tauschadresse = gegenstelle.URL + "/token"
		anbieter.Nutzeradresse = gegenstelle.URL + "/user"

		lauf := mussBeginnen(t, anbieter)

		if _, fehler := lauf.Einloesen(
			t.Context(), anmeldung.RuferUeber(gegenstelle.Client()), "c"); fehler == nil {
			t.Error("die abgeschnittene Antwort kam durch")
		}
	})
}

// **Ein Fehlerstatus ohne `error` im Rumpf zählt trotzdem.**
//
// **Beides und nicht eines**: GitHub antwortet auf einen verbrauchten Code mit
// **200** und `error` im Rumpf — dafür steht die andere Prüfung. Ein `500` mit
// leerem Rumpf hat kein `error`, und ohne diese Prüfung ginge er als geglückter
// Tausch ohne Zugangswort durch.
func TestEinFehlerstatusOhneFehlerImRumpfZaehlt(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.Header().Set("Content-Type", "application/json")
			schreiben.WriteHeader(http.StatusInternalServerError)
			antworte(t, schreiben, `{}`)
		}))
	t.Cleanup(gegenstelle.Close)

	anbieter := anmeldung.GitHub("kennung", "geheim")
	anbieter.Tauschadresse = gegenstelle.URL + "/token"

	lauf := mussBeginnen(t, anbieter)

	_, fehler := lauf.Einloesen(t.Context(), anmeldung.RuferUeber(gegenstelle.Client()), "c")
	if fehler == nil {
		t.Fatal("der Fehlerstatus kam durch")
	}

	if !strings.Contains(fehler.Error(), "500") {
		t.Errorf("die Meldung nennt den Status nicht: %v", fehler)
	}
}
