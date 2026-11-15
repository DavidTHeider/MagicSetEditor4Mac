package service

// **Die Fehlerwege des ZIP-Verzeichnisses** — jeder Lesefehler macht das
// Archiv unlesbar, keiner hält die Ansicht an.

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"
)

// stotternd liest wie ein bytes.Reader und scheitert ab dem n-ten Ruf.
type stotternd struct {
	roh  *bytes.Reader
	rufe int
	ab   int
}

var errStottert = errors.New("die Platte stottert")

func (s *stotternd) ReadAt(ziel []byte, stelle int64) (int, error) {
	s.rufe++
	if s.rufe >= s.ab {
		return 0, errStottert
	}

	return s.roh.ReadAt(ziel, stelle)
}

func TestJederLesefehlerImVerzeichnisendeKommtAn(t *testing.T) {
	t.Parallel()

	roh := handzip(nil, []handeintrag{{name: "a"}}, 0xffff, true)

	// Ruf 1: das Ende; Ruf 2: der zip64-Verweis; Ruf 3: das zip64-Ende.
	for ab := 1; ab <= 3; ab++ {
		quelle := &stotternd{roh: bytes.NewReader(roh), ab: ab}
		if _, _, fehler := verzeichnisLesen(quelle, int64(len(roh)), 10); !errors.Is(fehler, errStottert) {
			t.Errorf("ab Ruf %d: %v", ab, fehler)
		}
	}
}

func TestUngereimteEndenSindUnlesbar(t *testing.T) {
	t.Parallel()

	allein := handzip(nil, nil, 0xffff, false)

	// **Ein zip64-Merker ohne Verweis davor** fällt auf das alte Ende zurück
	// — und dessen Zahl passt dann nicht.
	ohneVerweis := handzip(nil, []handeintrag{{name: "a"}}, 0xffff, false)

	riesig := handzip(nil, []handeintrag{{name: "a"}}, 0xffff, true)
	ende := len(handverzeichnis([]handeintrag{{name: "a"}}))
	binary.LittleEndian.PutUint64(riesig[ende+40:], 1<<63)

	fern := handzip(nil, []handeintrag{{name: "a"}}, 0xffff, true)
	binary.LittleEndian.PutUint64(fern[len(fern)-22-12:], 1<<40)

	for name, roh := range map[string][]byte{
		"nur ein Ende mit zip64-Merker": allein, "ohne Verweis": ohneVerweis,
		"Verzeichnis über 63 Bit": riesig, "zip64-Ende hinter der Datei": fern,
	} {
		if _, _, fehler := verzeichnisLesen(bytes.NewReader(roh), int64(len(roh)), 10); fehler == nil {
			t.Errorf("%s: gelesen", name)
		}
	}
}

// **Ein Ende, das das Verzeichnis zu klein nennt**, ergäbe einen Vorsatz; steht
// an Versatz null ein Kopf, gilt null.
func TestEinFalscherVorsatzFaelltAufNullZurueck(t *testing.T) {
	t.Parallel()

	roh := handzip(nil, []handeintrag{{name: "a"}, {name: "b"}}, 2, false)
	groesse := binary.LittleEndian.Uint32(roh[len(roh)-10:])
	binary.LittleEndian.PutUint32(roh[len(roh)-10:], groesse-10)

	gelesen, _, fehler := verzeichnisLesen(bytes.NewReader(roh), int64(len(roh)), 10)
	if fehler != nil || len(gelesen) != 2 {
		t.Errorf("%v %v", gelesen, fehler)
	}
}

func TestEinKopfDerMehrVersprichtAlsDastehtIstUnlesbar(t *testing.T) {
	t.Parallel()

	for name, einer := range map[string]handeintrag{
		"Name":      {name: "a", namenslaenge: 4000},
		"Kommentar": {name: "a", kommentarlaenge: 4000},
	} {
		roh := handzip(nil, []handeintrag{einer}, 1, false)
		if _, _, fehler := verzeichnisLesen(bytes.NewReader(roh), int64(len(roh)), 10); fehler == nil {
			t.Errorf("%s: gelesen", name)
		}
	}
}

// **Die Größe steht im zip64-Feld, auch hinter einem anderen** — und fehlt
// sie, bleibt die volle alte.
func TestDieZip64GroesseWirdGesucht(t *testing.T) {
	t.Parallel()

	fremd := append(binary.LittleEndian.AppendUint16(nil, 0x5455), 1, 0, 0)
	zip64 := binary.LittleEndian.AppendUint64(append(binary.LittleEndian.AppendUint16(nil, 0x0001), 8, 0), 77)

	for name, fall := range map[string]struct {
		extra []byte
		soll  int64
	}{
		"hinter einem fremden Feld": {append(append([]byte{}, fremd...), zip64...), 77},
		"abgeschnitten":             {[]byte{0x01, 0x00, 0x20, 0x00, 1, 2}, 0xffffffff},
		"ohne Feld":                 {[]byte{}, 0xffffffff},
	} {
		roh := handzip(nil, []handeintrag{{name: "a", groesse: 0xffffffff, extra: fall.extra}}, 1, false)

		gelesen, _, fehler := verzeichnisLesen(bytes.NewReader(roh), int64(len(roh)), 10)
		if fehler != nil || gelesen[0].Groesse != fall.soll {
			t.Errorf("%s: %v %v", name, gelesen, fehler)
		}
	}
}

// **Ein zip64-Feld ohne die nötigen acht Bytes ist ein kaputter Kopf** — so
// weist es `archive/zip` ab (Go 1.27.1, `reader.go`, `readDirectoryHeader`:
// `ErrFormat`, wenn die alte Größe voll ist und das Feld zu kurz). Hier stand
// bis zum Nachlauf über UX-13 die volle alte Größe als Anzeige.
func TestEinZuKurzesZip64FeldIstUnlesbar(t *testing.T) {
	t.Parallel()

	kurz := append(binary.LittleEndian.AppendUint16(nil, 0x0001), 4, 0, 1, 2, 3, 4)
	roh := handzip(nil, []handeintrag{{name: "a", groesse: 0xffffffff, extra: kurz}}, 1, false)

	if _, _, fehler := verzeichnisLesen(bytes.NewReader(roh), int64(len(roh)), 10); fehler == nil {
		t.Error("ein zu kurzes zip64-Feld wurde gelesen")
	}
}
