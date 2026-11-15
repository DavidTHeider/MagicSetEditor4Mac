package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// druckleserDouble kennt Drucke nach Scryfall-Kennung und nach Zeile.
type druckleserDouble struct {
	nachScryfall map[string]*domain.Druck
	nachZeile    map[int64]*domain.Druck
	fehler       error
	gefragt      int
}

func (d *druckleserDouble) NachScryfall(_ context.Context, printing string) (*domain.Druck, error) {
	d.gefragt++

	return d.nachScryfall[printing], d.fehler
}

func (d *druckleserDouble) Fuer(_ context.Context, id int64) (*domain.Druck, error) {
	d.gefragt++

	if d.fehler != nil {
		return nil, d.fehler
	}

	druck, da := d.nachZeile[id]
	if !da {
		return nil, fmt.Errorf("%w: den Druck %d gibt es nicht", domain.ErrNotFound, id)
	}

	return druck, nil
}

// vorratDouble ist der Zwischenspeicher von Scryfall.
type vorratDouble map[string]*scryfall.Karte

func (v vorratDouble) KarteImVorrat(printing string) (*scryfall.Karte, bool) {
	karte, da := v[printing]

	return karte, da
}

// **Eine Handelszeile nennt eine Karte und keine Kennung** — P10
// (#2378),
// Befund C-13: am Fenster stand „1 × counterspell-7ed".
func TestHandelsdruckeNennenKarteSetNummerUndBild(t *testing.T) {
	t.Parallel()

	leser := &druckleserDouble{
		nachScryfall: map[string]*domain.Druck{
			"bolt-m11": {ID: 7, Name: "Blitzschlag", Setcode: "m11", Nummer: "149"},
		},
		nachZeile: map[int64]*domain.Druck{
			3: {ID: 3, Name: "Eigener Wächter", Setcode: "set", Nummer: "007"},
		},
	}
	vorrat := vorratDouble{
		"bolt-m11": {
			Kennung: "bolt-m11", Name: "Lightning Bolt", Set: "m11", Nummer: "149",
			Bilder: scryfall.Bilder{Klein: "https://cards.scryfall.io/small/bolt.jpg"},
		},
		"dfc-1": {
			Kennung: "dfc-1", Name: "Vorne // Hinten", Set: "dfc", Nummer: "1",
			Seiten: []scryfall.Seite{{Bilder: scryfall.Bilder{Klein: "https://cards.scryfall.io/small/vorne.jpg"}}},
		},
	}

	positionen := []domain.Handelsposition{
		{Von: "ich", Printing: "bolt-m11", Menge: 1},
		{Von: "du", Printing: "dfc-1", Menge: 1},
		{Von: "du", Printing: "unbekannt-9", Menge: 2},
		{Von: "ich", DruckID: 3, Menge: 1},
		// **Die Druckzeile des Gegenübers ist eine Zeile seines Bestands**:
		// hier hieße die Zeile 3 „Eigener Wächter" — und das wäre eine
		// andere Karte.
		{Von: "du", DruckID: 3, Menge: 1},
		{Von: "ich", Printing: "bolt-m11", Menge: 3},
		// **Eine eigene Zeile, die es nicht mehr gibt**, bleibt ein eigener
		// Druck ohne Namen — und kein Fehler: der Tausch steht trotzdem.
		{Von: "ich", DruckID: 9, Menge: 1},
	}

	drucke, fehler := service.Handelsdrucke(t.Context(), leser, vorrat, positionen, "ich")
	if fehler != nil {
		t.Fatalf("Handelsdrucke: %v", fehler)
	}

	erwartet := []service.Handelsdruck{
		// **Der Name aus dem eigenen Bestand**, das Bild aus dem Vorrat.
		{Name: "Blitzschlag", Setcode: "m11", Nummer: "149", Bild: "https://cards.scryfall.io/small/bolt.jpg"},
		{Name: "Vorne // Hinten", Setcode: "dfc", Nummer: "1", Bild: "https://cards.scryfall.io/small/vorne.jpg"},
		{},
		{Name: "Eigener Wächter", Setcode: "set", Nummer: "007", Eigen: true},
		{Eigen: true},
		{Name: "Blitzschlag", Setcode: "m11", Nummer: "149", Bild: "https://cards.scryfall.io/small/bolt.jpg"},
		{Eigen: true},
	}

	if len(drucke) != len(erwartet) {
		t.Fatalf("%d Drucke für %d Positionen", len(drucke), len(positionen))
	}

	for stelle := range erwartet {
		if drucke[stelle] != erwartet[stelle] {
			t.Errorf("Position %d: %+v, erwartet %+v", stelle+1, drucke[stelle], erwartet[stelle])
		}
	}

	// **Je Druck einmal nachgeschlagen**: drei Kennungen und zwei Zeilen.
	if leser.gefragt != 5 {
		t.Errorf("%d Abfragen für fünf verschiedene Drucke", leser.gefragt)
	}
}

// **Ein Lesefehler ist einer** — er wird nicht zu „kein Name".
func TestHandelsdruckeGebenLesefehlerWeiter(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("die Platte ist weg")

	for name, position := range map[string]domain.Handelsposition{
		"über die Kennung": {Von: "ich", Printing: "x", Menge: 1},
		"über die Zeile":   {Von: "ich", DruckID: 4, Menge: 1},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, fehler := service.Handelsdrucke(t.Context(),
				&druckleserDouble{fehler: kaputt}, vorratDouble{},
				[]domain.Handelsposition{position}, "ich")
			if !errors.Is(fehler, kaputt) {
				t.Errorf("Fehler %v, erwartet %v", fehler, kaputt)
			}
		})
	}
}

// **Ein Druck, den nur der Vorrat kennt**, hat Name, Set und Nummer von dort.
func TestHandelsdruckeNehmenDenVorratOhneEigenenBestand(t *testing.T) {
	t.Parallel()

	drucke, fehler := service.Handelsdrucke(t.Context(), &druckleserDouble{},
		vorratDouble{"z": {Kennung: "z", Name: "Sturmrufer", Set: "zwe", Nummer: "3"}},
		[]domain.Handelsposition{{Von: "du", Printing: "z", Menge: 1}}, "ich")
	if fehler != nil {
		t.Fatalf("Handelsdrucke: %v", fehler)
	}

	if want := (service.Handelsdruck{Name: "Sturmrufer", Setcode: "zwe", Nummer: "3"}); drucke[0] != want {
		t.Errorf("%+v, erwartet %+v", drucke[0], want)
	}
}

// **Ein leerer Eintrag im eigenen Bestand überschreibt nichts** — gefunden
// von der Hüllenprobe `p10handel09` (#2378): der Preisabruf legt eine
// Druckzeile mit Kennung und ohne Namen an, und danach stand an der Zeile
// wieder „Unbekannter Druck".
func TestHandelsdruckeBehaltenDenVorratGegenEineLeereZeile(t *testing.T) {
	t.Parallel()

	leser := &druckleserDouble{nachScryfall: map[string]*domain.Druck{
		"z": {ID: 1, Name: "", Setcode: "", Nummer: "4"},
	}}

	drucke, fehler := service.Handelsdrucke(t.Context(), leser,
		vorratDouble{"z": {
			Kennung: "z", Name: "Sturmrufer", Set: "zwe", Nummer: "3",
			Bilder: scryfall.Bilder{Klein: "bild"},
		}},
		[]domain.Handelsposition{{Von: "du", Printing: "z", Menge: 1}}, "ich")
	if fehler != nil {
		t.Fatalf("Handelsdrucke: %v", fehler)
	}

	// **Was der Bestand nennt, gilt; was er leer lässt, nicht.**
	if want := (service.Handelsdruck{Name: "Sturmrufer", Setcode: "zwe", Nummer: "4", Bild: "bild"}); drucke[0] != want {
		t.Errorf("%+v, erwartet %+v", drucke[0], want)
	}
}
