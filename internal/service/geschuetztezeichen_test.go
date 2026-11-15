package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Zwei Steuerzeichen in einem Kartenwert sind keine Steuerzeichen** —
// #687.
//
//	/// Escaped '<'
//	const Char ESCAPED_LANGLE = _('\1');
//	#define CONNECTION_SPACE _('\xEB00') // in private use area, untags to ' '
//
//	wxUniChar untag_char(wxUniChar c) {
//	  if (c == ESCAPED_LANGLE) return _('<');
//	  else if (c == CONNECTION_SPACE) return _(' ');
//	  else return c;
//	}
//
// `src/util/tagged_string.hpp` Zeile 20, `src/util/string.hpp` Zeile 85,
// `src/util/tagged_string.cpp` Zeilen 14–18, Fassung f9d9356 vom 2021-07-06.
//
// **Wer sie zeichnet, zeichnet ein unsichtbares Zeichen** an der Stelle eines
// gewöhnlichen. Ein getipptes `<` steht so in der Datei — sonst begänne beim
// nächsten Lesen eine Marke —, und ein aufgelöstes Schlüsselwort setzt das
// verbindende Leerzeichen.
func TestGeschuetzteZeichenWerdenGezeichnet(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "text", Links: "0", Oben: "0", Breite: "375", Hoehe: "100",
			Schrift: msepaket.Schrift{Name: "Helvetica", Groesse: "10"},
		}},
	}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{{
		Field: "text", Text: "Stärke 3 \u0001 4 und ein\ueb00Wort",
	}}}

	plan, _, err := service.NeuerPlanbauer().Stilplan(
		t.Context(), karte, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art != render.ArtText {
			continue
		}

		if ebene.Klartext() != "Stärke 3 < 4 und ein Wort" {
			t.Errorf("der Text lautet %q, erwartet die geschützten Zeichen aufgelöst",
				ebene.Klartext())
		}

		return
	}

	t.Fatal("keine Textebene")
}
