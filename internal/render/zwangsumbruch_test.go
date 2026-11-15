package render_test

// **Ein Wort, das allein nicht in seine Zeile passt, wird zwischen zwei
// Zeichen gebrochen** (Befund 3 des Nachlaufs über E/F).
//
// MSE prüft den Umbruch Zeichen für Zeichen. Reicht ein Wort, das die Zeile
// beginnt, über `lineRight` hinaus, erzwingt es einen Umbruch: das Wort bis
// zum vorigen Zeichen bleibt, das überstehende Zeichen beginnt die nächste
// Zeile — außer es ist das einzige, dann bleibt es stehen
// (`TextViewer::prepareLinesAtScale`, `src/render/text/viewer.cpp`, Zeilen
// 614–633 und 651–658, Fassung f9d9356 vom 2021-07-06). `stop_if_too_long`
// ist an jedem Aufruf `false` (Zeilen 453–541); in einem **mehrzeiligen**
// Feld wird also immer gebrochen, nie aufgegeben. Ein einzeiliges bricht gar
// nicht (Zeile 618) — das hält `einzeilig_test.go`.

import (
	"image"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

func TestEinZuLangesWortWirdZwischenZeichenGebrochen(t *testing.T) {
	for _, fall := range []struct {
		name, text string
		breite     float64
		erwartet   []string
	}{
		// Je Zeichen 5 Punkte, acht passen in 40.
		{name: "allein", text: "aaaaaaaaaaaa", breite: 40, erwartet: []string{"aaaaaaaa", "aaaa"}},
		{name: "nach einem Wort", text: "ab aaaaaaaaaaaa", breite: 40, erwartet: []string{"ab", "aaaaaaaa", "aaaa"}},
		{name: "mit einem Wort danach", text: "aaaaaaaaaaaa b", breite: 40, erwartet: []string{"aaaaaaaa", "aaaa b"}},
		{name: "über drei Zeilen", text: "aaaaaaaaaaaaaaaaaaa", breite: 40, erwartet: []string{"aaaaaaaa", "aaaaaaaa", "aaa"}},
		// **Ein einzelnes Zeichen bleibt stehen**, auch wenn es übersteht.
		{name: "ein Zeichen breiter als die Zeile", text: "ab", breite: 4, erwartet: []string{"a", "b"}},
		// Ein Wort, das passt, bleibt ganz.
		{name: "passend", text: "aaaaaaaa", breite: 40, erwartet: []string{"aaaaaaaa"}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			gesetzt := render.SetzeAbschnitte(
				render.Satz{Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben},
				render.Rechteck{Breite: fall.breite, Hoehe: 100},
				[]render.Abschnitt{{Text: fall.text}},
				render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

			zeilen := make([]string, 0, len(gesetzt.Zeilen))
			for _, zeile := range gesetzt.Zeilen {
				zeilen = append(zeilen, zeile.Text)
			}

			if !slices.Equal(zeilen, fall.erwartet) {
				t.Errorf("Zeilen %q, erwartet %q", zeilen, fall.erwartet)
			}
		})
	}
}

// **Die Breite des Kopfs ist gemessen, nicht geschätzt** — die Zeile, die
// ein Zwangsumbruch hinterlässt, ist so breit wie ihre Zeichen.
func TestDerKopfEinesGebrochenenWortsHatSeineBreite(t *testing.T) {
	gesetzt := render.SetzeAbschnitte(
		render.Satz{Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben},
		render.Rechteck{Breite: 40, Hoehe: 100},
		[]render.Abschnitt{{Text: "aaaaaaaaaaaa"}},
		render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

	if len(gesetzt.Zeilen) != 2 || gesetzt.Zeilen[0].Breite != 40 || gesetzt.Zeilen[1].Breite != 20 {
		t.Errorf("Zeilen %+v, erwartet 40 und 20 breit", gesetzt.Zeilen)
	}
}

// **Ein Symbolzug bleibt ein Zeichen** — gebrochen wird vor oder nach ihm,
// nie darin ([render.Setze]).
func TestEinSymbolzugBleibtBeimZwangsumbruchGanz(t *testing.T) {
	// „aaaa", das Symbol mit 18 Punkten (siehe
	// [TestEinSymbolMisstInGanzenPunkten]) und „aaaa": 58 in einer Zeile von
	// 40. Mit dem Symbol sind es 38, mit dem nächsten Zeichen 43.
	gesetzt := render.SetzeAbschnitte(
		render.Satz{Groesse: 10, Ausrichtung: render.LinksOben},
		render.Rechteck{Breite: 40, Hoehe: 100},
		[]render.Abschnitt{
			{Text: "aaaa"},
			{
				Text: "1", Symbole: []render.Symbol{{Bild: render.Blatt("/1.png")}},
				Symbolmasse: render.Symbolmasse{Groesse: 14, Bildgroesse: 135, AbstandWaagerecht: 2},
			},
			{Text: "aaaa"},
		},
		render.Schriftsatz{Aufrecht: einheitsschrift{}},
		bildmasse{da: map[string]image.Point{"/1.png": {X: 160, Y: 160}}})

	zeilen := make([]string, 0, len(gesetzt.Zeilen))
	for _, zeile := range gesetzt.Zeilen {
		zeilen = append(zeilen, zeile.Text)
	}

	if erwartet := []string{"aaaa1", "aaaa"}; !slices.Equal(zeilen, erwartet) {
		t.Errorf("Zeilen %q, erwartet %q", zeilen, erwartet)
	}
}
