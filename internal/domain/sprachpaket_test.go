package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine eigene Übersetzung hat eine gültige Kennung** — LOC-02.AK-01
// (#1688).
func TestDieKennungEinerEigenenSprache(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name       string
		kennung    string
		schluessel domain.Schluessel
	}{
		{name: "eine Sprache", kennung: "pt"},
		{name: "mit Gegend", kennung: "pt-BR"},
		{name: "mit Schrift", kennung: "zh-Hant"},
		{name: "mit Schrift und Gegend", kennung: "zh-Hant-TW"},
		{name: "mit Zahlengegend", kennung: "es-419"},
		{name: "drei Buchstaben", kennung: "yue"},
		{name: "privat", kennung: "ja-x-meine"},
		{name: "leer", kennung: "", schluessel: domain.SchluesselSprachkennungUngueltig},
		{name: "großgeschrieben", kennung: "PT", schluessel: domain.SchluesselSprachkennungUngueltig},
		{name: "Gegend klein", kennung: "pt-br", schluessel: domain.SchluesselSprachkennungUngueltig},
		{name: "Unterstrich", kennung: "pt_BR", schluessel: domain.SchluesselSprachkennungUngueltig},
		{name: "privater Teil zu lang", kennung: "ja-x-korrektur", schluessel: domain.SchluesselSprachkennungUngueltig},
		{name: "Pfad", kennung: "../de", schluessel: domain.SchluesselSprachkennungUngueltig},
		{name: "zu lang", kennung: "de-x-" + strings.Repeat("abcdefgh-", 5) + "a", schluessel: domain.SchluesselSprachkennungUngueltig},
		{name: "eingebaut", kennung: "ja", schluessel: domain.SchluesselSprachkennungEingebaut},
		{name: "eingebaut in BCP 47", kennung: "zh-Hans", schluessel: domain.SchluesselSprachkennungEingebaut},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			fehler := domain.PruefeSprachkennung(fall.kennung)
			if fall.schluessel == "" {
				if fehler != nil {
					t.Fatalf("abgewiesen: %v", fehler)
				}

				return
			}

			var feld *domain.FieldError
			if !errors.As(fehler, &feld) || feld.Key != fall.schluessel || feld.Field != "kennung" {
				t.Fatalf("%v statt %s", fehler, fall.schluessel)
			}

			if feld.Werte["kennung"] != fall.kennung {
				t.Errorf("Werte %v", feld.Werte)
			}
		})
	}
}

// **Die Angaben eines Pakets** — Name, Basis, Stand und die zwei Freitexte.
func TestDieAngabenEinesSprachpakets(t *testing.T) {
	t.Parallel()

	gut := domain.Sprachpaket{
		Kennung: "pt-BR", Name: "Português do Brasil", Basis: "en",
		Stand: domain.StandEntwurf, Lizenz: "CC0-1.0", Herkunft: "Ana",
	}

	faelle := []struct {
		name    string
		aendere func(*domain.Sprachpaket)
		feld    string
	}{
		{name: "vollständig", aendere: func(*domain.Sprachpaket) {}},
		{name: "ohne Lizenz und Herkunft", aendere: func(p *domain.Sprachpaket) { p.Lizenz, p.Herkunft = "", "" }},
		{name: "falsche Kennung", aendere: func(p *domain.Sprachpaket) { p.Kennung = "pt_BR" }, feld: "kennung"},
		{name: "ohne Namen", aendere: func(p *domain.Sprachpaket) { p.Name = "  " }, feld: "name"},
		{name: "zu langer Name", aendere: func(p *domain.Sprachpaket) { p.Name = strings.Repeat("ä", 61) }, feld: "name"},
		{name: "Steuerzeichen im Namen", aendere: func(p *domain.Sprachpaket) { p.Name = "a\nb" }, feld: "name"},
		{name: "fremde Basis", aendere: func(p *domain.Sprachpaket) { p.Basis = "pt" }, feld: "basis"},
		{name: "fremder Stand", aendere: func(p *domain.Sprachpaket) { p.Stand = "fertig" }, feld: "stand"},
		{name: "zu lange Lizenz", aendere: func(p *domain.Sprachpaket) { p.Lizenz = strings.Repeat("x", 201) }, feld: "lizenz"},
		{name: "Nullbyte in der Herkunft", aendere: func(p *domain.Sprachpaket) { p.Herkunft = "a\x00" }, feld: "herkunft"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			paket := gut
			fall.aendere(&paket)

			fehler := paket.Pruefe()
			if fall.feld == "" {
				if fehler != nil {
					t.Fatalf("abgewiesen: %v", fehler)
				}

				return
			}

			var feld *domain.FieldError
			if !errors.As(fehler, &feld) || feld.Field != fall.feld {
				t.Fatalf("%v statt eines Befunds an %s", fehler, fall.feld)
			}
		})
	}
}

// **Jeder Stand ist einer der vier**, und die Basis eine Fenstersprache.
func TestDieStaendeUndDieFenstersprachen(t *testing.T) {
	t.Parallel()

	if got := strings.Join(domain.Fenstersprachen(), " "); got != "de en es fr it ja ru zh" {
		t.Errorf("Fenstersprachen %q", got)
	}

	sprachen := domain.Fenstersprachen()
	sprachen[0] = "xx"

	if domain.Fenstersprachen()[0] != "de" {
		t.Error("die Liste ist keine Kopie")
	}

	for _, stand := range []domain.Sprachstand{
		domain.StandEntwurf, domain.StandMaschinell, domain.StandGeprueft, domain.StandAbgenommen,
	} {
		if !stand.Gueltig() {
			t.Errorf("%s ungültig", stand)
		}
	}
}

// **Welche eigene Übersetzung gilt, ist eine Einstellung des Menschen** —
// leer heißt keine, sonst eine gültige Kennung, und der Befund steht am
// Wert (LOC-02).
func TestDieEinstellungDerEigenenSprache(t *testing.T) {
	t.Parallel()

	for _, wert := range []string{"", "pt-BR"} {
		if fehler := domain.PruefeEinstellung(domain.EinstellungEigeneSprache, wert); fehler != nil {
			t.Errorf("%q abgewiesen: %v", wert, fehler)
		}
	}

	var feld *domain.FieldError
	if fehler := domain.PruefeEinstellung(domain.EinstellungEigeneSprache, "de"); !errors.As(fehler, &feld) ||
		feld.Field != "wert" || feld.Key != domain.SchluesselSprachkennungEingebaut ||
		feld.Werte["kennung"] != "de" {
		t.Errorf("eine eingebaute: %v", fehler)
	}

	if ebene, bekannt := domain.EbeneVon(domain.EinstellungEigeneSprache); !bekannt || ebene != domain.EbeneNutzer {
		t.Errorf("Ebene %s", ebene)
	}
}
