package domain_test

// **Was ein Wunsch über sich sagt** — Werkstück `T4` von Phase 5.6
// (#1469).
//
// # Was hier gemessen wird
//
// **Dass beide Angaben eine Vorgabe haben, die nichts kostet.** Eine Zeile aus
// einem älteren Bestand trägt sie nicht; sie muss trotzdem eine Antwort geben,
// und die Antwort muss die neutrale sein — sonst verschöbe die Migration eine
// Reihenfolge, die niemand angefasst hat.
//
// **Und dass sie an einem Besitz nichts bedeuten.** Wer eine gewünschte Karte
// erwirbt, dreht `Wunsch` um; ein stehengebliebener Rang dürfte danach keine
// Reihenfolge mehr bestimmen.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einWunschexemplar baut eine gewünschte Zeile.
func einWunschexemplar(anders func(*domain.Exemplar)) domain.Exemplar {
	exemplar := domain.Exemplar{
		Printing: "a1b2c3d4", Menge: 2, Sprache: "de",
		Zustand: domain.ZustandNearMint, Wunsch: true,
	}

	if anders != nil {
		anders(&exemplar)
	}

	return exemplar
}

func TestEinWunschOhneRangIstNormal(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		exemplar domain.Exemplar
		erwartet domain.Wunschrang
	}{
		"ohne Angabe":     {einWunschexemplar(nil), domain.RangNormal},
		"mit Angabe":      {einWunschexemplar(func(e *domain.Exemplar) { e.Wunschrang = domain.RangHoch }), domain.RangHoch},
		"niedrig gewählt": {einWunschexemplar(func(e *domain.Exemplar) { e.Wunschrang = domain.RangNiedrig }), domain.RangNiedrig},

		// **Ein Besitz hat keine Dringlichkeit.** Wer die Karte bekommt,
		// dreht `Wunsch` um — und ein stehengebliebener Rang bestimmte
		// danach eine Reihenfolge, die niemand gewollt hat.
		"an einem Besitz": {einWunschexemplar(func(e *domain.Exemplar) {
			e.Wunsch, e.Wunschrang = false, domain.RangSehrHoch
		}), domain.RangNormal},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if gilt := fall.exemplar.Dringlichkeit(); gilt != fall.erwartet {
				t.Errorf("es gilt %q, erwartet war %q", gilt, fall.erwartet)
			}
		})
	}
}

func TestEinWunschOhneArtMeintGenauDiesenDruck(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		exemplar domain.Exemplar
		erwartet domain.Wunschart
	}{
		// **Genau und nicht „egal"**: jede heutige Wunschzeile nennt einen
		// Druck, und aus „ohne Angabe" ein „jeder Druck" zu machen weitete
		// beim Migrieren jeden Wunsch, den jemand genau so gemeint hat.
		"ohne Angabe": {einWunschexemplar(nil), domain.WunschGenau},
		"jeder Druck": {einWunschexemplar(func(e *domain.Exemplar) {
			e.Wunschart = domain.WunschJederDruck
		}), domain.WunschJederDruck},
		"an einem Besitz": {einWunschexemplar(func(e *domain.Exemplar) {
			e.Wunsch, e.Wunschart = false, domain.WunschJederDruck
		}), domain.WunschGenau},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if gilt := fall.exemplar.Druckbindung(); gilt != fall.erwartet {
				t.Errorf("es gilt %q, erwartet war %q", gilt, fall.erwartet)
			}
		})
	}
}

// **Was an einem Besitz nichts bedeutet, wird auch nicht aufbewahrt.** Sonst
// stünde die Angabe im Bestand und wartete darauf, dass jemand sie liest.
func TestNormalisierenRaeumtDenRangEinesBesitzes(t *testing.T) {
	t.Parallel()

	exemplar := einWunschexemplar(func(e *domain.Exemplar) {
		e.Wunsch = false
		e.Wunschrang, e.Wunschart = domain.RangSehrHoch, domain.WunschJederDruck
	})

	exemplar.Normalisieren()

	if exemplar.Wunschrang != "" || exemplar.Wunschart != "" {
		t.Errorf("es blieb %q und %q stehen", exemplar.Wunschrang, exemplar.Wunschart)
	}

	// **Die Gegenprobe**: an einem Wunsch bleibt beides stehen.
	wunsch := einWunschexemplar(func(e *domain.Exemplar) {
		e.Wunschrang, e.Wunschart = domain.RangSehrHoch, domain.WunschJederDruck
	})

	wunsch.Normalisieren()

	if wunsch.Wunschrang != domain.RangSehrHoch || wunsch.Wunschart != domain.WunschJederDruck {
		t.Errorf("am Wunsch blieb %q und %q", wunsch.Wunschrang, wunsch.Wunschart)
	}
}

func TestEinErfundenerRangIstEinBefund(t *testing.T) {
	t.Parallel()

	for feld, exemplar := range map[string]domain.Exemplar{
		"wunschrang": einWunschexemplar(func(e *domain.Exemplar) { e.Wunschrang = "dringend" }),
		"wunschart":  einWunschexemplar(func(e *domain.Exemplar) { e.Wunschart = "irgendein" }),
	} {
		t.Run(feld, func(t *testing.T) {
			t.Parallel()

			befunde := exemplar.Validate()
			if len(befunde) != 1 {
				t.Fatalf("es kamen %d Befunde: %v", len(befunde), befunde)
			}

			if befunde[0].Field != feld {
				t.Errorf("der Befund nennt %q", befunde[0].Field)
			}
		})
	}

	// **Die Gegenprobe, und sie ist die Hälfte der Aussage** (§10): ein
	// Befund über jeden Wert wäre kein Befund über den falschen. Jeder
	// gültige Rang und jede gültige Art gehen durch.
	for _, rang := range domain.Wunschraenge() {
		for _, art := range domain.Wunscharten() {
			gueltig := einWunschexemplar(func(e *domain.Exemplar) {
				e.Wunschrang, e.Wunschart = rang, art
			})

			if befunde := gueltig.Validate(); len(befunde) != 0 {
				t.Errorf("%q mit %q gab %v", rang, art, befunde)
			}
		}
	}
}

// **Die Reihenfolge ist die Aussage** — von niedrig nach sehr hoch. Eine
// Fläche zählt sie auf, und eine Sortierung nach Dringlichkeit liest sie.
func TestDieRaengeStehenInIhrerOrdnung(t *testing.T) {
	t.Parallel()

	raenge := domain.Wunschraenge()

	erwartet := []domain.Wunschrang{
		domain.RangNiedrig, domain.RangNormal, domain.RangHoch, domain.RangSehrHoch,
	}

	if len(raenge) != len(erwartet) {
		t.Fatalf("es sind %d Ränge", len(raenge))
	}

	for stelle, rang := range erwartet {
		if raenge[stelle] != rang {
			t.Errorf("an Stelle %d steht %q statt %q", stelle, raenge[stelle], rang)
		}
	}

	// **Gewicht statt Textvergleich**: „hoch" ist alphabetisch kleiner als
	// „normal", und eine Sortierung über die Zeichenkette drehte die
	// Reihenfolge still um.
	for stelle := 1; stelle < len(erwartet); stelle++ {
		if erwartet[stelle-1].Gewicht() >= erwartet[stelle].Gewicht() {
			t.Errorf("%q wiegt nicht weniger als %q", erwartet[stelle-1], erwartet[stelle])
		}
	}

	// **Ein erfundener Rang wiegt wie `normal`** und nicht wie „ganz oben".
	// Eine Zeile, die aus einer fremden Fassung kommt, darf sich nicht an
	// den Anfang der Liste schreiben, nur weil niemand ihr Wort kennt.
	if unbekannt := domain.Wunschrang("dringend").Gewicht(); unbekannt != domain.RangNormal.Gewicht() {
		t.Errorf("ein erfundener Rang wiegt %d", unbekannt)
	}
}
