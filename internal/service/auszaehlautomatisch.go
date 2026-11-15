package service

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// automatischeDimensionen gibt je Kartenfeld mit Statistik eine Dimension —
// #1985.
//
// Portiert aus `StatsDimension::StatsDimension(const Field&)`
// (`src/data/statistics.cpp`, Zeilen 27–56) und `Game::validate`
// (`src/data/game.cpp`, Zeilen 68–77; Fassung f9d9356 vom 2021-07-06): ein
// Auswahlfeld rechnet `primary_choice(card.<feld>)`, gruppiert nach seinen
// Wahlen der obersten Ebene und färbt nach `choice colors:`; jedes andere
// rechnet `card.<feld>`.
//
// **Eine erklärte Dimension gleichen Namens geht vor**, und die automatische
// entfällt — eine Abweichung mit Absicht: MSE zeigte beide, und eine Auswahl
// nach Namen fände dann nur noch die automatische. Im Pack kommt es nicht
// vor (gemessen am 2026-10-01).
// mitAutomatischen stellt die automatischen Dimensionen vor die erklärten —
// MSEs `insert(begin(), …)`, „push front".
func mitAutomatischen(felder []domain.Field, erklaert []Auszaehlangabe) []Auszaehlangabe {
	automatische := automatischeDimensionen(felder, erklaert)

	// **Nie `nil`**, auch ohne jede Dimension — über die JSON-Naht würde
	// daraus `null` statt einer leeren Liste (§3).
	alle := make([]Auszaehlangabe, 0, len(automatische)+len(erklaert))

	return append(append(alle, automatische...), erklaert...)
}

func automatischeDimensionen(felder []domain.Field, erklaert []Auszaehlangabe) []Auszaehlangabe {
	vergeben := make(map[string]bool, len(erklaert))
	for _, eine := range erklaert {
		vergeben[eine.Name] = true
	}

	var dimensionen []Auszaehlangabe

	for _, feld := range felder {
		if !feld.Statistik || vergeben[feld.Name] {
			continue
		}

		eine := Auszaehlangabe{
			Name: feld.Name, Beschreibung: feld.Beschreibung,
			Skript: "card." + msepaket.Kanonisch(feld.Name),
		}

		if feld.Kind == domain.KindChoice {
			eine.Skript = "primary_choice(" + eine.Skript + ")"
			eine.Gruppen = feld.Choices
			eine.Farben = feld.Statistikfarben
		}

		dimensionen = append(dimensionen, eine)
	}

	return dimensionen
}
