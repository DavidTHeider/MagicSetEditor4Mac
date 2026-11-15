package service

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Ein Feldstil nennt auch seine Textangaben — Polster und Zeilenhöhen.**
//
//	REFLECT(padding_left);  REFLECT(padding_right);
//	REFLECT(padding_top);   REFLECT(padding_bottom);
//	REFLECT(line_height_soft); REFLECT(line_height_hard);
//	REFLECT(line_height_line);
//	REFLECT(line_height_soft_max); … REFLECT(paragraph_height);
//
// `IMPLEMENT_REFLECTION(TextStyle)`, `src/data/field/text.cpp`, Zeilen 143–167,
// Fassung f9d9356 vom 2021-07-06.
//
// **Eine Vorlage rechnet mit den Angaben einer anderen Ebene**:
// `magic-genevensis` liest `card_style.text.line_height_hard`, um seine Stufen
// gleich hoch zu setzen — 14 Befunde über die 190 Stile von `M15-Magic-Pack`
// (#270).

// textangaben gibt die Textangaben eines Feldstils als Einträge.
//
// **Als Zahlen und nicht als Text.** Die Vorlagen rechnen damit
// (`card_style.text.line_height_hard * 2`); eine Zeichenkette dort brauchte
// überall ein `to_number`, und MSE führt sie als `Scriptable<double>`.
//
// **Was ein Skript ist, wird null.** Dieselbe Regel wie bei den Kanten
// ([kanten]): ausgerechnet wird eine Angabe beim Anordnen, und das ist eine
// Ebene später. Die Abweichung steht dort.
func textangaben(feld *msepaket.Feldstil) []msescript.Eintrag {
	polster, zeilen := feld.Polster, feld.Zeilenhoehen

	roh := []struct {
		name   string
		angabe string
	}{
		{"padding_left", polster.Links},
		{"padding_right", polster.Rechts},
		{"padding_top", polster.Oben},
		{"padding_bottom", polster.Unten},
		{"line_height_soft", zeilen.Weich},
		{"line_height_hard", zeilen.Hart},
		{"line_height_line", zeilen.Linie},
		{"line_height_soft_max", zeilen.WeichMax},
		{"line_height_hard_max", zeilen.HartMax},
		{"line_height_line_max", zeilen.LinieMax},
		{"paragraph_height", zeilen.Absatz},
	}

	eintraege := make([]msescript.Eintrag, 0, len(roh))
	for _, angabe := range roh {
		eintraege = append(eintraege,
			msescript.Eintrag{Name: angabe.name, Wert: feldmass(angabe.angabe)})
	}

	return eintraege
}
