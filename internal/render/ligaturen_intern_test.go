package render

import (
	"bytes"
	"encoding/binary"
	"slices"
	"testing"
)

// **Ligaturen aus `GSUB`, Merkmal `liga`** —
// #1805.
//
// Die Tabellen hier sind von Hand gebaut, Byte für Byte nach der Gestalt aus
// der OpenType-Spezifikation (`GSUB`, Lookup-Arten 4 und 7). Gegen die echten
// Schriften des Packs und gegen Core Text hält sie
// `TestDieNamensschriftSetztWieCoreText` (`kerncoretext_intern_test.go`).

// Glyphennummern der Tabellen unten — erfunden, nur ihre Rolle zählt.
const (
	gF, gI, gL, gX  = 10, 11, 12, 13
	gFFI, gFFL, gFF = 100, 101, 102
)

func TestDieLigaturenErsetzenWieDieSchriftSieNennt(t *testing.T) {
	t.Parallel()

	// **Die Reihenfolge im Satz ist die der Schrift**: `ffi` und `ffl` vor
	// `ff`, sonst käme `ff` immer zuerst und `ffi` nie.
	standard := gsubMit(
		[]testmerkmal{{"liga", []int{0}}},
		[]testnachschlag{{4, [][]byte{ligaturuntertabelle(false, map[uint16][][]uint16{
			gF: {{gFFI, gF, gI}, {gFFL, gF, gL}, {gFF, gF}},
		})}}},
	)

	for _, fall := range []struct {
		name   string
		gsub   []byte
		ein    []uint16
		aus    []uint16
		deckt  []int
		glyphn int
	}{
		{"ffi wird eine Glyphe", standard, []uint16{gF, gF, gI}, []uint16{gFFI}, []int{3}, 1000},
		{"ffl wird eine Glyphe", standard, []uint16{gF, gF, gL}, []uint16{gFFL}, []int{3}, 1000},
		{"ff ohne i bleibt ff als Ligatur", standard, []uint16{gF, gF, gX}, []uint16{gFF, gX}, []int{2, 1}, 1000},
		{"ein einzelnes f bleibt", standard, []uint16{gX, gF}, []uint16{gX, gF}, []int{1, 1}, 1000},
		{"fi gibt es in dieser Schrift nicht", standard, []uint16{gF, gI}, []uint16{gF, gI}, []int{1, 1}, 1000},
		{
			"zwei Ligaturen hintereinander", standard,
			[]uint16{gX, gF, gF, gI, gF, gF, gL},
			[]uint16{gX, gFFI, gFFL},
			[]int{1, 3, 3},
			1000,
		},
		{
			"die Reihenfolge der Schrift entscheidet, nicht die Länge",
			gsubMit([]testmerkmal{{"liga", []int{0}}}, []testnachschlag{{4, [][]byte{
				ligaturuntertabelle(false, map[uint16][][]uint16{gF: {{gFF, gF}, {gFFI, gF, gI}}}),
			}}}),
			[]uint16{gF, gF, gI},
			[]uint16{gFF, gI},
			[]int{2, 1},
			1000,
		},
		{
			"eine Deckung im Bereichsformat",
			gsubMit([]testmerkmal{{"liga", []int{0}}}, []testnachschlag{{4, [][]byte{
				ligaturuntertabelle(true, map[uint16][][]uint16{gF: {{gFFI, gF, gI}}, gI: {{gX, gI}}}),
			}}}),
			[]uint16{gF, gF, gI, gI, gI},
			[]uint16{gFFI, gX},
			[]int{3, 2},
			1000,
		},
		{
			"eine Untertabelle ohne Treffer lässt die nächste zu",
			gsubMit([]testmerkmal{{"liga", []int{0}}}, []testnachschlag{{4, [][]byte{
				ligaturuntertabelle(false, map[uint16][][]uint16{gF: {{gFFL, gF, gX}}}),
				ligaturuntertabelle(false, map[uint16][][]uint16{gF: {{gFFI, gF, gI}}}),
			}}}),
			[]uint16{gF, gF, gI},
			[]uint16{gFFI},
			[]int{3},
			1000,
		},
		{
			// **Die Lookups nacheinander**: der zweite sieht, was der erste
			// ersetzt hat.
			"ein späterer Lookup sieht die Ligatur des früheren",
			gsubMit([]testmerkmal{{"liga", []int{1, 0}}}, []testnachschlag{
				{4, [][]byte{ligaturuntertabelle(false, map[uint16][][]uint16{gF: {{gFF, gF}}})}},
				{4, [][]byte{ligaturuntertabelle(false, map[uint16][][]uint16{gFF: {{gFFI, gI}}})}},
			}),
			[]uint16{gF, gF, gI},
			[]uint16{gFFI},
			[]int{3},
			1000,
		},
		{
			"eine Erweiterung (Art 7) trägt ihre Ligaturen",
			gsubMit([]testmerkmal{{"liga", []int{0}}}, []testnachschlag{{7, [][]byte{
				erweiterung(4, ligaturuntertabelle(false, map[uint16][][]uint16{gF: {{gFFI, gF, gI}}})),
			}}}),
			[]uint16{gF, gF, gI},
			[]uint16{gFFI},
			[]int{3},
			1000,
		},
		{
			// **Nur `liga`** — `calt` und `dlig` setzt Core Text nicht
			// (`docs/SCHRIFTUNTERSUCHUNG.md`).
			"andere Merkmale gelten nicht",
			gsubMit([]testmerkmal{{"calt", []int{0}}, {"dlig", []int{0}}}, []testnachschlag{{4, [][]byte{
				ligaturuntertabelle(false, map[uint16][][]uint16{gF: {{gFFI, gF, gI}}}),
			}}}),
			[]uint16{gF, gF, gI},
			[]uint16{gF, gF, gI},
			[]int{1, 1, 1},
			1000,
		},
		{
			"eine andere Lookup-Art unter liga wird übergangen",
			gsubMit([]testmerkmal{{"liga", []int{0}}}, []testnachschlag{{6, [][]byte{
				ligaturuntertabelle(false, map[uint16][][]uint16{gF: {{gFFI, gF, gI}}}),
			}}}),
			[]uint16{gF, gF, gI},
			[]uint16{gF, gF, gI},
			[]int{1, 1, 1},
			1000,
		},
		{
			// Eine Glyphe, die es nicht gibt, ließe sich nicht zeichnen — und
			// das Zeichen dafür wäre verloren. Die nächste Ligatur des Satzes
			// kommt dann zum Zug.
			"eine Ligatur jenseits der Glyphenzahl gilt nicht",
			gsubMit([]testmerkmal{{"liga", []int{0}}}, []testnachschlag{{4, [][]byte{
				ligaturuntertabelle(false, map[uint16][][]uint16{gF: {{500, gF, gI}, {gFF, gF}}}),
			}}}),
			[]uint16{gF, gF, gI},
			[]uint16{gFF, gI},
			[]int{2, 1},
			500,
		},
		{
			"eine Ligatur aus einer Glyphe ersetzt sie",
			gsubMit([]testmerkmal{{"liga", []int{0}}}, []testnachschlag{{4, [][]byte{
				ligaturuntertabelle(false, map[uint16][][]uint16{gI: {{gL}}}),
			}}}),
			[]uint16{gF, gI},
			[]uint16{gF, gL},
			[]int{1, 1},
			1000,
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			liga := ligaturenAus(mitTabelle("GSUB", fall.gsub), fall.glyphn)

			aus, deckt := liga.anwenden(fall.ein)
			if !slices.Equal(aus, fall.aus) || !slices.Equal(deckt, fall.deckt) {
				t.Errorf("%v → %v deckt %v, erwartet %v deckt %v", fall.ein, aus, deckt, fall.aus, fall.deckt)
			}
		})
	}
}

// **Eine kaputte Tabelle setzt nichts** — und hängt nichts auf. Fremde
// Schriften sind der Normalfall (§8).
func TestEineKaputteLigaturtabelleSetztNichts(t *testing.T) {
	t.Parallel()

	ganz := gsubMit([]testmerkmal{{"liga", []int{0}}}, []testnachschlag{{7, [][]byte{
		erweiterung(4, ligaturuntertabelle(true, map[uint16][][]uint16{gF: {{gFFI, gF, gI}}})),
	}}})

	// **Jede Kürzung**: jede Grenzprüfung des Lesers liegt an einer anderen
	// Stelle, und eine Kürzung trifft jede einmal.
	for laenge := range len(ganz) {
		liga := ligaturenAus(mitTabelle("GSUB", ganz[:laenge]), 1000)

		if aus, _ := liga.anwenden([]uint16{gF, gF, gI}); len(aus) == 1 && laenge < len(ganz)-2 {
			t.Errorf("auf %d von %d Byte gekürzt: %v", laenge, len(ganz), aus)
		}
	}

	// Und ohne Tabelle, mit Versätzen ins Leere und ohne Datei.
	for name, roh := range map[string][]byte{
		"ohne GSUB":               mitTabelle("GPOS", ganz),
		"Merkmalsliste im Leeren": mitTabelle("GSUB", gsubVersetzt(ganz, 6, 0xFFF0)),
		"Lookupliste im Leeren":   mitTabelle("GSUB", gsubVersetzt(ganz, 8, 0xFFF0)),
		"leere Datei":             nil,
	} {
		if liga := ligaturenAus(roh, 1000); len(liga.nachschlaege) != 0 {
			t.Errorf("%s: %d Lookups gelesen", name, len(liga.nachschlaege))
		}
	}
}

// testmerkmal ist ein Eintrag der Merkmalsliste: Kennung und Lookups.
type testmerkmal struct {
	kennung      string
	nachschlaege []int
}

// testnachschlag ist ein Lookup: Art und Untertabellen.
type testnachschlag struct {
	art           int
	untertabellen [][]byte
}

// gsubMit baut eine `GSUB`-Tabelle ohne Schriftsystemliste — gelesen wird
// jedes Merkmal, wie beim Kerning.
func gsubMit(merkmale []testmerkmal, nachschlaege []testnachschlag) []byte {
	liste := make([]byte, 2+6*len(merkmale), 2+10*len(merkmale)+2*zahlDerNachschlaege(merkmale))
	setze(liste, 0, len(merkmale))

	var eintraege []byte

	for i, merkmal := range merkmale {
		copy(liste[2+6*i:], merkmal.kennung)
		setze(liste, 6+6*i, len(liste)+len(eintraege))

		eintrag := make([]byte, 4+2*len(merkmal.nachschlaege))
		setze(eintrag, 2, len(merkmal.nachschlaege))

		for j, nummer := range merkmal.nachschlaege {
			setze(eintrag, 4+2*j, nummer)
		}

		eintraege = append(eintraege, eintrag...)
	}

	merkmalsliste := liste
	merkmalsliste = append(merkmalsliste, eintraege...)

	verweise := make([]byte, 2+2*len(nachschlaege))
	setze(verweise, 0, len(nachschlaege))

	var koerper []byte

	for i, nachschlag := range nachschlaege {
		setze(verweise, 2+2*i, len(verweise)+len(koerper))

		kopf := make([]byte, 6+2*len(nachschlag.untertabellen))
		setze(kopf, 0, nachschlag.art)
		setze(kopf, 4, len(nachschlag.untertabellen))

		var daten []byte

		for j, tabelle := range nachschlag.untertabellen {
			setze(kopf, 6+2*j, len(kopf)+len(daten))
			daten = append(daten, tabelle...)
		}

		koerper = append(koerper, kopf...)
		koerper = append(koerper, daten...)
	}

	kopf := make([]byte, 10)
	setze(kopf, 0, 1)
	setze(kopf, 6, len(kopf))
	setze(kopf, 8, len(kopf)+len(merkmalsliste))

	return bytes.Join([][]byte{kopf, merkmalsliste, verweise, koerper}, nil)
}

// zahlDerNachschlaege zählt die Lookupverweise aller Merkmale.
func zahlDerNachschlaege(merkmale []testmerkmal) int {
	zahl := 0
	for _, merkmal := range merkmale {
		zahl += len(merkmal.nachschlaege)
	}

	return zahl
}

// gsubVersetzt setzt einen Versatz im Kopf einer `GSUB`-Tabelle neu.
func gsubVersetzt(gsub []byte, stelle, wert int) []byte {
	aus := slices.Clone(gsub)
	setze(aus, stelle, wert)

	return aus
}

// ligaturuntertabelle baut eine `LigatureSubst`-Untertabelle in Format 1: je
// erster Glyphe die Ligaturen als Ziel, gefolgt von den übrigen Bestandteilen.
func ligaturuntertabelle(bereiche bool, saetze map[uint16][][]uint16) []byte {
	ersten := make([]uint16, 0, len(saetze))
	for erste := range saetze {
		ersten = append(ersten, erste)
	}

	slices.Sort(ersten)

	kopf := make([]byte, 6+2*len(ersten), 6+2*len(ersten)+len(saetze)*64)
	setze(kopf, 0, 1)
	setze(kopf, 4, len(ersten))

	var koerper []byte

	for i, erste := range ersten {
		setze(kopf, 6+2*i, len(kopf)+len(koerper))

		satz := make([]byte, 2+2*len(saetze[erste]))
		setze(satz, 0, len(saetze[erste]))

		var ligaturen []byte

		for j, ligatur := range saetze[erste] {
			setze(satz, 2+2*j, len(satz)+len(ligaturen))

			eintrag := make([]byte, 4+2*(len(ligatur)-1))
			setze(eintrag, 0, int(ligatur[0]))
			setze(eintrag, 2, len(ligatur))

			for k, bestandteil := range ligatur[1:] {
				setze(eintrag, 4+2*k, int(bestandteil))
			}

			ligaturen = append(ligaturen, eintrag...)
		}

		koerper = append(koerper, satz...)
		koerper = append(koerper, ligaturen...)
	}

	setze(kopf, 2, len(kopf)+len(koerper))

	return append(append(kopf, koerper...), deckungAus(bereiche, ersten)...)
}

// deckungAus baut eine `Coverage`-Tabelle — als Liste oder je Glyphe ein
// Bereich.
func deckungAus(bereiche bool, glyphen []uint16) []byte {
	if !bereiche {
		aus := make([]byte, 4+2*len(glyphen))
		setze(aus, 0, 1)
		setze(aus, 2, len(glyphen))

		for i, glyphe := range glyphen {
			setze(aus, 4+2*i, int(glyphe))
		}

		return aus
	}

	aus := make([]byte, 4+6*len(glyphen))
	setze(aus, 0, 2)
	setze(aus, 2, len(glyphen))

	for i, glyphe := range glyphen {
		setze(aus, 4+6*i, int(glyphe))
		setze(aus, 6+6*i, int(glyphe))
		setze(aus, 8+6*i, i)
	}

	return aus
}

// erweiterung legt eine Untertabelle hinter eine `ExtensionSubst` (Art 7) —
// der Versatz ist dort vier Byte breit.
func erweiterung(art int, untertabelle []byte) []byte {
	kopf := make([]byte, 8, 8+len(untertabelle))
	setze(kopf, 0, 1)
	setze(kopf, 2, art)
	binary.BigEndian.PutUint32(kopf[4:], uint32(len(kopf)))

	return append(kopf, untertabelle...)
}
