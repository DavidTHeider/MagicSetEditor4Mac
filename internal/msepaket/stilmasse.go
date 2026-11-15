package msepaket

import "strings"

// **Die Maße, die ein Feldstil seinem Satz mitgibt** — Polster und
// Zeilenhöhen.
//
// Sie stehen hier und nicht in `stilschrift.go`, weil sie eine andere Frage
// beantworten: dort steht, **womit** ein Feld schreibt — hier, **wie eng**.

// Polsterangabe ist der `padding …:`-Rand eines Feldstils.
//
// **Roh und nicht gerechnet.** Die Angaben dürfen Skripte sein; ausgerechnet
// wird in `internal/service`, wo eine Umgebung dasteht (§2).
type Polsterangabe struct{ Links, Rechts, Oben, Unten string }

// Zeilenhoehenangabe sind `line height soft/hard/line`.
//
// Sie sind **Faktoren** auf die Zeilenhöhe, keine Abstände: MSE rechnet
// `line.top += line.line_height * multiplier`
// (`src/render/text/viewer.cpp`, Zeilen 678–681, Fassung f9d9356 vom
// 2021-07-06).
type Zeilenhoehenangabe struct {
	Weich, Hart, Linie string

	// WeichMax, HartMax und LinieMax sind `line height … max` — bis dorthin
	// darf ein zu kurzer Zug gedehnt werden. `hard max` und `line max` stehen
	// je 271-mal im Pack, `soft max` 22-mal.
	WeichMax, HartMax, LinieMax string

	// Absatz ist `paragraph height:` — die Höhe eines Absatzes, wenn eine
	// Vorlage sie festlegt (`REFLECT(paragraph_height)`,
	// `src/data/field/text.cpp`, Zeile 164, Fassung f9d9356 vom 2021-07-06).
	//
	// **Beim Setzen legt sie jeden Absatz in ein eigenes Band**
	// ([render.Satz.Absatzhoehe],
	// #479);
	// die Vorlagen lesen sie zudem über `card_style.<feld>.paragraph_height`.
	Absatz string
}

// Randangabe sind die vier Randbreiten eines **Farbfeldstils**.
//
//	REFLECT(left_width); REFLECT(right_width);
//	REFLECT(top_width);  REFLECT(bottom_width);
//
// `IMPLEMENT_REFLECTION(ColorStyle)`, `src/data/field/color.cpp`, Zeilen 69–72,
// Fassung f9d9356 vom 2021-07-06.
//
// **Roh und nicht gerechnet**, wie [Polsterangabe]: `magic-modules` schreibt
// `left width: { … }` am Prägesiegel.
type Randangabe struct{ Links, Rechts, Oben, Unten string }

// massangabeAus liest ein Maß in beiden Schreibweisen.
//
// **Einzeilig oder als `script:`-Block.** Steht ein Wert da, gilt er; sonst
// wird der eingerückte Block gelesen und in geschweifte Klammern gesetzt,
// damit die Deutung ihn wie jedes andere Skript behandelt.
//
// Die Klammern sind kein Kunstgriff, sondern die Gleichmachung der beiden
// Formen: `left: { … }` und `left:` mit `script:` darunter sind bei MSE
// dasselbe (`Scriptable<double>` liest beide über `ScriptValue`).
func massangabeAus(knoten *Knoten) string {
	// **Kein Knoten ist keine Angabe.** [Knoten.Ersten] gibt `nil`, wenn der
	// Schlüssel fehlt; die Aufrufer an `left:` und `font: color:` prüfen das
	// im Zweig davor, der an `image:` eines Symbols nicht
	// (#405).
	if knoten == nil {
		return ""
	}

	if strings.TrimSpace(knoten.Wert) != "" {
		return knoten.Wert
	}

	skript := knoten.Ersten("script")
	if skript == nil {
		return knoten.Wert
	}

	text := strings.TrimSpace(skript.Skripttext())
	if text == "" {
		return knoten.Wert
	}

	return "{\n" + text + "\n}"
}
