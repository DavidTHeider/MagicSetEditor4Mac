// **Ein Befund beim Zeichnen erreicht jede Fläche** —
// #1137.
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

// **Wer über die Kommandozeile, MCP oder REST zeichnet, erfährt, was fehlte.**
//
// # Der Befund
//
// `app.Zeichenwerk.SchreibeBild` gab Pfad, Breite und Höhe — und **keinen
// einzigen Befund**. Eine fehlende Schrift, ein unlesbares Bild, eine
// Verknüpfung, die dieses Programm nicht kann: im Fenster steht das alles
// neben der Karte. Über die drei anderen Flächen entstand dieselbe Datei, und
// niemand erfuhr, was ihr fehlt.
//
// Angestrichen vom Nachlauf des Zweitblicks über Phase Ω, und zwar an der
// Güte — die sagt seither an der Ausgabe, worin sie von MSE abweicht, und kam
// nur im Fenster an. **Der Befund ist größer als die Güte**: es fällt jeder.
//
// # Was hier gefahren wird
//
// Die **Güte** ist der Befund, den dieser Lauf erzeugt: `fast` weicht
// ausdrücklich ab und sagt es. Sie ist der einzige Befund, der sich ohne
// fehlende Schriften und ohne fremde Pakete herstellen lässt — und sie geht
// denselben Weg wie jeder andere.
func TestEinZeichenbefundErreichtJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	ziel := t.TempDir()

	setID := setkennungAus(t,
		mussKommandozeile(t, umgebung, "oeffnen", probeset(t), "--json"))

	karteID := ersteKarte(t,
		mussKommandozeile(t, umgebung, "karten", setID, "--json"))

	t.Run("Kommandozeile", func(t *testing.T) {
		bild := filepath.Join(ziel, "cli.png")

		aus, fehler, code := kommandozeile(t, umgebung,
			"zeichnen", karteID, bild, "--guete=fast")
		if code != 0 {
			t.Fatalf("zeichnen: %d (%s)", code, fehler)
		}

		nenntDieGuete(t, "die Kommandozeile", aus+fehler)
	})

	t.Run("MCP", func(t *testing.T) {
		bild := filepath.Join(ziel, "mcp.png")

		text, istFehler := mcpRuf(t, umgebung, "schreiben.karte.zeichnen",
			map[string]any{
				"karteID": zahl(t, karteID), "ziel": bild, "guete": "fast",
			})
		if istFehler {
			t.Fatalf("zeichnen: %s", text)
		}

		nenntDieGuete(t, "MCP", text)
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, umgebung)
		bild := filepath.Join(ziel, "rest.png")

		status, rumpf := an.ruf(t, http.MethodPost, "/karten/"+karteID+"/bild",
			`{"ziel":`+alsJSONText(t, bild)+`,"guete":"fast"}`)
		if status != http.StatusCreated {
			t.Fatalf("POST: %d (%s)", status, rumpf)
		}

		nenntDieGuete(t, "REST", rumpf)
	})
}

// nenntDieGuete prüft, dass der Befund der Güte in der Antwort steht.
func nenntDieGuete(t *testing.T, wer, antwort string) {
	t.Helper()

	if !strings.Contains(antwort, "fast") {
		t.Errorf("%s führt den Befund nicht: %s", wer, antwort)
	}
}
