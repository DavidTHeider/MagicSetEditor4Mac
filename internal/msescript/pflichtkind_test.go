package msescript_test

// **Ein Pflichtkind kommt aus dem Aufruf oder aus der Umgebung**
// (#473).
//
//	#define SCRIPT_PARAM(Type, name) SCRIPT_PARAM_N(Type, _(#name), name)
//	#define SCRIPT_PARAM_N(Type, str, name) \
//	    Type name = from_script<Type>(ctx.getVariable(str), str)
//
// `src/script/functions/util.hpp`, Zeilen 98–101, Fassung f9d9356 vom
// 2021-07-06. `SCRIPT_PARAM` liest aus dem **Kontext** — ein benanntes
// Argument bindet dort eine Variable, aber eine anderswo gebundene tut es
// genauso.
//
// **Das Pack lebt davon**: `magic-blends.mse-include/new-blends:91` schreibt
//
//	linear_blend := { do_linear_blend(x1: 0, x2: 0, y1: x1, y2: x2) }
//
// — eine Teilanwendung, deren `image1` und `image2` aus dem umgebenden
// Bereich kommen. Dasselbe steht in `magic-m15-split-fusable` und drei
// weiteren Stilen.
//
// **Vorher wurde ein fehlendes Kind übersprungen**, und geprüft war nur, ob
// irgendein Bild dastand: ein `masked_blend` mit zwei von drei Kindern kam
// durch, ohne dass jemand in der Umgebung nachgesehen hätte.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

func TestEinPflichtkindKommtAuchAusDerUmgebung(t *testing.T) {
	faelle := []struct {
		name, quelle string
		fehler       string
	}{
		{
			name:   "alle drei benannt",
			quelle: `masked_blend(light: "a.png", dark: "b.png", mask: "m.png")`,
		},
		{
			// Die Form aus dem Pack: eine Teilanwendung bindet die Bilder,
			// der Aufruf nennt nur die Zahlen.
			name: "aus der Umgebung",
			quelle: `[ image1: "a.png", image2: "b.png" ] +
				{ linear_blend(x1: 0, y1: 0, x2: 1, y2: 0) }`,
		},
		{
			name:   "eines fehlt ganz",
			quelle: `masked_blend(light: "a.png", dark: "b.png")`,
			fehler: "masked_blend braucht mask:",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			_, err := msescript.AuswertenQuelle(
				t.Context(), fall.quelle, msescript.Bibliothek())

			if fall.fehler == "" {
				if err != nil {
					t.Fatalf("ausrechnen: %v", err)
				}

				return
			}

			if err == nil {
				t.Fatal("kein Fehler")
			}

			if !strings.Contains(err.Error(), fall.fehler) {
				t.Errorf("der Fehler lautet %q, erwartet %q", err, fall.fehler)
			}
		})
	}
}

// **`to_image` reicht durch und baut keinen Knoten.**
//
//	SCRIPT_FUNCTION(to_image) {
//	  SCRIPT_PARAM_C(GeneratedImageP, input);
//	  return input;
//	}
//
// `src/script/functions/image.cpp`, Zeilen 26–29, Fassung f9d9356 vom
// 2021-07-06.
//
// **Gefunden durch #473**: der Saga-Stil schreibt
// `crop(to_image(…/textmask.png), …)`, und solange ein Befund neben einem Bild
// verlorenging, zeichnete das Programm die **unbeschnittene** Maske und
// meldete nichts.
func TestToImageReichtDurch(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(
		t.Context(), `to_image("a.png")`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("ausrechnen: %v", err)
	}

	baum, istBild := wert.AlsBild()
	if !istBild {
		t.Fatalf("%v ist kein Bild", wert)
	}

	// **Ein Blatt, keine Verknüpfung.** Ein eigener Knoten wäre eine
	// Verknüpfung, die nichts verknüpft — und der Rasterer müsste sie kennen.
	if baum.Knotenart != "datei" || baum.Pfad != "a.png" {
		t.Errorf("ergibt %q mit Pfad %q, erwartet das Blatt a.png",
			baum.Knotenart, baum.Pfad)
	}

	// **Was kein Bild ist, wird benannt.**
	if _, err := msescript.AuswertenQuelle(
		t.Context(), `to_image(3)`, msescript.Bibliothek()); err == nil {
		t.Error("to_image(3) hat keinen Fehler gegeben")
	}

	if _, err := msescript.AuswertenQuelle(
		t.Context(), `to_image()`, msescript.Bibliothek()); err == nil {
		t.Error("to_image() hat keinen Fehler gegeben")
	}
}
