package msepaket_test

// **Die Ressourcenregel für Archivpakete** —
// #1886.
//
// # Was hier gemessen wird
//
// **Dass ein Archiv über der Grenze abgewiesen wird, bevor sein Verzeichnis
// gelesen ist** — an drei Behauptungen des Verzeichnisendes (zu viele
// Einträge, ein zu großes Verzeichnis, dasselbe über zip64) und an einem
// Verzeichnis, das **länger ist, als das Ende sagt**: `archive/zip` liest
// Kopf um Kopf, bis einer nicht passt, und glaubte der Angabe nicht.
//
// **Und dass alles darunter wie bisher aufgeht** — gelesen wird danach mit
// denselben Werkzeugen.

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// archivende schreibt ein Verzeichnisende mit den genannten Angaben.
func archivende(zahl uint16, groesse, versatz uint32) []byte {
	ende := make([]byte, 22)
	binary.LittleEndian.PutUint32(ende, 0x06054b50)
	binary.LittleEndian.PutUint16(ende[8:], zahl)
	binary.LittleEndian.PutUint16(ende[10:], zahl)
	binary.LittleEndian.PutUint32(ende[12:], groesse)
	binary.LittleEndian.PutUint32(ende[16:], versatz)

	return ende
}

// zip64Ende schreibt ein zip64-Ende samt Verweis und ein volles altes Ende.
func zip64Ende(zahl, groesse uint64) []byte {
	var roh bytes.Buffer

	satz := make([]byte, 56)
	binary.LittleEndian.PutUint32(satz, 0x06064b50)
	binary.LittleEndian.PutUint64(satz[24:], zahl)
	binary.LittleEndian.PutUint64(satz[32:], zahl)
	binary.LittleEndian.PutUint64(satz[40:], groesse)
	roh.Write(satz)

	verweis := make([]byte, 20)
	binary.LittleEndian.PutUint32(verweis, 0x07064b50)
	binary.LittleEndian.PutUint32(verweis[16:], 1)
	roh.Write(verweis)
	roh.Write(archivende(0xffff, 0xffffffff, 0xffffffff))

	return roh.Bytes()
}

// minimalkopf ist ein Kopf des Verzeichnisses mit einem Namen aus einem
// Zeichen und sonst nichts — 47 Byte.
func minimalkopf() []byte {
	kopf := make([]byte, 46, 47)
	binary.LittleEndian.PutUint32(kopf, 0x02014b50)
	binary.LittleEndian.PutUint16(kopf[28:], 1)

	return append(kopf, 'x')
}

func ablegen(t *testing.T, roh []byte) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "gross.mse-style")
	if fehler := os.WriteFile(pfad, roh, 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	return pfad
}

func TestEinArchivUeberDerGrenzeWirdBenanntAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, roh := range map[string][]byte{
		"ein zu großes Verzeichnis":            archivende(3, uint32(msepaket.GrenzeArchivverzeichnis+1), 0),
		"zu viele Einträge über zip64":         zip64Ende(msepaket.GrenzeArchiveintraege+1, 47),
		"ein zu großes Verzeichnis über zip64": zip64Ende(1, uint64(msepaket.GrenzeArchivverzeichnis)+1),
	} {
		pfad := ablegen(t, roh)

		if _, fehler := msepaket.Oeffne(pfad); !errors.Is(fehler, msepaket.ErrArchivZuGross) {
			t.Errorf("%s — Oeffne: %v", name, fehler)
		}

		if _, fehler := msepaket.OeffneInstaller(pfad); !errors.Is(fehler, msepaket.ErrArchivZuGross) {
			t.Errorf("%s — OeffneInstaller: %v", name, fehler)
		}
	}
}

// **Ein Verzeichnis, das länger ist, als das Ende sagt**, wird nicht zu Ende
// gelesen — die Angabe nennt einen Kopf, dahinter stehen hunderttausend.
func TestEinVerzeichnisLaengerAlsAngegebenWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	roh := bytes.Repeat(minimalkopf(), 100_000)
	roh = append(roh, archivende(uint16(100_000%0x10000), 47, 0)...)

	if _, fehler := msepaket.Oeffne(ablegen(t, roh)); !errors.Is(fehler, msepaket.ErrArchivZuGross) {
		t.Errorf("Oeffne: %v", fehler)
	}
}

// **Darunter geht alles auf wie bisher** — und gelesen wird danach ohne
// Grenze, denn die gilt dem Verzeichnis und nicht dem Inhalt.
func TestEinArchivUnterDerGrenzeGehtAuf(t *testing.T) {
	t.Parallel()

	var roh bytes.Buffer

	schreiber := zip.NewWriter(&roh)
	for _, name := range []string{"style", "karte.png"} {
		datei, fehler := schreiber.Create(name)
		if fehler != nil {
			t.Fatal(fehler)
		}

		if _, fehler := datei.Write(bytes.Repeat([]byte("mse "), 70_000)); fehler != nil {
			t.Fatal(fehler)
		}
	}

	if fehler := schreiber.Close(); fehler != nil {
		t.Fatal(fehler)
	}

	paket, fehler := msepaket.Oeffne(ablegen(t, roh.Bytes()))
	if fehler != nil {
		t.Fatalf("Oeffne: %v", fehler)
	}

	t.Cleanup(func() { _ = paket.Schliessen() }) //nolint:errcheck // nur ein Lauf

	if inhalt, fehler := paket.Datei("karte.png"); fehler != nil || len(inhalt) != 280_000 {
		t.Errorf("Datei: %d Byte, %v", len(inhalt), fehler)
	}
}

// **Was kein Archiv ist, bleibt der Fehler, der es war** — die Regel macht
// aus einer kaputten Datei keine zu große.
func TestKeinArchivBleibtKeinArchiv(t *testing.T) {
	t.Parallel()

	_, fehler := msepaket.Oeffne(ablegen(t, []byte("das ist kein Archiv")))
	if fehler == nil || errors.Is(fehler, msepaket.ErrArchivZuGross) {
		t.Errorf("Oeffne: %v", fehler)
	}

	if !errors.Is(fehler, zip.ErrFormat) {
		t.Errorf("die Klasse ging verloren: %v", fehler)
	}
}
