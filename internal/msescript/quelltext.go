package msescript

import "strings"

// AlsQuelltext gibt einen Wert als Quelltext — `toCode`
// (#1070).
//
// # Wofür
//
// MSEs Kommandozeile zeigt das Ergebnis eines Ausdrucks damit und nicht mit
// `toString()` (`src/cli/cli_main.cpp`, Zeile 234, Fassung f9d9356 vom
// 2021-07-06). Der Unterschied ist der Punkt: `"a"` ist eine Zeichenkette,
// `a` wäre ein Name — und wer eine Konsole benutzt, will das unterscheiden.
//
// # Die drei Fälle
//
// `ScriptValue::toCode` gibt in der Grundform `toString()`
// (`src/script/value.cpp`, Zeilen 42–44); **zwei** Arten überschreiben das:
// die Zeichenkette (Zeile 326) und `nil` (Zeile 453). Sammlungen haben ihre
// eigene Form (Zeilen 467–484).
func (w Wert) AlsQuelltext() string {
	switch w.Art {
	case ArtNichts:
		return "nil"

	case ArtTextwert:
		return inAnfuehrungszeichen(w.Text)

	case ArtAbbildung:
		return alsListe(w)

	default:
		return w.AlsText()
	}
}

// inAnfuehrungszeichen ist `quote_string` (`src/script/value.cpp`, Zeilen
// 303–317, Fassung f9d9356 vom 2021-07-06).
//
// **`<` bleibt, wie es ist.** MSE schützt dort `ESCAPED_LANGLE` — ein
// Sonderzeichen seines eigenen Auszeichnungsformats, das dieses Programm
// nicht kennt; ein gewöhnliches `<` schreibt auch MSE unverändert.
func inAnfuehrungszeichen(text string) string {
	var gebaut strings.Builder

	gebaut.WriteByte('"')

	for _, zeichen := range text {
		switch zeichen {
		case '"', '\\':
			gebaut.WriteByte('\\')
			gebaut.WriteRune(zeichen)

		case '\n':
			gebaut.WriteString(`\n`)

		case '\r':
			gebaut.WriteString(`\r`)

		case '\t':
			gebaut.WriteString(`\t`)

		default:
			gebaut.WriteRune(zeichen)
		}
	}

	gebaut.WriteByte('"')

	return gebaut.String()
}

// alsListe ist `ScriptCollectionBase::toCode`.
//
// **Der Schlüssel steht davor, wo es einen gibt.** MSE prüft den Index: `-1`
// heißt „kein Zähler, also ein Name". Dieses Programm führt Liste und
// Abbildung als **eine** Art (`ArtAbbildung`) und unterscheidet am Namen —
// ein Eintrag ohne Namen ist ein Listenglied.
func alsListe(w Wert) string {
	var gebaut strings.Builder

	gebaut.WriteByte('[')

	for stelle, eintrag := range w.Eintraege {
		if stelle > 0 {
			gebaut.WriteByte(',')
		}

		if eintrag.Name != "" {
			gebaut.WriteString(inAnfuehrungszeichen(eintrag.Name))
			gebaut.WriteByte(':')
		}

		gebaut.WriteString(eintrag.Wert.AlsQuelltext())
	}

	gebaut.WriteByte(']')

	return gebaut.String()
}
