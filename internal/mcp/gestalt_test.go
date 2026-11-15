// **Die zugesagte Gestalt der Antworten**
// (#985).
//
// # Warum es diesen Wächter gibt
//
// Bis zum 2026-08-23 gaben `lesen.sets`, `lesen.karten` und `lesen.suche` die
// Entitäten aus `internal/domain` **unmittelbar** heraus — mit ihren
// Go-Feldnamen. MCP ist nach
// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md),
// Regel 5 eine Fläche nach **außen**: dort steht ein Gegenüber, das nicht
// mitgeliefert wird.
//
// **Ein `Origin` in `domain.SetHead`, das jemand in `Quelle` umbenennt, änderte
// damit stillschweigend den Vertrag** — und kein Test hätte es gesehen.
//
// # Was er hält und was nicht
//
// Er hält die **Feldnamen**, nicht ihre Bedeutung. Dass `karten` die Zahl der
// Karten ist und nicht die der Felder, liest ein Mensch.
package mcp_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// quelltext liest eine Datei dieses Pakets.
func quelltext(t *testing.T, name string) (string, error) {
	t.Helper()

	roh, err := os.ReadFile(filepath.Join(".", name)) //nolint:gosec // fester Name
	if err != nil {
		return "", err
	}

	return string(roh), nil
}

// zugesagteFelder ist die Gestalt, auf die ein Gegenüber sich verlassen darf.
//
// **Sie ist die von 0.1.0**, Zeichen für Zeichen. Eine Umbenennung wäre nach
// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md)
// Regel 7 eine **brechende** Änderung und bräuchte eine Hauptfassung — und der
// Gewinn wäre allein, dass die Namen wie die der Kommandozeile aussähen. Das
// ist kein Grund, ein Gegenüber zu brechen.
//
// **Was sich geändert hat, ist die Herkunft**: die Namen stehen jetzt an einem
// DTO in diesem Paket und nicht mehr an einer Entität, die sich mit der
// Fachlichkeit bewegt.
var zugesagteFelder = map[string][]string{
	"lesen.sets":   {"Cards", "Game", "ID", "Name", "Origin", "Style"},
	"lesen.karten": {"ID", "Notes", "Number", "SetID", "Values"},
}

func TestDieAntwortenHaltenIhreGestalt(t *testing.T) {
	bestand := probenbestand()

	for name, erwartet := range zugesagteFelder {
		t.Run(name, func(t *testing.T) {
			werkzeug := werkzeugMitNamen(t, bestand, name)

			roh, err := werkzeug.Ruf(t.Context(), json.RawMessage(`{"setID": 1}`))
			if err != nil {
				t.Fatalf("rufen: %v", err)
			}

			var zeilen []map[string]any
			if fehler := json.Unmarshal([]byte(roh), &zeilen); fehler != nil {
				t.Fatalf("kein JSON: %v\n%s", fehler, roh)
			}

			if len(zeilen) == 0 {
				t.Fatalf("die Antwort ist leer — die Vorrichtung gibt nichts her:\n%s", roh)
			}

			gesehen := make([]string, 0, len(zeilen[0]))
			for feld := range zeilen[0] {
				gesehen = append(gesehen, feld)
			}

			sort.Strings(gesehen)

			if strings.Join(gesehen, ",") != strings.Join(erwartet, ",") {
				t.Errorf("die Felder sind [%s], zugesagt sind [%s]",
					strings.Join(gesehen, ","), strings.Join(erwartet, ","))
			}
		})
	}
}

// **Und keine Antwort trägt eine Entität aus `internal/domain`.**
//
// Der Wächter darüber sieht die Namen; dieser sieht die **Herkunft**. Ohne ihn
// bliebe ein Werkzeug, das eine Entität durchreicht, unauffällig, solange
// deren Felder zufällig passen — und der nächste Umbau bräche den Vertrag.
func TestKeinWerkzeugGibtEineEntitaetHeraus(t *testing.T) {
	roh, err := quelltext(t, "bestand.go")
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	for _, verboten := range []string{
		"alsJSON(liste(koepfe))", "alsJSON(liste(karten))", "alsJSON(ergebnis)",
	} {
		if strings.Contains(roh, verboten) {
			t.Errorf("%q gibt eine Entität unmittelbar heraus — dazwischen "+
				"gehört ein DTO dieses Pakets (#985)", verboten)
		}
	}
}
