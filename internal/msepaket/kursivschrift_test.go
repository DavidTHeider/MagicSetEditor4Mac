package msepaket_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Kursiv wird nicht errechnet, sondern benannt.**
//
// # Der Befund, an der Quelle nachgelesen
//
// MSE2 schrägt keine Glyphen: `Font::toWxFont` nimmt bei gesetztem
// `FONT_ITALIC` die Schrift, deren **Namen der Stil angibt** —
//
//	} else if (flags & FONT_ITALIC && !italic_name().empty()) {
//	  font = wxFont(…, italic_name());
//
// Gelesen: `src/data/font.cpp`, Zeilen 100–101, Fassung f9d9356 vom
// 2021-07-06. `MPlantin` kursiv ist die Datei `MPlantin-Italic`, eine eigene
// Schrift mit eigenen Formen — eine geneigte Fassung derselben Glyphen sähe
// anders aus.
//
// **475 `italic name:` stehen in `M15-Magic-Pack`.** Ohne sie hat der Rasterer
// keinen Namen, unter dem er die Kursive suchen könnte, und der Flavortext
// bliebe aufrecht
// (#312).
func TestEinFeldstilNenntSeineKursivschrift(t *testing.T) {
	paket, err := msepaket.Oeffne("../../saat/probe-einzeiler.mse-style")
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	stil, befunde := paket.Stilmodell()
	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	feld, da := stil.Feld("name")
	if !da {
		t.Fatal("der Stil führt kein Feld `name`")
	}

	if feld.Schrift.KursivName != "Helvetica-Oblique" {
		t.Errorf("die Kursivschrift ist %q, erwartet Helvetica-Oblique",
			feld.Schrift.KursivName)
	}
}
