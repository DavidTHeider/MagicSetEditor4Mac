package main

// **Abschluss und Protokoll an der Kommandozeile** — DOM-04.AK-04
// (#1699).

import (
	"strings"
	"testing"
)

func TestAbschlussUndProtokollAnDerKommandozeile(t *testing.T) {
	ort, _ := eingefahren(t)

	turnierID := strings.TrimSpace(mussLaufen(t, ort, "turnier", "neu", "Freitag", "magic"))
	mussLaufen(t, ort, "turnier", "anmelden", turnierID, "Anna")

	// **Ohne Runde schließt nichts ab** — ein fachliches Nein.
	if code, _, _ := fahreAn(t, ort, "turnier", "abschliessen", turnierID); code != CodeFachlich {
		t.Errorf("ohne Runde: %d, erwartet %d", code, CodeFachlich)
	}

	mussLaufen(t, ort, "turnier", "runde", turnierID)
	mussLaufen(t, ort, "turnier", "abschliessen", turnierID)

	// **Abgeschlossen nimmt nichts mehr an.**
	if code, _, _ := fahreAn(t, ort, "turnier", "anmelden", turnierID, "Bert"); code != CodeFachlich {
		t.Errorf("anmelden nach dem Abschluss: %d, erwartet %d", code, CodeFachlich)
	}

	if aus := mussLaufen(t, ort, "turnier", turnierID); !strings.Contains(aus, "abgeschlossen") {
		t.Errorf("der Kopf nennt den Abschluss nicht: %q", aus)
	}

	mussLaufen(t, ort, "turnier", "wiedereroeffnen", turnierID)
	mussLaufen(t, ort, "turnier", "anmelden", turnierID, "Bert")

	protokoll := mussLaufen(t, ort, "turnier", "protokoll", turnierID)
	for _, vorgang := range []string{"angelegt", "angemeldet", "runde", "abgeschlossen", "wiedereroeffnet"} {
		if !strings.Contains(protokoll, vorgang) {
			t.Errorf("das Protokoll nennt %q nicht: %q", vorgang, protokoll)
		}
	}

	if roh := mussLaufen(t, ort, "turnier", "protokoll", turnierID, "--json"); !strings.Contains(roh, `"eintraege"`) {
		t.Errorf("die JSON-Gestalt: %q", roh)
	}

	for _, befehl := range []string{"abschliessen", "wiedereroeffnen", "protokoll"} {
		if code, _, _ := fahreAn(t, ort, "turnier", befehl); code != CodeBedienung {
			t.Errorf("%s ohne Kennung: %d, erwartet %d", befehl, code, CodeBedienung)
		}
	}
}
