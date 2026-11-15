package service

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Ein Zug aus zwei Blöcken trägt einen Trenner in seinem Layout.**
//
// Innen geprüft, und das ist kein Ausweichen: dieses Programm schneidet noch
// nicht an `<line>`, ein gesetzter Zug hat deshalb genau **einen** Block, und
// die Liste der Trenner bliebe über den Planbauer immer leer. Der Weg vom
// gesetzten Zug in die Umgebung ist trotzdem einer, und er wird hier begangen
// (#338).
func TestEinLayoutMitZweiBloeckenTraegtSeinenTrenner(t *testing.T) {
	mass := Inhaltsmass{
		Breite: 40, Hoehe: 48, Zeilen: 3,
		Satz: render.Satzergebnis{
			Bloecke: []render.Zeilengruppe{{Oben: 0, Hoehe: 24}, {Oben: 36, Hoehe: 12}},
			Trenner: []float64{30},
		},
	}

	layout := layoutAus(mass)

	trenner, da := eintrag(layout, "separators")
	if !da {
		t.Fatal("das Layout führt keine Trenner")
	}

	if len(trenner.Eintraege) != 1 || trenner.Eintraege[0].Wert.Zahl != 30 {
		t.Errorf("die Trenner sind %+v, erwartet einer bei 30", trenner.Eintraege)
	}

	// **Und die Liste ist über ihre Stelle erreichbar** — `separators[0]` in
	// einer Vorlage.
	if trenner.Eintraege[0].Name != "0" {
		t.Errorf("der erste Trenner heißt %q, erwartet \"0\"", trenner.Eintraege[0].Name)
	}

	bloecke, _ := eintrag(layout, "blocks")
	if len(bloecke.Eintraege) != 2 {
		t.Fatalf("%d Blöcke, erwartet zwei", len(bloecke.Eintraege))
	}

	zweiter, _ := eintrag(bloecke.Eintraege[1].Wert, "middle")
	if zweiter.Zahl != 42 {
		t.Errorf("die Mitte des zweiten Blocks ist %g, erwartet 42", zweiter.Zahl)
	}
}

// eintrag sucht einen benannten Eintrag in einer Abbildung.
func eintrag(wert msescript.Wert, name string) (msescript.Wert, bool) {
	for _, glied := range wert.Eintraege {
		if glied.Name == name {
			return glied.Wert, true
		}
	}

	return msescript.Nichts, false
}
