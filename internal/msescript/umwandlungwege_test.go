package msescript_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Die Wege durch die Umwandlung, die kein Skript von selbst nimmt.
//
// **Jede Art einmal**, denn eine Umwandlung, die für eine Art stillschweigend
// null liefert, ist genau der Fehler, den `AGENTS.md` §8 meint: sie sieht aus
// wie ein Ergebnis. Die Tabelle in `umwandlung.go` behauptet für Abbildung,
// Funktion und Farbe einen Fehler — hier steht die Gegenprobe dazu.
func TestJedeArtMeldetIhreUmwandlung(t *testing.T) {
	faelle := []struct {
		name   string
		quelle string
	}{
		{"eine Abbildung rechnet nicht", `[a: 1] * 2`},
		{"eine Funktion rechnet nicht", `{ 1 } * 2`},
		{"eine Farbe rechnet nicht", `rgb(1,2,3) * 2`},

		{"eine Abbildung ist keine Bedingung", `if [a: 1] then 1 else 2`},
		{"eine Funktion ist keine Bedingung", `if { 1 } then 1 else 2`},
		{"eine Farbe ist keine Bedingung", `if rgb(1,2,3) then 1 else 2`},

		{"eine Abbildung vergleicht nicht", `[a: 1] > 2`},
		{"rechts vergleicht nicht", `2 > [a: 1]`},
		{"rechts rechnet nicht", `2 * [a: 1]`},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			_, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err == nil {
				t.Fatal("kein Fehler — die Umwandlung gelang stillschweigend")
			}
		})
	}
}

// **`nil` gerufen mit einem Argument gibt dieses Argument zurück** — das ist
// der Fall, den der Kommentar in `value.cpp:456` wörtlich nennt:
// `nil(input) == input`.
func TestNichtsGerufenReichtSeinArgumentDurch(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`nil("durchgereicht")`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != "durchgereicht" {
		t.Errorf("%q", wert.AlsText())
	}
}

// **Und ohne Argument nimmt es das `input` aus der Umgebung.**
//
// So steht es in den Stilen: `{ module_card() }` läuft in einer Umgebung, in
// der `input` gesetzt ist — der Aufruf hat keine Klammerfüllung, und trotzdem
// ist die Eingabe da.
func TestNichtsGerufenNimmtDasInputDerUmgebung(t *testing.T) {
	umgebung := msescript.Bibliothek()
	umgebung.Setze("input", msescript.Text("aus der Umgebung"))

	wert, err := msescript.AuswertenQuelle(t.Context(), `nil()`, umgebung)
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != "aus der Umgebung" {
		t.Errorf("%q", wert.AlsText())
	}
}

// **Der Vergleich zweier Zeichenketten bleibt die Zeichenordnung.**
//
// Eine bewusste Abweichung von MSE, das dort `toInt` riefe und für jedes Wort
// würfe. Sie steht in `auswerter_verknuepfung.go` begründet; hier ist sie
// festgehalten, damit sie nicht beim nächsten Abgleich mit MSE2 stillschweigend
// verschwindet.
func TestZweiZeichenkettenVergleichenSichNachZeichen(t *testing.T) {
	faelle := []struct {
		quelle  string
		erwartt bool
	}{
		{`"Alpha" < "Beta"`, true},
		{`"Beta" < "Alpha"`, false},
		{`"gleich" <= "gleich"`, true},
	}

	for _, fall := range faelle {
		t.Run(fall.quelle, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.Wahrheit != fall.erwartt {
				t.Errorf("%v, erwartet war %v", wert.Wahrheit, fall.erwartt)
			}
		})
	}
}

// **Eine Zahl gegen einen Text vergleicht als Zahl.**
//
// Die Gegenprobe zur Bedingung „beide Seiten Text": sonst gewänne die
// Zeichenordnung, und `"9" < 10` wäre falsch, weil `"9" > "1"` gilt.
func TestZahlGegenTextVergleichtAlsZahl(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(), `"9" < 10`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if !wert.Wahrheit {
		t.Error("die Zeichenordnung hat gegen die Zahl gewonnen")
	}
}

// **Ein unlesbarer Text auf der rechten Seite wird ebenso benannt.**
//
// Die Meldung sagt, *welche* Seite — ohne das steht bei einem Feldstil mit zwei
// Aufrufen nur „rechnet mit Zahlen", und beide Seiten sind zu prüfen.
func TestDieSeiteStehtInDerMeldung(t *testing.T) {
	for _, fall := range []struct{ quelle, seite string }{
		{`"links" * 2`, "links"},
		{`2 * "rechts"`, "rechts"},
		{`"links" > 2`, "links"},
		{`2 > "rechts"`, "rechts"},
	} {
		t.Run(fall.quelle, func(t *testing.T) {
			_, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err == nil {
				t.Fatal("kein Fehler")
			}

			if !strings.Contains(err.Error(), fall.seite) {
				t.Errorf("die Meldung nennt die Seite nicht: %v", err)
			}
		})
	}
}
