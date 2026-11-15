package domain_test

// **Was ein Profil vom Menschen erbt** — Werkstück 9 von Phase 4.12
// (#1351),
// Punkt 9 der Kommentierung vom 2026-09-05.

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die vier Erscheinungseinstellungen stehen auf beiden Ebenen.**
//
// **Und nur diese vier.** Sie beschreiben, wie die Anwendung **aussieht und
// spricht** — das ist eine Eigenschaft des Menschen, nicht der
// Arbeitsumgebung. `aktualisierung` betrifft den Rechner, `bildkonflikt`
// einen Ausgabelauf, und beide erbt niemand.
func TestDieVierErscheinungenStehenAufBeidenEbenen(t *testing.T) {
	t.Parallel()

	beimNutzer := domain.Nutzereinstellungsreihe()

	for _, schluessel := range domain.Uebersteuerbare() {
		if !slices.Contains(beimNutzer, schluessel) {
			t.Errorf("%q lässt sich übersteuern, steht aber nicht beim Nutzer",
				schluessel)
		}

		if !slices.Contains(domain.ErlaubteEinstellungswerte(schluessel),
			domain.Erben) {
			t.Errorf("%q kennt %q nicht als Wert", schluessel, domain.Erben)
		}

		if domain.EinstellungsVorgabe(schluessel) != domain.Erben {
			t.Errorf("%q steht ohne Wahl auf %q statt auf %q", schluessel,
				domain.EinstellungsVorgabe(schluessel), domain.Erben)
		}
	}

	if len(domain.Uebersteuerbare()) != 4 {
		t.Errorf("es sind %d — die Zahl steht hier, damit eine fünfte auffällt",
			len(domain.Uebersteuerbare()))
	}
}

// **Der Nutzer kennt `erben` nicht.**
//
// **Weil er die unterste Ebene ist.** Wer dort „wie der Nutzer" wählen
// könnte, zeigte auf sich selbst — und die Frage hätte keine Antwort mehr.
func TestDerNutzerKenntErbenNicht(t *testing.T) {
	t.Parallel()

	for _, schluessel := range domain.Nutzereinstellungsreihe() {
		if slices.Contains(domain.ErlaubteNutzerwerte(schluessel), domain.Erben) {
			t.Errorf("%q kennt %q — die unterste Ebene erbt von niemandem",
				schluessel, domain.Erben)
		}
	}
}

// **Die Regel: das Profil sticht, sonst der Mensch, sonst die Vorgabe.**
func TestDieUebersteuerungFolgtDreiStufen(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		schluessel, profil, nutzer, erwartet string
	}{
		"das Profil hat gewählt": {
			domain.EinstellungErscheinungsbild, "dunkel", "hell", "dunkel",
		},
		"das Profil erbt, der Mensch hat gewählt": {
			domain.EinstellungErscheinungsbild, domain.Erben, "hell", "hell",
		},
		"beide erben": {
			domain.EinstellungErscheinungsbild, domain.Erben, "", "system",
		},
		"die Sprache hat keine Vorgabe": {
			domain.EinstellungSprache, domain.Erben, "", "",
		},
		"eine Einstellung, die niemand erbt": {
			domain.EinstellungAktualisierung, "melden", "holen", "melden",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gilt := domain.Verschmolzen(fall.schluessel, fall.profil, fall.nutzer)
			if gilt != fall.erwartet {
				t.Errorf("es gilt %q, erwartet war %q", gilt, fall.erwartet)
			}
		})
	}
}
