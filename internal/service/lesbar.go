package service

import (
	"strconv"
	"strings"
	"unicode"
)

// **Fremder Text in einer Meldung wird lesbar gemacht**
// (#675).
//
// # Warum es das gibt
//
// Ein Befund nennt oft, **woran** er hängt: ein Zeichen, für das die
// Symbolschrift nichts führt, ein Stück Text, das stehen bleibt. Roh
// eingesetzt zerreißt so ein Stück die Meldung — über den ganzen Kanon stand
//
//	die Symbolschrift magic-mana-small führt kein Symbol für
//	; es wird übersprungen
//
// weil das Zeichen ein **Zeilenumbruch** war. Wer das liest, sieht eine
// abgeschnittene Meldung und keinen Grund.
//
// # Was sie tut
//
// Sie setzt den Text in Anführungszeichen und schreibt Steuerzeichen aus.
// Nicht `strconv.Quote`: das schriebe `ä` für ein Ä und machte aus einer
// deutschen Meldung eine Bytefolge.
//
// **Und sie kürzt.** Ein Feldwert kann tausend Zeichen haben; eine Meldung,
// die ihn ganz enthält, liest niemand.
func lesbar(text string) string {
	if text == "" {
		return "eine leere Stelle"
	}

	gekuerzt := []rune(text)
	if len(gekuerzt) > LaengeInMeldung {
		gekuerzt = append(gekuerzt[:LaengeInMeldung], '…')
	}

	var gebaut strings.Builder

	gebaut.WriteByte('"')

	for _, zeichen := range gekuerzt {
		switch {
		case zeichen == '"' || zeichen == '\\':
			gebaut.WriteByte('\\')
			gebaut.WriteRune(zeichen)

		case unicode.IsControl(zeichen):
			// **`QuoteRune` und dann die Hochkommas weg.** Es schreibt `\n`,
			// `\t` und für alles Übrige `\u2028` — dieselbe Schreibweise, die
			// jeder aus einer Zeichenkette kennt.
			gequotet := strconv.QuoteRune(zeichen)
			gebaut.WriteString(gequotet[1 : len(gequotet)-1])

		default:
			gebaut.WriteRune(zeichen)
		}
	}

	gebaut.WriteByte('"')

	return gebaut.String()
}

// LaengeInMeldung ist die Zahl der Zeichen, die aus fremdem Text in eine
// Meldung übernommen werden.
//
// **40, und die Zahl hat einen Grund**: eine Meldung soll in eine Zeile
// passen, und vor dem Text steht schon ein halber Satz. Wer mehr sehen will,
// sieht es am Feld — der Befund nennt es.
const LaengeInMeldung = 40
