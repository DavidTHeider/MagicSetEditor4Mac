package app

// **Das Warten auf den alten Dienst, unabhängig von der Last** — Nachlauf
// über Phase 2 von 0.9.0, Befund 5
// (#2323).
//
// Die Läufe mit echtem `lsof` erreichten die Warteschleife nur, wenn `lsof`
// schneller war als ihre Frist: unter der Last einer Vorprüfung (21 auf zehn
// Kernen) lief die Frist schon bei der ersten Frage ab, und weder das Warten
// noch das Aufgeben darin war betreten. Hier antwortet eine Folge.

import (
	"context"
	"errors"
	"os/exec"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// altePID ist der Prozess des alten Dienstes in diesen Läufen.
const altePID = 4711

// halterAntworten setzt Prozess und Halterfrage: [altePID] läuft, und die
// Frage nach den Haltern sagt die ersten `halten` Male, er halte den Bestand.
func halterAntworten(t *testing.T, halten int) *atomic.Int32 {
	t.Helper()

	gefragt := &atomic.Int32{}
	vorherProzess, vorherFrage := dienstprozesse, halterFragen
	dienstprozesse = func(context.Context) ([]dienstprozess, error) { return nurDerNeue(altePID), nil }
	halterFragen = func(context.Context, string) (string, error) {
		if int(gefragt.Add(1)) <= halten {
			return "p" + strconv.Itoa(altePID) + "\nc mse-rest\n", nil
		}

		return "", nil
	}

	t.Cleanup(func() { dienstprozesse, halterFragen = vorherProzess, vorherFrage })

	return gefragt
}

// **Er hält noch, dann nicht mehr**: der Wärter fragt nach jedem Takt neu und
// ist fertig, sobald der Bestand frei ist.
func TestDerWaerterWartetBisDerDienstLoslaesst(t *testing.T) { //nolint:paralleltest // Nähte sind Paketvariablen
	gefragt := halterAntworten(t, 3)

	waerter := dienstwaerter{befehl: func(context.Context, []string) error { return nil }}
	if fehler := waerter.anhalten(t.Context(), "/ein/bestand.db"); fehler != nil {
		t.Fatal(fehler)
	}

	if gefragt.Load() != 4 {
		t.Errorf("%d Mal gefragt, erwartet vier: einmal davor, zweimal im Warten, einmal frei", gefragt.Load())
	}
}

// **Hält er den Bestand nicht, bleibt er**: ein Dienst, der einen anderen
// Bestand bedient, schreibt nicht in diesen, und angehalten wird er nicht.
func TestEinDienstDerDenBestandNichtHaeltBleibt(t *testing.T) { //nolint:paralleltest // dito
	halterAntworten(t, 0)

	gerufen := false
	waerter := dienstwaerter{befehl: func(context.Context, []string) error { gerufen = true; return nil }}

	if fehler := waerter.anhalten(t.Context(), "/ein/bestand.db"); fehler != nil || gerufen {
		t.Errorf("angehalten %v, Fehler %v", gerufen, fehler)
	}
}

// **Scheitert die Frage nach den Haltern, ist das kein Nein**: der Bestand
// ist belegt, und angehalten wird nicht.
func TestEineGescheiterteHalterfrageBelegtDenBestand(t *testing.T) { //nolint:paralleltest // dito
	halterAntworten(t, 0)

	halterFragen = func(context.Context, string) (string, error) { return "", errors.New("lsof: keine Antwort") }

	gerufen := false
	waerter := dienstwaerter{befehl: func(context.Context, []string) error { gerufen = true; return nil }}

	if fehler := waerter.anhalten(t.Context(), "/ein/bestand.db"); !errors.Is(fehler, domain.ErrBelegt) || gerufen {
		t.Errorf("angehalten %v, Fehler %v", gerufen, fehler)
	}
}

// **Er hält, bis die Frist abläuft**: dann ist der Bestand belegt, und es
// wird weder gesichert noch gewandert.
func TestDerWaerterGibtNachDerFristAuf(t *testing.T) { //nolint:paralleltest // dito
	halterAntworten(t, 1<<30)

	vorher := dienstanhaltefrist
	dienstanhaltefrist = 300 * time.Millisecond

	t.Cleanup(func() { dienstanhaltefrist = vorher })

	waerter := dienstwaerter{befehl: func(context.Context, []string) error { return nil }}
	if fehler := waerter.anhalten(t.Context(), "/ein/bestand.db"); !errors.Is(fehler, domain.ErrBelegt) {
		t.Errorf("nach der Frist: %v", fehler)
	}
}

// **Was `launchctl print` über einen laufenden Dienst sagt**, liest
// [launchctlPID] — hier aus einem Skript, das so antwortet.
func TestLaunchctlPIDLiestEinenLaufendenDienst(t *testing.T) { //nolint:paralleltest // die Naht ist eine Paketvariable
	for name, fall := range map[string]struct {
		skript string
		pid    int
		fehler bool
	}{
		"er läuft":          {skript: "printf 'gui/501/x = {\\n\\tstate = running\\n\\tpid = 4711\\n}\\n'", pid: 4711},
		"nicht geladen":     {skript: "exit 113"},
		"launchd scheitert": {skript: "echo kaputt >&2; exit 5", fehler: true},
	} {
		vorher := launchctlDrucken
		launchctlDrucken = func(ctx context.Context, _ string) ([]byte, error) {
			// #nosec G204 -- ein festes Skript dieses Laufs
			return exec.CommandContext(ctx, "/bin/sh", "-c", fall.skript).Output()
		}
		pid, fehler := launchctlPID(t.Context(), "probe")
		launchctlDrucken = vorher

		var beendet *exec.ExitError
		if pid != fall.pid || (fehler != nil) != fall.fehler || (fehler != nil && !errors.As(fehler, &beendet)) {
			t.Errorf("%s: %d, %v", name, pid, fehler)
		}
	}
}

// **Ein Fehler beim Fragen bricht ab** (Zweitblick über Phase 3, Befund 8):
// läuft der neue Dienst und scheitert das Fragen nach dem alten, ist das kein
// „nur der neue" — ein halber Nachweis ist kein Nein.
func TestLaufendeDiensteBrechenBeimFehlerAb(t *testing.T) { //nolint:paralleltest // die Naht ist eine Paketvariable
	vorher := launchctlDrucken

	t.Cleanup(func() { launchctlDrucken = vorher })

	launchctlDrucken = func(ctx context.Context, ziel string) ([]byte, error) {
		skript := "printf 'x = {\\n\\tstate = running\\n\\tpid = 4711\\n}\\n'"
		if strings.HasSuffix(ziel, "/"+altdienstlabel) {
			skript = "echo kaputt >&2; exit 5"
		}

		// #nosec G204 -- ein festes Skript dieses Laufs
		return exec.CommandContext(ctx, "/bin/sh", "-c", skript).Output()
	}

	if laeufe, fehler := laufendeDienste(t.Context()); fehler == nil {
		t.Errorf("das Fragen nach %s scheiterte, und es hieß %+v", altdienstlabel, laeufe)
	}
}
