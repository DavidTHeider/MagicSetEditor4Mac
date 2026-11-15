package render

import (
	"fmt"
	"strconv"
	"unicode"
)

// **Wie ein fehlendes Zeichen benannt wird.**
//
// Getrennt von `textzeichnen.go`, weil es eine andere Frage ist: dort steht,
// wie ein Glyph auf die Fläche kommt — hier, wie die **Meldung** darüber
// lautet, wenn keiner da ist. Was diese Namen leisten müssen, steht in
// `fehlzeichenmeldung_test.go`: ein Bindestrich und ein Geviertstrich sehen
// gleich aus, und „Zeichen fehlt" hülfe niemandem.

// zahlwort schreibt eine kleine Anzahl aus.
func zahlwort(anzahl int) string {
	if anzahl == 1 {
		return "eine Zeile"
	}

	return strconv.Itoa(anzahl) + " Zeilen"
}

// unsichtbar sagt, ob ein Zeichen von sich aus nichts zeichnet.
//
// **Kein Mangel, sondern die Bauform.** Ein Nullbreiten-Leerzeichen trennt
// Wörter, ohne selbst etwas zu malen; der weiche Trennstrich zeigt nur an, wo
// getrennt werden **darf**; die Reihenfolgemarke ist eine Angabe über die
// Datei. Eine Schrift führt für sie keine Glyphe, und eine Meldung darüber
// zeigt auf nichts — sie stünde in derselben Liste wie die echten
// (#270).
//
// **Die Liste ist kurz und benannt.** Alles, was Unicode als „Format" führt,
// wäre mehr, als dieser Rasterer je gesehen hat — und eine Regel über eine
// ganze Kategorie ist eine Behauptung über Zeichen, die niemand gemessen hat
// (§10).
// zeichenname benennt ein Zeichen so, dass man es in einer Meldung erkennt.
//
// **Ein Befund muss zeigen, was er gefunden hat**
// (#476). Roh
// in die Meldung gesetzt, läse sich ein Zeichen ohne Gestalt so:
//
//	in der Schrift MPlantin fehlen Zeichen:
//
// Wer das liest, kann nichts tun.
//
// **Ein sichtbares bleibt sichtbar**, in Anführungszeichen; alles andere
// bekommt seine Nummer.
func zeichenname(zeichen rune) string {
	if unicode.IsGraphic(zeichen) && !unicode.IsSpace(zeichen) {
		return strconv.Quote(string(zeichen))
	}

	return fmt.Sprintf("U+%04X", zeichen)
}

func unsichtbar(zeichen rune) bool {
	switch zeichen {
	case '\u200b', '\u200c', '\u200d', '\u00ad', '\ufeff':
		return true
	}

	return false
}

// zeilengrundlinie gibt die Grundlinie einer Zeile in Vorlagenpunkten.
//
// **Die Oberkante gibt [SetzeAbschnitte], den Rest die Schrift** — und sie
// rastet auf dasselbe Viertel wie die Zeilenhöhe, denn es ist dieselbe
// Messung: MSE zeichnet bei der Oberkante und überlässt wx die Grundlinie
// (`FontTextElement::draw`, `src/render/text/font.cpp`, Zeilen 15–25, Fassung
// f9d9356 vom 2021-07-06), und wx misst bei Vierfachgröße in ganzen
