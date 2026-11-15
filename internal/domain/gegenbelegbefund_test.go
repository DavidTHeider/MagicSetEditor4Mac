package domain

// **Ein Gegenbeleg, der nicht passt, wird festgehalten** —
// #2500.

import (
	"errors"
	"testing"
	"time"
)

// einBefund besteht jede Prüfung.
func einBefund() Gegenbelegbefund {
	return Gegenbelegbefund{
		Vorgang: "v-1", Von: "bert", Abdruck: "a1", Mangel: MangelZeilen,
		Grund: "der Beleg trägt 2 Zeilen und meiner 3", Empfangen: time.Now(),
	}
}

func TestEinGegenbelegbefundNenntVorgangAbsenderMangelUndGrund(t *testing.T) {
	faelle := []struct {
		name    string
		aendern func(*Gegenbelegbefund)
		feld    string
	}{
		{"ohne Vorgang", func(b *Gegenbelegbefund) { b.Vorgang = "" }, "vorgang"},
		{"ohne Absender", func(b *Gegenbelegbefund) { b.Von = "" }, "von"},
		{"mit einem Mangel, den es nicht gibt", func(b *Gegenbelegbefund) { b.Mangel = "farbe" }, "mangel"},
		{"ohne Grund", func(b *Gegenbelegbefund) { b.Grund = "" }, "grund"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			befund := einBefund()
			fall.aendern(&befund)

			var feldfehler *FieldError
			if fehler := befund.Pruefe(); !errors.As(fehler, &feldfehler) || feldfehler.Field != fall.feld {
				t.Errorf("Pruefe gab %v, erwartet ein Fehler am Feld %q", fehler, fall.feld)
			}
		})
	}

	t.Run("nimmt jeden der fünf Mängel an", func(t *testing.T) {
		for _, mangel := range DieGegenbelegmaengel {
			befund := einBefund()
			befund.Mangel = mangel

			if fehler := befund.Pruefe(); fehler != nil {
				t.Errorf("%s: %v", mangel, fehler)
			}
		}

		if len(DieGegenbelegmaengel) != 5 {
			t.Errorf("%d Mängel, erwartet fünf", len(DieGegenbelegmaengel))
		}
	})
}
