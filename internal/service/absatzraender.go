package service

// **`<margin:links:rechts:oben>` — und MSEs Fehlgriff darin**
// (#433,
// [0027](../../docs/entscheidungen/0027-mses-margin-fehlgriff-wird-nachgebaut.md)).

import (
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// margenAus liest die drei Zahlen einer `<margin:…>`-Auszeichnung.
//
// # MSE greift daneben, und das wird nachgebaut
//
//	text.substr(colon  + 1, colon2 - colon  - 2).ToDouble(&m.left);
//	text.substr(colon2 + 1, colon3 - colon2 - 2).ToDouble(&m.right);
//	text.substr(colon3 + 1, pos    - colon3 - 2).ToDouble(&m.top);
//
// `src/render/text/element.cpp`, Zeilen 181–183, Fassung f9d9356 vom
// 2021-07-06. Die Länge zieht **2** ab, wo für ein durch `:` beendetes Feld
// **1** richtig wäre — die ersten beiden Zahlen verlieren ihr letztes Zeichen.
// Beim dritten stimmt es, weil `pos` schon hinter dem `>` steht.
//
// Für `<margin:130:0:0>` heißt das: links **13**, rechts leer (bleibt null),
// oben 0.
//
// **Nachgebaut und nicht behoben**, weil das Ziel Parität ist: ein
// „richtiger" Rand rückte den Text um das Zehnfache ein und wäre von MSEs Bild
// weiter entfernt als gar kein Rand. Der Grund steht ausführlich in Vermerk
// 0027.
func margenAus(angabe string) *render.Margen {
	_, felder, ist := strings.Cut(angabe, ":")
	if !ist {
		return nil
	}

	// **Nachgestellt wird die Rechnung, nicht ihr Ergebnis.** Wer hier
	// `TrimSuffix` auf die Ziffer schriebe, träfe denselben Wert und
	// verlöre den Grund; die Länge ist das, was MSE falsch rechnet.
	teile := strings.SplitN(felder, ":", 3)
	if len(teile) != 3 {
		return nil
	}

	margen := render.Margen{
		Links:  gekuerzteZahl(teile[0]),
		Rechts: gekuerzteZahl(teile[1]),
		Oben:   zahlOderNull(teile[2]),
	}

	return &margen
}

// gekuerzteZahl liest eine Zahl **ohne ihr letztes Zeichen** — MSEs Fehlgriff.
func gekuerzteZahl(roh string) float64 {
	if roh == "" {
		return 0
	}

	return zahlOderNull(roh[:len(roh)-1])
}

// zahlOderNull liest eine Zahl; was keine ist, ist null.
//
// **Wie `wxString::ToDouble`**: es meldet einen Fehlschlag und lässt den Wert
// stehen, und der stand auf null (`Margins m = {0.,0.,0.};`).
func zahlOderNull(roh string) float64 {
	zahl, err := strconv.ParseFloat(strings.TrimSpace(roh), 64)
	if err != nil {
		return 0
	}

	return zahl
}
