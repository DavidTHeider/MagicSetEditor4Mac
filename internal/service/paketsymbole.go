// **Der Symbolindex eines Pakets** — Werkstück U7 von Phase 5.1
// (#1418).
//
// # Wofür
//
// Die Vorlagenwerkstatt soll zu einem Namen sagen, **wo er herkommt** — das
// Feld `casting cost` steht im Spiel, der Skriptname `kosten` im
// `init script` —, und wer ihn noch benutzt. Das Issue verlangt, dass Hover,
// Sprung zur Definition und Referenzen auf **belegte** Paket- und
// Skriptsymbole gestützt sind, nicht auf eine zweite Lesart des Formats in
// TypeScript. Deshalb steht der Index hier: er liest mit demselben Zerteiler,
// der das Paket lädt, und mit demselben, der die Skripte auswertet.
//
// # Was ein Symbol ist
//
// **Ein `name:` in einem definierenden Block** — `card field`, `set field`,
// `styling field`, `extra card field` — oder **eine Zuweisung auf oberster
// Ebene eines Skriptwerts** (`kosten := …`). Was in einem `card style:`
// steht, ist keine Definition, sondern eine Verwendung: der Stil richtet ein
// Feld ein, das das Spiel definiert hat.
//
// **Zuweisungen innerhalb einer Funktion zählen nicht.** `{ x := 1; x }`
// bindet `x` nur im Rumpf; wer es außen sucht, fände es nicht.
//
// # Woher die Regel kommt
//
// **new** — MSE2 hat keinen Vorlageneditor und keinen Index (gesucht in
// `src/gui/`, Fassung f9d9356 vom 2021-07-06: dort liegen Kartenliste,
// Editor und Konsole, kein Bauteil liest eine Vorlage als Text). Was ein Feld
// definiert, steht in `src/data/field.cpp` (`name`), und dass ein Skript ein
// Leerzeichen im Feldnamen als `_` schreibt, in `src/script/context.cpp`;
// beides ist hier nicht nachgebaut, sondern über [msepaket] und [msescript]
// gelesen.

package service

import (
	"sort"
	"strings"
	"unicode"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Paketdateitext ist eine bearbeitbare Datei eines Pakets samt Inhalt.
type Paketdateitext struct {
	Name   string
	Inhalt string
}

// Paketsymbol ist ein Name, den das Paket definiert, mit seiner Stelle.
type Paketsymbol struct {
	Name string

	// Art ist der definierende Block — `card field`, `set field`,
	// `styling field`, `extra card field` — oder `script` für eine Zuweisung.
	Art string

	Datei  string
	Zeile  int
	Spalte int
}

// Symbolverwendung ist eine Zeile, die einen Namen nennt.
type Symbolverwendung struct {
	Datei string
	Zeile int

	// Text ist die Zeile ohne ihre Einrückung — zum Zeigen, nicht zum Rechnen.
	Text string
}

// definierendeBloecke sind die Blöcke, deren `name:` etwas einführt —
// in der kanonischen Schreibweise, in der [msepaket.Kanonisch] vergleicht.
var definierendeBloecke = map[string]bool{
	"card_field":       true,
	"set_field":        true,
	"styling_field":    true,
	"extra_card_field": true,
}

// skriptart ist die Art einer Zuweisung im Index.
const skriptart = "script"

// Paketsymbole gibt jede Definition in den Dateien, je Datei nach Zeile.
func Paketsymbole(dateien []Paketdateitext) []Paketsymbol {
	symbole := []Paketsymbol{}

	for _, datei := range dateien {
		eigene := feldsymbole(datei)
		eigene = append(eigene, skriptsymbole(datei)...)

		sort.SliceStable(eigene, func(i, j int) bool { return eigene[i].Zeile < eigene[j].Zeile })

		symbole = append(symbole, eigene...)
	}

	return symbole
}

// feldsymbole liest die `name:`-Einträge der definierenden Blöcke.
//
// **Eine Datei, die sich nicht zerteilen lässt, trägt nichts bei** — und
// nimmt den anderen nichts (§8).
func feldsymbole(datei Paketdateitext) []Paketsymbol {
	baum, err := msepaket.Zerteile(datei.Name, []byte(datei.Inhalt))
	if err != nil {
		return nil
	}

	var symbole []Paketsymbol

	for _, block := range baum.Wurzel {
		kanonisch := msepaket.Kanonisch(block.Schluessel)
		if block.Zeilenart != msepaket.ArtEintrag || !definierendeBloecke[kanonisch] {
			continue
		}

		name := block.Ersten("name")
		if name == nil || strings.TrimSpace(name.Wert) == "" {
			continue
		}

		symbole = append(symbole, Paketsymbol{
			Name: strings.TrimSpace(name.Wert), Art: msepaket.Entkanonisch(kanonisch), Datei: datei.Name,
			Zeile: name.Stelle.Zeile, Spalte: name.Stelle.Spalte,
		})
	}

	return symbole
}

// skriptsymbole liest die Zuweisungen auf oberster Ebene jedes Skriptwerts.
//
// **Über dieselben Merkmale wie die Einfärbung und die Prüfung** — welche
// Schlüssel ein Skript tragen, weiß `msepaket` an einer Stelle; und die
// Versetzung von Skriptzeile auf Dateizeile ist dieselbe wie bei einem
// Fehler ([versetzt]).
func skriptsymbole(datei Paketdateitext) []Paketsymbol {
	var symbole []Paketsymbol

	for _, merkmal := range msepaket.Einfaerben(datei.Inhalt) {
		if merkmal.Art != msepaket.FarbeSkriptwert {
			continue
		}

		baum, err := msescript.Zerteile(merkmal.Text)
		if err != nil {
			continue
		}

		for _, glied := range obersteGlieder(baum) {
			if glied.Form != msescript.FormZuweisung {
				continue
			}

			zeile, spalte := versetzt(merkmal, glied.Zeile, glied.Spalte)

			symbole = append(symbole, Paketsymbol{
				Name: glied.Text, Art: skriptart, Datei: datei.Name, Zeile: zeile, Spalte: spalte,
			})
		}
	}

	return symbole
}

// obersteGlieder gibt die Anweisungen eines Skripts — eine oder die Folge.
func obersteGlieder(baum *msescript.Knoten) []*msescript.Knoten {
	if baum.Form == msescript.FormFolge {
		return baum.Kinder
	}

	return []*msescript.Knoten{baum}
}

// Symbolverwendungen gibt jede Zeile, die den Namen als ganzes Wort nennt —
// in der Schreibweise der Datei (`casting cost`) und der des Skripts
// (`casting_cost`). Die Zeile, die ihn definiert, zählt nicht.
func Symbolverwendungen(dateien []Paketdateitext, name string) []Symbolverwendung {
	if name == "" {
		return nil
	}

	definitionen := map[Symbolverwendung]bool{}

	for _, symbol := range Paketsymbole(dateien) {
		if symbol.Name == name {
			definitionen[Symbolverwendung{Datei: symbol.Datei, Zeile: symbol.Zeile}] = true
		}
	}

	var gefunden []Symbolverwendung

	for _, datei := range dateien {
		for nummer, zeile := range strings.Split(datei.Inhalt, "\n") {
			stelle := Symbolverwendung{Datei: datei.Name, Zeile: nummer + 1}
			if definitionen[stelle] || !nenntGanzesWort(zeile, name) {
				continue
			}

			stelle.Text = strings.TrimSpace(zeile)
			gefunden = append(gefunden, stelle)
		}
	}

	return gefunden
}

// nenntGanzesWort sagt, ob die Zeile den Namen in einer seiner beiden
// Schreibweisen als ganzes Wort enthält.
func nenntGanzesWort(zeile, name string) bool {
	for _, form := range []string{name, strings.ReplaceAll(name, " ", "_")} {
		if ganzesWort(zeile, form) {
			return true
		}
	}

	return false
}

// ganzesWort sucht jede Fundstelle und prüft ihre Ränder.
func ganzesWort(zeile, wort string) bool {
	zeichen := []rune(zeile)
	gesucht := []rune(wort)

	for von := 0; von+len(gesucht) <= len(zeichen); von++ {
		if string(zeichen[von:von+len(gesucht)]) != wort {
			continue
		}

		bis := von + len(gesucht)
		if (von == 0 || !wortzeichen(zeichen[von-1])) && (bis == len(zeichen) || !wortzeichen(zeichen[bis])) {
			return true
		}
	}

	return false
}

// wortzeichen sagt, ob ein Zeichen zu einem Namen gehören kann.
func wortzeichen(zeichen rune) bool {
	return zeichen == '_' || unicode.IsLetter(zeichen) || unicode.IsDigit(zeichen)
}
