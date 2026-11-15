// Package sammelliste liest eine eingefügte Kartenliste in Zeilen.
package sammelliste

// **Eine Kartenliste lesen** — Werkstück 2 von Phase 4.5
// (#1292).
//
// # Was dieses Paket ist
//
// Ein **Blatt** (§2): es kennt `internal/domain` und sonst nichts. Text
// hinein, Zeilen heraus — kein Netz, kein Bestand. Die **Zuordnung** einer
// Zeile auf einen bestimmten Druck braucht Scryfall und steht in
// `internal/service`.
//
// # Das Format
//
// Die gewachsene Konvention der Deckbaukästen, in der Gestalt, die sich aus
// MTGO, Arena, Moxfield und Archidekt überschneidet:
//
//	[<menge>[x]] <name> [(<setcode>) [<nummer>]] [*F*|*etched*]
//
// Alles außer dem Namen ist wahlfrei. `#` und `//` leiten einen Kommentar ein;
// Leerzeilen und Überschriften wie `Deck` oder `Sideboard` werden übergangen.
//
// **Es ist eine Produktentscheidung und keine Portierung.** Wo gesucht wurde,
// steht im Kopf von `liste_test.go` (§15): MSE2 kennt keine Sammlung, und
// `magic-set-generator` hat dazu ein TODO.

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Zeile ist eine gelesene Kartenzeile.
//
// **Ohne Druckkennung.** Die entsteht erst bei der Zuordnung; was hier steht,
// ist, was jemand geschrieben hat.
type Zeile struct {
	Menge int

	Name string

	// Setcode und Nummer sind der Hinweis auf einen **bestimmten** Druck —
	// beide wahlfrei.
	//
	// **Kleingeschrieben wie bei [domain.Exemplar]**: Scryfall führt Setcodes
	// klein, und `M11` neben `m11` wäre zweimal dasselbe Set.
	Setcode string
	Nummer  string

	// Finish ist die Ausführung — **leer heißt `nonfoil`** (T2, #1467).
	//
	// Wer „etched" schreibt, bekommt etched.
	Finish domain.Finish

	// Zeilennummer ist die Stelle in der Eingabe — ab 1.
	//
	// **Die der Datei und nicht die des Ergebnisses.** Ein Befund, der auf die
	// dritte *gelesene* Zeile zeigt, zeigt auf nichts, was jemand findet (§8).
	Zeilennummer int

	// Abschnitt ist die Überschrift, unter der diese Zeile steht —
	// kleingeschrieben, leer vor der ersten (T8, #1473).
	//
	// **Roh und nicht gedeutet.** Dieses Paket weiß nicht, wofür gelesen
	// wird: eine Liste für die **Sammlung** hat keine Bereiche, ein **Deck**
	// hat zwei — und was aus `maybeboard` wird, entscheidet der Aufrufer.
	//
	// **Hier fiel sie vorher weg.** `Lies` übergeht eine Überschrift, und
	// damit war eine Deckliste nach dem Lesen ein Stapel ohne Sideboard.
	// Eine Angabe zu erkennen und wegzuwerfen ist schlimmer, als sie nicht
	// zu kennen — derselbe Satz wie bei [Zeile.Finish].
	Abschnitt string

	// Roh ist die Zeile, wie sie dastand — geglättet.
	//
	// **Für die Rückmeldung.** Wer eine Zeile nicht zuordnen kann, will sie
	// sehen, nicht ihre Zerlegung.
	Roh string
}

// mengenmuster findet eine führende Menge — `4`, `4x` oder `4 x`.
//
// **`(?i)` für das `x`**, weil beide Schreibweisen vorkommen. Das Leerzeichen
// dahinter ist Pflicht: ohne es wäre `4x` in `4x4 Truck` eine Menge.
var mengenmuster = regexp.MustCompile(`^(\d+)\s*[xX]?\s+`)

// foilmuster findet die Ausführungsmarke am Ende — `*F*`, `*foil*`,
// `(foil)`, `*etched*`.
//
// **Die Marke wird gelesen und nicht nur erkannt**: bis `T2` traf dieses
// Muster `etched` bereits und machte daraus `Foil = true`. Eine Angabe zu
// erkennen und wegzuwerfen ist schlimmer, als sie nicht zu kennen — der
// Bestand sah aus, als hätte niemand sie gemacht.
var foilmuster = regexp.MustCompile(`(?i)\s*[*(](f|foil|etched)[*)]\s*$`)

// druckmuster findet Setcode und Nummer am Ende — `(M11) 149` oder `[M11]`.
//
// **Der Setcode ist drei bis fünf Zeichen.** Scryfall führt drei bis fünf;
// eine offene Länge machte aus `(the best one)` einen Setcode.
var druckmuster = regexp.MustCompile(`\s*[([]([A-Za-z0-9]{3,5})[)\]](?:\s+([A-Za-z0-9\-★]+))?\s*$`)

// ueberschriften sind Zeilen, die eine Liste gliedern statt eine Karte zu nennen.
//
// **Übergangen und nicht bemängelt.** Wer eine Deckliste einfügt, bringt sie
// mit; drei Befunde darüber liest niemand.
var ueberschriften = map[string]bool{
	"deck": true, "sideboard": true, "maybeboard": true, "commander": true,
	"companion": true, "tokens": true, "about": true,
}

// Lies zerteilt eine Liste in Zeilen.
//
// **Es liest so weit weiter, wie es verantwortbar ist** (§8): eine Zeile, die
// nicht aufgeht, wird zum Befund, und der Rest wird gelesen.
func Lies(text string) ([]Zeile, []domain.Befund) {
	zeilen := []Zeile{}
	befunde := []domain.Befund{}

	abschnitt := ""

	for nummer, roh := range strings.Split(text, "\n") {
		geglaettet := strings.TrimSpace(roh)

		if kopf, ist := ueberschrift(geglaettet); ist {
			abschnitt = kopf

			continue
		}

		if uebergehen(geglaettet) {
			continue
		}

		gelesen, fehler := zeileLesen(geglaettet)
		if fehler != nil {
			befunde = append(befunde, domain.Befund{
				Ort:     fmt.Sprintf("Zeile %d", nummer+1),
				Meldung: fehler.Error(),
				Art:     domain.ArtEingabe,
			})

			continue
		}

		gelesen.Zeilennummer = nummer + 1
		gelesen.Roh = geglaettet
		gelesen.Abschnitt = abschnitt
		zeilen = append(zeilen, gelesen)
	}

	return zeilen, befunde
}

// uebergehen sagt, ob eine Zeile keine Karte meint.
func uebergehen(zeile string) bool {
	return zeile == "" || strings.HasPrefix(zeile, "#") ||
		strings.HasPrefix(zeile, "//")
}

// ueberschrift sagt, ob eine Zeile gliedert — und wie sie heißt.
//
// **Getrennt von [uebergehen]**, weil sie zwei verschiedene Dinge sagt: eine
// Leerzeile trägt nichts, eine Überschrift trägt alles, was danach kommt.
// Solange beides dieselbe Antwort gab, fiel die Gliederung heraus.
func ueberschrift(zeile string) (string, bool) {
	klein := strings.ToLower(zeile)

	return klein, ueberschriften[klein]
}

// zeileLesen zerlegt eine einzelne Zeile.
func zeileLesen(zeile string) (Zeile, error) {
	gelesen := Zeile{Menge: 1}

	rest := zeile

	// **Die Menge zuerst**, damit `1 Black Lotus` nicht als Name `1 Black
	// Lotus` gelesen wird.
	if treffer := mengenmuster.FindStringSubmatch(rest); treffer != nil {
		menge, fehler := strconv.Atoi(treffer[1])
		if fehler != nil || menge < 1 {
			return Zeile{}, fmt.Errorf(
				"%w: %q ist keine Stückzahl — eine Zeile ohne Karte ist keine "+
					"Zeile", domain.ErrValidation, treffer[1])
		}

		gelesen.Menge = menge
		rest = rest[len(treffer[0]):]
	}

	// **Die Ausführung vor dem Druck**, weil die Marke hinter der Nummer
	// stehen darf.
	if treffer := foilmuster.FindStringSubmatch(rest); treffer != nil {
		gelesen.Finish = domain.FinishFoil
		if strings.EqualFold(treffer[1], string(domain.FinishEtched)) {
			gelesen.Finish = domain.FinishEtched
		}

		rest = foilmuster.ReplaceAllString(rest, "")
	}

	if treffer := druckmuster.FindStringSubmatch(rest); treffer != nil {
		gelesen.Setcode = strings.ToLower(treffer[1])
		gelesen.Nummer = treffer[2]
		rest = rest[:len(rest)-len(treffer[0])]
	}

	gelesen.Name = strings.TrimSpace(rest)
	if gelesen.Name == "" {
		return Zeile{}, fmt.Errorf(
			"%w: hier steht keine Karte — eine Zeile braucht einen Namen",
			domain.ErrValidation)
	}

	// **Eine Zeile aus lauter Ziffern ist keine Karte.**
	//
	// `7` allein ist eine verrutschte Stückzahl, kein Name: die Mengenmarke
	// verlangt ein Leerzeichen dahinter, und ohne das bliebe `7` als Name
	// stehen. Namen **mit** Ziffern gibt es („1996 World Champion", „Kaya,
	// Ghost Assassin"), und die fallen hier nicht auf — geprüft wird, ob
	// **nichts als** Ziffern dasteht.
	if nurZiffern(gelesen.Name) {
		return Zeile{}, fmt.Errorf(
			"%w: %q ist eine Zahl und kein Kartenname — fehlt dahinter die Karte?",
			domain.ErrValidation, gelesen.Name)
	}

	return gelesen, nil
}

// nurZiffern sagt, ob eine Zeichenkette aus nichts als Ziffern besteht.
func nurZiffern(text string) bool {
	for _, zeichen := range text {
		if zeichen < '0' || zeichen > '9' {
			return false
		}
	}

	return text != ""
}
