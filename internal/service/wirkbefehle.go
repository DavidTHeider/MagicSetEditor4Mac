package service

// **Wirken und Landspielen** — Werkstück M4 der Regelmaschine
// (#1606).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **`manabefehle.go` beantwortet, wie Mana entsteht**; hier steht, **wie man
// es ausgibt** — und wie ein Land auf den Tisch kommt, das gar nichts kostet.

import (
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// wirkbefehl legt einen Zauber auf den Stapel oder ein Land auf den Tisch.
func (s *Sandkasten) wirkbefehl(
	partie domain.Partie, befehl domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	objekt, steht := befehl.Nutzlast["objekt"].(string)
	if !steht {
		return nil, domain.Partie{}, fmt.Errorf(
			"%w: %q braucht die Objektkennung in \"objekt\"",
			domain.ErrValidation, befehl.Art)
	}

	if befehl.Art == BefehlLandspielen {
		danach, fehler := partie.LandSpielen(befehl.Absender, objekt)
		if fehler != nil {
			return nil, domain.Partie{}, fehler
		}

		return wirkereignis(danach, befehl, objekt, nil)
	}

	ziele, fehler := zielkennungen(befehl)
	if fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	danach, fehler := partie.Wirken(befehl.Absender, objekt, ziele)
	if fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	// **Gelesen wird beim Verrechnen und nicht hier** — CR 608.2c
	// (#1625).
	//
	// Zwischen Ankündigen und Verrechnen liegt mindestens eine
	// Prioritätsrunde. Was die Karte sagt, liest [domain.Wirkungsdeuter] —
	// eingesetzt vom [Sandkasten], gerufen von [domain.Partie.Verrechnen].
	return wirkereignis(danach, befehl, objekt, ziele)
}

// zielkennungen liest die Ziele aus der Nutzlast — CR 115.1.
//
// # Warum jedes einzeln geprüft wird
//
// **Weil JSON keine Liste von Zeichenketten kennt, sondern eine Liste von
// irgendetwas.** Eine Zahl darin käme sonst als leere Kennung durch, und die
// Domäne meldete „gibt es nicht" — ein Satz, der von der falschen Sache
// spricht (§8).
//
// **Keine Ziele ist der Normalfall**: die meisten Zauber haben keines.
func zielkennungen(befehl domain.Spielbefehl) ([]string, error) {
	roh, steht := befehl.Nutzlast["ziele"]
	if !steht {
		return nil, nil
	}

	liste, ist := roh.([]any)
	if !ist {
		return nil, fmt.Errorf("%w: \"ziele\" ist keine Liste",
			domain.ErrValidation)
	}

	ziele := make([]string, 0, len(liste))

	for _, eines := range liste {
		kennung, ist := eines.(string)
		if !ist {
			return nil, fmt.Errorf(
				"%w: ein Ziel muss eine Kennung sein, hier steht %v",
				domain.ErrValidation, eines)
		}

		ziele = append(ziele, kennung)
	}

	return ziele, nil
}

// wirkereignis macht aus dem Ausgang ein Ereignis.
//
// # Warum die Ziele öffentlich sind
//
// **Weil CR 115.1 sie zum Teil des Hinaufstellens macht.** Wer nicht sieht,
// worauf ein Zauber zeigt, kann nicht entscheiden, ob er antworten will — und
// genau dafür gibt es die Priorität (CR 117.1).
// **Ohne Fehlerweg**: beide Aufrufer prüfen selbst, und ein Fehlerzweig
// hier wäre einer, den keiner von ihnen betritt (§10).
func wirkereignis(
	danach domain.Partie, befehl domain.Spielbefehl,
	objekt string, ziele []string,
) ([]domain.Ereignis, domain.Partie, error) {
	oeffentlich := map[string]any{
		"sitz": befehl.Absender, "objekt": objekt,
	}

	if len(ziele) > 0 {
		oeffentlich["ziele"] = ziele
	}

	// **Kein Befund an diesem Ereignis**: seit CR 608.2c hier richtig sitzt
	// (#1625),
	// entsteht ein Befund beim **Verrechnen** und steht am Ereignis des
	// Passens. Ein Feld, das kein Weg mehr füllt, ist eine Zusage, die
	// niemand einlöst (§10).

	return []domain.Ereignis{{
		Art: befehl.Art, Oeffentlich: oeffentlich,
	}}, danach, nil
}
