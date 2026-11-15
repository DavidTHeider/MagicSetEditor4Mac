package service

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **`symbol_variation` rastert das Setsymbol in einer benannten Fassung.**
//
//	FOR_EACH(v, style->variations) {
//	  if (v->name == variation) {
//	    return make_intrusive<SymbolToImage>(value, filename, value->last_update, v);
//	  }
//	}
//	throw ScriptError(_("Variation of symbol not found ('") + variation + _("')"));
//
// `src/script/functions/image.cpp`, Zeilen 141–170, Fassung f9d9356 vom
// 2021-07-06.
//
// # Warum sie hier steht und nicht in der Bibliothek
//
// Sie braucht die Fassungen des Stils und des Spiels, und `msescript` kennt
// dieses Projekt nicht (§2). Wie schon bei `exists_in_package`: die Bibliothek
// trägt die Sprache, der Aufrufer setzt hinein, was die Umgebung darüber hinaus
// weiß (#323).

// symbolfassungsfrage baut die Funktion `symbol_variation`.
//
// **Sie gibt einen Bildbaum, kein Bild.** Gerastert wird in `internal/render`
// (Vermerk 0022) — hier entsteht die Beschreibung: welcher Pfad, welche
// Randbreite, welche Farben.
func symbolfassungsfrage(fassungen []msepaket.Symbolfassung) msescript.Wert {
	return msescript.Wert{
		Art: msescript.ArtFunktion,
		Funktion: &msescript.Funktionsteil{
			Eingebaut: func(_ context.Context, ruf msescript.Ruf) (msescript.Wert, error) {
				return fassungswert(ruf, fassungen)
			},
		},
	}
}

// fassungswert sucht die Fassung und baut den Baum.
func fassungswert(ruf msescript.Ruf, fassungen []msepaket.Symbolfassung) (msescript.Wert, error) {
	pfad, feld := "", ""

	// **Das Feld kommt mit dem Wert** — MSEs `value->fieldP`
	// (`src/script/functions/image.cpp`, Zeile 161, Fassung f9d9356 vom
	// 2021-07-06). `magic.mse-game` erklärt `symbol:` und `masterpiece
	// symbol:` mit gleichnamigen, verschieden gefärbten Fassungen; ohne das
	// Feld fände die Suche die des anderen.
	if wert, da := ruf.Benannter("symbol"); da {
		pfad, feld = strings.TrimSpace(wert.AlsText()), wert.Feld
	} else if len(ruf.Stellung) > 0 {
		pfad, feld = strings.TrimSpace(ruf.Stellung[0].AlsText()), ruf.Stellung[0].Feld
	}

	// **Ohne `variation:` eine freie Fassung** aus den Argumenten
	// ([freieFassung]) — MSE2s zweiter Zweig.
	wert, benannt := ruf.Benannter("variation")
	if !benannt {
		return freieFassung(ruf, pfad)
	}

	name := strings.TrimSpace(wert.AlsText())

	gefunden, gibt := ErsteFassung(fassungen, feld, name)

	if !gibt {
		// **Eine Fassung, die es nicht gibt, ist ein Fehler** — MSE wirft
		// `Variation of symbol not found` (Zeile 169). Eine erfundene Fassung
		// zeichnete ein Symbol in falschen Farben, und niemand sähe warum (§8).
		return msescript.Nichts, ruf.FehlerAn(
			"symbol_variation kennt die Fassung " + name + " nicht")
	}

	return fassungsbild(pfad, gefunden), nil
}

// ErsteFassung sucht eine Symbolfassung — **die erste, die passt**.
//
//	FOR_EACH(v, style->variations) {
//	  if (v->name == variation) {
//	    return make_intrusive<SymbolToImage>(value, filename, value->last_update, v);
//	  }
//	}
//
// `src/script/functions/image.cpp`, Zeilen 164–170, Fassung f9d9356 vom
// 2021-07-06 — **`return` in der Schleife.**
//
// # Warum die erste und nicht die letzte
//
// Ein Stilpaket **erbt** die Fassungen des Spiels und hängt seine eigenen
// dahinter: `set_info_style.cloneFrom(game->default_set_style)` steht **vor**
// `REFLECT(set_info_style)` (`src/data/stylesheet.cpp`, Zeilen 106–113), und
// `Reader::handle(name, vector<T>&)` macht `emplace_back` ohne zu leeren
// (`src/util/io/reader.hpp`, Zeilen 218–226).
//
// **Bei gleichem Namen gewinnt damit die Fassung des Spiels**
// (#548).
//
// **Am Artefakt belegt**: die elf `special`-Karten von `White` erscheinen in
// MSE2s eigener Ausgabe weiß — die Randfarbe der **Spiel**-Fassung.
//
// # Und was noch offen ist
//
// `feld` wählt zusätzlich nach dem Setfeld; leer heißt „gleich welches". Der
// Aufrufer `symbol_variation` kann es heute **nicht** angeben, weil er den
// Symbol**pfad** bekommt und nicht das Feld, aus dem er stammt. Solange
// `symbol:` in `magic.mse-game/set_fields` vor `masterpiece symbol:` steht,
// verdeckt die Erste-gewinnt-Regel den Unterschied — latent bleibt er.
func ErsteFassung(
	fassungen []msepaket.Symbolfassung, feld, name string,
) (msepaket.Symbolfassung, bool) {
	for stelle := range fassungen {
		if fassungen[stelle].Name != name {
			continue
		}

		if feld != "" && fassungen[stelle].Feld != feld {
			continue
		}

		return fassungen[stelle], true
	}

	return msepaket.Symbolfassung{}, false
}

// fassungsbild macht aus Pfad und Fassung einen Bildwert.
//
// **Die Farben werden hier gedeutet und nicht im Rasterer.** `rgb(0,0,0)` ist
// eine Angabe der Vorlage, und was sie bedeutet, weiß diese Schicht
// (`feldangaben.go`).
func fassungsbild(pfad string, fassung msepaket.Symbolfassung) msescript.Wert {
	eintraege := []msescript.Eintrag{
		{Name: "symbol", Wert: msescript.Text(pfad)},
		{Name: "fill_type", Wert: msescript.Text(fassung.Fuellart)},
		{Name: "border_radius", Wert: msescript.Zahl(fassung.Randbreite)},
		{Name: "center_x", Wert: msescript.Zahl(fassung.MitteX)},
		{Name: "center_y", Wert: msescript.Zahl(fassung.MitteY)},
		{Name: "end_x", Wert: msescript.Zahl(fassung.EndeX)},
		{Name: "end_y", Wert: msescript.Zahl(fassung.EndeY)},
	}

	for _, farbe := range []struct{ name, angabe string }{
		{"fill_color", fassung.Fuellfarbe},
		{"border_color", fassung.Randfarbe},
		{"fill_color_1", fassung.Fuellfarbe1},
		{"fill_color_2", fassung.Fuellfarbe2},
		{"border_color_1", fassung.Randfarbe1},
		{"border_color_2", fassung.Randfarbe2},
	} {
		if strings.TrimSpace(farbe.angabe) == "" {
			continue
		}

		gedeutet := farbeAus(farbe.angabe)
		eintraege = append(eintraege, msescript.Eintrag{
			Name: farbe.name,
			Wert: msescript.Farbwert(msescript.Farbe{
				R: gedeutet.R, G: gedeutet.G, B: gedeutet.B, A: gedeutet.A,
			}),
		})
	}

	return msescript.Wert{Art: msescript.ArtBild, Text: "symbol_variation", Eintraege: eintraege}
}

// symbolquelle macht aus einer Bildquelle eine Symbolquelle, wenn sie eine ist.
//
// **Dieselbe Bauform wie bei [Dateiquelle]**: der Planbauer bekommt eine
// Bildquelle; ob die auch Symbolzeichnungen hergibt, entscheidet ihr Typ. Wer
// keine liefert, lässt die Gestalt leer — und der Rasterer benennt das.
func symbolquelle(bilder render.Bildquelle) Symbolquelle {
	if quelle, kann := bilder.(Symbolquelle); kann {
		return quelle
	}

	return nil
}
