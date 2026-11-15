//go:build !darwin

package schluesselbund_test

// **Auf anderen Plattformen gibt es hier keinen Schlüsselbund** — und das ist
// ein **Fehler**, keine leere Antwort (§12).
//
// Der Lauf wird hier nicht gefahren; gebaut wird für macOS. Er steht da, damit
// der Bau für Windows und Linux nicht bricht und damit die Zusage aus
// `bund_andere.go` überhaupt eine Prüfung hat, sobald jemand dort baut.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

func TestOhneSchluesselbundGibtEsEinenFehler(t *testing.T) {
	t.Parallel()

	bund := schluesselbund.Anmeldebund()

	if err := bund.Schreibe(t.Context(), "mse", "konto", "geheim"); !errors.Is(
		err, domain.ErrValidation) {
		t.Errorf("Schreibe gab %v, erwartet eine Validierung", err)
	}

	// **Und beim Lesen kommt kein leerer Schlüssel zurück.** Wer hier
	// „nichts" bekäme, bekäme dieselbe Antwort wie bei einem Dienst ohne
	// hinterlegten Schlüssel — und das ist ein anderer Zustand (0063).
	wort, err := bund.Lies(t.Context(), "mse", "konto")
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Lies gab %v, erwartet eine Validierung", err)
	}

	if wort != "" {
		t.Errorf("Lies gab %q zurück", wort)
	}

	if err := bund.Loesche(t.Context(), "mse", "konto"); !errors.Is(
		err, domain.ErrValidation) {
		t.Errorf("Loesche gab %v, erwartet eine Validierung", err)
	}
}
