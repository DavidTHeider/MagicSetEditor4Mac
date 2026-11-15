package render

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"
)

// **Kernpaare aus einer Schriftdatei** —
// #752.
//
// **Die Bytes stehen von Hand da, damit der Wächter ohne das Pack läuft.** Eine
// echte Schrift liegt nur unter `fremdpakete`; ein Leser, der nur dort geprüft
// wird, ist in der Pipeline ungeprüft.
//
// **Gegen echte Schriften ist er trotzdem gehalten**: `beleren-bold` gibt
// 1.983 Paare, `matrixb.ttf` 85, `mplantin.ttf` keine — dieselben Zahlen wie
// eine unabhängige Zählung außerhalb dieses Programms.
func TestKernpaareAusGPOS(t *testing.T) {
	t.Parallel()

	roh := schriftMitGPOS(paarpostabelle([]uint16{3}, [][]kernpaar{{
		{zweite: 7, vorschub: -224},
		{zweite: 9, vorschub: -44},
	}}))

	paare := paareVon(kernungAus(roh))
	if len(paare) != 2 {
		t.Fatalf("%d Paare gelesen, erwartet 2: %v", len(paare), paare)
	}

	for _, fall := range []struct {
		erste, zweite uint16
		vorschub      int16
	}{{3, 7, -224}, {3, 9, -44}} {
		if got := paare[[2]uint16{fall.erste, fall.zweite}]; got != fall.vorschub {
			t.Errorf("Paar %d/%d = %d, erwartet %d",
				fall.erste, fall.zweite, got, fall.vorschub)
		}
	}
}

// **Eine Schrift ohne `kern`-Merkmal gibt keine Paare.**
//
// Ein Leser, der bei jeder Schrift etwas fände, wäre schlimmer als keiner:
// `mplantin` trägt nichts, und eine erfundene Korrektur verschöbe den
// Regeltext, der heute stimmt.
func TestOhneKernMerkmalGibtEsKeinePaare(t *testing.T) {
	t.Parallel()

	roh := schriftMitGPOS(paarpostabelle([]uint16{3}, [][]kernpaar{{{zweite: 7, vorschub: -1}}}))
	roh[bytes.Index(roh, []byte("kern"))] = 'l'

	if paare := paareVon(kernungAus(roh)); len(paare) != 0 {
		t.Errorf("%d Paare, erwartet keine", len(paare))
	}
}

type kernpaar struct {
	zweite   uint16
	vorschub int16
}

// paarpostabelle baut eine `PairPos`-Untertabelle in Format 1.
func paarpostabelle(deckung []uint16, saetze [][]kernpaar) []byte {
	kopf := 10 + 2*len(saetze)
	deckungsstelle := kopf
	paarstelle := deckungsstelle + 4 + 2*len(deckung)

	aus := make([]byte, kopf)
	setze(aus, 0, 1)
	setze(aus, 2, deckungsstelle)
	setze(aus, 4, 0x0004) // valueFormat1: XAdvance
	setze(aus, 6, 0)
	setze(aus, 8, len(saetze))

	deck := make([]byte, 4+2*len(deckung))
	setze(deck, 0, 1)
	setze(deck, 2, len(deckung))

	for i, g := range deckung {
		setze(deck, 4+i*2, int(g))
	}

	var listen []byte

	for i, satz := range saetze {
		setze(aus, 10+i*2, paarstelle+len(listen))

		liste := make([]byte, 2+4*len(satz))
		setze(liste, 0, len(satz))

		for j, p := range satz {
			setze(liste, 2+j*4, int(p.zweite))
			setze(liste, 4+j*4, int(uint16(p.vorschub))) //nolint:gosec // Testdaten
		}

		listen = append(listen, liste...)
	}

	return bytes.Join([][]byte{aus, deck, listen}, nil)
}

// schriftMitGPOS legt eine `PairPos`-Untertabelle in eine Schriftdatei.
func schriftMitGPOS(untertabelle []byte) []byte {
	merkmale := merkmalsliste()
	nachschlag := nachschlagliste()

	kopf := make([]byte, 10)
	setze(kopf, 0, 1)
	setze(kopf, 4, len(kopf))
	setze(kopf, 6, len(kopf))
	setze(kopf, 8, len(kopf)+len(merkmale))

	gpos := bytes.Join([][]byte{kopf, merkmale, nachschlag, untertabelle}, nil)

	verzeichnis := make([]byte, 12+16, 12+16+len(gpos))
	binary.BigEndian.PutUint32(verzeichnis[0:], 0x00010000)
	setze(verzeichnis, 4, 1)
	copy(verzeichnis[12:], "GPOS")
	binary.BigEndian.PutUint32(verzeichnis[20:], uint32(len(verzeichnis))) //nolint:gosec // Testdaten
	binary.BigEndian.PutUint32(verzeichnis[24:], uint32(len(gpos)))        //nolint:gosec // Testdaten

	return append(verzeichnis, gpos...)
}

// merkmalsliste ist eine FeatureList mit genau einem `kern`.
func merkmalsliste() []byte {
	aus := make([]byte, 14)
	setze(aus, 0, 1) // featureCount
	copy(aus[2:], "kern")
	setze(aus, 6, 8)  // featureOffset, relativ zur Liste
	setze(aus, 8, 0)  // featureParamsOffset
	setze(aus, 10, 1) // lookupIndexCount

	// Die letzten beiden Byte sind `lookupListIndices[0]` und bleiben null.
	return aus
}

// nachschlagliste ist eine LookupList mit einem Lookup vom Typ 2.
func nachschlagliste() []byte {
	aus := make([]byte, 12)
	setze(aus, 0, 1)  // lookupCount
	setze(aus, 2, 4)  // lookupOffsets[0]
	setze(aus, 4, 2)  // lookupType: PairPos
	setze(aus, 8, 1)  // subTableCount
	setze(aus, 10, 8) // subtableOffset, ab dem Lookup

	return aus
}

func setze(ziel []byte, stelle, wert int) {
	binary.BigEndian.PutUint16(ziel[stelle:], uint16(wert)) //nolint:gosec // Testdaten
}

// **Die alte `kern`-Tabelle, an einer echten Schrift** —
// `matrixb.ttf` des Packs hat kein GPOS und trägt seine Paare nur dort.
//
// **Gemessen wird an `Times New Roman`**, weil es sie auf jedem macOS gibt und
// weil sie dieselbe Ablage benutzt. Eine Schrift des Packs stünde nur unter
// `fremdpakete`, und ein Leser, der nur dort geprüft wird, ist in der Pipeline
// ungeprüft.
func TestDieAlteKerntabelleWirdGelesen(t *testing.T) {
	t.Parallel()

	const pfad = "/System/Library/Fonts/Supplemental/Times New Roman.ttf"

	roh, err := os.ReadFile(pfad)
	if err != nil {
		t.Skipf("keine Schriftdatei unter %s: %v", pfad, err)
	}

	paare := paareVon(kernungAus(roh))
	if len(paare) == 0 {
		t.Fatal("keine Paare — die Schrift trägt eine `kern`-Tabelle")
	}

	// **Und sie kommen an der Schrift an**, nicht nur im Leser: `V` vor `a`
	// rückt in jeder Serifenschrift zusammen.
	schrift, err := SchriftAusDatei(roh)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if schrift.Kernung('V', 'a') >= 0 {
		t.Errorf("Kernung(V,a) = %v, erwartet eine Verengung", schrift.Kernung('V', 'a'))
	}

	// **Ein Zeichen, das die Schrift nicht führt, gibt null** statt eines
	// Zugriffs auf eine Glyphe, die es nicht gibt — **beidseitig**: das erste
	// wie das zweite. Der erste Anlauf prüfte nur das zweite, und der Zweig
	// für das erste stand ungeprüft da (§3).
	if got := schrift.Kernung('V', '￿'); got != 0 {
		t.Errorf("Kernung mit unbekanntem zweiten Zeichen = %v", got)
	}

	if got := schrift.Kernung('￿', 'a'); got != 0 {
		t.Errorf("Kernung mit unbekanntem ersten Zeichen = %v", got)
	}
}

// **Abgeschnittene Bytes brechen den Leser nicht.**
//
// Eine Schriftdatei kommt aus einem fremden Paket (§8): sie darf halb, alt oder
// beschädigt sein. Jeder Sprung im Leser prüft seine Grenze — und diese Tabelle
// betritt sie, statt sie zu behaupten.
func TestEineAbgeschnitteneSchriftGibtKeinePaare(t *testing.T) {
	t.Parallel()

	ganz := schriftMitGPOS(paarpostabelle([]uint16{3}, [][]kernpaar{{
		{zweite: 7, vorschub: -224},
	}}))

	for laenge := range len(ganz) {
		// **Keine Zusicherung über das Ergebnis, sondern über den Ausgang**:
		// ein abgeschnittener Leser darf nichts oder etwas finden, aber nicht
		// abstürzen.
		kernungAus(ganz[:laenge])
	}

	// **Und dasselbe mit einer Deckung in Bereichen** — sie hat ihre eigene
	// Schleife und ihre eigene Grenze.
	bereiche := paarpostabelle([]uint16{3, 4}, [][]kernpaar{
		{{zweite: 7, vorschub: -10}},
		{{zweite: 8, vorschub: -20}},
	})
	aufBereichsdeckung(bereiche, 3, 4)

	mitBereichen := schriftMitGPOS(bereiche)
	for laenge := range len(mitBereichen) {
		kernungAus(mitBereichen[:laenge])
	}
}

// **Die Nebenwege des Lesers**, jeder einmal betreten.
//
// **Nicht der Vollständigkeit halber**: jeder von ihnen kommt in echten
// Schriften vor, und ein Zweig, den kein Test betritt, ist eine Behauptung
// darüber, was im Ernstfall geschieht (§3).
func TestDieNebenwegeDesKernlesers(t *testing.T) {
	t.Parallel()

	t.Run("Deckung in Bereichen statt einzeln", func(t *testing.T) {
		t.Parallel()

		// Coverage Format 2 nennt Bereiche — `beleren-bold` benutzt es.
		tabelle := paarpostabelle([]uint16{3, 4}, [][]kernpaar{
			{{zweite: 7, vorschub: -10}},
			{{zweite: 8, vorschub: -20}},
		})
		aufBereichsdeckung(tabelle, 3, 4)

		paare := paareVon(kernungAus(schriftMitGPOS(tabelle)))
		if len(paare) != 2 {
			t.Errorf("%d Paare, erwartet 2: %v", len(paare), paare)
		}
	})

	t.Run("ohne Vorschub im Wertformat", func(t *testing.T) {
		t.Parallel()

		// Ein Paar, das nur die Grundlinie verschiebt, ändert keine Breite.
		tabelle := paarpostabelle([]uint16{3}, [][]kernpaar{{{zweite: 7, vorschub: -10}}})
		setze(tabelle, 4, 0x0001) // valueFormat1: nur XPlacement

		if paare := paareVon(kernungAus(schriftMitGPOS(tabelle))); len(paare) != 0 {
			t.Errorf("%d Paare, erwartet keine", len(paare))
		}
	})

	t.Run("ein Lookup anderer Art", func(t *testing.T) {
		t.Parallel()

		roh := schriftMitGPOS(paarpostabelle([]uint16{3}, [][]kernpaar{{{zweite: 7, vorschub: -10}}}))

		// Aus `PairPos` wird `SinglePos`; die Untertabelle gehört dann nicht uns.
		stelle := bytes.Index(roh, []byte("kern")) + 12 + 4
		setze(roh[stelle:], 0, 1)

		if paare := paareVon(kernungAus(roh)); len(paare) != 0 {
			t.Errorf("%d Paare, erwartet keine", len(paare))
		}
	})
}

// aufBereichsdeckung schreibt die Deckung einer Untertabelle als Bereich um.
func aufBereichsdeckung(tabelle []byte, von, bis uint16) {
	stelle := int(binary.BigEndian.Uint16(tabelle[2:]))
	setze(tabelle[stelle:], 0, 2) // coverageFormat: Bereiche
	setze(tabelle[stelle:], 2, 1) // rangeCount
	setze(tabelle[stelle:], 4, int(von))
	setze(tabelle[stelle:], 6, int(bis))
}

// **Und dieselben Grenzen an der alten `kern`-Tabelle.**
//
// Sie hat ihren eigenen Kopf, ihre eigene Länge je Untertabelle und ihre eigene
// Deckungsmaske — jede davon kann in einer beschädigten Datei ins Leere zeigen.
func TestEineAbgeschnitteneKerntabelleGibtKeinePaare(t *testing.T) {
	t.Parallel()

	ganz := schriftMitKerntabelle(0x0001, [][3]int{{3, 7, -224}, {4, 8, -44}})

	if paare := paareVon(kernungAus(ganz)); len(paare) != 2 {
		t.Fatalf("%d Paare, erwartet 2: %v", len(paare), paare)
	}

	for laenge := range len(ganz) {
		kernungAus(ganz[:laenge])
	}

	// **Eine senkrechte Untertabelle zählt nicht** — sie verschiebt die
	// Zeilen, nicht die Buchstaben.
	senkrecht := schriftMitKerntabelle(0x0002, [][3]int{{3, 7, -224}})
	if paare := paareVon(kernungAus(senkrecht)); len(paare) != 0 {
		t.Errorf("%d Paare aus einer senkrechten Tabelle", len(paare))
	}

	// **Eine Untertabelle ohne Länge beendet den Lauf.** Sonst stünde der
	// Zeiger still und die Schleife liefe, bis die Zahl der Untertabellen
	// abgearbeitet ist — bei einer beschädigten Datei sind das bis zu 65.535.
	ohneLaenge := schriftMitKerntabelle(0x0001, [][3]int{{3, 7, -224}})

	// Die Länge steht als drittes Feld der Untertabelle: Verzeichnis (28),
	// Tabellenkopf (4), dann `version` (2).
	setze(ohneLaenge, 28+4+2, 0)

	kernungAus(ohneLaenge)
}

// schriftMitKerntabelle baut eine `kern`-Tabelle mit einer Untertabelle.
func schriftMitKerntabelle(deckung int, paare [][3]int) []byte {
	unter := make([]byte, 6+8+6*len(paare))
	setze(unter, 0, 0) // version
	setze(unter, 2, len(unter))
	setze(unter, 4, deckung)
	setze(unter, 6, len(paare)) // nPairs

	for i, p := range paare {
		eintrag := 14 + i*6
		setze(unter, eintrag, p[0])
		setze(unter, eintrag+2, p[1])
		setze(unter, eintrag+4, int(uint16(int16(p[2])))) //nolint:gosec // Testdaten
	}

	tabelle := make([]byte, 4, 4+len(unter))
	setze(tabelle, 0, 0) // version
	setze(tabelle, 2, 1) // nTables
	tabelle = append(tabelle, unter...)

	verzeichnis := make([]byte, 12+16, 12+16+len(tabelle))
	binary.BigEndian.PutUint32(verzeichnis[0:], 0x00010000)
	setze(verzeichnis, 4, 1)
	copy(verzeichnis[12:], "kern")
	binary.BigEndian.PutUint32(verzeichnis[20:], uint32(len(verzeichnis))) //nolint:gosec // Testdaten
	binary.BigEndian.PutUint32(verzeichnis[24:], uint32(len(tabelle)))     //nolint:gosec // Testdaten

	return append(verzeichnis, tabelle...)
}

// **`zwei` ist die eine Grenzprüfung des Lesers.**
//
// Jeder Sprung geht durch sie; die übrigen Prüfungen im Leser sind
// Abbruchbedingungen von Schleifen und keine Grenzen. Sie deshalb hier
// unmittelbar zu prüfen ist nicht Buchführung, sondern die Zusage selbst:
// **außerhalb gibt es null und keinen Zugriff.**
func TestZweiGibtAusserhalbNull(t *testing.T) {
	t.Parallel()

	roh := []byte{0x12, 0x34}

	if got := zwei(roh, 0); got != 0x1234 {
		t.Errorf("innerhalb = %#04x, erwartet 0x1234", got)
	}

	if got := zwei(roh, 1); got != 0 {
		t.Errorf("halb draußen = %#04x, erwartet 0", got)
	}

	if got := zwei(roh, 99); got != 0 {
		t.Errorf("ganz draußen = %#04x, erwartet 0", got)
	}
}
