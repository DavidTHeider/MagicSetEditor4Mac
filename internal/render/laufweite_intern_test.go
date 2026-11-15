package render

import (
	"testing"
	"unicode"
)

// gleichbreit gibt jeder Rune dieselbe Breite — außer den Marken.
//
// **Auch die Marken bekommen eine**, und das ist Absicht: eine Schrift, die für
// sie 0 gäbe, ließe den Test grün werden, ohne dass die Laufweite gezählt
// hätte. Eine Vorrichtung, die den Fall selbst wegräumt, prüft ihn nicht.
type gleichbreit struct{}

func (gleichbreit) Breite(zeichen rune) float64 {
	if unicode.Is(unicode.Mn, zeichen) {
		return 0.3
	}

	return 0.5
}

func (gleichbreit) Zeilenhoehe() float64 { return 1 }

// **Die Laufweite zählt Buchstaben, nicht Runen.**
//
// # Der Fehler
//
// `e` gefolgt von U+0301 (kombinierender Akut) ist **ein** Buchstabe aus zwei
// Runen. Wer Runen zählt, legt zwischen Grundzeichen und Marke eine Laufweite —
// und die Marke sitzt danach neben ihrem Buchstaben statt über ihm
// (#113).
//
// Der Fehler ist klein: 0,2 Punkte je kombinierendem Zeichen in `magic-m15`.
// Er wächst mit jeder Sprache, die Diakritika stapelt — Vietnamesisch legt zwei
// auf einen Vokal.
//
// # Was hier **nicht** gelöst ist
//
// Die vollständige Zerlegung nach UAX #29
// ([0016](../../docs/entscheidungen/0016-keine-unicode-segmentierung.md)):
// Emoji-Folgen mit Nullbreiten-Verbinder, Regionalindikatoren, Hangul-Jamo.
// Sie zählen weiter als mehrere Buchstaben, und das steht im Vermerk.
func TestDieLaufweiteZaehltBuchstabenUndNichtRunen(t *testing.T) {
	satz := Satz{Groesse: 10, Laufweite: 2}

	// **Zwei Runen, ein Buchstabe.** Zwischen ihnen gehört keine Laufweite.
	mitMarke := breiteVon("é", satz, gleichbreit{})
	ohneMarke := breiteVon("e", satz, gleichbreit{})

	// Die Marke bringt ihre eigene Breite mit (0,3 × 10) — aber keine Laufweite.
	if erwartet := ohneMarke + 3; mitMarke != erwartet {
		t.Errorf("%g, erwartet %g — die Laufweite wurde mitgezählt", mitMarke, erwartet)
	}

	// **Zwei echte Buchstaben bekommen ihre.**
	zwei := breiteVon("ab", satz, gleichbreit{})
	if erwartet := ohneMarke + 5 + satz.Laufweite; zwei != erwartet {
		t.Errorf("%g, erwartet %g", zwei, erwartet)
	}
}

// **Eine Zeile aus lauter Marken bleibt ein Buchstabe.**
//
// Der Fall ist unsinnig und trotzdem der, an dem eine Zählung mit `-1`
// stolpert: null Buchstaben minus eins ergäbe eine **negative** Laufweite und
// damit eine Zeile, die schmaler ist als ihre Zeichen.
func TestEineZeileAusLauterMarkenBleibtEinBuchstabe(t *testing.T) {
	satz := Satz{Groesse: 10, Laufweite: 2}

	breite := breiteVon("́́", satz, gleichbreit{})
	if breite != 6 {
		t.Errorf("%g, erwartet 6 — zwei Marken zu je 0,3 und keine Laufweite", breite)
	}
}

// **Was gemessen wird, wird auch gezeichnet.**
//
// # Der Befund, der diesen Test nötig macht
//
// Aus dem Zweitblick durch Codex über Phase 6. `breiteVon` zählt seit #113 die
// Laufweite zwischen **Buchstaben**; die Zeichenschleife addierte sie weiter
// nach **jeder Rune**. Die Zeile war damit richtig breit und ihre Glyphen
// falsch gesetzt — bei `e` plus U+0301 rückte alles hinter der Marke um eine
// Laufweite nach rechts.
//
// **Zwei Rechnungen über dieselbe Sache laufen auseinander**, und zwar an dem
// Tag, an dem eine von ihnen etwas lernt. Deshalb rechnet die Zeichenschleife
// jetzt mit derselben Funktion.
func TestWasGemessenWirdWirdAuchGezeichnet(t *testing.T) {
	satz := Satz{Groesse: 10, Laufweite: 2}

	for _, zeile := range []string{"ab", "é", "e", "́", "abc"} {
		gemessen := breiteVon(zeile, satz, gleichbreit{})

		runen := []rune(zeile)

		gezeichnet := 0.0
		for stelle, zeichen := range runen {
			markeFolgt := stelle+1 < len(runen) && !istBuchstabe(runen[stelle+1])
			gezeichnet += vorschub(0, zeichen, satz, gleichbreit{}, 1, markeFolgt)
		}

		// **Die Zeichenschleife legt hinter das letzte Zeichen noch eine
		// Laufweite** — die Messung nicht, denn sie steht zwischen den
		// Zeichen. Der Unterschied ist genau eine, wenn es welche gab.
		erwartet := gemessen
		if buchstabenIn(zeile) > 0 {
			erwartet += satz.Laufweite
		}

		if gezeichnet != erwartet {
			t.Errorf("%q: gezeichnet %g, gemessen %g (erwartet %g)",
				zeile, gezeichnet, gemessen, erwartet)
		}
	}
}

// **Die Laufweite steht hinter dem Buchstaben, nicht hinter dem Grundzeichen.**
//
// # Der Befund des Nachlaufs
//
// `vorschub` legte sie unmittelbar hinter das Grundzeichen — und damit **vor**
// seine kombinierende Marke. Die Marke saß danach um eine Laufweite zu weit
// rechts, also neben ihrem Buchstaben statt über ihm.
//
// **Das ist derselbe Fehler, den #113 beheben wollte**, eine Stelle weiter: die
// erste Korrektur hat die Zeichen **hinter** der Marke gerade gerückt und die
// Marke selbst stehen lassen. Der Fall wurde behandelt, nicht seine Klasse.
func TestDieMarkeSitztAufIhremBuchstaben(t *testing.T) {
	satz := Satz{Groesse: 10, Laufweite: 2}

	// **Die Marke rückt um die Breite ihres Grundzeichens vor, sonst nichts.**
	// Zwischen `e` und seinem Akut gehört keine Laufweite.
	if vorschub(0, 'e', satz, gleichbreit{}, 1, true) != 5 {
		t.Errorf("vor einer Marke: %g, erwartet 5",
			vorschub(0, 'e', satz, gleichbreit{}, 1, true))
	}

	// **Vor einem Buchstaben gehört sie dazu.**
	if vorschub(0, 'e', satz, gleichbreit{}, 1, false) != 7 {
		t.Errorf("vor einem Buchstaben: %g, erwartet 7",
			vorschub(0, 'e', satz, gleichbreit{}, 1, false))
	}

	// **Und eine Marke selbst schiebt nur ihre eigene Breite.**
	if vorschub(0, '́', satz, gleichbreit{}, 1, false) != 3 {
		t.Errorf("eine Marke: %g, erwartet 3",
			vorschub(0, '́', satz, gleichbreit{}, 1, false))
	}
}
