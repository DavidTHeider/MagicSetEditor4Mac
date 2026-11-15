// Symbole und Erinnerungstext im Regeltext.
//
// **Gemeldet an der laufenden Anwendung**: `Birds of Paradise` zeigte
// `{T}: Add one mana of any color.` — den rohen Text von Scryfall statt des
// Tap-Symbols.

package scryfall

import "strings"

// MitSymbolen macht aus Scryfalls Klammern MSEs Symbolmarken.
//
// # Was MSE wirklich schreibt
//
// **Nachgelesen am 2026-08-24** in `Green.mse-set` (MSE 2.1.2, kanonischer
// Satz, `stylesheet: m15-altered`):
//
//	rule_text:
//		Other Elves you control get +1/+1.
//		<sym>T</sym>: Add <sym>G</sym><sym>G</sym><sym>G</sym>.
//
// Ein `<sym>` **je Symbol**. Scryfall schreibt dieselbe Stelle als `{T}` und
// `{G}{G}{G}`.
//
// # Und der Erinnerungstext
//
// **`<i-auto>…</i-auto>`**, in derselben Datei:
//
//	Crew 3 <i-auto>(Tap any number of creatures you control …)</i-auto>
//
// Der Zusatz `auto` sagt: von MSE erzeugt, nicht von Hand gesetzt. Scryfall
// liefert ihn als gewöhnliche Klammer im `oracle_text`.
//
// # Was sie nicht tut
//
// **Sie prüft nicht, ob ein Symbol gezeichnet werden kann.** `{Zauberstab}`
// wird zu `<sym>Zauberstab</sym>`; ob es eine Glyphe dafür gibt, entscheidet
// die Symbolschrift der Vorlage — und die kennt dieses Paket nicht. Dieselbe
// Haltung wie bei [domain.ManakostenOhneKlammern].
//
// **Und sie erkennt keine kursive Auszeichnung außer der Klammer.** MSE kennt
// `<i>` für von Hand gesetzte Kursive; woran man sie in einem Oracle-Text
// erkennen sollte, sagt Scryfall nicht.
func MitSymbolen(roh string) string {
	if roh == "" {
		return ""
	}

	return klammernKursiv(symbolmarken(roh))
}

// symbolmarken ersetzt `{X}` durch `<sym>X</sym>`.
//
// **Eine unpaarige Klammer bleibt stehen.** `Kostet {2 mehr` ist kein Symbol,
// und wer daraus eines machte, schriebe eine Marke, die nie schließt.
func symbolmarken(roh string) string {
	var bau strings.Builder

	rest := roh

	for {
		auf := strings.IndexByte(rest, '{')
		if auf < 0 {
			bau.WriteString(rest)

			return bau.String()
		}

		zu := strings.IndexByte(rest[auf:], '}')
		if zu < 0 {
			bau.WriteString(rest)

			return bau.String()
		}

		bau.WriteString(rest[:auf])
		bau.WriteString("<sym>")
		bau.WriteString(rest[auf+1 : auf+zu])
		bau.WriteString("</sym>")

		rest = rest[auf+zu+1:]
	}
}

// klammernKursiv setzt geklammerte Abschnitte in `<i-auto>`.
//
// **Nur ganze Klammerpaare in einer Zeile.** Ein Erinnerungstext steht in
// MSEs Sätzen immer geschlossen; eine offene Klammer über einen Zeilenumbruch
// hinweg wäre eine Vermutung, und Vermutungen erfindet dieses Paket nicht.
func klammernKursiv(text string) string {
	var bau strings.Builder

	rest := text

	for {
		auf := strings.IndexByte(rest, '(')
		if auf < 0 {
			bau.WriteString(rest)

			return bau.String()
		}

		zeile := strings.IndexByte(rest[auf:], '\n')
		zu := strings.IndexByte(rest[auf:], ')')

		if zu < 0 || (zeile >= 0 && zeile < zu) {
			bau.WriteString(rest[:auf+1])

			rest = rest[auf+1:]

			continue
		}

		bau.WriteString(rest[:auf])
		bau.WriteString("<i-auto>")
		bau.WriteString(rest[auf : auf+zu+1])
		bau.WriteString("</i-auto>")

		rest = rest[auf+zu+1:]
	}
}
