package render_test

// **MSE2 sucht den Maßstab eines zu langen Texts in Schritten** —
// `prepareLinesTryScales`, `src/render/text/viewer.cpp`, Zeilen 444–543,
// Fassung f9d9356 vom 2021-07-06.
//
// Es beginnt bei 1, schätzt aus der Unterkante der letzten Zeile eine untere
// und eine obere Schranke (`bound_on_min_scale`, `bound_on_max_scale`, ebd.
// Zeilen 429–444) und halbiert, bis die Schranken weniger als einen Schritt
// auseinanderliegen — `1/(4·Schriftgröße)`, mindestens 0,01
// (`FontTextElement::scaleStep`, `src/render/text/font.cpp`, Zeilen 54–56).
// **Es findet damit nicht den größten passenden Maßstab**, sondern einen, der
// höchstens einen Schritt darunter liegt. Hier stand bis dahin eine
// Halbierung bis auf 0,007 Punkte; der Text stand bei MSE2 um bis zu 1,8 %
// kleiner.
//
// **Und „passt" heißt bei MSE2**: die Unterkante der **letzten** Zeile liegt
// über dem unteren Polster (`lines.back().bottom() <= … - padding_bottom`,
// ebd. Zeile 733) — auch wenn diese Zeile leer ist, und bei einem Feld mit
// `paragraph height:` die höchste Absatzhöhe mal die Zahl der Bänder, die
// der Kasten fasst (ebd. Zeilen 718–732).

import (
	"math"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

func setzeKlein(satz render.Satz, rahmen render.Rechteck, abschnitte ...render.Abschnitt) render.Satzergebnis {
	satz.Schrift, satz.Groesse, satz.Kleinstgroesse = "Probe", 10, 5

	return render.SetzeAbschnitte(satz, rahmen, abschnitte,
		render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)
}

// Zwei Wörter, 45 Punkte breit bei Größe 10, in einem Kasten von 40 × 20:
// in einer Zeile passen sie bis zum Maßstab 40/45 = 0,889. MSE2 schätzt
// 20/24 = 0,833, prüft 0,917 (zu breit), 0,875 (passt) und 0,896 (zu breit)
// und hört auf, weil 0,875 + 0,025 nicht unter 0,896 liegt — der Text steht
// in 8,75 Punkten, nicht in 8,89.
func TestDerMassstabFolgtMSE2sSuche(t *testing.T) {
	gesetzt := setzeKlein(render.Satz{}, render.Rechteck{Breite: 40, Hoehe: 20},
		render.Abschnitt{Text: "aaaa aaaa"})

	if math.Abs(gesetzt.Groesse-8.75) > 1e-9 {
		t.Errorf("gesetzt in %g Punkten, erwartet 8,75", gesetzt.Groesse)
	}

	if len(gesetzt.Zeilen) != 1 {
		t.Errorf("%d Zeilen, erwartet eine", len(gesetzt.Zeilen))
	}
}

// **Eine leere letzte Zeile zählt mit.** Hinter `<line>` steht sie um den
// Abstand `line height line` tiefer (12 · 1,5 = 18) und ist selbst null hoch;
// in einem Kasten von 15 passt der Text bei MSE2 deshalb nicht, obwohl seine
// einzige sichtbare Zeile 12 misst.
func TestEineLeereLetzteZeileZaehltBeimPassen(t *testing.T) {
	gesetzt := setzeKlein(
		render.Satz{Zeilenfaktoren: render.Zeilenfaktoren{Weich: 1, Hart: 1, Linie: 1.5}},
		render.Rechteck{Breite: 100, Hoehe: 15},
		render.Abschnitt{Text: "aaaa"}, render.Abschnitt{Text: "\n", Blockende: true})

	if gesetzt.Groesse >= 10 || gesetzt.Groesse < 8 {
		t.Errorf("gesetzt in %g Punkten, erwartet verkleinert auf etwa 15/18 · 10", gesetzt.Groesse)
	}
}

// **Mit festen Absatzbändern zählt die höchste Absatzhöhe** — mal die Zahl
// der Bänder, die der Kasten fasst: `floor(0,5 + 40/20)` = 2, die höchste
// Höhe 12, also 24 ≤ 40. Drei einzeilige Absätze passen damit bei MSE2 in
// voller Größe, obwohl das dritte Band unter dem Kasten liegt.
func TestMitAbsatzbaendernZaehltDieHoechsteAbsatzhoehe(t *testing.T) {
	gesetzt := setzeKlein(render.Satz{Absatzhoehe: 20, Ausrichtung: render.LinksOben},
		render.Rechteck{Breite: 100, Hoehe: 40},
		render.Abschnitt{Text: "aa\nbb\ncc"})

	if gesetzt.Groesse != 10 {
		t.Errorf("gesetzt in %g Punkten, erwartet 10", gesetzt.Groesse)
	}
}

// **Die Schranken haben einen Boden**: liegt die Unterkante samt Polster unter
// einem Punkt, ist die untere Schranke 0 und die obere 1 (`if (height < 1)`,
// `viewer.cpp`, Zeilen 433 und 441). Ein Kasten von einem halben Punkt und
// eine Schrift von 0,6 erreichen beide: bei 1 (gesetzt 0,5 Punkt) ist die
// Zeile 0,6 hoch und passt nicht, bei der Schranke 0,5 (0,25 Punkt) ist sie
// 0,3 hoch und passt, und bei 0,75 (wieder 0,5 Punkt) nicht.
func TestDieSchrankenEinesWinzigenKastens(t *testing.T) {
	gesetzt := render.SetzeAbschnitte(
		render.Satz{Schrift: "Probe", Groesse: 0.6, Kleinstgroesse: 0.3},
		render.Rechteck{Breite: 100, Hoehe: 0.5},
		[]render.Abschnitt{{Text: "a"}},
		render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

	// **Beim Maßstab 0,5, gesetzt auf ¼ Punkt** (#2302): 0,3 wird zu
	// `to_int(1,2) / 4` = 0,25 — der Maßstab ist derselbe, die Größe die,
	// mit der MSE2 die Schrift setzt.
	if math.Abs(gesetzt.Groesse-0.25) > 1e-9 {
		t.Errorf("gesetzt in %g Punkten, erwartet 0,25", gesetzt.Groesse)
	}
}
