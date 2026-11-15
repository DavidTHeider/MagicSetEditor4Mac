// **Die Fehlerwege von `archivBauen`** — Werkstück 2 von Phase 3.3
// (#1164).
//
// # Warum dieser Lauf im Paket steht und nicht daneben
//
// `archivBauen` ist paketintern, und das soll es bleiben: es ist ein Schritt
// des Schreibwegs und keine Fähigkeit für Aufrufer. Ein exportierter
// Testeinstieg daneben wäre eine Öffnung, die nur der Prüfung dient — und die
// nächste Änderung fände sie und benutzte sie.
//
// # Warum ein Schreiber und keine Datei
//
// `archivBauen` nimmt einen [io.Writer]. Anders wären seine Fehlerwege nicht
// zu betreten: ein Griff auf eine eben angelegte Datei scheitert beim
// Schreiben nicht auf eine Art, die sich herbeiführen ließe (§3). Dieselbe
// Gestalt und derselbe Grund wie bei `sicherung.hineinschreiben`.
package msepaket

import (
	"archive/zip"
	"bytes"
	"errors"
	"testing"
)

// **Ein Schreiber, der ablehnt, bringt jeden Weg des Archivbaus zum Scheitern.**
//
// **Zwei Stellen, zwei Läufe**: der Deckel läuft entweder schon beim
// Übernehmen des fremden Eintrags voll oder erst beim Anlegen des neuen. Wer
// nur einen prüfte, prüfte den anderen nicht.
func TestEinAblehnenderSchreiberBringtDenArchivbauZumScheitern(t *testing.T) {
	t.Parallel()

	lesen := probearchivImSpeicher(t)

	// **Die zweite Zahl liegt knapp über dem übernommenen Eintrag** — er
	// braucht rund 10.060 Byte samt seinem Kopf. Danach ist Schluss, und das
	// trifft das Anlegen des neuen Eintrags samt dem Verzeichnis am Ende.
	for name, deckel := range map[string]int{
		"beim Übernehmen": 0,
		"beim Anlegen":    10_100,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := archivBauen(&voll{uebrig: deckel}, lesen, "style", []byte("x\n")); fehler == nil {
				t.Fatal("der ablehnende Schreiber nahm alles an")
			}
		})
	}
}

// probearchivImSpeicher baut ein Archiv mit zwei Einträgen.
func probearchivImSpeicher(t *testing.T) *zip.Reader {
	t.Helper()

	var gebaut bytes.Buffer

	packer := zip.NewWriter(&gebaut)

	// **Der fremde Eintrag ist groß und ungepackt**, und beides mit Grund:
	// `zip.Writer` puffert über `bufio` (4 KiB), und ein kurzer Eintrag käme
	// deshalb gar nicht bis zum Schreiber — der Fehler fiele erst beim
	// Schließen an, und `Copy` bliebe unbetreten. Am 2026-08-26 gemessen: mit
	// neunzehn Byte ja, mit zehntausend nein.
	for _, eintrag := range []struct {
		name     string
		inhalt   []byte
		verfahre uint16
	}{
		{"style", []byte("mse version: 2.0.0\n"), zip.Deflate},
		{"unterordner/skript", langerInhalt(), zip.Store},
	} {
		strom, err := packer.CreateHeader(&zip.FileHeader{
			Name: eintrag.name, Method: eintrag.verfahre,
		})
		if err != nil {
			t.Fatalf("Eintrag: %v", err)
		}

		if _, fehler := strom.Write(eintrag.inhalt); fehler != nil {
			t.Fatalf("schreiben: %v", fehler)
		}
	}

	if err := packer.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	lesen, err := zip.NewReader(bytes.NewReader(gebaut.Bytes()), int64(gebaut.Len()))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	return lesen
}

// voll nimmt eine begrenzte Menge an und lehnt danach ab.
type voll struct{ uebrig int }

func (v *voll) Write(teil []byte) (int, error) {
	if len(teil) > v.uebrig {
		v.uebrig = 0

		return 0, errVoll
	}

	v.uebrig -= len(teil)

	return len(teil), nil
}

var errVoll = errors.New("kein Platz mehr")

// langerInhalt gibt mehr Bytes, als der Puffer eines [zip.Writer] fasst.
func langerInhalt() []byte {
	inhalt := make([]byte, 10_000)
	for i := range inhalt {
		inhalt[i] = byte(i % 251)
	}

	return inhalt
}
