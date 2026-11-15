package domain_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// TestEinSicherungskennwortHatZwoelfZeichen hält OPS-06.AK-02 (#1723): ein
// Kennwort, das eine Sicherung außer Haus schützt, wird nur geraten, wenn es
// kurz ist — gezählt werden Zeichen und nicht Bytes.
func TestEinSicherungskennwortHatZwoelfZeichen(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name, kennwort string
		gut            bool
	}{
		{"leer", "", false},
		{"elf Zeichen", "elfzeichen!", false},
		{"zwölf Zeichen", "zwölfzeichen", true},
		{"elf Umlaute sind elf Zeichen", "ääääääääääö", false},
		{"ein Satz", "korrekt pferd batterie heftklammer", true},
		{"nur Leerraum", "                ", false},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			fehler := domain.PruefeSicherungskennwort(fall.kennwort)

			var feld *domain.FieldError
			if fall.gut != (fehler == nil) {
				t.Fatalf("%q: %v", fall.kennwort, fehler)
			}

			if !fall.gut && (!errors.As(fehler, &feld) || feld.Field != "kennwort") {
				t.Errorf("kein Befund am Feld: %v", fehler)
			}
		})
	}
}
