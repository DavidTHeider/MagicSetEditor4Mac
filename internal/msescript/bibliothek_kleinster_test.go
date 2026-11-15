package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// `min` — die drittmeistvermisste Funktion, ausgezählt statt geraten.
//
// # Herkunft und Form
//
// `src/script/parser.cpp:632` von twanvl/MagicSetEditor2 behandelt `min` und
// `max` gemeinsam und **vorangestellt mit Klammer**: `min(x,y,z,…)`, paarweise
// zusammengefaltet. Sie sind damit keine Infix-Zeichen, obwohl `context.cpp`
// sie als Anweisungen `I_MIN`/`I_MAX` führt.
//
// Nachgesehen, weil es anders aussah: `grep` fand in `M15-Magic-Pack` Zeilen
// wie `cost min 5`, was nach einem Infix-Zeichen aussieht. Es sind
// Ausschnitte aus `min(cost, 5)` über mehrere Zeilen — der Zerteiler von MSE
// kennt keine Infix-Form.
//
// # Warum sie zählt
//
// Gemessen am 2026-08-16 über die 190 Stile von `M15-Magic-Pack`: **841**
// Feldstile scheiterten an ihr, verteilt auf `height`, `left`, `top`, `width`
// und `font size` — also genau auf die Angaben, die eine Ebene erst
// platzierbar machen.
//
// Fassung f9d9356 vom 2021-07-06.
func TestMinGibtDenKleinsten(t *testing.T) {
	faelle := []struct {
		name    string
		quelle  string
		erwartt float64
	}{
		{"zwei Zahlen", `min(3, 7)`, 3},
		{"die andere Seite", `min(7, 3)`, 3},
		{"mehr als zwei", `min(10, 4, 8, 2, 6)`, 2},
		{"eine einzige", `min(5)`, 5},
		{"Kommazahlen", `min(1.5, 1.25)`, 1.25},

		// **`nil` rechnet als null** — dieselbe Umwandlung wie überall
		// (`umwandlung.go`). `min(0, move_typebox())` steht so im Pack, und
		// `move_typebox` gibt oft nichts.
		{"nichts ist null", `min(3, nil)`, 0},

		// So steht es in `magic-m15-planeswalker.mse-style`, verschachtelt.
		{"in max verschachtelt", `min(1.0, max(0.0, 0.5))`, 0.5},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.Art != msescript.ArtZahlwert {
				t.Fatalf("Art %s", wert.Art)
			}

			if wert.Zahl != fall.erwartt {
				t.Errorf("%v, erwartet war %v", wert.Zahl, fall.erwartt)
			}
		})
	}
}

// **Ohne Argument ist `min` ein Tippfehler.**
func TestMinOhneArgumentWirdAbgewiesen(t *testing.T) {
	_, err := msescript.AuswertenQuelle(t.Context(), `min()`, msescript.Bibliothek())
	if err == nil {
		t.Fatal("kein Fehler")
	}
}

// **Ein Text ohne Zahl bleibt auch hier ein Fehler.**
//
// Die Gegenprobe: `min` benutzt dieselbe Umwandlung wie `*` und `>`, und die
// wirft für alles, was jemand gemeint haben könnte.
func TestMinMitEinemUnlesbarenWertMeldet(t *testing.T) {
	_, err := msescript.AuswertenQuelle(t.Context(), `min(1, "links")`,
		msescript.Bibliothek())
	if err == nil {
		t.Fatal("kein Fehler")
	}
}

// **Eine eingebaute Funktion nimmt `input` aus der Umgebung, wenn sie kein
// Stellungsargument bekommt.**
//
// # Herkunft
//
// MSE liest in jeder eingebauten Funktion die Kontextvariable `input`, nicht
// ein Stellungsargument — `SCRIPT_PARAM_C(String,input)` in
// `src/script/functions/editor.cpp:240` für `chosen`, und ebenso überall sonst.
// Ein Stellungsargument ist bei MSE nur eine **Schreibweise**, die `input`
// setzt.
//
// # Was das gemessen ausmacht
//
// 172 Feldstile scheiterten mit „chosen erwartet ein Argument, bekam kein
// Argument", weitere 162 mit „der Name `input` ist hier nicht bekannt". In den
// Vorlagen steht `chosen@(choice: "white")` — teilweise angewandt, und das
// Stellungsargument kommt erst später aus der Umgebung.
//
// Fassung f9d9356 vom 2021-07-06.
func TestEineEingebauteFunktionNimmtInputAusDerUmgebung(t *testing.T) {
	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("input", msescript.Text("white, blue"))

	for _, fall := range []struct{ name, quelle, erwartt string }{
		{"chosen", `chosen(choice: "white")`, "true"},
		{"chosen trifft nicht", `chosen(choice: "green")`, "false"},
		{"teilweise angewandt", `(chosen@(choice: "blue"))()`, "true"},
		{"to_upper", `to_upper()`, "WHITE, BLUE"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, umgebung)
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != fall.erwartt {
				t.Errorf("%q, erwartet war %q", wert.AlsText(), fall.erwartt)
			}
		})
	}
}

// **Ein Stellungsargument schlägt das `input` der Umgebung.**
//
// Die Gegenprobe: sonst wäre die Umgebung stärker als der Aufruf, und
// `to_upper("x")` gäbe in einer Umgebung mit `input` etwas anderes als daneben.
func TestEinStellungsargumentSchlaegtDasInput(t *testing.T) {
	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("input", msescript.Text("aus der Umgebung"))

	wert, err := msescript.AuswertenQuelle(t.Context(), `to_upper("ausdrücklich")`, umgebung)
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != "AUSDRÜCKLICH" {
		t.Errorf("%q", wert.AlsText())
	}
}
