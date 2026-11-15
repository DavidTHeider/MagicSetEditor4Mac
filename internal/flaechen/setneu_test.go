// **Ein neues Set über jede Fläche** —
// #891, seit
// C04 (#2229)
// auch über das Fenster: hier stand „ohne Fenster", weil die Wails-Bindung in
// keinem Go-Prozess lief. `fensterAus` ruft sie über `NeueBindungen`, mit
// Toren und Rufkontext (Zweitblick über Phase C, Befund 3).
package flaechen_test

import (
	"net/http"
	"strings"
	"testing"
)

// **Jede Fläche legt ein Set an, und jede legt ein eigenes.**
//
// # Was hier die Zusage ist
//
// Nicht „dieselbe Antwort" — die Kennung ist bei jedem Aufruf eine andere, und
// das ist Absicht: ein Set hat keinen Schlüssel außer seiner Kennung, und zwei
// Sets desselben Namens sind erlaubt.
//
// Die Zusage ist: **jede Fläche kann es, und danach steht das Set in der
// Liste** — mit Namen und Spiel, wie sie hineingegeben wurden. Der Wächter
// über `FAEHIGKEITEN.md` sieht, **dass** die Fähigkeit an vier Flächen steht;
// hier wird sie gefahren (§16).
//
// # Was er nicht hält
//
// **Dass die Antworten gleich aussehen** — geprüft ist die Liste danach,
// gelesen an der Kommandozeile.
func TestEinNeuesSetUeberJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	// **Drei Namen, damit sich die drei Sets unterscheiden lassen.**
	const (
		ausCLI     = "Aus der Kommandozeile"
		ausMCP     = "Aus dem Modell"
		ausREST    = "Aus der Rückschleife"
		ausFenster = "Aus dem Fenster"
	)

	if angelegt := fensterAus(t, umgebung).SetAnlegen(t.Context(), ausFenster, "probe", "probe-schlicht"); angelegt.Fehler != nil {
		t.Fatalf("SetAnlegen: %+v", angelegt.Fehler)
	}

	mussKommandozeile(t, umgebung, "set-neu", ausCLI, "probe", "probe-schlicht")

	text, istFehler := mcpRuf(t, umgebung, "schreiben.set.neu",
		map[string]any{"name": ausMCP, "spiel": "probe", "stil": "probe-schlicht"})
	if istFehler {
		t.Fatalf("schreiben.set.neu: %s", text)
	}

	an := starteDienst(t, umgebung)

	status, rumpf := an.ruf(t, http.MethodPost, "/sets/neu",
		`{"name":"`+ausREST+`","spiel":"probe","stil":"probe-schlicht"}`)
	if status != http.StatusCreated {
		t.Fatalf("POST /sets/neu: %d (%s)", status, rumpf)
	}

	// **Und jetzt stehen alle vier in der Liste** — über dieselbe Fläche
	// gelesen, damit der Vergleich nicht drei Gestalten hat.
	liste := mussKommandozeile(t, umgebung, "sets")

	for _, name := range []string{ausFenster, ausCLI, ausMCP, ausREST} {
		if !strings.Contains(liste, name) {
			t.Errorf("%q steht nicht in der Liste:\n%s", name, liste)
		}
	}

	// **Das Spiel steht dabei.** Ohne es wäre ein Set eine Zeile, die sich
	// nicht öffnen lässt — und der Lauf hätte drei leere Zeilen gezählt.
	if zahl := strings.Count(liste, "\tprobe\t"); zahl != 4 {
		t.Errorf("%d Sets mit dem Spiel »probe«, erwartet 4:\n%s", zahl, liste)
	}
}
