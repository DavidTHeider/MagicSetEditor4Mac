package netz_test

// **Die Erlaubnis vor jeder Anfrage** — Phase 4.8
// (#1321).
//
// **Sie sitzt hier und nicht beim Bauen des Klienten**, weil ein Schalter
// sich ändert: wer einen Dienst abschaltet, soll es sofort merken und nicht
// beim nächsten Start.

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

var errKeineErlaubnis = errors.New("nicht erlaubt")

// **Ohne Erlaubnis geht keine Anfrage hinaus.**
func TestOhneErlaubnisGehtNichtsHinaus(t *testing.T) {
	t.Parallel()

	var anfragen atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiber http.ResponseWriter, _ *http.Request) {
			anfragen.Add(1)
			schreiber.WriteHeader(http.StatusOK)
		}))
	defer server.Close()

	erlaubt := &atomic.Bool{}

	klient := netz.KlientUeber(server.Client()).
		MitErlaubnis(func(context.Context) error {
			if !erlaubt.Load() {
				return errKeineErlaubnis
			}

			return nil
		})

	_, err := netz.Hole(t.Context(), klient, nil,
		netz.Kenner{Name: "probe", Annahme: "*/*"}, server.URL, 1024)

	if !errors.Is(err, errKeineErlaubnis) {
		t.Fatalf("Fehler ist %v, erwartet »nicht erlaubt«", err)
	}

	if gezaehlt := anfragen.Load(); gezaehlt != 0 {
		t.Fatalf("der Server sah %d Anfragen, erwartet keine", gezaehlt)
	}

	// **Und mit Erlaubnis geht sie sofort** — ohne neuen Klienten.
	erlaubt.Store(true)

	if _, zweiter := netz.Hole(t.Context(), klient, nil,
		netz.Kenner{Name: "probe", Annahme: "*/*"}, server.URL, 1024); zweiter != nil {
		t.Fatalf("die zweite Anfrage: %v", zweiter)
	}

	if gezaehlt := anfragen.Load(); gezaehlt != 1 {
		t.Errorf("der Server sah %d Anfragen, erwartet genau eine", gezaehlt)
	}
}

// **Und ein Klient ohne Erlaubnis fragt niemanden.**
//
// `nil` heißt: es fragt niemand — dieselbe Bedeutung wie eine fehlende
// Drossel. Ohne diesen Lauf bliebe offen, ob ein Klient ohne Erlaubnis
// überhaupt noch etwas schickt.
func TestOhneGesetzteErlaubnisFragtNiemand(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiber http.ResponseWriter, _ *http.Request) {
			schreiber.WriteHeader(http.StatusOK)
		}))
	defer server.Close()

	if _, err := netz.Hole(t.Context(), netz.KlientUeber(server.Client()), nil,
		netz.Kenner{Name: "probe", Annahme: "*/*"}, server.URL, 1024); err != nil {
		t.Errorf("ohne Erlaubnis: %v", err)
	}
}

// **Auch ein Strom fragt vorher.**
//
// # Warum das ein eigener Lauf ist
//
// **Weil [netz.Strom] ein zweiter Weg zu demselben `Do` ist**, und der Lauf
// darüber prüft nur den ersten. Bis zum 2026-09-11 hat er nicht gefragt: die
// Zusage am Feld `erlaubnis` lautete „vor **jeder** Anfrage", eingelöst war
// sie in `holen`, und ein Strom ging hinaus, obwohl der Dienst aus war.
//
// Gefunden ist das nicht durch Lesen, sondern durch einen roten Lauf in
// `internal/app`: der Bulkbezug der Preise (A3, #1442) ist der erste Dienst
// **mit Schalter**, der strömt — `bezug` und `aktualisierung` haben je einen
// eigenen. Ein Wächter zählt seitdem nach, dass jeder Weg zum `Do` fragt
// ([erlaubnisvollstaendig_test.go]).
func TestAuchEinStromFragtVorher(t *testing.T) {
	t.Parallel()

	var anfragen atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiber http.ResponseWriter, _ *http.Request) {
			anfragen.Add(1)
			schreiber.WriteHeader(http.StatusOK)
		}))
	defer server.Close()

	erlaubt := &atomic.Bool{}

	klient := netz.KlientUeber(server.Client()).
		MitErlaubnis(func(context.Context) error {
			if !erlaubt.Load() {
				return errKeineErlaubnis
			}

			return nil
		})

	_, err := netz.Strom(t.Context(), klient, nil,
		netz.Kenner{Name: "probe", Annahme: "*/*"}, server.URL)

	if !errors.Is(err, errKeineErlaubnis) {
		t.Fatalf("Fehler ist %v, erwartet »nicht erlaubt«", err)
	}

	if gezaehlt := anfragen.Load(); gezaehlt != 0 {
		t.Fatalf("der Server sah %d Anfragen, erwartet keine", gezaehlt)
	}

	// **Und mit Erlaubnis strömt er sofort** — ohne neuen Klienten.
	erlaubt.Store(true)

	antwort, zweiter := netz.Strom(t.Context(), klient, nil,
		netz.Kenner{Name: "probe", Annahme: "*/*"}, server.URL)
	if zweiter != nil {
		t.Fatalf("die zweite Anfrage: %v", zweiter)
	}

	_ = antwort.Rumpf.Close() //nolint:errcheck // der Rumpf ist leer

	if gezaehlt := anfragen.Load(); gezaehlt != 1 {
		t.Errorf("der Server sah %d Anfragen, erwartet genau eine", gezaehlt)
	}
}
