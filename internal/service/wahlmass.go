package service

import (
	"context"
	"image"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// istWahlart sagt, ob ein Feld über MSE2s Wahlbetrachter läuft — eine Wahl,
// ein Ja/Nein-Feld (`BooleanStyle : ChoiceStyle`,
// `src/data/field/boolean.hpp`, Zeile 32, Fassung f9d9356 vom 2021-07-06)
// oder eine Mehrfachwahl.
func istWahlart(art domain.FieldKind) bool {
	return art == domain.KindChoice || art == domain.KindBoolean || art == domain.KindMultipleChoice
}

// wahlmass ist `content_width` und `content_height` einer Wahl: das Maß ihres
// **vorbereiteten Bilds**
// (#2303).
//
//	if (style.render_style & RENDER_IMAGE) {
//	  …
//	  if (img.update(ctx) && img.isReady()) {
//	    get_options(dc, viewer, style, img_options);
//	    img.generateCached(img_options, …, &size);
//	    style.content_width = size.width; style.content_height = size.height;
//
// `prepare_choice_viewer`, `src/render/value/choice.cpp`, Zeilen 35–58;
// die Optionen aus `get_options`, Zeilen 111–124: ohne Textanteil der
// Kasten in ganzen Punkten und `ASPECT_FIT`, außer die Ausrichtung sagt
// `stretch`; mit Textanteil die Originalgröße. Eine Mehrfachwahl als Liste
// bereitet nichts vor (`src/render/value/multiple_choice.cpp`, Zeilen
// 19–22). Fassung f9d9356 vom 2021-07-06.
//
// **Ohne Bild bleibt das Maß null** — `content_width(0.0)`
// (`src/data/field/choice.cpp`, Zeile 168). Hier stand bis #2303 der Text:
// eine Wahl mit Schrift wurde gemessen wie ein Textfeld, eine mit Bild bekam
// ihren Kasten, und ein Kasten aus einem Skript ergab null.
//
// **Portierungsmarke adapted**: MSE2 übernimmt ein Maß nur, wenn sich das
// Bild geändert hat, und hält es sonst über die Karten hinweg; hier wird je
// Karte gemessen. Das unterscheidet sich nur, wenn dasselbe Bild in einem
// anderen Kasten steht. Und die Drehung eines Felds (`opts.angle`) geht hier
// nicht ins Maß ein.
func wahlmass(
	ctx context.Context, feld *msepaket.Feldstil, wert string, vorlage Vorlage,
	bilder render.Bildquelle, umgebung *msescript.Umgebung,
) Inhaltsmass {
	bild, text, liste := wahldarstellung(feld.Darstellung)
	if !bild || (liste && feldart(vorlage, feld.Feld) == domain.KindMultipleChoice) {
		return Inhaltsmass{}
	}

	baum, gerechnet := wahlbaum(ctx, feld, wert, umgebung)
	if !gerechnet {
		return Inhaltsmass{}
	}

	var (
		gefragt   image.Point
		einpassen bool
	)

	if !text {
		_, breite, _ := gerechneteKanten(ctx, feld, umgebung, feld.Links, feld.Breite, feld.Rechts)
		_, hoehe, _ := gerechneteKanten(ctx, feld, umgebung, feld.Oben, feld.Hoehe, feld.Unten)
		gefragt = image.Point{X: int(breite), Y: int(hoehe)}

		ausrichtung := ausrichtungstext(ctx, feld, umgebung)
		einpassen = strings.TrimSpace(ausrichtung) != "" && !strings.Contains(ausrichtung, "stretch")
	}

	symbole := SymbolquelleAusAblage(vorlage.Bilderverzeichnis, symbolquelle(bilder), vorlage.Setpaket)

	mass, da := render.NeuerRasterer(bilder).ErzeugtesMass(
		alsZeichenbaum(baum, vorlage.Stil.Paket, symbole), gefragt, einpassen)
	if !da {
		return Inhaltsmass{}
	}

	return Inhaltsmass{Breite: float64(mass.X), Hoehe: float64(mass.Y)}
}

// wahldarstellung liest `render style:` — ob ein Bild, ein Text und eine
// Liste darin sind (`IMPLEMENT_REFLECTION_ENUM(ChoiceRenderStyle)`,
// `src/data/field/choice.cpp`, Zeilen 245–257, Fassung f9d9356 vom
// 2021-07-06). **Ohne Angabe gilt `text`** (Zeile 164).
func wahldarstellung(roh string) (bild, text, liste bool) {
	switch strings.TrimSpace(roh) {
	case "image", "image hidden":
		return true, false, false
	case "both":
		return true, true, false
	case "image checklist", "image list":
		return true, false, true
	case "both checklist", "both list":
		return true, true, true
	default:
		return false, true, false
	}
}

// wahlbaum ist das Stilbild einer Wahl für ihren Wert — wie MSE2 es baut:
// ein gerechnetes `image:`, sonst die Wahlbilder als Nachschlagetafel mit
// `or else nil` (`ChoiceStyle::initImage`, `src/data/field/choice.cpp`,
// Zeilen 171–198, Fassung f9d9356 vom 2021-07-06). Ein Wert ohne Wahlbild
// ist `nil` und damit ein leeres Bild, kein fehlendes.
//
// **gerechnet ist falsch, wo MSE2 das Bild nie neu rechnet** — bei einem
// festen `image:` meldet `update` keine Änderung (`src/script/image.cpp`,
// Zeilen 44–46), ohne Bild und Wahlbilder gibt es nichts, und ein Skript, das
// scheitert, wirft.
func wahlbaum(
	ctx context.Context, feld *msepaket.Feldstil, wert string, umgebung *msescript.Umgebung,
) (msescript.Bildbaum, bool) {
	eigene := umgebung.Darueber()
	eigene.Setze("input", msescript.Text(wert))

	if roh := strings.TrimSpace(feld.Bild); roh != "" {
		if !bildGerechnet(roh) {
			return msescript.Bildbaum{}, false
		}

		baum, befunde := bildangabe(ctx, feld.Feld, "image", roh, eigene)

		return baum, len(befunde) == 0
	}

	if len(feld.Wahlbilder) == 0 {
		return msescript.Bildbaum{}, false
	}

	baum, befunde := wahlbild(ctx, feld, wert, eigene)

	return baum, len(befunde) == 0
}
