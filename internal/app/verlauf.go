package app

import (
	"context"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Verlaufsfuehrer ist, was `app` zum Zurücknehmen braucht.
type Verlaufsfuehrer interface {
	// Zuruecknehmen nimmt den jüngsten geltenden Schritt des Sets zurück —
	// oder mit `vorgang` **genau diesen** (#1393). Das Fenster gibt ihn nie
	// mit: ⌘Z heißt „das da eben".
	Zuruecknehmen(ctx context.Context, setID int64, vorgang string) (service.Bewegung, error)
	Wiederherstellen(ctx context.Context, setID int64, vorgang string) (service.Bewegung, error)

	// Verlauf gibt die Einträge des Sets, der älteste zuerst (U6, #1417).
	Verlauf(ctx context.Context, setID int64) ([]domain.Verlaufseintrag, error)
}

// BewegungAntwort sagt, was eine Rücknahme bewegt hat.
type BewegungAntwort struct {
	// Bewegt ist falsch, wenn es nichts zu bewegen gab — **kein Fehler**.
	Bewegt bool `json:"bewegt"`

	// KarteID ist die betroffene Karte, damit die Oberfläche **nur sie**
	// nachlädt (§7). Als Zeichenkette wie überall an dieser Naht: ein `int64`
	// überlebt den Weg durch JavaScript nicht ungeteilt.
	//
	// **Ohne `omitempty`.** Eine Bewegung, die es nicht gab, trägt hier `"0"`,
	// und das ist eine Angabe; ein fehlendes Feld wäre `undefined` und zwänge
	// das Frontend zu einer Fallunterscheidung, die `Bewegt` schon trifft (§3).
	KarteID string `json:"karteID"`

	Feld string `json:"feld"`

	// Umfang ist die Zahl der Karten, die der Schritt berührt hat — eins bei
	// einer Einzeländerung, viele bei einer Massenänderung.
	//
	// **Er steht hier, obwohl heute niemand ihn zeigt.** Die Schichten darunter
	// rechnen ihn aus, weil die Zusage „eine Massenänderung ist **ein** Schritt"
	// sonst nirgends nachprüfbar wäre; ihn an dieser Naht fallen zu lassen
	// hieße, sie erst hier wieder zu verlieren.
	Umfang int `json:"umfang"`

	// Urheber sagt, wessen Änderung bewegt wurde — `fenster` oder `server`.
	// Ohne ihn ist eine Rücknahme im Zwei-Schreiber-Betrieb eine Vermutung (#52).
	Urheber string `json:"urheber"`

	Fehler *Fehler `json:"fehler"`
}

// Zuruecknehmen macht die jüngste noch geltende Änderung rückgängig — ⌘Z.
//
// **Ohne Vorgangskennung, und das ist die richtige Antwort** (#1393): ⌘Z heißt
// „das da eben", und ein Fenster hat den Stand vor Augen. Die gezielte
// Rücknahme brauchen die Flächen ohne Fenster, wo ein zweiter Prozess
// dazwischenschreiben kann — sie steht an [Verlaufsweg.Zuruecknehmen].
//
// **Mit `vorgang` genau diesen Schritt** — Werkstück U6 von Phase 5.1
// (#1417): aus der Verlaufsliste heraus, wie es die drei anderen Flächen seit
// #1393 können. Leer heißt ⌘Z: der jüngste noch geltende Schritt des Sets.
func (a *App) Zuruecknehmen(ctx context.Context, setID, vorgang string) BewegungAntwort {
	ctx = a.imProfil(ctx)

	return a.bewege(ctx, setID, vorgang, a.verlauf.Zuruecknehmen)
}

// Wiederherstellen führt die jüngste zurückgenommene Änderung wieder aus — ⇧⌘Z.
func (a *App) Wiederherstellen(ctx context.Context, setID, vorgang string) BewegungAntwort {
	ctx = a.imProfil(ctx)

	return a.bewege(ctx, setID, vorgang, a.verlauf.Wiederherstellen)
}

// bewege übersetzt beide Richtungen — sie unterscheiden sich nur im Aufruf.
func (a *App) bewege(
	ctx context.Context, setID, vorgang string,
	richtung func(context.Context, int64, string) (service.Bewegung, error),
) BewegungAntwort {
	kennung, befund := kennungAus(setID)
	if befund != nil {
		return BewegungAntwort{Fehler: befund}
	}

	bewegung, err := richtung(ctx, kennung, vorgang)
	if err != nil {
		return BewegungAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return BewegungAntwort{
		Bewegt:  bewegung.Bewegt,
		KarteID: strconv.FormatInt(bewegung.KarteID, 10),
		Feld:    bewegung.Feld,
		Umfang:  bewegung.Umfang,
		Urheber: string(bewegung.Urheber),
	}
}

// MitVerlauf setzt die Naht zum Verlauf — für Tests und den Composition Root.
func (a *App) MitVerlauf(fuehrer Verlaufsfuehrer) { a.verlauf = fuehrer }
