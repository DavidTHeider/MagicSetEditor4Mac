// **Merkmale zum Einfärben** — Werkstück 1 von Phase 3.3
// (#1164).
//
// # Warum es nicht `Zerlege` tut
//
// Zwei Gründe, und beide sind Eigenschaften des Zerteilens, nicht Mängel:
//
//   - **Kommentare fallen weg.** Der Zerteiler braucht sie nicht, ein Editor
//     schon — sie sind das Erste, was jemand eingefärbt sehen will.
//   - **Ein Fehler nimmt alles mit.** `Zerlege` gibt `nil` zurück, sobald ein
//     Zeichen nicht zur Sprache gehört. Beim Tippen ist der Ausdruck fast
//     immer unfertig; ein Editor, der dann gar nichts einfärbt, flackert bei
//     jedem zweiten Anschlag.
//
// # Und warum es trotzdem derselbe Lexer ist
//
// Ein zweiter, der „ungefähr dieselben" Regeln kennt, liefe an dem Tag
// auseinander, an dem einer von beiden eine Regel lernt — und die
// Einfärbung zeigte dann eine Sprache, die es nicht gibt (§10).
package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Ein Kommentar ist ein Merkmal.**
func TestEinKommentarIstEinMerkmal(t *testing.T) {
	t.Parallel()

	merkmale := msescript.Einfaerben("1 + 1 # das Ergebnis\n2")

	kommentare := 0

	for _, merkmal := range merkmale {
		if merkmal.Art != msescript.ArtKommentar {
			continue
		}

		kommentare++

		if merkmal.Text != "# das Ergebnis" {
			t.Errorf("Text %q", merkmal.Text)
		}

		if merkmal.Zeile != 1 || merkmal.Spalte != 7 {
			t.Errorf("Stelle %d:%d — erwartet war 1:7", merkmal.Zeile, merkmal.Spalte)
		}

		if merkmal.Breite != len("# das Ergebnis") {
			t.Errorf("Breite %d", merkmal.Breite)
		}
	}

	if kommentare != 1 {
		t.Errorf("%d Kommentare", kommentare)
	}
}

// **Der Zerteiler sieht ihn weiterhin nicht.**
//
// Die Gegenprobe: ohne sie sagte der Lauf darüber auch dann ja, wenn
// Kommentare **immer** im Merkmalsstrom stünden — und dann fiele jeder
// Ausdruck mit einem Kommentar darin auf einen Zerteilfehler.
func TestDerZerteilerSiehtKeineKommentare(t *testing.T) {
	t.Parallel()

	merkmale, fehler := msescript.Zerlege("1 + 1 # das Ergebnis")
	if fehler != nil {
		t.Fatalf("Zerlege: %v", fehler)
	}

	for _, merkmal := range merkmale {
		if merkmal.Art == msescript.ArtKommentar {
			t.Errorf("ein Kommentar im Merkmalsstrom: %v", merkmal)
		}
	}
}

// **Ein unfertiger Ausdruck färbt so weit, wie er lesbar ist.**
func TestEinUnfertigerAusdruckFaerbtSoWeitEsGeht(t *testing.T) {
	t.Parallel()

	// Die Zeichenkette hat kein schließendes Anführungszeichen — genau der
	// Stand, in dem ein Ausdruck beim Tippen die meiste Zeit ist.
	merkmale := msescript.Einfaerben(`card.name + "unfertig`)

	if len(merkmale) < 3 {
		t.Fatalf("nur %d Merkmale: %v", len(merkmale), merkmale)
	}

	if merkmale[0].Text != "card" {
		t.Errorf("das erste Merkmal ist %v", merkmale[0])
	}
}

// **Und ein Zeichen, das nicht zur Sprache gehört, hält es nicht auf.**
func TestEinFremdesZeichenHaeltDasEinfaerbenNichtAuf(t *testing.T) {
	t.Parallel()

	merkmale := msescript.Einfaerben("1 + 1 § 2")

	if len(merkmale) < 3 {
		t.Fatalf("nur %d Merkmale: %v", len(merkmale), merkmale)
	}
}

// **Die Arten sind die des Zerteilers.**
//
// Ohne diese Zusicherung könnte `Einfaerben` seine eigenen Namen erfinden, und
// die Farbe im Fenster hieße etwas anderes als der Fehler daneben.
func TestDieArtenSindDieDesZerteilers(t *testing.T) {
	t.Parallel()

	gefunden := map[msescript.Art]string{}

	for _, merkmal := range msescript.Einfaerben(`if card.name == "x" then 1 else 2`) {
		gefunden[merkmal.Art] = merkmal.Text
	}

	for _, art := range []msescript.Art{
		msescript.ArtBezeichner,
		msescript.ArtZahl,
		msescript.ArtZeichenkette,
		msescript.ArtSchluesselwort,
		msescript.ArtOperator,
	} {
		if _, da := gefunden[art]; !da {
			t.Errorf("die Art %q kam nicht vor: %v", art, gefunden)
		}
	}
}
