package domain

import "fmt"

// Nur gibt das Schema mit den genannten Feldern — **leer heißt: alle**.
//
// **Für die Frage an das Modell, nicht für die Prüfung der Antwort** —
// AI-05 (#2042): eine kleine Aufgabe schickt nur Name oder Flavortext; was
// zurückkommt, prüft weiter das ganze Schema. Ein Feld, das das Paket nicht
// kennt oder das eine KI nicht setzen darf, ist ein Fehler und kein leerer
// Auftrag — sonst fragte ein Tippfehler das Modell nach nichts.
func (s KISchema) Nur(namen []string) (KISchema, error) {
	if len(namen) == 0 {
		return s, nil
	}

	eng := KISchema{Paket: s.Paket, Fassung: s.Fassung}

	for _, name := range namen {
		feld, gibt := s.feld(name)
		if !gibt || !feld.Schreibbar {
			return KISchema{}, fmt.Errorf("%w: nach %q kann eine KI nicht gefragt werden — "+
				"das Paket kennt es nicht, oder es ist nicht für sie", ErrValidation, name)
		}

		eng.Felder = append(eng.Felder, feld)
	}

	return eng, nil
}
