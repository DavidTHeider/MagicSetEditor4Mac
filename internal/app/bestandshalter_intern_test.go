package app

// **Wer den Bestand sonst offen hält** — M02
// (#2346):
// die Vorprüfung vor einem Wechsel nennt jedes andere Programm, ohne diesen
// Prozess und ohne den eingerichteten Dienst, den die neue Fassung selbst neu
// startet.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"testing"
)

func TestDieHalterOhneDichUndOhneDenDienst(t *testing.T) {
	t.Parallel()

	eigene := strconv.Itoa(os.Getpid())
	lsof := "p" + eigene + "\ncMagicSetEditor4Mac\np4711\ncmse-rest\np815\ncsqlite3\np999\ncmse-rest\n"

	halter, fehler := bestandshalter{
		pfad:   "bestand.db",
		fragen: func(context.Context, string) (string, error) { return lsof, nil },
		dienst: func(context.Context) ([]int, error) { return []int{999}, nil },
	}.Halter(t.Context())
	if fehler != nil {
		t.Fatal(fehler)
	}

	if !slices.Equal(halter, []string{"mse-rest (4711)", "sqlite3 (815)"}) {
		t.Errorf("Halter %v", halter)
	}

	if _, fehler := (bestandshalter{
		fragen: func(context.Context, string) (string, error) { return "", errors.New("lsof fehlt") },
		dienst: func(context.Context) ([]int, error) { return nil, nil },
	}).Halter(t.Context()); fehler == nil {
		t.Error("eine Frage, die scheitert, gilt als beantwortet")
	}
}

// **Echt gefragt**: eine Datei, die niemand offen hat, hat keinen Halter —
// `lsof` endet dann mit 1 —, und eine, die dieser Prozess offen hält, nennt
// ihn.
func TestLsofWirklichGefragt(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "bestand.db")
	if err := os.WriteFile(pfad, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	if ausgabe, fehler := lsofFragen(t.Context(), pfad); fehler != nil || ausgabe != "" {
		t.Errorf("niemand hält: %q, %v", ausgabe, fehler)
	}

	offen, err := os.Open(pfad) //nolint:gosec // die eigene Probedatei
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = offen.Close() }() //nolint:errcheck // gehalten

	if ausgabe, fehler := lsofFragen(t.Context(), pfad); fehler != nil || !slices.Contains(splitZeilen(ausgabe), "p"+strconv.Itoa(os.Getpid())) {
		t.Errorf("dieser Prozess hält: %q, %v", ausgabe, fehler)
	}

	// **Und ein Fehler von `lsof` ist kein „niemand"** — Zweitblick über
	// Phase 2, Befund 9: `lsof` endet auch bei einem Fehler mit 1, und dessen
	// Meldung steht auf der Fehlerausgabe.
	if ausgabe, fehler := lsofFragen(t.Context(), filepath.Join(t.TempDir(), "gibt-es-nicht.db")); fehler == nil {
		t.Errorf("eine Datei, die es nicht gibt: %q ohne Fehler", ausgabe)
	}

	abgebrochen, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	if _, fehler := lsofFragen(abgebrochen, pfad); fehler == nil {
		t.Error("ein abgebrochener Ruf gab keinen Fehler")
	}
}

// **Ein Nachweis, der scheitert, ist kein „läuft nicht"** — Nachlauf über
// Phase 2 von 0.9.0, Befund 1 (#2323): bis dahin hieß jeder Fehler von
// `launchctl print` und jede Ausgabe ohne Prozessnummer „kein Dienst", und
// gesichert und gewandert wurde ohne Nachweis. **Nicht geladen** sagt
// `launchctl` mit 113 (am 2026-10-14 gefragt); `print` ist nach seiner
// Handbuchseite keine Schnittstelle, ein Zustand „läuft" ohne lesbare
// Prozessnummer ist deshalb ein Fehler und kein Nein.
func TestDiePIDDesDienstes(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		ausgabe string
		pid     int
		fehler  bool
	}{
		"läuft":                      {"gui/501/x = {\n\tstate = running\n\tpid = 4711\n}", 4711, false},
		"geladen und ruhend":         {"gui/501/x = {\n\tstate = not running\n}", 0, false},
		"läuft, ohne lesbare Nummer": {"gui/501/x = {\n\tstate = running\n\tpid = ?\n}", 0, true},
		"eine Ausgabe ohne Zustand":  {"", 0, true},
	} {
		pid, fehler := pidAus(fall.ausgabe)
		if pid != fall.pid || (fehler != nil) != fall.fehler {
			t.Errorf("%s: %d, %v — erwartet %d, Fehler %v", name, pid, fehler, fall.pid, fall.fehler)
		}
	}

	// **Und echt, lesend**: ein Etikett, das es nicht gibt, ist nicht
	// geladen — 0 und kein Fehler.
	if pid, fehler := launchctlPID(t.Context(), "de.magicseteditor4mac.probe.gibt.es.nicht"); pid != 0 || fehler != nil {
		t.Errorf("nicht geladen: %d, %v", pid, fehler)
	}

	// **Und ein Ruf, der nicht zu Ende kommt, ist ein Fehler.**
	abgebrochen, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	if _, fehler := launchctlPID(abgebrochen, dienstlabel); fehler == nil {
		t.Error("ein abgebrochener Ruf gab keinen Fehler")
	}
}
