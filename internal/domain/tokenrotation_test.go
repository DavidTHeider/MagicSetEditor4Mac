package domain_test

// **Rotation, letzte Nutzung und der benannte Principal** — Werkstück O2 von
// Phase 5.2 (#1426),
// nach [Vermerk 0080](../../docs/entscheidungen/0080-maschinenzugaenge-und-widerrufbare-sitzungen.md).

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Token nennt, für wen es arbeitet.**
//
// **Der Fall, ohne den „benannter Principal" eine Absichtserklärung wäre**:
// ein Feld, das leer bleiben darf, bleibt leer.
func TestEinTokenBrauchtSeinenPrincipal(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"leer":         "",
		"nur Leerraum": "   ",
		"mit Rändern":  " Nachtlauf ",
	}

	for name, principal := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			eines := gutesToken()
			eines.Principal = principal

			fehler := eines.Validate()
			if !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("%q wurde angenommen: %v", principal, fehler)
			}
		})
	}
}

// **Ein Vorgänger gilt bis zu seiner Frist und keinen Augenblick länger.**
func TestDerVorgaengerGiltBisZurFrist(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)

	faelle := map[string]struct {
		bis  string
		gilt bool
	}{
		"kein Vorgänger":   {"", false},
		"noch in Frist":    {jetzt.Add(time.Minute).Format(time.RFC3339), true},
		"genau abgelaufen": {jetzt.Format(time.RFC3339), false},
		"lange vorbei":     {jetzt.Add(-time.Hour).Format(time.RFC3339), false},
		"unlesbar":         {"irgendwann", false},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			eines := gutesToken()
			eines.VorgaengerBis = fall.bis

			if eines.VorgaengerGilt(jetzt) != fall.gilt {
				t.Errorf("VorgaengerGilt(%q) = %v, erwartet %v",
					fall.bis, !fall.gilt, fall.gilt)
			}
		})
	}
}

// **Die Übergangsfrist ist eine Zahl dieses Programms und keine Meinung des
// Aufrufers.**
func TestDieUebergangsfristStehtFest(t *testing.T) {
	t.Parallel()

	if domain.Tokenuebergang != time.Hour {
		t.Errorf("Tokenuebergang = %v, erwartet eine Stunde", domain.Tokenuebergang)
	}
}

// **Vier Ablaufwahlen, und jede rechnet aus derselben Uhr.**
func TestDieAblaufwahlenRechnenAusEinerUhr(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)

	erwartet := map[string]string{
		"tag":   jetzt.AddDate(0, 0, 1).Format(time.RFC3339),
		"monat": jetzt.AddDate(0, 0, 30).Format(time.RFC3339),
		"jahr":  jetzt.AddDate(1, 0, 0).Format(time.RFC3339),
		"nie":   "",
	}

	wahlen := domain.Ablaufwahlen(jetzt)
	if len(wahlen) != len(erwartet) {
		t.Fatalf("%d Wahlen, erwartet %d", len(wahlen), len(erwartet))
	}

	for _, wahl := range wahlen {
		wollte, bekannt := erwartet[wahl.Name]
		if !bekannt {
			t.Errorf("die Wahl %q kennt niemand", wahl.Name)

			continue
		}

		if wahl.Ablauf != wollte {
			t.Errorf("%q läuft ab am %q, erwartet %q", wahl.Name, wahl.Ablauf, wollte)
		}

		if fehler := (domain.Zugriffstoken{
			ProfilID: 1, Bezeichnung: "x", Principal: "y",
			Scope: domain.ScopeLesen, Ablauf: wahl.Ablauf,
		}).Validate(); fehler != nil {
			t.Errorf("die Wahl %q gibt einen Ablauf, den niemand annimmt: %v", wahl.Name, fehler)
		}
	}
}

// **Die Reihenfolge ist die Steigerung** — sie steht in der Oberfläche so da.
func TestDieAblaufwahlenStehenInDerSteigerung(t *testing.T) {
	t.Parallel()

	wahlen := domain.Ablaufwahlen(time.Now())
	namen := make([]string, 0, len(wahlen))

	for _, wahl := range wahlen {
		namen = append(namen, wahl.Name)
	}

	erwartet := []string{"tag", "monat", "jahr", "nie"}
	for stelle, name := range erwartet {
		if namen[stelle] != name {
			t.Errorf("an Stelle %d steht %q, erwartet %q", stelle, namen[stelle], name)
		}
	}
}
