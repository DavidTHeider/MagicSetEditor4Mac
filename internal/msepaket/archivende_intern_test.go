package msepaket

// **Die Wege des Verzeichnisendes** — jeder Lesefehler kommt an, jedes
// ungereimte Ende ist keines, und ein zip64-Merker ohne Verweis fällt auf die
// alten Angaben zurück, wie in `archive/zip` (#1886, umgezogen aus der
// Dateiansicht, #1885).

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

var errStottert = errors.New("die Platte stottert")

// stotternd liest wie ein bytes.Reader und scheitert ab dem n-ten Ruf.
type stotternd struct {
	roh  *bytes.Reader
	rufe int
	ab   int
}

func (s *stotternd) ReadAt(ziel []byte, stelle int64) (int, error) {
	s.rufe++
	if s.rufe >= s.ab {
		return 0, errStottert
	}

	return s.roh.ReadAt(ziel, stelle)
}

func altesEnde(zahl uint16, groesse, versatz uint32, kommentar uint16) []byte {
	ende := make([]byte, endlaenge)
	binary.LittleEndian.PutUint32(ende, endsignatur)
	binary.LittleEndian.PutUint16(ende[8:], zahl)
	binary.LittleEndian.PutUint16(ende[10:], zahl)
	binary.LittleEndian.PutUint32(ende[12:], groesse)
	binary.LittleEndian.PutUint32(ende[16:], versatz)
	binary.LittleEndian.PutUint16(ende[20:], kommentar)

	return ende
}

// mitZip64 legt vor ein volles altes Ende ein zip64-Ende samt Verweis.
func mitZip64(zahl, groesse uint64, signatur uint32) []byte {
	satz := make([]byte, zip64endelen, zip64endelen+zip64verweislen)
	binary.LittleEndian.PutUint32(satz, signatur)
	binary.LittleEndian.PutUint64(satz[32:], zahl)
	binary.LittleEndian.PutUint64(satz[40:], groesse)

	verweis := make([]byte, zip64verweislen)
	binary.LittleEndian.PutUint32(verweis, zip64verweis)
	binary.LittleEndian.PutUint32(verweis[16:], 1)

	roh := append(append(satz, verweis...), altesEnde(0xffff, 0xffffffff, 0xffffffff, 0)...)

	return roh
}

func TestJederLesefehlerAmEndeKommtAn(t *testing.T) {
	t.Parallel()

	roh := mitZip64(3, 100, zip64ende)

	// Ruf 1: das Ende; Ruf 2: der zip64-Verweis; Ruf 3: das zip64-Ende.
	for ab := 1; ab <= 3; ab++ {
		if _, fehler := LiesArchivende(&stotternd{roh: bytes.NewReader(roh), ab: ab},
			int64(len(roh))); !errors.Is(fehler, errStottert) {
			t.Errorf("ab Ruf %d: %v", ab, fehler)
		}
	}
}

func TestUngereimteEndenSindKeine(t *testing.T) {
	t.Parallel()

	riesig := mitZip64(1, 1<<63, zip64ende)
	fremd := mitZip64(1, 1, 0x12345678)

	for name, roh := range map[string][]byte{
		"ohne Signatur":                []byte("kein Ende weit und breit, nur Text"),
		"ein Kommentar über den Block": altesEnde(1, 47, 0, 400),
		"ein Verzeichnis über 63 Bit":  riesig,
		"ein zip64-Ende ohne Signatur": fremd,
	} {
		if _, fehler := LiesArchivende(bytes.NewReader(roh), int64(len(roh))); !errors.Is(fehler, errKeinArchivende) {
			t.Errorf("%s: %v", name, fehler)
		}
	}
}

// **Ein zip64-Merker ohne Verweis** — oder ohne Platz für einen — lässt die
// alten Angaben stehen; so liest es `archive/zip` auch.
func TestEinMerkerOhneVerweisBleibtBeimAltenEnde(t *testing.T) {
	t.Parallel()

	allein := altesEnde(0xffff, 47, 0, 0)
	vorn := append(make([]byte, 40, 40+len(allein)), allein...)

	for name, roh := range map[string][]byte{"ohne Platz": allein, "ohne Verweis": vorn} {
		ende, fehler := LiesArchivende(bytes.NewReader(roh), int64(len(roh)))
		if fehler != nil || ende.Zahl != 0xffff || ende.Groesse != 47 {
			t.Errorf("%s: %+v, %v", name, ende, fehler)
		}
	}
}

// **Ein zip64-Ende nennt Zahl und Größe über 16 und 32 Bit hinaus.**
func TestEinZip64EndeWirdGelesen(t *testing.T) {
	t.Parallel()

	roh := mitZip64(1<<20, 1<<33, zip64ende)

	ende, fehler := LiesArchivende(bytes.NewReader(roh), int64(len(roh)))
	if fehler != nil || ende.Zahl != 1<<20 || ende.Groesse != 1<<33 || ende.Endversatz != 0 {
		t.Errorf("%+v, %v", ende, fehler)
	}
}

// messlos ist eine Quelle, die sich nicht messen lässt.
type messlos struct{ *os.File }

func (messlos) Stat() (os.FileInfo, error) { return nil, errStottert }

func TestWasSichNichtMessenLaesstWirdNichtGelesen(t *testing.T) {
	t.Parallel()

	if _, fehler := archivLesen(messlos{}); !errors.Is(fehler, errStottert) {
		t.Errorf("archivLesen: %v", fehler)
	}

	if _, _, fehler := oeffneArchiv(filepath.Join(t.TempDir(), "fehlt.mse-style")); !errors.Is(fehler, os.ErrNotExist) {
		t.Errorf("oeffneArchiv: %v", fehler)
	}
}

// **Ist das Verzeichnis gelesen, gilt die Grenze nicht mehr** — sie gilt dem
// Verzeichnis und nicht dem Inhalt.
func TestDerZaehlerIstNachDemVerzeichnisFrei(t *testing.T) {
	t.Parallel()

	zaehler := &mitzaehlend{quelle: bytes.NewReader(make([]byte, 64)), budget: 8}

	if _, fehler := zaehler.ReadAt(make([]byte, 16), 0); !errors.Is(fehler, errUeberZusage) {
		t.Errorf("vor der Freigabe: %v", fehler)
	}

	zaehler.frei.Store(true)

	if gelesen, fehler := zaehler.ReadAt(make([]byte, 16), 0); fehler != nil || gelesen != 16 {
		t.Errorf("nach der Freigabe: %d, %v", gelesen, fehler)
	}
}
