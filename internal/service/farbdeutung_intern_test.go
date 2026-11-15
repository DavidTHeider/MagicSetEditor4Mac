package service

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Eine Farbangabe wird gedeutet, wie MSE sie deutet.**
//
//	optional<Color> parse_color(const String& v) {
//	  UInt r,g,b,a;
//	  if (wxSscanf(v.c_str(),_("rgb(%u,%u,%u)"),&r,&g,&b))        return Color(r,g,b);
//	  else if (wxSscanf(v.c_str(),_("rgba(%u,%u,%u,%u)"),&r,&g,&b,&a)) return Color(r,g,b,a);
//	  else if (v.size() > 0 && v[0] == '#') { … }
//	  else if (v == _("transparent")) return Color(0,0,0,0);
//	  else { wxColour c = wxTheColourDatabase->Find(v); … }
//	}
//
// `parse_color`, `src/gfx/color.cpp`, Zeilen 42–74, Fassung f9d9356 vom
// 2021-07-06.
//
// # Was das gekostet hat
//
// Hier stand ein `switch` über zwei Namen — `white` und `red` —, und alles
// andere war Schwarz. Gezählt über die Vorlagen von `M15-Magic-Pack` steht
// `color:` **567-mal** als `rgb(…)`, 236-mal als `rgb(255,255,255)` und dazu
// mit `grey` und `rgba(0,0,0,0)`: jede dieser Angaben wurde schwarz gezeichnet
// (#270).
func TestEineFarbangabeWirdGedeutet(t *testing.T) {
	faelle := []struct {
		angabe   string
		erwartet render.Farbe
	}{
		{"rgb(1,2,3)", render.Farbe{R: 1, G: 2, B: 3, A: 255}},
		{"rgb( 1 , 2 , 3 )", render.Farbe{R: 1, G: 2, B: 3, A: 255}},
		{"rgba(1,2,3,4)", render.Farbe{R: 1, G: 2, B: 3, A: 4}},
		{"white", render.Farbe{R: 255, G: 255, B: 255, A: 255}},
		{"black", render.Farbe{A: 255}},
		{"red", render.Farbe{R: 255, A: 255}},
		{"grey", render.Farbe{R: 128, G: 128, B: 128, A: 255}},
		{"gray", render.Farbe{R: 128, G: 128, B: 128, A: 255}},
		// **`transparent` ist ein eigener Fall**, und nicht Schwarz mit Alpha
		// 255: `rgba(0,0,0,0)` steht 28-mal im Pack, meist als „diese Ebene
		// zeichnet nichts".
		{"transparent", render.Farbe{}},
		{"#0a0b0c", render.Farbe{R: 10, G: 11, B: 12, A: 255}},
		{"#0a0b0c0d", render.Farbe{R: 10, G: 11, B: 12, A: 13}},
		// **Drei Ziffern werden verdoppelt** — MSEs `17 * r`, also `f` → 255.
		{"#123", render.Farbe{R: 17, G: 34, B: 51, A: 255}},
		{"#1234", render.Farbe{R: 17, G: 34, B: 51, A: 68}},
		// **Unbekanntes ist Schwarz**, wie MSEs `Color()` nach der Warnung.
		// `wxSscanf` schlägt fehl, und dann greift der nächste Zweig; keiner
		// von ihnen trifft, und `parse_color` gibt `nullopt`.
		{"rgb(1,2)", render.Farbe{A: 255}},
		{"rgb(1,2,3", render.Farbe{A: 255}},
		{"rgb(1,x,3)", render.Farbe{A: 255}},
		{"rgb(300,0,0)", render.Farbe{A: 255}},
		{"rgba(1,2,3)", render.Farbe{A: 255}},
		{"#zz0011", render.Farbe{A: 255}},
		{"#zzzz", render.Farbe{A: 255}},
		{"knallbunt", render.Farbe{A: 255}},
		{"#12345", render.Farbe{A: 255}},
		{"", render.Farbe{A: 255}},
	}

	for _, fall := range faelle {
		t.Run(fall.angabe, func(t *testing.T) {
			if gedeutet := farbeAus(fall.angabe); gedeutet != fall.erwartet {
				t.Errorf("%q wird %v, erwartet %v", fall.angabe, gedeutet, fall.erwartet)
			}
		})
	}
}
