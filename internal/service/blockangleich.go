package service

import (
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Die Blöcke einer Art — `card:` oder `keyword:` — gegen den Bestand.
//
// **Ein Weg für beide**, seit ENG-03
// (#1728).
// Er stand zweimal da, fast Zeile für Zeile, und **derselbe Fehler kam
// zweimal**: erst zählte der Kartenweg mit, statt nach der Stelle zuzuordnen
// (#621), dann der Schlüsselwortweg — angestrichen vom Zweitblick über Phase W.
// Die Regel steht jetzt einmal; was eine Art von der anderen unterscheidet, sind
// ihre Zuordnung, ihr Angleichen und ihre neuen Blöcke.

// bloeckeAngleichen bringt die Blöcke einer Art auf den Stand des Bestands.
//
// **Zugeordnet wird über die Stelle**: der n-te Block dieser Art gehört zum
// Eintrag mit der Stelle n, und `nachStelle` nennt ihn. **Ein Block ohne
// Eintrag ist ein gelöschter und geht mit** — ihn stehen zu lassen schriebe
// zurück, was jemand entfernt hat, und die Rücknahme dafür heißt ⌘Z, nicht
// Speichern.
//
// **Neue gehen hinter den letzten, den es gab** — auch hinter einen gelöschten,
// sonst wanderten sie hinter einen Block, den niemand versteht, sobald der
// letzte alte weg ist. Gab es keinen, gehen sie ans Ende.
func bloeckeAngleichen[T any](
	baum *msepaket.Baum,
	istBlock func(*msepaket.Knoten) bool,
	nachStelle map[int]T,
	angleichen func(*msepaket.Knoten, T),
	neue func(ab int, zeilenende string) []*msepaket.Knoten,
) {
	behalten := make([]*msepaket.Knoten, 0, len(baum.Wurzel))

	stelle, hinterDemLetzten, zeilenende := 0, -1, "\n"

	for _, knoten := range baum.Wurzel {
		if !istBlock(knoten) {
			behalten = append(behalten, knoten)

			continue
		}

		if knoten.Zeilenende != "" {
			zeilenende = knoten.Zeilenende
		}

		eintrag, da := nachStelle[stelle]
		stelle++

		if da {
			angleichen(knoten, eintrag)
			behalten = append(behalten, knoten)
		}

		hinterDemLetzten = len(behalten)
	}

	if hinterDemLetzten < 0 {
		hinterDemLetzten = len(behalten)
	}

	baum.Wurzel = slices.Insert(behalten, hinterDemLetzten, neue(stelle, zeilenende)...)
}
