package service

// **Handlungen mit Schlüsselwort** — Werkstück M7 der Regelmaschine
// (#1609).
//
// # Acht von neunundsechzig
//
// **Die, die das Regelwerk allein beschreibt** — welche das sind und warum,
// steht im Kopf von `internal/domain/schluesselhandlungen.go`.

import (
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// istHandlung sagt, ob dieser Befehl eine Handlung mit Schlüsselwort ist.
func istHandlung(art string) bool {
	switch art {
	case BefehlZerstoeren, BefehlOpfern, BefehlExilieren,
		BefehlNeutralisieren, BefehlKaempfen,
		BefehlMahlen, BefehlHellsehen, BefehlUeberwachen:
		return true
	}

	return false
}

// handlungsbefehl führt eine Handlung mit Schlüsselwort aus — CR 701.
func (s *Sandkasten) handlungsbefehl(
	partie domain.Partie, befehl domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	switch befehl.Art {
	case BefehlMahlen, BefehlHellsehen, BefehlUeberwachen:
		return s.bibliothekshandlung(partie, befehl)
	}

	return s.objekthandlung(partie, befehl)
}

// objekthandlung führt die fünf aus, die ein Objekt meinen.
func (s *Sandkasten) objekthandlung(
	partie domain.Partie, befehl domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	objekt, steht := befehl.Nutzlast["objekt"].(string)
	if !steht {
		return nil, domain.Partie{}, fmt.Errorf(
			"%w: %q braucht die Objektkennung in \"objekt\"",
			domain.ErrValidation, befehl.Art)
	}

	danach, fehler := partie, error(nil)

	switch befehl.Art {
	case BefehlZerstoeren:
		danach, fehler = partie.Zerstoeren(objekt)

	case BefehlOpfern:
		danach, fehler = partie.Opfern(befehl.Absender, objekt)

	case BefehlExilieren:
		danach, fehler = partie.Exilieren(objekt)

	case BefehlNeutralisieren:
		danach, fehler = partie.Neutralisieren(objekt)

	case BefehlKaempfen:
		danach, fehler = s.kaempfen(partie, befehl, objekt)
	}

	return handlungsereignis(danach, fehler, befehl,
		map[string]any{"objekt": objekt})
}

// kaempfen liest den zweiten Kämpfer aus der Nutzlast — CR 701.14a.
func (s *Sandkasten) kaempfen(
	partie domain.Partie, befehl domain.Spielbefehl, erste string,
) (domain.Partie, error) {
	zweite, steht := befehl.Nutzlast["gegen"].(string)
	if !steht {
		return partie, fmt.Errorf(
			"%w: %q braucht die zweite Kreatur in \"gegen\"",
			domain.ErrValidation, befehl.Art)
	}

	return partie.Kaempfen(erste, zweite)
}

// bibliothekshandlung führt die drei aus, die eine Bibliothek meinen.
//
// **Der Absender ist der Spieler** — CR 701.17a und CR 701.22a sprechen von
// „einem Spieler" und meinen den, der handelt. Eine fremde Bibliothek zu
// mahlen braucht einen Effekt, der es sagt, und der ist `M8`.
func (s *Sandkasten) bibliothekshandlung(
	partie domain.Partie, befehl domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	wie, fehler := ganzeZahl(befehl, "wie")
	if fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	if befehl.Art == BefehlMahlen {
		gemahlen, gescheitert := partie.Mahlen(befehl.Absender, wie)

		return handlungsereignis(gemahlen, gescheitert, befehl,
			map[string]any{"wie": float64(wie)})
	}

	gewaehlt, fehler := kennungsliste(befehl, "karten")
	if fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	sortieren := domain.Partie.Hellsehen
	if befehl.Art == BefehlUeberwachen {
		sortieren = domain.Partie.Ueberwachen
	}

	danach, fehler := sortieren(partie, befehl.Absender, wie, gewaehlt)

	// **Wie viele gewählt wurden, nicht welche** — was jemand sieht, bleibt
	// seine Sache: die Bibliothek ist auch für ihren Eigner verdeckt (G1),
	// und ein Ereignis mit Kennungen verriete sie allen.
	return handlungsereignis(danach, fehler, befehl, map[string]any{
		"wie": float64(wie), "gewaehlt": float64(len(gewaehlt)),
	})
}

// kennungsliste liest eine Liste von Kennungen aus der Nutzlast.
//
// **Ein fehlendes Feld ist keine leere Liste, sondern keine** — wer
// hellsieht und nichts wegsortiert, schickt gar nichts mit.
func kennungsliste(befehl domain.Spielbefehl, feld string) ([]string, error) {
	roh, steht := befehl.Nutzlast[feld]
	if !steht {
		return nil, nil
	}

	liste, ist := roh.([]any)
	if !ist {
		return nil, fmt.Errorf("%w: %q ist keine Liste",
			domain.ErrValidation, feld)
	}

	kennungen := make([]string, 0, len(liste))

	for _, eines := range liste {
		kennung, ist := eines.(string)
		if !ist {
			return nil, fmt.Errorf(
				"%w: in %q steht %v statt einer Kennung",
				domain.ErrValidation, feld, eines)
		}

		kennungen = append(kennungen, kennung)
	}

	return kennungen, nil
}

// handlungsereignis macht aus dem Ausgang ein Ereignis.
func handlungsereignis(
	danach domain.Partie, fehler error, befehl domain.Spielbefehl,
	dazu map[string]any,
) ([]domain.Ereignis, domain.Partie, error) {
	if fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	oeffentlich := map[string]any{"sitz": befehl.Absender}
	for name, wert := range dazu {
		oeffentlich[name] = wert
	}

	return []domain.Ereignis{{
		Art: befehl.Art, Oeffentlich: oeffentlich,
	}}, danach, nil
}
