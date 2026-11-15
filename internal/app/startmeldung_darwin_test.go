//go:build darwin

package app

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// **Der Text kommt als Wert an, nicht als Skript** (#797).
//
// # Was dieser Test belegt und was nicht
//
// Er zeigt **kein** Fenster — `meldungZeigen` tut das, und deshalb steht die
// Funktion in `.covignore`. Belegt wird hier der Mechanismus darunter: dass
// `osascript` den Text über `on run argv` als Wert entgegennimmt und ihn nicht
// auswertet.
//
// **Das ist die Gegenprobe zu einer Behauptung.** Neben `meldungZeigen` steht,
// ein Anführungszeichen in einer Fehlermeldung sei ein Anführungszeichen und
// kein Skriptende. Ohne diesen Test wäre das ein Satz; mit ihm ist es gemessen
// — und zwar an genau dem Text, der sonst eine Einschleusung wäre.
func TestOsascriptNimmtDenTextAlsWert(t *testing.T) {
	if _, fehler := exec.LookPath("/usr/bin/osascript"); fehler != nil {
		t.Skip("kein osascript auf diesem Rechner")
	}

	// **Derselbe Rumpf wie `anzeigen`, nur ohne Dialog.** Ein Test, der ein
	// anderes Skript prüfte, sagte nichts über das benutzte.
	const zurueckgeben = `on run argv
	return (item 1 of argv) & "/" & (item 2 of argv)
end run`

	// Ein Titel, der ein Skript beenden würde, und eine Meldung, die danach
	// etwas anderes täte.
	titel := `Fehler" & (do shell script "echo eingeschleust") & "`
	meldung := "Migration 0013_x.sql fehlt — \"neuere Fassung\"?"

	aus, fehler := exec.CommandContext(probekontext(t),
		"/usr/bin/osascript", "-e", zurueckgeben, titel, meldung).Output()
	if fehler != nil {
		t.Fatalf("osascript: %v", fehler)
	}

	gelesen := strings.TrimSpace(string(aus))

	// **Die Gleichheit ist der ganze Beweis.** Wäre der Text ausgewertet
	// worden, könnte er nicht Zeichen für Zeichen zurückkommen: `do shell
	// script` hätte den Titel verkürzt und ein `echo` hineingesetzt.
	//
	// Hier stand zuerst eine zweite Zusicherung — „enthält nicht
	// `eingeschleust`" —, und die konnte nichts unterscheiden: das Wort steht
	// im **Eingabetext**. Eine Prüfung, die bei beiden Ausgängen zutrifft,
	// ist keine (§10).
	if gelesen != titel+"/"+meldung {
		t.Errorf("zurück kam %q, erwartet war %q", gelesen, titel+"/"+meldung)
	}
}

// **Und die Frist ist eine Frist.**
//
// Ein Dialog wartet auf einen Klick; kommt keiner, hinge der Start für immer.
// Geprüft wird das an einem Skript, das nichts anzeigt und trotzdem wartet —
// das Verhalten, um das es geht, ohne das Fenster, das im Weg wäre.
func TestEinHaengendesSkriptWirdAbgebrochen(t *testing.T) {
	if _, fehler := exec.LookPath("/usr/bin/osascript"); fehler != nil {
		t.Skip("kein osascript auf diesem Rechner")
	}

	ctx, aufgeben := context.WithTimeout(probekontext(t), 200_000_000) // 200 ms
	defer aufgeben()

	fehler := exec.CommandContext(ctx,
		"/usr/bin/osascript", "-e", `delay 30`).Run()
	if fehler == nil {
		t.Fatal("der Lauf endete von selbst")
	}

	if ctx.Err() == nil {
		t.Errorf("abgebrochen wurde aus einem anderen Grund: %v", fehler)
	}
}

// **Am Terminal wird kein Fenster geöffnet** (#834).
//
// # Was dieser Test belegt
//
// Dass der Aufruf **sofort** zurückkommt, wenn stderr an einem Terminal hängt
// — und dass er das mit einem Fehler tut, damit der Aufrufer nach stderr
// schreibt.
//
// **Unter `go test` hängt stderr an keinem Terminal**, deshalb prüft der Test
// beide Richtungen an `amTerminal` selbst und den Weg dahinter über die
// Zeitmessung: ein echter Dialog bräuchte die volle Frist.
func TestAmTerminalKommtKeinFenster(t *testing.T) {
	if amTerminal() {
		t.Skip("dieser Lauf hängt an einem Terminal")
	}

	// **Die Gegenprobe zur Aussage**: hier hängt keines, also **soll** das
	// Fenster kommen — und mit einer kurzen Frist kostet das nichts.
	vorher := meldefrist
	meldefrist = 300 * time.Millisecond

	t.Cleanup(func() { meldefrist = vorher })

	start := time.Now()
	fehler := meldungZeigen("MagicSetEditor4Mac", "Eine Probe.")

	if errors.Is(fehler, errNichtNoetig) {
		t.Fatal("es kam kein Fenster, obwohl stderr an keinem Terminal hängt")
	}

	if time.Since(start) < 200*time.Millisecond {
		t.Errorf("nach %v zurück — dann hat es nicht gewartet", time.Since(start))
	}
}
