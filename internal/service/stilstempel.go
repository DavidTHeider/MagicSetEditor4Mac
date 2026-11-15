package service

// **Eine gemerkte Vorlage gilt, bis sich ihre Dateien ändern** — FMT-02,
// S10.NAT-03 (#1691).
//
// # Was vorher galt
//
// Über dem Gedächtnis des Stilspeichers stand, ein `.mse-style` ändere sich
// nicht, solange die Anwendung läuft. Seit der Werkstatt stimmt das nicht mehr
// — sie sichert, legt an, benennt um, löscht —, und für einen Texteditor
// daneben hat es nie gestimmt. Gemessen am nativen Fenster (Hüllenprobe
// `fmt02f`): die Werkstatt sicherte einen roten Grund, und die Karte im Set
// blieb weiß bis zum Neustart.
//
// # Wie es jetzt gilt
//
// **Gemerkt wird mit einem Stempel** über jede Datei der Pakete, aus denen die
// Vorlage entstand — Stil, Spiel und jedes Modul, das `include file:` las —:
// Pfad, Größe und Zeit. Weicht er ab, wird neu gelesen. **Auch ein Paket, das
// es beim Lesen nicht gab, steht im Stempel** (als fehlend): entsteht es
// später, gilt es ohne Neustart.
//
// # Was es kostet, und warum die Pause
//
// Ein Stempel ist ein Gang über die Dateien eines Pakets — bei einem großen
// Stil einige hundert. Ein Bogen zeichnet dreihundert Karten hintereinander;
// je Karte einmal gegangen wären das hunderttausend Abfragen. **Geprüft wird
// deshalb höchstens einmal je `stempelpause`** — eine Änderung, die jemand in
// der Werkstatt sichert und im Set ansieht, liegt länger zurück.
//
// # Was er nicht sieht
//
// **Eine Änderung binnen derselben Zeit auf dieselbe Größe** — manche
// Dateisysteme führen die Zeit grob. Auf APFS ist sie in Nanosekunden.

import (
	"fmt"
	"hash"
	"hash/fnv"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// stempelpause ist, wie lange ein gemerkter Stil ungeprüft gilt.
var stempelpause = time.Second

// protokollleser liest wie ein Paketleser und merkt sich, aus welchen Paketen.
type protokollleser struct {
	innen  msepaket.Paketquelle
	sperre sync.Mutex
	pakete []string
}

// Lies gibt eine Datei aus einem Paket und merkt sich das Paket.
func (l *protokollleser) Lies(pfad msepaket.Paketpfad) ([]byte, error) {
	l.merke(pfad.Paket)

	return l.innen.Lies(pfad)
}

// merke nimmt einen Paketnamen einmal auf.
func (l *protokollleser) merke(name string) {
	l.sperre.Lock()
	defer l.sperre.Unlock()

	if !slices.Contains(l.pakete, name) {
		l.pakete = append(l.pakete, name)
	}
}

// namen gibt die gemerkten Pakete.
func (l *protokollleser) namen() []string {
	l.sperre.Lock()
	defer l.sperre.Unlock()

	return slices.Clone(l.pakete)
}

// paketstempel faltet Pfad, Größe und Zeit jeder Datei der genannten Pakete
// zusammen — jedes in der ersten Wurzel, in der es liegt, wie beim Lesen.
func paketstempel(wurzeln, namen []string) string {
	summe := fnv.New64a()

	for _, name := range namen {
		stempelFuer(summe, wurzeln, name)
	}

	return strconv.FormatUint(summe.Sum64(), 16)
}

// stempelFuer schreibt den Stempel eines Pakets — oder dass es fehlt.
//
// **Ein Gang über ein Archiv besucht die eine Datei**, ein Gang über ein
// Verzeichnis jede darin; beide Formen eines Pakets gehen denselben Weg.
func stempelFuer(summe hash.Hash, wurzeln []string, name string) {
	for _, wurzel := range wurzeln {
		ort := filepath.Join(wurzel, name)
		if _, err := os.Stat(ort); err != nil {
			continue
		}

		//nolint:errcheck // ein Gang, der scheitert, gibt einen anderen Stempel — und damit ein neues Lesen
		_ = filepath.WalkDir(ort, func(pfad string, eintrag fs.DirEntry, err error) error {
			if err == nil && !eintrag.IsDir() {
				if angaben, fehler := eintrag.Info(); fehler == nil {
					summe.Write(fmt.Appendf(nil, "%s %d %d\n", pfad, angaben.Size(), angaben.ModTime().UnixNano()))
				}
			}

			return nil
		})

		return
	}

	summe.Write(fmt.Appendf(nil, "%s fehlt\n", name))
}

// stempelGilt sagt, ob ein gemerkter Stil noch gilt — und rückt die Zeit der
// letzten Prüfung vor, wenn er geprüft wurde und galt.
func (s *Stilspeicher) stempelGilt(gemerkt *gemerkterStil) bool {
	if time.Since(gemerkt.geprueft) < stempelpause {
		return true
	}

	if paketstempel(s.wurzeln, gemerkt.pakete) != gemerkt.stempel {
		// **Gemeldet, bevor neu gelesen wird** — wer Bilder vorhält, räumt
		// sie, ehe die neue Vorlage nach ihnen fragt (#2005).
		if s.paketwechsel != nil {
			s.paketwechsel()
		}

		return false
	}

	gemerkt.geprueft = time.Now()

	return true
}

// gestempelt gibt einen gemerkten Stil samt Stempel.
func (s *Stilspeicher) gestempelt(vorlage Vorlage, befunde []domain.Befund, da bool, pakete []string) gemerkterStil {
	return gemerkterStil{
		vorlage: vorlage, befunde: befunde, da: da,
		pakete: pakete, stempel: paketstempel(s.wurzeln, pakete), geprueft: time.Now(),
	}
}
