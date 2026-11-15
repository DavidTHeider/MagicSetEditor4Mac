package render_test

import (
	"image"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Ein Textzug besteht aus Abschnitten, nicht aus einer Zeichenkette.**
//
// # Warum das der nötige Umbau ist
//
// Kursiv steht **mitten** im Regeltext: `<i-flavor>` beginnt nach dem
// Regelabsatz und läuft bis zum Ende. Solange [render.Ebene] eine flache
// Zeichenkette trägt, gibt es keine Stelle, an der ein Schnitt wechseln könnte
// — es ist derselbe Umbau wie beim Bildbaum (#270), eine Ebene höher
// (#312).
//
// # Was hier zugesagt wird
//
// Der Umbruch geht **über** die Abschnittsgrenze: „Feuer" kursiv und „ball"
// aufrecht sind ein Wort und bleiben zusammen. Und jedes Stück trägt seine
// eigene Stelle, denn nur so kann der Rasterer es mit seiner eigenen Schrift
// zeichnen.
func TestEinAbschnittsweiseGesetzterZug(t *testing.T) {
	satz := render.Satz{Groesse: 10}
	schriften := render.Schriftsatz{
		Aufrecht: gleichbreit{breite: 0.5},
		// **Die Kursive ist breiter**, damit der Test sieht, welche gefragt
		// wurde. Eine gleich breite bewiese nichts.
		Kursiv: gleichbreit{breite: 1.0},
	}

	ergebnis := render.SetzeAbschnitte(satz, render.Rechteck{Breite: 200, Hoehe: 30},
		[]render.Abschnitt{
			{Text: "ab"},
			{Text: "cd", Kursiv: true},
		}, schriften, nil)

	if len(ergebnis.Zeilen) != 1 {
		t.Fatalf("%d Zeilen, erwartet eine", len(ergebnis.Zeilen))
	}

	zeile := ergebnis.Zeilen[0]
	if len(zeile.Stuecke) != 2 {
		t.Fatalf("%d Stücke, erwartet zwei: %+v", len(zeile.Stuecke), zeile.Stuecke)
	}

	// `ab` aufrecht: 2 × 0,5 × 10 = 10. `cd` kursiv: 2 × 1,0 × 10 = 20.
	if breite := zeile.Stuecke[0].Breite; breite != 10 {
		t.Errorf("das aufrechte Stück ist %.1f breit, erwartet 10", breite)
	}

	if breite := zeile.Stuecke[1].Breite; breite != 20 {
		t.Errorf("das kursive Stück ist %.1f breit, erwartet 20 — "+
			"es wurde mit der aufrechten Schrift gemessen", breite)
	}

	// **Das zweite Stück beginnt, wo das erste endet.**
	if links := zeile.Stuecke[1].Links; links != 10 {
		t.Errorf("das kursive Stück beginnt bei %.1f, erwartet 10", links)
	}

	if zeile.Breite != 30 {
		t.Errorf("die Zeile ist %.1f breit, erwartet 30", zeile.Breite)
	}
}

// **Ein Wort bleibt zusammen, auch wenn der Schnitt darin wechselt.**
//
// „Feuer" kursiv und „ball" aufrecht ist **ein** Wort. Es an der
// Abschnittsgrenze umzubrechen hieße, eine Zeile zu erzeugen, die niemand
// geschrieben hat — dieselbe Begründung wie bei `{T}` in [render.Setze].
func TestEinWortMitSchnittwechselBleibtZusammen(t *testing.T) {
	satz := render.Satz{Groesse: 10}
	schriften := render.Schriftsatz{Aufrecht: gleichbreit{breite: 1.0}}

	// Rahmen für 10 Zeichen; „ab Feuerball" hat zwölf. „ab Feuer" passte —
	// wer an der Abschnittsgrenze bräche, ließe „ball" allein auf Zeile zwei.
	//
	// Hier stand bis zum Nachlauf über E/F ein Rahmen für fünf Zeichen, in den
	// das Wort gar nicht passte; seit es dann zwischen zwei Zeichen gebrochen
	// wird wie bei MSE ([zwangsgebrochen]), fiele der Schnitt dort zufällig
	// auf die Abschnittsgrenze.
	ergebnis := render.SetzeAbschnitte(satz, render.Rechteck{Breite: 100, Hoehe: 60},
		[]render.Abschnitt{
			{Text: "ab "},
			{Text: "Feuer", Kursiv: true},
			{Text: "ball"},
		}, schriften, nil)

	if len(ergebnis.Zeilen) != 2 || ergebnis.Zeilen[1].Text != "Feuerball" {
		t.Fatalf("%d Zeilen, erwartet »ab« und »Feuerball« — das Wort wurde an der "+
			"Abschnittsgrenze getrennt: %+v", len(ergebnis.Zeilen), ergebnis.Zeilen)
	}
}

// **Ohne Kursivschrift wird die aufrechte genommen** — und nicht nichts.
//
// Ein Stil ohne `italic name:` gibt es: von den Feldstilen mit `font:` nennen
// ihn nicht alle. MSE2 fällt dort auf denselben Weg zurück wie ohne die Angabe
// (`src/data/font.cpp`, Zeile 100 — die Bedingung verlangt
// `!italic_name().empty()`).
//
// Fassung f9d9356 vom 2021-07-06.
func TestOhneKursivschriftWirdDieAufrechteGenommen(t *testing.T) {
	ergebnis := render.SetzeAbschnitte(
		render.Satz{Groesse: 10}, render.Rechteck{Breite: 200, Hoehe: 30},
		[]render.Abschnitt{{Text: "ab", Kursiv: true}},
		render.Schriftsatz{Aufrecht: gleichbreit{breite: 0.5}}, nil)

	if len(ergebnis.Zeilen) != 1 || len(ergebnis.Zeilen[0].Stuecke) != 1 {
		t.Fatalf("unerwartetes Ergebnis: %+v", ergebnis.Zeilen)
	}

	if breite := ergebnis.Zeilen[0].Stuecke[0].Breite; breite != 10 {
		t.Errorf("%.1f breit, erwartet 10 — die aufrechte Schrift wurde nicht genommen",
			breite)
	}
}

// **Der Klartext einer Ebene ist ihr Zug ohne die Schnitte.**
//
// Alles, was nur **misst** — `content_width`, die Suche, die Ausfuhr —, fragt
// danach; die Schnitte gehen es nichts an.
func TestDerKlartextEinerEbene(t *testing.T) {
	ebene := render.Ebene{Abschnitte: []render.Abschnitt{
		{Text: "Fliegend. "},
		{Text: "Er kam mit dem Sturm.", Kursiv: true},
	}}

	if klar := ebene.Klartext(); klar != "Fliegend. Er kam mit dem Sturm." {
		t.Errorf("der Klartext lautet %q", klar)
	}
}

// **Zwei Lücken nebeneinander ergeben kein Wort aus Leerzeichen.**
//
// Ein Abschnitt, der auf ein Leerzeichen endet, und der nächste, der mit einem
// beginnt, liefern zwei Lückenstücke hintereinander. Ohne die Bedingung in
// [worteAus] entstünde daraus ein „Wort", das der Umbruch wie ein Wort
// behandelte — und eine Zeile könnte mit Leerzeichen anfangen, die niemand
// gesetzt hat.
func TestZweiLueckenNebeneinanderErgebenKeinWort(t *testing.T) {
	ergebnis := render.SetzeAbschnitte(
		render.Satz{Groesse: 10}, render.Rechteck{Breite: 200, Hoehe: 30},
		[]render.Abschnitt{{Text: "ab "}, {Text: " cd", Kursiv: true}},
		render.Schriftsatz{Aufrecht: gleichbreit{breite: 0.5}}, nil)

	if len(ergebnis.Zeilen) != 1 {
		t.Fatalf("%d Zeilen, erwartet eine", len(ergebnis.Zeilen))
	}

	if text := ergebnis.Zeilen[0].Text; text != "ab  cd" {
		t.Errorf("die Zeile lautet %q, erwartet %q", text, "ab  cd")
	}
}

// **Ein schrumpfender Zug ohne Text bricht nicht und misst null.**
func TestEinGestauchterZugOhneTextMisstNull(t *testing.T) {
	ergebnis := render.SetzeAbschnitte(
		render.Satz{Groesse: 10, Stauchen: true, Einzeilig: true, NurBeiUeberlauf: true},
		render.Rechteck{Breite: 200, Hoehe: 30},
		render.Aufrecht(""),
		render.Schriftsatz{Aufrecht: gleichbreit{breite: 0.5}}, nil)

	if len(ergebnis.Zeilen) != 1 || ergebnis.Zeilen[0].Breite != 0 {
		t.Errorf("unerwartetes Ergebnis: %+v", ergebnis.Zeilen)
	}
}

// **Ein kursives Stück wird mit der Kursivschrift gezeichnet.**
//
// Der Nachweis geht bis aufs Bild und nicht bis zur Messung: die eine Schrift
// zeichnet ein Quadrat halber Breite, die andere eines voller Breite. Was
// rechts von der halben Breite steht, kann nur aus der Kursiven stammen.
func TestEinKursivesStueckWirdMitDerKursivschriftGezeichnet(t *testing.T) {
	plan := &render.Plan{Breite: 60, Hoehe: 20, Ebenen: []render.Ebene{{
		Art:        render.ArtText,
		Rahmen:     render.Rechteck{Breite: 60, Hoehe: 20},
		Abschnitte: []render.Abschnitt{{Text: "a", Kursiv: true}},
		Farbe:      render.Farbe{A: 255},
		Satz: render.Satz{
			Schrift: "Aufrecht", Kursivschrift: "Kursiv", Groesse: 10,
		},
		Deckung: 1,
	}}}

	bild, befunde, err := render.NeuerRasterer(nil).
		MitSchriften(schriften{
			"Aufrecht": einheitsschrift{},
			"Kursiv":   breiteschrift{},
		}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	// Die aufrechte malt 5 Punkte breit, die kursive 10. Bei x = 7 steht also
	// nur dann etwas, wenn die Kursive gefragt wurde.
	gefunden := false

	for y := range bild.Bounds().Dy() {
		if _, _, _, a := bild.At(7, y).RGBA(); a > 0 {
			gefunden = true
		}
	}

	if !gefunden {
		t.Error("bei x = 7 steht nichts — gezeichnet wurde mit der aufrechten Schrift")
	}
}

// **Ist die Kursivschrift nicht zu finden, wird aufrecht gesetzt.**
//
// Bestenfalls-Arbeit (§8): ein Flavortext, der aufrecht steht, ist eine Karte;
// einer, der fehlt, ist keine.
func TestOhneAuffindbareKursivschriftWirdAufrechtGesetzt(t *testing.T) {
	plan := &render.Plan{Breite: 60, Hoehe: 20, Ebenen: []render.Ebene{{
		Art:        render.ArtText,
		Rahmen:     render.Rechteck{Breite: 60, Hoehe: 20},
		Abschnitte: []render.Abschnitt{{Text: "a", Kursiv: true}},
		Farbe:      render.Farbe{A: 255},
		Satz: render.Satz{
			Schrift: "Aufrecht", Kursivschrift: "Gibtsnicht", Groesse: 10,
		},
		Deckung: 1,
	}}}

	_, befunde, err := render.NeuerRasterer(nil).
		MitSchriften(schriften{"Aufrecht": einheitsschrift{}}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) > 0 {
		t.Errorf("Befunde: %v — eine fehlende Kursive hält die Karte nicht auf", befunde)
	}
}

// **Eine Kursivschrift ohne Glyphen lässt ihr Stück aus, nicht die Zeile.**
func TestEineKursiveOhneGlyphenLaesstIhrStueckAus(t *testing.T) {
	plan := &render.Plan{Breite: 60, Hoehe: 20, Ebenen: []render.Ebene{{
		Art:    render.ArtText,
		Rahmen: render.Rechteck{Breite: 60, Hoehe: 20},
		Abschnitte: []render.Abschnitt{
			{Text: "a"}, {Text: "b", Kursiv: true},
		},
		Farbe: render.Farbe{A: 255},
		Satz: render.Satz{
			Schrift: "Aufrecht", Kursivschrift: "Ohne", Groesse: 10,
		},
		Deckung: 1,
	}}}

	bild, _, err := render.NeuerRasterer(nil).
		MitSchriften(schriften{
			"Aufrecht": einheitsschrift{},
			"Ohne":     gleichbreit{breite: 0.5},
		}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	// Das aufrechte `a` steht trotzdem da.
	gefunden := false

	for y := range bild.Bounds().Dy() {
		if _, _, _, a := bild.At(1, y).RGBA(); a > 0 {
			gefunden = true
		}
	}

	if !gefunden {
		t.Error("das aufrechte Stück fehlt — ein Stück ohne Glyphen hat die Zeile mitgenommen")
	}
}

// breiteschrift zeichnet Blöcke voller Breite — die Gegenprobe zur
// [einheitsschrift], die halbe zeichnet.
type breiteschrift struct{}

func (breiteschrift) Breite(rune) float64  { return 1.0 }
func (breiteschrift) Zeilenhoehe() float64 { return 1.2 }
func (breiteschrift) Grundlinie() float64  { return 0.8 }

func (breiteschrift) Maske(_ rune, groesse float64) (image.Image, image.Point, bool) {
	seite := max(int(groesse), 1)

	maske := image.NewAlpha(image.Rect(0, 0, seite, seite))
	for i := range maske.Pix {
		maske.Pix[i] = 255
	}

	return maske, image.Point{}, true
}

// **Ein weicher Trenner am Zeilenende zählt bei der Ausrichtung nicht mit.**
//
//	size_t end_or_soft; ///< Index just beyond the last non-soft character
//	double width() const { return positions[end_or_soft-start] - positions.front(); }
//
// `src/render/text/viewer.cpp`, Zeilen 17–35 und 610, Fassung f9d9356 vom
// 2021-07-06.
//
// **Nur der Schwanz fällt weg.** MSE merkt sich die Stelle hinter dem letzten
// nicht weichen Zeichen; ein weiches Stück **zwischen** zwei harten zählt also
// mit (#345).
func TestEinWeicherTrennerZaehltNichtMit(t *testing.T) {
	satz := render.Satz{
		Schrift: "Probe", Groesse: 10, Ausrichtung: render.MitteOben,
	}
	rahmen := render.Rechteck{Breite: 100, Hoehe: 40}

	nur := render.SetzeAbschnitte(satz, rahmen,
		[]render.Abschnitt{{Text: "AB"}}, render.Schriftsatz{Aufrecht: gleichbreit{breite: 0.5}}, nil)

	mit := render.SetzeAbschnitte(satz, rahmen, []render.Abschnitt{
		{Text: "AB"}, {Text: "CD", Weich: true},
	}, render.Schriftsatz{Aufrecht: gleichbreit{breite: 0.5}}, nil)

	if len(nur.Zeilen) != 1 || len(mit.Zeilen) != 1 {
		t.Fatalf("%d und %d Zeilen", len(nur.Zeilen), len(mit.Zeilen))
	}

	if nur.Zeilen[0].Links != mit.Zeilen[0].Links {
		t.Errorf("mit weichem Schwanz beginnt die Zeile bei %g, ohne bei %g",
			mit.Zeilen[0].Links, nur.Zeilen[0].Links)
	}

	// **Und ein weiches Stück mitten drin zählt mit.**
	drin := render.SetzeAbschnitte(satz, rahmen, []render.Abschnitt{
		{Text: "AB"}, {Text: "CD", Weich: true}, {Text: "EF"},
	}, render.Schriftsatz{Aufrecht: gleichbreit{breite: 0.5}}, nil)

	if drin.Zeilen[0].Links >= nur.Zeilen[0].Links {
		t.Errorf("ein weiches Stück mitten drin verschiebt die Zeile nicht: %g gegen %g",
			drin.Zeilen[0].Links, nur.Zeilen[0].Links)
	}
}
