package domain_test

// **Ein Preisstand trägt seine ganze Reihe** — Werkstück A3 von Phase 5.3
// (#1442).

import (
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// vollerStand ist ein Stand, an dem nichts fehlt.
func vollerStand() domain.Preisstand {
	return domain.Preisstand{
		Printing:   "00000000-0000-4000-8000-000000000001",
		Quelle:     domain.PreisquelleScryfall,
		Markt:      domain.MarktPapier,
		Art:        domain.PreisartVerkauf,
		Waehrung:   "usd",
		Finish:     domain.FinishNonfoil,
		Lage:       domain.PreisGenannt,
		Wert:       417,
		Beobachtet: time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC),
		Erhoben:    time.Date(2026, 9, 11, 8, 0, 0, 0, time.UTC),
	}
}

// **Ein voller Stand hat keine Befunde.**
func TestEinVollerPreisstandGehtDurch(t *testing.T) {
	t.Parallel()

	if befunde := vollerStand().Validate(); len(befunde) != 0 {
		t.Errorf("%+v", befunde)
	}
}

// **Und jede Angabe der Reihe wird geprüft.**
func TestJedeReihenangabeWirdGeprueft(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		aendere func(*domain.Preisstand)
		feld    string
	}{
		"ein Markt, den es nicht gibt": {
			func(p *domain.Preisstand) { p.Markt = "mond" }, "markt",
		},
		"kein Markt": {
			func(p *domain.Preisstand) { p.Markt = "" }, "markt",
		},
		"eine Preisart, die es nicht gibt": {
			func(p *domain.Preisstand) { p.Art = "tausch" }, "art",
		},
		"eine Lage, die es nicht gibt": {
			func(p *domain.Preisstand) { p.Lage = "vielleicht" }, "lage",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			stand := vollerStand()
			fall.aendere(&stand)

			befunde := stand.Validate()
			if len(befunde) != 1 || befunde[0].Field != fall.feld {
				t.Fatalf("%+v", befunde)
			}
		})
	}
}

// **Ein Beobachtungsdatum in der Zukunft ist keins.**
//
// **Und ein fehlendes ist erlaubt**: Scryfall nennt keines — „daily estimates",
// ohne zu sagen, welcher Tag gemeint ist. Wer eines erfände, behauptete eine
// Genauigkeit, die die Quelle nicht hergibt.
func TestDasBeobachtungsdatumDarfFehlenAberNichtLuegen(t *testing.T) {
	t.Parallel()

	ohne := vollerStand()
	ohne.Beobachtet = time.Time{}

	if befunde := ohne.Validate(); len(befunde) != 0 {
		t.Errorf("ohne Datum: %+v", befunde)
	}

	spaeter := vollerStand()
	spaeter.Beobachtet = spaeter.Erhoben.Add(48 * time.Hour)

	befunde := spaeter.Validate()
	if len(befunde) != 1 || befunde[0].Field != "beobachtet" {
		t.Fatalf("%+v", befunde)
	}
}

// **Was nicht genannt ist, hat keinen Wert.**
//
// Die Zusage aus [domain.PreisWeggefallen] und [domain.PreisUngeprueft]: dort
// steht null, und sie bedeutet nichts. Ein Wert daneben wäre eine Zahl, die
// aussieht wie ein Preis.
func TestEinNichtGenannterStandTraegtKeinenWert(t *testing.T) {
	t.Parallel()

	for _, lage := range []domain.Preislage{
		domain.PreisWeggefallen, domain.PreisUngeprueft,
	} {
		t.Run(string(lage), func(t *testing.T) {
			t.Parallel()

			stand := vollerStand()
			stand.Lage = lage

			befunde := stand.Validate()
			if len(befunde) != 1 || befunde[0].Field != "wert" {
				t.Fatalf("%+v", befunde)
			}

			// **Die Gegenprobe**: ohne Wert geht dieselbe Lage durch.
			stand.Wert = 0

			if ohne := stand.Validate(); len(ohne) != 0 {
				t.Errorf("ohne Wert: %+v", ohne)
			}
		})
	}
}

// **Die Reihe ist der Schlüssel, unter dem zwei Stände dasselbe meinen.**
func TestDieReiheUnterscheidetJedeAngabe(t *testing.T) {
	t.Parallel()

	grund := vollerStand().Reihe()

	faelle := map[string]func(*domain.Preisstand){
		"eine andere Quelle":   func(p *domain.Preisstand) { p.Quelle = "mtgjson" },
		"ein anderer Markt":    func(p *domain.Preisstand) { p.Markt = domain.MarktOnline },
		"eine andere Preisart": func(p *domain.Preisstand) { p.Art = domain.PreisartAnkauf },
		"eine andere Währung":  func(p *domain.Preisstand) { p.Waehrung = "eur" },
		"eine andere Ausführung": func(p *domain.Preisstand) {
			p.Finish = domain.FinishFoil
		},
		"ein anderer Zustand": func(p *domain.Preisstand) { p.Zustand = "NM" },
		"eine andere Sprache": func(p *domain.Preisstand) { p.Sprache = "de" },
	}

	for name, aendere := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			anders := vollerStand()
			aendere(&anders)

			if anders.Reihe() == grund {
				t.Errorf("dieselbe Reihe für %s: %q", name, grund)
			}
		})
	}

	// **Und was die Reihe nicht ausmacht, ändert sie nicht.**
	//
	// Ohne diese Gegenprobe sagte der Lauf nur, dass die Reihe **irgendetwas**
	// unterscheidet — auch eine, die jeden Stand für verschieden hielte (§10).
	anderswert := vollerStand()
	anderswert.Wert = 999
	anderswert.Erhoben = anderswert.Erhoben.Add(time.Hour)
	anderswert.Lage = domain.PreisWeggefallen
	anderswert.Wert = 0

	if anderswert.Reihe() != grund {
		t.Errorf("Wert, Zeitpunkt und Lage änderten die Reihe: %q gegen %q",
			anderswert.Reihe(), grund)
	}
}

// **Die Reihe trennt ihre Teile so, dass sie nicht ineinanderlaufen.**
//
// Ohne Trenner wären „usd“ + „nonfoil“ und „usdn“ + „onfoil“ dieselbe Reihe.
func TestDieReiheLaeuftNichtIneinander(t *testing.T) {
	t.Parallel()

	eine := vollerStand()
	eine.Waehrung = "usd"
	eine.Zustand = "NM"

	andere := vollerStand()
	andere.Waehrung = "usdNM"
	andere.Zustand = ""

	if eine.Reihe() == andere.Reihe() {
		t.Errorf("die Teile liefen ineinander: %q", eine.Reihe())
	}

	if strings.TrimSpace(eine.Reihe()) == "" {
		t.Error("die Reihe ist leer")
	}
}
