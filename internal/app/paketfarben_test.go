// **Die Bindung für die Einfärbung einer Paketdatei** —
// #1181.
//
// **Eine zweite Bindung neben `Skriptfarben` und kein Parameter**: es sind
// zwei Sprachen. Ein Schalter darin wäre eine fachliche Verzweigung in
// `internal/app`, und dort gehört keine hin (§2).
package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestPaketfarbenGibtSchluesselUndWert(t *testing.T) {
	t.Parallel()

	antwort := app.NeueApp("").Paketfarben("game: magic\n")

	if len(antwort.Merkmale) != 2 {
		t.Fatalf("%d Merkmale, erwartet zwei: %+v", len(antwort.Merkmale), antwort.Merkmale)
	}

	if antwort.Merkmale[0].Art != "Schlüssel" || antwort.Merkmale[1].Art != "Wert" {
		t.Errorf("Arten %q und %q, erwartet Schlüssel und Wert",
			antwort.Merkmale[0].Art, antwort.Merkmale[1].Art)
	}
}

// **Eine unlesbare Datei färbt, so weit sie lesbar ist** — und meldet nichts.
func TestPaketfarbenMeldenKeinenFehler(t *testing.T) {
	t.Parallel()

	if antwort := app.NeueApp("").Paketfarben(": ohne Namen\n"); len(antwort.Merkmale) != 0 {
		t.Errorf("Merkmale %+v, erwartet keins", antwort.Merkmale)
	}
}

// **Und die Bindung reicht durch.**
//
// Sie ist ein Durchreicher; geprüft wird an einem **Ergebnis** und nicht am
// Ausbleiben eines Fehlers — dieselbe Regel wie in `bindungen_test.go`.
func TestDieBindungPaketfarbenReichtDurch(t *testing.T) {
	t.Parallel()

	antwort := app.NeueBindungen(app.NeueApp("")).Paketfarben(t.Context(), "game: magic\n")

	if len(antwort.Merkmale) != 2 {
		t.Errorf("%d Merkmale, erwartet zwei: %+v", len(antwort.Merkmale), antwort.Merkmale)
	}
}
