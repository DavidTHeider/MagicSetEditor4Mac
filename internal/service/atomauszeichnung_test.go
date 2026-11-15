package service_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein schlichtes `<atom…>` wirkt auf einer gedruckten Karte nicht**
// (#555).
//
// # Was MSE damit tut
//
// `TextElements::fromString` baut daraus ein `AtomTextElement` mit einer
// Hintergrundfarbe (`src/render/text/element.cpp`, Zeilen 150–160), und dessen
// `draw` malt sie **nur im Editor**:
//
//	void AtomTextElement::draw(…, DrawWhat what, …) const {
//	  if (what & DRAW_ACTIVE) {
//	    dc.SetPen(*wxTRANSPARENT_PEN);
//	    dc.SetBrush(background_color);
//	    dc.DrawRectangle(rect);
//	  }
//	  CompoundTextElement::draw(dc, scale, rect, xs, what, start, end);
//	}
//
// `src/render/text/compound.cpp`, Zeilen 57–64. `DRAW_ACTIVE` heißt
// ausdrücklich „draw active editor stuff, such as hidden separators and atom
// highlights" (`src/data/draw_what.hpp`, Zeile 23) und ist beim Ausgeben eines
// Kartenbildes nicht gesetzt. Fassung f9d9356 vom 2021-07-06.
//
// **Der Inhalt wird gezeichnet**, die Auszeichnung nicht — genau das, was
// dieses Programm tut. Der Befund dafür war also einer für richtiges
// Verhalten, und über die Stichprobe stand er **50-mal** da.
//
// # Zwei Ausnahmen bleiben Befund
//
// `<atom-param>` und `<atom-kwpph>` klammern ihren Inhalt in ⟨ ⟩ und färben
// beziehungsweise mildern ihn (`element.cpp`, Zeilen 237–241 und 298–306).
// Das ist Verhalten, das hier fehlt, und es bleibt gemeldet.
func TestEinSchlichtesAtomIstKeinBefund(t *testing.T) {
	faelle := []struct {
		name, text, erwartet string
		befund               bool
	}{
		{
			name: "atom-sep aus dem Sprachverzeichnis",
			text: "Human<atom-sep> </atom-sep>Wizard",
			// **Der Inhalt bleibt**, das Leerzeichen steht zwischen den Worten.
			erwartet: "Human Wizard",
		},
		{
			name:     "atom-autorem",
			text:     "Trample<atom-autorem> (Rem.)</atom-autorem>",
			erwartet: "Trample (Rem.)",
		},
		{
			name:     "atom-param bleibt gemeldet",
			text:     "gib <atom-param>Ziel</atom-param> zurück",
			erwartet: "gib Ziel zurück",
			befund:   true,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			stil := msepaket.Stil{
				Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
				Felder: []msepaket.Feldstil{{
					Feld: "text", Links: "0", Oben: "0", Breite: "375", Hoehe: "100",
					Schrift: msepaket.Schrift{Name: "Helvetica", Groesse: "10"},
				}},
			}

			karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
				{Field: "text", Text: fall.text},
			}}

			plan, befunde, err := service.NeuerPlanbauer().Stilplan(
				t.Context(), karte, service.Vorlage{Stil: stil})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			var gemeldet []string

			for _, befund := range befunde {
				gemeldet = append(gemeldet, befund.Meldung)
			}

			if gab := len(gemeldet) > 0; gab != fall.befund {
				t.Errorf("Befunde %v, erwartet %v: %s",
					gab, fall.befund, strings.Join(gemeldet, " | "))
			}

			for _, ebene := range plan.Ebenen {
				if ebene.Art == render.ArtText && ebene.Klartext() != fall.erwartet {
					t.Errorf("der Text lautet %q, erwartet %q",
						ebene.Klartext(), fall.erwartet)
				}
			}
		})
	}
}
