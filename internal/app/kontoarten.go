// **Die vier Kontoarten am Fenster und an den drei Flächen** — Werkstück 6 von
// Phase 4.12
// (#1351),
// nach der Kommentierung vom 2026-09-05.
//
// # Was hier **nicht** geschieht
//
// **Kein Vergleich und keine Prüfung.** Diese Datei nimmt drei Zeichenketten
// entgegen und gibt einen Nutzer zurück; wer hereinkommt, entscheidet
// `internal/service` (§2).
//
// **Und das Kennwort geht nicht zurück.** Es kommt herein und bleibt hier —
// [NutzerDTO] trägt kein Feld dafür, und `TestDasKennwortErreichtDieAntwortNie`
// hält das über die JSON-Naht fest und nicht über den Go-Typ.

package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Kontofuehrer ist die Naht zu `internal/service`.
type Kontofuehrer interface {
	Eintragen(ctx context.Context, wunsch service.Kontowunsch) (domain.Nutzer, error)
	Anmelden(ctx context.Context, kennung, kennwort, nachweis string) (domain.Nutzer, error)
	Anonym(ctx context.Context) (domain.Nutzer, error)
}

// MitKontoarten setzt die Naht.
func (a *App) MitKontoarten(fuehrer Kontofuehrer) { a.kontoarten = fuehrer }

// KontoEintragen legt ein Konto unter einer Adresse oder einer Rufnummer an.
//
// **Der Name ist wahlfrei** — leer heißt: einen aus der Kennung. Er steht
// trotzdem in der Naht, weil das Fenster ihn anbietet und niemand danach
// nochmal fragen soll.
func (a *App) KontoEintragen(
	ctx context.Context, kennung, kennwort, name string,
) EinNutzerAntwort {
	ctx = a.imProfil(ctx)

	eingetragen, fehler := a.kontoarten.Eintragen(ctx, service.Kontowunsch{
		Name: name, Kennung: kennung, Kennwort: kennwort,
	})
	if fehler != nil {
		return EinNutzerAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return a.nutzerNach(ctx, eingetragen.ID)
}

// KontoAnmelden lässt einen Nutzer unter seiner Kennung herein.
//
// **Der Nachweis ist der zweite Faktor** — Werkstück 7 (#1351): ein Code aus
// der Authenticator-App, ein Wiederherstellungscode, oder **leer** für Touch
// ID. Wer keinen eingerichtet hat, wird nicht gefragt.
func (a *App) KontoAnmelden(
	ctx context.Context, kennung, kennwort, nachweis string,
) EinNutzerAntwort {
	ctx = a.imProfil(ctx)

	herein, fehler := a.kontoarten.Anmelden(ctx, kennung, kennwort, nachweis)
	if fehler != nil {
		return EinNutzerAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return a.nutzerNach(ctx, herein.ID)
}

// KontoAnonym legt ein Konto ohne jede Angabe an.
//
// **Ohne Argument, weil es nichts anzugeben gibt** — das ist der Punkt. Wer
// eines will, drückt einen Knopf und arbeitet.
func (a *App) KontoAnonym(ctx context.Context) EinNutzerAntwort {
	ctx = a.imProfil(ctx)

	angelegt, fehler := a.kontoarten.Anonym(ctx)
	if fehler != nil {
		return EinNutzerAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return a.nutzerNach(ctx, angelegt.ID)
}
