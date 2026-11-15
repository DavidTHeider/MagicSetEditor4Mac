package msescript_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// kurz beschreibt ein erwartetes Merkmal ohne seine Stelle.
//
// Die Stelle steht in eigenen Tests. Sie in jede Tabellenzeile zu schreiben
// machte die Tabellen unlesbar, und unlesbare Tabellen werden nicht gelesen.
type kurz struct {
	art  msescript.Art
	text string
}

func TestZerlegeErkenntDieFormenAusDenPaketen(t *testing.T) {
	faelle := []struct {
		name     string
		quelle   string
		erwartet []kurz
	}{
		{
			name:   "eine Zuweisung, wie sie in jedem Stil steht",
			quelle: `is_pendulum := {not is_nopendulum()}`,
			erwartet: []kurz{
				{msescript.ArtBezeichner, "is_pendulum"},
				{msescript.ArtOperator, ":="},
				{msescript.ArtOperator, "{"},
				{msescript.ArtSchluesselwort, "not"},
				{msescript.ArtBezeichner, "is_nopendulum"},
				{msescript.ArtOperator, "("},
				{msescript.ArtOperator, ")"},
				{msescript.ArtOperator, "}"},
			},
		},
		{
			name:   "Punktzugriff über mehrere Ebenen",
			quelle: `card_style.text.content_lines <= 2`,
			erwartet: []kurz{
				{msescript.ArtBezeichner, "card_style"},
				{msescript.ArtOperator, "."},
				{msescript.ArtBezeichner, "text"},
				{msescript.ArtOperator, "."},
				{msescript.ArtBezeichner, "content_lines"},
				{msescript.ArtOperator, "<="},
				{msescript.ArtZahl, "2"},
			},
		},
		{
			name:     "eine Zahl mit Nachkomma",
			quelle:   `45.5`,
			erwartet: []kurz{{msescript.ArtZahl, "45.5"}},
		},
		{
			name:   "eine Zeichenkette",
			quelle: `styling.image_frame=="no"`,
			erwartet: []kurz{
				{msescript.ArtBezeichner, "styling"},
				{msescript.ArtOperator, "."},
				{msescript.ArtBezeichner, "image_frame"},
				{msescript.ArtOperator, "=="},
				{msescript.ArtZeichenkette, "no"},
			},
		},
		{
			name:   "die Bindung benannter Argumente",
			quelle: `sort_text@(order: "XY")()`,
			erwartet: []kurz{
				{msescript.ArtBezeichner, "sort_text"},
				{msescript.ArtOperator, "@"},
				{msescript.ArtOperator, "("},
				{msescript.ArtBezeichner, "order"},
				{msescript.ArtOperator, ":"},
				{msescript.ArtZeichenkette, "XY"},
				{msescript.ArtOperator, ")"},
				{msescript.ArtOperator, "("},
				{msescript.ArtOperator, ")"},
			},
		},
		{
			name:   "eine Rechnung",
			quelle: `length(remove_tags(card.type_1))*17 + 2`,
			erwartet: []kurz{
				{msescript.ArtBezeichner, "length"},
				{msescript.ArtOperator, "("},
				{msescript.ArtBezeichner, "remove_tags"},
				{msescript.ArtOperator, "("},
				{msescript.ArtBezeichner, "card"},
				{msescript.ArtOperator, "."},
				{msescript.ArtBezeichner, "type_1"},
				{msescript.ArtOperator, ")"},
				{msescript.ArtOperator, ")"},
				{msescript.ArtOperator, "*"},
				{msescript.ArtZahl, "17"},
				{msescript.ArtOperator, "+"},
				{msescript.ArtZahl, "2"},
			},
		},
		{
			name:   "ein Kommentar bis Zeilenende verschwindet",
			quelle: "1 # 没阴影\n2",
			erwartet: []kurz{
				{msescript.ArtZahl, "1"},
				{msescript.ArtZeilenende, "\n"},
				{msescript.ArtZahl, "2"},
			},
		},
		{
			name:   "die Zeile trennt zwei Zuweisungen",
			quelle: "a := 1\nb := 2",
			erwartet: []kurz{
				{msescript.ArtBezeichner, "a"},
				{msescript.ArtOperator, ":="},
				{msescript.ArtZahl, "1"},
				{msescript.ArtZeilenende, "\n"},
				{msescript.ArtBezeichner, "b"},
				{msescript.ArtOperator, ":="},
				{msescript.ArtZahl, "2"},
			},
		},
		{
			name:   "ungleich ist ein Merkmal, nicht zwei",
			quelle: `a != nil`,
			erwartet: []kurz{
				{msescript.ArtBezeichner, "a"},
				{msescript.ArtOperator, "!="},
				{msescript.ArtSchluesselwort, "nil"},
			},
		},
		{
			// `notieren` fängt nicht mit dem Schlüsselwort `not` an, es
			// **enthält** es. Ein Lexer, der auf Präfix vergleicht, zerlegte
			// hier in `not` und `ieren` — und der Fehler zeigte sich erst an
			// einer Vorlage, die niemand geschrieben hat, um ihn zu finden.
			name:     "ein Bezeichner, der mit einem Schlüsselwort beginnt",
			quelle:   `notieren`,
			erwartet: []kurz{{msescript.ArtBezeichner, "notieren"}},
		},
		{
			name:     "leerer Quelltext ergibt nichts",
			quelle:   "",
			erwartet: nil,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			merkmale, err := msescript.Zerlege(fall.quelle)
			if err != nil {
				t.Fatalf("abgewiesen: %v", err)
			}

			vergleiche(t, merkmale, fall.erwartet)
		})
	}
}

// Jedes Merkmal trägt seine Stelle, und Spalten zählen in Zeichen.
//
// **Nicht in Bytes.** Kommentare in diesen Paketen sind chinesisch; eine Spalte
// in Bytes zeigte dort auf die Mitte eines Zeichens und wäre als Wegweiser
// unbrauchbar.
func TestJedesMerkmalTraegtSeineStelle(t *testing.T) {
	merkmale, err := msescript.Zerlege("a := 1\n  没 := 2")
	if err != nil {
		t.Fatalf("abgewiesen: %v", err)
	}

	erwartet := []msescript.Merkmal{
		{Art: msescript.ArtBezeichner, Text: "a", Zeile: 1, Spalte: 1, Breite: 1},
		{Art: msescript.ArtOperator, Text: ":=", Zeile: 1, Spalte: 3, Breite: 2},
		{Art: msescript.ArtZahl, Text: "1", Zeile: 1, Spalte: 6, Breite: 1},
		{Art: msescript.ArtZeilenende, Text: "\n", Zeile: 1, Spalte: 7, Breite: 1},
		{Art: msescript.ArtBezeichner, Text: "没", Zeile: 2, Spalte: 3, Breite: 1},
		{Art: msescript.ArtOperator, Text: ":=", Zeile: 2, Spalte: 5, Breite: 2},
		{Art: msescript.ArtZahl, Text: "2", Zeile: 2, Spalte: 8, Breite: 1},
	}

	if len(merkmale) != len(erwartet) {
		t.Fatalf("%d Merkmale, erwartet waren %d: %v", len(merkmale), len(erwartet), merkmale)
	}

	for i, soll := range erwartet {
		if merkmale[i] != soll {
			t.Errorf("Merkmal %d ist %v, erwartet war %v", i, merkmale[i], soll)
		}
	}
}

func TestZerlegeMeldetDieStelleDesFehlers(t *testing.T) {
	faelle := []struct {
		name   string
		quelle string
		zeile  int
		spalte int
	}{
		{
			name:   "eine Zeichenkette ohne Ende",
			quelle: "a := \"offen",
			zeile:  1,
			spalte: 6,
		},
		{
			name:   "eine Zeichenkette, die an der Zeile endet",
			quelle: "a := \"offen\nb := 1",
			zeile:  1,
			spalte: 6,
		},
		{
			name:   "ein Zeichen, das die Sprache nicht kennt",
			quelle: "a := 1\nb := 2 §",
			zeile:  2,
			spalte: 8,
		},
		{
			name:   "ein einzelnes Gleichheitszeichen ist keine Zuweisung",
			quelle: "a = 1",
			zeile:  1,
			spalte: 3,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			_, err := msescript.Zerlege(fall.quelle)

			var befund *msescript.Fehler
			if !errors.As(err, &befund) {
				t.Fatalf("Fehler %v, erwartet war ein *msescript.Fehler", err)
			}

			if befund.Zeile != fall.zeile || befund.Spalte != fall.spalte {
				t.Errorf("Stelle %d:%d, erwartet war %d:%d",
					befund.Zeile, befund.Spalte, fall.zeile, fall.spalte)
			}
		})
	}
}

// Ein einzelnes `=` ist die häufigste Verwechslung, die es geben wird.
//
// Deshalb sagt die Meldung, was gemeint war — nicht nur, dass etwas nicht geht.
func TestEinGleichheitszeichenNenntDieZuweisung(t *testing.T) {
	_, err := msescript.Zerlege("a = 1")

	var befund *msescript.Fehler
	if !errors.As(err, &befund) {
		t.Fatalf("Fehler %v, erwartet war ein *msescript.Fehler", err)
	}

	if !containsAll(befund.Meldung, ":=") {
		t.Errorf("die Meldung %q nennt die Zuweisung nicht", befund.Meldung)
	}
}

func containsAll(text string, teile ...string) bool {
	for _, teil := range teile {
		found := false
		for i := 0; i+len(teil) <= len(text); i++ {
			if text[i:i+len(teil)] == teil {
				found = true

				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}

func vergleiche(t *testing.T, merkmale []msescript.Merkmal, erwartet []kurz) {
	t.Helper()

	if len(merkmale) != len(erwartet) {
		t.Fatalf("%d Merkmale, erwartet waren %d: %v", len(merkmale), len(erwartet), merkmale)
	}

	for i, soll := range erwartet {
		if merkmale[i].Art != soll.art || merkmale[i].Text != soll.text {
			t.Errorf("Merkmal %d ist %s(%q), erwartet war %s(%q)",
				i, merkmale[i].Art, merkmale[i].Text, soll.art, soll.text)
		}
	}
}

// Beide Darstellungen tragen die Stelle mit.
//
// Sie sehen aus wie Bequemlichkeit und sind es nicht: eine fehlgeschlagene
// Tabelle zeigt `Operator(":=") at 2:5`, und ein Fehler ohne seine Stelle
// wäre in einem 1500-Zeilen-Stil unbrauchbar (AGENTS.md §8).
func TestBeideDarstellungenNennenDieStelle(t *testing.T) {
	merkmal := msescript.Merkmal{
		Art: msescript.ArtOperator, Text: ":=", Zeile: 2, Spalte: 5,
	}
	if merkmal.String() != `Operator(":=") at 2:5` {
		t.Errorf("Darstellung %q", merkmal.String())
	}

	fehler := &msescript.Fehler{Zeile: 7, Spalte: 3, Meldung: "etwas fehlt"}
	if fehler.Error() != "Zeile 7, Spalte 3: etwas fehlt" {
		t.Errorf("Meldung %q", fehler.Error())
	}
}

// **Ein Rückstrich flieht das nächste Zeichen.**
//
// # Die dritte Fassung dieses Tests, und die erste mit einem Beleg
//
// Zuerst stand hier „Zeichenketten kennen keine Maskierung", geschlossen
// daraus, dass keine vorkomme — sie kam vor, in derselben Datei, aus der die
// Vermutung stammte. Dann stand hier „ein Rückstrich reist unverändert mit",
// weil ein einzelnes `contains(match:"\n")` nicht entscheidet, ob das ein
// Umbruch ist oder zwei Zeichen.
//
// **Die Exportvorlagen entscheiden es** (#51):
// fünf von neunzehn schreiben `\"` mitten in einer Zeichenkette, und ohne
// Fluchtfolge endet sie dort — diese fünf Dateien wären nicht lesbar. `\"` und
// `\\` sind damit **belegt**; dass `\n` ein Umbruch ist, ist der folgerichtige
// Schluss daraus, dass dieselben Vorlagen `"\\n"` schreiben, wo sie die zwei
// Zeichen eines Regex meinen.
//
// **Ein Beleg aus einem Fehlen ist so viel wert wie die Suche, die nichts
// fand.** Dreimal derselbe Satz an derselben Stelle.
func TestEinRueckstrichFliehtDasNaechsteZeichen(t *testing.T) {
	merkmale, err := msescript.Zerlege(`contains(match:"\n")`)
	if err != nil {
		t.Fatalf("abgewiesen: %v", err)
	}

	vergleiche(t, merkmale, []kurz{
		{msescript.ArtBezeichner, "contains"},
		{msescript.ArtOperator, "("},
		{msescript.ArtBezeichner, "match"},
		{msescript.ArtOperator, ":"},
		{msescript.ArtZeichenkette, "\n"},
		{msescript.ArtOperator, ")"},
	})
}
