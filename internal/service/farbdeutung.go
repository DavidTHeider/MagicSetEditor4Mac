package service

// **Eine Farbangabe deuten** — `rgb(…)`, `rgba(…)`, `#rrggbb`, ein Name.
//
// # Die Regel, nachgelesen in MSE2
//
//	optional<Color> parse_color(const String& v) {
//	  UInt r,g,b,a;
//	  if (wxSscanf(v.c_str(),_("rgb(%u,%u,%u)"),&r,&g,&b))             return Color(r,g,b);
//	  else if (wxSscanf(v.c_str(),_("rgba(%u,%u,%u,%u)"),&r,&g,&b,&a)) return Color(r,g,b,a);
//	  else if (v.size() > 0 && v[0] == '#') { … 17 * r … }
//	  else if (v == _("transparent")) return Color(0,0,0,0);
//	  else { wxColour c = wxTheColourDatabase->Find(v); … }
//	}
//
// `parse_color`, `src/gfx/color.cpp`, Zeilen 42–74, Fassung f9d9356 vom
// 2021-07-06. Was sich nicht deuten lässt, wird bei MSE `Color()` mit einer
// Warnung — also **Schwarz**, und das ist hier dieselbe Antwort.

import (
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// farbeAus deutet eine Farbangabe des Stils; ungedeutet heißt schwarz.
func farbeAus(angabe string) render.Farbe {
	if farbe, da := gedeuteteFarbe(angabe); da {
		return farbe
	}

	return render.Farbe{A: 255}
}

// gedeuteteFarbe deutet eine Farbangabe und sagt, ob sie eine war.
//
// **Das `ob` ist nicht dasselbe wie „schwarz"**: ein Farbfeld ohne Wert nimmt
// seine Vorgabe, und dafür muss der leere Text von `rgb(0,0,0)` zu
// unterscheiden sein ([farbfeldwert]).
func gedeuteteFarbe(angabe string) (render.Farbe, bool) {
	angabe = strings.TrimSpace(angabe)

	switch {
	case strings.HasPrefix(angabe, "rgba("):
		return anteilsfarbe(angabe, "rgba(", 4)

	case strings.HasPrefix(angabe, "rgb("):
		return anteilsfarbe(angabe, "rgb(", 3)

	case strings.HasPrefix(angabe, "#"):
		return hexfarbe(angabe[1:])

	// **`transparent` ist ein eigener Fall** und kommt vor der Namensliste —
	// die Farbdatenbank von wx führt ihn nicht.
	case angabe == "transparent":
		return render.Farbe{}, true
	}

	farbe, da := farbnamen[strings.ToLower(angabe)]

	return farbe, da
}

// farbnamen ist der benutzte Teil von wxWidgets' Farbdatenbank.
//
// **Ein Teil, und das steht hier, weil es die Reichweite der Aussage ist**
// (§10): `wxTheColourDatabase` führt über hundert Namen, und welche davon eine
// Vorlage benutzt, ist nicht abschließend zu wissen. Gezählt über die
// Vorlagendateien von `M15-Magic-Pack` stehen als Farbname genau vier:
// `black` (716-mal), `white` (273), `red` (263) und `grey` (2). Kommt eine
// Vorlage mit einem fünften, wird sie schwarz gezeichnet — sichtbar, und dafür
// gibt es diesen Absatz.
var farbnamen = map[string]render.Farbe{
	"black":   {A: 255},
	"white":   {R: 255, G: 255, B: 255, A: 255},
	"red":     {R: 255, A: 255},
	"green":   {G: 255, A: 255},
	"blue":    {B: 255, A: 255},
	"yellow":  {R: 255, G: 255, A: 255},
	"cyan":    {G: 255, B: 255, A: 255},
	"magenta": {R: 255, B: 255, A: 255},
	"grey":    {R: 128, G: 128, B: 128, A: 255},
	"gray":    {R: 128, G: 128, B: 128, A: 255},
}

// anteilsfarbe liest `rgb(r,g,b)` und `rgba(r,g,b,a)`.
//
// **Leerzeichen sind erlaubt**, weil `wxSscanf` mit `%u` darüber hinwegliest;
// eine Vorlage, die `rgb(0, 0, 0)` schreibt, meint Schwarz und nicht nichts.
func anteilsfarbe(angabe, kopf string, anzahl int) (render.Farbe, bool) {
	rumpf, hatKlammer := strings.CutSuffix(angabe[len(kopf):], ")")
	if !hatKlammer {
		return render.Farbe{}, false
	}

	teile := strings.Split(rumpf, ",")
	if len(teile) != anzahl {
		return render.Farbe{}, false
	}

	anteile := [4]uint8{0, 0, 0, 255}

	for stelle, teil := range teile {
		zahl, err := strconv.Atoi(strings.TrimSpace(teil))
		if err != nil || zahl < 0 || zahl > 255 {
			return render.Farbe{}, false
		}

		// #nosec G115 -- oben auf 0…255 geprüft
		anteile[stelle] = uint8(zahl)
	}

	return render.Farbe{R: anteile[0], G: anteile[1], B: anteile[2], A: anteile[3]}, true
}

// hexfarbe liest die vier Längen hinter dem `#`.
//
//	if (v.size() == 4)      … 17 * r …
//	else if (v.size() == 5) … 17 * r … a
//	else if (v.size() == 7) … parse_hex(v[1], v[2]) …
//	else if (v.size() == 9) … a
//	return nullopt;
//
// Zeilen 48–62. **Drei Ziffern werden verdoppelt** — `#f00` ist `#ff0000`, und
// `17 * 15` ist genau das.
func hexfarbe(ziffern string) (render.Farbe, bool) {
	nach := func(anzahl int) ([]uint8, bool) {
		breite := len(ziffern) / anzahl

		werte := make([]uint8, 0, anzahl)

		for stelle := range anzahl {
			zahl, err := strconv.ParseUint(ziffern[stelle*breite:(stelle+1)*breite], 16, 8)
			if err != nil {
				return nil, false
			}

			if breite == 1 {
				zahl *= 17
			}

			// #nosec G115 -- ParseUint mit Bitbreite 8, mal 17 höchstens 255
			werte = append(werte, uint8(zahl))
		}

		return werte, true
	}

	switch len(ziffern) {
	case 3, 6:
		werte, gut := nach(3)
		if !gut {
			return render.Farbe{}, false
		}

		return render.Farbe{R: werte[0], G: werte[1], B: werte[2], A: 255}, true

	case 4, 8:
		werte, gut := nach(4)
		if !gut {
			return render.Farbe{}, false
		}

		return render.Farbe{R: werte[0], G: werte[1], B: werte[2], A: werte[3]}, true
	}

	return render.Farbe{}, false
}
