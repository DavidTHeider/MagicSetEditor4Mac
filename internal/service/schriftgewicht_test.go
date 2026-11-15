package service_test

// **`weight: bold` setzt das ganze Feld fett, `style: italic` das ganze Feld
// kursiv** — F02 (#2244).
//
// MSE2 macht aus beiden Angaben Schalter der **Schrift** —
//
//	flags = (flags & ~FONT_BOLD & ~FONT_ITALIC)
//	      | (weight() == _("bold")   ? FONT_BOLD   : FONT_NORMAL)
//	      | (style()  == _("italic") ? FONT_ITALIC : FONT_NORMAL);
//
// (`Font::update`, `src/data/font.cpp`, Zeilen 26–41, Fassung f9d9356 vom
// 2021-07-06) —, und `<b>` und `<i>` legen ihre Schalter darauf
// (`Font::make(add_flags, …)`, `f->flags |= add_flags`, ebd. Zeilen 54–56).
// Ein Abschnitt ist also fett, wenn die Schrift **oder** die Auszeichnung es
// sagt.
//
// **Verglichen wird genau**: `weight() == "bold"`, nicht ohne Rücksicht auf
// Groß- und Kleinschreibung. Im Pack steht überall `bold`.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestGewichtUndLageDerSchriftGeltenFuerJedenAbschnitt(t *testing.T) {
	faelle := []struct {
		name, gewicht, lage string
		fett, kursiv        []bool
	}{
		{"ohne Angabe", "", "", []bool{false, true}, []bool{false, false}},
		{"fett", "bold", "", []bool{true, true}, []bool{false, false}},
		{"kursiv", "", "italic", []bool{false, true}, []bool{true, true}},
		{"fett als Skript", `{if true then "bold" else ""}`, "", []bool{true, true}, []bool{false, false}},
		{"nicht fett als Skript", `{if false then "bold" else ""}`, "", []bool{false, true}, []bool{false, false}},
		{"ein Skript ohne else", `{if false then "bold"}`, "", []bool{false, true}, []bool{false, false}},
		{"genau verglichen", "Bold", "Italic", []bool{false, true}, []bool{false, false}},
		{"ein anderes Gewicht", "normal", "", []bool{false, true}, []bool{false, false}},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			ebene, befunde := textebeneMitSchrift(t, msepaket.Schrift{
				Name: "Matrix", Groesse: "12", Gewicht: fall.gewicht, Lage: fall.lage,
			}, "Name <b>fett</b>")
			if len(befunde) > 0 {
				t.Fatalf("Befunde: %+v", befunde)
			}

			if len(ebene.Abschnitte) != len(fall.fett) {
				t.Fatalf("%d Abschnitte, erwartet %d: %+v", len(ebene.Abschnitte), len(fall.fett), ebene.Abschnitte)
			}

			for i, abschnitt := range ebene.Abschnitte {
				if abschnitt.Fett != fall.fett[i] || abschnitt.Kursiv != fall.kursiv[i] {
					t.Errorf("Abschnitt %d %q: fett %v, kursiv %v — erwartet %v und %v",
						i, abschnitt.Text, abschnitt.Fett, abschnitt.Kursiv, fall.fett[i], fall.kursiv[i])
				}
			}
		})
	}
}

// **Ein Gewicht, das sich nicht ausrechnen lässt, ist ein Befund** — und das
// Feld bleibt aufrecht, statt still fett zu werden.
func TestEinGewichtOhneErgebnisIstEinBefund(t *testing.T) {
	ebene, befunde := textebeneMitSchrift(t, msepaket.Schrift{
		Name: "Matrix", Groesse: "12", Gewicht: "{gibt_es_nicht()}",
	}, "Name")

	if len(befunde) != 1 || !strings.Contains(befunde[0].Ort, "font weight") {
		t.Fatalf("Befunde: %+v, erwartet einer an »font weight«", befunde)
	}

	if ebene.Abschnitte[0].Fett {
		t.Error("das Feld ist fett, obwohl sein Gewicht nicht auszurechnen war")
	}
}

// textebeneMitSchrift baut den Plan über ein einziges Textfeld mit dieser Schrift.
func textebeneMitSchrift(t *testing.T, schrift msepaket.Schrift, text string) (render.Ebene, []domain.Befund) {
	t.Helper()

	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld: "name", Links: "10", Oben: "10", Breite: "300", Hoehe: "40", Schrift: schrift,
		}},
	}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{{Field: "name", Text: text}}}

	plan, befunde, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art == render.ArtText {
			return ebene, befunde
		}
	}

	t.Fatal("der Plan führt keine Textebene")

	return render.Ebene{}, nil
}
