// **`notes` steht genau einmal in der Umgebung** — Werkstück 6 von Phase 3.7
// (#1168,
// Befund Z1-01).
//
// # Was hier zusammengezogen ist
//
// Zwei Stellen bauen die Skriptumgebung einer Karte — [umgebungAus] für den
// Zeichenplan und [exportdaten] für die Exportvorlage. Beide hängten hinter
// die Feldwerte einen Eintrag `notes` aus `CardRow.Notes`, **zusätzlich** zu
// dem, den der Einleser aus der Datei gemacht hat. Bei einer Karte mit Notizen
// standen damit zwei Einträge desselben Namens da, und welcher galt, entschied
// die Reihenfolge in [msescript.Wert.Eintrag] — der erste.
//
// **Das war eine Zusage über eine Reihenfolge, die niemand sieht.** Wer die
// Zeilen tauschte, bekam an jeder Karte ein leeres `notes`, und
// `magic.mse-game` liest daraus die Kartennummer.
//
// **Jetzt steht die Regel an einer Stelle**: der Eintrag kommt aus der Liste,
// und der leere Rückfall nur, wenn die Karte keinen trägt.

package service

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// mitNotizen ergänzt `notes`, wenn die Karte keines mitbringt.
//
// **`notes` steht an jeder Karte, auch an einer ohne** — so hält es MSE
// (`REFLECT(notes)` in `IMPLEMENT_REFLECTION(Card)`, `src/data/card.cpp`,
// Zeile 91, Fassung `f9d9356` vom 2021-07-06), und fünf der neunzehn
// Exportvorlagen des Packs lesen es
// (#609).
//
// **`magic` erklärt gar kein Feld `notes`** — nur `sorted notes`; die
// Eigenschaft gibt es trotzdem.
func mitNotizen(
	felder []msescript.Eintrag, karte *domain.CardRow,
) []msescript.Eintrag {
	if _, hat := karte.Value(domain.Notizenfeld); hat {
		return felder
	}

	return append(felder, msescript.Eintrag{
		Name: domain.Notizenfeld, Wert: msescript.Text(""),
	})
}
