package app

// **Gedreht wird auch während der Sitzung** — PERF03
// (#2353),
// Befund A-20 des Audits 0.9.0: die Grenze galt nur beim Einrichten
// ([drehenMit]), und ein Dienst, der eine Woche lief, schrieb eine Datei ohne
// Ende.
//
// # Wie, mit mehreren Schreibern
//
// **Fenster, Dienst und Werkzeugserver schreiben in dieselbe Datei.** Jeder
// sieht vor jeder Zeile nach, ob die Datei am Ort noch die ist, die er offen
// hält — wenn nicht, hat ein anderer gedreht oder sie wurde gelöscht, und er
// öffnet neu. Ist sie die eigene und über der Grenze, dreht er: unter einer
// eigenen kurzen Sperre ([Drehsperre]), und dort fragt er noch einmal, ob
// nicht ein anderer schneller war. **Was eine Datei über die Grenze trägt,
// ist höchstens eine Zeile je Schreiber**; die Zeile eines anderen, der
// gerade dreht, landet in der Vorfassung.
//
// **Die Startsperre bleibt** ([drehenMit]): sie dreht beim Einrichten nur,
// wenn niemand sonst die Datei hält. Während der Sitzung muss niemand allein
// sein, weil jeder dem Drehen folgt.

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
)

// Drehsperre ist die Nachbardatei, unter der während der Sitzung gedreht
// wird.
const Drehsperre = Protokollname + ".drehen"

// protokolldatei ist die offene Datei samt ihrem Griff auf die Sperre — und
// dreht, wenn sie während der Sitzung über ihre Grenze wächst.
type protokolldatei struct {
	*os.File
	griff *dateisperre.Griff

	pfad         string
	grenze       int64
	aufbewahrung int
	sperre       sync.Mutex
}

// vorDemDrehen läuft, nachdem ein Schreiber die Grenze gesehen hat und bevor
// er die Drehsperre nimmt — eine Naht für den Lauf, in dem ein anderer
// schneller dreht. Im Betrieb tut sie nichts.
var vorDemDrehen = func() {}

// Write schreibt eine Zeile — nachdem es nachgesehen hat, ob gedreht wird.
func (d *protokolldatei) Write(zeile []byte) (int, error) {
	d.sperre.Lock()
	defer d.sperre.Unlock()

	d.nachsehen()

	return d.File.Write(zeile)
}

// WriteString schreibt über [protokolldatei.Write] — sonst nähme
// `io.WriteString` das `WriteString` der eingebetteten Datei und ginge am
// Nachsehen vorbei.
func (d *protokolldatei) WriteString(zeile string) (int, error) {
	return d.Write([]byte(zeile)) //nolint:gocritic // d.WriteString wäre diese Funktion selbst
}

// Close gibt die Sperre zurück und schließt die Datei.
func (d *protokolldatei) Close() error {
	d.sperre.Lock()
	defer d.sperre.Unlock()

	schliesseGriff(d.griff)

	return d.File.Close()
}

// nachsehen folgt einem fremden Drehen und dreht selbst, wenn die eigene
// Datei über der Grenze ist. **Jeder Fehlschlag heißt „weiter wie bisher"**:
// ein Protokoll, das beim Drehen scheitert, schreibt weiter, wohin es kann.
func (d *protokolldatei) nachsehen() {
	eigene, fehler := d.Stat()
	if fehler != nil {
		return
	}

	if amOrt, ortfehler := os.Stat(d.pfad); ortfehler == nil && os.SameFile(eigene, amOrt) {
		if eigene.Size() < d.grenze || !d.drehen(eigene) {
			return
		}
	}

	// #nosec G304 -- derselbe Ort, den das Einrichten geprüft hat
	neu, fehler := os.OpenFile(d.pfad, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if fehler != nil {
		return
	}

	_ = d.File.Close() //nolint:errcheck // die alte ist geschrieben, und ihr Fehler hätte keinen Empfänger
	d.File = neu
}

// drehen legt die Datei beiseite, wenn sie unter der Drehsperre noch die
// eigene ist — und sagt, ob am Ort jetzt eine andere steht.
func (d *protokolldatei) drehen(eigene os.FileInfo) bool {
	vorDemDrehen()

	griff, fehler := dateisperre.Oeffnen(filepath.Join(filepath.Dir(d.pfad), Drehsperre))
	if fehler != nil {
		return false
	}

	defer schliesseGriff(griff)

	// **Ein Erbe hält das Drehen nicht auf** (#2401): hielt ein Kind die
	// Drehsperre zwischen `fork` und `exec` mit, fiel jedes Drehen in dieser
	// Zeit aus, und die Datei wuchs über ihre Grenze.
	if griff.AlleinOhneErbe(context.Background()) != nil {
		return false
	}

	// **Ein anderer war schneller** — dann steht am Ort schon die neue, und
	// es bleibt nur, ihr zu folgen.
	if amOrt, ortfehler := os.Stat(d.pfad); ortfehler == nil && os.SameFile(eigene, amOrt) {
		vorfassungenRuecken(d.pfad, d.aufbewahrung)
	}

	return true
}

// vorfassungenRuecken rückt die Vorfassungen weiter und legt die Datei als
// `.1` beiseite.
//
// **Von hinten nach vorn**, damit keine Vorfassung eine andere überschreibt,
// bevor diese weitergerückt ist. **Umbenennen und nicht kopieren**:
// `os.Rename` ersetzt ein vorhandenes Ziel in einem Zug, und zwischen zwei
// Zügen läge ein Augenblick ohne Protokoll.
func vorfassungenRuecken(pfad string, aufbewahrung int) {
	for stelle := aufbewahrung - 1; stelle >= 1; stelle-- {
		_ = os.Rename(vorfassung(pfad, stelle), //nolint:errcheck // eine fehlende Vorfassung rückt nicht
			vorfassung(pfad, stelle+1))
	}

	_ = os.Rename(pfad, vorfassung(pfad, 1)) //nolint:errcheck // siehe oben
}
