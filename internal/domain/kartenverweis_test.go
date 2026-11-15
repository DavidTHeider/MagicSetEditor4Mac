package domain_test

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Verweis einer entworfenen Karte auf einen echten Druck** — Werkstück 2
// von Phase 3.6
// (#1167).
func TestKartenverweisValidateBefunde(t *testing.T) {
	t.Parallel()

	gesetzt := time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC)

	for name, fall := range map[string]struct {
		verweis  domain.Kartenverweis
		erwartet []string
	}{
		"vollständig": {
			domain.Kartenverweis{
				Printing: "516e137a-d9d9-4ba7-89fb-e680ddb2c2d1",
				Name:     "Lightning Bolt", Setcode: "lea", Nummer: "161",
				Adresse: "https://scryfall.com/card/lea/161", Gesetzt: gesetzt,
			},
			nil,
		},

		// **Ohne Druck ist es kein Verweis.** Genau das unterscheidet ihn von
		// einem getippten Kartennamen.
		"ohne Druck": {
			domain.Kartenverweis{Name: "Lightning Bolt", Gesetzt: gesetzt},
			[]string{"printing"},
		},

		// **Ohne Namen ist er ohne Netz unlesbar.** Eine UUID sagt niemandem,
		// worauf sie zeigt — §16, Regel 5: der letzte gute Stand bleibt, und
		// er muss etwas aussagen.
		"ohne Namen": {
			domain.Kartenverweis{Printing: "516e137a", Gesetzt: gesetzt},
			[]string{"name"},
		},

		"leer": {
			domain.Kartenverweis{},
			[]string{"printing", "name"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			befunde := fall.verweis.Validate()

			if len(befunde) != len(fall.erwartet) {
				t.Fatalf("%d Befunde, erwartet %d: %v", len(befunde), len(fall.erwartet), befunde)
			}

			for stelle, feld := range fall.erwartet {
				if befunde[stelle].Field != feld {
					t.Errorf("Befund %d betrifft %q, erwartet %q",
						stelle, befunde[stelle].Field, feld)
				}
			}
		})
	}
}

// **Was aus einer Eingabe kommt, wird geglättet.**
//
// Dieselbe Trennung wie beim Exemplar: eine Prüfung, die nebenbei ändert,
// lässt den Aufrufer im Unklaren, was gespeichert wurde.
//
// **Und die Kennung wird kleingeschrieben**, weil Scryfall sie so führt — eine
// Kennung in zwei Schreibweisen wären zwei Verweise auf denselben Druck, und
// die Sammlung fände neben dem einen den anderen nicht.
func TestKartenverweisNormalisieren(t *testing.T) {
	t.Parallel()

	verweis := domain.Kartenverweis{
		Printing: "  516E137A-D9D9-4BA7-89FB-E680DDB2C2D1 ",
		Name:     "  Lightning Bolt  ", Setcode: " LEA ", Nummer: " 161 ",
		Adresse: " https://scryfall.com/card/lea/161 ",
	}

	verweis.Normalisieren()

	if verweis.Printing != "516e137a-d9d9-4ba7-89fb-e680ddb2c2d1" {
		t.Errorf("Printing ist %q", verweis.Printing)
	}

	if verweis.Name != "Lightning Bolt" || verweis.Nummer != "161" {
		t.Errorf("Name %q, Nummer %q", verweis.Name, verweis.Nummer)
	}

	// **Der Setcode wird klein**, wie Scryfall ihn führt und wie ihn jede
	// Adresse trägt.
	if verweis.Setcode != "lea" {
		t.Errorf("Setcode ist %q", verweis.Setcode)
	}

	if verweis.Adresse != "https://scryfall.com/card/lea/161" {
		t.Errorf("Adresse ist %q", verweis.Adresse)
	}
}
