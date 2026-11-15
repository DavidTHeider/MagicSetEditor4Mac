package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Ein Block liest sein Stellungsargument als `input`**
// (#216).
//
// # Was hier vorher stand
//
// Im Kopf von `auswerter_aufruf.go` stand: „Ein Stellungsargument hat keinen
// Namen und ist damit **nicht erreichbar** — in den gesichteten Vorlagen wird
// auch keines gelesen." Der zweite Halbsatz war die Begründung, und er ist
// widerlegt: vier der neunzehn Exportvorlagen von `M15-Magic-Pack` lesen
// `input` in einem Block, und `magic.mse-game` tut es ebenfalls.
//
// Das ist §10 in einem Satz: die Erinnerung eines Werkzeugs ist kein
// Prüfergebnis. Die Annahme stimmte für die Vorlagen, die damals angesehen
// wurden, und sie stimmte nicht für die Menge.
func TestEinBlockLiestSeinStellungsargumentAlsInput(t *testing.T) {
	faelle := []struct{ name, quelle, erwartet string }{
		{"unmittelbar", `({ input + "!" })("hallo")`, "hallo!"},

		// `filter@(…)` in den Vorlagen: erst binden, dann anwenden.
		{"nach einer Bindung", `({ input + suffix })@(suffix: "?")("was")`, "was?"},

		// **Das Stellungsargument schlägt ein benanntes `input:`.**
		//
		// Hier stand bis zum 2026-08-20 das Gegenteil, mit der Begründung
		// „sonst verlöre `f(x, input: y)` sein ausdrücklich gesetztes
		// Argument". Das war eine Setzung, und das Orakel sagt das Gegenteil:
		// `({ input })("x", input: "y")` gibt bei MSE **`"x"`**.
		//
		// Der Grund ist seine Bauform: **MSE kennt gar kein
		// Stellungsargument.** Ein bloßes Argument wird beim Zerteilen zu
		// `input` (`src/script/parser.cpp`, Zeilen 881–893, Fassung f9d9356
		// vom 2021-07-06), und `f(x, input: y)` ist dort
		// `f(input: x, input: y)` — der **erste** gewinnt.
		//
		// Angestrichen vom Nachlauf zum Zweitblick über Phase K; die ganze
		// Reihenfolge steht in `TestWerSchlaegtWenBeimAufruf`.
		{"Stellung schlägt benannt", `({ input })("aus der Stelle", input: "benannt")`, "aus der Stelle"},

		// **Verschachtelt gilt das innere.** Ein Block im Block bekommt sein
		// eigenes `input`; das äußere bliebe sonst sichtbar und der innere
		// Ausdruck läse den falschen Wert.
		{"verschachtelt", `({ ({ input + "b" })("a") })("z")`, "ab"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("%q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **Ohne Stellungsargument bleibt `input` unbekannt.**
//
// Die Gegenprobe: wer es immer setzte, machte aus einem Tippfehler einen
// leeren Text — und der sieht wie ein Ergebnis aus.
func TestOhneStellungsargumentBleibtInputUnbekannt(t *testing.T) {
	if _, err := msescript.AuswertenQuelle(t.Context(), `({ input })()`,
		msescript.Bibliothek()); err == nil {
		t.Error("kein Fehler — input war ohne Stellungsargument gesetzt")
	}
}
