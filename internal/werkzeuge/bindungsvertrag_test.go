package werkzeuge_test

// **Das Fenster hat den Vertrag quittiert, den Go anbietet** — ENG-06.AK-04
// (#1731).
//
// # Wofür es ihn gibt
//
// Das Issue verlangt, dass Bindungen **nicht still gegen einen anderen
// Backendvertrag ausgeliefert** werden. Die von Wails erzeugten können es
// nicht — sie entstehen bei jedem Bau neu und liegen in keinem Commit. Die
// Fassade in `frontend/src/` ist von Hand geschrieben, und an ihr kann es:
// eine umbenannte JSON-Marke übersetzt `tsc` klaglos.
//
// Deshalb hält `frontend/src/bindungsvertrag.json` den Abdruck jeder
// gebundenen Methode fest, **wie das Fenster ihn zuletzt gesehen hat**. Ändert
// sich einer in Go, wird dieser Wächter rot und nennt die Methode — und wer
// die Datei nachführt, tut es im Verzeichnis des Fensters, sichtbar im Diff.
//
// # Was er nicht hält
//
// **Ob die Typen des Fensters danach stimmen.** Er erzwingt einen Blick, nicht
// den richtigen; den Weg selbst geht der Nahtlauf an der laufenden Anwendung.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// quittung ist die Gestalt von `frontend/src/bindungsvertrag.json`.
type quittung struct {
	Abdruck  string            `json:"abdruck"`
	Methoden map[string]string `json:"methoden"`
}

func TestDasFensterHatDenBindungsvertragQuittiert(t *testing.T) {
	ort := filepath.Join("..", "..", "frontend", "src", "bindungsvertrag.json")

	erwartet := quittung{Abdruck: app.Vertragsabdruck(), Methoden: app.Bindungsvertrag()}

	roh, fehler := os.ReadFile(ort) //nolint:gosec // ein fester Ort im Baum
	if fehler != nil {
		t.Fatalf("%v\n\n%s", fehler, vorlage(t, erwartet))
	}

	var gelesen quittung
	if fehler := json.Unmarshal(roh, &gelesen); fehler != nil {
		t.Fatalf("%s ist kein JSON: %v", ort, fehler)
	}

	geaendert := []string{}

	for name, abdruck := range erwartet.Methoden {
		if gelesen.Methoden[name] != abdruck {
			geaendert = append(geaendert, name)
		}
	}

	for name := range gelesen.Methoden {
		if _, gebunden := erwartet.Methoden[name]; !gebunden {
			geaendert = append(geaendert, name+" (nicht mehr gebunden)")
		}
	}

	slices.Sort(geaendert)

	if len(geaendert) != 0 || gelesen.Abdruck != erwartet.Abdruck {
		t.Errorf("der Vertrag hat sich geändert, und das Fenster hat es nicht "+
			"quittiert — %d Methoden: %v\n\nWer die Typen in frontend/src/ "+
			"angesehen hat, schreibt nach %s:\n\n%s",
			len(geaendert), geaendert, ort, vorlage(t, erwartet))
	}
}

// vorlage gibt die Datei, wie sie sein müsste — in der Form, die prettier
// stehen lässt.
func vorlage(t *testing.T, erwartet quittung) string {
	t.Helper()

	roh, fehler := json.MarshalIndent(erwartet, "", "  ")
	if fehler != nil {
		t.Fatal(fehler)
	}

	return string(roh) + "\n"
}
