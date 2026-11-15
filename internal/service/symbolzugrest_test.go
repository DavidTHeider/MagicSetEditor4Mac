package service_test

// **Ein unbekanntes Zeichen in `<sym>` wird übersprungen, nicht verworfen**
// (#517).
//
//	void SymbolFont::split(const String& text, SplitSymbols& out) const {
//	  …
//	    // unknown code, draw single character as text
//	    //out.push_back(DrawableSymbol(text.substr(pos, 1), _(""), defaultSymbol()));
//	    pos += 1;
//	  next_symbol:;
//	}
//
// `src/data/symbol_font.cpp`, Zeilen 192–231, Fassung f9d9356 vom
// 2021-07-06. **Der `push_back` ist auskommentiert**: MSE überspringt das
// Zeichen und zeichnet die übrigen Symbole.
//
// # Woran es aufgefallen ist
//
// Der Lauf über alle 23.168 kanonischen Karten meldete dreimal „die
// Symbolschrift magic-mana-small führt kein Symbol für …; `<sym>…</sym>`
// bleibt Text" — einmal für einen bloßen Zeilenumbruch. **Der ganze Zug wurde
// dabei zu Text**, und das ist sichtbar falsch: aus Manasymbolen wurden
// Buchstaben.
//
// #508 hat
// den Weg über `always symbol:` behoben; dies ist der Weg über die
// **Auszeichnung** `<sym>`, und er ist ein anderer.

import (
	"image"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinSymbolzugUeberspringtUnbekannte(t *testing.T) {
	stil := msepaket.Stil{
		Paket: "probe.mse-style", Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld: "text", Links: "0", Oben: "0", Breite: "300", Hoehe: "60",
			Symbolschrift: msepaket.Symbolangabe{Name: "probe-symbole", Groesse: "15"},
			Schrift:       msepaket.Schrift{Name: "Probe", Groesse: "12"},
		}},
	}

	vorlage := service.Vorlage{Stil: stil, Symbolschriften: map[string]msepaket.Symbolschrift{
		"probe-symbole": {
			Name: "probe-symbole", Bildgroesse: 30, AbstandWaagerecht: 2,
			Symbole: []msepaket.Symbol{{Code: "W", Bild: "w.png"}},
		},
	}}

	// Im `<sym>`-Block steht ein Umbruch, den die Schrift nicht kennt.
	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "text", Text: "<sym>W\nW</sym>"},
	}}

	quelle := bildmasse{da: map[string]image.Point{
		"/probe-symbole.mse-symbol-font/w.png": {X: 60, Y: 30},
	}}

	plan, befunde, err := service.NeuerPlanbauer().MitBildern(quelle).
		Stilplan(t.Context(), karte, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	var symbole int

	for i := range plan.Ebenen {
		if plan.Ebenen[i].Art != render.ArtText {
			continue
		}

		for _, abschnitt := range plan.Ebenen[i].Abschnitte {
			symbole += len(abschnitt.Symbole)
		}
	}

	// **Beide bekannten Symbole stehen da**, das unbekannte Zeichen nicht.
	if symbole != 2 {
		t.Errorf("%d Symbole, erwartet 2", symbole)
	}

	// **Und der Befund bleibt** — er sagt, dass die Vorlage ein Zeichen in
	// den Symbolblock gesetzt hat, das dort nichts zu suchen hat. Er ist
	// deshalb von der Art `Vorlage` und nicht `Portierung`.
	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet einer: %v", len(befunde), befunde)
	}

	if befunde[0].Art != domain.ArtVorlage {
		t.Errorf("die Art ist %q, erwartet %q", befunde[0].Art, domain.ArtVorlage)
	}
}

// **Ohne ein einziges Symbol bleibt es Text.**
//
// Ein Symbolzug ohne Symbole zeichnete nichts, und ein leerer Kasten ist
// schlechter als der Text, den die Vorlage hineingeschrieben hat (§8).
func TestEinSymbolzugOhneJedesSymbolBleibtText(t *testing.T) {
	stil := msepaket.Stil{
		Paket: "probe.mse-style", Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld: "text", Links: "0", Oben: "0", Breite: "300", Hoehe: "60",
			Symbolschrift: msepaket.Symbolangabe{Name: "probe-symbole", Groesse: "15"},
			Schrift:       msepaket.Schrift{Name: "Probe", Groesse: "12"},
		}},
	}

	vorlage := service.Vorlage{Stil: stil, Symbolschriften: map[string]msepaket.Symbolschrift{
		"probe-symbole": {
			Name: "probe-symbole", Bildgroesse: 30,
			Symbole: []msepaket.Symbol{{Code: "W", Bild: "w.png"}},
		},
	}}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "text", Text: "<sym>xyz</sym>"},
	}}

	plan, befunde, err := service.NeuerPlanbauer().MitBildern(bildmasse{}).
		Stilplan(t.Context(), karte, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for i := range plan.Ebenen {
		if plan.Ebenen[i].Art != render.ArtText {
			continue
		}

		for _, abschnitt := range plan.Ebenen[i].Abschnitte {
			if len(abschnitt.Symbole) != 0 {
				t.Errorf("der Abschnitt trägt Symbole: %+v", abschnitt)
			}

			if abschnitt.Text != "xyz" {
				t.Errorf("der Text ist %q, erwartet xyz", abschnitt.Text)
			}
		}
	}

	if len(befunde) != 1 || befunde[0].Art != domain.ArtVorlage {
		t.Errorf("Befunde: %v", befunde)
	}
}
