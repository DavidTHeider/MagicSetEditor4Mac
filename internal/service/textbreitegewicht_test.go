package service_test

// **Gemessen wird im Schnitt, in dem gezeichnet wird** — F02
// (#2244).
//
// `content_width` ist bei MSE eine Rückmeldung aus dem **tatsächlich**
// gesetzten Text (`TextViewer::prepareLines`, `src/render/text/viewer.cpp`,
// Zeilen 348–389, Fassung f9d9356 vom 2021-07-06), und der setzt jedes Zeichen
// in seiner Schrift — fett, wo `weight: bold` oder `<b>` es sagen. Hier wurde
// aufrecht gemessen und fett gezeichnet.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDieInhaltsbreiteMisstDenFettenSchnitt(t *testing.T) {
	// `schriftprobe`: aufrecht 0,5 je Zeichen, fett 1,0 — bei Größe 10 also
	// 5 und 10. Die Typzeile beginnt bei 10 und endet bei `content_width`.
	faelle := []struct {
		name, gewicht, text string
		breite              float64
	}{
		{"aufrecht", "", "abcd", 10},
		{"das ganze Feld fett", "bold", "abcd", 30},
		{"eine fette Stelle", "", "<b>ab</b>cd", 20},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			stil := msepaket.Stil{
				Breite: 375, Hoehe: 523,
				Felder: []msepaket.Feldstil{
					{
						Feld: "name", Links: "10", Oben: "10", Hoehe: "23", Breite: "100",
						Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10", Gewicht: fall.gewicht},
					},
					{
						Feld: "type", Links: "10", Oben: "40", Hoehe: "23",
						Rechts:  "{ card_style.name.content_width }",
						Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
					},
				},
			}

			karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
				{Field: "name", Text: fall.text},
				{Field: "type", Text: "x"},
			}}

			plan, _, err := service.NeuerPlanbauer().
				MitSchriften(schriftprobe{breite: 0.5}).
				Stilplan(t.Context(), karte, service.Vorlage{Stil: stil})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			for _, ebene := range plan.Ebenen {
				if ebene.Art == render.ArtText && ebene.Klartext() == "x" {
					if ebene.Rahmen.Breite != fall.breite {
						t.Errorf("die Typzeile ist %.0f breit, erwartet %.0f", ebene.Rahmen.Breite, fall.breite)
					}

					return
				}
			}

			t.Fatal("der Plan führt keine Typzeile")
		})
	}
}

// **Und kursiv in der Kursivschrift** — `italic name:`, wie der Rasterer sie
// sucht.
func TestDieInhaltsbreiteMisstDieKursivschrift(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{
			{
				Feld: "name", Links: "10", Oben: "10", Hoehe: "23", Breite: "100",
				Schrift: msepaket.Schrift{Name: "Probe", KursivName: "Probe Kursiv", Groesse: "10"},
			},
			{
				Feld: "type", Links: "10", Oben: "40", Hoehe: "23",
				Rechts:  "{ card_style.name.content_width }",
				Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
			},
		},
	}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "name", Text: "<i>ab</i>cd"},
		{Field: "type", Text: "x"},
	}}

	plan, _, err := service.NeuerPlanbauer().
		MitSchriften(kursivprobe{}).
		Stilplan(t.Context(), karte, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	// Kursiv 2,0 je Zeichen, aufrecht 0,5: 2·20 + 2·5 = 50, ab 10 also 40.
	for _, ebene := range plan.Ebenen {
		if ebene.Art == render.ArtText && ebene.Klartext() == "x" {
			if ebene.Rahmen.Breite != 40 {
				t.Errorf("die Typzeile ist %.0f breit, erwartet 40", ebene.Rahmen.Breite)
			}

			return
		}
	}

	t.Fatal("der Plan führt keine Typzeile")
}

// kursivprobe gibt der Kursivschrift ein eigenes Maß.
type kursivprobe struct{}

func (kursivprobe) Suche(name string, _ bool) (render.Schrift, error) {
	if name == "Probe Kursiv" {
		return gleichbreiteSchrift{breite: 2}, nil
	}

	return gleichbreiteSchrift{breite: 0.5}, nil
}

// **Ohne `italic name:` misst die Kursivstelle im Schnitt „Italic" der
// Familie** — Befund 3 des Zweitblicks über E/F, über
// [render.Kursivschnitte]; ohne einen solchen Schnitt aufrecht.
func TestDieInhaltsbreiteMisstDenKursivenSchnittDerFamilie(t *testing.T) {
	for _, fall := range []struct {
		name   string
		quelle render.Schriftquelle
		breite float64
	}{
		// Kursiv 2,0 je Zeichen, aufrecht 0,5: 2·20 + 2·5 = 50, ab 10 also 40.
		{name: "mit kursivem Schnitt der Familie", quelle: familienkursiv{mit: true}, breite: 40},
		// Alles aufrecht: 4·5 = 20, ab 10 also 10.
		{name: "ohne kursiven Schnitt aufrecht", quelle: familienkursiv{}, breite: 10},
	} {
		t.Run(fall.name, func(t *testing.T) {
			stil := msepaket.Stil{
				Breite: 375, Hoehe: 523,
				Felder: []msepaket.Feldstil{
					{
						Feld: "name", Links: "10", Oben: "10", Hoehe: "23", Breite: "100",
						Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
					},
					{
						Feld: "type", Links: "10", Oben: "40", Hoehe: "23",
						Rechts:  "{ card_style.name.content_width }",
						Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
					},
				},
			}

			plan, _, err := service.NeuerPlanbauer().MitSchriften(fall.quelle).
				Stilplan(t.Context(), &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
					{Field: "name", Text: "<i>ab</i>cd"}, {Field: "type", Text: "x"},
				}}, service.Vorlage{Stil: stil})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			for _, ebene := range plan.Ebenen {
				if ebene.Art == render.ArtText && ebene.Klartext() == "x" {
					if ebene.Rahmen.Breite != fall.breite {
						t.Errorf("die Typzeile ist %.0f breit, erwartet %.0f", ebene.Rahmen.Breite, fall.breite)
					}

					return
				}
			}

			t.Fatal("der Plan führt keine Typzeile")
		})
	}
}

// familienkursiv kennt „Probe Italic" — oder nicht.
type familienkursiv struct{ mit bool }

func (f familienkursiv) Suche(name string, _ bool) (render.Schrift, error) {
	switch name {
	case "Probe Italic":
		if f.mit {
			return gleichbreiteSchrift{breite: 2}, nil
		}

		return nil, domain.ErrNotFound
	case "Probe Bold Italic":
		return nil, domain.ErrNotFound
	}

	return gleichbreiteSchrift{breite: 0.5}, nil
}
