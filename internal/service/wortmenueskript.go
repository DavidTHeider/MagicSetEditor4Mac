// Wörter aus einem Skript
// (#1074).
//
// **77 der 432 Wörter im Pack haben kein festes Wort, sondern ein Skript** —
// gemessen am 2026-08-24 am bezogenen `M15-Magic-Pack`. `all_sub_types(setting:
// "is_land")` sammelt die Untertypen aller Länder **dieses Sets**; die Liste
// wächst also mit dem Set und steht nirgends fest.
//
// Portiert aus `DropDownWordList::addWordsFromScript`, `src/gui/value/text.cpp`,
// Zeilen 181–213, Fassung f9d9356 vom 2021-07-06.

package service

import (
	"sort"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// ausSkript hängt die Wörter eines Skriptworts ans Menü.
//
// # Der Strich wandert
//
// `line below` steht am **Wort**, nicht an seinen Ergebnissen — und ein Strich
// unter jedem wäre kein Trenner mehr. MSE nimmt ihn deshalb dem Vorgänger
// wieder weg, sobald ein weiterer kommt (`items[prev].flags = 0`, Zeile 208):
// am Ende trägt ihn der letzte.
//
// **Der leere Eintrag ist ausgenommen** (`!items[prev].name.empty()`). Er
// entsteht aus dem führenden Komma, das `all_sub_types` schreibt, und bedeutet
// »nichts davon«; sein Strich trennt ihn von den Wörtern.
func ausSkript(
	eintraege []Wortmenueeintrag,
	wort msepaket.Wortlisteneintrag,
	auswerten func(string) string,
) []Wortmenueeintrag {
	vorheriger := len(eintraege)

	for _, name := range zerteileWoerter(auswerten(wort.Skript)) {
		if vorheriger < len(eintraege) && name == eintraege[vorheriger].Name {
			continue
		}

		if vorheriger < len(eintraege) && eintraege[vorheriger].Name != "" {
			eintraege[vorheriger].ZeileDarunter = false
		}

		vorheriger = len(eintraege)

		eintraege = append(eintraege, Wortmenueeintrag{
			Name:          name,
			IstPraefix:    wort.IstPraefix,
			ZeileDarunter: wort.ZeileDarunter,
		})
	}

	return eintraege
}

// zerteileWoerter zerlegt das Ergebnis eines Skripts an den Kommas und
// sortiert.
//
// # Das Leerzeichen nach dem Komma bleibt stehen
//
//	prev = comma + 1;
//	if (prev + 1 < str.size() && str.GetChar(prev + 1) == _(' ')) ++prev;
//
// `src/gui/value/text.cpp`, Zeilen 195–196, Fassung f9d9356 vom 2021-07-06.
// Der Kommentar daneben sagt „skip space after comma" — geprüft wird aber das
// Zeichen **hinter** dem ersten nach dem Komma. Bei `"a, b"` steht dort `b`,
// also passiert nichts, und der Eintrag heißt `" b"`.
//
// **Mitportiert, nicht behoben** (§15.4): der Name ist das, was ins Feld
// geschrieben wird. Wer hier ein Leerzeichen wegnähme, ließe dieselbe Karte in
// MSE2 und hier verschieden aussehen. Im echten Pack tritt der Fall nicht auf —
// `all_sub_types` trennt ohne Leerzeichen.
func zerteileWoerter(ergebnis string) []string {
	var woerter []string

	anfang := 0

	for {
		komma := strings.IndexByte(ergebnis[anfang:], ',')
		if komma < 0 {
			break
		}

		komma += anfang

		woerter = append(woerter, ergebnis[anfang:komma])

		anfang = komma + 1
		if anfang+1 < len(ergebnis) && ergebnis[anfang+1] == ' ' {
			anfang++
		}
	}

	woerter = append(woerter, ergebnis[anfang:])

	sort.Strings(woerter)

	return woerter
}
