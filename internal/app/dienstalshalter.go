package app

// **Hält der Dienst den Bestand, sagt das Zurückspielen es** — M02
// (#2346),
// Befund A-11 des Audits 0.9.0.

import (
	"context"
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// alsDienstHalter nennt den Dienst, wenn **er** es ist, der den Bestand hält
// — gefragt bei `lsof` (Zweitblick über Phase 2, Befund 8): bis dahin
// genügte, dass er lief, und das Zurückspielen riet, den Dienst zu
// widerrufen, wo ein anderes Programm den Bestand hielt.
func alsDienstHalter(ctx context.Context, pfad string, fehler error) error {
	if errors.Is(fehler, domain.ErrConflict) && dienstHaeltDatei(ctx, pfad) {
		return &domain.DienstHaeltBestand{Grund: fehler}
	}

	return fehler
}

// dienstHaeltDatei sagt, ob der eingerichtete Dienst `pfad` offen hält. **Ein
// Fehler von `lsof` heißt nein** — dann bleibt die allgemeine Meldung, die
// niemandem etwas zuschreibt.
func dienstHaeltDatei(ctx context.Context, pfad string) bool {
	laeufe, fehler := dienstprozesse(ctx)
	if fehler != nil {
		return false
	}

	for _, lauf := range laeufe {
		if haelt, fehler := haeltDatei(ctx, pfad, lauf.pid); fehler == nil && haelt {
			return true
		}
	}

	return false
}

// dienstHaelt übersetzt einen Dienst, der den Bestand hält — oder gibt nichts.
func dienstHaelt(fehler error) *Fehler {
	var dienst *domain.DienstHaeltBestand
	if !errors.As(fehler, &dienst) {
		return nil
	}

	return mitSatz(CodeKonflikt, domain.SchluesselBestandVomDienstGehalten, fehler)
}
