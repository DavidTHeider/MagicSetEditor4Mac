package svg

import (
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// zahl schreibt eine Länge ohne überflüssige Nullen.
func zahl(wert float64) string {
	return strings.TrimSuffix(strings.TrimRight(
		fmt.Sprintf("%.3f", wert), "0"), ".")
}

// alsFarbe schreibt eine Farbe als `rgba(...)`.
//
// **Mit Alpha und nicht als `#rrggbb`**: eine Ebene kann durchsichtig sein, und
// ein Format, das das verschweigt, malt sie voll.
func alsFarbe(farbe render.Farbe) string {
	return fmt.Sprintf("rgba(%d,%d,%d,%s)",
		farbe.R, farbe.G, farbe.B, zahl(float64(farbe.A)/255))
}

// entschaerfe macht aus einem Text einen, der in XML stehen darf.
//
// # Fünf Zeichen und nicht drei
//
// Ein Kartenname enthält Anführungszeichen — „Ach! Hans, run!" ist eine echte
// Magic-Karte —, und dieser Text steht hier sowohl zwischen Marken als auch
// **in einem Attribut**.
//
// # Und Steuerzeichen, seit #852
//
// Der Zweitblick über Phase P hat angestrichen, dass „gültiges XML" mehr ist
// als fünf ersetzte Satzzeichen: **XML 1.0 verbietet die meisten
// Steuerzeichen** überhaupt, auch als Zahlenverweis (§2.2 der Empfehlung).
// Erlaubt sind unter U+0020 nur Tabulator, Zeilenvorschub und Wagenrücklauf.
//
// Ein Kartenfeld kommt aus einer fremden Datei von 2007 (§8) — ein `\x1b`
// darin machte die ganze SVG unlesbar, statt eine Stelle zu verlieren.
// **Ersetzt wird deshalb durch nichts**: ein Ersatzzeichen wäre eine Aussage
// über das Aussehen, die niemand getroffen hat.
func entschaerfe(text string) string {
	ersetzt := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&apos;",
	).Replace(text)

	return strings.Map(func(zeichen rune) rune {
		if erlaubtInXML(zeichen) {
			return zeichen
		}

		return -1
	}, ersetzt)
}

// erlaubtInXML sagt, ob ein Zeichen in einem XML-1.0-Dokument stehen darf.
//
//	Char ::= #x9 | #xA | #xD | [#x20-#xD7FF] | [#xE000-#xFFFD] | [#x10000-#x10FFFF]
//
// XML 1.0, fünfte Ausgabe, §2.2.
func erlaubtInXML(zeichen rune) bool {
	switch {
	case zeichen == '\t' || zeichen == '\n' || zeichen == '\r':
		return true

	case zeichen < 0x20:
		return false

	// **Die beiden Nichtzeichen am Ende der BMP.**
	case zeichen == 0xFFFE || zeichen == 0xFFFF:
		return false
	}

	// **Kein Zweig für halbe Ersatzzeichen** (U+D800…U+DFFF), obwohl XML sie
	// ebenso verbietet: eine Go-Zeichenkette kann keines tragen.
	// [strings.Map] läuft über Runen, und ungültiges UTF-8 kommt als U+FFFD
	// an — das ist erlaubt. Ein Zweig dafür wäre einer, den kein Aufruf
	// betreten kann (§3); nachgemessen, nicht angenommen.

	// **DEL und die C1-Steuerzeichen bleiben.** XML 1.0 verbietet sie nicht;
	// erst XML 1.1 verlangt für sie einen Zahlenverweis. Wer sie hier
	// zusätzlich striche, entfernte Zeichen, die ein Betrachter anstandslos
	// liest — und die Regel stünde weiter als ihre Quelle.
	return true
}
