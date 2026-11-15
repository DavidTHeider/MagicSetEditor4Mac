package render

import "math"

// textrasterung ist MSEs `text_scaling`.
//
//	const int text_scaling = 4;
//
// `src/gfx/resample_text.cpp`, Zeile 22, Fassung f9d9356 vom 2021-07-06. MSE
// setzt, misst und zeichnet Text in der **vierfachen** Größe und verkleinert
// das Ergebnis danach — daher der Name der Datei, in der die Zahl steht.
const textrasterung = 4

// gerastertesTextmass rastert ein Maß der Schrift so, wie MSE es misst —
// die Zeilenhöhe wie die Grundlinie.
//
// # Die Rechnung
//
// MSE fragt wxWidgets nach der Texthöhe, und wx antwortet in **ganzen
// Bildpunkten**. Zwei Rundungen sitzen darin, und beide bei Vierfachgröße:
//
//  1. Die Punktgröße der Schrift wird ganzzahlig —
//     `dc.SetFont(font.toWxFont(trS(scale) * text_scaling))`
//     (`src/util/rotation.cpp`, Zeile 326) und `int size_i = to_int(scale * size)`
//     (`src/data/font.cpp`, Zeile 84).
//  2. Die gemessene Höhe ist ganzzahlig und wird durch `zoomY * text_scaling`
//     zurückgeteilt — `RotatedDC::GetTextExtent`, Zeile 349.
//
// Beim Bildexport ist `zoom = 1` (`UnzoomedDataViewer::getRotation`,
// `src/data/format/image.cpp`, Zeilen 37–45), und die Vorlagenpunkte dieses
// Programms sind MSEs innere Einheiten. Übrig bleibt
//
//	ceil(verhaeltnis * to_int(4 * groesse)) / 4
//
// Alles Fassung f9d9356 vom 2021-07-06.
//
// # Und bei jeder anderen Vergrößerung rastert MSE feiner
//
// **`trS(scale)` ist `scale * zoomY`** (`src/util/rotation.hpp`, Zeile 63).
// Allgemein bleibt in inneren Einheiten
//
//	ceil(verhaeltnis * to_int(4 * zoom * groesse)) / (4 * zoom)
//
// — bei doppelter Vergrößerung also Achtelpunkte statt Viertel. **Diese
// Rechnung hier ist der Sonderfall `zoom = 1`**, und das ist für die Referenz
// richtig und für den Druckbogen eine Abweichung: dort rastert dieses Programm
// weiter auf Viertel, MSE feiner.
//
// **Sie steht als benannte Abweichung und nicht als Versehen**
// (#612) —
// angestrichen vom Zweitblick über Phase K. Der Satz dieses Programms ist
// auflösungsunabhängig (§2: der Plan steht in Vorlagenpunkten), MSEs ist es
// nicht.
//
// **Entschieden am 2026-08-20: sie bleibt**
// ([0035](../../docs/entscheidungen/0035-der-satz-bleibt-aufloesungsunabhaengig.md)).
// Was sie kostet, ist ausgerechnet — je Zeile, bei der Vergrößerung des
// Druckbogens (1,9843 für eine 63-mm-Karte bei 300 dpi über einem Plan von
// 375 Punkten):
//
//	Zeilenhöhe MPlantin      0,327 pt = 0,65 Bildpunkte
//	Grundlinie MPlantin      0,305 pt = 0,60 Bildpunkte
//	schlimmster Fall über
//	alle Verhältnisse 0,6–2,0
//	und Größen 4–40          0,513 pt = 1,02 Bildpunkte
//
// Über einen sechszeiligen Regeltext summiert sich das auf knapp vier
// Bildpunkte — **0,33 mm bei 300 dpi**. Dafür die Naht aufzugeben, die den
// ganzen Rasterpfad auflösungsunabhängig hält, wäre der schlechtere Tausch.
//
// **Eine Schranke von einem Viertelpunkt gilt nicht**, obwohl unser Raster das
// gröbere ist: `to_int(4 · zoom · groesse)` rundet die **skalierte** Größe, und
// das Teilen durch `4 · zoom` bringt den Fehler nicht auf unser Raster zurück.
// Das war die erste Annahme und ist an 2.528.805 durchgerechneten Paaren
// widerlegt.
//
// **Und was hier gerechnet und nicht gemessen ist** (§10): die Zahlen folgen
// aus **beiden Formeln**, und MSEs steht nachgelesen da — sie sind nicht an
// MSEs Ausgabe bei `zoom ≠ 1` abgelesen, denn die liegt nicht vor. Der
// Unterschied hat außerdem kein festes Vorzeichen: er zittert, er verschiebt
// nicht.
//
// # Warum nicht auf ganze Punkte
//
// **Das war der erste Anlauf, und er ist gemessen widerlegt.** Am 2026-08-19
// umgesetzt und zurückgenommen: `ceil(verhaeltnis * groesse)` machte den
// Abstand zu MSE2 von 22,542 auf 26,299, also um **3,757 schlechter**
// (#546).
// Der Faktor vier fehlte — bei `MPlantin` und Größe 14 sind es 16,25 und nicht
// 17, ein Unterschied von 0,75 Punkten je Zeile.
//
// # Was gemessen ist, und was nicht
//
// Dass wx' Texthöhe `ceil(verhaeltnis * punktgroesse)` ist, ist am 2026-08-20
// gegen **wx 3.3.3** an zwanzig Punktgrößen von 9 bis 120 gemessen, für
// `mplantin` (1,149) und `beleren-bold` (1,205566): **40 von 40**, die
// Vierfachgrößen 56, 96 und 120 darunter.
//
// **Für `MATRIX.TTF` trifft sie nicht** — dort nennt die Datei keine
// `hhea`-Metrik, und unser Rückfall liest den Tintenkasten (0,957), wo wx die
// `OS/2`-typo-Metrik nimmt (0,860). Das ist ein eigener Fehler und ein eigenes
// Issue (#597),
// keiner dieser Rechnung.
//
// **Und die Grundlinie folgt ihr nicht.** Sie steht als eigene Frage an
// #446; wx
// meldet für `mplantin` bei 120 pt einen Descent von 28, wo `hhea` 27 gäbe.
func gerastertesTextmass(verhaeltnis, groesse float64) float64 {
	punktgroesse := mseGanzzahl(textrasterung * groesse)

	return math.Ceil(verhaeltnis*punktgroesse) / textrasterung
}

// mseGanzzahl ist MSEs `to_int`.
//
//	inline int to_int(double d) {
//	  return static_cast<int>(d > 0 ? d + 0.4999995 : d - 0.4999995);
//	}
//
// `src/util/vector2d.hpp`, Zeilen 20–22, Fassung f9d9356 vom 2021-07-06.
//
// **Die 0,4999995 sind Absicht und kein Tippfehler**; der Kommentar daneben
// sagt, warum: knapp unter einem Halben gibt bei Ausdrücken wie `x/2` das
// gleichmäßigere Ergebnis. Wer hier `math.Round` schriebe, träfe MSE genau
// dort nicht, wo es darauf ankommt.
func mseGanzzahl(wert float64) float64 {
	if wert > 0 {
		return math.Trunc(wert + 0.4999995)
	}

	return math.Trunc(wert - 0.4999995)
}
