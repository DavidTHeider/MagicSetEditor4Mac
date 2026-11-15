package app

// **Die Schnittstellen dieses Rechners am Fenster** — G08
// (#2272).
//
// Für die Wahl, wo die eigenen Dienste lauschen: Name, Art, Zustand und
// Adressen jeder Schnittstelle — auch einer ausgeschalteten. Gewählt wird
// dann in `netz.lauschen`, über denselben Weg wie jede Einstellung.
//
// **Nur am Fenster** — sie dient der Wahl dort; an der Kommandozeile nennt
// `ifconfig` dieselben Schnittstellen, und die Wahl selbst lesen und setzen
// `mse einstellungen` und MCP wie jede andere Einstellung.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/entdeckung"
)

// NetzschnittstelleDTO ist eine Schnittstelle dieses Rechners.
type NetzschnittstelleDTO struct {
	Name     string   `json:"name"`
	Art      string   `json:"art"`
	Aktiv    bool     `json:"aktiv"`
	Adressen []string `json:"adressen"`
	Befund   string   `json:"befund"`
}

// NetzschnittstellenAntwort trägt die Liste oder einen Fehler.
type NetzschnittstellenAntwort struct {
	Schnittstellen []NetzschnittstelleDTO `json:"schnittstellen"`
	Fehler         *Fehler                `json:"fehler"`
}

// MitNetzschnittstellen setzt die Naht zur Liste — für Läufe ohne die
// Schnittstellen dieses Rechners.
func (a *App) MitNetzschnittstellen(lesen func() ([]entdeckung.Schnittstelleneintrag, error)) *App {
	a.netzliste = lesen

	return a
}

// Netzschnittstellen liest die Schnittstellen dieses Rechners.
func (a *App) Netzschnittstellen(ctx context.Context) NetzschnittstellenAntwort {
	ctx = a.imProfil(ctx)

	lesen := a.netzliste
	if lesen == nil {
		lesen = entdeckung.Schnittstellenliste
	}

	liste, fehler := lesen()
	if fehler != nil {
		return NetzschnittstellenAntwort{Schnittstellen: []NetzschnittstelleDTO{}, Fehler: Uebersetze(ctx, fehler)}
	}

	dtos := make([]NetzschnittstelleDTO, 0, len(liste))
	for _, eine := range liste {
		dtos = append(dtos, NetzschnittstelleDTO{
			Name: eine.Name, Art: eine.Art, Aktiv: eine.Aktiv, Adressen: eine.Adressen, Befund: eine.Befund,
		})
	}

	return NetzschnittstellenAntwort{Schnittstellen: dtos}
}
