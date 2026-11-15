package service_test

// **Ein Packfilter misst keinen Satz** —
// #2466.
//
// Der erste Aufruf von „Zufallspack“ brauchte mit *Blue* (3.693 Karten) an
// der Kommandozeile 80 s; das Profil zeigte zwei Fünftel davon in der
// Inhaltsmessung — Textsatz für `content_width`, je Karte zweimal, samt
// Schriftsuche. **MSE2 wertet einen Packfilter im Kontext der Karte aus,
// ohne sie zu setzen** (`PackInstance::PackInstance`, `src/data/pack.cpp`,
// Zeilen 110–117; gesetzt wird nur beim Zeichnen, `src/render/card/viewer.cpp`,
// Zeilen 42–50, Fassung f9d9356 vom 2021-07-06): die Inhaltsmaße eines nie
// gezeichneten Stils sind dort null (`TextStyle::layout`,
// `src/data/field/text.cpp`, Zeilen 113–119).

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinPackfilterMisstKeinenSatz(t *testing.T) {
	t.Parallel()

	var gefragt []string

	vorlage := service.Vorlage{
		Kartenfelder: []domain.Field{{Name: "name", Kind: domain.KindText}},
		Stil: msepaket.Stil{
			Paket: "satz.mse-style", Breite: 375, Hoehe: 523,
			Felder: []msepaket.Feldstil{{
				Feld: "name", Links: "10", Oben: "10", Hoehe: "23", Breite: "100",
				Schrift: msepaket.Schrift{Name: "Beleren", Groesse: "10"},
			}},
		},
		Packs: []msepaket.Pack{
			{Name: "alle", Filter: `card.name != ""`},
			// **Wie in MSE2 an einer nie gezeichneten Karte**: null.
			{Name: "ungesetzt", Filter: `card_style.name.content_width == 0`},
		},
	}

	karten := []*domain.CardRow{
		{ID: 1, SetID: 1, Values: []domain.Value{{Field: "name", Text: "Moorgeist"}}},
		{ID: 2, SetID: 1, Values: []domain.Value{{Field: "name", Text: "Waldboden"}}},
	}

	bereich := service.NeuerPackbereich(nil, setkarten{karten: karten},
		service.NeuerPlanbauer().MitStilen(stileJeKarte{set: vorlage}).
			MitSchriften(gefragteSchriften{namen: &gefragt}))

	angebot, _, fehler := bereich.Angebot(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("Angebot: %v", fehler)
	}

	if len(gefragt) != 0 {
		t.Errorf("gesucht wurden %d Schriften: %v", len(gefragt), gefragt)
	}

	for _, art := range angebot {
		if art.Karten != len(karten) {
			t.Errorf("%s: %d Karten, erwartet %d", art.Name, art.Karten, len(karten))
		}
	}
}

// **Mit mehreren Arbeitern dasselbe Ergebnis in derselben Reihenfolge** —
// #2466: die Karten eines großen Sets werden verteilt ausgewertet, wie in der
// Problemprüfung (PERF02). Ein Pack zieht aus der Liste in der Reihenfolge
// des Sets; eine andere Reihenfolge änderte, was derselbe Samen zieht.
func TestMehrereArbeiterFilternInDerReihenfolgeDesSets(t *testing.T) {
	t.Parallel()

	vorlage := service.Vorlage{
		Kartenfelder: []domain.Field{{Name: "name", Kind: domain.KindText}},
		Stil:         msepaket.Stil{Paket: "satz.mse-style"},
		Packs: []msepaket.Pack{
			{Name: "gerade", Auswahl: "no replace", Filter: `to_int(card.name) mod 2 == 0`},
			{Name: "alle", Auswahl: "no replace", Filter: `card.name != ""`},
			{
				Name: "booster", Auswahl: "all", Waehlbar: true,
				Posten: []msepaket.Packposten{
					{Name: "gerade", Anzahl: 5, Gewicht: 1},
					{Name: "alle", Anzahl: 5, Gewicht: 1},
				},
			},
		},
	}

	karten := make([]*domain.CardRow, 0, 200)
	for nummer := range 200 {
		karten = append(karten, &domain.CardRow{
			ID: int64(nummer + 1), SetID: 1,
			Values: []domain.Value{{Field: "name", Text: strconv.Itoa(nummer)}},
		})
	}

	planer := service.NeuerPlanbauer().MitStilen(stileJeKarte{set: vorlage})
	einzeln := service.NeuerPackbereich(nil, setkarten{karten: karten}, planer)
	verteilt := service.NeuerPackbereich(nil, setkarten{karten: karten}, planer).MitArbeitern(8)

	for _, same := range []uint64{7, 8, 9} {
		eins, _, fehler := einzeln.Ziehen(t.Context(), 1, "booster", same)
		if fehler != nil {
			t.Fatalf("einzeln: %v", fehler)
		}

		acht, _, fehler := verteilt.Ziehen(t.Context(), 1, "booster", same)
		if fehler != nil {
			t.Fatalf("verteilt: %v", fehler)
		}

		if len(eins) != 10 || !reflect.DeepEqual(kennungen(eins), kennungen(acht)) {
			t.Errorf("Samen %d: einzeln %v, verteilt %v", same, kennungen(eins), kennungen(acht))
		}
	}
}

// kennungen gibt die Kennungen gezogener Karten.
func kennungen(karten []*domain.CardRow) []int64 {
	alle := make([]int64, 0, len(karten))
	for _, karte := range karten {
		alle = append(alle, karte.ID)
	}

	return alle
}
