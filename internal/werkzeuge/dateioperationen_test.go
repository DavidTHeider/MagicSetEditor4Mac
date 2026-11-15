package werkzeuge

// **Die Spezifikation der Dateioperationen hält gegen den Dienst** —
// FMT-03.AK-01 (#1692).
//
// # Was er hält
//
// Jeder Weg des Dienstes, den `docs/DATEIOPERATIONEN.md` nennt, ist ein
// eingetragener Weg; jeder Auftragsweg (`rest.Auftragswege`) steht dort und
// trägt die Marke **Auftrag** — und keine andere Zeile trägt sie. Und jeder
// Weg, der eine Datei annimmt oder in den Ausgang schreibt
// (`rest.Uebertragungswege`), steht im Abschnitt über die Übertragung.
//
// # Was er nicht hält
//
// Ob die Sätze stimmen: Rollen, Grenzen, Fehlercodes, Nebenwirkungen. Die
// Grenzen stehen mit dem Namen ihrer Konstante daneben, damit ein Leser sie
// findet; eine geänderte Zahl sieht dieser Wächter nicht.

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// dateiweg trifft einen Weg des Dienstes in Backticks — bis zum Leerzeichen,
// zum Fragezeichen oder zum Ende des Codes.
var dateiweg = regexp.MustCompile("`(GET|POST|PUT|PATCH|DELETE) (/[^`\\s?]*)")

func dateioperationszeilen(t *testing.T) []string {
	t.Helper()

	roh, err := os.ReadFile(filepath.Join(wurzel, "docs", "DATEIOPERATIONEN.md")) //nolint:gosec // eine Datei im Repositorium
	if err != nil {
		t.Fatalf("DATEIOPERATIONEN.md lesen: %v", err)
	}

	return strings.Split(string(roh), "\n")
}

func TestDieDateioperationenNennenNurWegeDesDienstes(t *testing.T) {
	dienst := restwegeAusDienst(t)
	gezaehlt := 0

	for _, zeile := range dateioperationszeilen(t) {
		for _, treffer := range dateiweg.FindAllStringSubmatch(zeile, -1) {
			gezaehlt++

			if weg := treffer[1] + " " + treffer[2]; !slices.Contains(dienst, weg) {
				t.Errorf("DATEIOPERATIONEN.md nennt %q — den Weg gibt es nicht", weg)
			}
		}
	}

	if gezaehlt == 0 {
		t.Fatal("in DATEIOPERATIONEN.md fand sich kein Weg — der Leser greift daneben")
	}
}

func TestJederAuftragswegStehtMitMarkeInDenDateioperationen(t *testing.T) {
	auftrag := map[string]bool{}

	for _, zeile := range dateioperationszeilen(t) {
		if !strings.Contains(zeile, "**Auftrag**") {
			continue
		}

		for _, treffer := range dateiweg.FindAllStringSubmatch(zeile, -1) {
			auftrag[treffer[1]+" "+treffer[2]] = true
		}
	}

	for _, weg := range rest.Auftragswege() {
		if !auftrag[weg] {
			t.Errorf("der Auftragsweg %q steht nicht mit der Marke **Auftrag** in DATEIOPERATIONEN.md", weg)
		}

		delete(auftrag, weg)
	}

	for weg := range auftrag {
		t.Errorf("DATEIOPERATIONEN.md führt %q als Auftrag — der Dienst nimmt ihn nicht so an", weg)
	}
}

func TestJederUebertragungswegStehtImAbschnittDerUebertragung(t *testing.T) {
	zeilen := dateioperationszeilen(t)
	abschnitt := map[string]bool{}
	drin := false

	for _, zeile := range zeilen {
		if strings.HasPrefix(zeile, "## ") {
			drin = strings.Contains(zeile, "Ohne Pfad")
		}

		for _, treffer := range dateiweg.FindAllStringSubmatch(zeile, -1) {
			if drin {
				abschnitt[treffer[1]+" "+treffer[2]] = true
			}
		}
	}

	if len(abschnitt) == 0 {
		t.Fatal("DATEIOPERATIONEN.md hat keinen Abschnitt „Ohne Pfad“ mit Wegen")
	}

	for _, weg := range rest.Uebertragungswege() {
		if !abschnitt[weg] {
			t.Errorf("der Übertragungsweg %q fehlt in DATEIOPERATIONEN.md, Abschnitt „Ohne Pfad“", weg)
		}
	}
}
