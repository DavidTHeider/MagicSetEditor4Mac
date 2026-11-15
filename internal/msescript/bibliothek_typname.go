package msescript

import (
	"context"
	"math"
	"strings"
)

// **`type_name` nennt die Art eines Werts, `reverse` dreht eine Zeichenkette
// um.**
//
// Zwei kleine Funktionen aus `src/script/functions/basic.cpp` (Fassung f9d9356
// vom 2021-07-06), die zusammen **272** Befunde über die 190 Stile von
// `M15-Magic-Pack` tragen — beide auf dem **Bildweg**, wo aus ihnen ein
// Dateiname entsteht
// (#270).
func init() {
	eintragen("type_name", typname)
	eintragen("reverse", umgedreht)
}

// typname ist `type_name`.
//
//	SCRIPT_FUNCTION(type_name) {
//	  SCRIPT_PARAM_C(ScriptValueP, input);
//	  SCRIPT_RETURN(input->typeName());
//	}
//
// Zeilen 231–234.
//
// **Die Namen sind die englischen von MSE.** Die Vorlagen vergleichen sie
// (`type_name(x) != type_name(nil)`), und dort täte es jeder Name — aber
// `magic.mse-game` schreibt an anderer Stelle `== "nil"`, und dann zählt die
// Zeichenkette selbst.
func typname(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("type_name"); err != nil {
		return Nichts, err
	}

	return Text(artname(ruf.Stellung[0])), nil
}

// artname gibt MSEs `typeName()` zu einem Wert.
//
// **Eine Zeichenkette nennt ihren Inhalt mit:**
//
//	return _TYPE_("string") + _(" (\"") +
//	       (value.size() < 30 ? value : value.substr(0,30) + _("...")) + _("\")");
//
// `ScriptString::typeName`, `src/script/value.cpp`, Zeile 324. Die Grenze ist
// `< 30`, nicht `<= 30` — bei genau dreißig Zeichen kürzt MSE.
//
// **Und eine Zahl ist `integer` oder `double`**, je nachdem, ob sie ganz ist:
// MSE führt zwei Werttypen, dieses Programm einen (Zeilen 223 und 289).
//
// Fassung f9d9356 vom 2021-07-06.
func artname(wert Wert) string {
	switch wert.Art {
	case ArtZahlwert:
		if wert.Zahl == math.Trunc(wert.Zahl) {
			return "integer"
		}

		return "double"

	case ArtWahrheitswert:
		return "boolean"

	case ArtTextwert:
		return "string (\"" + gekuerzterInhalt(wert.Text) + "\")"

	case ArtFunktion:
		return "function"

	case ArtAbbildung:
		return "collection"

	case ArtFarbe:
		return "color"

	// **»date« und nicht »string«** — eine Vorlage, die den Typnamen
	// vergleicht, liefe sonst anders. Der Grund für die eigene Art steht in
	// `bibliothek_datum.go`.
	case ArtDatum:
		return "date"

	default:
		return "nil"
	}
}

// gekuerzterInhalt schneidet nach dreißig Zeichen ab.
func gekuerzterInhalt(text string) string {
	zeichen := []rune(text)
	if len(zeichen) < 30 {
		return text
	}

	return string(zeichen[:30]) + "..."
}

// umgedreht ist `reverse`.
//
// **In Zeichen, nicht in Bytes.** MSE rechnet in `wxChar`; ein Kartenname mit
// Umlaut gäbe byteweise Bruchstücke.
func umgedreht(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("reverse"); err != nil {
		return Nichts, err
	}

	zeichen := []rune(ruf.Stellung[0].AlsText())

	var bau strings.Builder

	for stelle := len(zeichen) - 1; stelle >= 0; stelle-- {
		bau.WriteRune(zeichen[stelle])
	}

	return Text(bau.String()), nil
}
