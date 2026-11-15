package domain

// **Was ein Objekt auf dem Tisch ist** — Werkstück M2 der Regelmaschine
// (#1604).
//
// # Warum das die Partie beantwortet und nicht das Objekt
//
// **Weil ein [Spielobjekt] seine Karte nicht kennt.** Es trägt einen Verweis
// ([Spielobjekt.Karte]) in den **eingefrorenen Abzug** seines Eigners; welcher
// Sitz das ist, weiß nur die Partie. Ein Objekt, das seine Eigenschaften
// selbst trüge, trüge sie doppelt — einmal hier und einmal im Abzug — und die
// beiden liefen auseinander, sobald jemand eine Karte im Bestand ändert.
//
// # Und warum gerechnet und nicht gespeichert wird
//
// **Weil kein Bestandsschema sich ändern soll.** Eine Partie von gestern trägt
// dieselben Felder wie eine von heute; gedeutet wird beim Lesen. Was das
// kostet, ist eine Zerlegung je Abfrage — und was es spart, ist eine
// Migration über jede laufende Partie.
//
// # Die Grenze
//
// **Hier stehen die gedruckten Werte plus Marken, sonst nichts.** Was
// Effekte, Schichten und textändernde Zauber daraus machen, ist CR 613 und
// damit `M10`
// (#1612);
// die Schichtreihenfolge zu erraten wäre genau der Fehler, den §15 verhindert.

// Eigenschaften gibt die Eigenschaften eines Objekts — **false heißt: keine**.
//
// **Keine ist der Normalfall für einen Spielstein** (CR 111): er kommt aus
// keinem Deck, und [Spielobjekt.Karte] ist leer. Ein Verweis, der ins Leere
// zeigt, kommt aus einem von Hand geänderten Bestand und bricht ebenso wenig
// etwas (§8).
func (p Partie) Eigenschaften(objekt Spielobjekt) (Karteneigenschaften, bool) {
	// **Verdeckt hat keine Eigenschaften außer den genannten** — CR 708.2a
	// (M12, #1614).
	// Was die Regel nennt, steht in [VerdeckteEigenschaften].
	if objekt.Verdeckt && objekt.Zone == ZoneTisch {
		return VerdeckteEigenschaften(), true
	}

	// **Ein Spielstein bringt seine Eigenschaften mit** — CR 111.1: er
	// vertritt keine Karte, und sein Erzeuger nennt sie (CR 701.7a).
	if len(objekt.Felder) > 0 {
		return EigenschaftenAus(p.Spiel, objekt.Felder), true
	}

	if objekt.Karte == "" {
		return Karteneigenschaften{}, false
	}

	for _, sitz := range p.Sitze {
		if sitz.Kennung != objekt.Eigner {
			continue
		}

		return eigenschaftenAusDeck(p.Spiel, sitz.Deck, objekt.Karte,
			objekt.Seite)
	}

	return Karteneigenschaften{}, false
}

// VerdeckteEigenschaften gibt, was ein verdecktes Permanent ist — CR 708.2a.
//
// **Eine 2/2 ohne Namen, ohne Typen, ohne Manakosten.** Die Regel nennt sie
// ausdrücklich, und sie gilt für jedes verdeckte Permanent, dem kein Effekt
// andere Eigenschaften gegeben hat.
//
// **Was dieses Programm nicht kann, ist der andere Fall**: eine Fähigkeit,
// die beim Umdrehen Eigenschaften **nennt** — Morph nennt sie, und Morph ist
// ein Schlüsselwort (`M14`). Das steht als Grenze in der Regelmatrix.
func VerdeckteEigenschaften() Karteneigenschaften {
	return Karteneigenschaften{
		Typzeile:         Typzeile{Typen: []Kartentyp{TypKreatur}},
		Staerke:          Kartenwert{Steht: true, Zahl: 2},
		Widerstandskraft: Kartenwert{Steht: true, Zahl: 2},
	}
}

// eigenschaftenAusDeck sucht die Karte im Abzug eines Sitzes.
//
// **Die Vorderseite gilt, solange nichts sie umdreht** — CR 712.8a. Welche
// Seite oben liegt, steht seit `M12`
// (#1614)
// am Objekt.
func eigenschaftenAusDeck(
	spiel string, deck Partiedeck, kennung string, seite int,
) (Karteneigenschaften, bool) {
	for _, karte := range deck.Karten {
		if karte.Kennung != kennung || len(karte.Seiten) == 0 {
			continue
		}

		// **Die Seite, die oben liegt** — CR 701.27a (M12). Eine Seite, die
		// es nicht gibt, fällt auf die Vorderseite zurück: eine Karte, die
		// nur eine hat, wurde nie umgedreht, und ein von Hand geänderter
		// Bestand darf nicht umfallen (§8).
		if seite < 0 || seite >= len(karte.Seiten) {
			seite = 0
		}

		return EigenschaftenAus(spiel, karte.Seiten[seite].Felder), true
	}

	return Karteneigenschaften{}, false
}

// Werte gibt Stärke und Widerstandskraft eines Objekts **mit seinen Marken** —
// CR 122.1a (Fassung 2026-08-19).
//
// **Die zweite Abnahme von M2**: eine Kreatur 2/3 mit zwei +1/+1-Marken ist
// 4/5 — gerechnet und nicht getippt.
//
// **Ohne gedruckte Werte bleiben sie ohne** (CR 208.3): eine Marke auf einem
// Land macht daraus keine 1/1. CR 613.4c legt die Marken über die gedruckten
// Werte, und wo keine stehen, liegt sie über nichts.
func (p Partie) Werte(objekt Spielobjekt) (staerke, widerstand Kartenwert) {
	eigen, da := p.Eigenschaften(objekt)
	if !da {
		return Kartenwert{}, Kartenwert{}
	}

	um := objekt.Zaehler[MarkePlusEins] - objekt.Zaehler[MarkeMinusEins]

	staerke, widerstand = eigen.Staerke, eigen.Widerstandskraft

	// **Und darüber die siebte Schicht** — CR 613.4 (M10,
	// #1612).
	//
	// **Vor den Marken**, und das ist die Regel: CR 613.4c nennt Effekte
	// **und** Marken in derselben Unterschicht, und CR 613.4b (setzen)
	// liegt davor. Wer die Marken zuerst rechnete, gäbe einer auf 5/5
	// gesetzten Kreatur mit einer Marke 5/5 statt 6/6.
	staerke, widerstand = p.werteMitEffekten(
		objekt.Kennung, staerke, widerstand)

	if staerke.Steht {
		staerke.Zahl += um
	}

	if widerstand.Steht {
		widerstand.Zahl += um
	}

	return staerke, widerstand
}

// markenVerrechnen entfernt Plus- und Minusmarken gegeneinander — CR 704.5q
// (Fassung 2026-08-19), die zustandsbasierte Aktion zu CR 122.3.
//
// **N von jeder Art, wo N die kleinere Zahl ist** — und nicht „alle": ein
// Objekt mit drei Plus und einer Minus behält zwei Plus.
//
// **Auf dem Tisch und nur dort.** CR 704.5q spricht vom Permanent, und ein
// Permanent liegt auf dem Schlachtfeld (CR 110.1).
func (p Partie) markenVerrechnen() (Partie, bool) {
	var geaendert []Spielobjekt

	for stelle, objekt := range p.Objekte {
		kleinere := min(objekt.Zaehler[MarkePlusEins],
			objekt.Zaehler[MarkeMinusEins])

		if objekt.Zone != ZoneTisch || kleinere <= 0 {
			continue
		}

		if geaendert == nil {
			geaendert = make([]Spielobjekt, len(p.Objekte))
			copy(geaendert, p.Objekte)
		}

		zaehler := make(map[string]int, len(objekt.Zaehler))
		for art, anzahl := range objekt.Zaehler {
			zaehler[art] = anzahl
		}

		zaehler[MarkePlusEins] -= kleinere
		zaehler[MarkeMinusEins] -= kleinere

		geaendert[stelle].Zaehler = zaehler
	}

	if geaendert == nil {
		return p, false
	}

	p.Objekte = geaendert

	return p, true
}
