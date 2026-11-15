package service

// **Das ZIP-Verzeichnis, gelesen bis zur Grenze und nicht weiter** — Befund 1
// des Zweitblicks über UX-13
// (#1685).
//
// Die Archive entstehen hier **von Hand**: nur so lassen sich ein kaputter
// Kopf hinter der Grenze, zip64 und ein Unix-Verweis genau setzen.

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

type handeintrag struct {
	name      string
	groesse   uint32
	zip64     uint64
	unixmodus uint32
	kaputt    bool

	// extra ersetzt das zip64-Feld; namenslaenge und kommentarlaenge
	// behaupten mehr, als dasteht.
	extra           []byte
	namenslaenge    uint16
	kommentarlaenge uint16
}

// handverzeichnis schreibt Köpfe des zentralen Verzeichnisses.
func handverzeichnis(eintraege []handeintrag) []byte {
	var roh bytes.Buffer

	for _, einer := range eintraege {
		signatur := uint32(0x02014b50)
		if einer.kaputt {
			signatur = 0x12345678
		}

		// **Beide Größen im zip64-Feld, die ausgepackte zuerst** — so liest
		// es `archive/zip`; eine allein weist es als kaputt ab.
		var extra []byte
		if einer.zip64 != 0 {
			extra = binary.LittleEndian.AppendUint16(extra, 0x0001)
			extra = binary.LittleEndian.AppendUint16(extra, 16)
			extra = binary.LittleEndian.AppendUint64(extra, einer.zip64)
			extra = binary.LittleEndian.AppendUint64(extra, einer.zip64)
		}

		if einer.extra != nil {
			extra = einer.extra
		}

		namenslaenge := uint16(len(einer.name))
		if einer.namenslaenge != 0 {
			namenslaenge = einer.namenslaenge
		}

		ersteller := uint16(20)
		if einer.unixmodus != 0 {
			ersteller |= 3 << 8
		}

		kopf := binary.LittleEndian.AppendUint32(nil, signatur)
		kopf = binary.LittleEndian.AppendUint16(kopf, ersteller)
		kopf = append(kopf, make([]byte, 14)...) // benötigt, Merker, Verfahren, Zeit, Datum, CRC
		kopf = binary.LittleEndian.AppendUint32(kopf, einer.groesse)
		kopf = binary.LittleEndian.AppendUint32(kopf, einer.groesse)
		kopf = binary.LittleEndian.AppendUint16(kopf, namenslaenge)
		kopf = binary.LittleEndian.AppendUint16(kopf, uint16(len(extra)))
		kopf = binary.LittleEndian.AppendUint16(kopf, einer.kommentarlaenge)
		kopf = append(kopf, 0, 0, 0, 0) // Scheibe, innere Merkmale
		kopf = binary.LittleEndian.AppendUint32(kopf, einer.unixmodus<<16)
		kopf = append(kopf, 0, 0, 0, 0)

		roh.Write(kopf)
		roh.WriteString(einer.name)
		roh.Write(extra)
	}

	return roh.Bytes()
}

// handzip setzt Verzeichnis und Verzeichnisende zusammen.
func handzip(vorsatz []byte, eintraege []handeintrag, zahl uint16, mitZip64 bool) []byte {
	verzeichnis := handverzeichnis(eintraege)
	roh := append(append([]byte{}, vorsatz...), verzeichnis...)

	if mitZip64 {
		ende := int64(len(roh))
		roh = binary.LittleEndian.AppendUint32(roh, 0x06064b50)
		roh = append(roh, make([]byte, 12)...)
		roh = append(roh, make([]byte, 8)...) // Scheiben
		roh = binary.LittleEndian.AppendUint64(roh, uint64(len(eintraege)))
		roh = binary.LittleEndian.AppendUint64(roh, uint64(len(eintraege)))
		roh = binary.LittleEndian.AppendUint64(roh, uint64(len(verzeichnis)))
		roh = binary.LittleEndian.AppendUint64(roh, 0)
		roh = binary.LittleEndian.AppendUint32(roh, 0x07064b50)
		roh = binary.LittleEndian.AppendUint32(roh, 0)
		roh = binary.LittleEndian.AppendUint64(roh, uint64(ende))
		roh = binary.LittleEndian.AppendUint32(roh, 1)
	}

	roh = binary.LittleEndian.AppendUint32(roh, 0x06054b50)
	roh = append(roh, 0, 0, 0, 0)
	roh = binary.LittleEndian.AppendUint16(roh, zahl)
	roh = binary.LittleEndian.AppendUint16(roh, zahl)

	groesse, versatz := uint32(len(verzeichnis)), uint32(0)
	if mitZip64 {
		groesse, versatz = 0xffffffff, 0xffffffff
	}

	roh = binary.LittleEndian.AppendUint32(roh, groesse)
	roh = binary.LittleEndian.AppendUint32(roh, versatz)

	return append(roh, 0, 0)
}

func groesseVon(t *testing.T, pfad string) int64 {
	t.Helper()

	angaben, fehler := os.Stat(pfad)
	if fehler != nil {
		t.Fatal(fehler)
	}

	return angaben.Size()
}

func zipAblegen(t *testing.T, roh []byte) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "hand.zip")
	if fehler := os.WriteFile(pfad, roh, 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	return pfad
}

// **Was hinter der Grenze liegt, wird nicht gelesen** — ein kaputter Kopf
// dort ändert nichts. `archive/zip` las das ganze Verzeichnis und scheiterte
// an ihm.
func TestHinterDerGrenzeWirdNichtGelesen(t *testing.T) {
	t.Parallel()

	eintraege := []handeintrag{{name: "a"}, {name: "b"}, {name: "c"}, {name: "d", kaputt: true}, {name: "e"}}
	pfad := zipAblegen(t, handzip(nil, eintraege, 5, false))

	gelesen, _, unlesbar := archivEintraege(pfad, groesseVon(t, pfad), 3)
	if unlesbar || len(gelesen) != 3 || gelesen[2].Name != "c" {
		t.Errorf("%v %v", gelesen, unlesbar)
	}

	// **Vor der Grenze ist ein kaputter Kopf ein kaputtes Archiv.**
	if _, _, unlesbar := archivEintraege(pfad, groesseVon(t, pfad), 10); !unlesbar {
		t.Error("ein kaputter Kopf vor der Grenze blieb unbemerkt")
	}
}

func TestZip64UndVorsatzWerdenGelesen(t *testing.T) {
	t.Parallel()

	eintraege := []handeintrag{
		{name: "riesig.png", groesse: 0xffffffff, zip64: 5_000_000_000},
		{name: "verweis", unixmodus: 0o120777},
		{name: "ordner/", unixmodus: 0o040755},
		{name: "../aussen"},
	}
	pfad := zipAblegen(t, handzip([]byte("#!/bin/sh vorsatz\n"), eintraege, 0xffff, true))

	gelesen, unsicher, unlesbar := archivEintraege(pfad, groesseVon(t, pfad), 10)
	if unlesbar || len(gelesen) != 4 {
		t.Fatalf("%v %v", gelesen, unlesbar)
	}

	if gelesen[0].Groesse != 5_000_000_000 {
		t.Errorf("Größe aus zip64: %d", gelesen[0].Groesse)
	}

	if !gelesen[1].Unsicher || !gelesen[2].Verzeichnis || !gelesen[3].Unsicher || unsicher != 2 {
		t.Errorf("%+v, unsicher %d", gelesen, unsicher)
	}
}

// **Ein überlanger Name wird gezeigt, aber nicht ganz behalten.**
func TestEinUeberlangerNameWirdGekuerzt(t *testing.T) {
	t.Parallel()

	lang := string(bytes.Repeat([]byte("ä"), 2000))
	pfad := zipAblegen(t, handzip(nil, []handeintrag{{name: lang}}, 1, false))

	gelesen, _, _ := archivEintraege(pfad, groesseVon(t, pfad), 10)
	if len(gelesen) != 1 || len(gelesen[0].Name) > namensgrenze+len("…") {
		t.Errorf("Länge %d", len(gelesen[0].Name))
	}
}

func TestEinArchivOhneEndeIstUnlesbar(t *testing.T) {
	t.Parallel()

	for name, roh := range map[string][]byte{
		"leer":                  nil,
		"kein Ende":             bytes.Repeat([]byte("x"), 70_000),
		"Kommentar zu lang":     append(handzip(nil, nil, 0, false)[:20], 0xff, 0xff),
		"Verzeichnis außerhalb": verschobenesVerzeichnis(),
		"zip64 ohne Ende":       kaputtesZip64(),
	} {
		if _, _, unlesbar := archivEintraege(zipAblegen(t, roh), int64(len(roh)), 10); !unlesbar {
			t.Errorf("%s: gelesen", name)
		}
	}

	if _, _, unlesbar := archivEintraege(filepath.Join(t.TempDir(), "fehlt.zip"), 0, 10); !unlesbar {
		t.Error("eine fehlende Datei wurde gelesen")
	}
}

// **Ein gewöhnliches Archiv aus `archive/zip`** liest sich wie bisher.
func TestEinGeschriebenesArchivLiestSichWieBisher(t *testing.T) {
	t.Parallel()

	var roh bytes.Buffer

	schreiber := zip.NewWriter(&roh)
	for _, name := range []string{"set", "bild.png", "ordner/"} {
		if _, fehler := schreiber.Create(name); fehler != nil {
			t.Fatal(fehler)
		}
	}

	if fehler := schreiber.SetComment("ein Kommentar"); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := schreiber.Close(); fehler != nil {
		t.Fatal(fehler)
	}

	gelesen, unsicher, unlesbar := archivEintraege(zipAblegen(t, roh.Bytes()), int64(roh.Len()), 10)
	if unlesbar || unsicher != 0 || len(gelesen) != 3 || !gelesen[2].Verzeichnis {
		t.Errorf("%+v %d %v", gelesen, unsicher, unlesbar)
	}
}

func verschobenesVerzeichnis() []byte {
	roh := handzip(nil, []handeintrag{{name: "a"}}, 1, false)
	// Ein Verzeichnis, größer als die Datei, begänne vor ihrem Anfang.
	binary.LittleEndian.PutUint32(roh[len(roh)-10:], 1_000_000)

	return roh
}

func kaputtesZip64() []byte {
	roh := handzip(nil, []handeintrag{{name: "a"}}, 0xffff, true)
	// Die Signatur des zip64-Endes zerstören.
	ende := len(handverzeichnis([]handeintrag{{name: "a"}}))
	roh[ende] = 0

	return roh
}
