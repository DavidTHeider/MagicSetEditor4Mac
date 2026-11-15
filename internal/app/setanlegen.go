// **Ein neues, leeres Set über die Naht** —
// #891.

package app

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// SetAnlegeAntwort ist der Ausgang eines Anlegens.
//
// **Eigen und nicht [AnlegeAntwort]**: jene nennt eine **Karte**, und ein
// Verbund, dessen Feld `karteID` heißt und eine Setkennung trägt, ist eine
// Falschangabe, die niemand beim Lesen bemerkt.
type SetAnlegeAntwort struct {
	ID     string  `json:"id"`
	Fehler *Fehler `json:"fehler"`
}

// SetAnlegen legt ein leeres Set an und gibt seine Kennung.
//
// **Die Kennung kommt zurück**, denn wer ein Set anlegt, will es aufschlagen —
// dieselbe Überlegung wie bei [App.KarteAnlegen].
func (a *App) SetAnlegen(ctx context.Context, name, spiel, stil string) SetAnlegeAntwort {
	ctx = a.imProfil(ctx)

	if a.importeur == nil {
		return SetAnlegeAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: der Import ist nicht eingerichtet", domain.ErrValidation))}
	}

	kennung, err := a.importeur.Neu(ctx, name, spiel, stil)
	if err != nil {
		return SetAnlegeAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return SetAnlegeAntwort{ID: kennungText(kennung)}
}
