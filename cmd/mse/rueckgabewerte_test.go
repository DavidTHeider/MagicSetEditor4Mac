package main

// **Die Rückgabewerte sind ein Vertrag** — F03
// (#2342),
// Fortschreibung von Vermerk 0033.
//
// Ein Skript verzweigt nach dem Rückgabewert, nicht nach dem Satz. Wer die
// **Bedeutung** eines Werts ändert, bricht es — das ist eine brechende
// Änderung nach 0033 (Regel 7) und braucht einen eigenen Übergangsvermerk.
// Dieser Wächter hält die Konstanten gegen die Tafel im Vermerk: eine
// Änderung an einer Seite allein macht ihn rot.
//
// **Was er nicht hält**: ob ein Befehl den richtigen Wert **wählt** — das
// prüfen die Läufe in `internal/flaechen` (`abweisungGleich` verlangt 1 für
// jede fachliche Abweisung).

import (
	"os"
	"regexp"
	"strconv"
	"testing"
)

func TestDieRueckgabewerteStehenImVertrag(t *testing.T) {
	roh, err := os.ReadFile("../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md")
	if err != nil {
		t.Fatal(err)
	}

	zeile := regexp.MustCompile("(?m)^\\| (\\d+) \\| `(Code[A-Za-z]+)` \\|")
	vertrag := map[string]int{}

	for _, treffer := range zeile.FindAllStringSubmatch(string(roh), -1) {
		wert, _ := strconv.Atoi(treffer[1]) //nolint:errcheck // das Muster lässt nur Ziffern durch
		vertrag[treffer[2]] = wert
	}

	gebaut := map[string]int{
		"CodeGut": CodeGut, "CodeFachlich": CodeFachlich, "CodeBedienung": CodeBedienung,
		"CodeAbgebrochen": CodeAbgebrochen, "CodeUngeklaert": CodeUngeklaert,
	}

	for name, wert := range gebaut {
		if vertragswert, da := vertrag[name]; !da || vertragswert != wert {
			t.Errorf("%s ist %d, der Vertrag (0033) sagt %v — eine Änderung der Bedeutung ist brechend", name, wert, vertrag[name])
		}
	}

	if len(vertrag) != len(gebaut) {
		t.Errorf("der Vertrag nennt %d Werte, gebaut sind %d", len(vertrag), len(gebaut))
	}
}
