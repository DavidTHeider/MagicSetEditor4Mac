package msescript_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Zwei Funktionen mit `+` ergeben eine, die beide nacheinander anwendet.**
//
// # Erschlossen, nicht erfunden
//
// `magic-mtgsalvation-forum.mse-export-template` schreibt:
//
//	mtgsally_mse_forum_cost := replace@(match:"2/W", replace:":sym2w:")+
//		replace@(match:"2/U", replace:":sym2u:")+
//		…
//
// Das ergibt nur einen Sinn, wenn `+` die Anwendungen **hintereinander** legt:
// erst die linke, dann die rechte auf deren Ergebnis. Dieselbe Form steht in
// `magic-m15-godzilla-planeswalker` als `replace@(…) + filter_text@(…)` —
// zusammen sieben der neunzehn bezogenen Vorlagen
// (#202).
func TestZweiFunktionenLassenSichVerketten(t *testing.T) {
	// **Gerufen wird mit einem Stellungsargument**, so wie
	// `magic-mtgsalvation-forum` es tut: `paragraph_count(card.rule_text)`.
	// Der Zwischenwert nimmt denselben Platz.
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`f := replace@(match:"a", replace:"b") + filter_text@(match:"b"); f("axa")`,
		msescript.Bibliothek().Darueber())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.Text != "bb" {
		t.Errorf("%q, erwartet bb", wert.Text)
	}
}

// **Die Reihenfolge ist links vor rechts.**
//
// Der Test steht eigens, weil die andere Richtung dieselbe Länge hat und bei
// einem symmetrischen Beispiel dasselbe Ergebnis gäbe.
func TestDieVerkettungLaeuftVonLinksNachRechts(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`f := replace@(match:"x", replace:"y") + replace@(match:"y", replace:"z"); f("x")`,
		msescript.Bibliothek().Darueber())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	// Links zuerst: x → y, dann y → z. Andersherum bliebe es bei y.
	if wert.Text != "z" {
		t.Errorf("%q, erwartet z — die Verkettung läuft verkehrt herum", wert.Text)
	}
}

// **Auch drei und mehr.**
//
// `mtgsally_mse_forum_cost` verkettet zehn.
func TestMehrereFunktionenLassenSichVerketten(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`f := replace@(match:"1", replace:"2") + replace@(match:"2", replace:"3") + `+
			`replace@(match:"3", replace:"4"); f("1")`,
		msescript.Bibliothek().Darueber())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.Text != "4" {
		t.Errorf("%q, erwartet 4", wert.Text)
	}
}

// **Eine Funktion und etwas anderes bleibt ein Fehler.**
//
// `"text" + to_upper` ist keine Verkettung und kein Text; wer daraus etwas
// machte, verstünde eine Vorlage anders als MSE.
func TestEineFunktionUndEinTextLassenSichNichtVerbinden(t *testing.T) {
	_, err := msescript.AuswertenQuelle(t.Context(), `"a" + replace@(match:"x")`,
		msescript.Bibliothek().Darueber())
	if err == nil {
		t.Error("kein Fehler")
	}
}

// **Auch eine Vorlagenfunktion lässt sich verketten.**
//
// Keine der neunzehn Vorlagen tut es — sie verketten eingebaute. Der Zweig
// steht trotzdem, weil ein Funktionswert aus `{ … }` genauso aussieht, und ein
// Zweig, den kein Test betritt, ist eine Behauptung darüber, was im Ernstfall
// geschieht (§3).
func TestAuchEinBlockLaesstSichVerketten(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`f := { input + "!" } + { input + "?" }; f(input: "hallo")`,
		msescript.Bibliothek().Darueber())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != "hallo!?" {
		t.Errorf("%q, erwartet hallo!?", wert.AlsText())
	}
}

// **Ein Fehler in der linken Hälfte kommt heraus.**
func TestEinFehlerInDerLinkenHaelfteKommtHeraus(t *testing.T) {
	_, err := msescript.AuswertenQuelle(t.Context(),
		`f := { gibtesnicht() } + { input }; f("x")`, msescript.Bibliothek().Darueber())
	if err == nil {
		t.Error("kein Fehler")
	}
}

// **Und einer in der rechten auch.**
func TestEinFehlerInDerRechtenHaelfteKommtHeraus(t *testing.T) {
	_, err := msescript.AuswertenQuelle(t.Context(),
		`f := { input } + { gibtesnicht() }; f("x")`, msescript.Bibliothek().Darueber())
	if err == nil {
		t.Error("kein Fehler")
	}
}

// **Ein Verweis auf ein anderes Skript desselben Netzes lässt sich nicht
// verketten — und die Meldung sagt das.**
//
// Er hat weder Rumpf noch Go-Funktion; ihn zu rufen ginge nur über das Netz,
// und die Verkettung kennt keines. Kein gesichtetes Paket tut es; ohne diesen
// Fall stünde dort trotzdem ein Zweig, den kein Test betritt (§3).
func TestEinNetzverweisLaesstSichNichtVerketten(t *testing.T) {
	netz := netzAus(t, map[string]string{
		"a": `"x"`,
		"b": `f := a + a; f("y")`,
	}, nil)

	_, err := netz.Wert(t.Context(), "b")
	if err == nil {
		t.Fatal("kein Fehler")
	}

	if !strings.Contains(err.Error(), "anderes Skript") {
		t.Errorf("%v — die Meldung nennt den Fall nicht", err)
	}
}

// **Ein Block mit vorab gebundenen Argumenten behält sie in der Verkettung.**
//
// `{ … }@(a: 1)` bindet, ohne zu rufen. Ohne diesen Fall bliebe der Zweig
// ungeprüft, der die Bindung in die Umgebung der verketteten Hälfte legt.
func TestEinGebundenerBlockBehaeltSeineArgumenteInDerVerkettung(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`f := { input + zusatz }@(zusatz: "!") + { input + "?" }; f(input: "hallo")`,
		msescript.Bibliothek().Darueber())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != "hallo!?" {
		t.Errorf("%q, erwartet hallo!?", wert.AlsText())
	}
}
