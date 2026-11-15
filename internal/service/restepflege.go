package service

// **Liegengebliebene eigene Reste wegräumen** — OPS04
// (#2350),
// Befund A-27 des Audits 0.9.0.
//
// Bilder, Pakete, `.mse-set`, Sicherungen und die Rückspielmarke entstehen in
// einer Nachbardatei oder einem Nachbarordner mit führendem Punkt und kommen
// erst fertig unter ihren Namen. **Ein Abbruch lässt den Nachbarn liegen** —
// das ist richtig, besser als ein halbes PNG unter dem richtigen Namen —, aber
// weggeräumt hat ihn bis dahin niemand, und die Platzübersicht zählte ihn
// unter „Sonstiges".
//
// # Woran ein eigener Rest zu erkennen ist
//
// **Am Ort und daran, wer ihn dort anlegt** — nicht an einer Namensgestalt
// allein (Zweitblick über Phase 4, Befund 1): ein `.notizen.2026` neben einem
// Bestand, den `MSE_BESTAND` nach `~/Dokumente` legt, gehört jemand anderem.
// [Restorte] nennt drei Arten von Ort:
//
//   - **neben dem Bestand** — ein Ordner, den diese Anwendung teilen kann.
//     Dort geht nur, was ein benannter Erzeuger dort anlegt ([nebenreste]),
//     und **kein Unterordner wird betreten**.
//   - **eigene Ordner** — das Beiwerk: dort schreibt nur diese Anwendung, und
//     es gilt die allgemeine Gestalt `.<name>.<zufall>` (`os.CreateTemp` in
//     [SchreibeAtomar], `storage.Sichern`) oder `.<name>-<zufall>`
//     (`os.MkdirTemp`), oben und eine Ebene tiefer (`beiwerk/<set>/…`).
//   - **flache eigene Ordner** — die Pakete: dieselbe Gestalt, aber nur oben;
//     darunter liegen bezogene Pakete, und ihr Inhalt ist fremd.
//
// **Und das Alter**: länger als [Restfrist] unverändert. Ein laufender
// Vorgang schreibt in seinen Nachbarn, und keiner dauert eine Stunde, ohne
// ihn anzufassen.
//
// **Und der Aufrufer belegt dazu, dass niemand sonst arbeitet** — das
// Fenster räumt nur vor dem Öffnen und nur, wenn kein anderer Prozess den
// Bestand hält (`internal/app/restepflege.go`).
//
// # Was nie geschieht
//
// **Ein Verweis wird nicht verfolgt**: eine Wurzel, die ein Verweis ist, wird
// nicht betreten; ein Rest, der auf etwas zeigt, geht selbst, sein Ziel
// bleibt.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// Restorte sagt, wo eigene Reste liegen können — und was dort als eigen gilt.
type Restorte struct {
	// Bestand ist die Datei des Bestands. **Ihr Ordner kann fremde Dateien
	// tragen**; dort gehen nur die [nebenreste].
	Bestand string

	// Eigene gehören nur dieser Anwendung — oben und eine Ebene tiefer.
	Eigene []string

	// Flach gehören ihr auch, aber nur oben: darunter ist der Inhalt fremd.
	Flach []string
}

// Rueckspielvorsatz und Sicherungsvorsatz sind die Zwischenplätze neben dem
// Bestand — beim Zurückspielen und beim Sichern mit Kennwort. **Ihre Erzeuger
// nehmen dieselben Zeichen** ([bestandsauspacken], [Bestandsweg.sichern]).
const (
	Rueckspielvorsatz = ".rueckspiel-"
	Sicherungsvorsatz = ".sicherung-"
)

// Vorsicherungsform ist die Gestalt des Namens einer Sicherung vor einer
// Migration — `bestand-vor-<fassung>.db`, die Fassung samt Stand, wo es einen
// gibt (`app.Sicherungsname`). **Die Platzübersicht und die Restepflege lesen
// dieselbe** (Nachlauf über Phase 4, Befund 1): abgeschrieben lief sie
// auseinander, und ein `bestand-vor-notizen.txt` galt als eigen.
const Vorsicherungsform = `bestand-vor-\d+\.\d+\.\d+(-[0-9a-f]{12})?\.db`

// nebenrest ist ein Zwischenname neben dem Bestand **samt der Art, die sein
// Erzeuger anlegt** — `os.MkdirTemp` einen Ordner, das Archiv, `storage.Sichern`
// und [SchreibeAtomar] eine Datei. Ein Eintrag der anderen Art ist nicht von
// hier (Nachlauf über Phase 4, Befund 1).
type nebenrest struct {
	muster *regexp.Regexp
	ordner bool
}

// nebenreste sind die Namen, unter denen diese Anwendung **neben ihrem
// Bestand** vorläufig schreibt — je Erzeuger einer. Ein neuer Zwischenname
// dort, der hier fehlt, bleibt liegen: das ist die sichere Seite.
var nebenreste = []nebenrest{
	// Zurückspielen und Sichern mit Kennwort — `os.MkdirTemp`.
	{regexp.MustCompile(`^` + regexp.QuoteMeta(Rueckspielvorsatz) + `\d+$`), true},
	{regexp.MustCompile(`^` + regexp.QuoteMeta(Sicherungsvorsatz) + `\d+$`), true},
	// Ein Archiv, dessen Ziel neben dem Bestand liegt (`internal/sicherung`).
	{halbesArchiv, false},
	// Die Sicherung vor einer Migration — `storage.Sichern` auf
	// [Vorsicherungsform] (`internal/app/sicherungvormigration.go`).
	{regexp.MustCompile(`^\.` + Vorsicherungsform + `\.entsteht-\d+$`), false},
	// Die Marken neben dem Bestand — [SchreibeAtomar].
	{regexp.MustCompile(`^\.(` + regexp.QuoteMeta(Rueckspielmarke) + `|` + regexp.QuoteMeta(Epochenmarke) + `)\.\d+$`), false},
}

// halbesArchiv ist der Name, unter dem ein Archiv entsteht.
var halbesArchiv = regexp.MustCompile(`^` + regexp.QuoteMeta(sicherung.Unfertigvorsatz) + `\d+` +
	regexp.QuoteMeta(sicherung.Unfertigendung) + `$`)

// halbeArchiveRaeumen räumt in einem Ordner die halben Archive, die älter als
// die [Restfrist] sind — **für den Ordner des Plans** (Zweitblick über Phase
// 4, Befund 12): dort entstehen sie, und brach ein Lauf ab, blieb eines
// liegen, groß wie der Bestand.
//
// **Nur eine Datei** (Nachlauf über Phase 4, Befund 1): der Archiver schreibt
// keine Ordner, und einen gleichnamigen nähme `RemoveAll` samt Inhalt.
func halbeArchiveRaeumen(ordner string, jetzt time.Time) ([]string, error) {
	return resteIn(ordner, jetzt, func(eintrag os.DirEntry) bool {
		return halbesArchiv.MatchString(eintrag.Name()) && eintrag.Type().IsRegular()
	}, false)
}

// Restfrist ist, wie lange ein Rest unverändert liegen muss, bevor er geht.
const Restfrist = time.Hour

// restname erkennt die Namen, unter denen diese Anwendung vorläufig schreibt.
var restname = regexp.MustCompile(`^\.[^/]+[.-]\d+$`)

// RestePflegen räumt an den [Restorte]n die Reste, die älter als die
// [Restfrist] sind, und gibt ihre Pfade. **Ein Ordner, den es nicht gibt, ist
// kein Fehler** — ein frischer Bestand hat noch keinen.
func RestePflegen(orte Restorte, jetzt time.Time) (entfernt []string, fehler error) {
	raeume := func(wurzel string, eigen func(os.DirEntry) bool, tiefer bool) {
		gefunden, err := resteIn(wurzel, jetzt, eigen, tiefer)
		entfernt = append(entfernt, gefunden...)
		fehler = errors.Join(fehler, err)
	}

	if orte.Bestand != "" {
		raeume(filepath.Dir(orte.Bestand), istNebenrest, false)
	}

	for _, wurzel := range orte.Eigene {
		raeume(wurzel, restgestalt, true)
	}

	for _, wurzel := range orte.Flach {
		raeume(wurzel, restgestalt, false)
	}

	return entfernt, fehler
}

// restgestalt ist die allgemeine Gestalt in einem eigenen Ordner — jede Art.
func restgestalt(eintrag os.DirEntry) bool { return restname.MatchString(eintrag.Name()) }

// istNebenrest sagt, ob ein Eintrag einer der [nebenreste] ist — nach Name
// **und** Art.
func istNebenrest(eintrag os.DirEntry) bool {
	for _, rest := range nebenreste {
		if !rest.muster.MatchString(eintrag.Name()) {
			continue
		}

		if rest.ordner && eintrag.IsDir() || !rest.ordner && eintrag.Type().IsRegular() {
			return true
		}
	}

	return false
}

// resteIn räumt einen Ordner — und, wenn `tiefer`, eine Ebene darunter.
//
// **Eine Wurzel, die ein Verweis ist, wird nicht betreten**: `os.ReadDir`
// folgte ihm, und ein Beiwerk, das auf `~/Bilder` zeigt, räumte dort.
func resteIn(ordner string, jetzt time.Time, eigen func(os.DirEntry) bool, tiefer bool) (entfernt []string, fehler error) {
	if lage, err := os.Lstat(ordner); err == nil && lage.Mode()&os.ModeSymlink != 0 {
		return nil, nil
	}

	eintraege, err := os.ReadDir(ordner)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("die Reste in %s lesen: %w", ordner, err)
	}

	for _, eintrag := range eintraege {
		pfad := filepath.Join(ordner, eintrag.Name())

		if eigen(eintrag) {
			weg, err := restRaeumen(pfad, eintrag, jetzt)
			if err != nil {
				fehler = errors.Join(fehler, err)
			} else if weg {
				entfernt = append(entfernt, pfad)
			}

			continue
		}

		// **`IsDir` ist falsch für einen Verweis** — eine Ebene tiefer folgt
		// ihm also auch nicht.
		if tiefer && eintrag.IsDir() && !strings.HasPrefix(eintrag.Name(), ".") {
			gefunden, err := resteIn(pfad, jetzt, eigen, false)
			entfernt = append(entfernt, gefunden...)
			fehler = errors.Join(fehler, err)
		}
	}

	return entfernt, fehler
}

// restRaeumen nimmt einen Rest, der älter ist als die [Restfrist].
//
// **`RemoveAll` folgt keinem Verweis** — es nimmt den Verweis.
func restRaeumen(pfad string, eintrag os.DirEntry, jetzt time.Time) (bool, error) {
	angaben, err := eintrag.Info()
	if err == nil && jetzt.Sub(angaben.ModTime()) <= Restfrist {
		return false, nil
	}

	if err == nil {
		err = os.RemoveAll(pfad)
	}

	if err != nil {
		return false, fmt.Errorf("den Rest %s räumen: %w", pfad, err)
	}

	return true, nil
}
