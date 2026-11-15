package app

// Eine neue Karte — #562.
//
// # Die Lücke, die das schließt
//
// Der Bestand konnte Sets **einlesen** und Feldwerte ändern. Eine neue Karte
// gab es nirgends: nicht im Speicher, nicht im Dienst, nicht in der
// Oberfläche. Wer ein Set anlegen wollte, musste es woanders schreiben und hier
// öffnen.
//
// # Das Löschen steht nicht daneben
//
// Es gehört in denselben Zug und geht nicht ohne **Rücknahme**: eine gelöschte
// Karte ohne ⌘Z ist Datenverlust mit einem Klick. Der Verlauf trägt heute nur
// Feldänderungen; ein Schritt „Karte entfernt" bräuchte eine eigene Art und
// müsste die Kennung zurückbringen, auf die die Verlaufszeilen zeigen. Das
// steht in #562.

import (
	"context"
	"strconv"
)

// Kartenanleger legt eine Karte an — die Naht zum Speicher.
type Kartenanleger interface {
	KarteAnlegen(ctx context.Context, setID int64) (int64, error)
}

// AnlegeAntwort trägt die Kennung der neuen Karte oder einen Fehler.
type AnlegeAntwort struct {
	// KarteID ist die neue Karte. **Sie kommt zurück**, denn wer eine Karte
	// anlegt, will sie aufschlagen — und die Oberfläche soll dafür nicht die
	// ganze Liste durchsuchen müssen.
	KarteID string  `json:"karteID"`
	Fehler  *Fehler `json:"fehler"`
}

// KarteAnlegen hängt eine leere Karte an ein Set.
func (a *App) KarteAnlegen(ctx context.Context, setID string) AnlegeAntwort {
	ctx = a.imProfil(ctx)

	// **Am Feld `setID` und nicht am gemeinsamen `id`** — dieselbe
	// Unterscheidung wie auf den sechs Wegen daneben (#808, #832): eine
	// Kartenänderung trägt zwei Kennungen, und der gemeinsame Name sagte
	// nicht, welche gemeint ist.
	kennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		return AnlegeAntwort{Fehler: befund}
	}

	neue, err := a.anleger.KarteAnlegen(ctx, kennung)
	if err != nil {
		return AnlegeAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return AnlegeAntwort{KarteID: strconv.FormatInt(neue, 10)}
}

// MitKartenanleger setzt die Naht — für Tests und den Composition Root.
func (a *App) MitKartenanleger(anleger Kartenanleger) { a.anleger = anleger }
