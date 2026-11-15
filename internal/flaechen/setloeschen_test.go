// **Ein Set löschen über jede Fläche ohne Fenster** —
// #1218.
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// **Jede Fläche löscht ein Set, und jede verlangt dafür den Namen.**
//
// # Was hier die Zusage ist
//
// Zweierlei, und beides wird **gefahren** und nicht abgezählt:
//
//  1. Ein falscher Name löscht **nichts** — an jeder Fläche. Das ist die
//     Bestätigung, und eine, die an einer Fläche fehlte, wäre die Fläche, über
//     die ein Versehen durchkommt.
//  2. Der richtige Name nimmt das Set aus der Liste.
//
// # Was er nicht hält
//
// **Das Fenster.** Er fährt Kommandozeile, MCP und REST, das Fenster nicht
// (gemessen in `FLAECHENPARITAET.md`). Was die Bindung tut, prüft
// `internal/app` gegen dieselbe Naht.
//
// **Und nicht, was auf der Platte liegt.** Die eingesetzten Bilder räumt
// `service.Setloeschung`; hier wird gegen einen Bestand ohne sie gefahren.
func TestEinSetLoeschenUeberJedeFlaecheOhneFenster(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	const (
		fuerCLI  = "Weg über die Kommandozeile"
		fuerMCP  = "Weg über das Modell"
		fuerREST = "Weg über die Rückschleife"
	)

	kennungen := map[string]string{}
	for _, name := range []string{fuerCLI, fuerMCP, fuerREST} {
		kennungen[name] = strings.TrimSpace(
			mussKommandozeile(t, umgebung, "set-neu", name, "probe", "probe-schlicht"))
	}

	an := starteDienst(t, umgebung)

	// **Zuerst der falsche Name an jeder Fläche.** Er steht vor dem richtigen,
	// weil danach das Set weg ist — und ein „gibt es nicht" sähe aus wie ein
	// „Name stimmt nicht" und wäre keiner.
	t.Run("ein falscher Name löscht nichts", func(t *testing.T) {
		if _, _, code := kommandozeile(t, umgebung,
			"set-loeschen", kennungen[fuerCLI], "Heißt anders"); code == 0 {
			t.Error("die Kommandozeile hat mit falschem Namen gelöscht")
		}

		text, istFehler := mcpRuf(t, umgebung, "schreiben.set.loeschen",
			map[string]any{"setID": alsZahl(t, kennungen[fuerMCP]), "name": "Heißt anders"})
		if !istFehler {
			t.Errorf("MCP hat mit falschem Namen gelöscht: %s", text)
		}

		status, rumpf := an.ruf(t, http.MethodDelete,
			"/sets/"+kennungen[fuerREST]+"?name=Heißt+anders", "")
		if status == http.StatusOK {
			t.Errorf("REST hat mit falschem Namen gelöscht: %s", rumpf)
		}

		// **Und alle drei stehen noch da.** Ohne diese Zeile prüften die drei
		// Abweisungen nur, dass ein Fehler kam — nicht, dass nichts geschah.
		liste := mussKommandozeile(t, umgebung, "sets")
		for name := range kennungen {
			if !strings.Contains(liste, name) {
				t.Errorf("%q ist trotz falschem Namen weg:\n%s", name, liste)
			}
		}
	})

	t.Run("der richtige Name nimmt das Set", func(t *testing.T) {
		mussKommandozeile(t, umgebung, "set-loeschen", kennungen[fuerCLI], fuerCLI)

		text, istFehler := mcpRuf(t, umgebung, "schreiben.set.loeschen",
			map[string]any{"setID": alsZahl(t, kennungen[fuerMCP]), "name": fuerMCP})
		if istFehler {
			t.Fatalf("schreiben.set.loeschen: %s", text)
		}

		status, rumpf := an.ruf(t, http.MethodDelete,
			"/sets/"+kennungen[fuerREST]+"?name="+wegtauglich(fuerREST), "")
		if status != http.StatusOK {
			t.Fatalf("DELETE /sets/%s: %d (%s)", kennungen[fuerREST], status, rumpf)
		}

		liste := mussKommandozeile(t, umgebung, "sets")
		for name := range kennungen {
			if strings.Contains(liste, name) {
				t.Errorf("%q steht noch in der Liste:\n%s", name, liste)
			}
		}
	})
}

// alsZahl liest eine Setkennung als Zahl — MCP nimmt sie so entgegen.
func alsZahl(t *testing.T, kennung string) json.Number {
	t.Helper()

	return json.Number(kennung)
}

// wegtauglich schreibt einen Namen so, dass er in eine Abfrage passt.
func wegtauglich(name string) string {
	return strings.ReplaceAll(name, " ", "+")
}
