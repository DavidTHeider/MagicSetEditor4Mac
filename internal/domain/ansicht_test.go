package domain_test

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Ansicht einer Liste wird gemerkt, je Bereich und Profil** —
// UX-11.AK-04 (#1683).
// Dieselbe Bauform wie der Arbeitsplatz: das Profil erbt vom Menschen.
func TestDieAnsichtEinesBereichsIstEineArbeitsplatzeinstellung(t *testing.T) {
	faelle := []struct {
		name       string
		schluessel string
		werte      []string
		vorgabe    string
	}{
		{"die Sets", domain.AnsichtSets, []string{"kacheln", "liste"}, "kacheln"},
		{"die Decks", domain.AnsichtDecks, []string{"liste", "kacheln"}, "liste"},
		{
			"die Sammlung", domain.AnsichtSammlung,
			[]string{"tabelle", "galerie", "gruppen"},
			"tabelle",
		},
	}
	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			if werte := domain.ErlaubteEinstellungswerte(fall.schluessel); !slices.Equal(
				werte, append([]string{domain.Erben}, fall.werte...)) {
				t.Fatalf("Werte im Profil %v", werte)
			}

			if fehler := domain.PruefeNutzereinstellung(fall.schluessel, fall.vorgabe); fehler != nil {
				t.Fatalf("beim Menschen abgewiesen: %v", fehler)
			}

			if ebene, _ := domain.EbeneVon(fall.schluessel); ebene != domain.EbeneArbeitsplatz {
				t.Fatalf("Ebene %q", ebene)
			}
		})
	}
}
