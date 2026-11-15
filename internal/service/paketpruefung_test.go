// **Die Skripte einer Paketdatei prüfen, bevor sie gespeichert wird** —
// Werkstück B von Phase 4.9
// (#1324).
//
// # Warum das eine eigene Frage ist
//
// Die Vorschau zeichnet eine Karte und meldet, was dabei schiefging — aber
// ihr Fehler kann aus **jeder** beteiligten Datei stammen: dem Spielskript,
// einem eingebundenen Paket, der bearbeiteten Datei. Eine Zeile daraus im
// offenen Editor anzustreichen wäre eine Behauptung, die niemand geprüft hat.
//
// Diese Prüfung liest **nur den Text, der im Feld steht**. Ihre Stelle
// bezeichnet damit sicher eine Zeile dieser Datei.
package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestPaketSkriptfehlerNenntSeineStelleInDerDatei(t *testing.T) {
	faelle := []struct {
		name   string
		datei  string
		zeile  int
		spalte int
	}{
		{
			name:  "kein Skript, kein Befund",
			datei: "mse version: 0.3.8\ngame: magic\n",
			zeile: 0, spalte: 0,
		},
		{
			name:  "ein Skript ohne Fehler",
			datei: "mse version: 0.3.8\nscript: 1 + 2\n",
			zeile: 0, spalte: 0,
		},
		// **Die Spalte zählt in der Datei, nicht im Ausdruck.** `script: `
		// ist acht Zeichen breit; die Klammer steht danach.
		{
			name:  "eine offene Klammer in einer Zeile",
			datei: "mse version: 0.3.8\nscript: f(1, 2\n",
			zeile: 2, spalte: 10,
		},
		// **Auch die Zeilen des Rumpfes tragen ihre Einrückung.** `\tf(1,`
		// ist die dritte Zeile der Datei; der Tabulator steht auf Spalte 1,
		// das `f` auf 2 und die offene Klammer auf 3. Gemeldet wird die
		// Klammer — hier stand einmal 2, und das war die Stelle des `f`.
		{
			name:  "ein mehrzeiliges Skript",
			datei: "mse version: 0.3.8\nscript:\n\tf(1,\n\t2\n",
			zeile: 3, spalte: 3,
		},
		{
			name:  "ein Fehler in der zweiten von zwei Zeilen",
			datei: "mse version: 0.3.8\nscript: 1 + 2\nstyle:\n\tscript: [1, 2\n",
			zeile: 4, spalte: 10,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			befund := service.PaketSkriptfehler(fall.datei)

			if fall.zeile == 0 {
				if befund != nil {
					t.Fatalf("Befund %v, erwartet war keiner", befund)
				}

				return
			}

			if befund == nil {
				t.Fatal("kein Befund, erwartet war einer")
			}

			if befund.Zeile != fall.zeile || befund.Spalte != fall.spalte {
				t.Errorf(
					"Stelle Zeile %d, Spalte %d, erwartet Zeile %d, Spalte %d (%s)",
					befund.Zeile, befund.Spalte, fall.zeile, fall.spalte, befund.Meldung,
				)
			}

			// **Der Schlüssel überlebt die Verschiebung.** Ohne ihn könnte das
			// Fenster den Satz nicht übersetzen (§9), und die Meldung stünde
			// im englischen Fenster deutsch da.
			if befund.Schluessel != msescript.SchluesselKlammerOffen {
				t.Errorf("Schlüssel %q, erwartet %q", befund.Schluessel, msescript.SchluesselKlammerOffen)
			}
		})
	}
}
