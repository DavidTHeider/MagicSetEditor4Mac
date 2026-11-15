package service_test

// **`set.set_info` ist dasselbe wie `set`**
// (#424).
//
// `REFLECT_N("set_info", data)` und `reflect_set_info_get_member(handler, data)`
// (`src/data/set.cpp`, Zeilen 196 und 206, Fassung f9d9356 vom 2021-07-06)
// legen dieselbe Abbildung zweimal an `set`: einmal unter ihrem Namen, einmal
// namenlos. Der Kommentar im Altcode sagt es — „in scripts, set.something is
// read from the set_info".
//
// Über `Full-Magic-Pack` steht `set.set_info.` **38-mal** da; `set.styling.`,
// `set.game.` und `set.stylesheet.` null-mal (§10).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestSetInfoIstDasselbeWieSet(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld: "name", Links: "10", Oben: "10", Hoehe: "23", Breite: "100",
			Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
		}},
	}

	vorlage := service.Vorlage{
		Stil: stil,
		Kartenfelder: []domain.Field{{
			Name: "name", Kind: domain.KindText,
			// **Beide Wege in einem Ausdruck**, damit der Test nicht grün wird,
			// weil einer von beiden leer ist.
			Default: `{ set.title + "/" + set.set_info.title }`,
		}},
		Setfelder:  []domain.Field{{Name: "title", Kind: domain.KindText}},
		Setangaben: []domain.Value{{Field: "title", Text: "Probe"}},
	}

	plan, befunde, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(), &domain.CardRow{ID: 1, Number: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Errorf("Befunde: %+v", befunde)
	}

	if breite := zugbreite(t, plan, "Probe/Probe"); breite == 0 {
		t.Error("der Plan führt den zusammengesetzten Namen nicht")
	}
}
