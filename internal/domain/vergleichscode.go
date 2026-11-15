package domain

// **Was zwei Menschen laut miteinander vergleichen** — NET-01
// (#1705).
//
// # Warum es den Fingerabdruck daneben noch gibt
//
// **Weil er eine andere Frage beantwortet.** Der Fingerabdruck gehört einer
// Kennung; wer ihn vergleicht, liest zwei Werte vor — den eigenen und den,
// den er vom anderen sieht. Der Vergleichscode gehört einem **Paar**: beide
// Seiten sehen dieselben dreißig Ziffern, und ein einziger Vergleich genügt.
// Das ist das Verfahren „Code vergleichen" aus dem Issue; der Abdruck bleibt
// als Angabe für Fachleute.
//
// # Das Vorbild, und was davon übernommen ist
//
// **Die Sicherheitsnummern von Signal** — nachgelesen in
// `signalapp/libsignal`, `rust/protocol/src/fingerprint.rs`, Fassung 4ea5b76
// vom 2025-09-03. **Übernommen ist das Verfahren, kein Code**: libsignal
// steht unter AGPL-3.0-only und verträgt sich mit GPL-2.0-or-later nicht.
// Das Verfahren (Kennzeichnung **adapted**):
//
//   - **Je Seite eine Hälfte**, und sie hängt nur am Schlüssel dieser Seite.
//     SHA-512 über Fassung und Schlüssel, danach 5.200-mal über das
//     Zwischenergebnis und den Schlüssel — dieselbe Zahl wie dort.
//   - **Je fünf Bytes fünf Ziffern**, als Zahl gelesen und durch 100.000
//     geteilt; der Rest ist die Gruppe.
//   - **Die kleinere Hälfte zuerst** — so sehen beide Seiten dieselbe Folge.
//
// **Abweichend ist die Länge**: drei Gruppen je Seite statt sechs, also
// dreißig Ziffern statt sechzig. Das Issue verlangt einen Code, den Menschen
// laut vergleichen; sechzig Ziffern liest am Telefon niemand zu Ende.
//
// # Was das kostet
//
// **Wer sich zwischen zwei Menschen stellt**, muss für jede Seite einen
// eigenen Schlüssel finden, dessen Hälfte die eines **bestimmten** Gegenübers
// trifft: fünfzehn Ziffern sind rund 2^50 Möglichkeiten, jede davon ein neuer
// Schlüssel und 5.200 Hashes. Das reicht nicht gegen einen Staat mit
// Rechenzentrum, und es steht hier, damit niemand es behauptet.
//
// **Die getrennten Hälften sind Signals Gestalt und kein Gewinn an Stärke.**
// Ein Code über beide Schlüssel zugleich hätte dieselbe Größenordnung — dort
// genügte ein Geburtstag über dreißig Ziffern, also ebenfalls rund 2^50.
// Hier stand zuerst das Gegenteil; nachgerechnet beim Schreiben des Vermerks
// [0150](../../docs/entscheidungen/0150-einladen-und-vergleichen.md).

import (
	"crypto/sha512"
	"fmt"
	"strings"
)

// vergleichsfassung trennt diese Ableitung von jeder anderen über dieselben
// Bytes.
const vergleichsfassung = "mse4mac-vergleich/1"

// vergleichsrunden ist die Zahl der Hashrunden je Hälfte — die von Signal.
const vergleichsrunden = 5200

// vergleichsgruppen ist die Zahl der Fünfergruppen je Seite.
const vergleichsgruppen = 3

// Vergleichscode gibt die dreißig Ziffern, die zwei Seiten vergleichen.
func Vergleichscode(eine, andere string) string {
	a, b := vergleichshaelfte(eine), vergleichshaelfte(andere)
	if b < a {
		a, b = b, a
	}

	return a + " " + b
}

// VergleichscodePasst sagt, ob eine Eingabe der Code dieses Paars ist.
//
// **Leerzeichen und Striche zählen nicht** — wer die Ziffern abtippt, gliedert
// sie, wie er sie gehört hat.
func VergleichscodePasst(eine, andere, eingabe string) bool {
	return VergleichscodeGleich(Vergleichscode(eine, andere), eingabe)
}

// VergleichscodeGleich sagt, ob eine Eingabe ein schon gerechneter Code ist.
func VergleichscodeGleich(code, eingabe string) bool {
	nurZiffern := func(text string) string {
		return strings.Map(func(zeichen rune) rune {
			if zeichen == ' ' || zeichen == '-' {
				return -1
			}

			return zeichen
		}, text)
	}

	return eingabe != "" && nurZiffern(eingabe) == nurZiffern(code)
}

// vergleichshaelfte rechnet die Hälfte einer Seite.
func vergleichshaelfte(kennung string) string {
	schluessel := []byte(kennung)

	summe := sha512.Sum512(append([]byte(vergleichsfassung), schluessel...))
	for range vergleichsrunden - 1 {
		summe = sha512.Sum512(append(summe[:], schluessel...))
	}

	gruppen := make([]string, 0, vergleichsgruppen)

	for gruppe := range vergleichsgruppen {
		var zahl uint64
		for _, stelle := range summe[gruppe*5 : gruppe*5+5] {
			zahl = zahl<<8 | uint64(stelle)
		}

		gruppen = append(gruppen, fmt.Sprintf("%05d", zahl%100_000))
	}

	return strings.Join(gruppen, " ")
}
