package service

import (
	"regexp"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// rechtschreibmarke trifft MSE2s Unterringelung — `<error-spelling:en_US>`,
// samt Zusatzwörterbuch, öffnend und schließend.
//
// **Sie ist kein Wert, und deshalb wird sie nicht gelesen** (F05, #2247).
// MSE2 schreibt sie mit in die Datei, nimmt sie aber vor jeder neuen Prüfung
// heraus (`check_spelling`, `src/script/functions/spelling.cpp`, Zeilen
// 57–71, Fassung f9d9356 vom 2021-07-06), und die Vorlagen nehmen sie vor dem
// Zeichnen heraus. Hier blieb sie stehen — und der Volltext fand „spelling"
// und „magic" in jeder Karte, die MSE2 unterringelt hatte.
//
// **Portierungsmarke adapted**: MSE2 hebt sie bis zur nächsten Prüfung auf;
// dieses Programm prüft nicht (`msescript/rechtschreibung.go`), also gäbe es
// keine nächste. Ein MSE2, das die Abschrift öffnet, prüft neu und setzt sie
// wieder.
var rechtschreibmarke = regexp.MustCompile(`</?error-spelling(?::[^<>]*)?>`)

// wertText gibt den Wert eines Feldes — einzeilig oder als eingerückter Block.
//
// # Warum beide Formen
//
// In einer MSE-Datei steht ein langer Text unter seinem Schlüssel:
//
//	rule_text:
//		Erste Zeile
//		Zweite Zeile
//
// Der Wert hinter dem Doppelpunkt ist dann leer. Wer nur ihn liest, bekommt ein
// leeres Feld — und weil ein Textfeld ohne Wert gar nicht gezeichnet wird, ohne
// Befund. Auf einer gewöhnlichen Karte trifft das Regeltext **und** Flavortext
// (#398).
//
// **[msepaket.Knoten.Blocktext] und nicht `Skripttext`** (#2298): ein Wert
// behält seine tiefere Einrückung, wie MSEs `Reader::getValue` sie behält.
// `extra_data:` trägt zwei Ebenen, und bis dahin kam die zweite flach zurück —
// beim Sichern stand ein Feld dann neben seiner Stilvorlage statt darunter.
func wertText(knoten *msepaket.Knoten) string {
	if knoten.Wert != "" {
		return knoten.Wert
	}

	return knoten.Blocktext()
}

// wieInDerQuelle sagt, ob ein Wert des Bestands dem entspricht, was die Zeile
// der Quelldatei schon trägt.
//
// **Bei `extra_data:` auch in der flachen Lesart von früher** (#2298): bis
// dahin verlor ein Block beim Lesen seine tiefere Einrückung
// ([msepaket.Knoten.Skripttext]). Ein Bestand, der so eingelesen wurde,
// trägt diese Fassung, und bei den Zusatzdaten ist sie kein Wert, sondern
// ein Lesefehler — MSE2 läse die Vorlage als leer. Sie über die Datei zu
// schreiben, verdürbe, was dort richtig steht.
//
// **Nur dort** (Befund 1 des Zweitblicks über E/F): bei einem Text ist die
// flache Form ein Wert, den jemand gesetzt haben kann, und wer ihn im
// Formular sieht, sieht ihn flach. Die Gleichheit mit der alten Lesart
// unterdrückte sonst eine absichtliche Änderung.
func wieInDerQuelle(wert string, knoten *msepaket.Knoten) bool {
	if wert == wertText(knoten) {
		return true
	}

	return msepaket.Kanonisch(knoten.Schluessel) == zusatzdatenfeld &&
		knoten.Wert == "" && wert == knoten.Skripttext()
}

// rohtext gibt einen unverstandenen Knoten samt Kindern **wörtlich** zurück.
//
// Aus den Originalzeilen, nicht aus Schlüssel und Wert zusammengesetzt. Der
// Zweitblick über Phase 2 hat gezeigt, was der Unterschied kostet: eine
// Rekonstruktion normalisiert Abstände, setzt überall `\n` und verliert den
// fehlenden letzten Umbruch. Sie sieht aus wie das Original und ist es nicht —
// und genau das würde ein Byte-für-Byte-Vergleich finden, nicht vorher.
func rohtext(knoten *msepaket.Knoten) string {
	var bau strings.Builder

	schreibe(&bau, knoten)

	return bau.String()
}

func schreibe(bau *strings.Builder, knoten *msepaket.Knoten) {
	bau.WriteString(knoten.Roh)
	bau.WriteString(knoten.Zeilenende)

	for _, kind := range knoten.Kinder {
		schreibe(bau, kind)
	}
}
