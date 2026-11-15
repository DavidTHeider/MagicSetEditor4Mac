package domain_test

// **Die Prüfungen der Netzeinstellungen** — Werkstück N5 von Phase 5.4
// (#1571).
//
// # Wonach hier gefragt wird
//
// Danach, dass ein Proxy **ohne Schema** abgewiesen wird — `httpproxy` läse
// `proxy:3128` als **Rechnernamen mit Port**, und wer `http://` vergisst,
// bekäme stillschweigend etwas anderes, als er meinte —, und danach, dass
// eine Frist von null nicht durchkommt: in Go heißt sie „gar keine", und eine
// Einstellung, die eine Sicherung abschaltet, tut es dann unsichtbar.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineProxyadresseWirdGeprueft(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		wert     string
		abweisen bool
	}{
		"leer ist die Umgebung": {"", false},
		"nur Leerzeichen":       {"   ", false},
		"mit Schema":            {"http://proxy.firma:3128", false},
		"https als Schema":      {"https://proxy.firma:3128", false},
		"ohne Schema":           {"proxy.firma:3128", true},
		"fremdes Schema":        {"socks5://proxy.firma:1080", true},
		"ohne Rechner":          {"http://", true},
		"keine Adresse":         {"http://%zz", true},

		// **Ein Kennwort gehört nicht in die Einstellungen** — S13.NAT-02
		// (#1706): es läge dort unverschlüsselt im Bestand und stünde in
		// jeder Anzeige. Ein Name allein ist kein Geheimnis.
		"mit Kennwort":        {"http://anna:geheim@proxy.firma:3128", true},
		"mit leerem Kennwort": {"http://anna:@proxy.firma:3128", true},
		"nur mit Namen":       {"http://anna@proxy.firma:3128", false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := domain.PruefeProxyadresse(fall.wert)

			if fall.abweisen && fehler == nil {
				t.Errorf("%q kam durch", fall.wert)
			}

			if !fall.abweisen && fehler != nil {
				t.Errorf("%q wurde abgewiesen: %v", fall.wert, fehler)
			}

			if fehler != nil && !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("der Fehler ist keine Validierung: %v", fehler)
			}
		})
	}
}

func TestEineNetzfristWirdGeprueft(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		wert     string
		abweisen bool
	}{
		"leer ist die Vorgabe": {"", false},
		"eine Zahl":            {"30", false},
		"die Obergrenze":       {"3600", false},
		"null ist keine Frist": {"0", true},
		"negativ":              {"-5", true},
		"über der Grenze":      {"3601", true},
		"keine Zahl":           {"dreißig", true},
		// **Das Produkt lief um** (J05, #2269): 9.223.372.037 Sekunden mal
		// `time.Second` ist negativ und lag damit unter der Obergrenze.
		"über die Dauer hinaus": {"9223372037", true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := domain.PruefeNetzfrist(fall.wert)

			if fall.abweisen && fehler == nil {
				t.Errorf("%q kam durch", fall.wert)
			}

			if !fall.abweisen && fehler != nil {
				t.Errorf("%q wurde abgewiesen: %v", fall.wert, fehler)
			}
		})
	}
}

// **Ein gedeuteter Wert, und die Vorgabe, wo keiner steht.**
//
// **Hier und nicht in `app`**, damit jede Fläche dieselbe Zahl bekommt — zwei
// Deutungen desselben Werts wären zwei Antworten auf eine Frage (§16).
func TestEineNetzfristWirdGedeutet(t *testing.T) {
	t.Parallel()

	vorgabe := 42 * time.Second

	for name, fall := range map[string]struct {
		wert     string
		erwartet time.Duration
	}{
		"eine Zahl":       {"30", 30 * time.Second},
		"mit Leerzeichen": {" 30 ", 30 * time.Second},
		"leer":            {"", vorgabe},
		"keine Zahl":      {"dreißig", vorgabe},
		"null":            {"0", vorgabe},
		"negativ":         {"-1", vorgabe},
		// Ein Wert, der vor der Prüfung gespeichert wurde, wird nicht negativ.
		"über die Dauer hinaus": {"9223372037", vorgabe},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if gab := domain.Netzfrist(fall.wert, vorgabe); gab != fall.erwartet {
				t.Errorf("%q gab %v, erwartet %v", fall.wert, gab, fall.erwartet)
			}
		})
	}
}

// **Ein Kennwort in einer Proxyadresse steht in keiner Anzeige** — S13.NAT-02
// (#1706):
// „keine Geheimnisse in Standardanzeige oder Export".
//
// Aus den Einstellungen kommt keines mehr; aus der **Umgebung** schon —
// `HTTPS_PROXY` nimmt jeder, und Go schickt das Kennwort als
// `Proxy-Authorization` mit. Gezeigt wird die Adresse trotzdem, denn wer
// sich wundert, woher ein Proxy kommt, braucht sie.
func TestEinProxykennwortWirdVerdeckt(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct{ roh, gezeigt string }{
		"ohne Anmeldung":    {"http://proxy.firma:3128", "http://proxy.firma:3128"},
		"nur ein Name":      {"http://anna@proxy.firma:3128", "http://anna@proxy.firma:3128"},
		"Name und Kennwort": {"http://anna:geheim@proxy.firma:3128", "http://anna:xxxxx@proxy.firma:3128"},
		"ohne Schema":       {"anna:geheim@proxy.firma:3128", "anna:xxxxx@proxy.firma:3128"},
		"@ im Kennwort":     {"http://anna:ge@heim@proxy.firma:3128", "http://anna:xxxxx@proxy.firma:3128"},
		"/ im Kennwort":     {"http://anna:ge/heim@proxy.firma:3128", "http://anna:xxxxx@proxy.firma:3128"},
		"leer":              {"", ""},
		"Ausnahmeliste":     {"localhost,.firma", "localhost,.firma"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if gezeigt := domain.ProxyadresseOhneKennwort(fall.roh); gezeigt != fall.gezeigt {
				t.Errorf("%q wurde zu %q, erwartet %q", fall.roh, gezeigt, fall.gezeigt)
			}
		})
	}
}

// **Und dieselbe Verdeckung gilt für die Einstellung**, nicht nur für den
// Netzstand: ein Wert aus der Zeit, bevor das Kennwort abgewiesen wurde,
// stünde sonst im Feld.
func TestEineEinstellungZeigtKeinProxykennwort(t *testing.T) {
	t.Parallel()

	roh := "http://anna:geheim@proxy.firma:3128" //nolint:gosec // ein Beispiel, kein Kennwort

	for _, schluessel := range []string{domain.NetzProxyHTTP, domain.NetzProxyHTTPS} {
		if gezeigt := domain.Einstellungsanzeige(schluessel, roh); gezeigt != "http://anna:xxxxx@proxy.firma:3128" { //nolint:gosec // verdeckt, kein Kennwort
			t.Errorf("%s zeigt %q", schluessel, gezeigt)
		}
	}

	// **Andere Einstellungen bleiben, wie sie sind** — ein Doppelpunkt und
	// ein @ sind in einem Ordnerpfad kein Kennwort.
	if gezeigt := domain.Einstellungsanzeige(domain.NetzProxyAus, "a:b@c"); gezeigt != "a:b@c" {
		t.Errorf("die Ausnahmeliste zeigt %q", gezeigt)
	}
}
