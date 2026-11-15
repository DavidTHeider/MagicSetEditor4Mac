// **Die Güte über jede Fläche ohne Fenster** — Werkstück Ω.4 aus
// #1029,
// Vermerk [0049](../../docs/entscheidungen/0049-die-guete-ist-eine-wahl.md).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
package flaechen_test

import (
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

// gueten sind die vier Namen, unter denen eine Güte über die Flächen läuft.
//
// **Von Hand aufgezählt und nicht aus `render.Gueten` gelesen.** Dieser Lauf
// prüft, was **außen** ankommt; nähme er die Liste aus dem Programm, prüfte
// er das Programm gegen sich selbst — und ein umbenannter Wert bliebe grün.
var gueten = []string{"fast", "medium", "high", "ultra"}

// **Jede Fläche zeichnet in jeder Güte, und keine kennt eine fünfte.**
//
// # Was hier die Zusage ist
//
// Nicht „dasselbe Bild" — die Güten unterscheiden sich ja gerade, und das
// belegt `internal/app` an einer echten Karte. Die Zusage ist:
//
//  1. **jede der vier Güten kommt über jede Fläche an** und schreibt ein Bild,
//  2. **ein Name, den es nicht gibt, wird von jeder abgewiesen** — und zwar
//     fachlich, mit der Aufzählung dessen, was es gibt.
//
// Die zweite ist die wichtigere. Eine Fläche, die einen Tippfehler still auf
// die Vorgabe fallen lässt, sieht in Lauf 1 genauso grün aus.
func TestJedeGueteUeberJedeFlaecheOhneFenster(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	ziel := t.TempDir()

	// **Ein eingefahrenes Probeset und keine leere Karte**: gezeichnet werden
	// soll etwas, an dem sich Güten überhaupt unterscheiden können.
	setID := setkennungAus(t,
		mussKommandozeile(t, umgebung, "oeffnen", probeset(t), "--json"))

	karteID := ersteKarte(t,
		mussKommandozeile(t, umgebung, "karten", setID, "--json"))

	an := starteDienst(t, umgebung)

	for _, guete := range gueten {
		t.Run("die Güte "+guete, func(t *testing.T) {
			ueberKommandozeile(t, umgebung, ziel, karteID, guete)
			ueberMCP(t, umgebung, ziel, karteID, guete)
			ueberREST(t, an, ziel, karteID, guete)
		})
	}

	t.Run("eine Güte, die es nicht gibt", func(t *testing.T) {
		abgewiesen(t, umgebung, an, ziel, karteID)
	})
}

// ueberKommandozeile zeichnet mit `--guete=`.
func ueberKommandozeile(t *testing.T, umgebung []string, ziel, karteID, guete string) {
	t.Helper()

	bild := filepath.Join(ziel, "cli-"+guete+".png")

	if _, fehler, code := kommandozeile(t, umgebung,
		"zeichnen", karteID, bild, "--guete="+guete); code != 0 {
		t.Fatalf("zeichnen --guete=%s: %d (%s)", guete, code, fehler)
	}

	istBild(t, bild)
}

// ueberMCP zeichnet über das Werkzeug.
func ueberMCP(t *testing.T, umgebung []string, ziel, karteID, guete string) {
	t.Helper()

	bild := filepath.Join(ziel, "mcp-"+guete+".png")

	text, istFehler := mcpRuf(t, umgebung, "schreiben.karte.zeichnen",
		map[string]any{"karteID": zahl(t, karteID), "ziel": bild, "guete": guete})
	if istFehler {
		t.Fatalf("schreiben.karte.zeichnen mit %q: %s", guete, text)
	}

	istBild(t, bild)
}

// ueberREST zeichnet über die Rückschleife.
func ueberREST(t *testing.T, an dienst, ziel, karteID, guete string) {
	t.Helper()

	bild := filepath.Join(ziel, "rest-"+guete+".png")

	status, rumpf := an.ruf(t, http.MethodPost, "/karten/"+karteID+"/bild",
		`{"ziel":`+alsJSONText(t, bild)+`,"guete":"`+guete+`"}`)
	if status != http.StatusCreated {
		t.Fatalf("POST /karten/%s/bild mit %q: %d (%s)", karteID, guete, status, rumpf)
	}

	istBild(t, bild)
}

// abgewiesen prüft, dass keine Fläche einen erfundenen Namen durchlässt.
//
// **Und dass die Meldung sagt, was es gibt.** „unbekannte Güte" allein
// schickt jeden in die Dokumentation; die vier Namen stehen im Fehler.
func abgewiesen(t *testing.T, umgebung []string, an dienst, ziel, karteID string) {
	t.Helper()

	const erfunden = "superb"

	_, fehler, code := kommandozeile(t, umgebung, "zeichnen", karteID,
		filepath.Join(ziel, "cli-erfunden.png"), "--guete="+erfunden)
	if code == 0 {
		t.Error("die Kommandozeile hat eine erfundene Güte angenommen")
	}

	nenntDieVier(t, "die Kommandozeile", fehler)

	text, istFehler := mcpRuf(t, umgebung, "schreiben.karte.zeichnen",
		map[string]any{
			"karteID": zahl(t, karteID),
			"ziel":    filepath.Join(ziel, "mcp-erfunden.png"),
			"guete":   erfunden,
		})
	if !istFehler {
		t.Error("MCP hat eine erfundene Güte angenommen")
	}

	nenntDieVier(t, "MCP", text)

	status, rumpf := an.ruf(t, http.MethodPost, "/karten/"+karteID+"/bild",
		`{"ziel":`+alsJSONText(t, filepath.Join(ziel, "rest-erfunden.png"))+
			`,"guete":"`+erfunden+`"}`)
	if status == http.StatusCreated {
		t.Error("REST hat eine erfundene Güte angenommen")
	}

	nenntDieVier(t, "REST", rumpf)
}

// nenntDieVier prüft, dass eine Meldung die vier Namen führt.
func nenntDieVier(t *testing.T, wer, meldung string) {
	t.Helper()

	for _, guete := range gueten {
		if !strings.Contains(meldung, guete) {
			t.Errorf("%s nennt %q nicht: %s", wer, guete, meldung)
		}
	}
}
