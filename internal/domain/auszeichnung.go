// **Vom rohen Wert zum Text und zurück** — die Auszeichnungen einer
// MSE-Zeichenkette.
//
// # Warum eine eigene Datei
//
// **Weil sie eine andere Frage beantwortet als `ersetzung.go`.** Dort steht,
// *was* ersetzt wird; hier, *wie* man in einem Wert wie
// `<i>Fliegend</i>. Sehr schnell.` überhaupt eine Stelle findet und sie
// zurückrechnet, ohne die Auszeichnungen mitzunehmen.
//
// Der Anlass ist die Zeilengrenze (§2): `ersetzung.go` ging über 400 Zeilen,
// als der Zweitblick über Phase 3.7 die zweite Grenzliste nötig machte.

package domain

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// OhneAuszeichnung gibt den Wert als reinen Text — MSEs `untag`.
//
// **Für eine Anzeige und nicht zum Schreiben.** Ein `<i>` sieht in einer
// Vorschau wie ein Fehler aus; es ist Werkzeug und kein Text. Der Wert selbst
// bleibt roh — dieselbe Trennung wie in `frontend/src/seiten/kartenname.ts`.
func OhneAuszeichnung(roh string) string {
	text, _, _ := entmarkt(roh, false)

	return text
}

// entmarkt gibt den Wert ohne Auszeichnungen und je Stelle darin die Stelle im
// rohen Wert.
//
// **Die zweite Rückgabe ist der Punkt.** Ohne sie ließe sich ein Fund im Text
// nicht in den rohen Wert zurückrechnen — und eine Ersetzung schriebe die
// Auszeichnungen weg. MSE hat dafür `untagged_to_index`
// (`src/util/tagged_string.cpp`, Fassung f9d9356 vom 2021-07-06); hier fällt
// die Zuordnung beim Entmarken ab.
//
// **Zwei Listen, je einen Eintrag pro Byte** des entmarkten Textes:
// `grenzen` nennt den Anfang des Zeichens im rohen Wert, `enden` die Stelle
// **hinter** ihm.
//
// **Zwei und nicht eine.** Der erste Anlauf führte nur die Anfänge und nahm
// als Ende den Anfang des **nächsten** Zeichens. Dazwischen können
// Auszeichnungen stehen — bei `a<i>b</i>` und `a → X` wurde das `<i>`
// mitgelöscht. Der Fall am Ende des Wertes war gesehen und mit
// `hinterDemLetzten` behoben; **jeder Fall in der Mitte war es nicht**
// (Befund 9 des Zweitblicks über Phase 3.7 und die Klasse dahinter).
//
//	if (c == _('<')) in_tag = true;
//	else if (!in_tag) ret += untag_char(c);
//	else if (c == _('>')) in_tag = false;
//
// **`kleinschreiben` gehört hierher und nicht dahinter** — Befund 9 des
// Zweitblicks über Phase 3.7
// (#1168).
//
// Der erste Anlauf baute die Grenzen über den Text **in seiner Schreibung**
// und suchte danach in `strings.ToLower(text)`. Das trägt nur, solange die
// Kleinschreibung die Bytelänge erhält — `İ` (U+0130, zwei Bytes) wird in Go
// zu `i` (ein Byte), und die Stellen liefen um eins auseinander:
// `Ersetze("İ")` gab `"Xİ"` statt `"X"`.
//
// **Kleingeschrieben wird deshalb je Zeichen, während die Grenzen entstehen**;
// alle Bytes eines kleingeschriebenen Zeichens zeigen auf den Anfang des
// **ursprünglichen** im rohen Wert, und die Obergrenze zählt dessen Länge.
//
// `src/util/tagged_string.cpp`, Zeilen 33–47.
func entmarkt(roh string, kleinschreiben bool) (text string, grenzen, enden []int) {
	var gebaut strings.Builder

	grenzen = make([]int, 0, len(roh)+1)
	enden = make([]int, 0, len(roh)+1)
	inMarke := false

	// hinterDemLetzten ist die rohe Stelle **hinter** dem zuletzt
	// übernommenen Zeichen.
	//
	// **Und nicht `len(roh)`.** Ein Wert, der auf `</i>` endet, hat hinter
	// seinem letzten Textzeichen noch vier Zeichen Auszeichnung; wer die Länge
	// des rohen Wertes als Obergrenze nähme, schnitte sie beim letzten Treffer
	// mit weg.
	hinterDemLetzten := 0

	for stelle, zeichen := range roh {
		switch {
		case zeichen == '<':
			inMarke = true
		case inMarke:
			if zeichen == '>' {
				inMarke = false
			}
		default:
			imRohen := zeichen

			// **`\x01` ist ein geschütztes Kleinerzeichen** — MSEs
			// `ESCAPED_LANGLE`, und `untag_char` gibt es als `<` zurück.
			if zeichen == '\x01' {
				zeichen = '<'
			}

			// **Je Byte ein Eintrag und nicht je Zeichen.** Gesucht wird über
			// Bytes (`strings.HasPrefix` auf `text[stelle:]`); eine Zuordnung
			// je Zeichen liefe hinter dem ersten Umlaut aus dem Takt und beim
			// letzten Treffer über das Ende hinaus.
			//
			// Alle Bytes desselben Zeichens zeigen auf **seinen** Anfang im
			// rohen Wert. Getroffen wird ohnehin nur an einer Zeichengrenze —
			// dort stimmt es genau.
			hinterDemLetzten = stelle + utf8.RuneLen(imRohen)

			if kleinschreiben {
				zeichen = unicode.ToLower(zeichen)
			}

			breite, _ := gebaut.WriteRune(zeichen)
			for range breite {
				grenzen = append(grenzen, stelle)
				enden = append(enden, hinterDemLetzten)
			}
		}
	}

	return gebaut.String(),
		append(grenzen, hinterDemLetzten), append(enden, hinterDemLetzten)
}

// marke ist MSEs `escape`: ein `<` im Ersatz wird geschützt.
//
//	String escape(const String& str) { … ret += tag_char(c); … }
//	Char tag_char(Char c) { if (c == '<') return ESCAPED_LANGLE; else return c; }
//
// **Sonst schriebe ein Ersatz eine Auszeichnung.** Wer `<i` einsetzt, meint
// zwei Zeichen und keine Kursivstellung — MSE tut dasselbe
// (`panel.editor->insert(escape(what.GetReplaceString()), _("Replace"))`,
// `src/gui/set/cards_panel.cpp`, Zeile 434, Fassung f9d9356 vom 2021-07-06).
func marke(text string) string {
	return strings.ReplaceAll(text, "<", "\x01")
}

// Markenausgleich sagt, was vor und hinter einem Ersatz stehen muss, damit
// der Wert ausgezeichnet bleibt.
//
// # Der Fall, für den es ihn gibt
//
// Ein Treffer kann **über** eine Marke hinweggehen. Wer den rohen Bereich
// dann stumpf herausschneidet, nimmt die Marke mit und lässt ihre Gegenmarke
// stehen: `a<i>b</i>` wird bei `ab → ""` zu `</i>`.
//
// **MSE2 gleicht aus** (`tagged_substr_replace`, `src/util/tagged_string.cpp`,
// Zeilen 622–632, Fassung `f9d9356` vom 2021-07-06): es sammelt die Marken im
// geschnittenen Bereich, hebt Paare darin auf und setzt, was übrig bleibt,
// **geschlossen davor** und **offen dahinter** — danach wird vereinfacht.
//
// # Was übrig bleibt, ist zweierlei
//
//   - Eine **offene** Marke ohne Gegenstück im Bereich wurde vor dem Schnitt
//     geöffnet und wird danach geschlossen: sie gehört **hinter** den Ersatz,
//     damit die Schließung dahinter wieder etwas zu schließen hat.
//   - Eine **schließende** ohne Gegenstück gehört **vor** den Ersatz, aus dem
//     spiegelbildlichen Grund.
//
// Die Reihenfolge zählt: Geschachteltes wird von innen nach außen geschlossen
// und von außen nach innen geöffnet.
func Markenausgleich(roh string, anfang, ende int) (davor, danach string) {
	var offen, geschlossen []string

	for _, marke := range markenIn(roh[anfang:ende]) {
		if rest, istSchliessend := strings.CutPrefix(marke, "/"); istSchliessend {
			// **Nur die zuletzt geöffnete hebt sie auf.** Ein `</i>` schließt
			// das `<i>` davor, nicht irgendeines weiter außen.
			if len(offen) > 0 && offen[len(offen)-1] == rest {
				offen = offen[:len(offen)-1]

				continue
			}

			geschlossen = append(geschlossen, rest)

			continue
		}

		offen = append(offen, marke)
	}

	// **Was vor dem Bereich offen war**, wird davor geschlossen — von innen
	// nach außen, also in umgekehrter Reihenfolge des Auftretens.
	var vorne strings.Builder
	for i := len(geschlossen) - 1; i >= 0; i-- {
		vorne.WriteString("</" + geschlossen[i] + ">")
	}

	// **Was im Bereich offen blieb**, wird dahinter wieder geöffnet — von
	// außen nach innen.
	var hinten strings.Builder
	for _, marke := range offen {
		hinten.WriteString("<" + marke + ">")
	}

	return vorne.String(), hinten.String()
}

// markenIn gibt die Marken eines rohen Abschnitts, ohne spitze Klammern.
//
// **`\x01` zählt nicht als Klammer** — es ist ein geschütztes Kleinerzeichen
// und kein Markenanfang (MSEs `ESCAPED_LANGLE`).
func markenIn(roh string) []string {
	var gefunden []string

	for stelle := 0; stelle < len(roh); {
		auf := strings.IndexByte(roh[stelle:], '<')
		if auf < 0 {
			break
		}

		auf += stelle

		zu := strings.IndexByte(roh[auf:], '>')
		if zu < 0 {
			break
		}

		zu += auf

		gefunden = append(gefunden, roh[auf+1:zu])
		stelle = zu + 1
	}

	return gefunden
}

// OhneLeerePaare hebt Marken auf, die unmittelbar nichts umschließen.
//
// **MSEs `simplify_tagged`** tut das nach jedem Ersatz
// (`src/util/tagged_string.cpp`, Zeile 639, Fassung `f9d9356` vom 2021-07-06). Ohne es bliebe von
// `a<i>b</i>` bei `ab → ""` ein `<i></i>` stehen — richtig ausgezeichnet und
// doch Unrat.
//
// **Was es nicht tut**: überlappende Marken zusammenlegen. MSE hat dafür
// `simplify_tagged_overlap`; hier entsteht keine Überlappung, weil der
// Ausgleich immer paarweise setzt.
func OhneLeerePaare(roh string) string {
	for {
		gekuerzt := leeresPaarWeg(roh)
		if gekuerzt == roh {
			return roh
		}

		roh = gekuerzt
	}
}

// leeresPaarWeg nimmt das erste unmittelbar leere Markenpaar heraus.
func leeresPaarWeg(roh string) string {
	for _, marke := range markenIn(roh) {
		if strings.HasPrefix(marke, "/") {
			continue
		}

		paar := "<" + marke + "></" + marke + ">"
		if stelle := strings.Index(roh, paar); stelle >= 0 {
			return roh[:stelle] + roh[stelle+len(paar):]
		}
	}

	return roh
}
