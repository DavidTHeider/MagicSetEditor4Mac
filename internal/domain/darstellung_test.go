package domain_test

// **Sprache, Darstellung, Zeitzone und Währung — vier Einstellungen und nicht
// eine** — LOC-03.AK-03
// (#1689):
// „Die volle Locale bleibt erhalten", und aus dem Ziel des Issues:
// „Sprache, regionale Darstellung, Zeitzone und Währung müssen getrennt
// konfigurierbar sein."

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Darstellung ist eine volle Locale** — mit Gegend, Schrift und
// allem, was `Intl` versteht. Leer heißt: die des Systems.
func TestDieDarstellungIstEineVolleLocale(t *testing.T) {
	t.Parallel()

	for _, wert := range []string{"", "de-CH", "en-GB", "ja-JP", "sr-Latn-RS"} {
		if fehler := domain.PruefeEinstellung(
			domain.EinstellungDarstellung, wert); fehler != nil {
			t.Errorf("%q wurde abgewiesen: %v", wert, fehler)
		}
	}
}

// **Was keine Locale ist, wird erklärt** — und eine, die anders geschrieben
// gehört, auch: `Intl` im Fenster wirft bei »de_CH«, und eine Einstellung,
// die das Fenster nicht versteht, wäre eine stille (AK-03).
func TestEineFremdeDarstellungWirdErklaert(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"deutsch": "keine Locale",
		"12":      "keine Locale",
		"de_CH":   "de-CH",
		"de-ch":   "de-CH",
	}

	for wert, teil := range faelle {
		fehler := domain.PruefeEinstellung(domain.EinstellungDarstellung, wert)

		if !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("%q: erwartet ErrValidation, bekam %v", wert, fehler)

			continue
		}

		if !strings.Contains(fehler.Error(), teil) {
			t.Errorf("%q: die Meldung %q nennt %q nicht", wert, fehler, teil)
		}
	}
}

// **Eine Zeitzone ist ein Name aus der Zonendatenbank** — leer heißt die
// des Systems.
func TestDieZeitzoneIstEinZonenname(t *testing.T) {
	t.Parallel()

	for _, wert := range []string{"", "Europe/Berlin", "Asia/Tokyo", "UTC"} {
		if fehler := domain.PruefeEinstellung(
			domain.EinstellungZeitzone, wert); fehler != nil {
			t.Errorf("%q wurde abgewiesen: %v", wert, fehler)
		}
	}

	// **»Local« ist kein Name, sondern Gos Wort für „das System"** — und
	// das System heißt hier leer.
	for _, wert := range []string{"Mond/Krater", "Local", "+02:00"} {
		fehler := domain.PruefeEinstellung(domain.EinstellungZeitzone, wert)

		if !errors.Is(fehler, domain.ErrValidation) ||
			!strings.Contains(fehler.Error(), wert) {
			t.Errorf("%q: erwartet ErrValidation mit dem Wert, bekam %v",
				wert, fehler)
		}
	}
}

// **Die Vorgabewährung ist ein Code aus drei Großbuchstaben** — und leer,
// solange niemand eine will. Umgerechnet wird nichts (AK-02).
func TestDieVorgabewaehrungIstEinCode(t *testing.T) {
	t.Parallel()

	for _, wert := range []string{"", "EUR", "JPY", "BHD"} {
		if fehler := domain.PruefeEinstellung(
			domain.VorgabeWaehrung, wert); fehler != nil {
			t.Errorf("%q wurde abgewiesen: %v", wert, fehler)
		}
	}

	faelle := map[string]string{
		"Euro": "drei Buchstaben",
		"€":    "drei Buchstaben",
		"eur":  "EUR",
	}

	for wert, teil := range faelle {
		fehler := domain.PruefeEinstellung(domain.VorgabeWaehrung, wert)

		if !errors.Is(fehler, domain.ErrValidation) ||
			!strings.Contains(fehler.Error(), teil) {
			t.Errorf("%q: erwartet ErrValidation mit %q, bekam %v",
				wert, teil, fehler)
		}
	}
}

// **Jede gehört dorthin, wo sie hingehört.** Die Darstellung dem Menschen,
// wie die Sprache; die Zeitzone dem Rechner, denn sie sagt, wo er steht;
// die Vorgabewährung dem Profil, wie der Vorgabezustand.
func TestDieVierEinstellungenStehenAufIhrerEbene(t *testing.T) {
	t.Parallel()

	faelle := map[string]domain.Ebene{
		domain.EinstellungSprache:     domain.EbeneNutzer,
		domain.EinstellungDarstellung: domain.EbeneNutzer,
		domain.EinstellungZeitzone:    domain.EbeneGeraet,
		domain.VorgabeWaehrung:        domain.EbeneProfil,
	}

	for schluessel, erwartet := range faelle {
		if ebene, bekannt := domain.EbeneVon(schluessel); !bekannt ||
			ebene != erwartet {
			t.Errorf("%s steht auf %q (%v), erwartet %q",
				schluessel, ebene, bekannt, erwartet)
		}
	}
}
