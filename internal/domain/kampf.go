package domain

// **Angreifen, blocken, Schaden** — Werkstück M6 der Regelmaschine
// (#1608).
//
// # Drei Handlungen und eine Reihenfolge
//
// **CR 506.1** nennt fünf Kampfschritte; drei davon haben Inhalt:
//
//   - **Angreifer erklären** (CR 508.1) — im gleichnamigen Schritt, vom
//     aktiven Spieler, und die Gewählten werden dabei getappt (CR 508.1f).
//   - **Blocker erklären** (CR 509.1) — im gleichnamigen Schritt, vom
//     Verteidiger, und **ohne** zu tappen.
//   - **Kampfschaden** (CR 510.1, CR 510.2) — jede teilt ihre Stärke aus, und
//     zwar **gleichzeitig**: niemand kommt dazwischen.
//
// # Was hier nicht steht
//
// **Keine Einschränkung aus einer Fähigkeit** — CR 508.1c und CR 509.1b
// fragen, ob ein Effekt das Angreifen oder Blocken verbietet oder verlangt.
// Solche Effekte stehen im **Text** der Karte und brauchen `M8`. Was hier
// geprüft wird, steht im **Regelwerk**: getappt, Kreatur, wem sie gehört.
//
// **Und kein Trampelschaden, kein Flugfähig** — die sind Schlüsselwörter
// (CR 702) und damit `M14`
// (#1616).

import "fmt"

// AngreiferErklaeren erklärt die Angreifer — CR 508.1.
//
// **Auch keine ist eine Erklärung** (CR 508.1a: „if any"): ein Kampf, in dem
// niemand angreift, ist ein erklärter Kampf und kein unerklärter.
func (p Partie) AngreiferErklaeren(wer string, angriffe []Angriff) (Partie, error) {
	switch {
	case p.Zug.Abschnitt.Schritt != SchrittAngreifer:
		return p, fmt.Errorf(
			"%w: Angreifer werden im Angreiferschritt erklärt (CR 508.1), "+
				"und es ist %q", ErrValidation, p.Zug.Abschnitt.Schritt)
	case p.Zug.Aktiv != wer:
		return p, fmt.Errorf(
			"%w: nur der aktive Spieler greift an (CR 506.2), und das ist "+
				"%q", ErrValidation, p.Zug.Aktiv)
	case p.Zug.Kampf.Erklaert:
		return p, fmt.Errorf(
			"%w: die Angreifer sind in diesem Kampf schon erklärt "+
				"(CR 508.1)", ErrValidation)
	}

	getappt := make([]Spielobjekt, len(p.Objekte))
	copy(getappt, p.Objekte)

	for _, angriff := range angriffe {
		stelle, fehler := p.angreiferstelle(wer, angriff)
		if fehler != nil {
			return p, fehler
		}

		// **Tappen ist keine Kosten** — CR 508.1f sagt es ausdrücklich: das
		// Angreifen **bewirkt** das Tappen. Wer es als Kosten läse, ließe
		// eine getappte Kreatur angreifen, die sich anders bezahlt.
		//
		// **Und Wachsamkeit hebt genau das auf** — CR 702.20b (M14,
		// #1616):
		// nicht das Angreifen, sondern das Tappen dabei.
		if !p.HatFaehigkeit(angriff.Kreatur, SchluesselWachsamkeit) {
			getappt[stelle].Getappt = true
		}
	}

	p.Objekte = getappt
	p.Zug.Kampf.Angriffe = append([]Angriff{}, angriffe...)
	p.Zug.Kampf.Erklaert = true

	return p, nil
}

// angreiferstelle prüft eine Kreatur und ihr Ziel — CR 508.1a.
func (p Partie) angreiferstelle(wer string, angriff Angriff) (int, error) {
	stelle, fehler := p.kampfkreatur(wer, angriff.Kreatur)
	if fehler != nil {
		return 0, fehler
	}

	if p.Objekte[stelle].Getappt {
		return 0, fmt.Errorf(
			"%w: %q ist getappt und greift nicht an (CR 508.1a)",
			ErrValidation, angriff.Kreatur)
	}

	// **Einsatzverzögerung** — CR 302.6, den CR 508.1a namentlich nennt:
	// eine Kreatur muss seit dem Beginn des laufenden Zuges unter derselben
	// Kontrolle stehen. **Eile hebt es auf** (CR 702.10b), und die Regel
	// nennt auch sie — der Grund, warum dieses Programm gerade dieses
	// Schlüsselwort liest, steht bei [Partie.hatErstschlag].
	if p.Zug.Nummer > 0 && p.Objekte[stelle].Seit == p.Zug.Nummer &&
		!p.HatFaehigkeit(angriff.Kreatur, SchluesselEile) {
		return 0, fmt.Errorf(
			"%w: %q ist in diesem Zug gekommen und hat keine Eile "+
				"(CR 302.6)", ErrValidation, angriff.Kreatur)
	}

	// **Ein Verteidiger greift nicht an** — CR 702.3b (M14,
	// #1616).
	if p.HatFaehigkeit(angriff.Kreatur, SchluesselVerteidiger) {
		return 0, fmt.Errorf(
			"%w: %q hat Verteidiger und greift nicht an (CR 702.3b)",
			ErrValidation, angriff.Kreatur)
	}

	if !p.verteidigerSteht(wer, angriff.Ziel) {
		return 0, fmt.Errorf(
			"%w: %q ist kein zulässiges Angriffsziel (CR 508.1a)",
			ErrValidation, angriff.Ziel)
	}

	return stelle, nil
}

// verteidigerSteht sagt, ob dieses Ziel angegriffen werden darf — CR 508.1a.
//
// **Ein gegnerischer Sitz, ein Planeswalker oder eine Kampfkarte** — und
// alle drei müssen einem **anderen** gehören. Wer sich selbst angriffe,
// spielte ein Spiel, das die Regeln nicht kennen.
func (p Partie) verteidigerSteht(wer, ziel string) bool {
	for _, sitz := range p.Sitze {
		if sitz.Kennung == ziel {
			return sitz.Kennung != wer
		}
	}

	for _, objekt := range p.Objekte {
		if objekt.Kennung != ziel {
			continue
		}

		if objekt.Zone != ZoneTisch || objekt.Beherrscher == wer {
			return false
		}

		eigen, da := p.Eigenschaften(objekt)

		return da && (eigen.Typzeile.Ist(TypPlaneswalker) ||
			eigen.Typzeile.Ist(TypKampf))
	}

	return false
}

// BlockerErklaeren erklärt die Blocker — CR 509.1.
func (p Partie) BlockerErklaeren(wer string, bloecke []Block) (Partie, error) {
	switch {
	case p.Zug.Abschnitt.Schritt != SchrittBlocker:
		return p, fmt.Errorf(
			"%w: Blocker werden im Blockerschritt erklärt (CR 509.1), und "+
				"es ist %q", ErrValidation, p.Zug.Abschnitt.Schritt)
	case p.Zug.Aktiv == wer:
		return p, fmt.Errorf(
			"%w: der aktive Spieler blockt nicht (CR 506.2)", ErrValidation)
	case p.Zug.Kampf.Geblockt:
		return p, fmt.Errorf(
			"%w: die Blocker sind in diesem Kampf schon erklärt (CR 509.1)",
			ErrValidation)
	}

	for _, block := range bloecke {
		if fehler := p.blockPruefen(wer, block); fehler != nil {
			return p, fehler
		}
	}

	// **Bedrohlich zählt die Blocker und nicht den einzelnen Block** —
	// CR 702.111b (M14,
	// #1616).
	// Deshalb steht es **hinter** der Schleife: die Frage lautet, wie viele
	// zusammen auf denselben Angreifer gehen.
	if fehler := p.bedrohlichPruefen(bloecke); fehler != nil {
		return p, fehler
	}

	p.Zug.Kampf.Bloecke = append([]Block{}, bloecke...)
	p.Zug.Kampf.Geblockt = true

	return p, nil
}

// blockPruefen prüft eine blockende Kreatur — CR 509.1a.
func (p Partie) blockPruefen(wer string, block Block) error {
	stelle, fehler := p.kampfkreatur(wer, block.Kreatur)
	if fehler != nil {
		return fehler
	}

	if p.Objekte[stelle].Getappt {
		return fmt.Errorf("%w: %q ist getappt und blockt nicht (CR 509.1a)",
			ErrValidation, block.Kreatur)
	}

	if !p.Zug.Kampf.greiftAn(block.Angreifer) {
		return fmt.Errorf("%w: %q greift gar nicht an (CR 509.1a)",
			ErrValidation, block.Angreifer)
	}

	// **Flugfähigkeit sagt, wer blocken darf** — CR 702.9b (M14,
	// #1616):
	// nur Kreaturen mit Flugfähigkeit oder Reichweite (CR 702.17b).
	if p.HatFaehigkeit(block.Angreifer, SchluesselFlug) &&
		!p.HatFaehigkeit(block.Kreatur, SchluesselFlug) &&
		!p.HatFaehigkeit(block.Kreatur, SchluesselReichweite) {
		return fmt.Errorf(
			"%w: %q fliegt, und %q hat weder Flugfähigkeit noch Reichweite "+
				"(CR 702.9b)", ErrValidation, block.Angreifer, block.Kreatur)
	}

	return nil
}

// bedrohlichPruefen hält die Zahl der Blocker gegen CR 702.111b.
//
// **Geblockt oder gar nicht**: ein Angreifer mit Bedrohlich, auf den genau
// **einer** geht, ist unzulässig geblockt; keiner ist zulässig — CR 509.1b
// verlangt keine Blocker, sondern eine zulässige Zuteilung.
func (p Partie) bedrohlichPruefen(bloecke []Block) error {
	gezaehlt := map[string]int{}
	for _, block := range bloecke {
		gezaehlt[block.Angreifer]++
	}

	for angreifer, wie := range gezaehlt {
		if wie == 1 && p.HatFaehigkeit(angreifer, SchluesselBedrohlich) {
			return fmt.Errorf(
				"%w: %q ist bedrohlich und wird von zweien oder mehr "+
					"geblockt oder von keinem (CR 702.111b)",
				ErrValidation, angreifer)
		}
	}

	return nil
}

// kampfkreatur findet eine Kreatur auf dem Schlachtfeld dieses Sitzes.
func (p Partie) kampfkreatur(wer, kennung string) (int, error) {
	for stelle := range p.Objekte {
		if p.Objekte[stelle].Kennung != kennung {
			continue
		}

		if p.Objekte[stelle].Beherrscher != wer ||
			p.Objekte[stelle].Zone != ZoneTisch {
			return 0, fmt.Errorf(
				"%w: %q liegt nicht auf dem Schlachtfeld von %q",
				ErrValidation, kennung, wer)
		}

		eigen, da := p.Eigenschaften(p.Objekte[stelle])
		if !da || !eigen.Typzeile.Ist(TypKreatur) {
			return 0, fmt.Errorf("%w: %q ist keine Kreatur (CR 508.1a)",
				ErrValidation, kennung)
		}

		return stelle, nil
	}

	return 0, fmt.Errorf("%w: das Objekt %q gibt es in dieser Partie nicht",
		ErrNotFound, kennung)
}
