package app_test

// **Aus heißt: kein einziger Aufruf geht hinaus** — Werkstück 2 von Phase 4.8
// (#1321),
// Abnahme 2, entschieden in
// [0063](../../docs/entscheidungen/0063-was-ein-dienst-hier-ist.md).
//
// **Gezählt wird an einem Testserver**, nicht an einer Vorrichtung: eine
// Zusage über Netzverkehr, die an einem Doppel geprüft wird, prüft das Doppel
// (§10, „zwei grüne Seiten einer Grenze sind kein Nachweis für die Grenze").

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// wahlprobe sagt, ob gefragt werden darf — und lässt sich umlegen.
type wahlprobe struct {
	an     atomic.Bool
	fehler error
}

func (w *wahlprobe) IstAn(_ context.Context, _ domain.Dienst) (bool, error) {
	return w.an.Load(), w.fehler
}

// **Ein abgeschalteter Dienst wird nicht gefragt.**
//
// **Gezählt wird am Server, nicht am Rückgabewert.** Ein Klient, der einen
// Fehler gibt und die Anfrage trotzdem abschickt, sähe von außen gleich aus.
func TestEinAbgeschalteterDienstWirdNichtGefragt(t *testing.T) {
	var anfragen atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiber http.ResponseWriter, _ *http.Request) {
			anfragen.Add(1)
			schreiber.WriteHeader(http.StatusOK)
		}))
	defer server.Close()

	t.Setenv("HOME", t.TempDir())

	// **Die Adresse wird über `MitScryfalladresse` umgelenkt, nicht über die
	// Umgebung.** `scryfalladresse` ist eine Paketveränderliche, die **einmal
	// beim Laden** aus der Umgebung liest; ein `t.Setenv` danach ändert
	// nichts — und der Lauf ginge an das **echte** Scryfall.
	//
	// Gefunden am 2026-09-01 an genau diesem Test: der Zähler blieb bei null,
	// obwohl der Dienst eingeschaltet war. Er blieb bei null, weil die
	// Anfrage woanders hinging (§3).
	vorher := app.MitScryfalladresse(server.URL)
	t.Cleanup(func() { app.MitScryfalladresse(vorher) })

	wahl := &wahlprobe{}

	klient, err := app.Scryfallklient(wahl)
	if err != nil {
		t.Fatalf("bauen: %v", err)
	}

	// **Der Klient entsteht auch bei abgeschaltetem Dienst.** Die Sperre
	// sitzt an der Anfrage, nicht am Bau — sonst hinge sie am Stand des
	// Starts, und ein Umlegen wirkte erst beim nächsten.
	_, _, _, suchfehler := klient.Suche(probekontext(t), "Lightning Bolt")

	if !domain.IstDienstAus(suchfehler) {
		t.Fatalf("Fehler ist %v, erwartet »der Dienst ist aus«", suchfehler)
	}

	if gezaehlt := anfragen.Load(); gezaehlt != 0 {
		t.Fatalf("der Server sah %d Anfragen, erwartet keine", gezaehlt)
	}

	// **Und das Umlegen wirkt sofort** — ohne Neustart, ohne neuen Klienten.
	//
	// Das ist die Zusage, die der erste Anlauf nicht halten konnte: er prüfte
	// beim Verdrahten, und der Schalter wurde einmal beim Start gelesen.
	wahl.an.Store(true)

	if _, _, _, zweiter := klient.Suche(
		probekontext(t), "Lightning Bolt"); domain.IstDienstAus(zweiter) {
		t.Errorf("nach dem Einschalten sagt er weiter »aus«: %v", zweiter)
	}

	if gezaehlt := anfragen.Load(); gezaehlt != 1 {
		t.Errorf("der Server sah %d Anfragen, erwartet genau eine", gezaehlt)
	}
}

// **Ein Fehler beim Lesen der Einstellung ist kein »an«.**
//
// Wer den Schalter nicht lesen kann, weiß nicht, ob er fragen darf — und
// fragt dann nicht. Die stille Umkehrung wäre die schlimmere: ein Dienst, der
// nach einem Lesefehler doch hinausgeht.
func TestEinUnlesbarerSchalterHeisstNicht(t *testing.T) {
	var anfragen atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(
		func(_ http.ResponseWriter, _ *http.Request) { anfragen.Add(1) }))
	defer server.Close()

	t.Setenv("HOME", t.TempDir())

	// **Die Adresse wird über `MitScryfalladresse` umgelenkt, nicht über die
	// Umgebung.** `scryfalladresse` ist eine Paketveränderliche, die **einmal
	// beim Laden** aus der Umgebung liest; ein `t.Setenv` danach ändert
	// nichts — und der Lauf ginge an das **echte** Scryfall.
	//
	// Gefunden am 2026-09-01 an genau diesem Test: der Zähler blieb bei null,
	// obwohl der Dienst eingeschaltet war. Er blieb bei null, weil die
	// Anfrage woanders hinging (§3).
	vorher := app.MitScryfalladresse(server.URL)
	t.Cleanup(func() { app.MitScryfalladresse(vorher) })

	wahl := &wahlprobe{fehler: domain.ErrNotFound}
	wahl.an.Store(true)

	klient, err := app.Scryfallklient(wahl)
	if err != nil {
		t.Fatalf("bauen: %v", err)
	}

	if _, _, _, suchfehler := klient.Suche(
		probekontext(t), "Lightning Bolt"); suchfehler == nil {
		t.Fatal("die Anfrage glückte, erwartet ein Fehler")
	}

	if gezaehlt := anfragen.Load(); gezaehlt != 0 {
		t.Errorf("der Server sah %d Anfragen, erwartet keine", gezaehlt)
	}
}

// **Der Start setzt die Dienstwahl** — sonst wäre die Sperre still
// abgeschaltet.
//
// [App.wahl] gibt ohne gesetzte Wahl `ImmerAn{}`, weil eine Anwendung ohne
// Bestand keinen Schalter hat, den sie lesen könnte. Der Preis dafür ist,
// dass ein **vergessener** Aufruf im Start genauso aussieht — und von außen
// wie eine funktionierende Sperre.
//
// **Dieser Lauf ist der Beleg, dass er nicht vergessen ist.**
func TestDerStartSetztDieDienstwahl(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		if err := anwendung.Beenden(); err != nil {
			t.Errorf("beenden: %v", err)
		}
	})

	if !anwendung.HatDienstwahl() {
		t.Error("nach dem Start ist keine Dienstwahl gesetzt")
	}
}
