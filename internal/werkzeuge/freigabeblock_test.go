// **Der Freigabeblock steht zweimal und muss zweimal gleich lauten** — auf
// Anweisung des Auftraggebers vom 2026-09-07.
//
// # Warum er zweimal steht
//
// `docs/RELEASE.md` hält ihn **zum Übernehmen**: wer eine Release-Notiz
// schreibt, kopiert ihn. `README.md` hält ihn, weil dort steht, was diese
// Anwendung ist — und „was diese Fassung nicht belegt" gehört zu jeder
// Fassung, nicht nur zu ihrer Ankündigung.
//
// **Ein Verweis genügte nicht**, und das ist bei der Release-Notiz schon
// entschieden (#874):
// wer die Notiz schreibt, müsste ihn aufschlagen, und wer sie liest, sähe ihn
// nie.
//
// # Warum es dafür einen Wächter braucht
//
// **Zwei Fassungen desselben Absatzes sind dieselbe Falle wie zwei Fassungen
// einer Leitlinie.** `AGENTS.md` und `CLAUDE.md` stehen aus demselben Grund
// unter einer Prüfung (§1): „Werkzeuge lesen das eine oder das andere, und
// keines soll dabei eine andere Regel bekommen."
//
// Hier ist es schlimmer: der Block trägt **Zahlen**. Ein Kanonlauf, der eine
// davon ändert, ändert sie an einer Stelle — und die andere behauptet danach
// etwas, das nicht gemessen wurde.
//
// # Was er hält und was nicht
//
//   - **Er hält**: beide Fassungen sind Zeichen für Zeichen gleich, nach dem
//     Abziehen der Zitatzeichen und des Leerraums am Zeilenende.
//   - **Er hält nicht**, ob die Zahlen darin stimmen. Das tut
//     `freigabezahlen_test.go` gegen den Kanonbeleg — seit dem Abschluss von
//     0.6.0; bis dahin stand der Block mit den Zahlen von 0.5.0 da.
package werkzeuge_test

import (
	"strings"
	"testing"
)

// blockanfang ist die Überschrift, mit der der Block beginnt.
const blockanfang = "> ## Was diese Fassung nicht belegt"

// blockende ist seine letzte Zeile.
const blockende = "> Alles Weitere bezieht die Anwendung auf dem Rechner der Nutzenden."

func TestDerFreigabeblockLautetZweimalGleich(t *testing.T) {
	t.Parallel()

	ausRelease := blockAus(t, "docs/RELEASE.md")
	ausReadme := blockAus(t, "README.md")

	if ausRelease == ausReadme {
		return
	}

	// **Die erste abweichende Zeile wird genannt**, nicht nur „ungleich".
	// Ein Wächter, der einen Unterschied meldet und ihn nicht zeigt, schickt
	// jemanden auf die Suche.
	linksZeilen := strings.Split(ausRelease, "\n")
	rechtsZeilen := strings.Split(ausReadme, "\n")

	for i := 0; i < len(linksZeilen) && i < len(rechtsZeilen); i++ {
		if linksZeilen[i] != rechtsZeilen[i] {
			t.Fatalf("der Freigabeblock läuft auseinander, ab Zeile %d des "+
				"Blocks:\n  RELEASE.md: %q\n  README.md:  %q",
				i+1, linksZeilen[i], rechtsZeilen[i])
		}
	}

	t.Fatalf("der Freigabeblock ist verschieden lang: RELEASE.md %d Zeilen, "+
		"README.md %d", len(linksZeilen), len(rechtsZeilen))
}

// blockAus schneidet den Freigabeblock aus einer Datei.
func blockAus(t *testing.T, datei string) string {
	t.Helper()

	inhalt := lies(t, datei)

	anfang := strings.Index(inhalt, blockanfang)
	if anfang < 0 {
		t.Fatalf("%s enthält %q nicht — heißt die Überschrift noch so?",
			datei, blockanfang)
	}

	ende := strings.Index(inhalt[anfang:], blockende)
	if ende < 0 {
		t.Fatalf("%s enthält die Schlusszeile des Blocks nicht", datei)
	}

	roh := inhalt[anfang : anfang+ende+len(blockende)]

	// **Leerraum am Zeilenende fällt weg.** `prettier` räumt ihn in der einen
	// Datei und lässt ihn in der anderen stehen; ein Unterschied daran wäre
	// ein Befund über den Formatierer und keiner über den Text.
	zeilen := strings.Split(roh, "\n")

	geglaettet := make([]string, 0, len(zeilen))
	for _, zeile := range zeilen {
		geglaettet = append(geglaettet, strings.TrimRight(zeile, " \t"))
	}

	return strings.Join(geglaettet, "\n")
}
