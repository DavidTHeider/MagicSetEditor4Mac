// Die Prüfungen über die **Stelle** eines Zerteilfehlers.
//
// Von `zerteiler_test.go` abgetrennt, weil die Datei über die Zeilengrenze aus
// AGENTS.md §2 gewachsen war. Der Schnitt folgt der Frage: dort steht, welchen
// Baum ein Ausdruck ergibt, hier, wohin der Finger zeigt, wenn keiner entsteht.
package msescript_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Ein Fehler tief im Ausdruck reist bis nach oben — mit seiner Stelle.
//
// Jeder dieser Fälle betritt einen anderen Zweig. Sie stehen hier nicht der
// Vollständigkeit halber: ein Fehlerpfad, der nur gebaut und nie begangen wird,
// ist eine Behauptung darüber, was im Ernstfall geschieht.
func TestEinFehlerTiefImAusdruckKommtObenAn(t *testing.T) {
	faelle := []struct {
		name   string
		quelle string
		spalte int
	}{
		{name: "die Bedingung eines if", quelle: "if * then 1 else 2", spalte: 4},
		{name: "der then-Zweig", quelle: "if a then * else 2", spalte: 11},
		{name: "der else-Zweig", quelle: "if a then 1 else *", spalte: 18},
		{name: "hinter not", quelle: "not *", spalte: 5},
		// **`a.1` war hier einmal ein Fehler und ist jetzt keiner mehr.**
		// `default_images[colors.0]` steht so in `magic-m15.mse-style`: eine
		// Liste ist eine Abbildung mit Zahlen als Namen
		// (#186).
		// Was hinter dem Punkt weiterhin nicht stehen darf, ist ein Operator.
		{name: "hinter dem Punkt steht ein Operator", quelle: "a.*", spalte: 3},
		{name: "hinter @ fehlt die Klammer", quelle: "f@x", spalte: 3},
		{name: "ein Schlüsselwort, wo ein Wert steht", quelle: "a and then", spalte: 7},
		{name: "ein Argument, das keines ist", quelle: "f(*)", spalte: 3},
		// `{a := 1` steht jetzt in `klammerstelle_test.go`: die Stelle ist
		// seit Werkstück B von Phase 4.9 die öffnende Klammer.
		{name: "eine leere geschweifte Klammer", quelle: "{}", spalte: 2},
		{name: "eine leere runde Klammer", quelle: "()", spalte: 2},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			_, err := msescript.Zerteile(fall.quelle)

			var befund *msescript.Fehler
			if !errors.As(err, &befund) {
				t.Fatalf("Fehler %v, erwartet war ein *msescript.Fehler", err)
			}

			if befund.Spalte != fall.spalte {
				t.Errorf("Spalte %d, erwartet war %d — %s",
					befund.Spalte, fall.spalte, befund.Meldung)
			}
		})
	}
}

// Die Stelle hinter dem letzten Merkmal stimmt auch nach einer Zeichenkette.
//
// **Gefunden vom Zweitblick über Phase 3.** Der Text eines
// Zeichenkettenmerkmals trägt die Anführungszeichen nicht; wer die Endstelle
// aus ihm rechnet, zeigt mitten in die Zeichenkette statt dahinter. Bei `("x"`
// meldete die fehlende Klammer Spalte 3 statt 5.
func TestDieEndstelleStimmtAuchNachEinerZeichenkette(t *testing.T) {
	faelle := []struct {
		name   string
		quelle string
		spalte int
	}{
		// **Das Fahrzeug ist ein offener Operator, nicht eine offene
		// Klammer** — seit Werkstück B von Phase 4.9
		// (#1324).
		//
		// Hier stand `("x"`, und die fehlende Klammer meldete das Quellende.
		// Sie meldet jetzt ihre **öffnende** Stelle, und damit wäre dieser
		// Test grün geworden, ohne die Endstelle noch anzusehen — ein
		// Wächter, der seine Zusage nicht mehr hält, sondern nur noch
		// dasteht. Ein Operator ohne rechte Seite meldet weiterhin das
		// Quellende und prüft dieselbe Rechnung.
		{name: "ein Operator nach einer Zeichenkette", quelle: `"x" +`, spalte: 6},
		{name: "eine Zuweisung nach einer Zeichenkette", quelle: `a := "xy" +`, spalte: 12},
		{name: "mehrbytige Zeichen in einer Zeichenkette", quelle: `"没阴影" +`, spalte: 8},
		{name: "ein Operator nach einem Namen", quelle: `abc +`, spalte: 6},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			_, err := msescript.Zerteile(fall.quelle)

			var befund *msescript.Fehler
			if !errors.As(err, &befund) {
				t.Fatalf("Fehler %v, erwartet war ein *msescript.Fehler", err)
			}

			if befund.Spalte != fall.spalte {
				t.Errorf("Spalte %d, erwartet war %d — %s",
					befund.Spalte, fall.spalte, befund.Meldung)
			}
		})
	}
}

// `if … then …` ohne `else` — die Form aus kamen-decade Zeile 63.
//
// `visible: { if card.background == "Normal" then true }`. Der erste Zerteiler
// wies sie ab, obwohl `docs/FORMATE.md` sie führt und ein echtes Paket sie
// benutzt. Gefunden vom Zweitblick über Phase 3.
func TestEinIfOhneElseErgibtNichts(t *testing.T) {
	knoten, err := msescript.Zerteile(`if a then 1`)
	if err != nil {
		t.Fatalf("abgewiesen: %v", err)
	}

	erwartet := `(Bedingung Name(a) Zahl(1) Nichts)`
	if knoten.String() != erwartet {
		t.Errorf("Baum %s, erwartet war %s", knoten.String(), erwartet)
	}

	// Und der fehlende Zweig ergibt beim Rechnen `nil`.
	wert, err := msescript.AuswertenQuelle(t.Context(), `if false then 1`, nil)
	if err != nil {
		t.Fatalf("abgewiesen: %v", err)
	}

	if wert.String() != "nil" {
		t.Errorf("%s, erwartet war nil", wert.String())
	}
}
