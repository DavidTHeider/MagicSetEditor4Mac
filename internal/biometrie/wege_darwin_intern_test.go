//go:build darwin

package biometrie

// **Die Wege, die dieser Rechner nicht hergibt** — Werkstück 7 von Phase 4.12
// (#1351).
//
// **Ein interner Test**, weil die Naht hier liegt. Was er prüft, ließe sich
// sonst nur auf drei verschiedenen Rechnern prüfen: einem ohne Touch ID,
// einem mit Face ID und einem, dessen Sensor gesperrt ist.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/tmc/apple/localauthentication"
)

// mitBindung setzt die Naht für einen Lauf.
func mitBindung(t *testing.T, doppel bindung) {
	t.Helper()

	vorher := oeffne
	oeffne = func() bindung { return doppel }

	t.Cleanup(func() { oeffne = vorher })
}

// stilleBindung ist eine, die nichts kann und nichts sagt.
func stilleBindung() bindung {
	return bindung{
		kann:      func() (bool, error) { return false, nil },
		merkmal:   func() string { return "" },
		pruefe:    func(context.Context, string) (bool, error) { return false, nil },
		schliesse: func() {},
	}
}

// **Ein Rechner ohne Touch ID sagt, dass er keines hat.**
func TestEinRechnerOhneTouchIDSagtEs(t *testing.T) {
	mitBindung(t, stilleBindung())

	auskunft := stand(t.Context())
	if auskunft.Vorhanden {
		t.Fatal("vorhanden, obwohl die Bindung nein sagt")
	}

	if !strings.Contains(auskunft.Grund, "Touch ID") {
		t.Errorf("der Grund ist %q", auskunft.Grund)
	}
}

// **Und der Grund des Frameworks schlägt den eigenen Satz.**
//
// **Weil „kein Touch ID eingerichtet" und „gesperrt nach zu vielen
// Versuchen" zwei Lagen mit zwei Antworten sind** — und nur das Framework
// kennt den Unterschied.
func TestDerGrundDesFrameworksSchlaegtDenEigenen(t *testing.T) {
	doppel := stilleBindung()
	doppel.kann = func() (bool, error) {
		return false, errors.New("Touch ID ist gesperrt")
	}

	mitBindung(t, doppel)

	if grund := stand(t.Context()).Grund; grund != "Touch ID ist gesperrt" {
		t.Errorf("der Grund ist %q", grund)
	}
}

// **Ein Rechner mit einem Merkmal nennt es.**
func TestEinRechnerMitMerkmalNenntEs(t *testing.T) {
	doppel := stilleBindung()
	doppel.kann = func() (bool, error) { return true, nil }
	doppel.merkmal = func() string { return "Face ID" }

	mitBindung(t, doppel)

	auskunft := stand(t.Context())
	if !auskunft.Vorhanden || auskunft.Art != "Face ID" {
		t.Errorf("angekommen ist %+v", auskunft)
	}
}

// **Eine Bindung, die abstürzt, wird zur Antwort und nicht zum Ende.**
func TestEineAbstuerzendeBindungWirdZurAntwort(t *testing.T) {
	doppel := stilleBindung()
	doppel.kann = func() (bool, error) { panic("purego ist weg") }

	mitBindung(t, doppel)

	auskunft := stand(t.Context())
	if auskunft.Vorhanden {
		t.Fatal("vorhanden, obwohl die Bindung abgestürzt ist")
	}

	if !strings.Contains(auskunft.Grund, "ausgesetzt") {
		t.Errorf("der Grund ist %q", auskunft.Grund)
	}
}

// **Die Abfrage unterscheidet drei Ausgänge.**
//
// **Abgelehnt und gescheitert sind nicht dasselbe**: „der Finger passte
// nicht" sagt, es noch einmal zu versuchen; „das Framework antwortet nicht"
// sagt, den Code aus der App zu nehmen (§8).
func TestDieAbfrageUnterscheidetIhreAusgaenge(t *testing.T) {
	faelle := map[string]struct {
		pruefe   func(context.Context, string) (bool, error)
		erwartet string
	}{
		"bestätigt": {
			pruefe:   func(context.Context, string) (bool, error) { return true, nil },
			erwartet: "",
		},
		"abgelehnt": {
			pruefe:   func(context.Context, string) (bool, error) { return false, nil },
			erwartet: "nicht bestätigt",
		},
		"gescheitert": {
			pruefe: func(context.Context, string) (bool, error) {
				return false, errors.New("der Sensor ist belegt")
			},
			erwartet: "der Sensor ist belegt",
		},
		"abgestürzt": {
			pruefe:   func(context.Context, string) (bool, error) { panic("purego ist weg") },
			erwartet: "ausgesetzt",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			doppel := stilleBindung()
			doppel.pruefe = fall.pruefe

			mitBindung(t, doppel)

			fehler := Bestaetige(t.Context(), "zum Wechseln des Profils")

			if fall.erwartet == "" {
				if fehler != nil {
					t.Fatalf("gemeldet wurde: %v", fehler)
				}

				return
			}

			if fehler == nil || !strings.Contains(fehler.Error(), fall.erwartet) {
				t.Errorf("gemeldet wurde: %v", fehler)
			}
		})
	}
}

// **Der Grund reist bis in den Dialog.**
//
// **Weil macOS ihn verlangt** und weil er der Satz ist, den die Person liest,
// bevor sie den Finger auflegt — ein leerer wäre eine Abfrage ohne Anlass.
func TestDerGrundReistBisInDenDialog(t *testing.T) {
	gesehen := ""

	doppel := stilleBindung()
	doppel.pruefe = func(_ context.Context, grund string) (bool, error) {
		gesehen = grund

		return true, nil
	}

	mitBindung(t, doppel)

	if fehler := Bestaetige(t.Context(), "zum Wechseln des Profils"); fehler != nil {
		t.Fatalf("bestätigen: %v", fehler)
	}

	if gesehen != "zum Wechseln des Profils" {
		t.Errorf("angekommen ist %q", gesehen)
	}
}

// **Jedes Merkmal bekommt seinen Namen — und ein unbekanntes keinen.**
//
// **Eigennamen werden nicht übersetzt** (§9): „Touch ID" heißt in jeder
// Sprache so. Ein unbekanntes Merkmal gibt **nichts** und keinen Platzhalter
// — die Oberfläche sagt dann „Biometrie" und nicht „LABiometryType(7)".
func TestJedesMerkmalBekommtSeinenNamen(t *testing.T) {
	faelle := map[localauthentication.LABiometryType]string{
		localauthentication.LABiometryTypeTouchID: "Touch ID",
		localauthentication.LABiometryTypeFaceID:  "Face ID",
		localauthentication.LABiometryTypeNone:    "",
	}

	for art, erwartet := range faelle {
		if gegeben := artNach(art); gegeben != erwartet {
			t.Errorf("%v gab %q, erwartet war %q", art, gegeben, erwartet)
		}
	}
}
