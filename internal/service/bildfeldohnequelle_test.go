package service_test

// **Ein Bildfeld ohne Bildquelle zeichnet nichts — und meldet nichts**
// (#506).
//
//	void draw_choice_viewer(RotatedDC& dc, ValueViewer& viewer,
//	                        ChoiceStyle& style, const String& value) {
//	  if (value.empty()) return;
//	  …
//	  if (img.isReady()) { … }
//	  else if (viewer.nativeLook()) { margin = 18; }
//
// `src/render/value/choice.cpp`, Zeilen 62–92, Fassung f9d9356 vom
// 2021-07-06. **Ist kein Bild bereit, geschieht nichts** — kein Fehler, keine
// Meldung.
//
// # Woran es aufgefallen ist
//
// Der Lauf über alle 23.168 kanonischen Karten meldete **20-mal** „dieses Feld
// zeichnet ein Bild, nennt aber weder `image:` noch `default:` noch Wahlbilder".
// Über das volle Pack sind es **sechs** Feldstile, und alle sechs heißen
// `card color`:
//
//	magic-genevensis-90-counter, magic-hecatomb, magic-m15-flip,
//	magic-m15-full-art-basic-land-symbol,
//	magic-m15-planeswalker-promo-black, magic-starwars
//
// Sie decken die ganze Karte bei `z index: 0` — der Rahmen kommt bei diesen
// Vorlagen aus einem eigenen `image`-Feld, und `card color` ist ein
// Platzhalter. Dieselbe Lage wie bei `render style: hidden` und einer Breite
// von null: die Vorlage sagt „hier nichts", und MSE tut nichts.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestEinBildfeldOhneQuelleSchweigt(t *testing.T) {
	stil := bildstil(msepaket.Feldstil{
		Feld:        "card color",
		Darstellung: "image",
		Links:       "0", Oben: "0", Breite: "375", Hoehe: "523",
	})

	// **Mit Kartenwert**, sonst fällt das Feld schon vorher still weg: ein
	// Feld, das die Karte nicht führt, zeichnet ohnehin nichts.
	plan, befunde := gedeutetMit(t,
		&domain.CardRow{Values: []domain.Value{{Field: "card color", Text: "black"}}},
		stil)

	if len(befunde) != 0 {
		t.Errorf("Befunde: %v — erwartet keine", befunde)
	}

	// **Und keine Ebene**: gezeichnet wird nichts.
	if len(plan.Ebenen) != 1 {
		t.Errorf("%d Ebenen, erwartet nur den Grund", len(plan.Ebenen))
	}
}

// **Ein Feld, das weder Schrift noch Bild nennt, meldet weiter.**
//
// Das ist die andere Hälfte von `ungedeutet` und bleibt: ein Feldstil ohne
// jede Angabe ist kein Platzhalter, sondern eine Lücke.
func TestEinFeldOhneSchriftUndBildMeldetWeiter(t *testing.T) {
	stil := bildstil(msepaket.Feldstil{
		Feld:  "irgendwas",
		Links: "0", Oben: "0", Breite: "10", Hoehe: "10",
	})

	_, befunde := gedeutetMit(t,
		&domain.CardRow{Values: []domain.Value{{Field: "irgendwas", Text: "da"}}},
		stil)

	if len(befunde) != 1 {
		t.Errorf("%d Befunde, erwartet einer: %v", len(befunde), befunde)
	}
}
