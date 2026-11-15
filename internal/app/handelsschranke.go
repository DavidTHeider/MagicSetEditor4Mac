package app

// **Die Handelsrichtlinie aus den Einstellungen** — #2435.
//
// **Über den Stand und nicht über den rohen Wert**, wie bei der Ortsregel der
// KI ([App.ortsregel]): die Faltung über Gerät, Nutzer und Profil macht
// `domain`, und was hier ankommt, ist der Wert, der **gilt**. Hier wird nur
// übersetzt — die Schranke selbst steht in `domain`, wann sie gefragt wird,
// in `service` (`handelsschranke.go`).
//
// **Ohne Bestand keine Richtlinie und kein Handel**: ein Fehler hier hält den
// Handel an, statt ihn ohne Schranke durchzulassen.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einstellungsschranke liest die Handelsrichtlinie des Profils dieses Rufs.
type einstellungsschranke struct{ a *App }

func (s einstellungsschranke) Handelsrichtlinie(ctx context.Context) (domain.Handelsrichtlinie, error) {
	if s.a.einstellungen == nil {
		return domain.Handelsrichtlinie{}, fehlenderBestand()
	}

	stand, fehler := s.a.einstellungen.Stand(ctx)
	if fehler != nil {
		return domain.Handelsrichtlinie{}, fehler
	}

	werte := make(map[string]string, len(stand))
	for schluessel, wahl := range stand {
		werte[schluessel] = wahl.Wirksam
	}

	return domain.HandelsrichtlinieAus(werte), nil
}
