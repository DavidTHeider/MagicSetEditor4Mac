package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestKontaktlage(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name    string
		kontakt domain.Peerkontakt
		lage    domain.Kontaktlage
	}{
		{"unbestätigt", domain.Peerkontakt{Stand: domain.KontaktUnbestaetigt}, domain.LageUnbestaetigt},
		{"bestätigt", domain.Peerkontakt{Stand: domain.KontaktBestaetigt}, domain.LageBestaetigt},
		{"gesperrt heißt widerrufen", domain.Peerkontakt{Stand: domain.KontaktGesperrt}, domain.LageWiderrufen},
		{
			"ein neuer Schlüssel für jemanden, den ich kenne",
			domain.Peerkontakt{Stand: domain.KontaktUnbestaetigt, Vorgaenger: "12D3KooWalt"},
			domain.LageSchluesselGeaendert,
		},
		{
			"nach dem Vergleich ist er bestätigt, Vorgänger oder nicht",
			domain.Peerkontakt{Stand: domain.KontaktBestaetigt, Vorgaenger: "12D3KooWalt"},
			domain.LageBestaetigt,
		},
		{
			"gesperrt bleibt widerrufen, Vorgänger oder nicht",
			domain.Peerkontakt{Stand: domain.KontaktGesperrt, Vorgaenger: "12D3KooWalt"},
			domain.LageWiderrufen,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if lage := fall.kontakt.Lage(); lage != fall.lage {
				t.Fatalf("Lage = %q, erwartet %q", lage, fall.lage)
			}
		})
	}
}

func TestKontaktlagenSindVerschieden(t *testing.T) {
	t.Parallel()

	// **AK-03 verlangt vier unterschiedliche Zustände**, und der fünfte ist
	// der, in dem es gar keinen Kontakt gibt.
	gesehen := map[domain.Kontaktlage]bool{}
	for _, lage := range domain.DieKontaktlagen {
		if gesehen[lage] {
			t.Fatalf("%q zweimal", lage)
		}

		gesehen[lage] = true
	}

	if len(gesehen) != 5 || !gesehen[domain.LageUnbekannt] {
		t.Fatalf("Lagen: %v", domain.DieKontaktlagen)
	}
}
