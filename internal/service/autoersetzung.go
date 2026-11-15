package service

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Autoersetzer wendet die `auto replace:`-Regeln eines Spiels an — portiert
// aus `TextValueEditor::tryAutoReplace`, `src/gui/value/text.cpp`, Zeilen
// 1033–1052, Fassung f9d9356 vom 2021-07-06.
//
// # Warum in `service` und nicht in `msepaket`
//
// `msepaket` liest und schreibt Paketformate (§2). **Wann** eine Regel greift,
// ist keine Frage des Formats, sondern eine fachliche: MSE2 ersetzt beim
// Tippen, an der Schreibmarke, rückwärts — und nirgends sonst.
//
// # Was hier vorher stand, und warum es weg ist
//
// `msepaket.Ersetzer` lief über den **ganzen** Text, von links nach rechts,
// mit den Regeln nach Musterlänge geordnet. Es hatte **keinen Aufrufer**
// (#897), und
// es wich in drei Punkten von der Vorlage ab:
//
//	| | dort | MSE2 |
//	| Reichweite | ganzer Text | rückwärts von der Marke |
//	| Reihenfolge | längstes Muster zuerst | Reihenfolge der Datei |
//	| Wortgrenze | „links und rechts kein Wortzeichen" | eine **Oder**-Bedingung über Satzzeichen, Auszeichnungen überlesend |
//
// Der dritte Unterschied ist der teuerste: `(C)` mitten in einem Wort wurde
// dort nie ersetzt, in MSE2 aber dann, wenn **eine** Seite eine Grenze ist.
//
// **Ersetzt wird nur beim Tippen.** In MSE2 gibt genau ein Aufruf
// `allow_auto_replace = true` mit — `onChar`, ebenda Zeilen 497 und 499.
// Einfügen, Löschen, Rückschritt, Zeilenumbruch, Symbol einsetzen und die
// Rechtschreibkorrektur ersetzen **nicht**.
type Autoersetzer struct {
	regeln []msepaket.AutoErsetzung
}

// NeuerAutoersetzer baut den Ersetzer über den Regeln eines Spiels.
//
// **In der Reihenfolge der Datei**, ungeordnet: sie ist die einzige Auskunft
// darüber, was der Urheber zuerst gemeint hat, und MSE2 läuft sie so ab.
func NeuerAutoersetzer(regeln []msepaket.AutoErsetzung) *Autoersetzer {
	return &Autoersetzer{regeln: regeln}
}

// AnStelle wendet die Regeln an der Schreibmarke an und gibt Text und Marke
// zurück.
//
// Die Schleife läuft **einmal** durch alle Regeln. Eine spätere Regel sieht
// das Ergebnis einer früheren; eine frühere sieht das Ergebnis einer späteren
// nicht. So steht es in `tryAutoReplace`, und es ist der Grund, warum es nicht
// endlos laufen kann.
func (a *Autoersetzer) AnStelle(text string, stelle int) (ersetzt string, marke int) {
	if stelle < 0 || stelle > len(text) {
		return text, stelle
	}

	ende := stelle

	for _, regel := range a.regeln {
		if regel.Muster == "" || len(regel.Muster) > ende {
			continue
		}

		beginn := ende - len(regel.Muster)
		if text[beginn:ende] != regel.Muster {
			continue
		}

		if regel.GanzesWort && (!wortgrenze(text, beginn) || !wortgrenze(text, ende)) {
			continue
		}

		text = text[:beginn] + regel.Ersatz + text[ende:]
		ende = beginn + len(regel.Ersatz)
	}

	return text, ende
}

// grenzzeichen sind die Zeichen, an denen ein Wort endet.
//
// `word_bound_chars` plus `isPunct` — `src/gui/value/text.cpp`, Zeilen 1116
// bis 1120, Fassung f9d9356 vom 2021-07-06. `isPunct` ist dort `wxIspunct`;
// hier steht Gos `unicode.IsPunct` **und** `unicode.IsSymbol`, weil Go
// `+`, `=`, `<` und `$` als Symbole führt und C sie als Satzzeichen.
//
// **Das ist eine benannte Abweichung** (§15): ohne `IsSymbol` wäre `+` hier
// keine Wortgrenze und dort eine.
func grenzzeichen(zeichen rune) bool {
	return strings.ContainsRune(" ,.:;()\n", zeichen) ||
		unicode.IsPunct(zeichen) || unicode.IsSymbol(zeichen)
}

// wortgrenze sagt, ob an dieser Stelle ein Wort endet oder beginnt.
//
// **Eine Oder-Bedingung** — `TextValueEditor::isWordBoundary`, ebenda Zeilen
// 1138 bis 1165. Zuerst wird nach **vorn** gesehen: steht dort das Textende
// oder ein Grenzzeichen, ist es eine Grenze. Sonst nach **hinten**: Textanfang
// oder Grenzzeichen.
//
// **Auszeichnungen werden dabei überlesen**, in beide Richtungen. Ohne das
// stünde ein `<b>` zwischen Muster und Grenze und machte jede
// `whole word`-Regel im ausgezeichneten Text unwirksam.
func wortgrenze(text string, stelle int) bool {
	if nachVorn(text, stelle) {
		return true
	}

	return nachHinten(text, stelle)
}

// nachVorn sieht ab `stelle` nach rechts, Auszeichnungen überlesend.
func nachVorn(text string, stelle int) bool {
	for {
		if stelle >= len(text) {
			return true
		}

		if text[stelle] != '<' {
			zeichen, _ := utf8.DecodeRuneInString(text[stelle:])

			return grenzzeichen(zeichen)
		}

		zu := strings.IndexByte(text[stelle:], '>')
		if zu < 0 {
			// Ein nicht geschlossenes `<` ist keine Auszeichnung. MSEs
			// `skip_tag` gibt dann das Textende — und das ist eine Grenze.
			return true
		}

		stelle += zu + 1
	}
}

// nachHinten sieht ab `stelle` nach links, Auszeichnungen überlesend.
func nachHinten(text string, stelle int) bool {
	for {
		if stelle == 0 {
			return true
		}

		if text[stelle-1] != '>' {
			zeichen, _ := utf8.DecodeLastRuneInString(text[:stelle])

			return grenzzeichen(zeichen)
		}

		auf := strings.LastIndexByte(text[:stelle-1], '<')
		if auf < 0 || strings.ContainsRune(text[auf+1:stelle-1], '>') {
			// **War keine Auszeichnung** — MSE2 gibt hier `false`, nicht
			// „weiter suchen": ein `>` ohne sein `<` ist ein Zeichen wie
			// jedes andere, und Satzzeichen ist es nicht.
			return false
		}

		stelle = auf
	}
}
