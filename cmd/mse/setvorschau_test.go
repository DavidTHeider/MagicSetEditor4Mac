package main

// **Ein Set vorab ansehen von der Kommandozeile** — FMT-03.AK-01 (#1692).

import (
	"strings"
	"testing"
)

const probesetdatei = "../../testdaten/probe.mse-set"

func TestDieKommandozeileZeigtEineSetvorschau(t *testing.T) {
	code, aus, _ := fahre(t, "set-vorschau", probesetdatei)
	if code != CodeGut || !strings.Contains(aus, "Mein Probeset\t3 Karten\tprobe/probe-stil") ||
		!strings.Contains(aus, "fehlt\tMein Probeset braucht probe.mse-game") {
		t.Errorf("%d:\n%s", code, aus)
	}

	code, aus, _ = fahre(t, "--json", "set-vorschau", probesetdatei)
	if code != CodeGut || !strings.Contains(aus, `"name": "Mein Probeset"`) || !strings.Contains(aus, `"bedarf"`) {
		t.Errorf("json %d:\n%s", code, aus)
	}
}

// **Nach dem Einfahren sagt sie, dass die Datei schon im Bestand liegt** —
// und nennt das Set, das ein zweites Einfahren zurückgäbe.
func TestDieSetvorschauNenntDasSchonOffeneSet(t *testing.T) {
	ort := frischerBestand(t)

	if code, _, fehler := fahreAn(t, ort, "oeffnen", probesetdatei); code != CodeGut {
		t.Fatalf("oeffnen: %d %s", code, fehler)
	}

	code, aus, _ := fahreAn(t, ort, "set-vorschau", probesetdatei)
	if code != CodeGut || !strings.Contains(aus, "schon offen\t") {
		t.Errorf("%d:\n%s", code, aus)
	}
}

// **Dieselbe Antwort wie `oeffnen`, auch im Rückgabewert** — eine Datei, die
// kein Set ist, ist „die Antwort ist nein" (1) und nicht „der Aufruf stimmt
// nicht" (2); ein Skript, das erst nachsieht und dann einliest, bekommt
// zweimal dasselbe.
func TestDieSetvorschauWeistAbWasKeinSetIst(t *testing.T) {
	code, _, _ := fahre(t, "oeffnen", "../../testdaten/probe.mse-game")
	if code != CodeFachlich {
		t.Fatalf("oeffnen: %d", code)
	}

	if code, _, fehler := fahre(t, "set-vorschau", "../../testdaten/probe.mse-game"); code != CodeFachlich || fehler == "" {
		t.Errorf("kein Set: %d %q", code, fehler)
	}

	if code, _, _ := fahre(t, "set-vorschau"); code != CodeBedienung {
		t.Errorf("ohne Pfad: %d", code)
	}
}
