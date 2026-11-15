package domain_test

// **Das Deck** — Werkstück 2 von Phase 4.6
// (#1293),
// nach Vermerk [0061](../../docs/entscheidungen/0061-was-ein-deck-hier-ist.md).
//
// # Warum `new` und nicht `ported`
//
// **Gesucht am 2026-09-01** in allen drei Quellen aus §11: MSE2 (`f9d9356`)
// kennt weder `deck` noch `sideboard` noch `legal`; die `.mse-game`- und
// `.mse-style`-Pakete des Packs führen die Wörter nur in Regeltexten und eine
// **Packart** namens „tournament pack"; `magic-set-generator` kennt nichts
// davon. Die Belege stehen im Vermerk.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// feld sagt, ob ein Befund zu diesem Feld dabei ist.
func feld(befunde []*domain.FieldError, name string) bool {
	for _, einer := range befunde {
		if einer.Field == name {
			return true
		}
	}

	return false
}

// TestEinDeckplatzNenntGenauEinenBezug ist die Grundregel aus dem Vermerk.
func TestEinDeckplatzNenntGenauEinenBezug(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name  string
		platz domain.Deckplatz
		gut   bool
	}{
		{
			"eine eigene Karte",
			domain.Deckplatz{KarteID: 7, Menge: 4, Bereich: domain.BereichHauptdeck},
			true,
		},
		{
			"ein echter Druck",
			domain.Deckplatz{
				Printing: "a1b2c3d4-0000-0000-0000-000000000000",
				Menge:    1, Bereich: domain.BereichSideboard,
			},
			true,
		},
		{
			// **Niemals beides** — dann bliebe offen, was gilt.
			"beides zugleich",
			domain.Deckplatz{
				KarteID:  7,
				Printing: "a1b2c3d4-0000-0000-0000-000000000000",
				Menge:    1, Bereich: domain.BereichHauptdeck,
			},
			false,
		},
		{
			// **Und niemals keines** — ein Platz ohne Bezug ist kein Platz.
			"nichts von beidem",
			domain.Deckplatz{Menge: 1, Bereich: domain.BereichHauptdeck},
			false,
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			befunde := fall.platz.Validate()

			if fall.gut && len(befunde) > 0 {
				t.Errorf("abgewiesen: %+v", befunde)
			}

			if !fall.gut && len(befunde) == 0 {
				t.Error("kam durch")
			}
		})
	}
}

// **Eine Menge unter eins ist kein Platz.**
func TestEineMengeUnterEinsWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for _, menge := range []int{0, -1} {
		platz := domain.Deckplatz{KarteID: 7, Menge: menge, Bereich: domain.BereichHauptdeck}

		if befunde := platz.Validate(); !feld(befunde, "menge") {
			t.Errorf("Menge %d kam durch: %+v", menge, befunde)
		}
	}
}

// **Ein unbekannter Bereich ebenso.**
//
// **Zwei und nicht mehr**: Hauptdeck und Sideboard. Was ein drittes wäre, weiß
// niemand, und ein leerer Bereich wäre eine stille Zuordnung.
func TestEinUnbekannterBereichWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for _, bereich := range []domain.Deckbereich{"", "vielleicht", "Hauptdeck"} {
		platz := domain.Deckplatz{KarteID: 7, Menge: 1, Bereich: bereich}

		if befunde := platz.Validate(); !feld(befunde, "bereich") {
			t.Errorf("Bereich %q kam durch: %+v", bereich, befunde)
		}
	}
}

// **Eine Druckkennung, die keine UUID ist, wird abgewiesen.**
//
// Dieselbe Prüfung wie am Exemplar und am Kartenverweis: ein Druck ist eine
// Scryfall-UUID, und was keine ist, findet dort nichts.
func TestEineUnbrauchbareDruckkennungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	platz := domain.Deckplatz{Printing: "kein-druck", Menge: 1, Bereich: domain.BereichHauptdeck}

	if befunde := platz.Validate(); !feld(befunde, "printing") {
		t.Errorf("kam durch: %+v", befunde)
	}
}

// TestEinDeckBrauchtNamenUndSpiel ist die Regel am Deck selbst.
func TestEinDeckBrauchtNamenUndSpiel(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name string
		deck domain.Deck
		gut  bool
	}{
		{"vollständig", domain.Deck{Name: "Mein Deck", Spiel: "magic"}, true},
		{"ohne Namen", domain.Deck{Spiel: "magic"}, false},
		{"nur Leerraum als Name", domain.Deck{Name: "   ", Spiel: "magic"}, false},

		// **Ohne Spiel gäbe es kein Format** — der Vermerk sagt es.
		{"ohne Spiel", domain.Deck{Name: "Mein Deck"}, false},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			befunde := fall.deck.Validate()

			if fall.gut && len(befunde) > 0 {
				t.Errorf("abgewiesen: %+v", befunde)
			}

			if !fall.gut && len(befunde) == 0 {
				t.Error("kam durch")
			}
		})
	}
}

// **Der Name wird geglättet, nicht bei der Prüfung** — dieselbe Trennung wie
// am Exemplar: eine Prüfung, die nebenbei ändert, lässt den Aufrufer im
// Unklaren, was gespeichert wurde.
func TestDerDeckIdentitaetsnameWirdGeglaettet(t *testing.T) {
	t.Parallel()

	deck := domain.Deck{Name: "  Mein Deck  ", Spiel: " magic "}
	deck.Normalisieren()

	if deck.Name != "Mein Deck" {
		t.Errorf("Name %q", deck.Name)
	}

	if deck.Spiel != "magic" {
		t.Errorf("Spiel %q", deck.Spiel)
	}
}

// **Gezählt werden Karten, nicht Plätze.**
//
// Vier Wald sind vier Karten und **ein** Platz — und jedes Format zählt die
// Karten. Ohne diese Unterscheidung hielte ein Deck aus zwanzig Plätzen zu je
// vier Karten die Mindestgröße 60 nicht ein und bekäme trotzdem grünes Licht.
func TestGezaehltWerdenKartenUndNichtPlaetze(t *testing.T) {
	t.Parallel()

	deck := domain.Deck{
		Name: "Zählprobe", Spiel: "magic",
		Plaetze: []domain.Deckplatz{
			{KarteID: 1, Menge: 4, Bereich: domain.BereichHauptdeck},
			{KarteID: 2, Menge: 4, Bereich: domain.BereichHauptdeck},
			{KarteID: 3, Menge: 2, Bereich: domain.BereichSideboard},
		},
	}

	if gezaehlt := deck.Zaehle(domain.BereichHauptdeck); gezaehlt != 8 {
		t.Errorf("Hauptdeck: %d, erwartet 8", gezaehlt)
	}

	if gezaehlt := deck.Zaehle(domain.BereichSideboard); gezaehlt != 2 {
		t.Errorf("Sideboard: %d, erwartet 2", gezaehlt)
	}
}

// **Ein leeres Deck zählt null** — und das ist ein Wert, kein Fehler.
func TestEinLeeresDeckZaehltNull(t *testing.T) {
	t.Parallel()

	leer := domain.Deck{Name: "Leer", Spiel: "magic"}

	if gezaehlt := leer.Zaehle(domain.BereichHauptdeck); gezaehlt != 0 {
		t.Errorf("%d, erwartet 0", gezaehlt)
	}
}

// **Und das Glätten fasst auch die Plätze an** — eine Druckkennung mit
// Leerraum käme sonst durch die Prüfung und stünde mit Leerraum im Bestand.
func TestDasGlaettenFasstAuchDiePlaetzeAn(t *testing.T) {
	t.Parallel()

	deck := domain.Deck{
		Name: "Probe", Spiel: "magic", Notiz: "  eine Notiz  ",
		Plaetze: []domain.Deckplatz{
			{
				Printing: "  a1b2c3d4-0000-0000-0000-000000000000  ",
				Menge:    1, Bereich: domain.BereichHauptdeck,
			},
		},
	}

	deck.Normalisieren()

	if deck.Notiz != "eine Notiz" {
		t.Errorf("Notiz %q", deck.Notiz)
	}

	if deck.Plaetze[0].Printing != "a1b2c3d4-0000-0000-0000-000000000000" {
		t.Errorf("Printing %q", deck.Plaetze[0].Printing)
	}
}
