package render_test

// **`<margin:l:r:t>` rückt einen Absatz ein**
// (#433).
//
// `paragraphs.back().margin_left/right/top` (`src/render/text/element.cpp`,
// Zeilen 192–194, Fassung f9d9356 vom 2021-07-06); angewandt beim Absatzwechsel
// in `viewer.cpp`, Zeilen 692–694: der linke Rand verschiebt die Zeile **und**
// verkürzt ihre Breite, der obere kommt auf den Absatzanfang.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

func TestEinAbsatzrandRuecktEin(t *testing.T) {
	satz := render.Satz{Groesse: 10, Ausrichtung: render.LinksOben}
	rahmen := render.Rechteck{Links: 5, Oben: 0, Breite: 100, Hoehe: 100}

	gesetzt := render.SetzeAbschnitte(satz, rahmen, []render.Abschnitt{
		{Text: "a\n"},
		{Text: "b", Margen: &render.Margen{Links: 20, Rechts: 30, Oben: 7}},
	}, render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

	if len(gesetzt.Zeilen) != 2 {
		t.Fatalf("%d Zeilen, erwartet zwei", len(gesetzt.Zeilen))
	}

	// **Die erste Zeile bleibt, wo sie war** — die Gegenprobe: ohne sie
	// bestünde der Test auch, wenn der Rand für alle gälte.
	if gesetzt.Zeilen[0].Links != 5 {
		t.Errorf("erste Zeile bei %g, erwartet 5", gesetzt.Zeilen[0].Links)
	}

	if gesetzt.Zeilen[0].Oben != 0 {
		t.Errorf("erste Zeile oben bei %g, erwartet 0", gesetzt.Zeilen[0].Oben)
	}

	// Die zweite: 5 + 20 links, und 12 (Zeilenhöhe) + 7 oben.
	if gesetzt.Zeilen[1].Links != 25 {
		t.Errorf("zweite Zeile bei %g, erwartet 25", gesetzt.Zeilen[1].Links)
	}

	if gesetzt.Zeilen[1].Oben != 19 {
		t.Errorf("zweite Zeile oben bei %g, erwartet 19", gesetzt.Zeilen[1].Oben)
	}
}

// **Und der Rand verkürzt die Zeile, statt sie nur zu schieben.**
//
// `line_size` beginnt bei `line.margin_left` (`viewer.cpp:548`), und die
// nutzbare Breite endet bei `width() - margin_right` — ein eingerückter Absatz
// bricht deshalb früher um.
func TestEinAbsatzrandVerkuerztDieZeile(t *testing.T) {
	satz := render.Satz{Groesse: 10, Ausrichtung: render.LinksOben}
	rahmen := render.Rechteck{Breite: 40, Hoehe: 100}

	// Vier Zeichen à 5 sind 20 breit; ohne Rand passt „ab cd" in eine Zeile
	// von 40, mit 20 links nicht mehr.
	ohne := render.SetzeAbschnitte(satz, rahmen,
		[]render.Abschnitt{{Text: "ab cd"}},
		render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

	mit := render.SetzeAbschnitte(satz, rahmen,
		[]render.Abschnitt{{Text: "ab cd", Margen: &render.Margen{Links: 20}}},
		render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

	if len(ohne.Zeilen) != 1 {
		t.Errorf("ohne Rand %d Zeilen, erwartet eine", len(ohne.Zeilen))
	}

	if len(mit.Zeilen) != 2 {
		t.Errorf("mit Rand %d Zeilen, erwartet zwei", len(mit.Zeilen))
	}
}

// **Der obere Rand des ersten Absatzes gilt nicht.**
//
// MSE legt die erste Zeile auf `style.padding_top` und gibt ihr nur den
// linken und rechten Rand des ersten Absatzes mit; `margin_top` kommt erst
// beim **Wechsel** zum nächsten Absatz auf `line.top`
// (`TextViewer::prepareLinesAtScale`, `src/render/text/viewer.cpp`, Zeilen
// 566–569 und 694 — die einzige Stelle, die `margin_top` liest; Fassung
// f9d9356 vom 2021-07-06). Hier stand bis zum 2026-10-11, MSE lege ihn auch
// beim Anlegen der ersten Zeile auf; Zeile 568 setzt dort `margin_right`.
func TestDerObereRandDesErstenAbsatzesGiltNicht(t *testing.T) {
	satz := render.Satz{Groesse: 10, Ausrichtung: render.LinksOben}
	rahmen := render.Rechteck{Breite: 100, Hoehe: 100}

	gesetzt := render.SetzeAbschnitte(satz, rahmen, []render.Abschnitt{
		{Text: "a\nb", Margen: &render.Margen{Oben: 7}},
	}, render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

	if len(gesetzt.Zeilen) != 2 {
		t.Fatalf("%d Zeilen, erwartet zwei", len(gesetzt.Zeilen))
	}

	// Die erste Zeile oben, die zweite nach 12 (Zeilenhöhe) und 7 (Rand des
	// zweiten Absatzes).
	if gesetzt.Zeilen[0].Oben != 0 || gesetzt.Zeilen[1].Oben != 19 {
		t.Errorf("Zeilen oben bei %g und %g, erwartet 0 und 19",
			gesetzt.Zeilen[0].Oben, gesetzt.Zeilen[1].Oben)
	}
}
