package app

// **Nach dem Aufstieg läuft der Dienst noch unter dem alten Etikett** — R02
// (#2334)
// zusammen mit M02 (#2346).
//
// Das Fenster zieht ihn beim Start auf das neue Etikett um. **Wandert zuerst
// die Kommandozeile**, ist er noch nicht umgezogen — und ein Wärter, der nur
// das neue Etikett fragt, fände keinen Dienst und wanderte unter einem
// Dienst aus 0.8.0, der weiterschreibt: genau A-11.

import (
	"context"
	"os/exec"
	"strings"
	"sync/atomic"
	"testing"
)

// launchctlMit lässt `launchctl print` für ein Etikett einen laufenden Dienst
// melden und für jedes andere „nicht geladen".
func launchctlMit(t *testing.T, laufend string) {
	t.Helper()

	vorher := launchctlDrucken
	launchctlDrucken = func(ctx context.Context, ziel string) ([]byte, error) {
		skript := "exit 113"
		if strings.HasSuffix(ziel, "/"+laufend) {
			skript = "printf 'x = {\\n\\tstate = running\\n\\tpid = 4711\\n}\\n'"
		}

		// #nosec G204 -- ein festes Skript dieses Laufs
		return exec.CommandContext(ctx, "/bin/sh", "-c", skript).Output()
	}

	t.Cleanup(func() { launchctlDrucken = vorher })
}

func TestDerWaerterFindetDenDienstUnterDemAltenEtikett(t *testing.T) { //nolint:paralleltest // Nähte sind Paketvariablen
	for _, laufend := range []string{altdienstlabel, dienstlabel} {
		t.Run(laufend, func(t *testing.T) {
			launchctlMit(t, laufend)

			var gefragt atomic.Int32

			vorher := halterFragen
			halterFragen = func(context.Context, string) (string, error) {
				if gefragt.Add(1) == 1 {
					return "p4711\nc mse-rest\n", nil
				}

				return "", nil
			}

			t.Cleanup(func() { halterFragen = vorher })

			var befehle [][]string

			waerter := dienstwaerter{befehl: func(_ context.Context, befehl []string) error {
				befehle = append(befehle, befehl)

				return nil
			}}
			if fehler := waerter.anhalten(t.Context(), "/ein/bestand.db"); fehler != nil {
				t.Fatal(fehler)
			}

			// **Neu gestartet wird, was läuft** — unter seinem Etikett.
			if len(befehle) != 1 || befehle[0][1] != "kickstart" || !strings.HasSuffix(befehle[0][3], "/"+laufend) {
				t.Errorf("Befehle %v", befehle)
			}
		})
	}
}

// **Läuft keiner, gibt es nichts anzuhalten** — und die Nummer, die die
// Vorprüfung vor einem Wechsel ausnimmt, ist die des Dienstes unter welchem
// Etikett auch immer.
func TestOhneLaufendenDienstIstNichtsAnzuhalten(t *testing.T) { //nolint:paralleltest // Nähte sind Paketvariablen
	launchctlMit(t, "kein.etikett")

	if laeufe, fehler := laufendeDienste(t.Context()); len(laeufe) != 0 || fehler != nil {
		t.Errorf("%v %v", laeufe, fehler)
	}

	launchctlMit(t, altdienstlabel)

	if pids, fehler := dienstPIDs(t.Context()); len(pids) != 1 || pids[0] != 4711 || fehler != nil {
		t.Errorf("der Dienst unter dem alten Etikett: %v %v", pids, fehler)
	}
}

// **Laufen beide, wird jeder gefragt** — Befund 8 des Zweitblicks über Phase
// 3 von 0.9.0: der Wärter nahm den ersten laufenden und übersah einen alten,
// der den Bestand hielt, während der neue einen anderen bediente.
func TestLaufenBeideDiensteWirdJederGefragt(t *testing.T) { //nolint:paralleltest // Nähte sind Paketvariablen
	vorher := launchctlDrucken
	launchctlDrucken = func(ctx context.Context, ziel string) ([]byte, error) {
		pid := "100"
		if strings.HasSuffix(ziel, "/"+altdienstlabel) {
			pid = "200"
		}

		// #nosec G204 -- ein festes Skript dieses Laufs
		return exec.CommandContext(ctx, "/bin/sh", "-c", "printf 'x = {\\n\\tstate = running\\n\\tpid = "+pid+"\\n}\\n'").Output()
	}

	t.Cleanup(func() { launchctlDrucken = vorher })

	var gefragt atomic.Int32

	vorherFragen := halterFragen
	halterFragen = func(context.Context, string) (string, error) {
		if gefragt.Add(1) <= 2 {
			return "p200\nc mse-rest\n", nil
		}

		return "", nil
	}

	t.Cleanup(func() { halterFragen = vorherFragen })

	var befehle [][]string

	waerter := dienstwaerter{befehl: func(_ context.Context, befehl []string) error {
		befehle = append(befehle, befehl)

		return nil
	}}
	if fehler := waerter.anhalten(t.Context(), "/ein/bestand.db"); fehler != nil {
		t.Fatal(fehler)
	}

	if len(befehle) != 1 || !strings.HasSuffix(befehle[0][3], "/"+altdienstlabel) {
		t.Errorf("Befehle %v — der alte Dienst hielt den Bestand und lief weiter", befehle)
	}

	// **Und die Vorprüfung nimmt beide aus** — sie sind der Dienst, kein
	// fremdes Programm.
	if halter, fehler := (bestandshalter{
		pfad: "/ein/bestand.db", dienst: dienstPIDs,
		fragen: func(context.Context, string) (string, error) { return "p100\ncmse-rest\np200\ncmse-rest\n", nil },
	}).Halter(t.Context()); fehler != nil || len(halter) != 0 {
		t.Errorf("Halter %v (%v)", halter, fehler)
	}
}

// nurDerNeue ist der Dienst unter dem neuen Etikett — keiner, wo `pid` null ist.
func nurDerNeue(pid int) []dienstprozess {
	if pid == 0 {
		return nil
	}

	return []dienstprozess{{etikett: dienstlabel, pid: pid}}
}
