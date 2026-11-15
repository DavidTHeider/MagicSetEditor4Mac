package service_test

// **Die Leerwahl kommt nach der Vorgabe, nicht davor** — gefunden am
// Kanonlauf vom 2026-10-04: der mittlere Abstand zu MSE2 stieg von 9 auf 31,
// weil jede Karte des Kanons farblos gezeichnet wurde.
//
// Seit #2040 schreibt eine leere Mehrfachwahl ihren Leernamen (`colorless`
// an `card color`). Der erste Durchgang lief **vor** den Vorgabeskripten,
// und die rechnen nur leere Felder — `card color` war danach nie mehr leer.
// MSE2 schreibt den Leernamen über `mutateDontChangeDefault`: der Wert bleibt
// Vorgabe, und das Vorgabeskript rechnet ihn neu
// (`MultipleChoiceValue::normalForm`, `src/data/field/multiple_choice.cpp`,
// Zeilen 110–113, Fassung f9d9356 vom 2021-07-06).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// farbvorlage ist eine Vorlage mit `card color` wie in `magic.mse-game`.
//
// **Die Vorgabe ruft eine Funktion des Spiels** — wie dort
// `card_color(…)`, die erst das `init script:` anlegt. Im ersten Durchgang
// ist sie noch unbekannt, und das Ergebnis ist leer.
func farbvorlage(farbe string) service.Vorlage {
	vorlage := vorgabenvorlage()
	vorlage.Spielskript = "farbe_der_karte := { " + farbe + " }"
	vorlage.Kartenfelder[1] = domain.Field{
		Name: "card color", Kind: domain.KindMultipleChoice,
		Choices: []string{"white", "black"}, Leerwahl: "colorless",
		Default: "farbe_der_karte()",
	}

	return vorlage
}

func TestEineLeereMehrfachwahlNimmtErstIhreVorgabe(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ name, vorgabe, erwartet string }{
		{"die Vorgabe rechnet eine Farbe", `"black"`, "black"},
		{"die Vorgabe rechnet nichts", `""`, "colorless"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			plan, _, err := service.NeuerPlanbauer().
				MitSchriften(schriftprobe{breite: 0.5}).
				Stilplan(t.Context(), &domain.CardRow{ID: 1, Number: 1}, farbvorlage(fall.vorgabe))
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			// **Gesucht, nicht nur geprüft**: fehlte das Feld, bliebe eine
			// Schleife über die Vorgaben still grün (§3).
			gefunden := ""
			for _, vorgabe := range plan.Vorgaben {
				if vorgabe.Feld == "card color" {
					gefunden = vorgabe.Wert
				}
			}

			if gefunden != fall.erwartet {
				t.Errorf("card color = %q, erwartet %q — %v", gefunden, fall.erwartet, plan.Vorgaben)
			}
		})
	}
}
