// **Eine Paketdatei bekommt ihre eigene Einfärbung** —
// #1181.
//
// # Der Befund
//
// `Quelltextfeld` nimmt eine Sprache, und im Vorlagenbereich stand `keine` —
// Zeilennummern ja, Farbe nein. Das war **richtig herum geirrt**: der
// Skript-Lexer liest aus `game: magic` Bezeichner und Doppelpunkte und färbt
// entsprechend nichts Sinnvolles, und keine Farbe ist besser als eine falsche.
//
// # Warum das hier steht und nicht daneben
//
// **`zerteiler.go` liest dieses Format schon** — samt Einrückung,
// mehrzeiligen Werten und der Frage, wo ein Schlüssel aufhört. Ein zweiter
// Leser daneben liefe an dem Tag auseinander, an dem einer von beiden eine
// Regel lernt (§10). Die Einfärbung geht deshalb durch [Zerteile] und liest
// die Knoten, statt den Text ein zweites Mal zu zerlegen.
package msepaket_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// stellen beschreibt einen erwarteten Abschnitt.
type stelle struct {
	art    string
	zeile  int
	spalte int
	breite int
}

func TestEinePaketdateiWirdEingefaerbt(t *testing.T) {
	t.Parallel()

	quelle := "# ein Kommentar\ngame: magic\ncard style:\n\tname:\n\t\tscript: to_title(value)\n"

	merkmale := msepaket.Einfaerben(quelle)

	erwartet := []stelle{
		{"Kommentar", 1, 1, 15},
		{"Schlüssel", 2, 1, 4},
		{"Wert", 2, 7, 5},
		{"Schlüssel", 3, 1, 10},
		{"Schlüssel", 4, 2, 4},
		{"Schlüssel", 5, 3, 6},
	}

	if len(merkmale) < len(erwartet) {
		t.Fatalf("%d Merkmale, erwartet mindestens %d:\n%+v",
			len(merkmale), len(erwartet), merkmale)
	}

	for i, will := range erwartet {
		hat := merkmale[i]
		if string(hat.Art) != will.art || hat.Zeile != will.zeile ||
			hat.Spalte != will.spalte || hat.Breite != will.breite {
			t.Errorf("Merkmal %d ist %+v, erwartet %+v", i, hat, will)
		}
	}
}

// **Ein `script:`-Wert wird als Skriptwert benannt** — gefärbt wird er im
// Dienst.
//
// **Warum nicht hier**: `msepaket` darf `msescript` nicht kennen (§2). Was ein
// Skriptwert **ist**, weiß das Paketformat; **wie ein Skript aussieht**, weiß
// die Sprache.
func TestEinSkriptwertWirdBenannt(t *testing.T) {
	t.Parallel()

	merkmale := msepaket.Einfaerben("\tscript: if x then 1 else 2\n")

	if len(merkmale) != 2 || string(merkmale[1].Art) != "Skriptwert" {
		t.Fatalf("Merkmale %+v, erwartet Schlüssel und Skriptwert", merkmale)
	}

	if merkmale[1].Text != "if x then 1 else 2" {
		t.Errorf("der Text ist %q — wer ihn weiterfärbt, braucht ihn",
			merkmale[1].Text)
	}
}

// **Eine leere Zeile bekommt kein Merkmal.**
func TestEineLeereZeileBekommtKeinMerkmal(t *testing.T) {
	t.Parallel()

	if merkmale := msepaket.Einfaerben("\n\n"); len(merkmale) != 0 {
		t.Errorf("%d Merkmale, erwartet keins: %+v", len(merkmale), merkmale)
	}
}

// **Eine Fortsetzungszeile ist ein Wert.**
//
// Sie gehört zu dem Block über ihr — `]` am Ende einer mehrzeiligen
// Skriptklammer ist kein Schlüssel.
func TestEineFortsetzungIstEinWert(t *testing.T) {
	t.Parallel()

	merkmale := msepaket.Einfaerben("\tliste:\n\t\teins\n")

	if len(merkmale) != 2 || string(merkmale[1].Art) != "Wert" {
		t.Errorf("Merkmale %+v, erwartet Schlüssel und Wert", merkmale)
	}
}

// **Eine Datei, die sich nicht zerteilen lässt, färbt nichts.**
//
// Dieselbe Haltung wie beim Skript-Lexer: wo ein Fehler steht, sagt der
// Zerteiler; das Fenster fragt ohnehin beides.
func TestEineUnzerteilbareDateiFaerbtNichts(t *testing.T) {
	t.Parallel()

	// **Ein leerer Schlüssel** — der Zerteiler weist das ab.
	if merkmale := msepaket.Einfaerben(": ohne Namen\n"); len(merkmale) != 0 {
		t.Errorf("%d Merkmale, erwartet keins: %+v", len(merkmale), merkmale)
	}
}

// **Eine leere Fortsetzungszeile bekommt kein Merkmal.**
//
// Sie steht im Rumpf eines mehrzeiligen Skripts und trägt nichts als ihre
// Einrückung.
func TestEineLeereFortsetzungBekommtKeinMerkmal(t *testing.T) {
	t.Parallel()

	merkmale := msepaket.Einfaerben("\tliste:\n\t\t\n\t\teins\n")

	for _, merkmal := range merkmale {
		if merkmal.Breite == 0 {
			t.Errorf("ein Merkmal ohne Breite: %+v", merkmal)
		}
	}
}

// **Ein Kommentar mitten in einem Block bekommt seine Farbe.**
//
// Er steht als Kind unter seinem Vater, und der rekursive Lauf muss ihn
// erreichen — ein Lauf nur über die Wurzel sähe ihn nicht.
func TestEinKommentarImBlockWirdGefaerbt(t *testing.T) {
	t.Parallel()

	merkmale := msepaket.Einfaerben("card style:\n\t# ein Hinweis\n\tname: x\n")

	kommentare := 0
	for _, merkmal := range merkmale {
		if string(merkmal.Art) == "Kommentar" {
			kommentare++
		}
	}

	if kommentare != 1 {
		t.Errorf("%d Kommentare, erwartet einen: %+v", kommentare, merkmale)
	}
}

// **Ein Skriptschlüssel aus zwei Wörtern ist einer.** `init script:` trägt
// denselben Rumpf wie `script:` — und bis zum 2026-09-09 fiel er durch:
// nachgeschlagen wurde die kanonische Form (`init_script`), die Liste führte
// die Form der Datei. Gefunden beim Bau des Symbolindex (U7, #1418): kein
// `init script` eines Spiels hatte je eine Farbe oder einen Befund, und jede
// Zuweisung darin las der Zerteiler als Schlüssel `kosten ` mit Wert `= 1`.
func TestEinMehrwortigerSkriptschluesselTraegtEinSkript(t *testing.T) {
	t.Parallel()

	// **`choice images` stand hier bis FMT-02** (#1691) — es ist aber kein
	// Skript, sondern eine Liste von Bildern; `sort script` trägt dieselbe
	// Frage nach dem Schlüssel aus zwei Wörtern.
	for _, schluessel := range []string{"init script", "init_script", "sort script"} {
		var skripte []msepaket.Farbmerkmal

		for _, merkmal := range msepaket.Einfaerben(schluessel + ":\n\tkosten := 1\n") {
			if merkmal.Art == msepaket.FarbeSkriptwert {
				skripte = append(skripte, merkmal)
			}
		}

		if len(skripte) != 1 || skripte[0].Text != "kosten := 1" {
			t.Errorf("%s: Skriptwerte %+v", schluessel, skripte)
		}
	}
}
