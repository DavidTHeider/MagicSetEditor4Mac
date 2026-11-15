package anmeldung_test

// **Der Horcher auf der Schleifenadresse** — Werkstück 4 von Phase 4.12
// (#1351).

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/anmeldung"
)

// **Der Horcher nennt seine Adresse und nimmt einen Rückruf entgegen.**
//
// # Warum die Schleifenadresse und kein eigenes Schema
//
// **Nachgelesen am 2026-09-05**: Google verlangt für Desktop-Anwendungen
// „http://127.0.0.1:port oder http://[::1]:port" und sagt zu eigenen Schemata
// „Custom URI schemes are no longer supported due to the risk of app
// impersonation".
//
// **Der Port ist frei gewählt**, weil er es sein muss: „start an HTTP listener
// on a random available port". Deshalb kennt die Rückadresse erst der Horcher,
// und deshalb reicht ihn [anmeldung.Beginne] herein statt ihn zu wissen.
func TestDerHorcherNimmtEinenRueckrufEntgegen(t *testing.T) {
	t.Parallel()

	horcher, fehler := anmeldung.Horche()
	if fehler != nil {
		t.Fatalf("horchen: %v", fehler)
	}

	horcher.Erwarte("s")

	t.Cleanup(func() {
		if fehler := horcher.Schliesse(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	if !strings.HasPrefix(horcher.Adresse, "http://127.0.0.1:") {
		t.Errorf("die Adresse ist %q", horcher.Adresse)
	}

	go func() {
		antwort, ruffehler := http.Get(horcher.Adresse + "?code=c&state=s") //nolint:noctx // Testruf
		if ruffehler == nil {
			// **Der Rumpf wird geschlossen, sein Fehler nicht gemeldet**: der
			// Test hat, was er wollte, und ein `t.Error` aus einer Goroutine
			// nach dem Ende des Laufs bringt den Testlauf zum Absturz.
			_ = antwort.Body.Close() //nolint:errcheck // siehe oben
		}
	}()

	werte, wartefehler := horcher.Warte(t.Context())
	if wartefehler != nil {
		t.Fatalf("warten: %v", wartefehler)
	}

	if werte.Get("code") != "c" || werte.Get("state") != "s" {
		t.Errorf("Werte %v", werte)
	}
}

// **Ein Abbruch beendet das Warten.**
//
// **Sonst hinge das Fenster**, bis jemand im Browser fertig wird — und wer die
// Anmeldung abbricht, hat entschieden, dass er das nicht tut.
func TestEinAbbruchBeendetDasWarten(t *testing.T) {
	t.Parallel()

	horcher, fehler := anmeldung.Horche()
	if fehler != nil {
		t.Fatalf("horchen: %v", fehler)
	}

	horcher.Erwarte("s")

	t.Cleanup(func() {
		if fehler := horcher.Schliesse(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	ctx, brich := context.WithCancel(t.Context())
	brich()

	if _, wartefehler := horcher.Warte(ctx); !errors.Is(
		wartefehler, context.Canceled) {
		t.Errorf("Fehler %v, erwartet war context.Canceled", wartefehler)
	}
}

// **Und der Browser bekommt einen Satz, keinen leeren Bildschirm** (§6).
func TestDerBrowserBekommtEinenSatz(t *testing.T) {
	t.Parallel()

	horcher, fehler := anmeldung.Horche()
	if fehler != nil {
		t.Fatalf("horchen: %v", fehler)
	}

	horcher.Erwarte("s")

	t.Cleanup(func() {
		if fehler := horcher.Schliesse(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	fertig := make(chan string, 1)

	go func() {
		antwort, ruffehler := http.Get(horcher.Adresse + "?code=c&state=s") //nolint:noctx // Testruf
		if ruffehler != nil {
			fertig <- ""

			return
		}

		defer func() { _ = antwort.Body.Close() }() //nolint:errcheck // siehe oben

		roh, lesefehler := io.ReadAll(io.LimitReader(antwort.Body, 4096))
		if lesefehler != nil {
			fertig <- ""

			return
		}

		fertig <- string(roh)
	}()

	if _, wartefehler := horcher.Warte(t.Context()); wartefehler != nil {
		t.Fatalf("warten: %v", wartefehler)
	}

	select {
	case seite := <-fertig:
		if !strings.Contains(seite, "MagicSetEditor") {
			t.Errorf("die Seite sagt nichts: %q", seite)
		}

	case <-time.After(5 * time.Second):
		t.Error("der Browser bekam keine Antwort")
	}
}

// **Zweimal schließen ist kein Fehler.**
//
// **Weil der Aufrufer es doppelt tut** — einmal im `defer`, einmal nach einem
// geglückten Rückruf. Ein Fehler beim zweiten Mal wäre eine Meldung über
// etwas, das in Ordnung ist (§8).
func TestZweimalSchliessenIstKeinFehler(t *testing.T) {
	t.Parallel()

	horcher, fehler := anmeldung.Horche()
	if fehler != nil {
		t.Fatalf("horchen: %v", fehler)
	}

	horcher.Erwarte("s")

	if schliessfehler := horcher.Schliesse(); schliessfehler != nil {
		t.Fatalf("erstes Schließen: %v", schliessfehler)
	}

	if schliessfehler := horcher.Schliesse(); schliessfehler != nil {
		t.Errorf("zweites Schließen: %v", schliessfehler)
	}
}
