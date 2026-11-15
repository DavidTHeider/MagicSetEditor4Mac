package service

// **Die automatischen Dimensionen** —
// #1985.
//
// MSE legt für jedes Kartenfeld mit `show statistics:` eine Dimension an,
// **vor** den erklärten (`Game::validate`, `src/data/game.cpp`, Zeilen
// 68–77). Bei einem Auswahlfeld rechnet sie `primary_choice(card.<feld>)`,
// gruppiert nach den Wahlen der obersten Ebene und färbt nach
// `choice colors:` (`StatsDimension::StatsDimension(const Field&)`,
// `src/data/statistics.cpp`, Zeilen 27–56); sonst rechnet sie
// `card.<feld>`. Ein Bildfeld geht ohne Angabe nicht hinein
// (`src/data/field/image.hpp`, Zeile 27; Fassung f9d9356 vom 2021-07-06).
//
// Gemessen am 2026-10-01 über `full-magic-pack/data`: `magic.mse-game`
// schaltet die Statistik an allen Feldern ab, die übrigen sechs Spiele
// bekommen je zwei bis fünf automatische Dimensionen.

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestJedesFeldMitStatistikWirdEineDimension(t *testing.T) {
	t.Parallel()

	felder := []domain.Field{
		{
			Name: "rarity", Kind: domain.KindChoice, Statistik: true,
			Beschreibung:    "Die Seltenheit",
			Choices:         []string{"common", "mythic rare"},
			Statistikfarben: map[string]string{"mythic rare": "rgb(213,100,13)"},
		},
		{Name: "rule text", Kind: domain.KindText, Statistik: true},
		{Name: "notes", Kind: domain.KindText},
		{Name: "image", Kind: domain.KindImage, Statistik: true},
	}

	erwartet := []Auszaehlangabe{
		{
			Name: "rarity", Beschreibung: "Die Seltenheit",
			Skript:  "primary_choice(card.rarity)",
			Gruppen: []string{"common", "mythic rare"},
			Farben:  map[string]string{"mythic rare": "rgb(213,100,13)"},
		},
		{Name: "rule text", Skript: "card.rule_text"},
		{Name: "image", Skript: "card.image"},
	}

	if gefunden := automatischeDimensionen(felder, nil); !reflect.DeepEqual(gefunden, erwartet) {
		t.Errorf("Dimensionen\n%+v\nerwartet\n%+v", gefunden, erwartet)
	}
}

// **Eine erklärte Dimension gleichen Namens geht vor** — eine Abweichung mit
// Absicht: MSE zeigte beide, und eine Auswahl nach Namen fände dann nur
// noch die automatische. Im Pack kommt das nicht vor (gemessen, siehe oben).
func TestEineErklaerteDimensionGehtVor(t *testing.T) {
	t.Parallel()

	felder := []domain.Field{
		{Name: "farbe", Kind: domain.KindChoice, Statistik: true},
		{Name: "name", Kind: domain.KindText, Statistik: true},
	}
	erklaert := []Auszaehlangabe{{Name: "farbe", Skript: "card.farbe"}}

	gefunden := automatischeDimensionen(felder, erklaert)
	if len(gefunden) != 1 || gefunden[0].Name != "name" {
		t.Errorf("Dimensionen %+v", gefunden)
	}
}

// **Und im Spiel stehen sie vorn** — an der einzigen Stelle, an der das
// Spielmodell entsteht.
func TestDieAutomatischenStehenVorDenErklaerten(t *testing.T) {
	t.Parallel()

	_, speicher := saatplatz(t)

	vorlage, befunde, da := speicher.StilFuerSet(t.Context(), 7)
	if !da {
		t.Fatalf("kein Stil, Befunde %v", befunde)
	}

	namen := make([]string, 0, len(vorlage.Auszaehldimensionen))
	for _, eine := range vorlage.Auszaehldimensionen {
		namen = append(namen, eine.Name)
	}

	if len(namen) < 3 || namen[0] != "name" || namen[1] != "text" || namen[len(namen)-1] != "farbe" {
		t.Errorf("Reihenfolge %v — erwartet die Felder vorn und die erklärte `farbe` am Ende", namen)
	}
}

// auszaehlkarten gibt die Karten eines Sets aus dem Speicher.
type auszaehlkarten []*domain.CardRow

func (k auszaehlkarten) Karten(context.Context, int64) ([]*domain.CardRow, error) {
	return k, nil
}

// **Der ganze Weg: Kartenwerte hinein, gefärbte Balken heraus.** Die erklärte
// `farbe` heißt in dieser Kopie der Saat anders, damit das Feld seine eigene
// Dimension bekommt; das Feld bekommt `choice colors:`, in **beiden**
// Schreibweisen — `mythic_rare`-artig mit Unterstrich und mit Leerzeichen.
func TestEineAutomatischeDimensionZaehltGefaerbt(t *testing.T) {
	t.Parallel()

	wurzel, speicher := saatplatz(t)
	spiel := filepath.Join(wurzel, "probe.mse-game", "game")
	zeileErsetzen(t, spiel, "\tchoice: rot\n\tchoice: blau\n",
		"\tchoice: rot\n\tchoice: hell blau\n\tchoice colors:\n\t\trot: rgb(1,2,3)\n\t\thell_blau: rgb(4,5,6)\n")
	zeileErsetzen(t, spiel, "statistics dimension:\n\tname: farbe\n",
		"statistics dimension:\n\tname: farbe erklaert\n")

	dienst := NeuerAuszaehldienst(NeueVorlagenauszaehlung(speicher, auszaehlkarten{
		{ID: 1, SetID: 7, Values: []domain.Value{{Field: "farbe", Text: "rot"}}},
		{ID: 2, SetID: 7, Values: []domain.Value{{Field: "farbe", Text: "hell blau"}}},
		{ID: 3, SetID: 7, Values: []domain.Value{{Field: "farbe", Text: "rot"}}},
	}))

	zaehlung, fehler := dienst.Auszaehlen(t.Context(), 7, "farbe")
	if fehler != nil {
		t.Fatalf("Auszaehlen: %v", fehler)
	}

	farben := map[string]string{}
	karten := map[string]int{}

	for _, gruppe := range zaehlung.Gruppen {
		farben[gruppe.Name] = gruppe.Farbe
		karten[gruppe.Name] = gruppe.Karten
	}

	if farben["rot"] != "rgb(1,2,3)" || farben["hell blau"] != "rgb(4,5,6)" ||
		karten["rot"] != 2 || karten["hell blau"] != 1 {
		t.Errorf("Gruppen %+v", zaehlung.Gruppen)
	}
}
