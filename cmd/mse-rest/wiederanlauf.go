package main

// **Wer vor einem neueren Bestand endet, endet nicht mit 0** — M02
// (#2346),
// Abnahme 2: Anhalten **und Wiederanlauf**.
//
// Die plist des eingerichteten Dienstes sagt `SuccessfulExit` `false`
// (`internal/app/schnittstellenziele.go`): launchd startet ihn nach einem
// Ende mit einem Code ungleich 0 neu, **höchstens alle zehn Sekunden**
// (`man launchd.plist`, `ThrottleInterval`, nachgelesen am 2026-10-14), und
// nach einem Ende mit 0 nicht. Endete er nach der Fassungswache mit 0, bliebe
// er nach einer Wanderung über die Kommandozeile aus — und der Satz an den
// Rufer verspräche einen Neustart, der nie kommt. **Neu gestartet wird das
// Programm, auf das die plist zeigt**: das im Bündel, nach dem Tausch also das
// der neuen Fassung.

import (
	"errors"
	"sync/atomic"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// errVeraltetBeendet ist der Grund des Endes — `main` schreibt ihn auf die
// Fehlerausgabe, und unter launchd landet er in `dienst.log` (SEC02, #2355).
var errVeraltetBeendet = errors.New("dieser Dienst ist älter als der Bestand und endet — " +
	"eingerichtet startet launchd ihn mit dem Programm der neuen Fassung neu")

// mitWiederanlauf hängt den Grund an, wenn die Fassungswache das Ende auslöste.
func mitWiederanlauf(veraltet *atomic.Bool, fehler error) error {
	if veraltet.Load() {
		return errors.Join(fehler, errVeraltetBeendet)
	}

	return fehler
}

// endcode ist der Rückgabewert eines Laufs, der mit einem Fehler endete —
// **0 über einem Bestand einer neueren Fassung** (Zweitblick über Phase 2,
// Befund 5): zeigt die plist auf ein Programm einer älteren Fassung, startete
// launchd es nach jedem Ende mit 1 wieder, alle zehn Sekunden, ohne Ende. Ein
// Programm, das den Bestand nicht kennt, hat nichts zu tun; der Grund steht
// in der Fehlerausgabe und damit in `dienst.log`.
//
// **Nur der laufende Bestand** ([domain.BestandNeuer], Nachlauf über Phase 2,
// Befund 2): auch eine Vorsicherung aus einer neueren Fassung trägt
// [domain.NeuereFassung] — sie ist dann ein Fehler der Vorbereitung, nach dem
// launchd wieder anwerfen soll.
func endcode(fehler error) int {
	var neuer *domain.BestandNeuer
	if errors.As(fehler, &neuer) {
		return 0
	}

	return 1
}
