package render_test

import (
	"math"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Ein zu kurzer Zug wird auf die Kastenhöhe gedehnt.**
//
// MSE verteilt den freien Platz auf die Abstände **zwischen** den Zeilen, und
// zwar bis zu `line height … max` je Umbruchart (`alignParagraph`,
// `src/render/text/viewer.cpp`, Zeilen 780–815, Fassung f9d9356 vom
// 2021-07-06):
//
//	d_hard = max(0, line_height_hard_max - line_height_hard);
//	…
//	to_add = min(stop, (s.height - height) / sum);
//	l.top += add;  if (matching) add += to_add * l.line_height;
//
// `magic-m15` schreibt für den Regeltext `hard: 1,2 / max 1,3` und
// `line: 1,5 / max 1,6` — der Text darf sich also um ein Zehntel je Absatz
// auseinanderziehen, wenn der Kasten Platz lässt
// (#345).
//
// **Die letzte Zeile zählt nicht mit.** Ihr `break_after` ist bei MSE
// `LineBreak::NO` (`Line()`, Zeile 29) — hinter ihr kommt nichts, was Abstand
// bräuchte.
func TestEinZuKurzerZugWirdGedehnt(t *testing.T) {
	// Drei Zeilen à 12, zwei harte Umbrüche; der Kasten ist 100 hoch.
	gesetzt := render.SetzeAbschnitte(
		render.Satz{
			Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben,
			Zeilenfaktoren: render.Zeilenfaktoren{Hart: 1, HartMax: 1.5},
		},
		render.Rechteck{Breite: 200, Hoehe: 100},
		[]render.Abschnitt{{Text: "aaa\nbbb\nccc"}},
		render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

	if len(gesetzt.Zeilen) != 3 {
		t.Fatalf("%d Zeilen, erwartet drei", len(gesetzt.Zeilen))
	}

	// d_hard = 0,5; to_add = min(0,5; (100−36)/24) = 0,5.
	// Zeile 0 bleibt, Zeile 1 rückt um 0,5 × 12 = 6, Zeile 2 um 12.
	for stelle, soll := range []float64{0, 18, 36} {
		if ist := gesetzt.Zeilen[stelle].Oben; ist != soll {
			t.Errorf("Zeile %d sitzt bei %g, erwartet %g", stelle, ist, soll)
		}
	}

	if gesetzt.Hoehe != 48 {
		t.Errorf("der Zug ist %g hoch, erwartet 48", gesetzt.Hoehe)
	}
}

// **Gedehnt wird nur, so weit der Kasten Platz lässt.**
//
//	to_add = min(stop, (s.height - height) / sum);
func TestGedehntWirdNurSoWeitDerKastenReicht(t *testing.T) {
	// Drei Zeilen à 12 = 36; der Kasten ist 42 hoch, es sind also 6 zu
	// verteilen: to_add = min(0,5; 6/24) = 0,25.
	gesetzt := render.SetzeAbschnitte(
		render.Satz{
			Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben,
			Zeilenfaktoren: render.Zeilenfaktoren{Hart: 1, HartMax: 1.5},
		},
		render.Rechteck{Breite: 200, Hoehe: 42},
		[]render.Abschnitt{{Text: "aaa\nbbb\nccc"}},
		render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

	for stelle, soll := range []float64{0, 15, 30} {
		if ist := gesetzt.Zeilen[stelle].Oben; ist != soll {
			t.Errorf("Zeile %d sitzt bei %g, erwartet %g", stelle, ist, soll)
		}
	}
}

// **Ein Zug, der schon passt, wird nicht gedehnt.**
//
// `if (height < s.height)` — und ohne `… max` ist der Abstand null.
func TestOhneHoechstwertWirdNichtGedehnt(t *testing.T) {
	gesetzt := render.SetzeAbschnitte(
		render.Satz{
			Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben,
			Zeilenfaktoren: render.Zeilenfaktoren{Hart: 1},
		},
		render.Rechteck{Breite: 200, Hoehe: 100},
		[]render.Abschnitt{{Text: "aaa\nbbb"}},
		render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

	if gesetzt.Zeilen[1].Oben != 12 || gesetzt.Hoehe != 24 {
		t.Errorf("gesetzt wurde %+v bei Höhe %g", gesetzt.Zeilen, gesetzt.Hoehe)
	}
}

// **Und der gedehnte Zug wird danach ausgerichtet.**
//
// MSE dehnt zuerst und rechnet den senkrechten Versatz aus der **gedehnten**
// Höhe (`align_delta_y(alignment, s.height, height)`, Zeile 819).
func TestDerGedehnteZugWirdDanachAusgerichtet(t *testing.T) {
	gesetzt := render.SetzeAbschnitte(
		render.Satz{
			Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksUnten,
			Zeilenfaktoren: render.Zeilenfaktoren{Hart: 1, HartMax: 1.5},
		},
		render.Rechteck{Breite: 200, Hoehe: 100},
		[]render.Abschnitt{{Text: "aaa\nbbb\nccc"}},
		render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

	// Gedehnt ist der Zug 48 hoch; unten heißt 100 − 48 = 52.
	if ist := gesetzt.Zeilen[0].Oben; ist != 52 {
		t.Errorf("die erste Zeile sitzt bei %g, erwartet 52", ist)
	}
}

// **Ein Zug ohne Zeilen wird nicht gedehnt.**
//
// Ein leeres Feld kommt vor — `magic-m15` misst auch Felder, die die Karte
// nicht ausgefüllt hat. Ohne diese Frage teilte die Rechnung durch nichts.
func TestEinZugOhneZeilenWirdNichtGedehnt(t *testing.T) {
	gesetzt := render.SetzeAbschnitte(
		render.Satz{
			Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben,
			Zeilenfaktoren: render.Zeilenfaktoren{Hart: 1, HartMax: 1.5},
		},
		render.Rechteck{Breite: 200, Hoehe: 100},
		nil,
		render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

	// **Eine leere Zeile bleibt eine Zeile** — sie trennt zwei Absätze.
	if len(gesetzt.Zeilen) != 1 || gesetzt.Zeilen[0].Breite != 0 {
		t.Errorf("gesetzt wurde %+v", gesetzt.Zeilen)
	}
}

// **Wo alle drei Spielräume null sind, bleibt es beim Abstand.**
//
// Der Fall steht im Pack 22-mal weniger oft als sein Gegenteil: `soft max`
// fehlt fast überall, und dann ist `d_soft` null.
func TestOhneSpielraumBleibtEsBeimAbstand(t *testing.T) {
	gesetzt := render.SetzeAbschnitte(
		render.Satz{
			Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben,
			Zeilenfaktoren: render.Zeilenfaktoren{
				Weich: 0.9, Hart: 1.2, Linie: 1.5,
				WeichMax: 0.5, HartMax: 1.0, LinieMax: 1.0,
			},
		},
		render.Rechteck{Breite: 200, Hoehe: 300},
		[]render.Abschnitt{{Text: "aaa\nbbb"}},
		render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

	// Alle Höchstwerte liegen **unter** den Faktoren: kein Spielraum. Der
	// Abstand ist 12 × 1,2 — verglichen wird mit einer Toleranz, denn
	// 12 × 1,2 ist in Gleitkomma nicht genau 14,4.
	if abstand := gesetzt.Zeilen[1].Oben - gesetzt.Zeilen[0].Oben; math.Abs(abstand-14.4) > 1e-9 {
		t.Errorf("der Abstand ist %g, erwartet 14,4", abstand)
	}
}
