// Die Argumente des schreibenden Werkzeugs.
//
// **Eigene Datei**, weil `werkzeuge_test.go` sonst über die Zeilengrenze aus §2
// liefe — und weil hier eine eigene Frage steht: was ein Aufruf **bedeutet**,
// der etwas weglässt.
package mcp_test

import (
	"strings"
	"testing"
)

// **Ein Aufruf ohne »wert« und ohne »entfernen« leert nicht das Feld.**
//
// # Woher diese Prüfung kommt
//
// Aus dem Nachlauf zum Zweitblick über Phase 9. Seit `entfernen` dazukam, ist
// `wert` kein Pflichtfeld mehr — und ein Aufruf, der beides weglässt, kam als
// `Wert: ""`, `Entfernen: false` an und **leerte** das Feld. Ein Modell, das
// nach den Argumenten fragt und eines vergisst, löschte damit einen Wert.
func TestEinAufrufOhneWertUndOhneEntfernenAendertNichts(t *testing.T) {
	bestand := probenbestand()

	zeilen := sprich(t, bestand,
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":`+
			`{"name":"schreiben.feld","arguments":{"setID": 7, "karteID":7,"feld":"name"}}}`)

	if !strings.Contains(zeilen[0], `"isError":true`) {
		t.Errorf("Antwort: %s", zeilen[0])
	}

	if len(bestand.gesetzt) != 0 {
		t.Errorf("es wurde geschrieben: %v", bestand.gesetzt)
	}
}

// Ein **ausdrücklich** leerer Wert leert das Feld — das ist etwas anderes.
func TestEinAusdruecklichLeererWertLeertDasFeld(t *testing.T) {
	bestand := probenbestand()

	zeilen := sprich(t, bestand,
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":`+
			`{"name":"schreiben.feld","arguments":{"setID": 7, "karteID":7,"feld":"name","wert":""}}}`)

	if !strings.Contains(zeilen[0], `"isError":false`) {
		t.Errorf("Antwort: %s", zeilen[0])
	}

	if len(bestand.gesetzt) != 1 || bestand.gesetzt[0] != "name=" {
		t.Errorf("beim Bestand kam an: %v", bestand.gesetzt)
	}
}
