package domain_test

// **Die vier Pfade eines eigenen Dienstzeugnisses** — NET-03.AK-01
// (#1707).
//
// Geprüft wird hier nur die **Gestalt**: leer oder ein absoluter Pfad. Ob
// die Datei da ist, ob das Paar passt und ob die Kette trägt, prüft der
// Dienst vor seinem Start (`internal/leitung`, `PruefeZeugnis`) — die Domäne
// fasst keine Platte an.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinZeugnispfadWirdGeprueft(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		wert     string
		abweisen bool
	}{
		"leer heißt: das Zeugnis der Peeridentität": {"", false},
		"ein absoluter Pfad":                        {"/Users/anna/pki/dienst.pem", false},
		"ein relativer Pfad":                        {"pki/dienst.pem", true},
		"eine Tilde wird nicht aufgelöst":           {"~/pki/dienst.pem", true},
		"ein Nullzeichen":                           {"/pki/dienst\x00.pem", true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := domain.PruefeZeugnispfad(fall.wert)

			if fall.abweisen != (fehler != nil) {
				t.Errorf("%q: %v", fall.wert, fehler)
			}

			if fehler != nil && !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("keine Validierung: %v", fehler)
			}
		})
	}
}

// **Alle vier gehören dem Gerät** — ein Pfad auf diesem Rechner reist nicht
// mit, und sie wirken beim nächsten Start des Dienstes.
func TestDieZeugnispfadeGehoerenDemGeraet(t *testing.T) {
	t.Parallel()

	for _, schluessel := range domain.Zeugnispfade() {
		if ebene, _ := domain.EbeneVon(schluessel); ebene != domain.EbeneGeraet {
			t.Errorf("%s steht auf der Ebene %q", schluessel, ebene)
		}

		if wirkung := domain.EinstellungsWirkung(schluessel); wirkung != domain.WirkungNeustart {
			t.Errorf("%s wirkt %q", schluessel, wirkung)
		}

		if domain.PruefeEinstellung(schluessel, "relativ.pem") == nil {
			t.Errorf("%s nimmt einen relativen Pfad", schluessel)
		}
	}

	if len(domain.Zeugnispfade()) != 4 {
		t.Errorf("es sind %d Pfade", len(domain.Zeugnispfade()))
	}
}
