// Der Wächter über die Paketlisten in `README.md` und `docs/ARCHITEKTUR.md`.
//
// # Wofür
//
// §9 verlangt, dass die Architekturdokumente mit der Änderung fortgeschrieben
// werden. `internal/svg` entstand am 2026-08-22 und stand in **keinem** von
// beiden — aufgefallen ist es erst beim Durchgehen der Tabelle aus §9, und das
// ist die Sorte Lücke, die niemand sucht.
//
// **Beide Listen sind Prosa**, und Prosa hat keinen Wächter (§2). Diese hier
// aber schon: sie ist eine **Aufzählung von Verzeichnissen**, und die lässt
// sich zählen.
//
// # Was er nicht prüft
//
// **Ob die Beschreibung daneben stimmt.** Er sieht `internal/svg/`, nicht was
// dahintersteht — dafür gibt es keinen Maßstab außer dem Lesen.
//
// **Die Reihenfolge und die Schichtzuordnung.** Ob `service` über `storage`
// steht, entscheidet kein Test.
package werkzeuge

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// ohneEintragInDenListen sind die Pakete, die in keiner der beiden Listen
// stehen — **jedes mit seinem Grund**.
var ohneEintragInDenListen = map[string]string{
	// **`werkzeuge` ist kein Paket der Anwendung**, sondern die Sammlung der
	// Wächter über sie. Es enthält keine ausgelieferte Zeile.
	"werkzeuge": "die Wächter selbst — nichts davon wird ausgeliefert",

	// **`schichten` ebenso**: es hält die Importregel und ist sonst leer.
	"schichten": "der Wächter über die Schichtenregel",

	// **Und `flaechen` ebenso**: es baut `cmd/mse` und `cmd/mse-mcp` und
	// fährt denselben Arbeitsablauf über beide — die Abnahme von Phase T
	// (#886).
	// Ausgeliefert wird davon keine Zeile.
	"flaechen": "die Abnahme über beide Flächen ohne Fenster",
}

// paketzeile trifft `internal/<name>` — mit oder ohne Schrägstrich dahinter.
//
// **Beide Schreibweisen**, weil beide Dateien es anders halten: `README.md`
// zeichnet einen Verzeichnisbaum mit Schrägstrich, `ARCHITEKTUR.md` eine
// Tabelle mit `internal/svg` in Backticks. Ein Muster für eine der beiden
// hätte die andere für vollständig gehalten.
var paketzeile = regexp.MustCompile(`internal/([a-z]+)\b`)

func TestBeideArchitekturlistenFuehrenJedesPaket(t *testing.T) {
	vorhanden := paketverzeichnisse(t)
	if len(vorhanden) < 10 {
		t.Fatalf("%d Pakete unter internal/ gefunden — der Leser greift daneben: %v",
			len(vorhanden), vorhanden)
	}

	for _, datei := range []string{"../../README.md", "../../docs/ARCHITEKTUR.md"} {
		roh, err := os.ReadFile(datei) //nolint:gosec // ein fester Pfad im Repositorium
		if err != nil {
			t.Fatalf("%s lesen: %v", datei, err)
		}

		genannt := map[string]bool{}
		for _, treffer := range paketzeile.FindAllStringSubmatch(string(roh), -1) {
			genannt[treffer[1]] = true
		}

		var fehlend []string

		for _, paket := range vorhanden {
			if _, begruendet := ohneEintragInDenListen[paket]; begruendet {
				continue
			}

			if !genannt[paket] {
				fehlend = append(fehlend, paket)
			}
		}

		if len(fehlend) > 0 {
			t.Errorf("%s nennt diese Pakete nicht: %s", datei, strings.Join(fehlend, ", "))
		}
	}
}

// TestDieAusnahmelisteIstNichtVeraltet ist die Gegenprobe.
func TestDieAusnahmelisteIstNichtVeraltet(t *testing.T) {
	vorhanden := paketverzeichnisse(t)

	for paket := range ohneEintragInDenListen {
		if !enthaelt(vorhanden, paket) {
			t.Errorf("ohneEintragInDenListen führt %q — das Paket gibt es nicht mehr", paket)
		}
	}
}

// paketverzeichnisse gibt die Namen der Pakete unter `internal/`.
func paketverzeichnisse(t *testing.T) []string {
	t.Helper()

	eintraege, err := os.ReadDir("..")
	if err != nil {
		t.Fatalf("internal lesen: %v", err)
	}

	namen := make([]string, 0, len(eintraege))

	for _, eintrag := range eintraege {
		if eintrag.IsDir() {
			namen = append(namen, eintrag.Name())
		}
	}

	sort.Strings(namen)

	return namen
}
