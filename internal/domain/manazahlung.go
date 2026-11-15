package domain

// **Was womit bezahlbar ist** — Werkstück M3 der Regelmaschine
// (#1605).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan steht, was ein Vorrat ist**; hier steht, **wie man ihn
// ausgibt** — CR 118, CR 107.4.
//
// # Warum das eine Suche ist und keine Abrechnung
//
// **Weil ein Hybridsymbol zwei Wege hat** (CR 107.4e). `[W/U][W/U]` gegen
// einen Vorrat aus einem Weißen und einem Blauen geht auf; wer stur von links
// nach rechts das Weiße nähme, stünde beim zweiten Symbol mit einem Blauen da
// und sagte „geht nicht". Die Reihenfolge entscheidet also, und deshalb wird
// probiert statt gerechnet.
//
// **Die Suche bleibt klein**: durchsucht werden nur die Symbole mit **mehr
// als einem** Weg, und Manakosten haben davon selten mehr als drei.

import "fmt"

// zahlweise ist **ein** Weg, ein Symbol zu bezahlen.
//
// **Entweder eine Art oder eine Zahl**: `{W}` will ein weißes, `{2}` will zwei
// beliebige. Ein Symbol mit beidem gibt es nicht — `{2/B}` ist **zwei**
// Zahlweisen und nicht eine.
type zahlweise struct {
	art       Manaart
	generisch int
}

// Bezahlbar sagt, ob dieser Vorrat die Kosten deckt — CR 118.3.
//
// **Ganz oder gar nicht** (CR 601.2h): halbe Zahlungen gibt es nicht, und ein
// Vorrat, der neun von zehn deckt, deckt sie nicht.
//
// **Symbole, die dieses Programm nicht zahlen kann, machen die Kosten
// unbezahlbar** — welche das sind, sagt [Unbezahlbar], damit eine Fläche den
// Grund nennen kann, statt nur „nein" zu sagen (§8).
func (v Manavorrat) Bezahlbar(kosten Manakosten) bool {
	_, ging := v.Bezahle(kosten)

	return ging
}

// Bezahle nimmt die Kosten aus dem Vorrat und gibt, was bleibt.
//
// **Der zweite Wert sagt, ob es ging** — ging es nicht, ist der erste der
// **unveränderte** Vorrat: eine gescheiterte Zahlung lässt keinen halben Stand
// zurück (CR 601.2h).
//
// # Welche Art für generische Kosten ausgegeben wird
//
// **Farblos zuerst, dann WUBRG.** Das ist eine Entscheidung dieses Programms
// und keine Regel: CR 118 lässt dem Spieler die Wahl. Sie ist die am wenigsten
// einschränkende — farbloses Mana zahlt nur `{C}` und generische Kosten,
// farbiges zahlt mehr —, und sie steht hier, damit sie nachlesbar ist statt
// erraten (§15, Punkt 4). **In der Regelmatrix steht sie als Grenze.**
func (v Manavorrat) Bezahle(kosten Manakosten) (Manavorrat, bool) {
	if len(Unbezahlbar(kosten)) > 0 {
		return v, false
	}

	fest, wahlen, generisch := forderungen(kosten)

	uebrig := v.kopie()

	for _, art := range fest {
		if uebrig[art] == 0 {
			return v, false
		}

		uebrig = uebrig.Mehr(art, -1)
	}

	danach, ging := zahleWahlen(uebrig, wahlen, generisch)
	if !ging {
		return v, false
	}

	return danach, true
}

// forderungen zerlegt die Kosten in das, was feststeht, und das, was nicht.
//
// **Drei Töpfe, weil sie verschieden behandelt werden**: feste Arten kosten
// keine Überlegung, Wahlen werden durchprobiert, und generisches wird **zum
// Schluss** aus dem Rest gezahlt — es nimmt alles, und wer es zuerst zahlte,
// nähme einer Farbe ihr Mana weg.
func forderungen(kosten Manakosten) (fest []Manaart, wahlen [][]zahlweise, generisch int) {
	for _, symbol := range kosten.Symbole {
		wege := zahlweisen(symbol)

		switch {
		case len(wege) == 1 && wege[0].art != "":
			fest = append(fest, wege[0].art)
		case len(wege) == 1:
			generisch += wege[0].generisch
		default:
			wahlen = append(wahlen, wege)
		}
	}

	return fest, wahlen, generisch
}

// zahlweisen gibt die Wege, ein Symbol zu bezahlen — CR 107.4.
//
// **Ein leeres Ergebnis heißt: dieses Programm kann es nicht.** Gefiltert wird
// das vorher von [Unbezahlbar]; hier stünde sonst eine Forderung ohne Weg, und
// die sähe aus wie „nicht genug Mana" statt „das kennen wir nicht".
func zahlweisen(symbol Manasymbol) []zahlweise {
	var wege []zahlweise

	// **Die Farben zuerst** — ein Hybridsymbol trägt zwei, und jede ist ein
	// eigener Weg (CR 107.4e).
	for _, farbe := range symbol.Farben {
		wege = append(wege, zahlweise{art: manaartZuFarbe(farbe)})
	}

	if symbol.Farblos {
		wege = append(wege, zahlweise{art: ManaFarblos})
	}

	// **Und die Zahl als eigener Weg.** Bei `{2}` ist sie der einzige, bei
	// `{2/B}` der zweite (CR 107.4e).
	if symbol.Generisch > 0 || (symbol.Zahl && len(wege) == 0) {
		wege = append(wege, zahlweise{generisch: symbol.Generisch})
	}

	return wege
}

// zahleWahlen probiert die Hybridsymbole durch und zahlt zuletzt generisch.
//
// **Rekursiv und nicht als Schleife**, weil jede Wahl die nächste einschränkt.
// Die Tiefe ist die Zahl der Hybridsymbole einer Manakosten — bei den
// erhobenen Paketen nie mehr als drei.
func zahleWahlen(
	vorrat Manavorrat, wahlen [][]zahlweise, generisch int,
) (Manavorrat, bool) {
	if len(wahlen) == 0 {
		return zahleGenerisch(vorrat, generisch)
	}

	for _, weg := range wahlen[0] {
		if weg.art != "" {
			if vorrat[weg.art] == 0 {
				continue
			}

			danach, ging := zahleWahlen(
				vorrat.Mehr(weg.art, -1), wahlen[1:], generisch)
			if ging {
				return danach, true
			}

			continue
		}

		danach, ging := zahleWahlen(vorrat, wahlen[1:], generisch+weg.generisch)
		if ging {
			return danach, true
		}
	}

	return vorrat, false
}

// zahleGenerisch nimmt beliebiges Mana — farblos zuerst.
func zahleGenerisch(vorrat Manavorrat, offen int) (Manavorrat, bool) {
	if vorrat.Gesamt() < offen {
		return vorrat, false
	}

	danach := vorrat.kopie()

	for _, art := range generischeReihenfolge() {
		for offen > 0 && danach[art] > 0 {
			danach = danach.Mehr(art, -1)
			offen--
		}
	}

	return danach, offen == 0
}

// generischeReihenfolge gibt die Arten, aus denen generisch gezahlt wird.
//
// **Farblos zuerst** — der Grund steht bei [Manavorrat.Bezahle].
func generischeReihenfolge() []Manaart {
	return []Manaart{
		ManaFarblos, ManaWeiss, ManaBlau, ManaSchwarz, ManaRot, ManaGruen,
	}
}

// Unbezahlbar nennt die Symbole, die dieses Programm nicht zahlen kann — §8.
//
// **Ein Befund ist kein Fehler**: die Kosten sind lesbar und der Manawert
// stimmt; was fehlt, ist ein Weg, sie **auszugeben**. Wer nur `false` bekäme,
// suchte den Fehler in seinem Vorrat.
//
// **Drei Sorten, und jede hat ihren Grund:**
//
//   - `{X}` — der Wert wird beim Wirken angesagt (CR 601.2b), und das Wirken
//     ist `M4` (#1606).
//   - `{S}` — Schneemana ist Mana aus einer Schneequelle (CR 107.4h); dieses
//     Programm führt die Herkunft eines Manas nicht.
//   - Phyrexianisch — es lässt sich mit **zwei Leben** zahlen (CR 107.4f), und
//     das ist eine Wahl zwischen zwei Währungen. Wer sie trifft, ist der
//     Spieler, und einen Weg, ihn zu fragen, gibt es am Tisch nicht.
//   - Alles, was CR 107 nicht kennt (`Unbekannt`) — dasselbe wie überall bei
//     fremden Daten.
func Unbezahlbar(kosten Manakosten) []string {
	var befunde []string

	for _, symbol := range kosten.Symbole {
		switch {
		case symbol.Variabel:
			befunde = append(befunde, fmt.Sprintf(
				"%s: der Wert von X wird beim Wirken angesagt (CR 601.2b)",
				symbol.Roh))
		case symbol.Schnee:
			befunde = append(befunde, fmt.Sprintf(
				"%s: Schneemana kommt aus einer Schneequelle (CR 107.4h)",
				symbol.Roh))
		case symbol.Phyrexianisch:
			befunde = append(befunde, fmt.Sprintf(
				"%s: zahlbar auch mit zwei Leben — das ist eine Wahl "+
					"(CR 107.4f)", symbol.Roh))
		case symbol.Unbekannt:
			befunde = append(befunde, fmt.Sprintf(
				"%s: kein Symbol aus CR 107.4", symbol.Roh))
		}
	}

	return befunde
}
