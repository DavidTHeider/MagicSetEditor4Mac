package render

import (
	"image"
	"image/color"
	"strings"
)

// **`symbol_variation` rastert eine Vektorzeichnung.**
//
//	return make_intrusive<SymbolToImage>(value, filename, value->last_update, v);
//
// `src/script/functions/image.cpp`, Zeile 148, Fassung f9d9356 vom 2021-07-06
// — eine der fünf Bildquellen aus `generated_image.hpp`, und die letzte, die
// hier gefehlt hat.
//
// **Die Zeichnung steht im Plan.** Wer die `.mse-symbol` liest, ist
// `internal/service`; hier kommt die Gestalt an, wie [RastereSymbol] sie
// braucht (#323).

// Symbolseite ist die Untergrenze, unter die keine Rasterung geht.
//
// **Sie ist MSEs `max(100, …)`** und keine Wahl dieses Programms
// (`SymbolToImage::generate`, `src/gfx/generated_image.cpp`, Zeile 463,
// Fassung f9d9356 vom 2021-07-06).
const Symbolseite = 100

// Symbolflaeche ist die Fläche, in der ein `symbol_variation` gerastert wird.
//
//	int size = max(100, 3*max(opt.width,opt.height));
//	if (opt.width <= 1 || opt.height <= 1) {
//	  return render_symbol(…, size, size);
//	} else {
//	  int width  = size * opt.width  / max(opt.width,opt.height);
//	  int height = size * opt.height / max(opt.width,opt.height);
//	  return render_symbol(…, width, height, false, true);
//	}
//
// Dieselbe Stelle, Zeilen 463–470.
//
// # Warum das Dreifache
//
// Weil eine Verkleinerung sauber ist und eine Vergrößerung nicht. Wer 200
// Punkte breit zeichnen will, rastert 600 und rechnet herunter.
//
// **`<= 1` und nicht `== 0`**: ein Maß von einem Punkt trägt kein
// Seitenverhältnis, und MSE fällt dort auf das Quadrat zurück.
//
// **Die Ganzzahldivision steht so da wie in C++** — erst mal, dann geteilt.
// Wer zuerst teilte, bekäme bei 21 × 21 die 0.
func Symbolflaeche(gefragt image.Point) (breite, hoehe int, kleinerErlaubt bool) {
	laengste := max(gefragt.X, gefragt.Y)
	seite := max(Symbolseite, 3*laengste)

	if gefragt.X <= 1 || gefragt.Y <= 1 {
		return seite, seite, false
	}

	return seite * gefragt.X / laengste, seite * gefragt.Y / laengste, true
}

// symbolbild rastert die Zeichnung eines `symbol_variation`-Knotens.
func (r *Rasterer) symbolbild(baum Bildbaum) (quelle image.Image, befund string) {
	if len(baum.Symbol) == 0 {
		return nil, "die Symbolzeichnung " + baum.Texte["symbol"] + " liegt nicht vor"
	}

	breite, hoehe, kleinerErlaubt := Symbolflaeche(r.gefragteGroesse)

	return FaerbeSymbol(
		RastereSymbol(baum.Symbol, breite, hoehe,
			baum.Zahlen["border_radius"], kleinerErlaubt),
		filterAus(baum)), ""
}

// filterAus wählt den Filter nach `fill type:`.
//
//	if (fill_type == _("solid") || fill_type.empty()) …
//	else if (fill_type == _("linear gradient")) …
//	else if (fill_type == _("radial gradient")) …
//
// `src/script/functions/image.cpp`, Zeilen 172–195. **Ohne Angabe gilt
// `solid`** — und eine unbekannte Angabe ebenfalls: MSE wirft dort, hier
// zeichnet das Symbol in seinen einfachen Farben, und der Unterschied wäre eine
// fehlende Ebene statt einer falschen Farbe.
//
// Fassung f9d9356 vom 2021-07-06.
func filterAus(baum Bildbaum) Symbolfilter {
	switch strings.TrimSpace(baum.Texte["fill_type"]) {
	case "linear gradient":
		return Verlaufsfilter{
			Fuellfarbe1: nrgbaVon(baum.Farben["fill_color_1"]),
			Fuellfarbe2: nrgbaVon(baum.Farben["fill_color_2"]),
			Randfarbe1:  nrgbaVon(baum.Farben["border_color_1"]),
			Randfarbe2:  nrgbaVon(baum.Farben["border_color_2"]),
			MitteX:      baum.Zahlen["center_x"], MitteY: baum.Zahlen["center_y"],
			EndeX: baum.Zahlen["end_x"], EndeY: baum.Zahlen["end_y"],
		}

	case "radial gradient":
		return Kreisfilter{
			Fuellfarbe1: nrgbaVon(baum.Farben["fill_color_1"]),
			Fuellfarbe2: nrgbaVon(baum.Farben["fill_color_2"]),
			Randfarbe1:  nrgbaVon(baum.Farben["border_color_1"]),
			Randfarbe2:  nrgbaVon(baum.Farben["border_color_2"]),
		}

	default:
		return Vollfilter{
			Fuellfarbe: nrgbaVon(baum.Farben["fill_color"]),
			Randfarbe:  nrgbaVon(baum.Farben["border_color"]),
		}
	}
}

// nrgbaVon macht aus einer Planfarbe eine Bildfarbe.
//
// **Zwei Typen für dieselbe Sache**, und das ist keine Umständlichkeit: [Farbe]
// ist eine Angabe des Plans, `color.NRGBA` die des Bildes. Die Umrechnung an
// einer Stelle zu haben heißt, sie einmal zu prüfen.
func nrgbaVon(farbe Farbe) color.NRGBA {
	return color.NRGBA{R: farbe.R, G: farbe.G, B: farbe.B, A: farbe.A}
}
