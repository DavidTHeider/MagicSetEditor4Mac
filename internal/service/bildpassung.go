package service

// **Ein Wahlbild wird eingepasst, ein Bildfeld gestreckt.**
//
// MSE wählt den Zeichner nach dem Typ des Feldes, und die beiden gehen
// verschieden mit dem Seitenverhältnis um:
//
//	// Auswahlfeld, `render style: image`:
//	opts.width  = (int) rot.trX(style.width);
//	opts.height = (int) rot.trY(style.height);
//	opts.preserve_aspect = (style.alignment & ALIGN_STRETCH) ? ASPECT_STRETCH : ASPECT_FIT;
//
// `get_options`, `src/render/value/choice.cpp`, Zeilen 111–124, und
//
//	// Bildfeld — das Kunstfenster:
//	if (image_load_file(image, *image_file)) {
//	  image.Rescale(w, h);
//	}
//
// `ImageValueViewer::draw`, `src/render/value/image.cpp`, Zeilen 36–39 —
// Fassung f9d9356 vom 2021-07-06.
//
// # Was das gekostet hat
//
// Das Seltenheitssymbol steht in einem Feld mit `alignment: middle right` und
// `render style: image`, und sein Kasten ist breiter als hoch. Gestreckt wurde
// daraus ein Balken quer über die Typzeile — sichtbar auf jeder Karte
// (#270).

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// bildpassung sagt, ob dieses Feld sein Bild einpasst — nil heißt strecken.
//
// Drei Bedingungen, jede aus dem Quelltext:
//
//  1. Das Feld ist eine **Auswahl** — nur der `ChoiceValueViewer` fragt nach
//     `preserve_aspect`.
//  2. Es zeichnet **kein** Text daneben: `else if (style.render_style &
//     RENDER_TEXT)` lässt die Maße dann ungesetzt und damit die Originalgröße.
//  3. Seine Ausrichtung sagt nicht `stretch`.
func bildpassung(
	ctx context.Context, feld *msepaket.Feldstil, vorlage Vorlage, umgebung *msescript.Umgebung,
) *render.Bildpassung {
	if !istAuswahlfeld(vorlage, feld.Feld) {
		return nil
	}

	// **`render style:` zählt beide Wörter.** `render style: both` zeichnet Bild
	// **und** Text, und dann gilt die Originalgröße statt einer Einpassung in
	// den Kasten — das ist bei MSE der Zweig, der `opts.width` gar nicht setzt.
	darstellung := strings.Fields(feld.Darstellung)
	if !enthaelt(darstellung, "image") || enthaelt(darstellung, "text") ||
		enthaelt(darstellung, "both") {
		return nil
	}

	ausrichtung := ausrichtungstext(ctx, feld, umgebung)

	// **Ohne `alignment:` gilt `stretch`, nicht „links oben".**
	//
	//	ChoiceStyle::ChoiceStyle(…)
	//	  : alignment(ALIGN_STRETCH)
	//
	// `src/data/field/choice.cpp`, Zeile 167, Fassung f9d9356 vom 2021-07-06.
	// Die **Vorgabe** eines Auswahlstils ist Strecken; erst ein
	// ausdrückliches `alignment:` ohne `stretch` macht daraus eine Einpassung.
	//
	// **Was das gekostet hat:** `corners` zeichnet ein 5 × 5 großes Bild über
	// die ganze Karte und schneidet es mit einer 375 × 523 großen Maske.
	// Eingepasst wurde daraus ein 375 × 375 großes Bild — und die Maske mit ihm
	// gestaucht. Ihre untere Rundung landete 148 Zeilen zu hoch und riss zwei
	// weiße Kerben in den Kartenrand
	// (#388).
	if strings.TrimSpace(ausrichtung) == "" || strings.Contains(ausrichtung, "stretch") {
		return nil
	}

	return &render.Bildpassung{Ausrichtung: ausrichtungAus(ausrichtung)}
}

// istAuswahlfeld sagt, ob dieses Feld im Spiel oder im Stil eine Auswahl ist.
func istAuswahlfeld(vorlage Vorlage, name string) bool {
	return feldart(vorlage, name) == domain.KindChoice
}

// istBildfeld sagt, ob dieses Feld eines ist, an dem die Karte ein Bild führt.
//
// **Die Art des Feldes und nicht sein `default:`.** MSE wählt den Zeichner nach
// dem Typ (`IMPLEMENT_VALUE_VIEWER(Image)`); ein Kunstfenster ohne Ersatzbild
// ist damit immer noch eines
// (#380).
func istBildfeld(vorlage Vorlage, feld *msepaket.Feldstil) bool {
	art := feldart(vorlage, feld.Feld)
	if art != "" {
		return art == domain.KindImage
	}

	// **Ohne Spielpaket bleibt `default:` der Marker.** Die Art eines Feldes
	// steht im `.mse-game`; fehlt es, ist sie unbekannt — und dann ist das
	// Ersatzbild der beste Hinweis, den es gibt: über alle 190 Stile von
	// `M15-Magic-Pack` steht `default:` an genau zwei Feldnamen, `image` und
	// `image 2`, und nie neben einem `image:`.
	//
	// **Zurückfallen statt aufgeben**, weil die Alternative die schlechtere
	// ist: ein fehlendes Spielpaket ließe sonst die Illustration wortlos
	// verschwinden, und die Karte sähe aus wie eine ohne.
	return feld.Standardbild != ""
}

// feldart gibt die Art eines Feldes — aus dem Spiel oder aus den Zusatzfeldern
// des Stils. **Die des Spiels zuerst**, denn ein Zusatzfeld ergänzt und ersetzt
// nicht.
func feldart(vorlage Vorlage, name string) domain.FieldKind {
	for _, felder := range [][]domain.Field{vorlage.Kartenfelder, vorlage.Stil.Zusatzfelder} {
		for i := range felder {
			if felder[i].Name == name {
				return felder[i].Kind
			}
		}
	}

	return ""
}

// einzeiligesFeld sagt, ob ein Textfeld `multi line: false` trägt — dann
// bricht MSE es nicht um (siehe [render.Satz.Einzeilig]; J05, #2269).
//
// **Ohne Spielfeld nicht**: ein Feld ohne Definition gibt es in einem echten
// Paket nicht, und dort bliebe es beim Umbruch, wie es bis dahin stand.
func einzeiligesFeld(vorlage Vorlage, name string) bool {
	for _, felder := range [][]domain.Field{vorlage.Kartenfelder, vorlage.Stil.Zusatzfelder} {
		for i := range felder {
			if felder[i].Name == name {
				return felder[i].Kind == domain.KindText && !felder[i].Mehrzeilig
			}
		}
	}

	return false
}

// enthaelt sagt, ob ein Wort in der Liste steht.
func enthaelt(woerter []string, gesucht string) bool {
	for _, wort := range woerter {
		if wort == gesucht {
			return true
		}
	}

	return false
}
