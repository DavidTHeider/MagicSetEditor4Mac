// **Der Text, den man sieht** — Werkstück U5 von Phase 5.1
// (#1416).
//
// Ein Kartenfeld trägt Auszeichnungen (`<b>`, `<sym>`, `<sep-soft>`); was ein
// Mensch liest und wonach er sucht, ist der Text **ohne** sie. MSE vergleicht
// den Schnellfilter mit `TextValue::toString()`, und das ist
// `untag_hide_sep(value)` (`src/data/field/text.cpp:171–173`, Fassung
// f9d9356 vom 2021-07-06).
//
// **Hier und nicht in `service`**, weil der Schnellfilter in der Domäne
// entscheidet, ob eine Karte passt — und `service` diese Domäne kennt, nicht
// umgekehrt. `service.ohneMarken` und `service.entschuetzt` reichen seit
// diesem Werkstück hierher durch.

package domain

import "strings"

// Entschuetzt macht aus den zwei geschützten Zeichen wieder das, was der
// Mensch geschrieben hat: U+0001 ist ein `<`, das keine Marke beginnt, und
// U+EB00 ein Leerzeichen, das nicht umbrechen soll.
//
// Portiert aus `untag_char`, `src/util/tagged_string.hpp:23`, Fassung f9d9356.
func Entschuetzt(text string) string {
	return schutzloeser.Replace(text)
}

// schutzloeser hält die beiden Ersetzungen — **einmal gebaut**, denn er läuft
// über jeden Wert jeder Karte.
var schutzloeser = strings.NewReplacer("\u0001", "<", "\ueb00", " ")

// OhneMarken nimmt die Auszeichnungen aus einem Stück Kartentext.
//
// Portiert aus `untag`, `src/util/tagged_string.cpp:33–46`, Fassung f9d9356.
// **Jedes `<` beginnt eine Marke**, und was bis zum `>` steht, fällt weg.
func OhneMarken(text string) string {
	var klar strings.Builder

	inMarke := false

	for _, zeichen := range text {
		switch {
		case zeichen == '<':
			inMarke = true
		case zeichen == '>':
			inMarke = false
		case !inMarke:
			klar.WriteRune(zeichen)
		}
	}

	return Entschuetzt(klar.String())
}

// weicherTrenner ist die Marke, deren Inhalt beim Lesen verborgen ist —
// ein Trennzeichen, das MSE nur zum Umbrechen einsetzt.
const weicherTrenner = "<sep-soft"

// Sichttext ist der Text ohne Marken und ohne den Inhalt weicher Trenner.
//
// Portiert aus `untag_hide_sep` und `remove_tag_contents`,
// `src/util/tagged_string.cpp:65–67` und `581–596`, Fassung f9d9356.
// **Fehlt die schließende Marke, endet der Text vor der öffnenden** — so
// macht es das Original, und ein halber Trenner ist kein Sichttext.
func Sichttext(text string) string {
	return OhneMarken(ohneTrennerinhalt(text))
}

func ohneTrennerinhalt(text string) string {
	anfang := strings.Index(text, weicherTrenner)
	if anfang < 0 {
		return text
	}

	var rest strings.Builder

	for anfang >= 0 {
		rest.WriteString(text[:anfang])

		ende := strings.Index(text[anfang:], "</sep-soft>")
		if ende < 0 {
			return rest.String()
		}

		text = text[anfang+ende+len("</sep-soft>"):]
		anfang = strings.Index(text, weicherTrenner)
	}

	rest.WriteString(text)

	return rest.String()
}
