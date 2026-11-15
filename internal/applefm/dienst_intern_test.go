package applefm

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// **Antwortet der Dienst nicht binnen der Frist, ist das ein Fehler** — mit
// einer kurzen Frist, damit der Lauf nicht zehn Sekunden wartet.
func TestDerStartHatEineFrist(t *testing.T) {
	t.Parallel()

	ordner, fehler := os.MkdirTemp("", "fm")
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	t.Cleanup(func() { _ = os.RemoveAll(ordner) }) //nolint:errcheck // nur zum Aufräumen

	dienst := NeuerDienst(ordner, func(context.Context, ...string) (Prozess, error) {
		return befehlsprozessAus(t), nil
	})
	dienst.frist = 100 * time.Millisecond

	if _, fehler := dienst.Klient(t.Context()); fehler == nil || !strings.Contains(fehler.Error(), "binnen") {
		t.Errorf("%v", fehler)
	}
}

// befehlsprozessAus startet ein harmloses Programm, das keinen Socket anlegt.
func befehlsprozessAus(t *testing.T) Prozess {
	t.Helper()

	prozess, fehler := starte(exec.CommandContext(t.Context(), "/bin/sleep", "30"))
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	return prozess
}

// **Was ein Absturz zurückließ, räumt der nächste Start ab** — AI-05.AK-07:
// ein Socket, dessen Prozess nicht mehr lebt, und das `fm serve` dahinter.
// Der Socket eines lebenden Prozesses bleibt.
func TestVerwaisteSocketsWerdenAufgeraeumt(t *testing.T) {
	t.Parallel()

	ordner, fehler := os.MkdirTemp("", "fm")
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	t.Cleanup(func() { _ = os.RemoveAll(ordner) }) //nolint:errcheck // nur zum Aufräumen

	unter := ordner + "/mse-fm"
	if fehler := os.Mkdir(unter, 0o700); fehler != nil {
		t.Fatalf("%v", fehler)
	}

	for _, name := range []string{"fm-11.sock", "fm-22.sock", "fremd.txt"} {
		if fehler := os.WriteFile(unter+"/"+name, nil, 0o600); fehler != nil {
			t.Fatalf("%v", fehler)
		}
	}

	var beendet []string

	dienst := NeuerDienst(ordner, func(context.Context, ...string) (Prozess, error) {
		return nil, os.ErrNotExist
	})
	dienst.lebt = func(kennung int) bool { return kennung == 22 }
	dienst.verwaisteBeenden = func(_ context.Context, pfad string) error {
		beendet = append(beendet, pfad)

		return nil
	}

	_, _ = dienst.Klient(t.Context()) //nolint:errcheck // der Start scheitert absichtlich

	if len(beendet) != 1 || !strings.HasSuffix(beendet[0], "fm-11.sock") {
		t.Errorf("beendet: %v", beendet)
	}

	for name, soll := range map[string]bool{"fm-11.sock": false, "fm-22.sock": true, "fremd.txt": true} {
		if _, fehler := os.Stat(unter + "/" + name); (fehler == nil) != soll {
			t.Errorf("%s: da %v, soll %v", name, fehler == nil, soll)
		}
	}
}

// **Die echten Prüfer**: der eigene Prozess lebt, ein erfundener nicht.
func TestDieEchtenPrueferDesAufraeumens(t *testing.T) {
	t.Parallel()

	if !prozessLebt(os.Getpid()) || prozessLebt(1<<30) {
		t.Error("prozessLebt irrt")
	}
}

// **Der Weg eines Befehls, mit harmlosen Programmen** — `fm` selbst steht
// hinter dem Riegel, und der Weg darunter soll trotzdem geprüft sein.
func TestDerExitcodeEinesBefehls(t *testing.T) {
	t.Parallel()

	if code, fehler := exitcodeVon(exec.CommandContext(t.Context(), "/usr/bin/true")); code != 0 || fehler != nil {
		t.Errorf("true: %d %v", code, fehler)
	}

	if code, fehler := exitcodeVon(exec.CommandContext(t.Context(), "/usr/bin/false")); code != 1 || fehler != nil {
		t.Errorf("false: %d %v", code, fehler)
	}

	if _, fehler := exitcodeVon(exec.CommandContext(t.Context(), "/gibt/es/nicht")); fehler == nil {
		t.Error("ein fehlendes Programm lief")
	}
}

// **Ein gestarteter Befehl wird beendet und abgewartet** — ein zweites
// Beenden meldet, dass er schon vorbei ist.
func TestEinBefehlStartetUndEndet(t *testing.T) {
	t.Parallel()

	prozess := befehlsprozessAus(t)

	if fehler := prozess.Beende(); fehler != nil {
		t.Errorf("beenden: %v", fehler)
	}

	if fehler := prozess.Beende(); fehler == nil {
		t.Error("ein zweites Beenden ging durch")
	}

	if _, fehler := starte(exec.CommandContext(t.Context(), "/gibt/es/nicht")); fehler == nil {
		t.Error("ein fehlendes Programm startete")
	}

	if _, fehler := Prozessstarter(t.Context(), "serve"); fehler == nil {
		t.Error("fm startete im Testlauf")
	}
}

// **Der Riegel greift im Testlauf und im Kindprozess eines Laufs** — und
// sonst nicht.
func TestDerRiegelGreiftAuchImKindprozess(t *testing.T) {
	t.Parallel()

	leer := func(string) string { return "" }
	gesetzt := func(name string) string {
		if name == Sperre {
			return "an"
		}

		return ""
	}

	if !gesperrtFuer("/pfad/applefm.test", leer) {
		t.Error("ein Testlauf ist nicht gesperrt")
	}

	if !gesperrtFuer("/pfad/mse-mcp", gesetzt) {
		t.Error("ein Kindprozess mit Sperre ist nicht gesperrt")
	}

	if gesperrtFuer("/Applications/MagicSetEditor4Mac.app/Contents/MacOS/MagicSetEditor4Mac", leer) {
		t.Error("das Programm selbst ist gesperrt")
	}
}
