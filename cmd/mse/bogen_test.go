package main

// **Der Druckbogen an der Kommandozeile** —
// #1291.
//
// # Der Befund, der ihn verlangt
//
// `docs/FAEHIGKEITEN.md` führte ihn so:
//
//	| Einen Druckbogen legen | `BogenLegen` | — | — | — | ✋ |
//
// Drei Striche und ein Haken: **nur am Fenster**. §16 sagt, was das Fenster
// kann, können Kommandozeile, MCP und REST — oder es steht begründet, warum
// nicht. Für einen Druckbogen steht dort nichts.
//
// # Warum er ohne Fenster überhaupt geht
//
// Weil der Dialog nur in der **Bindung** steckt: `Bindungen.BogenLegen` fragt
// nach einem Ort und ruft dann `App.BogenLegen`, das einen Pfad entgegennimmt.
// Der Weg dahinter — `service.Bogenleger` — kennt kein Fenster.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// **`bogen` schreibt eine PDF-Datei — und sie ist ein PDF.**
func TestBogenSchreibtEinPDF(t *testing.T) {
	ort, _ := eingefahren(t)

	ziel := filepath.Join(t.TempDir(), "bogen.pdf")

	code, ausgabe, fehler := fahreAn(t, ort, "bogen", "1", ziel)
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	roh, err := os.ReadFile(ziel) // #nosec G304 -- Testpfad aus t.TempDir()
	if err != nil {
		t.Fatalf("die Bogendatei fehlt: %v", err)
	}

	// **Ein PDF und keine Datei mit Inhalt.** Der Unterschied ist der ganze
	// Punkt — eine leere Datei sähe von außen genauso aus.
	if !strings.HasPrefix(string(roh), "%PDF-") {
		t.Errorf("die Datei beginnt nicht mit %%PDF-: %q", roh[:min(20, len(roh))])
	}

	if !strings.Contains(ausgabe, ziel) {
		t.Errorf("die Ausgabe nennt den Pfad nicht: %q", ausgabe)
	}
}

// **Ohne Ziel ist es ein Bedienfehler**, kein fachliches Nein.
func TestBogenOhneZielIstEinBedienfehler(t *testing.T) {
	ort, _ := eingefahren(t)

	if code, _, _ := fahreAn(t, ort, "bogen", "1"); code != CodeBedienung {
		t.Errorf("Rückgabewert %d, erwartet %d", code, CodeBedienung)
	}
}

// **Eine Kennung, die keine ist, ebenso.**
func TestBogenMitUnbrauchbarerKennung(t *testing.T) {
	ort, _ := eingefahren(t)

	ziel := filepath.Join(t.TempDir(), "bogen.pdf")

	if code, _, _ := fahreAn(t, ort, "bogen", "keine Zahl", ziel); code != CodeBedienung {
		t.Errorf("Rückgabewert %d, erwartet %d", code, CodeBedienung)
	}
}

// **Ein Set, das es nicht gibt, ist ein fachliches Nein.**
func TestBogenFuerEinFehlendesSet(t *testing.T) {
	ort, _ := eingefahren(t)

	ziel := filepath.Join(t.TempDir(), "bogen.pdf")

	if code, _, _ := fahreAn(t, ort, "bogen", "999", ziel); code != CodeFachlich {
		t.Errorf("Rückgabewert %d, erwartet %d", code, CodeFachlich)
	}
}
