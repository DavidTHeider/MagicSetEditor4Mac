package domain_test

// **Die Marke über einen Stand** — Werkstück O3 von Phase 5.2
// (#1427).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Derselbe Stand gibt dieselbe Marke — auch in anderer Reihenfolge
// geschrieben.**
//
// Eine `map` durchläuft Go absichtlich in wechselnder Reihenfolge; eine Marke,
// die davon abhinge, wechselte zwischen zwei Aufrufen, und jede Übernahme
// scheiterte an einem Konflikt, den es nicht gibt.
func TestDieselbeLageGibtDieselbeMarke(t *testing.T) {
	t.Parallel()

	erste := domain.Standmarke(7, 3,
		map[string]string{"sprache": "de", "erscheinungsbild": "hell"},
		map[string]string{"barrierefreiheit.kontrast": "hoch"})

	zweite := domain.Standmarke(7, 3,
		map[string]string{"erscheinungsbild": "hell", "sprache": "de"},
		map[string]string{"barrierefreiheit.kontrast": "hoch"})

	if erste != zweite {
		t.Errorf("%q gegen %q", erste, zweite)
	}

	if erste == "" {
		t.Error("die Marke ist leer")
	}
}

// **Jede Änderung am Stand ändert die Marke.**
func TestJedeAenderungAendertDieMarke(t *testing.T) {
	t.Parallel()

	grund := domain.Standmarke(7, 3,
		map[string]string{"sprache": "de"},
		map[string]string{"barrierefreiheit.kontrast": "hoch"})

	faelle := map[string]string{
		"ein anderes Profil": domain.Standmarke(8, 3,
			map[string]string{"sprache": "de"},
			map[string]string{"barrierefreiheit.kontrast": "hoch"}),
		"ein anderer Mensch": domain.Standmarke(7, 4,
			map[string]string{"sprache": "de"},
			map[string]string{"barrierefreiheit.kontrast": "hoch"}),
		"ein anderer Profilwert": domain.Standmarke(7, 3,
			map[string]string{"sprache": "en"},
			map[string]string{"barrierefreiheit.kontrast": "hoch"}),
		"ein Profilwert mehr": domain.Standmarke(7, 3,
			map[string]string{"sprache": "de", "schriftgroesse": "gross"},
			map[string]string{"barrierefreiheit.kontrast": "hoch"}),
		"ein anderer Nutzerwert": domain.Standmarke(7, 3,
			map[string]string{"sprache": "de"},
			map[string]string{"barrierefreiheit.kontrast": "system"}),
		"ein Nutzerwert weniger": domain.Standmarke(7, 3,
			map[string]string{"sprache": "de"}, map[string]string{}),

		// **Die Trennung der Ebenen zählt mit.** Ohne sie beschriebe dieselbe
		// Marke zwei verschiedene Lagen: ein Schlüssel im Profil und
		// derselbe beim Menschen sind nicht dasselbe.
		"derselbe Wert auf der anderen Ebene": domain.Standmarke(7, 3,
			map[string]string{},
			map[string]string{
				"sprache": "de", "barrierefreiheit.kontrast": "hoch",
			}),
	}

	for name, marke := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if marke == grund {
				t.Errorf("%s gibt dieselbe Marke %q", name, marke)
			}
		})
	}
}
