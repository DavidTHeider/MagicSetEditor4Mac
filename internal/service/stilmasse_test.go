package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **`stylesheet.card_width` und `.card_height` sind die Maße der Vorlage.**
//
// # Die Maßproblematik
//
// Über die 190 Stile von `M15-Magic-Pack` scheiterten **464** Maßangaben mit
// „die Angabe ergab keine Zahl" — und der Wert dahinter war jedes Mal **leer**.
// 350 davon am Feld `corners`, 102 an `card color`.
//
// Die Spur führt über drei Ecken:
//
//	corners:  width: { face_coordinates_map(1).width }
//	magic.mse-game/script:2905  face_coordinates_map := { fc[input-1] or else […] }
//	magic.mse-game/script:2891  faces_coordinates := { [[ … width: stylesheet.card_width … ]] }
//
// **`stylesheet` trug bei uns nur die Styling-Optionen.** `card_width` stand
// nicht darin, also war die Breite nichts, also die Abbildung nichts, also die
// Maßangabe leer — und die Meldung sagte nur, dass sie leer war.
//
// Gezählt im Pack: **499 Zugriffe auf `stylesheet.card_width`, 506 auf
// `.card_height`.** Es ist damit der meistgelesene Wert überhaupt.
func TestStylesheetTraegtDieKartenmasse(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "rand", Links: "0", Oben: "0",
			Breite: "{ stylesheet.card_width }",
			Hoehe:  "{ stylesheet.card_height }",
			Bild:   "rand.png",
		}},
	}

	plan, befunde, err := service.NeuerPlanbauer().Stilplan(
		t.Context(), &domain.CardRow{ID: 1, Number: 1}, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art != render.ArtBild {
			continue
		}

		if ebene.Rahmen.Breite != 375 || ebene.Rahmen.Hoehe != 523 {
			t.Errorf("die Ebene ist %.0fx%.0f, erwartet 375x523",
				ebene.Rahmen.Breite, ebene.Rahmen.Hoehe)
		}

		return
	}

	t.Fatal("der Plan führt keine Bildebene")
}

// **Eine leere Maßangabe sagt jetzt, dass sie leer ist.**
//
// „die Angabe ergab keine Zahl: " mit nichts dahinter war die Meldung, die 464
// Befunde trugen — sie nennt weder den Wert noch seine Art, und wer sie liest,
// weiß nicht, ob die Rechnung fehlschlug oder ein Name fehlte.
func TestEineLeereMassangabeNenntIhreArt(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "rand", Links: "0", Oben: "0",
			Breite: "{ styling.gibtsnicht }", Hoehe: "10", Bild: "rand.png",
		}},
	}

	_, befunde, err := service.NeuerPlanbauer().Stilplan(
		t.Context(), &domain.CardRow{ID: 1, Number: 1}, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) == 0 {
		t.Fatal("eine leere Maßangabe ergibt keinen Befund")
	}

	// **Die Meldung nennt den Zugriff, der ins Leere ging.** Seit dem
	// verzögerten Fehlerwert steht dort nicht mehr „nichts", sondern welches
	// Glied gefehlt hat — das ist die Auskunft, die jemand braucht
	// (siehe [msescript.Fehlerwert]).
	if meldung := befunde[0].Meldung; !contains(meldung, "gibtsnicht") {
		t.Errorf("die Meldung lautet %q und nennt das fehlende Glied nicht", meldung)
	}
}

func contains(text, teil string) bool {
	for i := 0; i+len(teil) <= len(text); i++ {
		if text[i:i+len(teil)] == teil {
			return true
		}
	}

	return false
}

// **Was keine Zahl ist, wird mit Art und Wert genannt.**
//
// `nichts` bekommt ein eigenes Wort — ein Name, den es gibt, mit einem
// Schlüssel, den es nicht gibt, sieht sonst aus wie ein leerer Text.
func TestEineMassangabeNenntArtUndWert(t *testing.T) {
	faelle := map[string]string{
		"{ styling.gibtsnicht }": "gibtsnicht",
		`{ "kein Maß" }`:         "Zeichenkette",
		"{ nil }":                "nichts",
	}

	for angabe, erwartet := range faelle {
		t.Run(angabe, func(t *testing.T) {
			stil := msepaket.Stil{
				Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
				Felder: []msepaket.Feldstil{{
					Feld: "rand", Links: "0", Oben: "0",
					Breite: angabe, Hoehe: "10", Bild: "rand.png",
				}},
			}

			_, befunde, err := service.NeuerPlanbauer().Stilplan(
				t.Context(), &domain.CardRow{ID: 1, Number: 1}, service.Vorlage{Stil: stil})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			if len(befunde) == 0 {
				t.Fatal("kein Befund")
			}

			if !contains(befunde[0].Meldung, erwartet) {
				t.Errorf("die Meldung lautet %q und nennt %q nicht", befunde[0].Meldung, erwartet)
			}
		})
	}
}
