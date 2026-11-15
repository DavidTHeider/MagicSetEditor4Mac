package main

// **Die Parameterhilfen an der Kommandozeile** — Werkstück D von Phase 4.11
// (#1342).
//
// **Die fachliche Zusage prüft `internal/flaechen`.** Hier stehen die
// Bedienfehler.

import (
	"strings"
	"testing"
)

// **Ohne Wortkennung ist es ein Bedienfehler.**
func TestParameterOhneWortkennungIstEinBedienfehler(t *testing.T) {
	ort, setID, _ := eingefahrenesSet(t)

	if code, _, _ := fahreAn(t, ort,
		"schluesselwort", setID, "parameter"); code != CodeBedienung {
		t.Errorf("Rückgabewert %d, erwartet %d", code, CodeBedienung)
	}
}

// **Eine Kennung, die keine ist, ebenso.**
func TestParameterMitUnbrauchbarerWortkennung(t *testing.T) {
	ort, setID, _ := eingefahrenesSet(t)

	if code, _, _ := fahreAn(t, ort,
		"schluesselwort", setID, "parameter", "keine Zahl"); code != CodeBedienung {
		t.Errorf("Rückgabewert %d, erwartet %d", code, CodeBedienung)
	}
}

// **Ein Schlüsselwort, das es nicht gibt, ist ein fachliches Nein.**
func TestParameterEinesFehlendenSchluesselworts(t *testing.T) {
	ort, setID, _ := eingefahrenesSet(t)

	if code, _, _ := fahreAn(t, ort,
		"schluesselwort", setID, "parameter", "999"); code != CodeFachlich {
		t.Error("das fehlende Schlüsselwort kam durch")
	}
}

// **Und zu einem eigenen stehen beide Menüs da.**
func TestDieParameterhilfenEinesEigenenSchluesselworts(t *testing.T) {
	ort, setID, _ := eingefahrenesSet(t)

	_, wortID, _ := fahreAn(t, ort, "schluesselwort", setID, "anlegen")

	code, _, fehler := fahreAn(t, ort, "schluesselwort", setID,
		"parameter", trimZeile(wortID), "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}
}

// trimZeile nimmt den Zeilenumbruch von einer Kennung.
func trimZeile(aus string) string {
	return strings.TrimSpace(aus)
}
