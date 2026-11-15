//go:build darwin

package app

import (
	"context"
	"fmt"
	"io"
	"os/exec"
)

// systemprotokoll gibt den Schreiber ins Unified Log.
//
// # Warum über `/usr/bin/logger`
//
// `os_log` ist eine C-Schnittstelle, und cgo ist hier ausgeschlossen
// ([0002](../../docs/entscheidungen/0002-sqlite-als-arbeitsformat.md),
// [0042](../../docs/entscheidungen/0042-keine-verschluesselung-des-bestands.md)).
// `logger` ist der Weg, den macOS ohne jede Abhängigkeit mitbringt —
// **nachgemessen am 2026-08-22**: ein Aufruf, und `log show --last 60s`
// findet den Eintrag.
//
// # Was er kostet
//
// **Ein laufender `logger`** liest von seiner Standardeingabe, eine Zeile je
// Eintrag — gemessen am 2026-08-22: zwei Zeilen, eine PID, zwei Einträge. Bis
// PERF03 startete jede Zeile einen eigenen, 2,7 ms und synchron im
// Schreibenden; jetzt legt der Schreibende die Zeile in einen begrenzten
// Kanal und geht (`systemkanal.go`).
//
// # Was er nicht kann
//
// **Kein Subsystem und keine Kategorie.** Die setzt nur `os_log`; über
// `logger` steht der Eintrag unter dem Prozessnamen `logger`. Gefiltert wird
// deshalb über den Text, nicht über ein Prädikat auf das Subsystem.
func systemprotokoll() io.Writer { return neuerSystemkanal(loggerStarterMit((*exec.Cmd).Start)) }

// loggerStarterMit gibt den Starter eines `logger`, der von der
// Standardeingabe liest — `starte` startet ihn, damit ein Lauf einen
// scheiternden belegen kann. **Der Pfad bleibt ein Literal**:
// `internal/schichten` weist ein Programm aus einer Veränderlichen ab (#1170).
func loggerStarterMit(starte func(*exec.Cmd) error) func() (laufenderProzess, error) {
	return func() (laufenderProzess, error) {
		// #nosec G204 -- Pfad und Marke sind fest; was er schreibt, liest er
		// von der Standardeingabe und nie als Befehl (#797)
		befehl := exec.CommandContext(context.Background(), "/usr/bin/logger", "-t", "MagicSetEditor4Mac")

		// **Die Leitung des Prozesses selbst und kein `io.Pipe`** (Zweitblick
		// über Phase 3, Befund 6): über einen `io.Reader` kopiert `os/exec` in
		// eine eigene Leitung, endet beim ersten Schreibfehler und schließt
		// den Reader nicht — die nächste Zeile wartete auf einen Leser, den es
		// nicht mehr gab. Über `StdinPipe` gibt ein toter `logger` `EPIPE`
		// zurück, und der Kanal startet neu. **`StdinPipe` scheitert hier
		// nicht** — es scheitert nur an einem gesetzten `Stdin` oder einem
		// schon gestarteten Prozess —, und ein eigener Zweig dafür wäre einer,
		// den kein Lauf betritt (§3).
		ein, fehler := befehl.StdinPipe()
		if fehler == nil {
			if fehler = starte(befehl); fehler != nil {
				_ = ein.Close() //nolint:errcheck // nie benutzt
			}
		}

		if fehler != nil {
			return laufenderProzess{}, fmt.Errorf("logger starten: %w", fehler)
		}

		return laufenderProzess{ein: ein, warten: befehl.Wait, abbrechen: befehl.Process.Kill}, nil
	}
}
