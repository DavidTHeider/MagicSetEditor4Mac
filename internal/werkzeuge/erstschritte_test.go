// **Das Handbuch nennt die Schritte der Startseite in ihrer Reihenfolge** —
// #1170,
// Befund 6 des ersten Zweitblicks über die Freigabe.
//
// # Der Befund
//
// Das Handbuch erklärte „Vorlagen beziehen → Set öffnen → Karten bearbeiten"
// zur notwendigen Reihenfolge; die Startseite führt „Set öffnen → Karten
// bearbeiten → Vorlagen beziehen". Ein Neuling bekam von den beiden
// gegenteilige Anweisungen.
//
// **Die Startseite hatte recht.** Seit die Saat mit ins Bündel kam
// ([0037](../../docs/entscheidungen/0037-die-vorlagen-gehen-ins-buendel.md))
// bringt die Anwendung Spiel, Stilbögen und ein Beispielset mit; der erste
// Schritt ist eine Karte und keine Installation. Umgestellt war die
// Startseite, der Absatz nicht.
//
// # Was dieser Wächter hält
//
// Die drei Schlüssel aus `Start.tsx` stehen im Handbuch **in derselben
// Reihenfolge**, gemessen an der Stelle ihrer Überschrift.
//
// # Was er nicht hält
//
//   - **Ob die Erklärung daneben stimmt.** Er vergleicht Reihenfolgen, nicht
//     Sätze; ein Handbuch, das den richtigen Schritt falsch erklärt, kommt
//     durch. Das liest ein Mensch.
//   - **Ob die Startseite die richtige Reihenfolge hat.** Sie ist der Maßstab,
//     nicht der Gegenstand.
package werkzeuge

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// schrittImFenster findet `{ schluessel: 'setOeffnen', …`.
var schrittImFenster = regexp.MustCompile(`\{\s*schluessel:\s*'(\w+)',\s*ziel:`)

// Die Überschriften, unter denen ein Schritt im Handbuch steht.
//
// **Von Hand zugeordnet, und das ist die Grenze.** Ein Schlüssel ist ein
// Bezeichner, eine Überschrift ein Satz; sie automatisch zu verbinden hieße,
// den deutschen Text zu erraten. Wer einen Schritt umbenennt, trägt ihn hier
// ein — und ein Schlüssel ohne Eintrag macht diesen Lauf rot.
var ueberschriftZuSchritt = map[string]string{
	"setOeffnen":       "**Ein Set öffnen**",
	"kartenBearbeiten": "**Karten bearbeiten**",
	"paketeHolen":      "**Vorlagen beziehen**",
}

func TestDasHandbuchNenntDieErstenSchritteInIhrerReihenfolge(t *testing.T) {
	// **Seit dem 2026-09-03 stehen die Schritte in der Einführung**
	// (#1325),
	// nicht mehr auf der Startseite: sie war eine Hälfte davon, die
	// verschwand, sobald der Bestand ein Set führte — und weil „noch nicht
	// geladen" wie „kein Set" aussah, sprang die Seite um. Die Zusage dieses
	// Wächters ändert sich dabei nicht: das Handbuch nennt dieselben
	// Schritte in derselben Reihenfolge.
	const ort = "../../frontend/src/bausteine/Einfuehrung.tsx"

	quelle, err := os.ReadFile(ort) //nolint:gosec // ein fester Pfad
	if err != nil {
		t.Fatalf("%s lesen: %v", ort, err)
	}

	treffer := schrittImFenster.FindAllStringSubmatch(string(quelle), -1)
	if len(treffer) == 0 {
		t.Fatalf("in %s steht kein Schritt — der Leser greift daneben", ort)
	}

	handbuch, lesefehler := os.ReadFile("../../docs/HANDBUCH.md") //nolint:gosec // ein fester Pfad
	if lesefehler != nil {
		t.Fatalf("HANDBUCH.md lesen: %v", lesefehler)
	}

	text := string(handbuch)
	vorige := -1

	for stelle, eintrag := range treffer {
		ueberschrift, gefuehrt := ueberschriftZuSchritt[eintrag[1]]
		if !gefuehrt {
			t.Errorf("der Schritt %q ist in ueberschriftZuSchritt nicht geführt — "+
				"ohne Eintrag sagt dieser Lauf über ihn nichts", eintrag[1])

			continue
		}

		jetzt := strings.Index(text, ueberschrift)
		if jetzt < 0 {
			t.Errorf("das Handbuch nennt %s nicht", ueberschrift)

			continue
		}

		if jetzt < vorige {
			t.Errorf("das Handbuch nennt %s vor dem Schritt davor — die Startseite "+
				"führt ihn an %d. Stelle (docs/HANDBUCH.md, Abschnitt »Der erste Start«)",
				ueberschrift, stelle+1)
		}

		vorige = jetzt
	}
}
