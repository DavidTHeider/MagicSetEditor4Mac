// Die Prüfungen über Dateien, die **kein Baum** sind.
//
// Von `zerteiler_test.go` abgetrennt, weil die Datei über die Zeilengrenze aus
// AGENTS.md §2 gewachsen war. Der Schnitt folgt der Frage: dort steht, wie ein
// Baum entsteht, hier, was mit einer Datei geschieht, die keiner ist
// ([0014](../../docs/entscheidungen/0014-eine-mse-datei-ist-nicht-notwendig-ein-baum.md)).
package msepaket_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Eine Wurzelzeile ohne Doppelpunkt ist Text, kein Fehler**
// ([0014](../../docs/entscheidungen/0014-eine-mse-datei-ist-nicht-notwendig-ein-baum.md)).
//
// # Warum das keine Nachgiebigkeit ist
//
// Eine MSE-Datei ist nicht notwendig ein Baum. Von 555 Textdateien in
// `M15-Magic-Pack` scheitern zwölf an dieser Regel, und **keine davon ist eine
// Hauptdatei**: es sind Skriptrümpfe, Unit-Tests, JavaScript und CSS. Ein
// Zerteiler, der `.karte {` abweist, weist CSS ab — und CSS war nie ein Baum.
//
// Das Prüfstück bildet beide Formen nach, die im Pack vorkommen: die geöffnete
// Klammer unter einer Zuweisung (`magic.mse-game/script`) und die freistehende
// Anweisung (`magic.mse-game/unit_tests`).
func TestEineWurzelzeileOhneDoppelpunktIstText(t *testing.T) {
	faelle := map[string]struct {
		datei     string
		rohzeilen []string
	}{
		"ein Skriptrumpf": {
			datei:     "rumpf",
			rohzeilen: []string{"{", "assert(crop_safe(\"\") == nil)"},
		},
		"CSS": {
			datei:     "style css",
			rohzeilen: []string{".karte {", "}"},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			inhalt, err := os.ReadFile(
				filepath.Join(testdaten, "skriptdatei.mse-include", fall.datei))
			if err != nil {
				t.Fatalf("lesen: %v", err)
			}

			baum, err := msepaket.Zerteile(fall.datei, inhalt)
			if err != nil {
				t.Fatalf("zerteilen: %v", err)
			}

			// **Die Bytes zuerst.** Was hier gelesen wird, geht beim
			// Zurückschreiben wieder hinaus — und das ist die Zusage aus #48,
			// die auch für eine Datei gilt, deren Baum niemandem etwas sagt.
			if geschrieben := msepaket.Schreibe(baum); !bytes.Equal(geschrieben, inhalt) {
				t.Errorf("der Rundlauf hat die Bytes verändert:\n%q\n%q",
					string(geschrieben), string(inhalt))
			}

			// Und die Zeilen, um die es geht, sind als **Text** angekommen.
			for _, gesucht := range fall.rohzeilen {
				if !alsFortsetzungDa(baum.Wurzel, gesucht) {
					t.Errorf("%q steht nicht als Fortsetzung im Baum", gesucht)
				}
			}
		})
	}
}

// alsFortsetzungDa sucht eine Rohzeile, die als Fortsetzung gelesen wurde.
func alsFortsetzungDa(knoten []*msepaket.Knoten, roh string) bool {
	for _, kind := range knoten {
		if kind.Zeilenart == msepaket.ArtFortsetzung && kind.Roh == roh {
			return true
		}

		if alsFortsetzungDa(kind.Kinder, roh) {
			return true
		}
	}

	return false
}
