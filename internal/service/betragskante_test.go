package service_test

// **Eine negative Länge wird zum Betrag**
// (#504).
//
//	if (width  < 0) width  = -width;
//	if (height < 0) height = -height;
//
// `Style::update`, `src/data/field.cpp`, Zeilen 185–186, Fassung f9d9356 vom
// 2021-07-06 — die beiden letzten Zeilen, nach jeder Ableitung aus `left`,
// `right`, `width` und ihren senkrechten Geschwistern.
//
// # Woran es aufgefallen ist
//
// Der Lauf über alle 23.168 kanonischen Karten meldete **20-mal** „ein
// Textrahmen von −309×23". Das Namensfeld von `magic-m15` rechnet
//
//	left:  { if card.card_symbol == "none" then 32 else 50 }
//	right: { 341 - card_style.casting_cost.content_width }
//
// — bei sehr breiten Manakosten liegt `right` links von `left`, und die
// Breite wird negativ. **MSE dreht sie um und zeichnet**; dieses Programm
// meldete einen Mangel und ließ die Ebene fallen.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

func TestEineNegativeLaengeWirdZumBetrag(t *testing.T) {
	// `right` liegt links von `left`: 50 … 341 − 600 = −259.
	stil := bildstil(msepaket.Feldstil{
		Feld:   "name",
		Links:  "50",
		Rechts: "-259",
		Oben:   "30",
		Hoehe:  "23",
		Schrift: msepaket.Schrift{
			Name: "Probe", Groesse: "16",
		},
	})

	plan, befunde := gedeutetMit(t,
		&domain.CardRow{Values: []domain.Value{{Field: "name", Text: "Weit"}}}, stil)

	for _, befund := range befunde {
		t.Logf("  Befund %s: %s", befund.Ort, befund.Meldung)
	}

	var gefunden bool

	for i := range plan.Ebenen {
		ebene := &plan.Ebenen[i]
		if ebene.Art != render.ArtText {
			continue
		}

		gefunden = true

		// **309, nicht −309.**
		if ebene.Rahmen.Breite != 309 {
			t.Errorf("die Breite ist %v, erwartet 309", ebene.Rahmen.Breite)
		}
	}

	if !gefunden {
		t.Fatal("die Namensebene steht nicht im Plan")
	}
}
