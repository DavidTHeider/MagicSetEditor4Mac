// **Die Schema- und Fassungsprüfung läuft beim Lesen mit** —
// Befund 7 des Zweitblicks über Phase V, 2026-08-24.
//
// # Der Befund
//
// `mtgjson.Pruefe` war vollständig gebaut und vollständig getestet — und
// **kein Aufrufer** im Importweg. Ein `grep` über `internal/` und `cmd/` fand
// sie nur in `internal/app/mtgjson_rundlauf_test.go`, also in einem Test.
//
// Wer über das Fenster, die Kommandozeile oder MCP eine MTG-JSON-Datei öffnet,
// bekam damit **nie** einen Schemabefund und nie den Hinweis auf eine
// unbekannte Hauptfassung.
//
// # Die Klasse
//
// Dieselbe wie #852
// und #1013:
// ein Paket mit voller Abdeckung und ohne Aufrufer ist keine
// Anwendungsfunktion. Die Abdeckung sagt darüber **nichts** — sie zählt
// betretene Zeilen, und ein Test betritt sie.
package mtgjson_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
)

func TestDieSchemapruefungLaeuftBeimLesenMit(t *testing.T) {
	t.Parallel()

	t.Run("eine unbekannte Hauptfassung erscheint als Befund", func(t *testing.T) {
		t.Parallel()

		inhalt := `{"meta":{"date":"2027-01-01","version":"6.0.0+20270101"},
		            "data":{"code":"PRF","name":"Prüfsatz","cards":[
		              {"name":"Gut","layout":"normal","uuid":"x"}]}}`

		quelle, fehler := mtgjson.LiesInhalt([]byte(inhalt), "probe.json",
			mtgjson.ZuordnungMagic)
		if fehler != nil {
			t.Fatalf("LiesInhalt: %v", fehler)
		}

		if !nenntEiner(quelle.Befunde, "Hauptfassung 6") {
			t.Errorf("kein Befund über die Fassung: %+v", quelle.Befunde)
		}
	})

	t.Run("ein Schlüssel mit der falschen Art erscheint als Befund", func(t *testing.T) {
		t.Parallel()

		// **`isReserved` ist ein Wahrheitswert**, hier steht eine Zeichenkette.
		// Die Zuordnung übergeht den Schlüssel — sie kennt ihn nicht —, und
		// ohne die Schemaprüfung fiele das lautlos durch.
		inhalt := `{"data":{"code":"PRF","name":"Prüfsatz","cards":[
		              {"name":"Gut","layout":"normal","isReserved":"ja","uuid":"x"}]}}`

		quelle, fehler := mtgjson.LiesInhalt([]byte(inhalt), "probe.json",
			mtgjson.ZuordnungMagic)
		if fehler != nil {
			t.Fatalf("LiesInhalt: %v", fehler)
		}

		if !nenntEiner(quelle.Befunde, "isReserved") {
			t.Errorf("kein Befund über die falsche Art: %+v", quelle.Befunde)
		}
	})

	t.Run("eine gute Datei bleibt ohne Schemabefund", func(t *testing.T) {
		t.Parallel()

		// **Ohne diesen Fall wäre nur belegt, dass die Prüfung läuft** — nicht,
		// dass sie unterscheidet. Eine, die immer meldet, wäre nutzlos und
		// bestünde den Lauf darüber.
		inhalt := `{"meta":{"date":"2026-08-24","version":"5.2.2"},
		            "data":{"code":"PRF","name":"Prüfsatz","cards":[
		              {"name":"Gut","layout":"normal","isReserved":false,
		               "types":["Instant"],"uuid":"x"}]}}`

		quelle, fehler := mtgjson.LiesInhalt([]byte(inhalt), "probe.json",
			mtgjson.ZuordnungMagic)
		if fehler != nil {
			t.Fatalf("LiesInhalt: %v", fehler)
		}

		for _, befund := range quelle.Befunde {
			if strings.Contains(befund.String(), "Art") ||
				strings.Contains(befund.String(), "Hauptfassung") {
				t.Errorf("unerwarteter Schemabefund: %s", befund)
			}
		}
	})
}

// nenntEiner sagt, ob ein Befund den Text führt.
func nenntEiner(befunde []domain.Befund, text string) bool {
	for _, befund := range befunde {
		if strings.Contains(befund.String(), text) {
			return true
		}
	}

	return false
}
