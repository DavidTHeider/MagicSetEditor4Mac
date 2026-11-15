package domain_test

// **Wohin ein Dienst bindet** — Werkstück N6 von Phase 5.4
// (#1572).
//
// # Wonach hier gefragt wird
//
// Nach der einen Frage, an der TLS und Zugangstoken hängen: **verlässt diese
// Bindung den Rechner?** Sie wird an der Adresse beantwortet und nicht am
// Wunsch — und was keine Adresse ist, zählt als Öffnung nach außen.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineBindungSagtObSieDenRechnerVerlaesst(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		adresse   string
		verlaesst bool
	}{
		"leer ist die Schleife":    {"", false},
		"nur Leerzeichen auch":     {"   ", false},
		"die Schleife":             {"127.0.0.1", false},
		"eine andere Schleife":     {"127.0.0.2", false},
		"die Schleife in IPv6":     {"::1", false},
		"jede Schnittstelle":       {"0.0.0.0", true},
		"jede Schnittstelle in v6": {"::", true},
		"eine echte Schnittstelle": {"192.168.1.7", true},
		"ein Name zählt als außen": {"localhost", true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			bindung := domain.Bindung{Adresse: fall.adresse}

			if gab := bindung.VerlaesstDenRechner(); gab != fall.verlaesst {
				t.Errorf("%q verlässt den Rechner: %v", fall.adresse, gab)
			}
		})
	}
}

// **Ein Name wird abgewiesen und nicht aufgelöst.**
//
// **Das ist die Zusage, die zählt**: `localhost` sähe aus wie die Schleife
// und ist es auf einem fremd eingerichteten Rechner nicht. Eine
// Sicherheitsfrage über einen Namen zu beantworten, den dieser Rechner erst
// zur Laufzeit nachschlägt, ist kein Weg.
func TestEineBindungWirdGeprueft(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		bindung  domain.Bindung
		abweisen bool
	}{
		"leer":               {domain.Bindung{}, false},
		"die Schleife":       {domain.Bindung{Adresse: "127.0.0.1"}, false},
		"jede Schnittstelle": {domain.Bindung{Adresse: "0.0.0.0", Port: 8080}, false},
		"Port 0":             {domain.Bindung{Port: 0}, false},
		"höchster Port":      {domain.Bindung{Port: 65535}, false},
		"ein Name":           {domain.Bindung{Adresse: "localhost"}, true},
		"ein Rechnername":    {domain.Bindung{Adresse: "mein.mac"}, true},
		"Port zu groß":       {domain.Bindung{Port: 65536}, true},
		"Port negativ":       {domain.Bindung{Port: -1}, true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := domain.PruefeBindung(fall.bindung)

			if fall.abweisen && fehler == nil {
				t.Errorf("%+v kam durch", fall.bindung)
			}

			if !fall.abweisen && fehler != nil {
				t.Errorf("%+v wurde abgewiesen: %v", fall.bindung, fehler)
			}

			if fehler != nil && !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("der Fehler ist keine Validierung: %v", fehler)
			}
		})
	}
}

// **Leer heißt die Schleife — an einer Stelle und nicht an dreien.**
func TestEineLeereBindungIstDieSchleife(t *testing.T) {
	t.Parallel()

	// **Die Vorgabe ist die leere Bindung selbst** — es gab hier eine
	// `Schleifenbindung()`, die sie gebaut hätte, und **keinen produktiven
	// Aufrufer** dafür (§10, erste Dimension). Weg damit: wer die Vorgabe
	// meint, lässt das Feld leer, und genau das wird hier gemessen.
	vorgabe := domain.Bindung{}

	if vorgabe.Wirkadresse() != domain.Schleifenadresse {
		t.Errorf("die Vorgabe zeigt auf %q", vorgabe.Wirkadresse())
	}

	if vorgabe.VerlaesstDenRechner() {
		t.Error("die Vorgabe verlässt den Rechner")
	}

	if (domain.Bindung{}).Wirkadresse() != domain.Schleifenadresse {
		t.Error("eine leere Bindung zeigt nicht auf die Schleife")
	}

	if gab := (domain.Bindung{Port: 8080}).Adresstext(); gab != "127.0.0.1:8080" {
		t.Errorf("der Adresstext ist %q", gab)
	}

	// **Und die IPv6-Gestalt bekommt ihre Klammern** — `net.JoinHostPort`
	// setzt sie, und ohne sie läse jeder Lauscher die Adresse falsch.
	if gab := (domain.Bindung{Adresse: "::1", Port: 80}).Adresstext(); gab != "[::1]:80" {
		t.Errorf("der Adresstext ist %q", gab)
	}
}

// **Zwei Dienstarten und nicht beliebig viele.**
func TestEsGibtZweiDienstarten(t *testing.T) {
	t.Parallel()

	arten := domain.Dienstarten()
	if len(arten) != 2 {
		t.Fatalf("es sind %d Dienstarten", len(arten))
	}

	if arten[0] != domain.DienstREST || arten[1] != domain.DienstMCP {
		t.Errorf("die Arten sind %v", arten)
	}
}
