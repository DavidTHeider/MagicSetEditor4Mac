package msescript

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/dlclark/regexp2"
)

// Ersetze setzt an jeden Treffer den Ersatztext — in der Fassung von sed.
//
// # Herkunft
//
// `RegexReplacer::apply` in `src/script/functions/regex.cpp` von MSE2 ruft
// `results.format(replacement_string)`, und `Regex::Results::format` in
// `src/util/regex.hpp` Zeile 63 gibt `boost::format_sed` mit. Damit gilt:
//
//	\1 … \9   die Gruppe dieser Nummer
//	&         der ganze Treffer
//	\&        ein kaufmännisches Und
//	\\        ein Rückstrich
//	sonst     wörtlich
//
// # Warum das ein echter Befund war
//
// Bis #240
// ersetzte dieses Programm **wörtlich**: aus `replace: "\\1"` wurde die
// Zeichenfolge `\1` im Kartentext. Ausgezählt am 2026-08-16 stehen in
// `M15-Magic-Pack` **33** verschiedene Ersatztexte mit Rückverweis — darunter
// der Umbau der Erinnerungstexte, `\2<blob>\4\3<moved>\1</moved></blob>`.
//
// # Warum nicht Gos eigene Ersetzung
//
// `Regexp.ReplaceAllString` kennt `$1`, nicht `\1` — und ein `$` im Ersatztext
// wäre dann kein Dollarzeichen mehr. In den Vorlagen steht es als Text. Der
// erste Anlauf schrieb deshalb `ReplaceAllLiteralString`, und das war die
// andere Hälfte des Fehlers: er nahm auch `\1` wörtlich.
func (m *Muster) Ersetze(text, vorlage string) (string, error) {
	if m.kontext != nil {
		return m.ersetzeImKontext(text, vorlage)
	}

	if m.schnell != nil {
		return m.ersetzeSchnell(text, vorlage), nil
	}

	var gebaut strings.Builder

	zeichen := []rune(text)
	gelesen := 0

	err := m.jederTreffer(text, func(treffer *regexp2.Match) {
		gebaut.WriteString(string(zeichen[gelesen:treffer.Index]))
		gebaut.WriteString(fuelle(vorlage, func(nummer int) string {
			gruppe := treffer.GroupByNumber(nummer)
			if gruppe == nil {
				return ""
			}

			return gruppe.String()
		}))

		gelesen = treffer.Index + treffer.Length
	})
	if err != nil {
		return "", err
	}

	gebaut.WriteString(string(zeichen[gelesen:]))

	return gebaut.String(), nil
}

func (m *Muster) ersetzeSchnell(text, vorlage string) string {
	var gebaut strings.Builder

	gelesen := 0

	for _, stellen := range m.schnell.FindAllStringSubmatchIndex(text, -1) {
		gebaut.WriteString(text[gelesen:stellen[0]])
		gebaut.WriteString(fuelle(vorlage, func(nummer int) string {
			// **Eine Gruppe, die nicht getroffen hat, ist leer statt ein
			// Fehler.** `a(b)?c` gegen `ac` lässt Gruppe 1 offen; sed setzt
			// dort nichts ein, und die Vorlagen rechnen damit.
			if 2*nummer+1 >= len(stellen) || stellen[2*nummer] < 0 {
				return ""
			}

			return text[stellen[2*nummer]:stellen[2*nummer+1]]
		}))

		gelesen = stellen[1]
	}

	gebaut.WriteString(text[gelesen:])

	return gebaut.String()
}

// fuelle setzt die Rückverweise einer sed-Vorlage ein.
//
// `gruppe(0)` ist der ganze Treffer — dieselbe Zählung wie in sed und in beiden
// Maschinen, und deshalb steht `&` hier nicht als Sonderfall daneben.
func fuelle(vorlage string, gruppe func(int) string) string {
	var gebaut strings.Builder

	zeichen := []rune(vorlage)

	for stelle := 0; stelle < len(zeichen); stelle++ {
		if zeichen[stelle] == '&' {
			gebaut.WriteString(gruppe(0))

			continue
		}

		// **Ein Rückstrich am Ende bleibt ein Rückstrich.** Ein Ersatztext, der
		// so endet, ist in keiner Vorlage gemeint — aber er wäre der Griff
		// hinter das Ende der Zeichenkette, und fremde Daten sind der Normalfall.
		if zeichen[stelle] != '\\' || stelle+1 == len(zeichen) {
			gebaut.WriteRune(zeichen[stelle])

			continue
		}

		stelle++

		if naechstes := zeichen[stelle]; naechstes >= '0' && naechstes <= '9' {
			gebaut.WriteString(gruppe(int(naechstes - '0')))

			continue
		}

		gesetzt, weiter := fluchtzeichen(zeichen, stelle)
		gebaut.WriteRune(gesetzt)
		stelle = weiter
	}

	return gebaut.String()
}

// einfacheFluchten sind die Steuerzeichen, die `boost::format_sed` kennt.
var einfacheFluchten = map[rune]rune{
	'a': '\a', 'e': 0x1b, 'f': '\f', 'n': '\n', 'r': '\r', 't': '\t', 'v': '\v',
}

// fluchtzeichen deutet, was hinter einem Rückstrich steht — ohne die Gruppen,
// die [fuelle] schon abgefangen hat.
//
// Es gibt das gesetzte Zeichen und die Stelle, an der weiterzulesen ist.
//
// # Herkunft
//
// Die sed-Fassung von Boost führt neben den Gruppen und `&` ausdrücklich
// `\a \e \f \n \r \t \v`, `\xDD`, `\x{DDDD}` und `\cX`; ein Rückstrich vor
// irgendetwas anderem gibt dieses Zeichen wörtlich.
//
// **Ein unvollständiges `\x` fällt auf den Auffangfall zurück.** Fremde Daten
// sind der Normalfall, und ein Ersatztext, der auf `\x` endet, darf nicht
// hinter das Ende der Zeichenkette greifen.
func fluchtzeichen(zeichen []rune, stelle int) (gesetzt rune, weiter int) {
	if bekanntes, bekannt := einfacheFluchten[zeichen[stelle]]; bekannt {
		return bekanntes, stelle
	}

	if zeichen[stelle] == 'c' && stelle+1 < len(zeichen) {
		// `\cA` ist das Steuerzeichen 1. Die unteren fünf Bit des Buchstabens,
		// gleich ob groß oder klein geschrieben.
		return zeichen[stelle+1] & 0x1f, stelle + 1
	}

	if zeichen[stelle] == 'x' {
		if wert, bis, gut := hexadezimal(zeichen, stelle+1); gut {
			return wert, bis
		}
	}

	return zeichen[stelle], stelle
}

// hexadezimal liest `41` oder `{20AC}` hinter einem `\x`.
//
// Es gibt das Zeichen, die letzte gelesene Stelle und ob die Angabe taugte.
func hexadezimal(zeichen []rune, stelle int) (gelesen rune, bis int, gut bool) {
	ende := stelle + 2
	geklammert := stelle < len(zeichen) && zeichen[stelle] == '{'

	if geklammert {
		stelle++

		for ende = stelle; ende < len(zeichen) && zeichen[ende] != '}'; ende++ {
		}

		// Ohne schließende Klammer ist es keine Angabe, sondern Text.
		if ende == len(zeichen) {
			return 0, 0, false
		}
	}

	if ende > len(zeichen) || ende == stelle {
		return 0, 0, false
	}

	// **Über `strconv` und nicht über eine eigene Ziffernschleife.** Der erste
	// Anlauf rechnete von Hand und lief in `G115`: aus einem `int` wurde ohne
	// Prüfung ein `rune`. Ein Ersatztext aus einem bezogenen Paket ist fremde
	// Eingabe, und `\x{FFFFFFFF}` liefe dort über.
	wert, err := strconv.ParseUint(string(zeichen[stelle:ende]), 16, 21)
	if err != nil || wert > utf8.MaxRune {
		return 0, 0, false
	}

	if geklammert {
		// Die schließende Klammer mitnehmen.
		return rune(wert), ende, true
	}

	return rune(wert), ende - 1, true
}
