package render

import (
	"bytes"
	"testing"
)

// **Kerning über Klassen — `PairPos` Format 2** — QA-03.AK-02
// (#1656).
//
// `beleren-bold`, die Schrift jedes Kartennamens, legt ihr Kerning in **drei**
// Untertabellen eines Lookups ab: zwei mit Einzelpaaren, dahinter eine mit
// Klassen. Hier stand, die Namensschrift brauche die dritte nicht. Gemessen am
// 2026-09-23 über 25.011 Kartennamen des Kanons gegen Core Text — den Weg, auf
// dem MSE2 unter macOS setzt —: **mit** den Klassen stimmt die Breite jedes
// Namens auf ein Tausendstel Punkt, **ohne** sie nur bei 2.287.
//
// # Die Reihenfolge ist die Regel
//
// Innerhalb eines Lookups gilt die **erste** Untertabelle, die ein Paar kennt.
// Die Einzelpaare stehen vorn und sind die Ausnahmen; die Klassen dahinter
// sind der Normalfall. **Ein Einzelpaar mit dem Wert null sperrt die Klasse** —
// sonst bekäme ein Paar, das die Schrift ausdrücklich nicht kernt, den Wert
// seiner Klasse.
func TestKlassenpaareGeltenNachDenEinzelpaaren(t *testing.T) {
	t.Parallel()

	einzeln := paarpostabelle([]uint16{3}, [][]kernpaar{{
		{zweite: 7, vorschub: -224},
		{zweite: 8, vorschub: 0},
	}})

	klassen := klassenpaartabelle(klassenpaarbau{
		deckung: []uint16{3, 4},
		erste:   map[uint16]uint16{3: 1},
		zweite:  map[uint16]uint16{7: 1, 8: 1, 9: 2},
		werte: [][]int16{
			{0, -11, -22},
			{-30, -40, -50},
		},
	})

	kern := kernungAus(schriftMitNachschlaegen([][][]byte{{einzeln, klassen}}))

	for _, fall := range []struct {
		name          string
		erste, zweite uint16
		erwartet      int
	}{
		{"das Einzelpaar geht vor", 3, 7, -224},
		{"ein Einzelpaar mit null sperrt die Klasse", 3, 8, 0},
		{"ohne Einzelpaar gilt die Klasse", 3, 9, -50},
		{"eine zweite Glyphe ohne Klasse ist Klasse null", 3, 5, -30},
		{"eine erste Glyphe ohne Klasse ist Klasse null", 4, 9, -22},
		{"eine erste Glyphe außerhalb der Deckung kernt nicht", 6, 9, 0},
	} {
		if got := kern.fuer(fall.erste, fall.zweite); got != fall.erwartet {
			t.Errorf("%s: %d/%d = %d, erwartet %d", fall.name, fall.erste, fall.zweite, got, fall.erwartet)
		}
	}
}

// **Zwei Lookups addieren sich; zwei Untertabellen eines Lookups nicht.**
//
// Die Lookups eines Merkmals laufen nacheinander über die Zeile, und jeder
// setzt seine Verschiebung dazu. Innerhalb eines Lookups endet die Suche bei
// der ersten Untertabelle, die das Paar kennt — auch wenn eine spätere es
// ebenfalls kennt.
func TestLookupsAddierenSichUntertabellenNicht(t *testing.T) {
	t.Parallel()

	erste := paarpostabelle([]uint16{3}, [][]kernpaar{{{zweite: 7, vorschub: -100}}})
	zweite := paarpostabelle([]uint16{3}, [][]kernpaar{{{zweite: 7, vorschub: -7}}})

	getrennt := kernungAus(schriftMitNachschlaegen([][][]byte{{erste}, {zweite}}))
	if got := getrennt.fuer(3, 7); got != -107 {
		t.Errorf("zwei Lookups = %d, erwartet -107", got)
	}

	gemeinsam := kernungAus(schriftMitNachschlaegen([][][]byte{{erste, zweite}}))
	if got := gemeinsam.fuer(3, 7); got != -100 {
		t.Errorf("zwei Untertabellen eines Lookups = %d, erwartet -100 — die erste gilt", got)
	}
}

// **Mit GPOS gilt die alte `kern`-Tabelle nicht mehr.**
//
// Eine Schrift, die beides trägt, meint dasselbe zweimal: `vanilla whale.ttf`
// des Packs führt jedes Paar der alten Tabelle auch in `GPOS`, mit demselben
// Wert. Core Text zählt es einmal — über 32 Paare am 2026-09-23 nachgemessen;
// hier stand eine Karte, in die beide schrieben.
func TestMitGPOSGiltDieAlteKerntabelleNicht(t *testing.T) {
	t.Parallel()

	gpos := schriftMitGPOS(paarpostabelle([]uint16{3}, [][]kernpaar{{{zweite: 7, vorschub: -100}}}))
	alt := schriftMitKerntabelle(0x0001, [][3]int{{3, 7, -100}, {4, 8, -20}})

	beide := mitZweiterTabelle(gpos, alt)

	kern := kernungAus(beide)
	if got := kern.fuer(3, 7); got != -100 {
		t.Errorf("3/7 = %d, erwartet -100 — einmal, nicht zweimal", got)
	}

	if got := kern.fuer(4, 8); got != 0 {
		t.Errorf("4/8 = %d, erwartet 0 — das Paar steht nur in der alten Tabelle", got)
	}

	if got := kernungAus(alt).fuer(4, 8); got != -20 {
		t.Errorf("ohne GPOS: 4/8 = %d, erwartet -20", got)
	}
}

// **Eine Klassentabelle, die mehr verspricht, als die Datei trägt**, wird
// nicht zugeteilt.
//
// Zwei Zahlen im Kopf bestimmen, wie groß die Wertematrix ist — bis zu
// 65.535 mal 65.535 Einträge. Wer sie glaubt, bevor er nachsieht, lässt eine
// beschädigte Datei den Arbeitsspeicher bestimmen.
func TestEineZuGrosseKlassentabelleWirdNichtZugeteilt(t *testing.T) {
	t.Parallel()

	klassen := klassenpaartabelle(klassenpaarbau{
		deckung: []uint16{3},
		erste:   map[uint16]uint16{3: 0},
		zweite:  map[uint16]uint16{9: 1},
		werte:   [][]int16{{0, -5}},
	})

	// class1Count und class2Count stehen an Stelle 12 und 14.
	setze(klassen, 12, 0xFFFF)
	setze(klassen, 14, 0xFFFF)

	kern := kernungAus(schriftMitNachschlaegen([][][]byte{{klassen}}))
	if got := kern.fuer(3, 9); got != 0 {
		t.Errorf("3/9 = %d, erwartet 0 — die Tabelle trägt nicht, was sie nennt", got)
	}
}

// **Abgeschnittene Klassentabellen brechen den Leser nicht** — wie in
// [TestEineAbgeschnitteneSchriftGibtKeinePaare], mit beiden Formaten der
// Klassendefinition.
func TestEineAbgeschnitteneKlassentabelleBrichtNichts(t *testing.T) {
	t.Parallel()

	for _, bereiche := range []bool{false, true} {
		klassen := klassenpaartabelle(klassenpaarbau{
			deckung:  []uint16{3, 4},
			erste:    map[uint16]uint16{3: 1, 4: 1},
			zweite:   map[uint16]uint16{7: 1, 8: 2},
			werte:    [][]int16{{0, -1, -2}, {-3, -4, -5}},
			bereiche: bereiche,
		})

		ganz := schriftMitNachschlaegen([][][]byte{{klassen}})
		if got := kernungAus(ganz).fuer(3, 8); got != -5 {
			t.Fatalf("ganz (Bereiche %v): 3/8 = %d, erwartet -5", bereiche, got)
		}

		for laenge := range len(ganz) {
			kern := kernungAus(ganz[:laenge])
			kern.fuer(3, 8)
			kern.fuer(4, 7)
		}
	}
}

// klassenpaarbau beschreibt eine `PairPos`-Untertabelle in Format 2.
type klassenpaarbau struct {
	deckung       []uint16
	erste, zweite map[uint16]uint16
	werte         [][]int16

	// bereiche legt die Klassendefinitionen in Format 2 ab statt in Format 1.
	bereiche bool
}

// klassenpaartabelle baut eine `PairPos`-Untertabelle in Format 2 mit
// Vorschub im ersten Wert.
func klassenpaartabelle(bau klassenpaarbau) []byte {
	zeilen, spalten := len(bau.werte), len(bau.werte[0])

	kopf := make([]byte, 16+2*zeilen*spalten)
	setze(kopf, 0, 2)
	setze(kopf, 4, 0x0004) // valueFormat1: XAdvance
	setze(kopf, 6, 0)      // valueFormat2
	setze(kopf, 12, zeilen)
	setze(kopf, 14, spalten)

	for i, zeile := range bau.werte {
		for j, wert := range zeile {
			setze(kopf, 16+2*(i*spalten+j), int(uint16(wert))) //nolint:gosec // Testdaten
		}
	}

	deck := make([]byte, 4+2*len(bau.deckung))
	setze(deck, 0, 1)
	setze(deck, 2, len(bau.deckung))

	for i, g := range bau.deckung {
		setze(deck, 4+i*2, int(g))
	}

	erste := klassendefinition(bau.erste, bau.bereiche)
	zweite := klassendefinition(bau.zweite, bau.bereiche)

	setze(kopf, 2, len(kopf))
	setze(kopf, 8, len(kopf)+len(deck))
	setze(kopf, 10, len(kopf)+len(deck)+len(erste))

	return bytes.Join([][]byte{kopf, deck, erste, zweite}, nil)
}

// klassendefinition baut eine `ClassDef`-Tabelle — Format 1 als Feld ab der
// kleinsten Glyphe, Format 2 als ein Bereich je Glyphe.
func klassendefinition(klassen map[uint16]uint16, bereiche bool) []byte {
	kleinste, groesste := uint16(0xFFFF), uint16(0)
	for g := range klassen {
		kleinste, groesste = min(kleinste, g), max(groesste, g)
	}

	if bereiche {
		aus := make([]byte, 4, 4+6*len(klassen))
		setze(aus, 0, 2)
		setze(aus, 2, len(klassen))

		for g := kleinste; g <= groesste; g++ {
			klasse, da := klassen[g]
			if !da {
				continue
			}

			eintrag := make([]byte, 6)
			setze(eintrag, 0, int(g))
			setze(eintrag, 2, int(g))
			setze(eintrag, 4, int(klasse))
			aus = append(aus, eintrag...)
		}

		return aus
	}

	anzahl := int(groesste-kleinste) + 1
	aus := make([]byte, 6+2*anzahl)
	setze(aus, 0, 1)
	setze(aus, 2, int(kleinste))
	setze(aus, 4, anzahl)

	for g, klasse := range klassen {
		setze(aus, 6+2*int(g-kleinste), int(klasse))
	}

	return aus
}

// **Eine Deckung, deren Bereich bei der letzten Glyphennummer endet, endet.**
//
// Die Schleife über einen Bereich lief mit einer 16-Bit-Zahl: bei `0xFFFF`
// springt sie auf null zurück und hört nie auf. Gültig ist so eine Schrift
// nicht — die höchste Glyphennummer ist 65.534 —, aber eine Schriftdatei kommt
// aus einem fremden Paket (§8), und eine beschädigte hing den Rasterer auf.
//
// **Und ein Bereich über alle Nummern teilt nicht mehr zu, als es Glyphen
// geben kann**, gleich wie oft er in der Datei steht.
func TestEineDeckungBisZurLetztenGlypheEndet(t *testing.T) {
	t.Parallel()

	tabelle := paarpostabelle([]uint16{3}, [][]kernpaar{{{zweite: 7, vorschub: -10}}})
	aufBereichsdeckung(tabelle, 0xFFF0, 0xFFFF)

	kernungAus(schriftMitGPOS(tabelle))

	klassen := klassenpaartabelle(klassenpaarbau{
		deckung:  []uint16{3},
		erste:    map[uint16]uint16{3: 0},
		zweite:   map[uint16]uint16{9: 1},
		werte:    [][]int16{{0, -5}},
		bereiche: true,
	})

	// Die zweite Klassendefinition bekommt einen Bereich bis 0xFFFF.
	zweite := int(zwei(klassen, 10))
	setze(klassen[zweite:], 4, 0)
	setze(klassen[zweite:], 6, 0xFFFF)

	kern := kernungAus(schriftMitNachschlaegen([][][]byte{{klassen}}))
	if got := kern.fuer(3, 0xFFFE); got != -5 {
		t.Errorf("3/0xFFFE = %d, erwartet -5 — der Bereich reicht bis ans Ende", got)
	}
}

// **Eine Klasse jenseits der Matrix gilt nicht — und die Untertabelle dahinter
// kommt zum Zug.** Eine Klassentabelle ohne Vorschub dagegen gilt, mit null,
// und sperrt, was nach ihr kommt.
func TestWannEineKlassentabelleGilt(t *testing.T) {
	t.Parallel()

	dahinter := paarpostabelle([]uint16{3}, [][]kernpaar{{{zweite: 9, vorschub: -77}}})

	jenseits := klassenpaartabelle(klassenpaarbau{
		deckung: []uint16{3},
		erste:   map[uint16]uint16{3: 0},
		zweite:  map[uint16]uint16{9: 5},
		werte:   [][]int16{{0, -1, -2}},
	})

	if got := kernungAus(schriftMitNachschlaegen([][][]byte{{jenseits, dahinter}})).fuer(3, 9); got != -77 {
		t.Errorf("Klasse 5 bei drei Spalten: 3/9 = %d, erwartet -77 aus der Untertabelle dahinter", got)
	}

	ohneVorschub := klassenpaartabelle(klassenpaarbau{
		deckung: []uint16{3},
		erste:   map[uint16]uint16{3: 0},
		zweite:  map[uint16]uint16{9: 1},
		werte:   [][]int16{{0, -5}},
	})
	setze(ohneVorschub, 4, 0x0001) // valueFormat1: nur XPlacement

	if got := kernungAus(schriftMitNachschlaegen([][][]byte{{ohneVorschub, dahinter}})).fuer(3, 9); got != 0 {
		t.Errorf("Klassen ohne Vorschub: 3/9 = %d, erwartet 0 — sie gelten und sperren", got)
	}
}

// **Eine Klassendefinition macht nicht mehr Schritte, als es Glyphennummern
// gibt** — auch wenn sie denselben Bereich über alle Nummern zweimal nennt.
// Und das erste Format zählt nicht über die letzte Nummer hinaus.
func TestEineKlassendefinitionHatEinEnde(t *testing.T) {
	t.Parallel()

	bereiche := make([]byte, 4+12)
	setze(bereiche, 0, 2)
	setze(bereiche, 2, 2)
	setze(bereiche, 6, 0xFFFF)
	setze(bereiche, 8, 1)
	setze(bereiche, 12, 0xFFFF)
	setze(bereiche, 14, 2)

	klassen := klassenAus(bereiche, 0)
	if klassen[0xFFFE] != 1 || klassen[0] != 1 {
		t.Errorf("0xFFFE → %d, 0 → %d; erwartet beide in Klasse 1 — der zweite Bereich endet am Deckel",
			klassen[0xFFFE], klassen[0])
	}

	feld := make([]byte, 6+4)
	setze(feld, 0, 1)
	setze(feld, 2, 0xFFFF)
	setze(feld, 4, 2)
	setze(feld, 6, 7)
	setze(feld, 8, 8)

	if klassen := klassenAus(feld, 0); len(klassen) != 1 || klassen[0xFFFF] != 7 {
		t.Errorf("%v, erwartet nur 0xFFFF → 7 — hinter der letzten Nummer ist keine Glyphe", klassen)
	}
}
