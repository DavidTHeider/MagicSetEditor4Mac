package service

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Polster und Zeilenhöhen eines Feldstils.**
//
// Sie stehen hier und nicht in `stilplan.go`, weil sie eine eigene Frage
// beantworten: dort steht, **was** ein Feld zeichnet — hier, **wie eng** es
// gesetzt wird.

// polsterAus rechnet `padding left/right/top/bottom` aus.
//
// **Eine Angabe, die sich nicht ausrechnen lässt, ist null.** Ein Polster ist
// eine Zugabe; wo keines steht, steht keines. Ein Befund je fehlender Angabe
// stünde auf jeder Karte und sagte nichts.
func polsterAus(
	ctx context.Context, feld *msepaket.Feldstil, umgebung *msescript.Umgebung,
) render.Polster {
	return render.Polster{
		Links:  masszahlOderNull(ctx, feld.Polster.Links, Polstervorgabe, umgebung),
		Rechts: masszahlOderNull(ctx, feld.Polster.Rechts, Polstervorgabe, umgebung),
		Oben:   masszahlOderNull(ctx, feld.Polster.Oben, Polstervorgabe, umgebung),
		Unten:  masszahlOderNull(ctx, feld.Polster.Unten, Polstervorgabe, umgebung),
	}
}

// zeilenfaktorenAus rechnet `line height soft/hard/line` aus.
//
// **Null heißt eins** — das entscheidet [render.Zeilenfaktoren], nicht diese
// Stelle: dort steht auch der Grund.
func zeilenfaktorenAus(
	ctx context.Context, feld *msepaket.Feldstil, umgebung *msescript.Umgebung,
) render.Zeilenfaktoren {
	return render.Zeilenfaktoren{
		Weich: masszahlOderNull(ctx, feld.Zeilenhoehen.Weich, Zeilenhoehenvorgabe, umgebung),
		Hart:  masszahlOderNull(ctx, feld.Zeilenhoehen.Hart, Zeilenhoehenvorgabe, umgebung),
		Linie: masszahlOderNull(ctx, feld.Zeilenhoehen.Linie, Zeilenhoehenvorgabe, umgebung),
		WeichMax: masszahlOderNull(ctx, feld.Zeilenhoehen.WeichMax,
			Hoechstzeilenhoehenvorgabe, umgebung),
		HartMax: masszahlOderNull(ctx, feld.Zeilenhoehen.HartMax,
			Hoechstzeilenhoehenvorgabe, umgebung),
		LinieMax: masszahlOderNull(ctx, feld.Zeilenhoehen.LinieMax,
			Hoechstzeilenhoehenvorgabe, umgebung),
	}
}

// kleinstgroesseAus rechnet `scale down to:` aus — die kleinste Schriftgröße,
// auf die ein zu langer Text verkleinert werden darf.
//
// **Null heißt „nicht verkleinern".** MSEs Vorgabe ist 100000
// (`src/data/font.cpp`, Zeile 18, Fassung f9d9356 vom 2021-07-06), also größer
// als jede Schrift — dort ist `min(size, scale_down_to)` dann `size`, und der
// Maßstab bleibt 1 (#498).
func kleinstgroesseAus(
	ctx context.Context, feld *msepaket.Feldstil, umgebung *msescript.Umgebung,
) float64 {
	return masszahlOderNull(ctx, feld.Schrift.Kleinstgroesse,
		Kleinstgroessenvorgabe, umgebung)
}

// winkelAus rechnet `angle:` aus — die Drehung eines Feldes in Grad.
//
// **Gerechnet, weil er gerechnet sein darf**: im vollen Pack stehen 33 der
// 213 Angaben als `{if …}`
// (#486).
func winkelAus(
	ctx context.Context, feld *msepaket.Feldstil, umgebung *msescript.Umgebung,
) float64 {
	return masszahlOderNull(ctx, feld.Winkel, Winkelvorgabe, umgebung)
}

// absatzhoeheAus rechnet `paragraph height:` aus.
//
// **Sie ist keine Zeilenhöhe, obwohl sie daneben steht**: die Faktoren
// verschieben eine Zeile gegen die vorige, diese Zahl legt jeden Absatz in ein
// Band fester Höhe ([render.Satz.Absatzhoehe],
// #479).
func absatzhoeheAus(
	ctx context.Context, feld *msepaket.Feldstil, umgebung *msescript.Umgebung,
) float64 {
	return masszahlOderNull(ctx, feld.Zeilenhoehen.Absatz, Absatzvorgabe, umgebung)
}

// masszahlOderNull rechnet eine Angabe aus; was nicht geht, ist null.
func masszahlOderNull(
	ctx context.Context, roh string, vorwert float64,
	umgebung *msescript.Umgebung,
) float64 {
	if strings.TrimSpace(roh) == "" {
		return 0
	}

	zahl, befund := masszahl(ctx, "", "", roh, vorwert, umgebung)
	if len(befund) > 0 {
		return 0
	}

	return zahl
}
