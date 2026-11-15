package app_test

// **Eine leere Ausgabe einer Exportvorlage** —
// #1347.
//
// **Eine eigene Datei**, damit `vorlagenausfuhr_test.go` unter der
// Zeilengrenze aus §2 bleibt; die Vorrichtung (`mitVorlagen`,
// `vorlagenprobe`) steht dort.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Eine Vorlage, die nichts schreibt, schreibt eine leere Datei — und sagt es.**
//
// # Was MSE2 tut, nachgelesen
//
// `src/gui/html_export_window.cpp:81–87` (Fassung `f9d9356` vom 2021-07-06)
// schreibt `result->toString()` **ohne jede Prüfung**. Eine Vorlage, die
// nichts schreibt, erzeugt dort eine Null-Byte-Datei und meldet Erfolg.
//
// **§15 sagt: Semantik konservativ.** Die Datei entsteht also auch hier — was
// dazukommt, ist der Befund, denn MSE2 schweigt, und eine leere Datei sieht
// wie ein Fehlschlag aus
// (#1347).
func TestEineLeereAusgabeWirdGeschriebenUndGesagt(t *testing.T) {
	probe := &vorlagenprobe{
		vorlagen: []msepaket.Exportvorlage{probevorlage()},
		ausgabe:  "",
	}

	anwendung := mitVorlagen(t, probe)
	ziel := filepath.Join(t.TempDir(), "leer.txt")

	antwort := anwendung.VorlageAusfuehren(t.Context(), "1",
		"probe.mse-export-template", ziel, nil, nil)
	if antwort.Fehler != nil {
		t.Fatalf("%v", antwort.Fehler)
	}

	// **Die Datei ist da** — wie bei MSE2.
	inhalt, err := os.ReadFile(ziel) // #nosec G304 -- t.TempDir()
	if err != nil {
		t.Fatalf("die Datei fehlt: %v", err)
	}

	if len(inhalt) != 0 {
		t.Errorf("die Datei ist nicht leer: %q", inhalt)
	}

	// **Und der Befund steht daneben.**
	if len(antwort.Befunde) != 1 ||
		!strings.Contains(antwort.Befunde[0], "nichts geschrieben") {
		t.Errorf("die Befunde lauten %v", antwort.Befunde)
	}
}

// **Eine Ausgabe mit Inhalt bekommt keinen Befund.**
//
// Die Gegenprobe: ohne sie wäre ein Befund, der an **jeder** Ausfuhr hinge,
// grün — und die Meldung stünde bei jedem geglückten Lauf.
func TestEineAusgabeMitInhaltBekommtKeinenBefund(t *testing.T) {
	probe := &vorlagenprobe{
		vorlagen: []msepaket.Exportvorlage{probevorlage()},
		ausgabe:  "etwas",
	}

	antwort := mitVorlagen(t, probe).VorlageAusfuehren(probekontext(t), "1",
		"probe.mse-export-template", filepath.Join(t.TempDir(), "voll.txt"), nil, nil)

	if len(antwort.Befunde) != 0 {
		t.Errorf("die Befunde lauten %v", antwort.Befunde)
	}
}
