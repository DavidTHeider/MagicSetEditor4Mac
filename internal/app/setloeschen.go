// **Ein Set löschen über die Naht** —
// #1218.

package app

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Setloescher nimmt ein Set aus dem Bestand.
type Setloescher interface {
	Loeschen(ctx context.Context, setID int64, name string) ([]domain.Befund, error)
}

// SetLoeschAntwort ist der Ausgang eines Löschens.
//
// **Befunde ohne Fehler sind der Regelfall dieser Antwort** (§7, Regel 3): das
// Set ist weg, und was daneben liegen blieb, steht darin. Wer beides in ein
// Feld legte, machte aus „aufgeräumt bis auf die Bilder" ein
// „fehlgeschlagen".
type SetLoeschAntwort struct {
	Befunde []string `json:"befunde"`
	Fehler  *Fehler  `json:"fehler"`
}

// MitSetloeschung setzt die Naht.
func (a *App) MitSetloeschung(loescher Setloescher) { a.setloeschung = loescher }

// SetLoeschen nimmt ein Set aus dem Bestand — endgültig.
//
// **Der Name ist die Bestätigung** und wird durchgereicht, nicht hier geprüft:
// die Prüfung gehört an die Stelle, an der sie mit dem Löschen in **einer**
// Transaktion steht (`storage`). Eine zweite hier sähe nach Sorgfalt aus und
// wäre ein Wettlauf.
func (a *App) SetLoeschen(ctx context.Context, setID, name string) SetLoeschAntwort {
	ctx = a.imProfil(ctx)

	if a.setloeschung == nil {
		return SetLoeschAntwort{Befunde: []string{}, Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: das Löschen ist nicht eingerichtet", domain.ErrValidation))}
	}

	kennung, fehler := kennungAus(setID)
	if fehler != nil {
		return SetLoeschAntwort{Befunde: []string{}, Fehler: fehler}
	}

	befunde, loeschfehler := a.setloeschung.Loeschen(ctx, kennung, name)
	if loeschfehler != nil {
		return SetLoeschAntwort{Befunde: []string{}, Fehler: Uebersetze(ctx, loeschfehler)}
	}

	return SetLoeschAntwort{Befunde: befundtexte(befunde)}
}
