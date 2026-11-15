package service

import (
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// gemessenerWert liest einen Feldwert **so, wie er gezeichnet wird**.
//
// Aus den bereits mit Vorgabewerten ergänzten Feldern, und ohne Auszeichnungen.
// Der Name ist der des Kartenfelds mit Leerzeichen; die Umgebung führt ihn mit
// Unterstrichen.
func gemessenerWert(felder []msescript.Eintrag, feld string) ([]rohabschnitt, bool) {
	gesucht := strings.ReplaceAll(feld, " ", "_")

	for _, eintrag := range felder {
		if eintrag.Name != gesucht {
			continue
		}

		// **Gemessen wird der Klartext.** Die Inhaltsbreite fragt, wie breit
		// der Zug ist; welcher Teil davon kursiv steht, ändert die Breite —
		// aber die Kursivschrift kennt diese Schicht nicht (§2), und eine
		// Messung mit der aufrechten ist die ehrlichere Näherung als keine.
		roh, _ := ohneAuszeichnungen(feld, eintrag.Wert.AlsText())

		return roh, true
	}

	return nil, false
}

// klartextVon fügt die Texte roher Abschnitte zusammen.
func klartextVon(roh []rohabschnitt) string {
	var bau strings.Builder
	for _, teil := range roh {
		bau.WriteString(teil.Text)
	}

	return bau.String()
}
