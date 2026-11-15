package app

// **Der PATH der Shell** — Befund 9 des Zweitblicks über Phase 6.4
// (#1665).
//
// # Warum gefragt wird
//
// **Ein Programm, das über LaunchServices startet, erbt nicht den PATH der
// Shell**, sondern den von `launchd` — `/usr/bin:/bin:/usr/sbin:/sbin`.
// „`mse` liegt im PATH" hieß am Fenster deshalb nichts: ein `mse` unter
// `/opt/homebrew/bin` blieb unerkannt, und `~/.local/bin` fehlte immer.
//
// # Wie gefragt wird
//
// **zsh, als Anmeldeshell und interaktiv**, weil Homebrew und die meisten
// Einrichtungen den PATH in `.zprofile` oder `.zshrc` setzen — und zsh ist
// seit macOS 10.15 die Vorgabe. **Nicht `$SHELL`**: ein Programm, dessen Name
// aus der Umgebung kommt, weist `internal/schichten` zu Recht ab (#1170). Wer
// eine andere Shell benutzt, bekommt, was zsh daraus macht, mindestens also
// `/etc/paths` über `path_helper`.
//
// Gelesen wird nur, was zwischen zwei Marken steht: eine Shell, die beim
// Start etwas sagt, verdirbt die Antwort sonst. **Mit einer Frist**, und
// scheitert es, bleibt der eigene PATH — das ist der bisherige Stand, nie ein
// schlechterer.

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// pfadmarke umschließt den PATH in der Antwort der Shell.
const pfadmarke = "<<MSE-PFAD>>"

// shellpfad fragt zsh nach ihrem PATH — und gibt `rueckfall`, wenn sie nicht
// antwortet.
//
// **`laufen` kommt als Wert**, damit ein Lauf die Antwort vorgeben kann, ohne
// die Einrichtung des Rechners auszuführen; das Fenster reicht
// `(*exec.Cmd).Output`.
func shellpfad(ctx context.Context, laufen func(*exec.Cmd) ([]byte, error), rueckfall string) string {
	ctx, aufgeben := context.WithTimeout(ctx, 3*time.Second)
	defer aufgeben()

	befehl := exec.CommandContext(ctx, "/bin/zsh", "-ilc",
		`printf '%s%s%s' '`+pfadmarke+`' "$PATH" '`+pfadmarke+`'`)

	// **Die Frist bindet auch die Ausgabe** — Nachlauf über Phase 6.4, N7:
	// startet `.zshrc` einen Hintergrundprozess, der sie offen hält, wartete
	// `Output` ohne diese Zeile, bis er endet (`os/exec`, `WaitDelay`).
	befehl.WaitDelay = time.Second

	antwort, fehler := laufen(befehl)
	if fehler != nil {
		return rueckfall
	}

	// **Und ohne beide Marken gilt der eigene** — eine Shell, die mit 0
	// endet und nichts sagt, gab sonst einen leeren PATH.
	_, rest, anfang := strings.Cut(string(antwort), pfadmarke)
	pfad, _, ende := strings.Cut(rest, pfadmarke)

	if !anfang || !ende {
		return rueckfall
	}

	return pfad
}
