// **Der Schnellfilter der Kartenliste** — Werkstück U5 von Phase 5.1
// (#1416).
//
// Portiert aus MSE2: der Zerteiler aus `parse_quicksearch_query`
// (`src/data/filter.cpp:1–41`), das Treffen aus `QuickFilterQuery::match`
// und `match_quicksearch_query` (`src/data/filter.hpp:38–62`) und
// `Card::contains` (`src/data/card.cpp:55–61`), Fassung f9d9356 vom
// 2021-07-06.
//
// # Die Sprache, in einem Satz
//
// Leerraum trennt Teile, `"…"` hält zusammen, `feld:wert` bindet einen Teil
// an die Felder, deren Name `feld` enthält, und ein führendes `-` verneint.
// Eine Karte bleibt, wenn **jeder** Teil so ausgeht, wie er soll: getroffen,
// oder verneint und nicht getroffen. Verglichen wird ohne Rücksicht auf
// Groß- und Kleinschreibung, als Teilwort, und gegen den [Sichttext].
//
// # Was die Notizen angeht
//
// MSE prüft sie eigens unter dem Schlüssel `notes` (`card.cpp:59`). Hier sind
// sie ein Wert wie jeder andere ([Notizenfeld]) — die Schleife über die Werte
// sieht sie mit, unter demselben Schlüssel. Dasselbe Ergebnis, eine Regel.

package domain

import (
	"strings"
	"unicode"
)

// Filterteil ist ein Teil einer Anfrage.
type Filterteil struct {
	// Typ ist der Feldbezug — leer heißt: jedes Feld.
	Typ string

	// Wert ist, was im Feld stehen soll.
	Wert string

	// Soll sagt, ob der Teil treffen soll (true) oder gerade nicht (false).
	Soll bool
}

// ZerlegeSchnellfilter liest eine Anfrage in ihre Teile.
func ZerlegeSchnellfilter(anfrage string) []Filterteil {
	var teile []Filterteil

	teil := Filterteil{Soll: true}
	inAnfuehrung := false

	abschliessen := func() {
		if teil.Wert != "" {
			teile = append(teile, teil)
		}

		teil = Filterteil{Soll: true}
	}

	for _, zeichen := range anfrage {
		switch {
		case unicode.IsSpace(zeichen) && !inAnfuehrung:
			abschliessen()
		case zeichen == '"':
			inAnfuehrung = !inAnfuehrung
		case zeichen == ':' && teil.Typ == "" && !inAnfuehrung:
			teil.Typ = teil.Wert
			teil.Wert = ""
		case zeichen == '-' && teil.Wert == "" && teil.Typ == "" && !inAnfuehrung:
			teil.Soll = false
		default:
			teil.Wert += string(zeichen)
		}
	}

	abschliessen()

	return teile
}

// trifft sagt, ob dieser Teil auf ein Feld mit seinem Wert passt —
// `QuickFilterQuery::match`.
func (t Filterteil) trifft(feld, wert string) bool {
	return (t.Typ == "" || enthaeltOhneGross(feld, t.Typ)) && enthaeltOhneGross(wert, t.Wert)
}

// enthaeltOhneGross ist MSEs `find_i`: Teilwort, Groß und Klein egal.
func enthaeltOhneGross(heuhaufen, nadel string) bool {
	return strings.Contains(strings.ToLower(heuhaufen), strings.ToLower(nadel))
}

// enthaelt ist `Card::contains` — irgendein Wert passt.
func enthaelt(karte CardRow, teil Filterteil) bool {
	for _, wert := range karte.Values {
		if teil.trifft(wert.Field, Sichttext(wert.Text)) {
			return true
		}
	}

	return false
}

// SchnellfilterTrifft sagt, ob eine Karte durch den Filter kommt —
// `match_quicksearch_query`: jeder Teil geht so aus, wie er soll.
func SchnellfilterTrifft(teile []Filterteil, karte CardRow) bool {
	for _, teil := range teile {
		if enthaelt(karte, teil) != teil.Soll {
			return false
		}
	}

	return true
}
