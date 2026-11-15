package service

// **Wie gesichert wird** — OPS-06.AK-02
// (#1723).
//
// # Warum im Kontext und nicht als Argument
//
// **Weil vier Wege dasselbe Kennwort brauchen** — Sichern, Beschauen, der
// Rücksprungvergleich und Zurückspielen — und jeder von ihnen hinter einer
// Naht einer Fläche steht, die es heute nicht kennt. Der Wunsch reist wie die
// Vorgangskennung und das gezeigte Profil: die Fläche setzt ihn, der Weg liest
// ihn, und keine Schicht dazwischen muss ihn durchreichen.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// Sicherungswunsch sagt, ob eine Sicherung ein Kennwort trägt.
type Sicherungswunsch struct {
	// Kennwort verschlüsselt die Sicherung — und öffnet eine verschlüsselte.
	// Leer heißt: ohne.
	Kennwort string

	// Klartext sagt ausdrücklich „ohne Kennwort", auch wenn der Bestand
	// selbst verschlüsselt ist. Ohne diese Angabe gibt ein verschlüsselter
	// Bestand keinen Klartext her.
	Klartext bool

	// Flaeche und Anlass stehen im Vermerk der Sicherung (OPS01, #2347) —
	// welche Fläche sicherte, und ob von Hand oder nach Plan. **Ein leerer
	// Anlass heißt von Hand**; eine leere Fläche besteht die Prüfung des
	// Vermerks nicht, und das soll sie: jede Fläche nennt sich.
	Flaeche domain.Flaeche
	Anlass  domain.Sicherungsanlass
}

type sicherungswunschSchluessel struct{}

// MitSicherungswunsch legt den Wunsch in den Kontext.
func MitSicherungswunsch(ctx context.Context, wunsch Sicherungswunsch) context.Context {
	return context.WithValue(ctx, sicherungswunschSchluessel{}, wunsch)
}

// sicherungswunschAus liest ihn — ohne Angabe: ohne Kennwort, nicht
// ausdrücklich im Klartext.
func sicherungswunschAus(ctx context.Context) Sicherungswunsch {
	wunsch, _ := ctx.Value(sicherungswunschSchluessel{}).(Sicherungswunsch) //nolint:errcheck // ohne Angabe gilt der leere Wunsch

	return wunsch
}

// SicherungVerschluesselt sagt, ob eine Datei eine Sicherung mit Kennwort ist
// — damit eine Fläche nach dem Kennwort fragen kann, bevor sie etwas zeigt.
func SicherungVerschluesselt(pfad string) bool { return sicherung.IstVerschluesselt(pfad) }
