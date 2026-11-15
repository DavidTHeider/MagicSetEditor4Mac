package service_test

import (
	"errors"
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Maske eines Textfelds reist zum Satz** — gezeichnet wie gemessen
// (#2299).
//
// MSE bricht jede Zeile an der Breite um, die die Maske an ihrer Oberkante
// lässt (`TextViewer::lineLeft`/`lineRight`, `src/render/text/viewer.cpp`,
// Zeilen 736–741, Fassung f9d9356 vom 2021-07-06). Das gilt für den Satz,
// den es zeichnet, und damit auch für `content_lines`, das die Vorlagen aus
// ihm lesen. Ohne die Maske im Messsatz stünden die Treuekästen eines
// Planeswalkers neben anderen Zeilen als der Text.

// maskenbilder gibt unter `/probe.mse-style/maske.png` ein Bild von zehn mal
// sechs Punkten: oben ganz weiß, ab der vierten Zeile nur die linke Hälfte.
type maskenbilder struct{}

func (maskenbilder) Bild(pfad string) (image.Image, error) {
	if pfad != "/probe.mse-style/maske.png" {
		return nil, errors.New("kein Bild unter " + pfad)
	}

	bild := image.NewNRGBA(image.Rect(0, 0, 10, 6))
	for y := range 6 {
		for x := range 10 {
			farbe := color.NRGBA{A: 255}
			if y < 3 || x < 5 {
				farbe = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
			}

			bild.SetNRGBA(x, y, farbe)
		}
	}

	return bild, nil
}

func TestDieMaskeEinesTextfeldsReistZumSatz(t *testing.T) {
	// Neunzehn Wörter zu 20 Punkten: ohne Maske vier je Zeile (95 von 100),
	// fünf Zeilen. Mit ihr — auf 100 mal 60 gebracht — ebenfalls vier, solange
	// die Zeile oberhalb von 30 beginnt (99), darunter zwei (49): Zeilen bei
	// 0, 12, 24, 36, 48 und 60, die letzte unter dem Feld wieder breit.
	faelle := []struct {
		name           string
		maske          string
		zeilen         float64
		erwarteterPfad string
	}{
		{name: "ohne Maske", maske: "", zeilen: 5, erwarteterPfad: ""},
		{name: "mit Maske", maske: "maske.png", zeilen: 6, erwarteterPfad: "/probe.mse-style/maske.png"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			plan := maskenplan(t, fall.maske)

			var text, probe *render.Ebene

			for nummer := range plan.Ebenen {
				ebene := &plan.Ebenen[nummer]
				if ebene.Art != render.ArtText {
					continue
				}

				if ebene.Rahmen.Hoehe == 12 {
					probe = ebene
				} else {
					text = ebene
				}
			}

			if text == nil || probe == nil {
				t.Fatalf("der Plan führt nicht beide Textfelder: %v", plan.Ebenen)
			}

			if text.Satz.Zeilenmaske != fall.erwarteterPfad {
				t.Errorf("der Satz trägt die Maske %q, erwartet %q",
					text.Satz.Zeilenmaske, fall.erwarteterPfad)
			}

			// Das Feld `probe` legt seine linke Kante auf `content_lines`.
			if probe.Rahmen.Links != fall.zeilen {
				t.Errorf("content_lines ist %g, erwartet %g", probe.Rahmen.Links, fall.zeilen)
			}
		})
	}
}

func maskenplan(t *testing.T, maske string) *render.Plan {
	t.Helper()

	stil := msepaket.Stil{
		Paket: "probe.mse-style", Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{
			{
				Feld: "text", Links: "0", Oben: "100", Breite: "100", Hoehe: "60",
				Maske:   maske,
				Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
			},
			{
				Feld: "probe", Links: "{ card_style.text.content_lines }",
				Oben: "0", Breite: "50", Hoehe: "12",
				Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
			},
		},
	}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "text", Text: strings.TrimSpace(strings.Repeat("aaaa ", 19))},
		{Field: "probe", Text: "x"},
	}}

	plan, _, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		MitBildern(maskenbilder{}).
		Stilplan(t.Context(), karte, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	return plan
}
