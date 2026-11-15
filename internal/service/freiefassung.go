package service

import (
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// freieFassung baut die Fassung eines `symbol_variation` ohne `variation:`
// aus seinen Argumenten — MSE2s „custom variation":
//
//	SCRIPT_PARAM(double, border_radius);
//	SCRIPT_OPTIONAL_PARAM_(String, fill_type);
//	if (fill_type == _("solid") || fill_type.empty()) { … }
//	else if (fill_type == _("linear gradient")) { … }
//	else if (fill_type == _("radial gradient")) { … }
//	else throw ScriptError(_("Unknown fill type for symbol_variation: ") + fill_type);
//
// `src/script/functions/image.cpp`, Zeilen 171–200, Fassung f9d9356 vom
// 2021-07-06. **Zwölf Aufrufe im Pack** schreiben so, die meisten mit
// Verlauf; bis dahin scheiterten sie hier, und die Karte zeigte kein
// Seltenheitssymbol (gefunden am Orakel zu #2303).
//
// **Was MSE2 verlangt, wird verlangt**: `SCRIPT_PARAM` wirft, wenn ein
// Argument fehlt, und eine unbekannte Füllart ebenso.
func freieFassung(ruf msescript.Ruf, pfad string) (msescript.Wert, error) {
	rand, err := fassungszahl(ruf, "border_radius")
	if err != nil {
		return msescript.Nichts, err
	}

	art := "solid"
	if wert, da := ruf.Benannter("fill_type"); da && strings.TrimSpace(wert.AlsText()) != "" {
		art = strings.TrimSpace(wert.AlsText())
	}

	var farben, zahlen []string

	switch art {
	case "solid":
		farben = []string{"fill_color", "border_color"}
	case "linear gradient":
		farben = []string{"fill_color_1", "border_color_1", "fill_color_2", "border_color_2"}
		zahlen = []string{"center_x", "center_y", "end_x", "end_y"}
	case "radial gradient":
		farben = []string{"fill_color_1", "border_color_1", "fill_color_2", "border_color_2"}
	default:
		return msescript.Nichts, ruf.FehlerAn("symbol_variation kennt die Füllart " + art + " nicht")
	}

	eintraege := []msescript.Eintrag{
		{Name: "symbol", Wert: msescript.Text(pfad)},
		{Name: "fill_type", Wert: msescript.Text(art)},
		{Name: "border_radius", Wert: msescript.Zahl(rand)},
	}

	for _, name := range farben {
		// **Auch als Text** (J05, #2269): MSE liest das Argument als `Color`
		// und wandelt eine Zeichenkette über `parse_color`
		// (`src/script/functions/image.cpp`, Zeile 178; `src/gfx/color.cpp`,
		// Zeile 48, Fassung f9d9356 vom 2021-07-06) — derselbe Weg wie
		// [gedeuteteWertfarbe].
		wert, da := ruf.Benannter(name)
		farbe, istFarbe := gedeuteteWertfarbe(wert)

		if !da || !istFarbe {
			return msescript.Nichts, ruf.FehlerAn("symbol_variation braucht eine Farbe »" + name + "«")
		}

		eintraege = append(eintraege, msescript.Eintrag{Name: name, Wert: msescript.Farbwert(
			msescript.Farbe{R: farbe.R, G: farbe.G, B: farbe.B, A: farbe.A})})
	}

	for _, name := range zahlen {
		zahl, err := fassungszahl(ruf, name)
		if err != nil {
			return msescript.Nichts, err
		}

		eintraege = append(eintraege, msescript.Eintrag{Name: name, Wert: msescript.Zahl(zahl)})
	}

	return msescript.Wert{Art: msescript.ArtBild, Text: "symbol_variation", Eintraege: eintraege}, nil
}

// fassungszahl liest ein Zahlargument, das `symbol_variation` verlangt.
func fassungszahl(ruf msescript.Ruf, name string) (float64, error) {
	wert, da := ruf.Benannter(name)
	if !da {
		return 0, ruf.FehlerAn("symbol_variation braucht »" + name + "«")
	}

	zahl, err := wert.AlsZahl()
	if err != nil {
		return 0, ruf.FehlerAn("symbol_variation: »" + name + "« ist keine Zahl")
	}

	return zahl, nil
}
