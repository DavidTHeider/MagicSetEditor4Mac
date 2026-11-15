// **Die Kartenauswahl an der Fassade** — Werkstück 4 von Phase 3.6
// (#1167).
//
// **Eine eigene Datei**, weil zwei Ausgabewege sie brauchen — der Druckbogen
// und die Exportvorlage — und weil beide über die Zeilengrenze gingen (§2).

package app

import (
	"fmt"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// kartenauswahlAus liest die Kennungen, die eine Fläche geschickt hat.
//
// **Als Zeichenketten hinein, als Zahlen heraus** — wie jede Kennung an dieser
// Naht: JavaScript rechnet mit `double`, und eine Kennung jenseits von 2⁵³ käme
// verändert an.
//
// **Eine unbrauchbare Kennung wird abgewiesen und nicht übergangen.** Wer sie
// stillschweigend wegließe, gäbe eine Auswahl aus, die jemand anders gemeint
// hat — und bei genau einer unbrauchbaren wäre das eine leere Auswahl, also
// das ganze Set.
func kartenauswahlAus(kennungen []string) (domain.Kartenauswahl, *Fehler) {
	if len(kennungen) == 0 {
		return domain.Kartenauswahl{}, nil
	}

	karten := make([]int64, 0, len(kennungen))

	for _, roh := range kennungen {
		zahl, err := strconv.ParseInt(roh, 10, 64)
		if err != nil {
			return domain.Kartenauswahl{}, Nutzerfehler(fmt.Errorf(
				"%w: %q ist keine Kartenkennung", domain.ErrValidation, roh))
		}

		karten = append(karten, zahl)
	}

	return domain.Kartenauswahl{Karten: karten}, nil
}
