package service

// **Eine Datei ansehen, ohne dass etwas geschieht** — UX-13.AK-01/03
// (#1685),
// Marke `new`.
//
// # Was hier nie geschieht
//
// **Ausgepackt wird nichts, gefolgt wird keinem Verweis, ausgeführt wird
// nichts.** Ein Archiv zeigt die Namen seiner Einträge, ein Paketordner die
// seiner Dateien — ein Verweis darin wird als **unsicher** gezählt und nicht
// betreten, ein Name mit `..` oder einem Anfang bei `/` ebenso. Gelesen wird
// höchstens, was die Regel der Art erlaubt ([domain.AnsichtsregelFuer]).

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	// **Die Decoder, deren Maße eine Ansicht nennt** — PNG, JPEG und GIF aus
	// der Standardbibliothek, BMP, TIFF und WebP aus `golang.org/x/image`.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// Dateiansicht ist, was eine Datei über sich hergibt.
type Dateiansicht struct {
	Name        string
	Pfad        string
	Regel       domain.Ansichtsregel
	Groesse     int64
	Geaendert   time.Time
	Verzeichnis bool

	// Text ist der Anfang eines Textes — gültiges UTF-8, Gekuerzt sagt, ob
	// danach noch etwas kommt.
	Text     string
	Gekuerzt bool

	// Bild sind die Bytes eines Bildes samt seinem Typ und, wo sie sich
	// lesen lassen, seinen Maßen.
	Bild    []byte
	Bildtyp string
	Breite  int
	Hoehe   int

	// Seiten ist die Zahl der Seiten eines PDF.
	Seiten int

	// Eintraege sind die Einträge eines Pakets oder Archivs, Unsicher ihre
	// Zahl, die niemand betreten sollte.
	Eintraege []Dateieintrag
	Unsicher  int

	// Verschluesselt, ZuGross und Unlesbar sagen, warum eine Vorschau fehlt.
	Verschluesselt bool
	ZuGross        bool
	Unlesbar       bool
}

// Dateieintrag ist ein Eintrag eines Pakets oder Archivs.
type Dateieintrag struct {
	Name        string
	Groesse     int64
	Verzeichnis bool
	Unsicher    bool
}

// bildtypen sind die Typen, unter denen ein Bild im Fenster erscheint.
var bildtypen = map[string]string{
	"png": "image/png", "jpg": "image/jpeg", "jpeg": "image/jpeg", "gif": "image/gif",
	"webp": "image/webp", "bmp": "image/bmp", "tif": "image/tiff", "tiff": "image/tiff",
	"svg": "image/svg+xml",
}

// DateiAnsehen sagt, was in einer Datei liegt — nach ihrer Regel.
func DateiAnsehen(pfad string) (Dateiansicht, error) {
	angaben, fehler := os.Stat(pfad)
	if errors.Is(fehler, fs.ErrNotExist) {
		return Dateiansicht{}, fmt.Errorf("%w: %s gibt es nicht", domain.ErrNotFound, filepath.Base(pfad))
	}

	if fehler != nil {
		return Dateiansicht{}, domain.MitGrund(domain.ErrValidation, fehler, "die Datei lässt sich nicht lesen")
	}

	ansicht := Dateiansicht{
		Name: filepath.Base(pfad), Pfad: pfad, Regel: domain.AnsichtsregelFuer(pfad),
		Groesse: angaben.Size(), Geaendert: angaben.ModTime(), Verzeichnis: angaben.IsDir(),
	}

	switch {
	case ansicht.Verzeichnis && ansicht.Regel.Art == domain.DateiartPaket:
		ansicht.Eintraege, ansicht.Unsicher = ordnerEintraege(pfad, int(ansicht.Regel.Hoechstens))
		ansicht.Groesse = 0
	case ansicht.Verzeichnis:
		ansicht.Regel = domain.AnsichtsregelFuer("")
	case ansicht.Regel.Vorschau == domain.VorschauKeine && ansicht.Regel.Art != domain.DateiartPDF:
	case ansicht.Groesse > ansicht.Regel.Hoechstens && ansicht.Regel.Vorschau == domain.VorschauBild:
		ansicht.ZuGross = true
	default:
		inhaltAnsehen(&ansicht)
	}

	return ansicht, nil
}

// inhaltAnsehen liest, was die Regel erlaubt, und legt es in die Ansicht.
func inhaltAnsehen(ansicht *Dateiansicht) {
	switch ansicht.Regel.Art {
	case domain.DateiartArchiv, domain.DateiartPaket:
		// **Eine Sicherung mit Kennwort** (OPS-06) beginnt mit dem Kopf von
		// age; in ihr liegt kein lesbares Verzeichnis.
		if sicherung.IstVerschluesselt(ansicht.Pfad) {
			ansicht.Verschluesselt = true

			return
		}

		ansicht.Eintraege, ansicht.Unsicher, ansicht.Unlesbar = archivEintraege(ansicht.Pfad, ansicht.Groesse, int(ansicht.Regel.Hoechstens))

		return
	}

	roh, gekuerzt := anfang(ansicht.Pfad, ansicht.Regel.Hoechstens)

	switch ansicht.Regel.Art {
	case domain.DateiartText:
		ansicht.Text, ansicht.Gekuerzt = strings.ToValidUTF8(string(roh), "�"), gekuerzt
	case domain.DateiartPDF:
		ansicht.Seiten = pdfseiten(roh)
		ansicht.Unlesbar = !bytes.HasPrefix(roh, []byte("%PDF-"))
	default:
		endung := strings.ToLower(strings.TrimPrefix(filepath.Ext(ansicht.Name), "."))
		ansicht.Bild, ansicht.Bildtyp = roh, bildtypen[endung]

		if masse, _, fehler := image.DecodeConfig(bytes.NewReader(roh)); fehler == nil {
			ansicht.Breite, ansicht.Hoehe = masse.Width, masse.Height
		}
	}
}

// anfang liest höchstens `grenze` Bytes — und sagt, ob danach noch etwas kam.
//
// **Ein Lesefehler heißt: nichts gelesen.** Die Datei war eben noch da; wer
// sie zwischen Stat und Lesen wegnimmt, bekommt eine Ansicht ohne Inhalt.
func anfang(pfad string, grenze int64) (roh []byte, gekuerzt bool) {
	datei, fehler := os.Open(pfad) // #nosec G304 -- die Datei, die jemand ansehen will
	if fehler != nil {
		return nil, false
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // nur gelesen

	roh, _ = io.ReadAll(io.LimitReader(datei, grenze+1)) //nolint:errcheck // siehe oben
	if int64(len(roh)) > grenze {
		return roh[:grenze], true
	}

	return roh, false
}

// pdfseiten zählt die Seitenobjekte eines PDF — `/Type /Page`, nicht
// `/Pages`. Das ist eine Zählung und kein Lesen: ein PDF mit Objektströmen
// nennt so weniger Seiten, als es hat.
func pdfseiten(roh []byte) int {
	text := strings.ReplaceAll(string(roh), "/Type /Page", "/Type/Page")

	return strings.Count(text, "/Type/Page") - strings.Count(text, "/Type/Pages")
}

// ordnergang geht durch einen Ordner — eine Naht, damit ein Lauf zählen kann,
// wie weit er geht.
var ordnergang = filepath.WalkDir

// ordnerEintraege listet einen Paketordner, ohne einem Verweis zu folgen.
func ordnerEintraege(wurzel string, hoechstens int) (eintraege []Dateieintrag, unsicher int) {
	_ = ordnergang(wurzel, func(pfad string, eintrag fs.DirEntry, unterwegs error) error { //nolint:errcheck // was sich nicht lesen lässt, fehlt in der Liste
		// **Was sich nicht lesen lässt, fehlt in der Liste** — ein Fehler
		// unterwegs bricht die Ansicht nicht ab, sie zeigt, was sie zeigen kann.
		if unterwegs != nil {
			return nil //nolint:nilerr // siehe oben
		}

		if pfad == wurzel {
			return nil
		}

		// **An der Grenze endet der Gang** — nicht nur die Liste. Sonst stiege
		// `WalkDir` durch jeden weiteren Unterordner (Nachlauf über UX-13).
		if len(eintraege) >= hoechstens {
			return fs.SkipAll
		}

		name := filepath.ToSlash(strings.TrimPrefix(pfad, wurzel+string(filepath.Separator)))
		verweis := eintrag.Type()&fs.ModeSymlink != 0

		var groesse int64
		if angaben, fehler := eintrag.Info(); fehler == nil && eintrag.Type().IsRegular() {
			groesse = angaben.Size()
		}

		eintraege = append(eintraege, Dateieintrag{Name: name, Groesse: groesse, Verzeichnis: eintrag.IsDir(), Unsicher: verweis})
		if verweis {
			unsicher++
		}

		return nil
	})

	return eintraege, unsicher
}

// DateiUebergabePruefen sagt, ob eine Datei an macOS gehen darf — zum Zeigen
// im Finder jede, die es gibt, zum Öffnen nur, was ihre Regel erlaubt
// (UX-13.AK-04).
//
// **Nur ein ganzer Pfad**: ein Name, der mit `-` beginnt, läse `open` als
// Option. **Und kein Ordner zum Öffnen**: ein Programmbündel ist einer.
func DateiUebergabePruefen(pfad string, oeffnen bool) error {
	if !filepath.IsAbs(pfad) {
		return domain.MitGrund(domain.ErrValidation, fmt.Errorf("%q ist kein ganzer Pfad", pfad),
			"nur ein ganzer Pfad geht an macOS")
	}

	angaben, fehler := os.Stat(pfad)
	if fehler != nil {
		return fmt.Errorf("%w: %s gibt es nicht", domain.ErrNotFound, filepath.Base(pfad))
	}

	if oeffnen && (angaben.IsDir() || !domain.AnsichtsregelFuer(pfad).Oeffnen) {
		return fmt.Errorf("%w: %s wird nicht geöffnet — es könnte ein Programm sein; "+
			"zeige es im Finder", domain.ErrValidation, filepath.Base(pfad))
	}

	return nil
}
