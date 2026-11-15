package render

import "unicode"

// **Wie eine Ligatur durch einen Satz aus Zeichen kommt** —
// #1805.
//
// Dieser Rasterer misst, kernt und zeichnet **Zeichen**: [Schrift.Breite],
// [Kernschrift] und [Glyphenschrift.Maske] nehmen eine Rune. Eine Ligatur hat
// keine — `ffi` ist eine Glyphe für drei Zeichen. Sie reist deshalb als
// **Glyphenzeichen**: eine Rune jenseits von Unicode, deren Wert die
// Glyphennummer trägt. Aus gültigem Text entsteht keine solche Rune — Go
// dekodiert jede ungültige Folge zu U+FFFD —, und jede Stelle, die eine Rune
// misst oder zeichnet, bleibt, wie sie war.
//
// **Gesetzt wird je Zug**, den [gesetzt] gibt: die Vorsilbe, die Zeilenbreite,
// die Textbreite und der Zeichner nehmen ihn alle aus derselben Funktion.
// Zwei Rechnungen über dieselbe Sache laufen sonst auseinander, sobald eine
// von ihnen etwas lernt — der Satz steht bei [kernungZwischen], und er gilt
// hier genauso.

// Ligaturschrift zieht Zeichenfolgen zu Glyphen zusammen.
//
// **Eine eigene Schnittstelle, wie [Kernschrift]**: eine Schrift ohne
// Ligaturen soll nichts davon wissen müssen, und keine Attrappe der Tests
// braucht eine Methode, die ihr Fall nicht kennt.
type Ligaturschrift interface {
	// Ligiere gibt die Zeichen eines Zugs, wie die Schrift sie setzt — eine
	// Ligatur als **ein** Glyphenzeichen.
	Ligiere(runen []rune) []rune
}

// glyphenzeichenbasis ist die erste Rune jenseits von Unicode.
const glyphenzeichenbasis = unicode.MaxRune + 1

// glyphenzeichen gibt die Rune, die für eine Glyphe ohne eigenes Zeichen
// steht.
func glyphenzeichen(kennung uint16) rune { return glyphenzeichenbasis + rune(kennung) }

// glyphennummer gibt die Glyphe eines Glyphenzeichens — und ob es eines ist.
func glyphennummer(zeichen rune) (uint16, bool) {
	if zeichen < glyphenzeichenbasis || zeichen > glyphenzeichenbasis+0xFFFF {
		return 0, false
	}

	return uint16(zeichen - glyphenzeichenbasis), true //nolint:gosec // oben begrenzt
}

// gesetzt gibt die Zeichen eines Zugs, wie die Schrift sie setzt.
//
// **Mit Laufweite keine Ligatur.** Der Abstand steht zwischen den Buchstaben,
// und eine Ligatur hat keinen Zwischenraum; Satzprogramme lassen sie deshalb
// in gesperrtem Text weg. Im Betrieb ist die Laufweite null ([vorsilbe.nimm]).
func gesetzt(text string, satz Satz, schrift Schrift) []rune {
	runen := []rune(text)

	ligiert, kann := schrift.(Ligaturschrift)
	if !kann || satz.Laufweite != 0 {
		return runen
	}

	return ligiert.Ligiere(runen)
}

// Ligiere setzt die Ligaturen der Schrift ein — [Ligaturschrift].
//
// **Ein Zeichen, das die Schrift nicht führt, bleibt eines** und trennt:
// Glyphe null steht in keiner Deckung. Und **was nicht ersetzt wurde, bleibt
// seine Rune**, damit die Ersatzschrift und die Namen fehlender Zeichen es
// weiter erkennen.
func (s *OpenTypeSchrift) Ligiere(runen []rune) []rune {
	if len(s.liga.nachschlaege) == 0 || len(runen) == 0 {
		return runen
	}

	glyphen := make([]uint16, len(runen))

	s.schloss.Lock()
	for i, zeichen := range runen {
		kennung, _ := s.kennung(zeichen)
		glyphen[i] = uint16(kennung)
	}
	s.schloss.Unlock()

	aus, deckt := s.liga.anwenden(glyphen)

	gesetzt := make([]rune, 0, len(aus))
	stelle := 0

	for i, glyphe := range aus {
		if deckt[i] == 1 && glyphe == glyphen[stelle] {
			gesetzt = append(gesetzt, runen[stelle])
		} else {
			gesetzt = append(gesetzt, glyphenzeichen(glyphe))
		}

		stelle += deckt[i]
	}

	return gesetzt
}
