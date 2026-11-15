package domain

// **Ein Land spielen** — Werkstück M4 der Regelmaschine
// (#1606).
//
// # Die einzige der zwölf Sonderhandlungen, die hier geht
//
// **CR 116.2 zählt zwölf.** Elf davon brauchen einen Kartentext oder eine
// Variante, die dieses Programm nicht führt: eine verdeckte Kreatur umdrehen
// (116.2b), Effekte beenden (116.2c/d), eine benannte Karte abwerfen
// (116.2e), aussetzen (116.2f), ein Gefährte (116.2g), vorhersagen (116.2h),
// der Planarwürfel (116.2i), eine Verschwörung (116.2j), planen (116.2k).
// Wer sie bauen will, braucht `M14`
// (#1616)
// oder eine Variante, die #1593 ausgrenzt.
//
// **Ein Land zu spielen braucht keinen Text** — CR 116.2a steht im Regelwerk,
// so wie die innewohnende Manafähigkeit in CR 305.6. Deshalb geht diese eine.
//
// # Und sie geht nicht über den Stapel
//
// **CR 116.1**: Sonderhandlungen benutzen ihn nicht. Niemand kann darauf
// antworten, und es gibt nichts zu verrechnen — das Land liegt sofort da.

import "fmt"

// LandSpielen legt ein Land aus der Hand auf das Schlachtfeld — CR 116.2a.
func (p Partie) LandSpielen(wer, objekt string) (Partie, error) {
	if fehler := p.darfLandSpielen(wer); fehler != nil {
		return p, fehler
	}

	stelle, fehler := p.landkarte(wer, objekt)
	if fehler != nil {
		return p, fehler
	}

	// **Ein neues Objekt auf dem Schlachtfeld** — CR 400.7 (M18, #1619).
	p = p.zonenwechsel(stelle, ZoneTisch)
	p.Zug.Laender++

	return p, nil
}

// darfLandSpielen prüft die drei Bedingungen aus CR 116.2a und CR 305.2.
func (p Partie) darfLandSpielen(wer string) error {
	// **Eine Partie ohne Zugbegriff ist ein Sandkasten** — dort gilt CR
	// 116.2a nicht, weil es keine Hauptphase gibt (G4, §15 Punkt 4).
	if p.Zug.Nummer == 0 {
		return nil
	}

	haupt := p.Zug.Abschnitt.Phase == PhaseHauptVorKampf ||
		p.Zug.Abschnitt.Phase == PhaseHauptNachKampf

	switch {
	case p.Zug.Prioritaet != wer:
		return fmt.Errorf("%w: %q hat keine Priorität (CR 116.2a)",
			ErrValidation, wer)
	case p.Zug.Aktiv != wer:
		return fmt.Errorf("%w: es ist nicht der Zug von %q (CR 116.2a)",
			ErrValidation, wer)
	case !haupt:
		return fmt.Errorf(
			"%w: ein Land geht nur in einer Hauptphase (CR 116.2a), und es "+
				"ist %q", ErrValidation, p.Zug.Abschnitt.Phase)
	case p.StapelOben() != nil:
		return fmt.Errorf(
			"%w: ein Land geht nur bei leerem Stapel (CR 116.2a)",
			ErrValidation)
	case p.Zug.Laender > 0:
		return fmt.Errorf(
			"%w: in diesem Zug ist schon ein Land gespielt (CR 305.2)",
			ErrValidation)
	}

	return nil
}

// landkarte findet die Karte und prüft, ob sie ein Land ist.
func (p Partie) landkarte(wer, objekt string) (int, error) {
	for stelle := range p.Objekte {
		if p.Objekte[stelle].Kennung != objekt {
			continue
		}

		if p.Objekte[stelle].Beherrscher != wer {
			return 0, fmt.Errorf(
				"%w: %q beherrscht das Objekt %q nicht (CR 116.2a)",
				ErrValidation, wer, objekt)
		}

		eigen, da := p.Eigenschaften(p.Objekte[stelle])

		switch {
		case !da:
			return 0, fmt.Errorf(
				"%w: zu %q steht keine Karte im eingefrorenen Abzug",
				ErrNotFound, objekt)
		case !eigen.Typzeile.Ist(TypLand):
			return 0, fmt.Errorf("%w: %q ist kein Land (CR 305)",
				ErrValidation, eigen.Name)
		}

		return stelle, nil
	}

	return 0, fmt.Errorf("%w: das Objekt %q gibt es in dieser Partie nicht",
		ErrNotFound, objekt)
}
