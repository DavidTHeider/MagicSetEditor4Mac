package domain

// **Welche eigene Übersetzung gilt** — LOC-02
// (#1688).
//
// **Eine Einstellung neben `sprache` und nicht ein Wert darin.** `sprache`
// ist eine der acht Fenstersprachen, und an ihr hängen Menüleiste und
// Dialoge; eine eigene Übersetzung liegt **über** ihrer Basis. Gibt es sie
// nicht mehr, gilt wieder `sprache` — die Einstellung zeigt dann ins Leere,
// und das Fenster fällt zurück, statt leer zu stehen (AK-03).

import "errors"

// EinstellungEigeneSprache ist die Kennung der eigenen Übersetzung — leer
// heißt: keine.
const EinstellungEigeneSprache = "sprache.eigene"

// PruefeEigeneSprache urteilt über ihren Wert — leer oder eine Kennung, die
// eine eigene sein kann. **Ob es das Paket gibt, sagt der Bestand**, und ein
// fehlendes ist kein Fehler, sondern der Rückfall.
func PruefeEigeneSprache(wert string) error {
	if wert == "" {
		return nil
	}

	var feld *FieldError
	if errors.As(PruefeSprachkennung(wert), &feld) {
		return NewFieldError("wert", feld.Key, "%s", feld.Message).MitWerten(feld.Werte)
	}

	return nil
}
