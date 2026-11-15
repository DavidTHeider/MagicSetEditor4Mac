package main

// **Der Stand eines langen Befehls** — FMT-03.AK-02
// (#1692).
//
// `mse bogen`, `mse oeffnen`, `mse vorlagen … ausfuehren` und die übrigen
// langen Läufe melden über `service.Schritt`, wie weit sie sind. Hier wird
// daraus **eine Zeile auf dem Fehlerstrom, die sich überschreibt** — und nur,
// wenn er ein Terminal ist: in einer Pipe wäre sie Rauschen. Abbrechen ist
// ⌃C (`main.go`), wiederholen ist derselbe Befehl noch einmal.

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// zeilenpause ist, wie lange zwischen zwei Zeilen mindestens liegt.
const zeilenpause = 100 * time.Millisecond

// terminalmelder gibt einen Melder, wenn der Fehlerstrom ein Zeichengerät ist.
func terminalmelder(fehlerstrom io.Writer) (domain.Fortschrittsmelder, bool) {
	datei, ist := fehlerstrom.(*os.File)
	if !ist {
		return nil, false
	}

	info, fehler := datei.Stat()
	if fehler != nil || info.Mode()&os.ModeCharDevice == 0 {
		return nil, false
	}

	return zeilenmelder(datei, time.Now), true
}

// zeilenmelder schreibt steigende Stände in eine Zeile — gedrosselt, die
// erste und die letzte immer, und nach der letzten ein Umbruch.
func zeilenmelder(ziel io.Writer, jetzt func() time.Time) domain.Fortschrittsmelder {
	var (
		sperre   sync.Mutex
		zuletzt  = -1
		gesendet time.Time
	)

	return func(stand domain.Fortschritt) {
		sperre.Lock()
		defer sperre.Unlock()

		letzter := stand.Gesamt > 0 && stand.Erledigt >= stand.Gesamt
		if stand.Erledigt <= zuletzt || (zuletzt >= 0 && !letzter && jetzt().Sub(gesendet) < zeilenpause) {
			return
		}

		ende := ""
		if letzter {
			ende = "\n"
		}

		// **Ein Schreibfehler hält den Befehl nicht an** — der Stand ist
		// eine Zugabe, und die Ausgabe des Befehls sagt ihn ohnehin.
		_, _ = fmt.Fprintf(ziel, "\r%s\x1b[K%s", stand.Satz(), ende) //nolint:errcheck // siehe darüber

		zuletzt, gesendet = stand.Erledigt, jetzt()
	}
}
