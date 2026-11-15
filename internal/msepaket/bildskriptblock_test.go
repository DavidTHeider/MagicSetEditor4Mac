package msepaket_test

// **`image:` darf seinen Ausdruck als `script:`-Unterblock schreiben**
// (#454).
//
// So schreibt es `magic-modules.mse-include/stamps/card_fields` für das Feld
// `card stamp`. Über `Full-Magic-Pack` steht der Unterblock **56-mal** an
// `image:` und zweimal an `mask:`; gezählt an der Angabe (§10).

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestEinBildDarfEinenSkriptblockTragen(t *testing.T) {
	pfad := probestilpaket(t, "mse version: 2.0.0\n"+
		"card style:\n"+
		"\tstempel:\n"+
		"\t\tleft: 166\n"+
		"\t\ttop: 473\n"+
		"\t\twidth: 43\n"+
		"\t\theight: 43\n"+
		"\t\timage:\n"+
		"\t\t\tscript:\n"+
		"\t\t\t\tcard_stamp_image(field: 1)\n"+
		"\t\tmask:\n"+
		"\t\t\tscript:\n"+
		"\t\t\t\t\"maske.png\"\n")

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	stil, _ := paket.Stilmodell()

	if len(stil.Felder) != 1 {
		t.Fatalf("%d Feldstile, erwartet einer", len(stil.Felder))
	}

	feld := stil.Felder[0]

	if !strings.Contains(feld.Bild, "card_stamp_image") {
		t.Errorf("Bild %q", feld.Bild)
	}

	if !strings.Contains(feld.Maske, "maske.png") {
		t.Errorf("Maske %q", feld.Maske)
	}
}
