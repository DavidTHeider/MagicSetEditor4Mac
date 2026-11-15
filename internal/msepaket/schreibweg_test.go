// **Eine Datei im Paket zurückschreiben** — Werkstück 2 von Phase 3.3
// (#1164).
//
// # Woher der Weg kommt
//
// Aus `Package::saveToZipfile` und `Package::saveToDirectory`
// (`src/util/io/package.cpp`, Zeilen 402–491, Fassung f9d9356 vom 2021-07-06).
// MSE2 schreibt zuerst in eine Nachbardatei, benennt die alte in `.bak` um und
// erst dann die neue an ihren Platz — und **unveränderte Einträge eines
// Archivs kopiert es als Einträge** (`CopyEntry`), nicht neu gepackt.
//
// # Was MSE2 dabei nicht tut
//
// Ein **Verzeichnis**paket sichert es nicht: `saveToDirectory` löscht die alte
// Datei und schiebt die neue darüber. Hier wird auch dort eine `.bak`
// angelegt — **adapted**, und der Grund steht im Issue: „nichts überschreiben,
// was nicht zurückzuholen ist". Wer eine Vorlage im Fenster bearbeitet, hat
// keinen zweiten Weg an den vorigen Stand; wer in MSE2 eine Datei mit einem
// Texteditor ändert, hatte ihn nie anders.
package msepaket_test

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// summe gibt die SHA-256 eines Inhalts in der Form, die der Schreibweg erwartet.
func summe(inhalt []byte) string {
	roh := sha256.Sum256(inhalt)

	return hex.EncodeToString(roh[:])
}

// probeverzeichnis legt ein Verzeichnispaket mit zwei Dateien an.
func probeverzeichnis(t *testing.T) string {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "probe.mse-style")
	if err := os.MkdirAll(filepath.Join(ort, "unterordner"), 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	for name, inhalt := range map[string]string{
		"style":              "mse version: 2.0.0\ngame: probe\n",
		"unterordner/skript": "1 + 1\n",
	} {
		if err := os.WriteFile(filepath.Join(ort, filepath.FromSlash(name)),
			[]byte(inhalt), 0o600); err != nil {
			t.Fatalf("schreiben: %v", err)
		}
	}

	return ort
}

// probearchiv legt ein Archivpaket mit zwei Dateien an.
func probearchiv(t *testing.T) string {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "probe.mse-style")

	datei, err := os.Create(ort) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	packer := zip.NewWriter(datei)

	// **Ein Eintrag liegt ungepackt darin, und das ist Absicht.** Ein Archiv
	// aus einer fremden Hand mischt die Verfahren — MSE2 schreibt über
	// `wxZipOutputStream`, und was dort wie herauskommt, bestimmt nicht dieses
	// Programm. Ein Schreibweg, der jeden Eintrag durch **seinen** Packer
	// schickt, macht aus `Store` ein `Deflate`; genau daran ist zu sehen, ob
	// er die Bytes übernimmt oder neu herstellt.
	for _, eintrag := range []struct {
		name     string
		inhalt   string
		verfahre uint16
	}{
		{"style", "mse version: 2.0.0\ngame: probe\n", zip.Deflate},
		{"unterordner/skript", "1 + 1\n", zip.Store},
	} {
		strom, fehler := packer.CreateHeader(&zip.FileHeader{
			Name: eintrag.name, Method: eintrag.verfahre,
		})
		if fehler != nil {
			t.Fatalf("Eintrag: %v", fehler)
		}

		if _, fehler := strom.Write([]byte(eintrag.inhalt)); fehler != nil {
			t.Fatalf("schreiben: %v", fehler)
		}
	}

	if fehler := errors.Join(packer.Close(), datei.Close()); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	return ort
}

// **Der geänderte Inhalt steht danach da — in beiden Formen.**
func TestEineGeaenderteDateiStehtDanachImPaket(t *testing.T) {
	t.Parallel()

	for name, anlegen := range map[string]func(*testing.T) string{
		"Verzeichnis": probeverzeichnis,
		"Archiv":      probearchiv,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ort := anlegen(t)
			alt := "mse version: 2.0.0\ngame: probe\n"
			neu := "mse version: 2.0.0\ngame: probe\nfull name: Probe\n"

			if err := msepaket.DateiSchreiben(t.Context(), ort, "style", []byte(neu), summe([]byte(alt))); err != nil {
				t.Fatalf("DateiSchreiben: %v", err)
			}

			if gelesen := dateiAus(t, ort, "style"); gelesen != neu {
				t.Errorf("gelesen %q", gelesen)
			}
		})
	}
}

// **Die anderen Dateien bleiben Byte für Byte, wie sie waren.**
//
// Das ist die Rundlaufzusage aus dem Issue, an der Stelle, an der sie brechen
// könnte: ein Archiv, das beim Schreiben neu gepackt wird, ändert jede Datei
// darin — und MSE2 liest zwar noch, aber ein Vergleich stimmt nie wieder.
func TestDieUnberuehrtenDateienBleibenUnberuehrt(t *testing.T) {
	t.Parallel()

	for name, anlegen := range map[string]func(*testing.T) string{
		"Verzeichnis": probeverzeichnis,
		"Archiv":      probearchiv,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ort := anlegen(t)
			vorher := dateiAus(t, ort, "unterordner/skript")

			if err := msepaket.DateiSchreiben(t.Context(), ort, "style", []byte("x\n"),
				summe([]byte("mse version: 2.0.0\ngame: probe\n"))); err != nil {
				t.Fatalf("DateiSchreiben: %v", err)
			}

			if nachher := dateiAus(t, ort, "unterordner/skript"); nachher != vorher {
				t.Errorf("aus %q wurde %q", vorher, nachher)
			}
		})
	}
}

// **Und im Archiv bleiben sie auch als Bytes dieselben.**
//
// # Warum der Lauf darüber das nicht zeigt
//
// Er liest über [msepaket.Paket.Datei], und das packt aus. Ein Schreibweg, der
// jeden Eintrag auspackt und neu packt, käme dort durch — der **Inhalt** wäre
// ja gleich. Die Zusage aus dem Issue ist aber „bytegenau", und die entscheidet
// sich an der gepackten Form: Verfahren, Länge, Prüfsumme.
//
// Das ist die Gegenprobe zu [zip.Writer.Copy]. Am 2026-08-26 gemessen: mit
// `Create` und `io.Copy` statt `Copy` wird aus den acht rohen Bytes des
// ungepackten Eintrags ein vierzehn Byte langer Deflate-Strom — dieser Lauf
// rot, der darüber grün.
//
// **Der erste Anlauf der Gegenprobe zeigte nichts** (§10: eine Gegenprobe, die
// nichts ändert, ist keine). Damals lagen beide Einträge gepackt im Archiv,
// und derselbe Packer über sechs Byte ergibt zweimal dieselben Bytes.
// Unterscheidbar wird es erst an einem Eintrag, dessen Form der Schreiber
// **nicht von selbst** herstellt — deshalb liegt einer ungepackt darin.
func TestEinUnberuehrterEintragBehaeltSeineBytes(t *testing.T) {
	t.Parallel()

	ort := probearchiv(t)
	vorher := rohbytes(t, ort, "unterordner/skript")

	if err := msepaket.DateiSchreiben(t.Context(), ort, "style", []byte("x\n"),
		summe([]byte("mse version: 2.0.0\ngame: probe\n"))); err != nil {
		t.Fatalf("DateiSchreiben: %v", err)
	}

	if nachher := rohbytes(t, ort, "unterordner/skript"); !bytes.Equal(nachher, vorher) {
		t.Errorf("aus %v wurde %v", vorher, nachher)
	}
}

// rohbytes gibt einen Eintrag so, wie er im Archiv liegt — gepackt.
func rohbytes(t *testing.T, ort, name string) []byte {
	t.Helper()

	archiv, err := zip.OpenReader(ort)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := archiv.Close(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	for _, eintrag := range archiv.File {
		if eintrag.Name != name {
			continue
		}

		strom, oeffnefehler := eintrag.OpenRaw()
		if oeffnefehler != nil {
			t.Fatalf("roh öffnen: %v", oeffnefehler)
		}

		roh, lesefehler := io.ReadAll(strom)
		if lesefehler != nil {
			t.Fatalf("lesen: %v", lesefehler)
		}

		// **Verfahren und Prüfsumme gehören dazu.** Dieselben gepackten Bytes
		// unter einem anderen Verfahren wären eine andere Datei.
		return append(roh, byte(eintrag.Method), byte(eintrag.CRC32))
	}

	t.Fatalf("%s liegt nicht im Archiv", name)

	return nil
}

// **Ein fremder Stand wird abgewiesen, und nichts ist geschrieben.**
//
// Der Konfliktweg aus dem Issue: die Datei hat sich seit dem Öffnen geändert.
// **Beides gehört geprüft** — dass es einen Fehler gibt und dass er nichts
// hinterlässt; ein Abbruch nach dem halben Schreiben wäre schlimmer als kein
// Schutz, weil er wie einer aussieht.
func TestEinFremderStandWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, anlegen := range map[string]func(*testing.T) string{
		"Verzeichnis": probeverzeichnis,
		"Archiv":      probearchiv,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ort := anlegen(t)
			vorher := dateiAus(t, ort, "style")

			fehler := msepaket.DateiSchreiben(t.Context(), ort, "style", []byte("x\n"), summe([]byte("etwas anderes")))
			if !errors.Is(fehler, domain.ErrConflict) {
				t.Fatalf("der Fehler ist %v — will ErrConflict", fehler)
			}

			if nachher := dateiAus(t, ort, "style"); nachher != vorher {
				t.Errorf("es wurde doch geschrieben: %q", nachher)
			}
		})
	}
}

// **Der vorige Stand ist danach noch zu lesen.**
//
// Der Rückweg. Bei einem Archiv ist es MSEs eigener Griff, beim Verzeichnis
// unserer — siehe den Kopf dieser Datei.
//
// **Gelesen und nicht nur gezählt**: eine `.bak`, die dasteht und nicht
// aufgeht, ist kein Rückweg. Beim Archiv ist sie ein ganzes Paket, und dass
// sie sich als eines öffnen lässt, ist die halbe Zusage.
func TestDerVorigeStandBleibtLesbar(t *testing.T) {
	t.Parallel()

	alt := "mse version: 2.0.0\ngame: probe\n"

	for name, fall := range map[string]struct {
		anlegen func(*testing.T) string
		lesen   func(*testing.T, string) string
	}{
		"Verzeichnis": {probeverzeichnis, func(t *testing.T, ort string) string {
			t.Helper()

			roh, err := os.ReadFile( // #nosec G304 -- aus t.TempDir()
				filepath.Join(ort+msepaket.Sicherungsendung, "style"))
			if err != nil {
				t.Fatalf("der Rückweg fehlt: %v", err)
			}

			return string(roh)
		}},
		// **Von Hand aufgemacht und nicht über `Oeffne`.** Eine `.bak` trägt
		// keine Paketendung mehr, und `Oeffne` weist sie deshalb ab — bei
		// MSE2 genauso. Sie ist ein Rückweg für einen Menschen, der sie
		// zurückbenennt, nicht für das Programm.
		"Archiv": {probearchiv, func(t *testing.T, ort string) string {
			t.Helper()

			archiv, err := zip.OpenReader(ort + msepaket.Sicherungsendung)
			if err != nil {
				t.Fatalf("der Rückweg fehlt: %v", err)
			}

			t.Cleanup(func() {
				if fehler := archiv.Close(); fehler != nil {
					t.Errorf("schließen: %v", fehler)
				}
			})

			strom, oeffnefehler := archiv.Open("style")
			if oeffnefehler != nil {
				t.Fatalf("style in der Sicherung: %v", oeffnefehler)
			}

			roh, lesefehler := io.ReadAll(strom)
			if lesefehler != nil {
				t.Fatalf("lesen: %v", lesefehler)
			}

			return string(roh)
		}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ort := fall.anlegen(t)

			if err := msepaket.DateiSchreiben(t.Context(), ort, "style", []byte("x\n"), summe([]byte(alt))); err != nil {
				t.Fatalf("DateiSchreiben: %v", err)
			}

			if gesichert := fall.lesen(t, ort); gesichert != alt {
				t.Errorf("die Sicherung trägt %q", gesichert)
			}
		})
	}
}

// dateiAus liest eine Datei aus einem Paket — gleich welcher Form.
func dateiAus(t *testing.T, ort, name string) string {
	t.Helper()

	paket, err := msepaket.Oeffne(ort)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	inhalt, lesefehler := paket.Datei(name)
	if lesefehler != nil {
		t.Fatalf("%s lesen: %v", name, lesefehler)
	}

	return string(inhalt)
}
