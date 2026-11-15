//go:build darwin

package schluesselbund_test

// **Was auf der Befehlszeile mitreist, wird geprüft** — Dienst und Konto.
//
// Sie sind Bezeichner dieses Programms und keine Eingabe; wer einen aus einer
// Eingabe baut, bekommt ein Nein statt einer verstümmelten Befehlszeile.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinUnbrauchbarerNameWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name   string
		dienst string
		konto  string
	}{
		{"Dienst mit Leerzeichen", "mse probe", "konto"},
		{"Konto mit Leerzeichen", "mse", "kon to"},
		{"Dienst mit Gegenschrägstrich", `mse\probe`, "konto"},
		{"Konto mit Anführungszeichen", "mse", `kon"to`},
		{"Dienst leer", "", "konto"},
		{"Konto leer", "mse", ""},
		{"Konto mit Umlaut", "mse", "schlüssel"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			bund := probebund(t)

			// **Alle drei Wege prüfen** — sonst käme einer davon durch, und
			// zwar ausgerechnet der, den niemand geprüft hat.
			if err := bund.Schreibe(t.Context(), fall.dienst, fall.konto,
				"geheim"); !errors.Is(err, domain.ErrValidation) {
				t.Errorf("Schreibe gab %v, erwartet eine Validierung", err)
			}

			if _, err := bund.Lies(t.Context(), fall.dienst,
				fall.konto); !errors.Is(err, domain.ErrValidation) {
				t.Errorf("Lies gab %v, erwartet eine Validierung", err)
			}

			if err := bund.Loesche(t.Context(), fall.dienst,
				fall.konto); !errors.Is(err, domain.ErrValidation) {
				t.Errorf("Loesche gab %v, erwartet eine Validierung", err)
			}
		})
	}
}

// **Und ein brauchbarer Name kommt durch** — sonst prüfte der Lauf oben nur,
// dass die Prüfung immer Nein sagt.
func TestEinBrauchbarerNameKommtDurch(t *testing.T) {
	t.Parallel()

	bund := probebund(t)

	for _, name := range []string{"mse", "MSE4Mac", "scrydex.api-key", "a_b.c-1"} {
		if err := bund.Schreibe(t.Context(), name, name, "geheim"); err != nil {
			t.Errorf("%q wurde abgewiesen: %v", name, err)
		}
	}
}
