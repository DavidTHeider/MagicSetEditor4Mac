package app

// **Was einem Set an Paketen fehlt** — DOM-01.AK-04
// (#1696).
//
// Die Anwendung reicht durch; gerechnet wird in `service`
// (`Paketverwalter.Setbedarf`). Die Antwort hat die Gestalt der
// Abhängigkeiten ([AbhaengigkeitenAntwort]), denn es sind dieselben Befunde —
// nur von einem Set aus gesucht statt über den ganzen Katalog.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Paketbedarf nennt, was dem Set an Paketen fehlt oder zu alt ist.
func (a *App) Paketbedarf(ctx context.Context, setID string) AbhaengigkeitenAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := kennungAus(setID)
	if befund != nil {
		return AbhaengigkeitenAntwort{Befunde: []AbhaengigkeitDTO{}, Fehler: befund}
	}

	befunde, err := a.pakete.Setbedarf(ctx, kennung)
	if err != nil {
		return AbhaengigkeitenAntwort{Befunde: []AbhaengigkeitDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	return AbhaengigkeitenAntwort{Befunde: alsAbhaengigkeiten(befunde)}
}

// alsAbhaengigkeiten übersetzt Befunde — **eine Stelle** für den Katalog,
// den Plan und den Bedarf eines Sets.
func alsAbhaengigkeiten(befunde []service.Abhaengigkeitsbefund) []AbhaengigkeitDTO {
	dtos := make([]AbhaengigkeitDTO, 0, len(befunde))
	for _, befund := range befunde {
		dtos = append(dtos, AbhaengigkeitDTO{
			Paket: befund.Paket, Braucht: befund.Braucht, Art: string(befund.Art),
			Verlangt: befund.Verlangt, Da: befund.Da,
		})
	}

	return dtos
}
