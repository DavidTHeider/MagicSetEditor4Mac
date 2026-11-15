package service_test

// **Eine Mehrfachwahl beginnt leer, eine Wahl mit ihrer ersten** — F02
// (#2244).
//
//	inline MultipleChoiceValue(const MultipleChoiceFieldP& field) : ChoiceValue(field, false) {}
//
// (`src/data/field/multiple_choice.hpp`, Zeile 52, Fassung f9d9356 vom
// 2021-07-06) — `initial_first_choice` ist **falsch**, und `ChoiceValue` nimmt
// dann `initial:` oder die leere Zeichenkette (`src/data/field/choice.cpp`,
// Zeilen 279–285). Hier bekam jede Mehrfachwahl ohne Wert ihre erste Wahl:
// `magic-nouveau` zeichnete auf jede Karte das Symbol für Todesberührung,
// auf die ganze Breite der Schlüsselwortleiste gezogen.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDerAnfangswertKenntDieMehrfachwahl(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name     string
		feld     domain.Field
		erwartet string
	}{
		{"eine Wahl beginnt mit ihrer ersten", domain.Field{Kind: domain.KindChoice, Choices: []string{"none", "tombstone"}}, "none"},
		{"eine Mehrfachwahl beginnt leer", domain.Field{Kind: domain.KindMultipleChoice, Choices: []string{"deathtouch", "flying"}}, ""},
		{"eine Mehrfachwahl mit initial", domain.Field{Kind: domain.KindMultipleChoice, Initial: "flying", Choices: []string{"deathtouch", "flying"}}, "flying"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if wert := service.Anfangswert(fall.feld); wert != fall.erwartet {
				t.Errorf("%q, erwartet %q", wert, fall.erwartet)
			}
		})
	}
}

// **Und ein leeres Zusatzfeld mit Wahlbildern zeichnet nichts** — der Fall
// aus `magic-nouveau`.
func TestEineLeereMehrfachwahlZeichnetKeinWahlbild(t *testing.T) {
	t.Parallel()

	feld := msepaket.Feldstil{
		Feld: "keyword icons", Links: "42", Oben: "321", Breite: "297", Hoehe: "24",
		Darstellung: "image list",
		Wahlbilder:  []msepaket.Wahlbild{{Wahl: "deathtouch", Bild: "deathtouch.png"}, {Wahl: "flying", Bild: "flying.png"}},
	}

	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Zusatzstile: []msepaket.Feldstil{feld},
		Zusatzfelder: []domain.Field{{
			Name: "keyword icons", Kind: domain.KindMultipleChoice, Choices: []string{"deathtouch", "flying"},
		}},
	}

	plan, _, err := service.NeuerPlanbauer().Stilplan(t.Context(),
		&domain.CardRow{ID: 1, Number: 1}, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art == render.ArtBild && ebene.Feld == "keyword icons" {
			t.Errorf("die leere Mehrfachwahl zeichnet %s", ebene.Bild.Beschreibung())
		}
	}
}
