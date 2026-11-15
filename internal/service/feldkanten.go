package service

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Ein Feldstil hat sechs Maße, und drei davon rechnet MSE aus.**
//
//	if      (automatic_side & AUTO_LEFT)   left   = right - width;
//	else if (automatic_side & AUTO_WIDTH)  width  = right - left;
//	else if (automatic_side & AUTO_RIGHT)  right  = left + width;
//	else  { int lr = int(left + right); left = (lr - width) / 2; right = (lr + width) / 2; }
//
// `Style::update`, `src/data/field.cpp`, Zeilen 152–172, Fassung f9d9356 vom
// 2021-07-06. Senkrecht steht dasselbe mit `top`, `height` und `bottom`.
//
// # Welches der drei fehlt
//
// MSE fragt nicht die Datei, sondern den **Wert**: ein nicht beschriebenes Maß
// trägt 1.000.000, und `is_set` heißt `x < 100000` (Zeilen 133–138). Für die
// Breite gilt eine eigene Frage — `is_setw`, `|x| > 0.001` —, denn ihre Vorgabe
// ist die Null, und eine Breite von null ist keine Breite.
//
// **Ein Skript gilt als gesetzt**, gleich was es ergibt (`x.isScripted()`).
// Hier ist das die Angabe, die nicht leer ist; ausgerechnet wird sie erst beim
// Anordnen, und diese Abweichung steht unten.
//
// # Wozu die Karte das braucht
//
// Die Saga-Vorlagen stapeln ihre Stufen aufeinander:
// `top: { card_style.level_1_text.bottom }`. Ohne `right` und `bottom` in der
// Umgebung ergab jede dieser Angaben „keine Zahl, sondern nichts" — zwölfmal
// über die 190 Stile von `M15-Magic-Pack`
// (#291).
func kanten(
	ctx context.Context, feld *msepaket.Feldstil, umgebung *msescript.Umgebung,
) []msescript.Eintrag {
	links, breite, rechts := gerechneteKanten(ctx, feld, umgebung,
		feld.Links, feld.Breite, feld.Rechts)
	oben, hoehe, unten := gerechneteKanten(ctx, feld, umgebung,
		feld.Oben, feld.Hoehe, feld.Unten)

	return []msescript.Eintrag{
		{Name: "left", Wert: msescript.Zahl(links)},
		{Name: "top", Wert: msescript.Zahl(oben)},
		{Name: "width", Wert: msescript.Zahl(breite)},
		{Name: "height", Wert: msescript.Zahl(hoehe)},
		{Name: "right", Wert: msescript.Zahl(rechts)},
		{Name: "bottom", Wert: msescript.Zahl(unten)},
	}
}

// gerechneteKanten ergänzt aus zweien von dreien das dritte.
//
// **Die Reihenfolge der Fragen ist die von MSE** und nicht austauschbar: fehlt
// die Gegenkante, wird sie gerechnet; sonst die Länge; sonst die Kante; und
// wenn alle drei dastehen, gilt die Länge und steht **mittig** zwischen den
// Kanten.
func gerechneteKanten(
	ctx context.Context, feld *msepaket.Feldstil, umgebung *msescript.Umgebung,
	roheKante, roheLaenge, roheGegenkante string,
) (kante, laenge, gegenkante float64) {
	kante, kanteDa := massangabe(ctx, feld.Feld, "kante", roheKante,
		Kantenvorgabe, umgebung)
	laenge, laengeDa := massangabe(ctx, feld.Feld, "länge", roheLaenge,
		Laengenvorgabe, umgebung)
	gegenkante, gegenkanteDa := massangabe(ctx, feld.Feld, "gegenkante",
		roheGegenkante, Kantenvorgabe, umgebung)

	// **Eine Länge von null ist keine Länge** — `is_setw`.
	laengeDa = laengeDa && (laenge > 0.001 || laenge < -0.001)

	switch {
	case !gegenkanteDa:
		gegenkante = kante + laenge

	case !laengeDa:
		laenge = gegenkante - kante

	case !kanteDa:
		kante = gegenkante - laenge

	default:
		mitte := kante + gegenkante
		kante = (mitte - laenge) / 2
		gegenkante = (mitte + laenge) / 2
	}

	return kante, laenge, gegenkante
}

// massangabe wertet eine Maßangabe aus und sagt, ob sie dasteht.
//
// # Hier stand die Abweichung, und der Fall ist eingetreten
//
// „Ein Skript zählt als gesetzt und wiegt null" stand hier bis
// #449, mit
// dem Satz daneben: „Wer sie hier braucht, findet den nächsten Fall — dann ist
// dieser Aufruf der Ort."
//
// **Der Fall ist `Green / A-Tyvar Kell`.** Die Vorlage rechnet die Stelle
// jeder Treuemarke über `card_style.text.top + …layout.blocks[N].middle`
// (`magic-m15-mainframe-planeswalker.mse-style`, Zeile 182). Das `top:` des
// Textfeldes ist ein Skript, `card_style.text.top` war damit **null**, und die
// Marken standen oben links über Titel und Bild.
//
// Über `Full-Magic-Pack` sind **4.566 von 28.420** Kantenangaben Skripte —
// jede sechste; gezählt an der Zeile in den Dateien `style` (§10).
//
// # Kein Knoten
//
// Liest eine Kante selbst `card_style`, bekommt sie im ersten Durchgang den
// Stand von vorher. Das ist der zweistufige Satz, den MSE genauso fährt
// (`Style::update`, `src/data/field.cpp`, Zeilen 141–151, Fassung f9d9356 vom
// 2021-07-06).
//
// # Warum die Befunde hier verfallen
//
// Aus demselben Grund wie bei [zahlVorSatz]: dieselbe Angabe wird von der
// Zeichenstufe ein zweites Mal ausgerechnet und **dort** gemeldet. Zweimal
// gemeldet läse sich wie zwei Mängel.
func massangabe(
	ctx context.Context, feld, angabe, roh string, vorwert float64,
	umgebung *msescript.Umgebung,
) (float64, bool) {
	if strings.TrimSpace(roh) == "" {
		return 0, false
	}

	zahl, _ := masszahl(ctx, feld, angabe, roh, vorwert, umgebung)

	return zahl, true
}
