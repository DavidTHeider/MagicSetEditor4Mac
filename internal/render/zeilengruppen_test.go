package render_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Zwischen zwei Blöcken liegt ein Trenner — dort zeichnet MSE eine Linie.**
//
//	layout->separators.push_back((blocks[i]->bottom() + blocks[i+1]->top)/2);
//
// `src/render/text/viewer.cpp`, Zeilen 386–388, Fassung f9d9356 vom
// 2021-07-06.
//
// **Geprüft wird die Rechnung, nicht der Weg dorthin.** Dieses Programm
// schneidet noch nicht an `<line>`; ein gesetzter Zug hat deshalb genau einen
// Block, und ein Wächter über [render.SetzeAbschnitte] beträte diese Rechnung
// nie (§10, #338).
func TestZwischenZweiBloeckenLiegtEinTrenner(t *testing.T) {
	erst := render.Zeilengruppe{Oben: 0, Hoehe: 24}
	zweit := render.Zeilengruppe{Oben: 36, Hoehe: 12}

	trenner := render.TrennerZwischen([]render.Zeilengruppe{erst, zweit})

	if len(trenner) != 1 || trenner[0] != 30 {
		t.Errorf("die Trenner sind %v, erwartet einer bei 30", trenner)
	}

	if leer := render.TrennerZwischen([]render.Zeilengruppe{erst}); len(leer) != 0 {
		t.Errorf("ein einziger Block ergibt %v, erwartet keinen Trenner", leer)
	}
}

// **Ein `<line>` mitten in einem Absatz schneidet trotzdem am Umbruch.**
//
// MSE setzt die Auszeichnung um einen Zeilenumbruch (`<line>\n</line>`), und
// dort fällt der Schnitt. Steht sie an einem Abschnitt **ohne** eigenen
// Umbruch, endet der Block mit dem Absatz, in dem sie steht — der Schnitt
// wandert nicht in die Mitte einer Zeile
// (#338).
func TestEineBlockgrenzeOhneEigenenUmbruchSchliesstDenAbsatz(t *testing.T) {
	gesetzt := render.SetzeAbschnitte(
		render.Satz{Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben},
		render.Rechteck{Breite: 200, Hoehe: 60},
		[]render.Abschnitt{
			{Text: "aa"},
			{Text: "bb", Blockende: true},
			{Text: "\ncc"},
		},
		render.Schriftsatz{Aufrecht: gleichbreit{breite: 0.5}}, nil)

	if len(gesetzt.Zeilen) != 2 {
		t.Fatalf("%d Zeilen, erwartet zwei: %+v", len(gesetzt.Zeilen), gesetzt.Zeilen)
	}

	if len(gesetzt.Bloecke) != 2 {
		t.Errorf("%d Blöcke, erwartet zwei", len(gesetzt.Bloecke))
	}

	// **Die erste Zeile trägt beide Abschnitte** — geschnitten wird hinter ihr.
	if text := gesetzt.Zeilen[0].Text; text != "aabb" {
		t.Errorf("die erste Zeile lautet %q, erwartet %q", text, "aabb")
	}
}
