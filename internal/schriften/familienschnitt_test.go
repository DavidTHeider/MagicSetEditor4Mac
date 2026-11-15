package schriften_test

// **Fett ist der Schnitt „Bold“ der Familie — auch wenn kein Name das so
// zusammenschreibt.**
//
// MSE reicht für `name: Matrix` samt `weight: bold` die **Familie** und das
// **Gewicht** an wxWidgets weiter
// (`wxFont(size_i, wxFONTFAMILY_DEFAULT, style_i, weight_i, underline(), familyName)`,
// `Font::toWxFont`, `src/data/font.cpp`, Zeilen 83–110, Fassung f9d9356 vom
// 2021-07-06), und das System nimmt den fetten Schnitt dieser Familie. Im Pack
// liegt er so:
//
//	MATRIX.TTF    Familie "Matrix"   Unterfamilie "Medium"
//	matrixb.ttf   Familie "Matrix"   Unterfamilie "Bold"   voller Name "Matrix-Bold"
//
// Diese Suche fragt nach „Matrix Bold“ (0023). Keine Datei heißt so, kein
// voller Name lautet so — und sie fiel auf `MATRIX.TTF` zurück. Der Name auf
// jeder Archenemy-Karte war dadurch **5 Prozent schmaler** als bei MSE2
// (101 statt 106 Bildpunkte für „All in Good Time“, F02,
// #2210).

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"unicode/utf16"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften"
)

func TestFettFindetDenSchnittBoldDerFamilie(t *testing.T) {
	ort := t.TempDir()
	legeBenannt(t, ort, "Probe.ttf", map[uint16]string{1: "Probe", 2: "Medium", 4: "Probe"})
	legeBenannt(t, ort, "probeb.ttf", map[uint16]string{1: "Probe", 2: "Bold", 4: "Probe-Bold"})

	fundort, err := schriften.Neue([]string{ort}).Fundort("Probe", true)
	if err != nil {
		t.Fatalf("suchen: %v", err)
	}

	if filepath.Base(fundort) != "probeb.ttf" {
		t.Errorf("gefunden wurde %s, erwartet der Schnitt Bold der Familie: probeb.ttf", filepath.Base(fundort))
	}
}

// **Dasselbe in den Paketen** — dort liegt `matrixb.ttf` im Pack, und der
// Paketindex liest dieselben Namen.
func TestFettFindetDenSchnittBoldAuchInDenPaketen(t *testing.T) {
	paket := filepath.Join(t.TempDir(), "schriften.mse-include")
	if err := os.MkdirAll(paket, 0o750); err != nil {
		t.Fatal(err)
	}

	legeBenannt(t, paket, "Probe.ttf", map[uint16]string{1: "Probe", 2: "Medium", 4: "Probe"})
	legeBenannt(t, paket, "probeb.ttf", map[uint16]string{1: "Probe", 2: "Bold", 4: "Probe-Bold"})

	fundort, err := schriften.Neue(nil).MitPaketen(filepath.Dir(paket)).Fundort("Probe", true)
	if err != nil {
		t.Fatalf("suchen: %v", err)
	}

	if filepath.Base(fundort) != "probeb.ttf" {
		t.Errorf("gefunden wurde %s, erwartet probeb.ttf", filepath.Base(fundort))
	}
}

// **Und aufrecht bleibt es beim aufrechten Schnitt** — auch wenn der fette im
// Verzeichnis vorn steht.
//
// Beide Schnitte führen denselben Familiennamen. Bis hierher gewann der zuerst
// gefundene, und `NamenAus` nannte das folgenlos, „solange eine Familie nur mit
// einem Schnitt im Pack liegt" — `Matrix` liegt dort mit zweien. MSE reicht
// `wxFONTWEIGHT_NORMAL` weiter, und das System nimmt den aufrechten.
func TestAufrechtBleibtBeimAufrechtenSchnitt(t *testing.T) {
	ort := t.TempDir()
	legeBenannt(t, ort, "probe-b.ttf", map[uint16]string{1: "Probe", 2: "Bold", 4: "Probe-Bold"})
	legeBenannt(t, ort, "probe-n.ttf", map[uint16]string{1: "Probe", 2: "Regular", 4: "Probe Normal"})

	suche := schriften.Neue([]string{ort})

	faelle := []struct {
		name, gesucht, datei string
		fett                 bool
	}{
		{"aufrecht über die Familie", "Probe", "probe-n.ttf", false},
		{"der volle Name des fetten Schnitts", "Probe-Bold", "probe-b.ttf", false},
		{"Familie und Schnitt zusammen", "Probe Bold", "probe-b.ttf", false},
		{"fett über die Familie", "Probe", "probe-b.ttf", true},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			fundort, err := suche.Fundort(fall.gesucht, fall.fett)
			if err != nil {
				t.Fatalf("suchen: %v", err)
			}

			if filepath.Base(fundort) != fall.datei {
				t.Errorf("gefunden wurde %s, erwartet %s", filepath.Base(fundort), fall.datei)
			}
		})
	}
}

// legeBenannt legt Go Regular mit einer **eigenen Namenstabelle** ab — eine
// echte Schrift, deren Namen der Test bestimmt.
func legeBenannt(t *testing.T, ort, dateiname string, namen map[uint16]string) {
	t.Helper()

	roh := mitNamenstabelle(goregular.TTF, namenstabelle(namen))
	if err := os.WriteFile(filepath.Join(ort, dateiname), roh, 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}
}

// namenstabelle baut eine `name`-Tabelle im Format 0, Plattform Windows,
// Unicode BMP, Sprache en-US — die Einträge nach Kennung sortiert, wie die
// Spezifikation es verlangt.
func namenstabelle(namen map[uint16]string) []byte {
	kennungen := make([]uint16, 0, len(namen))
	for kennung := range namen {
		kennungen = append(kennungen, kennung)
	}

	slices.Sort(kennungen)

	texte := make([][]byte, len(kennungen))
	gesamt := 0

	for i, kennung := range kennungen {
		for _, einheit := range utf16.Encode([]rune(namen[kennung])) {
			texte[i] = binary.BigEndian.AppendUint16(texte[i], einheit)
		}

		gesamt += len(texte[i])
	}

	kopflaenge := 6 + 12*len(kennungen)
	tabelle := make([]byte, kopflaenge, kopflaenge+gesamt)
	binary.BigEndian.PutUint16(tabelle[2:], uint16(len(kennungen))) //nolint:gosec // Testdaten
	binary.BigEndian.PutUint16(tabelle[4:], uint16(kopflaenge))     //nolint:gosec // Testdaten

	for i, kennung := range kennungen {
		satz := tabelle[6+12*i:]
		binary.BigEndian.PutUint16(satz[0:], 3)
		binary.BigEndian.PutUint16(satz[2:], 1)
		binary.BigEndian.PutUint16(satz[4:], 0x409)
		binary.BigEndian.PutUint16(satz[6:], kennung)
		binary.BigEndian.PutUint16(satz[8:], uint16(len(texte[i])))            //nolint:gosec // Testdaten
		binary.BigEndian.PutUint16(satz[10:], uint16(len(tabelle)-kopflaenge)) //nolint:gosec // Testdaten

		tabelle = append(tabelle, texte[i]...)
	}

	return tabelle
}

// mitNamenstabelle ersetzt die `name`-Tabelle einer echten Schriftdatei —
// das Verzeichnis neu und nach Kennung sortiert, jede Tabelle auf vier Byte
// ausgerichtet, wie `sfnt` es verlangt.
func mitNamenstabelle(schrift, tabelle []byte) []byte {
	type eintrag struct {
		kennung string
		daten   []byte
	}

	anzahl := int(binary.BigEndian.Uint16(schrift[4:]))

	eintraege := make([]eintrag, 0, anzahl+1)
	eintraege = append(eintraege, eintrag{"name", tabelle})

	for i := range anzahl {
		stelle := 12 + 16*i
		if kennung := string(schrift[stelle : stelle+4]); kennung != "name" {
			versatz := binary.BigEndian.Uint32(schrift[stelle+8:])
			laenge := binary.BigEndian.Uint32(schrift[stelle+12:])
			eintraege = append(eintraege, eintrag{kennung, schrift[versatz : versatz+laenge]})
		}
	}

	slices.SortFunc(eintraege, func(a, b eintrag) int { return strings.Compare(a.kennung, b.kennung) })

	aus := make([]byte, 12+16*len(eintraege))
	copy(aus, schrift[:4])
	binary.BigEndian.PutUint16(aus[4:], uint16(len(eintraege))) //nolint:gosec // Testdaten

	for i, e := range eintraege {
		for len(aus)%4 != 0 {
			aus = append(aus, 0)
		}

		stelle := 12 + 16*i
		copy(aus[stelle:], e.kennung)
		binary.BigEndian.PutUint32(aus[stelle+8:], uint32(len(aus)))      //nolint:gosec // Testdaten
		binary.BigEndian.PutUint32(aus[stelle+12:], uint32(len(e.daten))) //nolint:gosec // Testdaten

		aus = append(aus, e.daten...)
	}

	return aus
}
