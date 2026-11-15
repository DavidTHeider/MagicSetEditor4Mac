//go:build darwin

package app

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"time"
)

// errNichtNoetig sagt, dass die Meldung schon zu sehen ist.
//
// **Ein Fehler und kein stilles Nichts**: der Aufrufer schreibt sie dann nach
// stderr, und genau das ist hier das Richtige — dort hängt ein Terminal.
var errNichtNoetig = errors.New("die Meldung steht schon auf der Fehlerausgabe")

// meldungZeigen zeigt unter macOS ein Fenster, bevor die Anwendung eines hat.
//
// # Warum osascript
//
// Es ist der Weg, den macOS ohne jede Abhängigkeit mitbringt, und er braucht
// weder ein laufendes Wails noch ein Bündel. Ein zweites Fenster-Werkzeug
// aufzunehmen, nur um einen Satz zu zeigen, wäre teurer als der Satz.
//
// # Warum der Text über argv geht und nicht in das Skript
//
// **Ein Anführungszeichen in einer Fehlermeldung ist ein Anführungszeichen, kein
// Skriptende.** Wer die Meldung in den Quelltext des AppleScript einsetzt,
// baut sich eine Einschleusung — und Fehlermeldungen enthalten hier Dateinamen
// und Migrationsnamen, also fremden Text. Über `on run argv` kommt er als
// **Wert** an und wird nie ausgewertet.
//
// # Warum eine Frist
//
// Ein Dialog wartet auf einen Klick. Kommt keiner — weil niemand am Rechner
// sitzt, etwa in einem Bauplan —, hinge der Start für immer. Nach zwei Minuten
// gibt der Aufruf auf; der Text steht dann noch immer auf stderr.
const anzeigen = `on run argv
	display dialog (item 2 of argv) with title (item 1 of argv) ¬
		buttons {"OK"} default button "OK" with icon stop
end run`

// meldefrist ist die Zeit, die auf einen Klick gewartet wird.
var meldefrist = 2 * time.Minute

// amTerminal sagt, ob die Fehlerausgabe an einem Terminal hängt.
//
// **Dann braucht es kein Fenster** — der Satz steht schon da, wo jemand
// hinsieht (#834).
//
// # Warum genau dieses Signal
//
// Der Zweitblick über Phase O hat angestrichen, dass ein Startfehler einen
// unbeaufsichtigten Lauf bis zu zwei Minuten hinhalten kann: der Dialog wartet
// auf einen Klick, und wo niemand sitzt, kommt keiner.
//
// **Dieses Projekt kennt genau zwei Umgebungen** ([0007](../../docs/entscheidungen/0007-kein-entfernter-lauf.md)):
// den Lauf aus einem Terminal — `make ci`, `go run`, ein Entwicklungsstart —
// und das doppelgeklickte Bündel. Im ersten hängt stderr an einem Terminal und
// die Meldung ist zu sehen; im zweiten hängt es an nichts, und genau dort wird
// das Fenster gebraucht. Ein entfernter Lauf, für den das Signal nicht
// entschiede, findet nicht statt.
//
// **Was es nicht abdeckt**: ein Lauf, der stderr in eine Datei umlenkt und
// trotzdem niemanden am Rechner hat. Der wartet weiter bis zur Frist — und die
// Frist ist der Grund, warum er nicht ewig wartet.
func amTerminal() bool {
	zustand, fehler := os.Stderr.Stat()
	if fehler != nil {
		return false
	}

	return zustand.Mode()&os.ModeCharDevice != 0
}

func meldungZeigen(titel, text string) error {
	if amTerminal() {
		return errNichtNoetig
	}

	ctx, aufgeben := context.WithTimeout(context.Background(), meldefrist)
	defer aufgeben()

	// #nosec G204 -- der Pfad ist fest, und der veränderliche Teil steht in
	// argv und nicht im Skript.
	return exec.CommandContext(ctx,
		"/usr/bin/osascript", "-e", anzeigen, titel, text).Run()
}
