// **Was eine Einstellung annimmt** — OPS-01.AK-01
// (#1718):
// ein dokumentiertes Schema nennt je Einstellung ihren **Typ**. Drei Arten,
// und jede folgt aus der Zeile, nicht aus einer zweiten Angabe daneben.
package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineEinstellungSagtWasSieAnnimmt(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name       string
		schluessel string
		art        string
	}{
		{"eine Liste", domain.EinstellungErscheinungsbild, domain.ArtListe},
		{"eine Regel ohne Liste", domain.EinstellungZeitzone, domain.ArtGeprueft},
		{"eine Belegung nach einer Regel", domain.EinstellungZiffernbelegung, domain.ArtGeprueft},
		{"weder Liste noch Regel", domain.AnbieterKennung(domain.Anbieter()[0]), domain.ArtFrei},
		{"ein unbekannter Schlüssel", "gibtsnicht", ""},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if art := domain.EinstellungsArt(fall.schluessel); art != fall.art {
				t.Errorf("EinstellungsArt(%q) = %q, erwartet %q", fall.schluessel, art, fall.art)
			}
		})
	}
}

func TestEineNutzereinstellungSagtWasSieAnnimmt(t *testing.T) {
	t.Parallel()

	for schluessel, art := range map[string]string{
		domain.NutzerKontrast: domain.ArtListe,
		"gibtsnicht":          "",
	} {
		if gelesen := domain.NutzerArt(schluessel); gelesen != art {
			t.Errorf("NutzerArt(%q) = %q, erwartet %q", schluessel, gelesen, art)
		}
	}
}

// **Jede Einstellung hat eine Art** — auch die aus den erzeugten Familien.
// Ohne diesen Lauf stünde eine neue Einstellung ohne Werte als `frei` da,
// ohne dass jemand es entschieden hätte; so steht sie in der Tafel von
// `docs/EINSTELLUNGEN.md`, und der Wächter dort nennt sie.
func TestJedeEinstellungHatEineArt(t *testing.T) {
	t.Parallel()

	for _, schluessel := range domain.Einstellungsreihe() {
		if domain.EinstellungsArt(schluessel) == "" {
			t.Errorf("%s hat keine Art", schluessel)
		}
	}

	for _, schluessel := range domain.Nutzereinstellungsreihe() {
		if domain.NutzerArt(schluessel) == "" {
			t.Errorf("%s hat keine Art beim Menschen", schluessel)
		}
	}
}
