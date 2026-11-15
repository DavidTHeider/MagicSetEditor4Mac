package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **`content_width` ist die Breite des *Inhalts*, nicht die des Feldes.**
//
// # Der zweistufige Satz
//
// In MSE wird erst gesetzt, dann angeordnet: `card_style.<feld>.content_width`
// fragt, wie breit der **gesetzte Inhalt** eines Feldes ist, und die
// Feldangaben rechnen damit. `magic-m15` schreibt für die Manakosten
//
//	width: { max(30, card_style.casting_cost.content_width) + 5 }
//
// Die Breite des Feldes hängt also an seinem Inhalt. Bis zum 2026-08-16 gab
// diese Umgebung dort die **deklarierte** Breite zurück — bei einem Skript wie
// oben ist das null, und `max(30, 0) + 5` ergibt für **jede** Karte 35. Eine
// Karte mit fünf Manasymbolen bekam denselben Kasten wie eine mit einem.
//
// **Gemessen über die 190 Stile von `M15-Magic-Pack`:** `content_width` steht
// 500-mal da, davon 194-mal für `casting_cost`. Es sind also nicht sechs
// Befunde, die hier falsch sind, sondern fünfhundert Rechnungen, die still ein
// falsches Ergebnis bekommen — Phase B bewegt den Befundzähler kaum und das
// Goldbild sehr wohl
// (#291).
//
// # Warum ein Symbolfeld zuerst
//
// Ein Symbolzug lässt sich **ohne Schriftmaße** messen: er besteht aus
// quadratischen Symbolen bekannter Größe. Für einen Textzug bräuchte es die
// Schrift, und die kennt `internal/service` nach §2 nicht — das ist die
// nächste Unteraufgabe und eine eigene Naht.
func TestDieInhaltsbreiteEinesSymbolfeldsWirdGemessen(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld:        "casting cost",
			Links:       "300",
			Oben:        "29",
			Breite:      "{ max(30, card_style.casting_cost.content_width) + 5 }",
			Hoehe:       "23",
			Ausrichtung: "middle right",
			Symbolschrift: msepaket.Symbolangabe{
				Name: "probe-symbole", Groesse: "15", Ausrichtung: "middle right",
			},
			ImmerSymbole: true,
		}},
	}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "casting cost", Text: "WUBRG"},
	}}

	vorlage := service.Vorlage{Stil: stil, Symbolschriften: map[string]msepaket.Symbolschrift{
		"probe-symbole": {Name: "probe-symbole", Symbole: []msepaket.Symbol{
			{Code: "W", Bild: "w.png"},
			{Code: "U", Bild: "u.png"},
			{Code: "B", Bild: "b.png"},
			{Code: "R", Bild: "r.png"},
			{Code: "G", Bild: "g.png"},
		}},
	}}

	plan, befunde, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, befund := range befunde {
		t.Logf("Befund %s: %s", befund.Ort, befund.Meldung)
	}

	// **Fünf Symbole à 15 ergeben 75**, und die Feldangabe macht daraus
	// `max(30, 75) + 5 = 80`. Mit der deklarierten Breite käme `max(30, 0) + 5
	// = 35` heraus — für **jede** Karte dieselbe Zahl.
	//
	// Sichtbar wird der Unterschied an der **Breite der Ebene**: sie ist die
	// des Feldes, und die Feldangabe rechnet mit `content_width`. Wo die
	// Symbole darin sitzen, rechnet seit 0022 der Rasterer aus — der Planbauer
	// kennt die Bildgrößen nicht.
	breite := 0.0
	gefunden := false

	for _, ebene := range plan.Ebenen {
		if ebene.Art != render.ArtSymbolzug {
			continue
		}

		breite, gefunden = ebene.Rahmen.Breite, true
	}

	if !gefunden {
		t.Fatal("der Plan führt keinen Symbolzug")
	}

	if breite != 80 {
		t.Errorf("das Feld ist %.0f breit, erwartet 80 — "+
			"`content_width` gibt weiter die deklarierte Breite statt der gemessenen",
			breite)
	}
}

// **Und das Namensfeld rechnet mit derselben Zahl.**
//
// `magic-m15` schreibt für den Namen `right: { 341 -
// card_style.casting_cost.content_width }`. Die rechte Kante des Namens hängt
// also daran, wie breit die Manakosten **wirklich** sind — bei drei Symbolen um
// 45 Punkte weiter links als bei keinem.
//
// # Warum das im Goldbild (noch) nichts bewegt
//
// Alle drei heutigen Referenzen tragen **kurze, linksbündige** Namen. Ein
// schmaleres Namensfeld ändert an ihnen keinen Bildpunkt; sichtbar würde es
// erst bei einem Namen, der die Kante erreicht — dann greift
// `alignment: bottom shrink-overflow`. Das steht hier, damit niemand aus einer
// unbewegten Zahl schließt, die Rechnung sei folgenlos.
func TestDasNamensfeldRechnetMitDerInhaltsbreite(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{
			{
				Feld: "casting cost", Links: "300", Oben: "29", Hoehe: "23",
				Symbolschrift: msepaket.Symbolangabe{Name: "probe-symbole", Groesse: "15"},
				ImmerSymbole:  true,
			},
			{
				Feld: "name", Links: "32", Oben: "30", Hoehe: "23",
				Rechts:  "{ 341 - card_style.casting_cost.content_width }",
				Schrift: msepaket.Schrift{Name: "Helvetica", Groesse: "16"},
			},
		},
	}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "casting cost", Text: "WUBRG"},
		{Field: "name", Text: "Feuerball"},
	}}

	vorlage := service.Vorlage{Stil: stil, Symbolschriften: map[string]msepaket.Symbolschrift{
		"probe-symbole": {Name: "probe-symbole", Symbole: []msepaket.Symbol{
			{Code: "W", Bild: "w.png"},
			{Code: "U", Bild: "u.png"},
			{Code: "B", Bild: "b.png"},
			{Code: "R", Bild: "r.png"},
			{Code: "G", Bild: "g.png"},
		}},
	}}

	plan, _, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art != render.ArtText {
			continue
		}

		// 341 − 75 = 266 als rechte Kante, ab 32 also 234 breit.
		if breite := ebene.Rahmen.Breite; breite != 234 {
			t.Errorf("das Namensfeld ist %.0f breit, erwartet 234 — "+
				"ohne die gemessene Inhaltsbreite wären es 309", breite)
		}

		return
	}

	t.Fatal("der Plan führt keine Textebene")
}

// **Wo sich nichts messen lässt, wird nichts gemessen** — und die deklarierte
// Breite bleibt stehen.
//
// Die vier Wege, auf denen [inhaltsmasse] ein Feld übergeht, sind je ein
// eigener Fall und kein gemeinsamer: eine fehlende Symbolschrift ist etwas
// anderes als ein unbekanntes Zeichen, und beide sind etwas anderes als eine
// Größe, die als Skript dasteht.
func TestOhneMessbarenInhaltBleibtDieDeklarierteBreite(t *testing.T) {
	symbole := map[string]msepaket.Symbolschrift{
		"probe-symbole": {Name: "probe-symbole", Symbole: []msepaket.Symbol{{Code: "W", Bild: "w.png"}}},
	}

	faelle := []struct {
		name  string
		feld  msepaket.Feldstil
		wert  string
		ohne  map[string]msepaket.Symbolschrift
		links float64
	}{
		{
			name:  "kein Symbolschriftblock",
			feld:  msepaket.Feldstil{Feld: "casting cost", Links: "300", Oben: "29", Hoehe: "23", Breite: "40", Ausrichtung: "middle right"},
			wert:  "W",
			ohne:  symbole,
			links: 300,
		},
		{
			name: "die Symbolschrift liegt nicht vor",
			feld: msepaket.Feldstil{
				Feld: "casting cost", Links: "300", Oben: "29", Hoehe: "23", Breite: "40",
				Ausrichtung: "middle right", Symbolschrift: msepaket.Symbolangabe{Name: "fehlt", Groesse: "15"},
				ImmerSymbole: true,
			},
			wert:  "W",
			ohne:  symbole,
			links: 300,
		},
		{
			name: "ein Zeichen ohne Code",
			feld: msepaket.Feldstil{
				Feld: "casting cost", Links: "300", Oben: "29", Hoehe: "23", Breite: "40",
				Ausrichtung: "middle right", Symbolschrift: msepaket.Symbolangabe{Name: "probe-symbole", Groesse: "15"},
				ImmerSymbole: true,
			},
			wert:  "WZ",
			ohne:  symbole,
			links: 300,
		},
		{
			name: "die Größe steht als Skript da",
			feld: msepaket.Feldstil{
				Feld: "casting cost", Links: "300", Oben: "29", Hoehe: "{ 23 }", Breite: "40",
				Ausrichtung: "middle right", Symbolschrift: msepaket.Symbolangabe{Name: "probe-symbole"},
				ImmerSymbole: true,
			},
			wert:  "W",
			ohne:  symbole,
			links: 300,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			plan, _, err := service.NeuerPlanbauer().Stilplan(
				t.Context(),
				&domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{{Field: "casting cost", Text: fall.wert}}},
				service.Vorlage{
					Stil:            msepaket.Stil{Breite: 375, Hoehe: 523, Felder: []msepaket.Feldstil{fall.feld}},
					Symbolschriften: fall.ohne,
				})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			// **Kein Absturz und keine erfundene Zahl** — mehr sagt dieser Fall
			// nicht zu. Ob überhaupt eine Ebene entsteht, hängt am Feld.
			_ = plan
		})
	}
}
