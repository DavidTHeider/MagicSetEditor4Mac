package laufbeleg

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Eingabe ist eine Pflichteingabe eines Laufs — ein Verzeichnis oder eine
// Datei, samt dem, was darin lag (AK-01: „Eingabe-/Fixturehashes").
type Eingabe struct {
	Name    string `json:"name"`
	Pfad    string `json:"pfad"`
	Muster  string `json:"muster,omitempty"`
	Hash    string `json:"hash,omitempty"`
	Dateien int    `json:"dateien"`
	Passend int    `json:"passend"`
	Bytes   int64  `json:"bytes"`
	Fehlt   string `json:"fehlt,omitempty"`
}

// Erhebe prüft eine Eingabe und rechnet ihren Hash.
//
// # Wann sie fehlt
//
// Wenn sie **nicht gesetzt** ist, **nicht vorliegt**, **leer** ist, nichts
// enthält, was zum Muster passt, oder **nicht lesbar** ist. Der vierte Fall
// ist der aus #1656: das Verzeichnis lag da — im eigenen Zuhause des Laufs,
// und leer.
//
// # Der Hash
//
// **Über Namen und Inhalt, nicht über Zeitstempel**: je Datei eine Zeile
// `<relativer Pfad>\x00<SHA-256 des Inhalts>\n`, in der Ordnung von
// `filepath.WalkDir` — lexikalisch und damit auf jedem Rechner dieselbe. Eine
// einzelne Datei hat den Hash ihres Inhalts.
//
// **Verknüpfungen** zählen mit ihrem Ziel, wenn es eine Datei ist; auf ein
// Verzeichnis werden sie nicht verfolgt, sonst liefe ein Kreis ohne Ende.
// **Verborgenes** — alles, was mit `.` beginnt — zählt gar nicht.
func Erhebe(name, pfad, muster string) Eingabe {
	eingabe := Eingabe{Name: name, Pfad: pfad, Muster: muster}

	eingabe.Fehlt = eingabe.erheben()
	if eingabe.Fehlt != "" {
		eingabe.Hash = ""
	}

	return eingabe
}

// erheben zählt und rechnet — und gibt zurück, was fehlt.
func (e *Eingabe) erheben() string {
	if e.Pfad == "" {
		return "nicht gesetzt"
	}

	// **Die Wurzel wird aufgelöst, innen nichts** — MSEs `data` ist eine
	// Verknüpfung auf das Paketverzeichnis, und `WalkDir` steigt in eine
	// verknüpfte Wurzel nicht hinab: die Eingabe hieß »leer« (2026-09-22).
	// Die Wurzel nennt, wer den Lauf startet; einem Zeiger **darin** folgt
	// der Lauf weiterhin nicht.
	wurzel, fehler := filepath.EvalSymlinks(e.Pfad)
	info, statfehler := os.Stat(wurzel)

	if fehler != nil || statfehler != nil {
		return "liegt nicht vor"
	}

	if !info.IsDir() {
		inhalt, lesefehler := inhaltshash(wurzel)
		if lesefehler != nil {
			return "nicht lesbar: " + filepath.Base(e.Pfad)
		}

		e.zaehle(filepath.Base(e.Pfad), info.Size())
		e.Hash = "sha256:" + inhalt

		return e.menge()
	}

	summe := sha256.New()

	fehler = filepath.WalkDir(wurzel, func(pfad string, eintrag fs.DirEntry, fehler error) error {
		if fehler != nil {
			return fmt.Errorf("nicht lesbar: %w", fehler)
		}

		// **Verborgenes zählt nicht**: `.DS_Store` schreibt der Finder, wann
		// er will, und der Hash änderte sich ohne neue Daten.
		if pfad != wurzel && strings.HasPrefix(eintrag.Name(), ".") {
			if eintrag.IsDir() {
				return filepath.SkipDir
			}

			return nil
		}

		// **Stat folgt einer Verknüpfung**, `eintrag` nicht: so zählt ein
		// Zeiger auf eine Datei mit ihrem Inhalt, und einer auf ein
		// Verzeichnis oder ins Leere zählt gar nicht.
		ziel, statfehler := os.Stat(pfad)
		if statfehler != nil {
			return nil //nolint:nilerr // eine Verknüpfung ins Leere ist kein Inhalt und zählt nicht
		}

		if eintrag.IsDir() || !ziel.Mode().IsRegular() {
			return nil
		}

		relativ, _ := filepath.Rel(wurzel, pfad) //nolint:errcheck // pfad liegt unter der Wurzel, WalkDir hat ihn von dort

		inhalt, fehler := inhaltshash(pfad)
		if fehler != nil {
			return fmt.Errorf("nicht lesbar: %s", filepath.ToSlash(relativ))
		}

		_, _ = fmt.Fprintf(summe, "%s\x00%s\n", filepath.ToSlash(relativ), inhalt) //nolint:errcheck // ein Hash nimmt jede Zeile an
		e.zaehle(eintrag.Name(), ziel.Size())

		return nil
	})
	if fehler != nil {
		return fehler.Error()
	}

	e.Hash = "sha256:" + hex.EncodeToString(summe.Sum(nil))

	return e.menge()
}

// zaehle nimmt eine Datei in die Zählung auf.
func (e *Eingabe) zaehle(name string, groesse int64) {
	e.Dateien++
	e.Bytes += groesse

	if passt, _ := filepath.Match(e.Muster, name); passt || e.Muster == "" { //nolint:errcheck // ein kaputtes Muster passt auf nichts, und das sagt `menge`
		e.Passend++
	}
}

// menge sagt, ob genug da war.
func (e *Eingabe) menge() string {
	switch {
	case e.Dateien == 0:
		return "leer"
	case e.Passend == 0:
		return "kein " + e.Muster + " darin"
	default:
		return ""
	}
}

// inhaltshash gibt den SHA-256 eines Dateiinhalts.
func inhaltshash(pfad string) (string, error) {
	datei, fehler := os.Open(pfad) // #nosec G304 -- der Pfad ist eine Pflichteingabe, die der Aufruf nennt
	if fehler != nil {
		return "", fehler
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // nur gelesen

	summe := sha256.New()
	_, fehler = io.Copy(summe, datei)

	return hex.EncodeToString(summe.Sum(nil)), fehler
}
