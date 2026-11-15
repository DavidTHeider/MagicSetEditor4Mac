package domain

import "strings"

// ManakostenOhneKlammern macht aus `{2}{R}` die Schreibweise `2R`.
//
// # Warum diese Umrechnung in der Domäne steht
//
// Weil **zwei** fremde Quellen dieselbe Schreibweise benutzen: MTG JSON und
// Scryfall schreiben beide `{2}{R}`. `internal/mtgjson` und
// `internal/scryfall` sind Blätter und kennen einander nicht (§2) — eine Kopie
// in jedem wäre genau der Fall, vor dem `msepaket/paket.go` warnt.
//
// **Und es ist eine Aussage über MSE, nicht über einen Anbieter**: so schreibt
// MSE seine Manakosten, gleich wer sie liefert.
//
// # Woher die Regel stammt
//
// **Belegt, nicht geraten:** `add_cards_scripts` in `magic.mse-game` legt
// Karten mit `casting_cost: "W"` an — ohne geschweifte Klammern. Was in einer
// Klammer steht, bleibt zusammen: aus `{W/U}` wird `W/U`.
//
// # Was sie nicht tut
//
// **Sie prüft nicht, ob die Symbole gültig sind.** `{Zauberstab}` wird zu
// `Zauberstab`. Ob ein Symbol gezeichnet werden kann, entscheidet die
// Symbolschrift der Vorlage — und die kennt dieses Paket nicht.
func ManakostenOhneKlammern(roh string) string {
	if roh == "" {
		return ""
	}

	var bau strings.Builder

	for _, teil := range strings.Split(roh, "}") {
		bau.WriteString(strings.TrimPrefix(strings.TrimSpace(teil), "{"))
	}

	return bau.String()
}

// Manasymbole zerlegt einen Kosten in seine Symbole — **mit** Klammern.
//
// # Warum mit Klammern
//
// Weil Scryfall sie so führt: die Symbologie nennt `{W}`, nicht `W`. Wer sie
// hier abschnitte, müsste sie an jeder Vergleichsstelle wieder anfügen — und
// eine davon vergäße es.
//
// # Was sie nicht tut
//
// **Sie prüft nichts.** `{Zauberstab}` kommt als `{Zauberstab}` zurück; ob es
// das gibt, weiß dieses Paket nicht. Die Frage beantwortet die Symbologie von
// Scryfall (`service.Auskunft.Manakostenpruefung`) — und sie ist noch einmal
// eine andere als die, ob die **Vorlage** das Symbol zeichnen kann. Drei
// Fragen, drei Orte.
//
// **Ein halb geschriebenes Symbol wird nicht geraten**: aus `{2}{R` kommt nur
// `{2}`. Der Rest hat keine schließende Klammer und ist damit kein Symbol —
// ihn zu ergänzen hieße, eine Eingabe zu erfinden.
func Manasymbole(kosten string) []string {
	var symbole []string

	rest := kosten

	for {
		auf := strings.Index(rest, "{")
		if auf < 0 {
			return symbole
		}

		zu := strings.Index(rest[auf:], "}")
		if zu < 0 {
			return symbole
		}

		symbole = append(symbole, rest[auf:auf+zu+1])
		rest = rest[auf+zu+1:]
	}
}
