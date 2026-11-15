package domain

// **Erstschlag und Doppelschlag teilen den Schadensschritt** — Werkstück M6
// der Regelmaschine
// (#1608).
//
// # Warum zwei von 194 Schlüsselwörtern hier stehen
//
// **Weil CR 510.4 sie namentlich nennt.** Erstschlag (CR 702.7a) und
// Doppelschlag (CR 702.4a) sind **statische** Fähigkeiten, deren ganze
// Bedeutung in CR 510 steht: sie tun nichts, was ein Text ausführen müsste —
// sie **teilen den Schadensschritt**. Sie zu erkennen heißt, ein Wort zu
// lesen, so wie `M2` die Typzeile liest; einen Text **auszuführen** wäre
// `M8`.
//
// **Seit `M14`** (#1616)
// sind es **zwölf**, und welche, folgt einer Zählung — der Grund und die
// Zahlen stehen bei [SchluesselwoerterAus].
//
// # Und wie gelesen wird
//
// **Die Zeile und keine Teilzeichenkette** (CR 702.1): eine Teilzeichenkette
// läse „Zielkreatur verliert Erstschlag" mit. `M14` hat den Preis
// **gemessen** — bei `flying` 45 Prozent zu viel.

// regeltextfelder sind die Feldnamen, unter denen ein Regeltext steht.
//
// **Aus `docs/FORMATE.md`** — MSE nennt das Feld `rule text`; MTG JSON und
// Scryfall schreiben `text`.
var regeltextfelder = []string{"rule text", "rules text", "text", "rules"}

// hatErstschlag sagt, ob diese Kreatur im ersten Schadensschritt austeilt —
// CR 510.4.
//
// **Doppelschlag zählt mit**: CR 702.4b sagt, dass eine Kreatur mit
// Doppelschlag in **beiden** Schritten austeilt, also auch im ersten.
func (p Partie) hatErstschlag(kennung string) bool {
	return p.HatFaehigkeit(kennung, SchluesselErstschlag) ||
		p.HatFaehigkeit(kennung, SchluesselDoppelschlag)
}

// hatDoppelschlag sagt, ob diese Kreatur **auch** im zweiten Schritt
// austeilt — CR 702.4b.
func (p Partie) hatDoppelschlag(kennung string) bool {
	return p.HatFaehigkeit(kennung, SchluesselDoppelschlag)
}

// regeltext gibt den gedruckten Regeltext einer Karte — leer, wo keiner steht.
func (p Partie) regeltext(kennung string) string {
	for _, objekt := range p.Objekte {
		if objekt.Kennung != kennung {
			continue
		}

		for _, sitz := range p.Sitze {
			if sitz.Kennung != objekt.Eigner {
				continue
			}

			return regeltextAusDeck(sitz.Deck, objekt.Karte)
		}
	}

	return ""
}

// regeltextAusDeck sucht den Regeltext im eingefrorenen Abzug.
func regeltextAusDeck(deck Partiedeck, karte string) string {
	for _, eine := range deck.Karten {
		if eine.Kennung != karte || len(eine.Seiten) == 0 {
			continue
		}

		for _, feld := range regeltextfelder {
			if wert := eine.Seiten[0].Felder[feld]; wert != "" {
				return wert
			}
		}
	}

	return ""
}

// erstschlagImKampf sagt, ob überhaupt jemand Erstschlag hat — CR 510.4.
func (p Partie) erstschlagImKampf() bool {
	for _, angriff := range p.Zug.Kampf.Angriffe {
		if p.hatErstschlag(angriff.Kreatur) {
			return true
		}
	}

	for _, block := range p.Zug.Kampf.Bloecke {
		if p.hatErstschlag(block.Kreatur) {
			return true
		}
	}

	return false
}

// teiltInDiesemSchrittAus sagt, ob diese Kreatur jetzt austeilt — CR 510.4.
func (p Partie) teiltInDiesemSchrittAus(kennung string) bool {
	switch {
	// **Im zweiten Schritt teilt aus, wer keinen Erstschlag hat** — und wer
	// Doppelschlag hat, ein zweites Mal (CR 702.4b).
	case p.Zug.Kampf.Schadensschritte > 0:
		return !p.hatErstschlag(kennung) || p.hatDoppelschlag(kennung)

	// **Gibt es einen Erstschläger, teilt im ersten Schritt nur er aus.**
	case p.erstschlagImKampf():
		return p.hatErstschlag(kennung)
	}

	// **Sonst teilen alle aus, und es gibt nur diesen einen Schritt.**
	return true
}

// zweiterSchadensschritt sagt, ob nach diesem noch einer kommt — CR 510.4.
//
// **Nur nach dem ersten und nur mit einem Erstschläger.** Ohne beides ginge
// es vom Schadensschritt geradewegs ins Kampfende (CR 506.1).
func (p Partie) zweiterSchadensschritt() bool {
	return p.Zug.Kampf.Schadensschritte == 1 && p.erstschlagImKampf()
}
