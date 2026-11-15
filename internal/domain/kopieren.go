package domain

// **Kopiereffekte** — CR 707, Werkstück M18 der Regelmaschine
// (#1619).
//
// # Was ein kopierbarer Wert in diesem Programm ist
//
// **CR 707.2** sagt: die Kopie nimmt die **kopierbaren Werte** der Merkmale
// des Originals an — die gedruckten Werte, verändert nur durch andere
// Kopiereffekte und durch Effekte, die sie beim Ins-Spiel-Kommen setzen.
//
// **Hier gibt es keine Effekte.** Also sind die kopierbaren Werte genau die
// gedruckten, und die stehen an zwei Stellen: im eingefrorenen Deckabzug
// (eine Karte) oder am Objekt selbst (ein Spielstein, CR 111.1). Kopiert
// wird deshalb der **Verweis** und nicht das Ergebnis — wer die Merkmale
// ausrechnete und ablegte, hätte sie doppelt, und die zwei liefen beim
// nächsten Blick auseinander (derselbe Grund wie bei `Partie.Eigenschaften`).
//
// # Und was nicht kopiert wird
//
// **Der Zustand** — CR 707.3 mit CR 110.5: getappt bleibt getappt, Marken
// bleiben liegen, und wer sie beherrscht, beherrscht sie weiter. Das ist der
// Unterschied zum Zonenwechsel nebenan, wo genau das wegfällt.
//
// **Und die Wahlen eines Zaubers** (Modus, X, Ziele): CR 707.2 nennt sie
// ausdrücklich, und es gibt sie ohne `M8` nicht.

import "fmt"

// Kopieren macht aus einem Objekt die Kopie eines anderen — CR 707.2.
//
// **Zwei Kennungen und keine Richtung im Namen**: `objekt` wird zur Kopie,
// `vorlage` bleibt, was sie war. Ein `Kopiere(a, b)` ließe offen, welches
// welches ist.
func (p Partie) Kopieren(objekt, vorlage string) (Partie, error) {
	stelle, fehler := p.objektstelle(objekt)
	if fehler != nil {
		return p, fehler
	}

	quelle, fehler := p.objektstelle(vorlage)
	if fehler != nil {
		return p, fehler
	}

	if stelle == quelle {
		return p, fmt.Errorf(
			"%w: %q kopiert sich nicht selbst (CR 707.2)",
			ErrValidation, objekt)
	}

	objekte := make([]Spielobjekt, len(p.Objekte))
	copy(objekte, p.Objekte)

	// **Die gedruckten Werte werden eingefroren und nicht verwiesen.**
	//
	// CR 707.3 sagt es so: die kopierbaren Werte der Kopie **werden** die
	// kopierte Information. Ein Verweis auf die Karte der Vorlage ginge
	// nicht: [Partie.Eigenschaften] sucht sie im Deckabzug des **Eigners**,
	// und der Eigner ändert sich beim Kopieren nicht (CR 108.3 — er ist,
	// wer mit der Karte angefangen hat). Wer ihn mitkopierte, legte die
	// Kopie beim Sterben in den falschen Friedhof.
	felder, da := p.gedruckteFelder(objekte[quelle])
	if !da {
		return p, fmt.Errorf(
			"%w: von %q ist nicht bekannt, was darauf steht (CR 707.2)",
			ErrNotFound, vorlage)
	}

	objekte[stelle].Felder = felder
	objekte[stelle].Name = objekte[quelle].Name

	p.Objekte = objekte

	return p, nil
}

// gedruckteFelder gibt die Felder, die auf einem Objekt stehen — CR 707.2.
//
// **Zwei Herkünfte**: ein Spielstein trägt sie am Objekt (CR 111.1), eine
// Karte im eingefrorenen Abzug ihres Eigners. Verdeckt liegt nichts
// Kopierbares (CR 708.2a nennt keine Felder).
func (p Partie) gedruckteFelder(objekt Spielobjekt) (map[string]string, bool) {
	if len(objekt.Felder) > 0 {
		return kopierteFelder(objekt.Felder), true
	}

	for _, sitz := range p.Sitze {
		if sitz.Kennung != objekt.Eigner {
			continue
		}

		for _, karte := range sitz.Deck.Karten {
			if karte.Kennung != objekt.Karte || len(karte.Seiten) == 0 {
				continue
			}

			seite := objekt.Seite
			if seite < 0 || seite >= len(karte.Seiten) {
				seite = 0
			}

			return kopierteFelder(karte.Seiten[seite].Felder), true
		}
	}

	return nil, false
}

// kopierteFelder gibt eine eigene Tabelle — oder keine.
//
// **Eine geteilte wäre ein geteiltes Objekt.** Wer danach ein Feld der Kopie
// änderte, änderte die Vorlage mit; dieselbe Falle wie bei `Sandkasten.kopie`
// und derselbe Grund.
func kopierteFelder(felder map[string]string) map[string]string {
	if len(felder) == 0 {
		return nil
	}

	eigene := make(map[string]string, len(felder))
	for name, wert := range felder {
		eigene[name] = wert
	}

	return eigene
}
