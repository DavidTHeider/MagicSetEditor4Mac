package msepaket

// **Die Ressourcenregel für Archivpakete** —
// #1886,
// die Regel mit ihrer Messung in `docs/FORMATE.md`.
//
// # Warum vor `archive/zip`
//
// **Weil es das ganze Verzeichnis liest, bevor es etwas zurückgibt** —
// nachgelesen in Go 1.27.1, `src/archive/zip/reader.go`, `(*Reader).init`:
// die Zahl aus dem Verzeichnisende bestimmt die Kapazität der Liste, sobald
// die Datei groß genug ist, und danach liest die Schleife Kopf um Kopf,
// **bis einer nicht passt** — die Größe aus dem Ende begrenzt sie nicht.
//
// Deshalb zwei Riegel:
//
//  1. **Das Ende wird zuerst gelesen** ([Archivende]) und abgewiesen, was
//     mehr Einträge oder ein größeres Verzeichnis nennt als die Grenzen.
//  2. **Beim Lesen des Verzeichnisses zählt ein Leser mit** und bricht ab,
//     sobald mehr gelesen ist, als das Ende zugesagt hat — ein Verzeichnis,
//     das länger ist als angegeben, kommt nicht weiter als seine Angabe.
//     Danach liest derselbe Leser ohne Grenze: sie gilt dem Verzeichnis und
//     nicht dem Inhalt, und den hält [GrenzePaketdatei].

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"sync/atomic"
)

// Die Grenzen eines Archivpakets.
//
// **Gemessen am 2026-10-01**: das größte der zehn kanonischen Sets
// (`Black.mse-set`) hat 3.838 Einträge und ein Verzeichnis von 210 KB; das
// ganze `full-magic-pack` — 34.683 Dateien, als ein einziges Archiv gedacht —
// hätte eines von etwa 3,5 MB. **Hunderttausend Einträge und 16 MiB** lassen
// das Dreifache davon und halten ein Archiv aus Millionen leerer Einträge auf.
const (
	GrenzeArchiveintraege   = 100_000
	GrenzeArchivverzeichnis = 16 << 20
)

// verzeichnisspielraum ist, was `archive/zip` jenseits des Verzeichnisses
// liest: die Suche nach dem Ende (bis 65 KiB samt Kommentar), das zip64-Ende
// und den Vorlauf seines Puffers (4 KiB).
const verzeichnisspielraum = 128 << 10

// ErrArchivZuGross meldet ein Archiv über [GrenzeArchiveintraege] oder
// [GrenzeArchivverzeichnis].
var ErrArchivZuGross = errors.New(
	"das Archiv hat mehr Einträge oder ein größeres Verzeichnis, als ein Paket haben darf")

// errUeberZusage ist der Abbruch des mitzählenden Lesers.
var errUeberZusage = errors.New("das Verzeichnis ist länger, als sein Ende sagt")

// Archivende ist, was das Verzeichnisende eines Archivs zusagt.
type Archivende struct {
	Zahl, Groesse, Versatz uint64

	// Endversatz ist, wo das Ende steht — beim zip64-Ende dessen Anfang.
	Endversatz int64
}

// oeffneArchiv öffnet ein Archivpaket nach der Ressourcenregel.
//
// **`zip.ErrInsecurePath` kommt samt Leser zurück**, wie bei
// `zip.OpenReader`: der Installer benennt unsichere Namen einzeln.
func oeffneArchiv(pfad string) (*zip.Reader, io.Closer, error) {
	datei, fehler := os.Open(pfad) // #nosec G304 -- das Paket, das jemand öffnen will
	if fehler != nil {
		return nil, nil, fehler
	}

	leser, fehler := archivLesen(datei)
	if fehler != nil && !errors.Is(fehler, zip.ErrInsecurePath) {
		return nil, nil, errors.Join(fehler, datei.Close())
	}

	return leser, datei, fehler
}

// quelle ist, was ein Archiv hergibt.
type quelle interface {
	io.ReaderAt
	Stat() (os.FileInfo, error)
}

func archivLesen(datei quelle) (*zip.Reader, error) {
	angaben, fehler := datei.Stat()
	if fehler != nil {
		return nil, fmt.Errorf("das Archiv messen: %w", fehler)
	}

	groesse := angaben.Size()

	ende, fehler := LiesArchivende(datei, groesse)
	if fehler != nil {
		// **Kein lesbares Ende ist kein zu großes** — die Klasse bleibt die
		// von `archive/zip`, damit ein Aufrufer „kaputt" weiterhin erkennt.
		return nil, fmt.Errorf("%w: %w", zip.ErrFormat, fehler)
	}

	if ende.Zahl > GrenzeArchiveintraege || ende.Groesse > GrenzeArchivverzeichnis {
		return nil, fmt.Errorf("%w (%d Einträge, Verzeichnis %d Byte; höchstens %d und %d)",
			ErrArchivZuGross, ende.Zahl, ende.Groesse, GrenzeArchiveintraege, GrenzeArchivverzeichnis)
	}

	zaehler := &mitzaehlend{quelle: datei, budget: int64(ende.Groesse) + verzeichnisspielraum} //nolint:gosec // oben begrenzt

	leser, fehler := zip.NewReader(zaehler, groesse)
	zaehler.frei.Store(true)

	if errors.Is(fehler, errUeberZusage) {
		return nil, fmt.Errorf("%w: %w", ErrArchivZuGross, fehler)
	}

	return leser, fehler
}

// mitzaehlend liest und bricht ab, sobald mehr gelesen ist als das Budget —
// bis es frei ist.
type mitzaehlend struct {
	quelle  io.ReaderAt
	budget  int64
	gelesen atomic.Int64
	frei    atomic.Bool
}

func (m *mitzaehlend) ReadAt(ziel []byte, stelle int64) (int, error) {
	if !m.frei.Load() && m.gelesen.Add(int64(len(ziel))) > m.budget {
		return 0, errUeberZusage
	}

	return m.quelle.ReadAt(ziel, stelle)
}
