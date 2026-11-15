//go:build darwin

package app

import (
	"errors"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// **Der Weg ins Unified Log ist an der laufenden Maschine belegt** (#848).
//
// # Warum nicht an einer Attrappe
//
// Eine Attrappe zeigte, dass `Write` gerufen wird — nicht, dass der Satz
// ankommt. Genau dazwischen liegt die Frage: `os_log` ist eine
// C-Schnittstelle, cgo ist ausgeschlossen, und ob `/usr/bin/logger` das
// Unified Log **erreicht**, ist eine Aussage über macOS und nicht über
// unseren Quelltext.
//
// # Was er kostet
//
// Er liest über `log show` zurück und braucht dafür ein paar Sekunden. Das ist
// der Preis dafür, dass die Zusage keine Behauptung ist.
func TestDerSatzErreichtDasUnifiedLog(t *testing.T) {
	if _, fehler := exec.LookPath("/usr/bin/logger"); fehler != nil {
		t.Skip("kein logger auf diesem Rechner")
	}

	// **Eine Marke, die es sonst nirgends gibt.** Ohne sie fände der Lauf
	// irgendeinen Eintrag und wäre grün, ohne den eigenen gesehen zu haben.
	marke := "mse4mac-pruefmarke-" + strconv.FormatInt(time.Now().UnixNano(), 36)

	pfad := filepath.Join(t.TempDir(), "mse.log")

	schreiber, schliessen, fehler := Protokoll(pfad)
	if fehler != nil {
		t.Fatalf("Protokoll anlegen: %v", fehler)
	}

	schreiber.Info(marke)

	if fehler := schliessen(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	// **`log show` sieht einen Eintrag nicht sofort.** Der Puffer wird
	// geschrieben, wenn er voll ist oder die Zeit reicht; deshalb wird bis zu
	// einer Frist nachgesehen statt einmal.
	frist := time.Now().Add(20 * time.Second)

	for time.Now().Before(frist) {
		aus, fehler := exec.CommandContext(probekontext(t), "/usr/bin/log", "show",
			"--last", "2m", "--style", "compact").Output()
		if fehler == nil && strings.Contains(string(aus), marke) {
			return
		}

		time.Sleep(2 * time.Second)
	}

	t.Errorf("die Marke %q steht nach 20 s nicht im Unified Log", marke)
}

// **Eine leere Zeile wird nicht ins Unified Log geschickt** — und startet
// keinen `logger`.
//
// `slog` schreibt zeilenweise, und ein `MultiWriter` reicht auch den
// Zeilenumbruch allein weiter. Ohne diesen Zweig entstünde je Zeile ein
// zweiter, leerer Eintrag.
func TestEineLeereZeileGehtNichtInsSystemprotokoll(t *testing.T) {
	t.Parallel()

	kanal := neuerSystemkanal(loggerStarterMit(func(*exec.Cmd) error {
		t.Error("für eine leere Zeile gestartet")

		return nil
	}))

	geschrieben, fehler := kanal.Write([]byte("\n"))
	if fehler != nil || geschrieben != 1 {
		t.Errorf("%d Bytes, %v — erwartet 1 ohne Fehler", geschrieben, fehler)
	}

	if fehler := kanal.Close(); fehler != nil {
		t.Errorf("Schluss: %v", fehler)
	}
}

// **Ein fehlender `logger` meldet sich** — Befund 10 des Zweitblicks über
// Phase 6.4: hier wurde der Fehler verschluckt, und die Weiche nannte das
// Unified Log „an". Seit PERF03 kommt er mit der nächsten Zeile.
func TestEinFehlenderLoggerMeldetSich(t *testing.T) {
	t.Parallel()

	kanal := neuerSystemkanal(loggerStarterMit(func(*exec.Cmd) error { return exec.ErrNotFound }))

	t.Cleanup(func() { _ = kanal.Close() }) //nolint:errcheck // Aufräumen

	var gemeldet error

	for range 200 {
		if _, gemeldet = kanal.Write([]byte("eine Zeile\n")); gemeldet != nil {
			break
		}

		time.Sleep(5 * time.Millisecond)
	}

	if !errors.Is(gemeldet, exec.ErrNotFound) {
		t.Errorf("der fehlende logger blieb stumm: %v", gemeldet)
	}
}

// **Ein Logger, der stirbt, gibt seinen Tod an die nächste Zeile weiter** —
// Befund 6 des Zweitblicks über Phase 3 von 0.9.0: über einen `io.Pipe`
// kopierte `os/exec` in eine zweite Leitung, endete beim ersten Schreibfehler
// und ließ die erste offen; die nächste Zeile wartete auf einen Leser, den es
// nicht mehr gab. Gestorben wird hier echt: `/usr/bin/true` liest nichts.
func TestEinGestorbenerLoggerHaeltDieLeitungNichtAuf(t *testing.T) {
	t.Parallel()

	starter := loggerStarterMit(func(befehl *exec.Cmd) error {
		befehl.Path, befehl.Args = "/usr/bin/true", []string{"true"}

		return befehl.Start()
	})

	lauf, err := starter()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = lauf.ein.Close(); _ = lauf.warten() }) //nolint:errcheck // Aufräumen

	time.Sleep(200 * time.Millisecond) // er endet sofort

	geschrieben := make(chan error, 1)

	go func() {
		var fehler error
		for range 3 {
			if _, fehler = lauf.ein.Write([]byte("ins Leere\n")); fehler != nil {
				break
			}
		}
		geschrieben <- fehler
	}()

	select {
	case fehler := <-geschrieben:
		if fehler == nil {
			t.Error("drei Zeilen gingen an einen toten Logger, ohne dass es jemand merkte")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("das Schreiben hing an einem toten Logger")
	}
}
