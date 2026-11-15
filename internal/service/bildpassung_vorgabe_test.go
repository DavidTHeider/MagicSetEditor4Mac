package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// wahlbildvorlage baut eine Vorlage mit **einem** Auswahlfeld, das ein Bild
// zeichnet.
func wahlbildvorlage(ausrichtung string) service.Vorlage {
	feld := flaeche("ecken")
	feld.Darstellung = "image"
	feld.Ausrichtung = ausrichtung
	feld.Wahlbilder = []msepaket.Wahlbild{{Wahl: "rund", Bild: "rund.png"}}

	return service.Vorlage{
		Stil:         bildstil(feld),
		Kartenfelder: []domain.Field{{Name: "ecken", Kind: domain.KindChoice}},
	}
}

// **Ein Auswahlbild ohne `alignment:` wird gestreckt, nicht eingepasst.**
//
// # Die Quelle
//
//	ChoiceStyle::ChoiceStyle(…)
//	  : alignment(ALIGN_STRETCH)
//
// `src/data/field/choice.cpp`, Zeile 167, Fassung f9d9356 vom 2021-07-06. Die
// **Vorgabe** eines Auswahlstils ist `stretch`; erst ein ausdrückliches
// `alignment:` ohne `stretch` macht daraus eine Einpassung
// (`get_options`, `src/render/value/choice.cpp`, Zeilen 111–124).
//
// # Was das gekostet hat
//
// Das Feld `corners` der M15-Vorlagen zeichnet ein **5 × 5** großes Bild über
// die ganze Karte und schneidet es mit einer 375 × 523 großen Maske. Eingepasst
// wird daraus ein 375 × 375 großes Bild — und die Maske wird mit ihm gestaucht.
// Sie sitzt dann 148 Zeilen zu hoch: die untere Rundung landete auf halber
// Kartenhöhe und riss zwei weiße Kerben in den Kartenrand
// (#388).
func TestEinAuswahlbildOhneAusrichtungWirdGestreckt(t *testing.T) {
	plan, befunde := kunstplan(t,
		&domain.CardRow{ID: 1, Values: []domain.Value{{Field: "ecken", Text: "rund"}}},
		wahlbildvorlage(""))

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	gezeichnet := ebenen(t, plan)
	if len(gezeichnet) != 1 {
		t.Fatalf("%d Ebenen, erwartet eine", len(gezeichnet))
	}

	if gezeichnet[0].Passung != nil {
		t.Errorf("Passung %+v, erwartet keine — MSEs Vorgabe ist stretch",
			*gezeichnet[0].Passung)
	}
}

// **Mit ausdrücklicher Ausrichtung bleibt es bei der Einpassung.**
//
// Das ist der Fall des Seltenheitssymbols: `alignment: middle right` in einem
// Kasten, der breiter als hoch ist. Gestreckt wäre daraus wieder der Balken
// quer über der Typzeile (#270).
func TestEinAuswahlbildMitAusrichtungWirdEingepasst(t *testing.T) {
	plan, _ := kunstplan(t,
		&domain.CardRow{ID: 1, Values: []domain.Value{{Field: "ecken", Text: "rund"}}},
		wahlbildvorlage("middle right"))

	gezeichnet := ebenen(t, plan)
	if len(gezeichnet) != 1 {
		t.Fatalf("%d Ebenen, erwartet eine", len(gezeichnet))
	}

	if gezeichnet[0].Passung == nil {
		t.Error("keine Passung, erwartet eine Einpassung")
	}

	if gezeichnet[0].Art != render.ArtBild {
		t.Errorf("Art %q", gezeichnet[0].Art)
	}
}
