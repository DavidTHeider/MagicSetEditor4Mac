package service

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// PrivaterOrdner legt einen Ordner an, den nur sein Besitzer betritt — 0700 —,
// oder nimmt einem bestehenden, was Gruppe und andere darin dürfen. **Einem
// Verweis an seiner Stelle wird nicht gefolgt.**
//
// Zweitblick über Phase 5, Befund 3 (SEC02, #2355): `os.MkdirAll(ordner,
// 0o700)` ändert die Rechte eines bestehenden Ordners nicht und folgt einem
// Verweis auf einen anderen. Die Zusage „0600 im 0700-Ordner" galt so nur für
// einen frischen.
//
// # Der ganze Weg
//
// **Nachlauf über Phase 5, Befund 3**: `MkdirAll` folgt auch Verweisen **im
// Elternpfad**, und geprüft war nur die letzte Komponente — bei `eltern →
// fremd` entstand der Ordner im fremden Ziel. Ein Verweis ist dort nicht
// verdächtig (`/var` ist einer, und `~/Library` kann einer sein); verdächtig
// ist, **wer den Weg verändern kann**. Deshalb wird **vor** dem Anlegen jeder
// Vorfahr des aufgelösten Wegs geprüft, bis zur Wurzel: er gehört diesem
// Nutzer oder `root`, und weder Gruppe noch andere dürfen darin schreiben —
// außer mit Sticky-Bit wie `/tmp`, wo niemand einen fremden Eintrag
// umbenennen kann. Dieselbe Regel wie OpenSSHs `StrictModes`.
//
// **Geändert wird über den geöffneten Ordner**: geöffnet ohne einem Verweis
// zu folgen, die Rechte über das Handle — so trifft die Änderung, was
// geprüft wurde, und nicht, was inzwischen unter dem Namen liegt.
//
// # Was das nicht abfängt
//
// **Einen Tausch durch denselben Nutzer** zwischen dieser Prüfung und dem
// Schreiben der Datei, die der Aufrufer danach über den Pfad anlegt: wer den
// Weg verändern kann, ist nach der Prüfung oben nur noch er selbst oder
// `root`. Und auf den beschriebenen Plattformen ohne `unix` prüft
// `privaterordner_sonst.go` keine Besitzer.
func PrivaterOrdner(ordner string) error {
	if fehler := wegPruefen(ordner); fehler != nil {
		return fehler
	}

	if fehler := os.MkdirAll(ordner, 0o700); fehler != nil {
		return fmt.Errorf("den Ordner %s anlegen: %w", ordner, fehler)
	}

	datei, fehler := ordnerOeffnen(ordner)
	if fehler != nil {
		return fmt.Errorf("%w: %s ist kein eigener Ordner — einem Verweis wird nicht gefolgt",
			domain.ErrConflict, ordner)
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // nur gelesen

	besitzer, rechte, fehler := eigenesVon(datei)
	if fehler != nil || besitzer != ich {
		return fmt.Errorf("%w: %s gehört einem anderen", domain.ErrConflict, ordner)
	}

	// **Nur wegnehmen, nie dazugeben**: was Gruppe und andere dürfen, fällt;
	// ein Ordner, in dem der Besitzer selbst nicht schreiben darf, bleibt so —
	// und nimmt dann keine Ansage an.
	if rechte&0o077 == 0 {
		return nil
	}

	return datei.Chmod(rechte &^ 0o077) //nolint:wrapcheck // der Aufrufer sagt, wofür der Ordner ist
}

// wegPruefen prüft jeden Vorfahren des Ordners, den es schon gibt — aufgelöst
// und bis zur Wurzel.
func wegPruefen(ordner string) error {
	if voll, fehler := filepath.Abs(ordner); fehler == nil {
		ordner = voll
	}

	vorhanden := filepath.Dir(ordner)

	for {
		_, fehler := os.Lstat(vorhanden)
		if fehler == nil {
			break
		}

		if !errors.Is(fehler, fs.ErrNotExist) {
			return fmt.Errorf("den Weg zu %s prüfen: %w", ordner, fehler)
		}

		vorhanden = filepath.Dir(vorhanden)
	}

	weg, fehler := filepath.EvalSymlinks(vorhanden)
	if fehler != nil {
		return fmt.Errorf("den Weg zu %s auflösen: %w", ordner, fehler)
	}

	for {
		ort, fehler := ortVon(weg)
		if fehler != nil || (ort.besitzer != ich && ort.besitzer != 0) || (ort.offen && !ort.klebend) {
			return fmt.Errorf("%w: auf dem Weg zu %s liegt %s, den ein anderer verändern kann",
				domain.ErrConflict, ordner, weg)
		}

		oben := filepath.Dir(weg)
		if oben == weg {
			return nil
		}

		weg = oben
	}
}

// ort ist, was an einem Ordner auf dem Weg zählt.
type ort struct {
	besitzer uint32

	// offen heißt: Gruppe oder andere dürfen darin schreiben.
	offen bool

	// klebend ist das Sticky-Bit — dann benennt niemand einen fremden Eintrag
	// um.
	klebend bool
}
