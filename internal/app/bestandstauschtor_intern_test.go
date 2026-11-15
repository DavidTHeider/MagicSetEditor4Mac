package app

// **Das Tor beim Tausch des Bestands, als Klasse** — A02 und A05
// (#2216,
// #2219).
//
// Der Weg über die Bindungen steht in `bestandstauschtor_test.go`. Hier
// stehen die Zustände, die von außen nicht herzustellen sind: **während**
// des Tauschs, ein echter Schluss mittendrin und ein Wiederöffnen, das
// scheitert.

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// torAntwort ruft eine gebundene Methode und gibt den Fehler zurück, mit dem
// das Tor abweist — oder nil, wenn es durchließ.
func torAntwort(t *testing.T, a *App) *Fehler {
	t.Helper()

	var abgewiesen *Fehler

	func() {
		defer func() {
			grund := recover()
			if grund == nil {
				return
			}

			var tor torfehler
			if fehler, ist := grund.(error); !ist || !errors.As(fehler, &tor) {
				panic(grund)
			}

			abgewiesen = tor.fehler
		}()

		NeueBindungen(a).Einstellungen(t.Context())
	}()

	return abgewiesen
}

// tauschbereit startet eine Anwendung über einer eigenen Datei.
func tauschbereit(t *testing.T) (a *App, bestand string) {
	t.Helper()

	bestand = filepath.Join(t.TempDir(), "bestand.db")
	a = NeueApp(bestand)

	if fehler := a.Starten(t.Context()); fehler != nil {
		t.Fatalf("starten: %v", fehler)
	}

	t.Cleanup(func() { _ = a.Beenden() }) //nolint:errcheck // nach dem Lauf ist nichts zu retten

	return a, bestand
}

func TestNachDemNeuanfangGehtDasTorWiederAuf(t *testing.T) {
	a, _ := tauschbereit(t)

	if _, fehler := a.neuanfangImTausch(t.Context()); fehler != nil {
		t.Fatalf("Neuanfang: %v", fehler)
	}

	if abgewiesen := torAntwort(t, a); abgewiesen != nil {
		t.Fatalf("nach dem Neuanfang weist das Tor ab: %+v", abgewiesen)
	}
}

// **Während des Tauschs sagt das Tor, was los ist** — und nicht „ein
// unerwarteter Fehler", wie es das Fenster bis A02 zeigte. **Und es bleibt
// zu, bis der ganze Tausch durch ist** — auch nach einem Wiederöffnen, denn
// danach kann noch eine Rücknahme kommen (Befund 3 des Zweitblicks über
// Phase A: bis dahin ging das Tor mit dem ersten Öffnen auf).
func TestWaehrendDesTauschsNenntDasTorDenTausch(t *testing.T) {
	a, _ := tauschbereit(t)

	fehler := a.imTausch(func() error {
		if fehler := a.zumTauschSchliessen(); fehler != nil {
			return fehler
		}

		abgewiesen := torAntwort(t, a)
		if abgewiesen == nil {
			t.Fatal("während des Tauschs ließ das Tor durch — auf einen geschlossenen Bestand")
		}

		if abgewiesen.Code != CodeBelegt || abgewiesen.Schluessel != "bestandWirdGetauscht" {
			t.Errorf("abgewiesen mit %+v, erwartet war belegt/bestandWirdGetauscht", abgewiesen)
		}

		if fehler := a.wiederOeffnen(t.Context()); fehler != nil {
			return fehler
		}

		if torAntwort(t, a) == nil {
			t.Error("nach dem Wiederöffnen, aber vor dem Ende des Tauschs, ließ das Tor durch")
		}

		return nil
	})
	if fehler != nil {
		t.Fatalf("Tausch: %v", fehler)
	}

	if abgewiesen := torAntwort(t, a); abgewiesen != nil {
		t.Fatalf("nach dem Tausch weist das Tor ab: %+v", abgewiesen)
	}
}

// **Zwei Tausche zugleich gibt es nicht** — Befund 2 des zweiten Durchgangs
// des Zweitblicks über Phase A: zwei Aufrufe passierten das offene Tor, der
// zweite ignorierte das gescheiterte Umstellen und räumte mit ab. **Der
// zweite wird abgewiesen, bevor er etwas anfasst.**
func TestEinZweiterTauschWirdAbgewiesenBevorErEtwasAnfasst(t *testing.T) {
	a, _ := tauschbereit(t)

	angefasst := false

	fehler := a.imTausch(func() error {
		return a.imTausch(func() error {
			angefasst = true

			return nil
		})
	})

	if !errors.Is(fehler, domain.ErrBelegt) {
		t.Errorf("der zweite Tausch gab %v, erwartet war belegt", fehler)
	}

	if angefasst {
		t.Error("der zweite Tausch fasste den Bestand an")
	}

	if abgewiesen := torAntwort(t, a); abgewiesen != nil {
		t.Errorf("nach beiden steht das Tor nicht offen: %+v", abgewiesen)
	}
}
