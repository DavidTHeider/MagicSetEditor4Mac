// Der Wächter über `docs/STILOPTIONEN.md`.
//
// Die Datei hält fest, welche `styling field:`-Arten das Pack führt und wo
// jede im Fenster bedient wird
// (#630).
// Sie entsteht aus einer Zählung von Hand, und eine solche veraltet still:
// jemand ergänzt ein Bedienelement, und die Datei sagt nichts dazu.
//
// # Was er prüft
//
//   - Jede Art, für die `service.moeglichkeitenVon` eine Wahlliste baut, hat
//     eine Zeile und ist dort mit ✅ geführt. **Das ist die Stelle, die es
//     entscheidet** — seit #634
//     schaltet das Fenster nicht mehr auf die Art, sondern darauf, ob eine
//     Liste dasteht.
//   - Jede Art, die `Stiloption.tsx` **ausdrücklich** behandelt
//     (`option.art === '…'`), hat eine Zeile und ist dort mit ✅ geführt. So
//     kommt `info` zu seinem Haken: es bekommt keine Wahlliste, sondern eine
//     Überschrift.
//   - Jede Art, die eine dieser beiden Dateien beim Namen nennt, hat eine
//     Zeile.
//   - Jede Art, die `msepaket.feldarten` kennt, hat eine Zeile.
//   - Die ✅-Zeilen sind **genau** die mit Wahlliste oder eigenem Zweig, plus
//     `text` — das ist der Rückfall, und ein Textfeld ist für einen Text die
//     richtige Antwort.
//
// # Was er nicht prüft
//
// **Die Zahlen.** Das Pack liegt nicht im Repositorium
// ([0005](../../docs/entscheidungen/0005-keine-fremden-vorlagen.md)), und ein
// Lauf dagegen liefe nicht in der Pipeline. Er hält die **Namen**.
//
// **Ob die Bedienstelle die richtige ist.** Dass `color` ein Textfeld bekommt,
// sieht er als „vom Rückfall bedient" und nicht als Mangel. Der Mangel steht in
// der Spalte „Trägt", und die liest ein Mensch.
//
// **Ob das Pack noch dieselben Arten führt.** Eine zehnte Art fiele hier durch
// und stünde als Text im Fenster; bemerkt würde sie beim nächsten Zählen.
//
// Beides steht auch in der Datei selbst. Ein Wächter, der die halbe Zusage
// hält, macht die andere unsichtbar (§10).
package werkzeuge

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// genannteArten findet die Arten, die das Fenster beim Namen nennt — gleich ob
// als Gleichheit oder als Ausnahme.
var genannteArten = regexp.MustCompile(`option\.art [!=]== '([^']+)'`)

// eigeneZweige findet die Arten mit einem **eigenen** Zweig. Eine Ausnahme
// (`!==`) zählt nicht: sie sagt, wo eine Art **nicht** hingehört.
var eigeneZweige = regexp.MustCompile(`option\.art === '([^']+)'`)

// wahlarten findet die Fälle des Schalters in `service.moeglichkeitenVon`.
var wahlarten = regexp.MustCompile(`(?m)^\tcase (domain\.Kind\w+(?:, domain\.Kind\w+)*):`)

// kindnamen bildet einen Go-Bezeichner auf den Namen im Paket ab.
var kindnamen = map[string]string{
	"domain.KindText":           "text",
	"domain.KindChoice":         "choice",
	"domain.KindImage":          "image",
	"domain.KindColor":          "color",
	"domain.KindBoolean":        "boolean",
	"domain.KindInfo":           "info",
	"domain.KindMultipleChoice": "multiple choice",
	"domain.KindPackageChoice":  "package choice",
}

// gekannteArten findet die Schlüssel von `msepaket.feldarten`.
var gekannteArten = regexp.MustCompile(`(?m)^\s*"([^"]+)":\s*domain\.Kind`)

// artenzeile findet eine Zeile der Artentabelle: Art, dann irgendwann ✅/❌.
var artenzeile = regexp.MustCompile("(?m)^\\| `([^`]+)` \\|.*\\|\\s*([✅❌])\\s*\\|$")

// derRueckfall wird von der Stelle bedient, die alles übrige auffängt.
//
// **Und das ist keine Lücke**: ein Textfeld ist für einen Text die richtige
// Antwort, nicht der Rückfall aus Verlegenheit.
const derRueckfall = "text"

func TestJedeStiloptionsartStehtInIhrerDatei(t *testing.T) {
	datei := lesen(t, filepath.Join("..", "..", "docs", "STILOPTIONEN.md"))

	gefuehrt := map[string]string{}

	for _, treffer := range artenzeile.FindAllStringSubmatch(datei, -1) {
		gefuehrt[treffer[1]] = treffer[2]
	}

	if len(gefuehrt) == 0 {
		t.Fatal("docs/STILOPTIONEN.md führt keine einzige Art — der Wächter prüfte nichts")
	}

	bedient := map[string]bool{derRueckfall: true}

	dienst := lesen(t, filepath.Join("..", "..", "internal", "service", "paketwahl.go"))
	for _, treffer := range wahlarten.FindAllStringSubmatch(dienst, -1) {
		for _, bezeichner := range strings.Split(treffer[1], ", ") {
			art, bekannt := kindnamen[bezeichner]
			if !bekannt {
				t.Errorf("`moeglichkeitenVon` nennt %s — dieser Wächter kennt den Bezeichner nicht",
					bezeichner)

				continue
			}

			bedient[art] = true

			if gefuehrt[art] != "✅" {
				t.Errorf("`moeglichkeitenVon` baut eine Wahlliste für %q, "+
					"docs/STILOPTIONEN.md führt es als %q", art, gefuehrt[art])
			}
		}
	}

	fenster := lesen(t, filepath.Join("..", "..", "frontend", "src", "seiten", "Stiloption.tsx"))
	for _, treffer := range genannteArten.FindAllStringSubmatch(fenster, -1) {
		if _, da := gefuehrt[treffer[1]]; !da {
			t.Errorf("`Stiloption.tsx` nennt %q, docs/STILOPTIONEN.md nicht", treffer[1])
		}
	}

	// **Ein eigener Zweig zählt wie eine Wahlliste.** `info` bekommt keine —
	// es ist eine Überschrift, kein Feld (#639).
	for _, treffer := range eigeneZweige.FindAllStringSubmatch(fenster, -1) {
		bedient[treffer[1]] = true

		if gefuehrt[treffer[1]] != "✅" {
			t.Errorf("`Stiloption.tsx` behandelt %q, docs/STILOPTIONEN.md führt es als %q",
				treffer[1], gefuehrt[treffer[1]])
		}
	}

	paket := lesen(t, filepath.Join("..", "..", "internal", "msepaket", "spiel.go"))
	for _, treffer := range gekannteArten.FindAllStringSubmatch(paket, -1) {
		if _, da := gefuehrt[treffer[1]]; !da {
			t.Errorf("`msepaket.feldarten` kennt %q, docs/STILOPTIONEN.md nicht", treffer[1])
		}
	}

	// **Die Gegenrichtung**: ein ✅ ohne Bedienstelle wäre eine Zusage, die
	// niemand einlöst.
	var zuviel []string

	for art, marke := range gefuehrt {
		if marke == "✅" && !bedient[art] {
			zuviel = append(zuviel, art)
		}
	}

	sort.Strings(zuviel)

	if len(zuviel) > 0 {
		t.Errorf("docs/STILOPTIONEN.md führt %v als bedient — weder baut "+
			"`service.moeglichkeitenVon` ihnen eine Wahlliste noch hat "+
			"`Stiloption.tsx` einen eigenen Zweig dafür", zuviel)
	}
}

// lesen holt eine Datei oder bricht ab.
func lesen(t *testing.T, pfad string) string {
	t.Helper()

	inhalt, err := os.ReadFile(pfad) // #nosec G304 -- feste Pfade im Repositorium
	if err != nil {
		t.Fatalf("%s: %v", pfad, err)
	}

	return strings.ReplaceAll(string(inhalt), "\r\n", "\n")
}
