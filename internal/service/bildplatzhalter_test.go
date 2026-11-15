package service_test

// **Ein Bildfeld ohne Bild zeigt das Schachbrett** — F02
// (#2244).
//
//	// checkerboard placeholder
//	if (!image.Ok() && !bitmap.Ok() && style().width > 40) {
//	  bitmap = imagePlaceholder(dc, w, h, wxNullImage, what & DRAW_EDITING);
//
// `ImageValueViewer::draw`, `src/render/value/image.cpp`, Zeilen 54–61,
// Fassung f9d9356 vom 2021-07-06 — **auch beim Export**: nur der Satz „double
// click to load image" hängt am Editor. Hier blieb das Fenster leer, und bei
// einem Rahmen aus JPG — `magic-nouveau`, `magic-planeshifted` — stand dort
// Schwarz, wo MSE2 ein weiß-graues Schachbrett zeigt.
//
// **Nur an einem Bildfeld der Karte** (`ImageValueViewer`), nicht an einer
// Wahl, die ein Bild nennt — die schweigt ohne Bild (#506).

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinLeeresBildfeldZeigtDasSchachbrett(t *testing.T) {
	faelle := []struct {
		name, wert, breite, ersatz string
		knoten                     string
		befunde                    int
	}{
		{"ohne Bild", "", "300", "", "platzhalter", 0},
		{"breiter als 40", "", "41", "", "platzhalter", 0},
		{"40 breit und nicht mehr", "", "40", "", "", 0},
		{"mit Ersatzbild", "", "300", "ersatz.png", "datei", 0},
		{"ein Kartenbild ohne Set-Paket", "fehlt.png", "300", "", "platzhalter", 1},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			stil := msepaket.Stil{
				Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
				Felder: []msepaket.Feldstil{{
					Feld: "image", Links: "30", Oben: "60", Breite: fall.breite, Hoehe: "200",
					Maske: "maske.png", Standardbild: fall.ersatz,
				}},
			}

			karte := &domain.CardRow{ID: 1, Number: 1}
			if fall.wert != "" {
				karte.Values = []domain.Value{{Field: "image", Text: fall.wert}}
			}

			plan, befunde, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, service.Vorlage{
				Stil: stil, Kartenfelder: []domain.Field{{Name: "image", Kind: domain.KindImage}},
			})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			if len(befunde) != fall.befunde {
				t.Errorf("Befunde: %+v, erwartet %d", befunde, fall.befunde)
			}

			knoten := ""

			for _, ebene := range plan.Ebenen {
				if ebene.Art != render.ArtBild || ebene.Feld != "image" {
					continue
				}

				knoten = ebene.Bild.Knotenart
				if !strings.Contains(ebene.Maske.Beschreibung(), "maske.png") {
					t.Errorf("die Maske des Felds fehlt: %q", ebene.Maske.Beschreibung())
				}
			}

			if knoten != fall.knoten {
				t.Errorf("die Ebene zeichnet %q, erwartet %q", knoten, fall.knoten)
			}
		})
	}
}
