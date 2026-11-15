package service

// **Was ein Befehl am Objekt ändert** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan steht, welcher Befehl wohin geht**; hier steht, **was er tut** —
// tappen, umdrehen, Zähler setzen, anhängen. `sandkasten.go` stand mit beidem
// bei 408 Zeilen (§2).

import (
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// wirken ändert das Objekt und sagt, was dabei geschah.
func (s *Sandkasten) wirken(
	danach *domain.Partie, stelle int, befehl domain.Spielbefehl,
) (map[string]any, error) {
	objekt := &danach.Objekte[stelle]

	switch befehl.Art {
	case BefehlTappen:
		objekt.Getappt = true

	case BefehlEnttappen:
		objekt.Getappt = false

	case BefehlUmdrehen:
		objekt.Verdeckt = !objekt.Verdeckt

	case BefehlSchieben:
		// **Als Einziges mit der ganzen Partie**: wer eine Karte einsortiert,
		// rückt jede darunter. Der Weg steht in `sandkastenschieben.go`.
		return s.schieben(danach, stelle, befehl)

	case BefehlZaehler:
		return s.zaehlen(objekt, befehl)

	case BefehlAnhaengen:
		return s.anhaengen(objekt, befehl)

	case BefehlLoesen:
		objekt.HaengtAn = ""

	// **Der Haufen braucht zwei Objekte** — wie das Schieben, und deshalb
	// geht er über die ganze Partie
	// (#1756).
	case BefehlHaeufen:
		return s.haeufen(danach, stelle, befehl)

	case BefehlEnthaeufen:
		return s.enthaeufen(danach, stelle)

	case BefehlSchaden:
		// **Markierter Schaden** (G9, #1592) — CR 120.3, und er geht im
		// Aufräumschritt weg (CR 514.2).
		return s.schadenbefehl(danach, stelle, befehl)

	default:
		return nil, fmt.Errorf(
			"%w: %q ist kein Befehl, den der Sandkasten kennt",
			domain.ErrValidation, befehl.Art)
	}

	return map[string]any{}, nil
}

// zaehlen legt Zähler auf ein Objekt oder nimmt sie herunter.
func (s *Sandkasten) zaehlen(
	objekt *domain.Spielobjekt, befehl domain.Spielbefehl,
) (map[string]any, error) {
	art, _ := befehl.Nutzlast["art"].(string) //nolint:errcheck // dito
	if art == "" {
		return nil, fmt.Errorf(
			"%w: ein Zähler ohne Art lässt sich nicht ablegen",
			domain.ErrValidation)
	}

	// **`float64`, weil JSON keine ganzen Zahlen kennt.** Wer hier `int`
	// erwartete, bekäme aus jeder Fläche eine Null.
	menge, _ := befehl.Nutzlast["menge"].(float64) //nolint:errcheck // dito

	if objekt.Zaehler == nil {
		objekt.Zaehler = map[string]int{}
	}

	neu := objekt.Zaehler[art] + int(menge)

	// **Unter null gibt es keine Zähler.** Ein Befehl, der mehr abnehmen
	// will, als daliegt, nimmt alles — und sagt, wie viele es wurden.
	if neu < 0 {
		neu = 0
	}

	if neu == 0 {
		delete(objekt.Zaehler, art)
	} else {
		objekt.Zaehler[art] = neu
	}

	return map[string]any{"art": art, "jetzt": float64(neu)}, nil
}

// anhaengen hängt ein Objekt an ein anderes.
func (s *Sandkasten) anhaengen(
	objekt *domain.Spielobjekt, befehl domain.Spielbefehl,
) (map[string]any, error) {
	an, _ := befehl.Nutzlast["an"].(string) //nolint:errcheck // dito

	if an == "" {
		return nil, fmt.Errorf(
			"%w: ein Anhängen ohne Ziel hängt nirgends", domain.ErrValidation)
	}

	objekt.HaengtAn = an

	return map[string]any{"an": an}, nil
}

// haeufen legt ein Objekt auf den Haufen eines anderen — CR 400.5.
//
// **Die Regel steht in der Domäne** (`domain.Partie.Haeufe`): *ob* zwei
// Objekte einen Haufen bilden dürfen, ist eine Aussage über die Partie und
// nicht über diesen Dienst. Hier steht nur, wie der Befehl ankommt.
func (s *Sandkasten) haeufen(
	danach *domain.Partie, stelle int, befehl domain.Spielbefehl,
) (map[string]any, error) {
	worauf, _ := befehl.Nutzlast["worauf"].(string) //nolint:errcheck // dito
	if worauf == "" {
		return nil, fmt.Errorf(
			"%w: ein Haufen braucht etwas, worauf gelegt wird",
			domain.ErrValidation)
	}

	gehaeuft, fehler := danach.Haeufe(danach.Objekte[stelle].Kennung, worauf)
	if fehler != nil {
		return nil, fehler
	}

	*danach = gehaeuft

	return map[string]any{"worauf": worauf}, nil
}

// enthaeufen nimmt ein Objekt aus seinem Haufen.
func (s *Sandkasten) enthaeufen(
	danach *domain.Partie, stelle int,
) (map[string]any, error) {
	gelockert, fehler := danach.Enthaeufe(danach.Objekte[stelle].Kennung)
	if fehler != nil {
		return nil, fehler
	}

	*danach = gelockert

	return map[string]any{}, nil
}
