package domain_test

// **Wo eine KI-Laufzeit lauscht** — Werkstück A7 von Phase 5.3
// (#1446).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine brauchbare Adresse kommt durch, eine unbrauchbare nicht.**
func TestNurBrauchbareKIAdressenKommenDurch(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name    string
		wert    string
		erlaubt bool
	}{
		{"leer heißt die Vorgabe", "", true},
		{"nur Leerzeichen ebenso", "   ", true},
		{"die Schleifenadresse", "http://127.0.0.1:11434", true},
		{"mit TLS", "https://ki.example.org", true},
		{"ohne Schema", "127.0.0.1:11434", false},
		{"mit fremdem Schema", "ftp://127.0.0.1", false},
		{"ohne Rechner", "http://", false},
		{"gar keine Adresse", "http://%%", false},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			fehler := domain.PruefeKIAdresse(fall.wert)

			if fall.erlaubt && fehler != nil {
				t.Fatalf("%q wurde abgewiesen: %v", fall.wert, fehler)
			}

			if !fall.erlaubt && !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("%q gab %v", fall.wert, fehler)
			}
		})
	}
}

// **Die Einstufung ist einseitig.**
//
// # Warum das der Kern ist
//
// **Eine Adresse kann ein Nein beweisen und ein Ja nicht.** Nicht auf der
// Schleife heißt: dort rechnet jemand anderes. Auf der Schleife heißt nur:
// dieses Programm hat keinen Gegenbeweis — „Loopback allein beweist keine
// lokale Inferenz".
func TestDieEinstufungNachAdresseIstEinseitig(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name  string
		wert  string
		lokal bool
	}{
		{"leer ist die Vorgabe auf diesem Rechner", "", true},
		{"die Schleife in Zahlen", "http://127.0.0.1:11434", true},
		{"und irgendwo sonst in 127.0.0.0/8", "http://127.42.7.9", true},
		{"ihr Name", "http://localhost:11434", true},
		{"in Großbuchstaben", "http://LOCALHOST:11434", true},
		{"die Schleife in IPv6", "http://[::1]:11434", true},
		{"ein fremder Rechner", "http://ki.example.org:11434", false},
		{"eine fremde Zahl", "http://10.0.0.7:11434", false},

		// **Eine Schreibweise, die dieselbe Adresse meint und hier nicht
		// zählt** — und das ist der sichere Ausgang: unter »nur lokal« wird
		// dort nicht gefragt.
		{"die Schleife als Dezimalzahl", "http://2130706433", false},
		{"und als kurze Form", "http://127.1", false},
		{"127 als Name und nicht als Zahl", "http://127.0.0.x", false},
		{"was gar keine Adresse ist", "://", false},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			ort := domain.KIOrtFuerAdresse(fall.wert)

			erwartet := domain.KIOrtCloud
			if fall.lokal {
				erwartet = domain.KIOrtLokal
			}

			if ort != erwartet {
				t.Errorf("%q gilt als %q, erwartet %q", fall.wert, ort, erwartet)
			}
		})
	}
}
