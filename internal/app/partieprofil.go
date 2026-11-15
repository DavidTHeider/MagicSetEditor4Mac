package app

// **Ein Sitz und sein Profil** — Werkstück G8 von Phase 5.5
// (#1465).
//
// # Wofür es diese Datei gibt
//
// **Für eine einzige Zusage aus der Abnahme**: „Ein Profil, das es nicht mehr
// gibt, macht die Partie nicht unbenutzbar: der Sitz bleibt, seine Bindung
// wird als Befund benannt."
//
// **Ein Befund ist kein Fehler** (§8). Wer daraus einen machte, machte eine
// Partie unbenutzbar, weil jemand ein Profil gelöscht hat — und die Karten
// stünden noch da.
//
// # Warum das hier steht und nicht in `service`
//
// **Weil es eine Auskunft über diesen Rechner ist und keine Spielregel.** Ob
// ein Profil existiert, weiß der Bestand; was eine Partie ist, weiß sie auch
// ohne ihn. Ein Spielkern, der Profile nachschlüge, hinge an einer Ablage
// (§2).

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Profilkenner sagt, welche Profile es gibt.
//
// **Ein Verfahren** (§2): gefragt wird nach den Profilen dessen, dem ein
// Profil gehört, und geschrieben wird keines.
//
// # Warum von einem Profil aus und nicht von einem Nutzer
//
// **Weil jede Fläche nur das Profil kennt, in dem ihr Ruf arbeitet**
// (UXR-F21, #2041).
// Hier stand „die Profile des Nutzers, der arbeitet" — und das gibt es nur am
// Fenster, dessen Zeiger ihn nennt. Kommandozeile, MCP und Dienst stellen ihr
// Profil in den Kontext ([storage.MitGezeigtemProfil]); ohne diese Frage sagten
// sie zu einer Bindung ins Leere nichts, und eine Partie mit einem Profil, das
// es nicht gibt, sähe dort aus wie eine ohne Bindung (§16).
type Profilkenner interface {
	ProfileDesBesitzersVon(ctx context.Context, profilID int64) ([]domain.Profil, error)
}

// MitProfilkenner setzt die Naht — für Läufe ohne Bestand (§3).
func (a *App) MitProfilkenner(kenner Profilkenner) *App {
	a.profilkenner = kenner

	return a
}

// profilbefunde nennt jede Sitzbindung, deren Profil es nicht mehr gibt.
//
// # Warum ein Fehler beim Nachschlagen keiner ist
//
// **Weil er die Partie nicht betrifft.** Wer die Profile nicht lesen kann,
// weiß nicht, ob eine Bindung ins Leere zeigt — und aus „ich weiß es nicht"
// einen Befund zu machen hieße, eine Auffälligkeit zu behaupten, die niemand
// gesehen hat (§10). Dann steht eben keiner da.
//
// **Und ohne eingesetzten Kenner ebenso.** Eine Fläche ohne Bestand hat keine
// Profile; ein Befund von dort wäre eine Aussage über nichts.
func (a *App) profilbefunde(
	ctx context.Context, sitze []SitzDTO,
) []string {
	befunde := []string{}

	if a.profilkenner == nil {
		return befunde
	}

	gebunden := false

	for _, sitz := range sitze {
		if sitz.ProfilID != "" {
			gebunden = true
		}
	}

	// **Ohne gebundenen Sitz wird gar nicht erst gefragt.** Der Normalfall
	// ist eine Partie ohne Profile; ein Ruf an den Bestand je Tischbild wäre
	// dort reine Arbeit.
	if !gebunden {
		return befunde
	}

	bekannte, fehler := a.bekannteProfile(ctx)
	if fehler != nil {
		return befunde
	}

	for _, sitz := range sitze {
		if sitz.ProfilID == "" || bekannte[sitz.ProfilID] {
			continue
		}

		befunde = append(befunde, fmt.Sprintf(
			"der Sitz %q ist an ein Profil gebunden, das es nicht mehr gibt "+
				"(%s) — gespielt wird weiter, die Sicht wählst du von Hand",
			sitz.Kennung, sitz.ProfilID))
	}

	return befunde
}

// bekannteProfile gibt die Kennungen der Profile dieses Nutzers.
//
// **Dieses Nutzers ist der, dem das Profil gehört, in dem der Ruf arbeitet** —
// am Fenster das gezeigte, an den anderen Flächen das, das sie in den Kontext
// gestellt haben.
func (a *App) bekannteProfile(ctx context.Context) (map[string]bool, error) {
	profilID, fehler := gezeigtesProfil(ctx)
	if fehler != nil {
		return nil, fehler
	}

	profile, fehler := a.profilkenner.ProfileDesBesitzersVon(ctx, profilID)
	if fehler != nil {
		return nil, fehler
	}

	bekannte := make(map[string]bool, len(profile))
	for _, profil := range profile {
		bekannte[profilkennung(profil.ID)] = true
	}

	return bekannte, nil
}
