package render

import (
	"image"
	"testing"
)

// ebenenquelle baut ein kleines Bild, dessen Punkte sich unterscheiden lassen.
func ebenenquelle(ton uint8) *image.NRGBA {
	bild := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for i := range bild.Pix {
		bild.Pix[i] = 255
	}

	bild.Pix[0] = ton

	return bild
}

// **Dieselbe Quelle in derselben Größe wird einmal abgetastet**
// (#672).
func TestEineEbeneWirdEinmalAbgetastet(t *testing.T) {
	t.Parallel()

	r := NeuerRasterer(nil)
	quelle := ebenenquelle(10)
	rahmen := image.Rect(0, 0, 8, 8)

	erste := r.abgetasteteEbene(quelle, rahmen, GueteMittel, true)
	zweite := r.abgetasteteEbene(quelle, rahmen, GueteMittel, true)

	if erste != zweite {
		t.Error("die Ebene wurde zweimal abgetastet")
	}

	// **Die Lage geht nicht ein, nur die Größe** — `abtasten` rechnet aus
	// `x-rahmen.Min.X`, und derselbe Rahmen zwei Punkte weiter gibt dasselbe
	// Bild.
	verschoben := r.abgetasteteEbene(quelle, rahmen.Add(image.Pt(2, 3)), GueteMittel, true)
	if erste != verschoben {
		t.Error("ein verschobener Rahmen gilt als andere Ebene")
	}

	// **Eine andere Größe ist eine andere Ebene.**
	if andere := r.abgetasteteEbene(quelle, image.Rect(0, 0, 9, 8), GueteMittel, true); erste == andere {
		t.Error("zwei Größen teilen sich eine Ebene")
	}
}

// **Eine Quelle ohne Zeiger wird nicht gemerkt.**
//
// `ausschnitt` ist ein Wert; als Schlüssel einer Abbildung verglichen er
// seinen Inhalt statt seiner Identität — und wäre damit kein Schlüssel mehr.
func TestEineQuelleOhneZeigerWirdNichtGemerkt(t *testing.T) {
	t.Parallel()

	r := NeuerRasterer(nil)
	quelle := ausschnitt{ebenenquelle(10), image.Rect(0, 0, 4, 4)}
	rahmen := image.Rect(0, 0, 8, 8)

	erste := r.abgetasteteEbene(quelle, rahmen, GueteMittel, true)
	zweite := r.abgetasteteEbene(quelle, rahmen, GueteMittel, true)

	if erste == zweite {
		t.Error("eine Quelle ohne Zeiger wurde gemerkt")
	}

	// **Und sie wird trotzdem richtig abgetastet** — nicht gemerkt heißt
	// langsam, nicht falsch.
	if farbeAusEbene(erste, 0, 0) != farbeAusEbene(zweite, 0, 0) {
		t.Error("zwei Läufe über dieselbe Quelle geben verschiedene Farben")
	}
}

// **Die Grenze räumt ganz, und danach wird wieder abgetastet.**
func TestDieEbenengrenzeRaeumt(t *testing.T) {
	t.Parallel()

	r := NeuerRasterer(nil)
	quelle := ebenenquelle(10)
	rahmen := image.Rect(0, 0, 8, 8)

	erste := r.abgetasteteEbene(quelle, rahmen, GueteMittel, true)

	// Ein Bild, das die Grenze allein sprengt.
	gross := image.Rect(0, 0, 4096, 4096)
	r.abgetasteteEbene(ebenenquelle(20), gross, GueteMittel, true)

	if zweite := r.abgetasteteEbene(quelle, rahmen, GueteMittel, true); erste == zweite {
		t.Error("die Grenze hat nicht geräumt")
	}
}
