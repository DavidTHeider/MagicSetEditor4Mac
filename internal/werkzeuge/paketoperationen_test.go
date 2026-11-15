package werkzeuge_test

// **Jede Paketart kennt jede Operation — mit Zustand, Grund und Test** —
// FMT-01 (#1690):
// „Jede Zelle nennt implementiert, fehlend, fehlerhaft oder fachlich nicht
// anwendbar mit Begründung und ausführbarem Test. Ein vorhandener Handler
// allein genügt nicht."
//
// # Was er hält
//
//   - **Jede der sieben Arten** aus `msepaket.Arten` steht mit **jeder** der
//     elf Operationen genau einmal in der Matrix.
//   - **Ein Zustand aus vieren**, und keiner sonst.
//   - **Jede Zelle nennt einen Test, den es gibt** — auch »fehlend« und
//     »nicht anwendbar«: dort den Gegenfall, der rot wird, sobald sich der
//     Zustand ändert. Jeder genannte Test steht als `func Test…(` in einer
//     Testdatei dieses Repositoriums.
//   - **Alles außer »implementiert« hat einen Grund.**
//
// # Was er nicht hält
//
// **Ob der genannte Test die Operation wirklich prüft.** Er sieht, dass es
// ihn gibt — dass er an dieser Fläche das Richtige tut, sieht ein Mensch,
// und dafür stehen die Flächenläufe aus `internal/flaechen` in der Spalte.

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// paketoperationen sind die elf aus FMT-01.AK-01.
var paketoperationen = []string{
	"Erstellen", "Lesen", "Ändern", "Löschen", "Kopieren", "Speichern",
	"Speichern unter", "Prüfen", "Import", "Export", "Nutzung",
}

// paketzustaende sind die vier aus FMT-01.AK-02.
var paketzustaende = []string{"implementiert", "fehlend", "fehlerhaft", "nicht anwendbar"}

// alleVorlagen ist die Kurzform für die sechs Arten außer dem Set.
const alleVorlagen = "alle sechs Vorlagenarten"

// matrixzeile ist eine Zeile der Operationsmatrix.
type matrixzeile struct {
	arten     []string
	operation string
	zustand   string
	tests     []string
	grund     string
}

// **Jede Art, jede Operation, genau einmal — und jede Zelle gedeckt.**
func TestJedePaketartKenntJedeOperation(t *testing.T) {
	t.Parallel()

	zeilen := operationsmatrix(t)
	tests := alleTestnamen(t)
	gesehen := map[string]int{}

	for _, zeile := range zeilen {
		for _, art := range zeile.arten {
			gesehen[art+" × "+zeile.operation]++
		}

		pruefeMatrixzeile(t, zeile, tests)
	}

	for _, art := range msepaket.Arten() {
		for _, operation := range paketoperationen {
			schluessel := string(art) + " × " + operation

			if anzahl := gesehen[schluessel]; anzahl != 1 {
				t.Errorf("%s steht %d-mal in der Matrix, erwartet einmal",
					schluessel, anzahl)
			}
		}
	}
}

// pruefeMatrixzeile hält eine Zeile gegen die Regeln.
func pruefeMatrixzeile(t *testing.T, zeile matrixzeile, tests map[string]bool) {
	t.Helper()

	name := strings.Join(zeile.arten, ", ") + " × " + zeile.operation

	if !slices.Contains(paketoperationen, zeile.operation) {
		t.Errorf("%s: keine Operation aus FMT-01", name)
	}

	if !slices.Contains(paketzustaende, zeile.zustand) {
		t.Errorf("%s: Zustand %q ist keiner der vier", name, zeile.zustand)
	}

	// **Jede Zelle nennt einen Test, nicht nur die implementierten** — für
	// »fehlend« und »nicht anwendbar« den Gegenfall, der rot wird, sobald sich
	// der Zustand ändert. Hier stand bis zum Zweitblick über Phase 6.0
	// `zeile.zustand == "implementiert"`, und eine später eingebaute Kopie
	// hieße weiter »fehlend«.
	if len(zeile.tests) == 0 {
		t.Errorf("%s ist %s und nennt keinen Test — ein Zustand ohne ausführbaren Test ist "+
			"eine Behauptung (FMT-01.AK-02)", name, zeile.zustand)
	}

	if zeile.zustand != "implementiert" && zeile.grund == "" {
		t.Errorf("%s ist %s und nennt keinen Grund", name, zeile.zustand)
	}

	for _, test := range zeile.tests {
		if !tests[test] {
			t.Errorf("%s nennt %s — den Test gibt es nicht", name, test)
		}
	}
}

// operationsmatrix liest die Tabelle unter »Operationen je Paketart«.
func operationsmatrix(t *testing.T) []matrixzeile {
	t.Helper()

	roh, fehler := os.ReadFile(filepath.Join("..", "..", "docs", "FAEHIGKEITEN.md"))
	if fehler != nil {
		t.Fatalf("FAEHIGKEITEN.md lesen: %v", fehler)
	}

	_, abschnitt, gefunden := strings.Cut(string(roh), "\n## Operationen je Paketart\n")
	if !gefunden {
		t.Fatal("FAEHIGKEITEN.md hat keinen Abschnitt »Operationen je Paketart«")
	}

	abschnitt, _, _ = strings.Cut(abschnitt, "\n## ")

	var zeilen []matrixzeile

	for _, zeile := range strings.Split(abschnitt, "\n") {
		zellen := tabellenzellen(zeile)
		if len(zellen) != 9 || zellen[0] == "Paketart" ||
			strings.HasPrefix(zellen[0], "---") {
			continue
		}

		zeilen = append(zeilen, matrixzeile{
			arten:     artenAus(zellen[0]),
			operation: zellen[1],
			zustand:   strings.Trim(zellen[2], "*"),
			tests:     testnamenAus(zellen[7]),
			grund:     strings.TrimSpace(zellen[8]),
		})
	}

	if len(zeilen) < 20 {
		t.Fatalf("nur %d Zeilen gelesen — der Leser greift daneben", len(zeilen))
	}

	return zeilen
}

// tabellenzellen zerlegt eine Tabellenzeile in ihre Zellen.
//
// **An jedem Strich und nicht an » | «** — eine leere Zelle steht als `| |`
// da, und eine Trennung an » | « ließe sie verschwinden; beim ersten Lauf
// fielen so siebzehn von neunundzwanzig Zeilen heraus.
func tabellenzellen(zeile string) []string {
	if !strings.HasPrefix(zeile, "|") || !strings.HasSuffix(zeile, "|") {
		return nil
	}

	zellen := strings.Split(zeile[1:len(zeile)-1], "|")
	for i := range zellen {
		zellen[i] = strings.TrimSpace(zellen[i])
	}

	return zellen
}

// artenAus liest die Arten einer Zeile — die Kurzform eingeschlossen.
func artenAus(zelle string) []string {
	if strings.Contains(zelle, alleVorlagen) {
		vorlagen := make([]string, 0, len(msepaket.Arten())-1)

		for _, art := range msepaket.Arten()[1:] {
			vorlagen = append(vorlagen, string(art))
		}

		return vorlagen
	}

	var arten []string

	for _, teil := range strings.Split(zelle, ",") {
		arten = append(arten, strings.Trim(strings.TrimSpace(teil), "`"))
	}

	return arten
}

// testnamenmuster trifft einen Testnamen in Backticks.
var testnamenmuster = regexp.MustCompile("`(Test\\w+)`")

// testnamenAus liest die Testnamen einer Zelle.
func testnamenAus(zelle string) []string {
	treffer := testnamenmuster.FindAllStringSubmatch(zelle, -1)
	namen := make([]string, 0, len(treffer))

	for _, treffer := range treffer {
		namen = append(namen, treffer[1])
	}

	return namen
}

// testkopf trifft den Kopf einer Testfunktion.
var testkopf = regexp.MustCompile(`(?m)^func (Test\w+)\(`)

// alleTestnamen sammelt jede Testfunktion im Repositorium.
func alleTestnamen(t *testing.T) map[string]bool {
	t.Helper()

	namen := map[string]bool{}

	for _, wurzel := range []string{"internal", "cmd"} {
		fehler := filepath.WalkDir(filepath.Join("..", "..", wurzel),
			func(pfad string, eintrag os.DirEntry, fehler error) error {
				if fehler != nil || eintrag.IsDir() ||
					!strings.HasSuffix(pfad, "_test.go") {
					return fehler
				}

				roh, lesefehler := os.ReadFile(pfad) //nolint:gosec // feste Wurzeln im Baum
				if lesefehler != nil {
					return lesefehler
				}

				for _, treffer := range testkopf.FindAllStringSubmatch(string(roh), -1) {
					namen[treffer[1]] = true
				}

				return nil
			})
		if fehler != nil {
			t.Fatalf("%s durchgehen: %v", wurzel, fehler)
		}
	}

	return namen
}
