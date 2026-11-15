package service

// Die **Ablage** für Bilder, die jemand in der Anwendung einsetzt —
// #246.
//
// # Wohin, und warum dorthin
//
// [0029](../../docs/entscheidungen/0029-bilder-liegen-neben-dem-set.md): neben
// das Set, nicht hinein. Die Bilder eines Sets machen **98,9 %** seines
// Gewichts aus — gemessen über die zehn kanonischen Sets, 1.996 MB gegen 22,9
// MB —, und ein Bestand, der sie führte, wöge das Neunzigfache.
//
// Ins **Herkunftspaket** zu schreiben, so wie MSE es tut, geht hier nicht: MSE
// *ist* das Paket, dieses Programm hat den Bestand dazwischen. Ein
// Herkunftspaket darf schreibgeschützt sein, auf einem abgezogenen Datenträger
// liegen oder gar nicht existieren.

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Bildablage schreibt Bilder in ein Verzeichnis je Set.
type Bildablage struct {
	wurzel string

	// hoechstens ist die Schranke dieser Ablage — 0 heißt [hoechsteBildzahl].
	hoechstens int
}

// NeueBildablage baut die Ablage über einer Wurzel.
//
// **Eine leere Wurzel ist erlaubt und scheitert beim Ablegen.** Die Wurzel
// kommt aus dem Benutzerverzeichnis, und das lässt sich nicht ermitteln, wenn
// die Umgebung kaputt ist; ein `NeueBildablage`, das dann selbst scheiterte,
// verhinderte den Start der ganzen Anwendung wegen einer Sache, die die
// wenigsten benutzen.
func NeueBildablage(wurzel string) *Bildablage {
	return &Bildablage{wurzel: wurzel}
}

// Wurzel gibt das Verzeichnis, unter dem die Sets liegen — leer, wenn es
// keines gibt.
//
// **Sie steht hier und wird nicht daneben mitgeführt.** Das Zeichenwerk braucht
// sie, um eingesetzte Bilder zu finden; ein zweiter Weg zu derselben Zahl liefe
// mit diesem auseinander.
func (b *Bildablage) Wurzel() string { return b.wurzel }

// MitHoechsterBildzahl setzt die Schranke dieser Ablage — **für Tests**.
//
// # Warum an der Ablage und nicht am Paket
//
// Ein Paketwert wäre ein Zustand, den **alle** Läufe teilen — und die Läufe
// dieses Pakets sind nebenläufig. Ein Lauf, der ihn herunterdreht, ließe
// jeden anderen, der gerade ein Bild ablegt, an der Schranke scheitern.
// Genau das ist am 2026-08-26 passiert: zehn Läufe wurden rot, und keiner
// von ihnen hatte mit Schranken zu tun.
//
// **Ohne sie müsste ein Lauf, der das Aufgeben prüfen will,
// [hoechsteBildzahl] Namen belegen**; das ist keine Prüfung, sondern eine
// Wartezeit (§3).
func (b *Bildablage) MitHoechsterBildzahl(neu int) *Bildablage {
	b.hoechstens = neu

	return b
}

// schranke gibt die Zahl, bei der diese Ablage aufgibt.
func (b *Bildablage) schranke() int {
	if b.hoechstens > 0 {
		return b.hoechstens
	}

	return hoechsteBildzahl
}

// Verzeichnis gibt den Ort der Bilder eines Sets — leer ohne Wurzel.
func (b *Bildablage) Verzeichnis(setID int64) string {
	if b.wurzel == "" {
		return ""
	}

	return filepath.Join(b.wurzel, strconv.FormatInt(setID, 10))
}

// Ablegen schreibt ein Bild und gibt den Namen, unter dem es im Set steht.
//
// # Der Name folgt MSE
//
//	name = prefix; name << ++infix; name += suffix;
//	name = normalize_internal_filename(name);
//	FileInfos::iterator it = files.find(name);
//	if (it == files.end()) { … return name; }
//
// `Package::newFileName`, `src/util/io/package.cpp`, Zeilen 248–267, Fassung
// f9d9356 vom 2021-07-06, gerufen mit dem **Feldnamen** als Präfix
// (`ImageValueEditor::sliceImage`, `src/gui/value/image.cpp`, Zeile 49).
// `normalize_internal_filename` schreibt klein und macht aus jedem
// Rückwärtsschrägstrich einen Schrägstrich (`src/util/file_utils.cpp`, Zeilen
// 24–31).
//
// **Die Zahl läuft je Verzeichnis, nicht je Feld.** MSE fragt gegen **alle**
// Dateien des Pakets; ein Zähler je Feld gäbe zwei Karten denselben Namen, und
// die zweite überschriebe die erste.
//
// # Und PNG, gleich was hereinkam
//
//	img.SaveFile(…, wxBITMAP_TYPE_PNG); // always use PNG images, see #69.
//
// `src/gui/value/image.cpp`, Zeile 51.
func (b *Bildablage) Ablegen(setID int64, feld string, bild image.Image) (string, error) {
	ort := b.Verzeichnis(setID)
	if ort == "" {
		return "", fmt.Errorf("%w: für Bilder ist kein Verzeichnis eingerichtet",
			domain.ErrValidation)
	}

	if err := os.MkdirAll(ort, 0o750); err != nil {
		return "", fmt.Errorf("Bildverzeichnis anlegen: %w", err)
	}

	// **Erst in den Speicher, dann auf die Platte.** Ein `png.Encode` in die
	// offene Datei spart einen Puffer und handelt sich zwei Fehlerwege ein, die
	// sich nicht prüfen lassen: ein gescheitertes `Close` nach einem geglückten
	// Schreiben ist auf einem Netzlaufwerk möglich und in keinem Test.
	var gepackt bytes.Buffer
	if err := png.Encode(&gepackt, bild); err != nil {
		return "", fmt.Errorf("das Bild ließ sich nicht kodieren: %w", err)
	}

	return namenBelegen(ort, feld, gepackt.Bytes(), b.schranke(), fuelleUndSchliesse)
}

// Wegnehmen entfernt ein abgelegtes Bild wieder.
//
// # Wofür es das gibt
//
// [Bildablage.Ablegen] schreibt die Datei, **bevor** ein Feld auf sie zeigt.
// Scheitert das Schreiben des Feldes, liegt eine Datei ohne Verweis da: das
// Bilderverzeichnis wüchse bei jedem Fehlversuch um eine, die niemand je
// wieder findet (#1163,
// Befund Z7-02).
//
// **Ein fehlendes Bild ist kein Fehler.** Der Aufrufer räumt auf einem
// Fehlerweg auf; dass dort schon jemand aufgeräumt hat, ist keine schlechte
// Nachricht.
//
// **Und ein leerer Name auch nicht.** Er heißt, dass gar nichts abgelegt
// wurde — dann ist nichts wegzunehmen.
func (b *Bildablage) Wegnehmen(setID int64, name string) error {
	ort := b.Verzeichnis(setID)
	if ort == "" || name == "" {
		return nil
	}

	// **Nur der Name, kein Weg.** Ein `../` darin führte aus dem
	// Bilderverzeichnis hinaus, und was hier gelöscht wird, entscheidet
	// nicht der Aufrufer.
	if name != filepath.Base(name) {
		return fmt.Errorf("%w: %q ist kein Bildname", domain.ErrValidation, name)
	}

	if fehler := os.Remove(filepath.Join(ort, name)); fehler != nil &&
		!errors.Is(fehler, os.ErrNotExist) {
		return fmt.Errorf("das Bild wegnehmen: %w", fehler)
	}

	return nil
}

// hoechsteBildzahl ist die Schranke, an der [freierName] aufgibt.
//
// **MSE hat keine** — dort ist die Dateiliste im Speicher, und `find`
// scheitert nicht (`Package::newFileName`, `src/util/io/package.cpp`, Zeilen
// 248–267, Fassung f9d9356 vom 2021-07-06). Hier fragt jeder Durchlauf die
// Platte, und ein Verzeichnis, das jeden Namen belegt meldet (ein kaputter
// Einhängepunkt etwa), drehte für immer.
//
// **Eine Variable und keine Konstante**, damit der Weg dahinter geprüft werden
// kann, ohne hunderttausend Dateien anzulegen — die Zahl selbst ist so hoch,
// dass sie im Betrieb niemanden trifft: das größte kanonische Set trägt 3.837
// Bilder.
const hoechsteBildzahl = 100000

// freierName sucht den ersten Namen, den es noch nicht gibt.
// **Die Endung gehört in die Prüfung, nicht dahinter.** Wer sie danach anhängt,
// sucht nach `symbol1` und schreibt `symbol1.mse-symbol` — und findet beim
// nächsten Mal wieder nichts, weil er wieder nach `symbol1` sucht. Zwei
// Zeichnungen bekämen denselben Namen, und die zweite nähme der ersten ihren
// Platz.
func freierName(ort, feld, endung string, hoechstens int) (string, error) {
	rumpf := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(feld), `\`, "/"))

	for zahl := 1; zahl <= hoechstens; zahl++ {
		name := rumpf + strconv.Itoa(zahl) + endung

		_, err := os.Stat(filepath.Join(ort, name))
		if errors.Is(err, os.ErrNotExist) {
			return name, nil
		}

		if err != nil {
			return "", fmt.Errorf("Bildverzeichnis lesen: %w", err)
		}
	}

	return "", fmt.Errorf("%w: für %q ist kein freier Name mehr da",
		domain.ErrConflict, feld)
}

// namenBelegen sucht einen freien Namen und **belegt** ihn beim Anlegen.
//
// # Warum nicht [freierName] und dann schreiben
//
// `freierName` fragt mit `Stat` und lässt den Namen dann liegen. Zwei
// gleichzeitige Einsetzungen im selben Set bekamen beide `kunst1` und
// schrieben in dieselbe Datei — und wenn danach **einer** von beiden
// Datenbankzügen scheiterte, nahm sein Aufräumen dem anderen das Bild weg.
// Übrig blieb ein gültiger Kartenwert ohne Bilddatei.
//
// **`O_EXCL` entscheidet, wer den Namen bekommt.** Wer ihn nicht bekommt,
// sucht weiter — dieselbe Schleife, nur ist das Ergebnis jetzt eine Zusage
// und keine Momentaufnahme (Nachlauf über Phase 3.2,
// #1163).
func namenBelegen(
	ort, feld string, inhalt []byte, hoechstens int,
	fuelle func(io.WriteCloser, []byte) error,
) (string, error) {
	rumpf := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(feld), `\`, "/"))

	// **Und hier noch einmal, für jeden Aufrufer.**
	//
	// `internal/service/ablagebilder.go` weist einen Weg schon ab, bevor er
	// hierher kommt. Diese Zeile ist die Grenze selbst: sie gilt auch für den
	// **nächsten** Aufrufer, der einen Namen von außen bekommt und nicht daran
	// denkt.
	//
	// Der Grund steht bei [Bildablage.Bytes]: `filepath.Join` **räumt** `..`
	// auf, statt es abzuweisen — aus `../1/bild` wird eine Datei im
	// Verzeichnis eines anderen Sets. Der `#nosec`-Vermerk weiter unten sagt,
	// der Name entstehe intern; das stimmte, bis ein Aufrufer dazukam, für
	// den es nicht stimmte (#1282,
	// Befund 3).
	if rumpf == "" || strings.Contains(rumpf, "/") || strings.Contains(rumpf, "..") {
		return "", fmt.Errorf(
			"%w: %q ist kein einfacher Dateiname", domain.ErrValidation, feld)
	}

	for zahl := 1; zahl <= hoechstens; zahl++ {
		name := rumpf + strconv.Itoa(zahl)
		pfad := filepath.Join(ort, name)

		// #nosec G304 -- der Name entsteht hier und kommt nicht als Eingabe
		datei, err := os.OpenFile(pfad, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if errors.Is(err, os.ErrExist) {
			// **Belegt.** Von einem früheren Bild oder von einer anderen
			// Einsetzung, die im selben Augenblick läuft — der Unterschied
			// zählt hier nicht.
			continue
		}

		// **Durch den Griff, den die Reservierung schon hält.**
		//
		// Hier wurde vorher geschlossen und derselbe Pfad ein zweites Mal
		// geöffnet. Scheiterte dieser zweite Zug, blieb die reservierte Datei
		// liegen — und ihr Name war für immer verbraucht, weil die Schleife
		// ihn ab dann als belegt überspringt. Angestrichen vom zweiten
		// Zweitblick über die Freigabe 0.3.0 (#1170, Befund 4).
		if err == nil {
			err = fuelle(datei, inhalt)
		}

		if err != nil {
			// **Die Reservierung wird zurückgenommen.** Ein leerer Name, den
			// niemand mehr vergibt, ist die zweite Sorte Schaden — und er ginge
			// in die nächste Sicherung mit.
			return "", fmt.Errorf("Bilddatei schreiben: %w",
				errors.Join(err, os.Remove(pfad)))
		}

		return name, nil
	}

	return "", fmt.Errorf("%w: für %q ist kein freier Name mehr da",
		domain.ErrConflict, feld)
}

// Bytes legt fertige Bildbytes unter einem neuen Namen ab.
//
// **Für die Zwischenablage** (#1237):
// dort kommt ein Bild als Bytes an, nicht als [image.Image]. Es noch einmal zu
// entschlüsseln und neu zu kodieren kostete Zeit und Genauigkeit — ein PNG
// bliebe eines, ein JPEG würde eines.
//
// **Der Name wird neu vergeben**, mit dem alten als Präfix: im Zielset kann
// `image1` schon belegt sein, und ihn zu überschreiben nähme einer anderen
// Karte ihr Bild.
func (b *Bildablage) Bytes(setID int64, name string, inhalt []byte) (string, error) {
	ort := b.Verzeichnis(setID)
	if ort == "" {
		return "", fmt.Errorf("%w: für Bilder ist kein Verzeichnis eingerichtet",
			domain.ErrConflict)
	}

	if err := os.MkdirAll(ort, 0o700); err != nil {
		return "", fmt.Errorf("das Bilderverzeichnis anlegen: %w", err)
	}

	return namenBelegen(ort, ohneZiffern(name), inhalt, b.schranke(), fuelleUndSchliesse)
}

// ohneZiffern gibt den Namensstamm — `image1` wird zu `image`.
//
// **Sonst hieße die Datei `image11`**, und der nächste `image111`. Der Stamm
// ist das, was [Bildablage.Ablegen] als Präfix nimmt.
func ohneZiffern(name string) string {
	schnitt := len(name)
	for schnitt > 0 && name[schnitt-1] >= '0' && name[schnitt-1] <= '9' {
		schnitt--
	}

	if schnitt == 0 {
		return name
	}

	return name[:schnitt]
}
