package app

// **Die Ansage eines Dienstes unter launchd, je Start neu** — SEC02
// (#2355),
// Befund A-16 des Audits 0.9.0.
//
// # Was vorher war
//
// **launchd hängte an.** `StandardOutPath` und `StandardErrorPath` zeigten
// beide auf `schnittstellen/dienst.txt`; jede Ansage samt Marke kam dazu,
// die alten blieben, und die Fehlerausgabe wuchs ohne Grenze in dieselbe
// Datei. Nach zehn Neustarts standen zehn Marken darin — gültig ist nur die
// letzte, aber wer die Datei las, sah sie alle.
//
// # Was jetzt gilt
//
// **Der Dienst schreibt seine Ansage selbst** (`-ansage`): je Start eine
// Nachbardatei mit 0600, die unter den Namen getauscht wird — die vorige
// Marke ist danach fort, und eine halbe Ansage steht nie unter dem Namen.
// Der Ordner entsteht mit 0700. Die Fehlerausgabe geht in eine eigene
// Datei, und die hat eine Grenze ([FehlerausgabeBegrenzen]). Der
// Lebenszyklus der Marke steht in Vermerk 0192.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Fehlerausgabegrenze ist, ab wann die Fehlerausgabe eines Dienstes unter
// launchd beim Start neu beginnt.
const Fehlerausgabegrenze = 1 << 20

// AnsageInDatei schreibt die Ansage nach `pfad` und ersetzt, was dort stand.
func AnsageInDatei(pfad string, ansage Dienstansage) error {
	ordner := filepath.Dir(pfad)
	if fehler := service.PrivaterOrdner(ordner); fehler != nil {
		return fmt.Errorf("den Ordner der Ansage anlegen: %w", fehler)
	}

	zwischen, fehler := os.CreateTemp(ordner, ".dienst-*.txt")
	if fehler != nil {
		return fmt.Errorf("die Ansage anlegen: %w", fehler)
	}

	_, schreibfehler := zwischen.WriteString(ansage.Text())

	fehler = errors.Join(schreibfehler, zwischen.Sync(), zwischen.Close())
	if fehler == nil {
		fehler = os.Rename(zwischen.Name(), pfad)
	}

	if fehler != nil {
		_ = os.Remove(zwischen.Name()) //nolint:errcheck // aufgeräumt wird, was nicht getauscht wurde

		return fmt.Errorf("die Ansage schreiben: %w", fehler)
	}

	return nil
}

// FehlerausgabeBegrenzen beginnt die Fehlerausgabe neu, wenn sie eine Datei
// über `grenze` ist — und sagt es in ihrer ersten Zeile. **Ein Rohr oder ein
// Endgerät bleibt, wie es ist**: dort hat niemand etwas aufzuheben.
func FehlerausgabeBegrenzen(ausgabe *os.File, grenze int64) error {
	angaben, fehler := ausgabe.Stat()
	if fehler != nil || !angaben.Mode().IsRegular() || angaben.Size() <= grenze {
		return nil //nolint:nilerr // ohne lesbare Größe gibt es nichts zu kürzen
	}

	fehler = ausgabe.Truncate(0)
	if fehler == nil {
		_, fehler = fmt.Fprintf(ausgabe, "%s die Fehlerausgabe begann neu — sie war über %d Bytes gekürzt\n",
			time.Now().UTC().Format(time.RFC3339), grenze)
	}

	if fehler != nil {
		return fmt.Errorf("die Fehlerausgabe kürzen: %w", fehler)
	}

	return nil
}
