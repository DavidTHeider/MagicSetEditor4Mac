// **Die Argumente des Austauschs** — Werkstück O3 von Phase 5.2
// (#1427).
// Was die drei Werkzeuge über willige und böse Nähte tun, prüfen die Läufe
// über `alleNeuenFaelle`; hier steht, was keiner von ihnen sieht: Argumente,
// die kein JSON sind, und Argumente, die fehlen.
package mcp_test

import (
	"encoding/json"
	"testing"
)

func TestDerAustauschWeistArgumenteAbDieKeinJSONSind(t *testing.T) {
	t.Parallel()

	for _, name := range []string{
		"lesen.einstellungen.ausfuhr",
		"lesen.einstellungen.vorschau",
		"schreiben.einstellungen.einfuhr",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			werkzeug := werkzeugMitNeuenNaehten(t, neueStilprobe(), neueBestandsprobeWeit(), name)

			if _, err := werkzeug.Ruf(t.Context(), json.RawMessage(`[`)); err == nil {
				t.Error("Argumente, die kein JSON sind, kamen durch")
			}
		})
	}
}

// **Ohne Argumente gilt die Vorgabe**: die Ausfuhr ohne »abschnitte« gibt
// Nutzer und Profil — `argumenteLesen` lässt leere Argumente stehen.
func TestDieAusfuhrOhneArgumenteGibtDieVorgabe(t *testing.T) {
	t.Parallel()

	werkzeug := werkzeugMitNeuenNaehten(t, neueStilprobe(), neueBestandsprobeWeit(),
		"lesen.einstellungen.ausfuhr")

	aus, err := werkzeug.Ruf(t.Context(), nil)
	if err != nil || aus == "" {
		t.Fatalf("ohne Argumente: %q, %v", aus, err)
	}
}
