package render

import (
	"math"
	"slices"
)

// zeilenlagen rechnet je Zeile ihre Höhe und ihre Oberkante — vom inneren
// Kasten an, **vor** Dehnung und Ausrichtung, so wie MSE sie beim Umbruch
// führt (`line.top` in `prepareLinesAtScale`, `src/render/text/viewer.cpp`,
// Zeilen 557–735, Fassung f9d9356 vom 2021-07-06). Mit dieser Oberkante
// fragt der Umbruch die Maske ([zeilenraum]).
func zeilenlagen(
	roh []rohzeile, arten []umbruchart, raum zeilenraum, satz Satz,
	schriften Schriftsatz, bilder Bildquelle,
) (oben, hoehen []float64) {
	hoehen = make([]float64, len(roh))
	oben = make([]float64, len(roh))

	for nummer := range roh {
		hoehen[nummer] = zeilenhoehe(roh[nummer].stuecke, satz, schriften, bilder)

		// **Eine leere Zeile nach einem `<line>` hat die Höhe null.**
		//
		//	if (line.break_after == LineBreak::LINE) line.line_height = 0;
		//
		// `src/render/text/viewer.cpp`, Zeile 684, Fassung f9d9356 vom
		// 2021-07-06. MSE setzt die Höhe der **neuen** Zeile auf null; trägt
		// sie Inhalt, wächst sie wieder mit ihm, und nur die leere bleibt bei
		// null. Erst damit greift die Rückwärtsschleife in [gesamthoehe]
		// (#528).
		if nummer > 0 && arten[nummer-1] == umbruchLinie && istLeer(roh[nummer]) {
			hoehen[nummer] = 0
		}

		// **Und die erste Zeile ohne Inhalt bleibt bei null.**
		//
		//	Line() : start(0), end_or_soft(0), top(0), line_height(0) …
		//
		//	if (line_size.height < 0.01 && !lines.empty()) {
		//	  // if a line has 0 height, use the height of the line above it,
		//	  // but at most once
		//	} else {
		//	  line.line_height = line_size.height;
		//	}
		//
		// `src/render/text/viewer.cpp`, Zeilen 27–30 und 668–673, Fassung
		// f9d9356 vom 2021-07-06. **Die Vererbung greift nur, wenn es eine
		// Zeile davor gibt** (`!lines.empty()`); die erste bleibt bei null.
		//
		// Angestrichen vom Zweitblick über Phase J
		// (#553).
		//
		// **Ist sie die einzige, bekommt sie doch eine Höhe** — die eines
		// Zeichens der Schrift: „no text, find a dummy height for the single
		// line" (`TextViewer::prepareLines`, `src/render/text/viewer.cpp`,
		// Zeilen 397–405, Fassung f9d9356 vom 2021-07-06). Hier stand bis zum
		// 2026-10-11, ein ganz leerer Block habe die Höhe null; das Orakel sagt
		// anderes (#2298), und die Höhe geht als `content_height` in die
		// Skripte der Vorlagen ein.
		if nummer == 0 && istLeer(roh[nummer]) && len(roh) > 1 {
			hoehen[nummer] = 0
		}

		if nummer > 0 {
			// **Wie weit es weitergeht, hängt am Umbruch der Zeile davor.**
			oben[nummer] = oben[nummer-1] +
				hoehen[nummer-1]*satz.Zeilenfaktoren.faktorFuer(arten[nummer-1])
		}

		// **Der obere Absatzrand kommt oben drauf** — MSE setzt ihn beim
		// Absatzwechsel auf `line.top` (`viewer.cpp:694`) und **nur dort**;
		// welche Zeile ihn trägt, entscheidet [zeilenAus].
		oben[nummer] += roh[nummer].margenOben()

		// **Und passt an ihrer Stelle nichts, rückt sie hinunter** — bevor die
		// nächste von ihr aus weiterzählt.
		oben[nummer] = raum.abgerueckt(oben[nummer], roh[nummer].margen)
	}

	return oben, hoehen
}

// umbrochen bricht den Zug in Zeilen.
//
// **Mit einer Maske hängt die Breite einer Zeile an ihrer Oberkante** und die
// Oberkante an den Zeilen davor; MSE rechnet beides in einem Zug. Hier wird
// so oft gebrochen, bis sich nichts mehr ändert: jede Zeile bekommt die Breite
// an der Oberkante, die der Lauf davor für sie fand. Nach dem k-ten Lauf
// stehen damit die ersten k Zeilen fest, und ein Umbruch, der sich nicht mehr
// ändert, ist der, den MSE in einem Zug fände.
//
// Ohne Maske ist jede Zeile gleich breit, und ein Lauf genügt.
func umbrochen(
	abschnitte []Abschnitt, raum zeilenraum, satz Satz, schriften Schriftsatz,
	bilder Bildquelle,
) []rohzeile {
	breite := func(oben []float64) func(int) float64 {
		return func(zeile int) float64 {
			lage := 0.0
			if len(oben) > 0 {
				lage = oben[min(zeile, len(oben)-1)]
			}

			_, frei := raum.bei(lage)

			return frei
		}
	}

	// **Ein einzeiliges Feld hat keinen rechten Rand** (J05, #2269): MSE
	// bricht es nur an einem harten Umbruch, und der kommt aus dem Text, nicht
	// aus der Breite — siehe [Satz.Einzeilig].
	if satz.Einzeilig {
		return zeilenAus(abschnitte, func(int) float64 { return math.Inf(1) }, satz, schriften, bilder)
	}

	roh := zeilenAus(abschnitte, breite(nil), satz, schriften, bilder)
	if raum.maske == nil {
		return roh
	}

	for lauf := 0; lauf <= len(roh); lauf++ {
		oben, _ := zeilenlagen(roh, umbruchartenVon(roh), raum, satz, schriften, bilder)

		neu := zeilenAus(abschnitte, breite(oben), satz, schriften, bilder)
		if gleichGebrochen(neu, roh) {
			break
		}

		roh = neu
	}

	return roh
}

// gleichGebrochen sagt, ob zwei Umbrüche dieselben Zeilen ergeben.
func gleichGebrochen(a, b []rohzeile) bool {
	return slices.EqualFunc(a, b, func(x, y rohzeile) bool {
		return textAus(x.stuecke) == textAus(y.stuecke)
	})
}

// unterkante gibt, woran MSE misst, ob ein Text passt — vom oberen Rand des
// Felds an:
//
//	return lines.back().bottom() <= dc.getInternalSize().height - style.padding_bottom;
//
// `prepareLinesAtScale`, `src/render/text/viewer.cpp`, Zeile 733, Fassung
// f9d9356 vom 2021-07-06. **Die letzte Zeile, auch wenn sie leer ist** —
// hinter `<line>` steht sie um den Abstand tiefer und ist null hoch; die
// Ausrichtung dagegen lässt leere Zeilen am Ende weg ([gesamthoehe]).
//
// **Mit festen Absatzbändern** gilt eine eigene Rechnung (ebd. Zeilen
// 718–732): die höchste Absatzhöhe mal die Zahl der Bänder, die der Kasten
// fasst, `floor(0,5 + (Höhe − padding_bottom) / paragraph_height)`. Ein
// Absatz endet dort, wo der Umbruch nicht weich ist.
func unterkante(
	oben, hoehen []float64, arten []umbruchart, satz Satz, feldhoehe float64,
) float64 {
	letzte := len(oben) - 1
	if satz.Absatzhoehe <= 0 {
		return satz.Polster.Oben + oben[letzte] + hoehen[letzte]
	}

	hoechste, anfang := 0.0, 0

	for zeile := range oben {
		if arten[zeile] == umbruchWeich {
			continue
		}

		hoechste = max(hoechste, oben[zeile]+hoehen[zeile]-oben[anfang])
		anfang = zeile + 1
	}

	baender := math.Floor(0.5 + (feldhoehe-satz.Polster.Unten)/satz.Absatzhoehe)

	return hoechste * baender
}
