package anmeldung

// **Die Fehlerwege des Horchers** — Werkstück 4 von Phase 4.12
// (#1351).
//
// **Paketintern**, weil die beiden Nähte hier liegen: ein Horchen, das
// scheitert, und ein Schließen, das klagt, kommen auf einem gesunden Rechner
// nicht zustande — im Betrieb schon.

import (
	"errors"
	"net"
	"net/http"
	"testing"
)

// **Ein Horchen, das scheitert, wird gemeldet.**
func TestEinGescheitertesHorchenWirdGemeldet(t *testing.T) {
	vorher := lauschen

	t.Cleanup(func() { lauschen = vorher })

	lauschen = func(string, string) (net.Listener, error) {
		return nil, errors.New("kein Port frei")
	}

	if _, fehler := Horche(); fehler == nil {
		t.Error("das gescheiterte Horchen kam durch")
	}
}

// **Und ein Schließen, das klagt, ebenso.**
//
// **`http.ErrServerClosed` ist dabei kein Fehler**: er heißt, dass schon
// zugemacht war — eine Meldung darüber wäre eine über etwas, das in Ordnung
// ist (§8).
func TestEinKlagendesSchliessenWirdGemeldet(t *testing.T) {
	horcher, fehler := Horche()
	if fehler != nil {
		t.Fatalf("horchen: %v", fehler)
	}

	// **Der Server wird zugemacht, sein Fehler nicht gemeldet**: die Naht
	// darüber ist gerade absichtlich verbogen, und was `Close` dann sagt, ist
	// keine Aussage über das Geprüfte.
	t.Cleanup(func() { _ = horcher.server.Close() }) //nolint:errcheck // siehe oben

	horcher.schliessen = func() error { return errors.New("geht nicht zu") }

	if schliessfehler := horcher.Schliesse(); schliessfehler == nil {
		t.Error("der Fehler ging verloren")
	}
}

// **Ein bereits geschlossener Server ist keiner.**
func TestEinBereitsGeschlossenerServerIstKeinFehler(t *testing.T) {
	horcher, fehler := Horche()
	if fehler != nil {
		t.Fatalf("horchen: %v", fehler)
	}

	// **Der Server wird zugemacht, sein Fehler nicht gemeldet**: die Naht
	// darüber ist gerade absichtlich verbogen, und was `Close` dann sagt, ist
	// keine Aussage über das Geprüfte.
	t.Cleanup(func() { _ = horcher.server.Close() }) //nolint:errcheck // siehe oben

	horcher.schliessen = func() error { return http.ErrServerClosed }

	if schliessfehler := horcher.Schliesse(); schliessfehler != nil {
		t.Errorf("ErrServerClosed kam als Fehler an: %v", schliessfehler)
	}
}
