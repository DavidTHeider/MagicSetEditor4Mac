package service_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine Güte, die von MSE abweicht, sagt es an der Ausgabe** — Werkstück Ω.4
// aus #1029:
// „Wo eine Güte etwas weglässt, sagt sie es — an der Ausgabe, nicht in der
// Dokumentation."
//
// # Warum dieser Test hier steht und nicht in `render`
//
// `Guete.Hinweis` gab den Satz von Anfang an her, und ein Wächter bestätigte,
// dass es ihn gibt. **Gerufen hat ihn niemand.** Der Zweitblick über Phase Ω
// hat es angestrichen: „Code ohne Aufrufer", und die Zusage im CHANGELOG stand
// ohne Deckung.
//
// Der Satz gehört an die Stelle, an der die Güte **gewählt** wird — also an
// die Ausfuhr, nicht in den Rasterer. Von dort erreicht er alle vier Flächen
// über denselben Befundkanal.
func TestEineAbweichendeGueteSagtEsAnDerAusgabe(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name   string
		guete  render.Guete
		format service.Bildformat
		sagt   bool
	}{
		{"fast", render.GueteSchnell, service.FormatPNG, true},
		{"high", render.GueteHoch, service.FormatPNG, true},
		{"ultra", render.GueteHoechst, service.FormatPNG, true},

		// **Die Vorgabe schweigt.** Ein Befund bei jedem gewöhnlichen Bild
		// wäre Lärm, und Lärm liest niemand.
		{"medium schweigt", render.GueteMittel, service.FormatPNG, false},

		// **Und SVG sagt etwas anderes**: dort zweigt der Weg vor dem Rasterer
		// ab, und die Güte bewirkt gar nichts. Das ist ein eigener Satz, kein
		// Schweigen — die Oberfläche bietet die Wahl auch neben SVG an.
		{"fast bei SVG", render.GueteSchnell, service.FormatSVG, true},

		// **Und `medium` schweigt auch bei SVG.** Dort wirkt keine Güte, und
		// wer die Vorgabe wählt, bekommt genau das, was er erwartet — es gibt
		// nichts anzusagen. Der Fall stand nicht im Lauf, und damit war
		// unbelegt, dass die Reihenfolge der beiden Prüfungen stimmt
		// (Nachlauf des Zweitblicks über Phase Ω).
		{"medium bei SVG schweigt", render.GueteMittel, service.FormatSVG, false},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			ergebnis, err := bildausfuhr(t, planDouble{}).Schreibe(t.Context(), 1,
				t.TempDir(), "karte"+fall.format.Endung(), fall.format, 0,
				domain.BildkonfliktUeberschreiben, fall.guete)
			if err != nil {
				t.Fatalf("schreiben: %v", err)
			}

			if gesagt := nenntDieGuete(ergebnis.Befunde); gesagt != fall.sagt {
				t.Errorf("sagt die Güte an: %v, erwartet %v — Befunde: %v",
					gesagt, fall.sagt, ergebnis.Befunde)
			}
		})
	}
}

// nenntDieGuete sagt, ob ein Befund von der Güte spricht.
func nenntDieGuete(befunde []domain.Befund) bool {
	for _, befund := range befunde {
		if strings.Contains(befund.Meldung, "Güte") {
			return true
		}
	}

	return false
}
