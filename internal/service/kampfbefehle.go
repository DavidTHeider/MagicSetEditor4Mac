package service

// **Angreifen und Blocken** — Werkstück M6 der Regelmaschine
// (#1608).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan steht, wie etwas auf den Stapel kommt**; hier steht, **wer wen
// angreift**. Und beide Nutzlasten sind **Listen von Paaren** — die einzige
// Gestalt im ganzen Sandkasten, die das ist.

import (
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// kampfbefehl erklärt Angreifer oder Blocker.
func (s *Sandkasten) kampfbefehl(
	partie domain.Partie, befehl domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	if befehl.Art == BefehlBlocker {
		bloecke, fehler := bloeckeAus(befehl)
		if fehler != nil {
			return nil, domain.Partie{}, fehler
		}

		danach, fehler := partie.BlockerErklaeren(befehl.Absender, bloecke)

		return kampfereignis(danach, fehler, befehl, len(bloecke))
	}

	angriffe, fehler := angriffeAus(befehl)
	if fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	danach, fehler := partie.AngreiferErklaeren(befehl.Absender, angriffe)

	return kampfereignis(danach, fehler, befehl, len(angriffe))
}

// angriffeAus liest die Angriffe aus der Nutzlast — CR 508.1a.
func angriffeAus(befehl domain.Spielbefehl) ([]domain.Angriff, error) {
	paare, fehler := paareAus(befehl, "angriffe", "kreatur", "ziel")
	if fehler != nil {
		return nil, fehler
	}

	angriffe := make([]domain.Angriff, 0, len(paare))
	for _, paar := range paare {
		angriffe = append(angriffe,
			domain.Angriff{Kreatur: paar[0], Ziel: paar[1]})
	}

	return angriffe, nil
}

// bloeckeAus liest die Blöcke aus der Nutzlast — CR 509.1a.
func bloeckeAus(befehl domain.Spielbefehl) ([]domain.Block, error) {
	paare, fehler := paareAus(befehl, "bloecke", "kreatur", "angreifer")
	if fehler != nil {
		return nil, fehler
	}

	bloecke := make([]domain.Block, 0, len(paare))
	for _, paar := range paare {
		bloecke = append(bloecke,
			domain.Block{Kreatur: paar[0], Angreifer: paar[1]})
	}

	return bloecke, nil
}

// paareAus liest eine Liste von Paaren aus der Nutzlast.
//
// # Warum eine Stelle für beide
//
// **Weil die Gestalt dieselbe ist** und nur die Feldnamen anders heißen. Zwei
// Ableser wären zwei Stellen, an denen derselbe Fehlerweg auseinanderläuft
// (§2).
//
// **Ein fehlendes Feld ist keine leere Liste, sondern keine** — CR 508.1a
// lässt „keine Angreifer" ausdrücklich zu, und das ist der Normalfall.
func paareAus(
	befehl domain.Spielbefehl, feld, erstes, zweites string,
) ([][2]string, error) {
	roh, steht := befehl.Nutzlast[feld]
	if !steht {
		return nil, nil
	}

	liste, ist := roh.([]any)
	if !ist {
		return nil, fmt.Errorf("%w: %q ist keine Liste",
			domain.ErrValidation, feld)
	}

	paare := make([][2]string, 0, len(liste))

	for _, eines := range liste {
		paar, ist := eines.(map[string]any)
		if !ist {
			return nil, fmt.Errorf(
				"%w: in %q steht %v statt eines Paares aus %q und %q",
				domain.ErrValidation, feld, eines, erstes, zweites)
		}

		links, fehler := kennungAus(paar, feld, erstes)
		if fehler != nil {
			return nil, fehler
		}

		rechts, fehler := kennungAus(paar, feld, zweites)
		if fehler != nil {
			return nil, fehler
		}

		paare = append(paare, [2]string{links, rechts})
	}

	return paare, nil
}

// kennungAus liest eine Kennung aus einem Paar.
func kennungAus(paar map[string]any, feld, name string) (string, error) {
	wert, steht := paar[name].(string)
	if !steht || wert == "" {
		return "", fmt.Errorf("%w: in %q fehlt %q an einem Paar",
			domain.ErrValidation, feld, name)
	}

	return wert, nil
}

// kampfereignis macht aus dem Ausgang ein Ereignis.
//
// # Warum nur die Zahl und nicht die Paare
//
// **Weil die Paare im Tischbild stehen.** Wer wen angreift, sieht jeder am
// Kampfstand der Partie; ein Ereignis, das es noch einmal aufzählte, führte
// dieselbe Wahrheit zweimal — und liefe beim ersten Abräumen auseinander
// (CR 506.4).
func kampfereignis(
	danach domain.Partie, fehler error, befehl domain.Spielbefehl, zahl int,
) ([]domain.Ereignis, domain.Partie, error) {
	if fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	return []domain.Ereignis{{
		Art: befehl.Art,
		Oeffentlich: map[string]any{
			"sitz": befehl.Absender, "zahl": float64(zahl),
		},
	}}, danach, nil
}
