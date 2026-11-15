// Die Zeichenkette in einer Einbettung — `"a{f("b")}c"`
// (#468).
//
// Sie steht in einer eigenen Datei, weil `lexer_test.go` mit ihr über 400
// Zeilen ginge (§2). Was hier geprüft wird, ist **eine** Frage: wie weit eine
// Zeichenkette reicht, wenn geschweifte Klammern darin stehen.
package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Eine Zeichenkette in einer Einbettung**
// (#468).
//
// MSE2 hält einen Klammerstapel: ein `{` in einer Zeichenkette kehrt zum
// normalen Lesen zurück, ein `"` darin legt einen **zweiten** `BRACE_STRING`
// auf den Stapel, und erst das `}` liest die äußere weiter
// (`src/script/parser.cpp`, Zeilen 220–228 und 268–307, Fassung f9d9356 vom
// 2021-07-06).
func TestZerlegeZeichenketteInEinbettung(t *testing.T) {
	faelle := []struct {
		name   string
		quelle string
		text   string
	}{
		{
			name:   "eine Zeichenkette in der Einbettung",
			quelle: `"a{f("b")}c"`,
			text:   `a{f("b")}c`,
		},
		{
			name:   "zwei Zeichenketten in der Einbettung",
			quelle: `"x{if p then "j" else "n"}y"`,
			text:   `x{if p then "j" else "n"}y`,
		},
		{
			name:   "die leere Zeichenkette in der Einbettung",
			quelle: `"borders/room{if a then "_cfuse" else ""}.png"`,
			text:   `borders/room{if a then "_cfuse" else ""}.png`,
		},
		{
			name:   "eine geschweifte Klammer in der inneren Zeichenkette",
			quelle: `"a{f("}")}b"`,
			text:   `a{f("}")}b`,
		},
		{
			// **Ein geflohenes Anführungszeichen beendet die innere nicht.**
			name:   "ein Anführungszeichen in der inneren Zeichenkette",
			quelle: `"a{f("x\"y")}b"`,
			text:   `a{f("x\"y")}b`,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			marken, err := msescript.Zerlege(fall.quelle)
			if err != nil {
				t.Fatalf("Zerlege(%q) = %v, erwartet kein Fehler", fall.quelle, err)
			}

			if len(marken) != 1 || marken[0].Art != msescript.ArtZeichenkette {
				t.Fatalf("Zerlege(%q) = %v, erwartet genau eine Zeichenkette",
					fall.quelle, marken)
			}

			if marken[0].Text != fall.text {
				t.Errorf("Text = %q, erwartet %q", marken[0].Text, fall.text)
			}
		})
	}
}

// **Ein Umbruch in einer eingebetteten Zeichenkette zählt mit**
// (#468).
//
// Ohne diese Zählung zeigte jeder Fehler danach auf eine falsche Zeile — der
// Schaden, den der Zweitblick über Phase 6 schon einmal an der Fluchtfolge vor
// dem Zeilenende gefunden hat.
func TestEineEingebetteteZeichenketteZaehltIhreUmbrueche(t *testing.T) {
	quelle := "\"a{f(\"x\ny\")}b\"\nz"

	marken, err := msescript.Zerlege(quelle)
	if err != nil {
		t.Fatalf("Zerlege: %v", err)
	}

	letzte := marken[len(marken)-1]
	if letzte.Text != "z" {
		t.Fatalf("die letzte Marke ist %q, erwartet `z`", letzte.Text)
	}

	// Die Zeichenkette geht über zwei Zeilen, `z` steht auf der dritten.
	if letzte.Zeile != 3 || letzte.Spalte != 1 {
		t.Errorf("`z` steht bei %d:%d, erwartet 3:1", letzte.Zeile, letzte.Spalte)
	}
}

// **Eine innere Zeichenkette ohne Ende ist keine**
// (#468).
//
// Die äußere endet dann an ihrem Anführungszeichen, wie vor der Änderung —
// `"[\x{41]"` schreibt eine Angabe für den Mustervergleich und keine
// Einbettung. MSE2 bricht hier ab (`src/script/parser.cpp`, Zeile 260,
// Fassung f9d9356 vom 2021-07-06); dieses Programm liest weiter (§8).
func TestEineInnereZeichenketteOhneEndeIstKeine(t *testing.T) {
	marken, err := msescript.Zerlege(`"a{b" + c`)
	if err != nil {
		t.Fatalf("Zerlege: %v", err)
	}

	if len(marken) != 3 || marken[0].Text != "a{b" || marken[2].Text != "c" {
		t.Errorf("Zerlege = %v, erwartet `a{b`, `+` und `c`", marken)
	}
}
