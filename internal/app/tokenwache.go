// **Die Wache über vorgezeigte Zugriffstoken** — Werkstück 10 von Phase 4.12
// (#1351),
// Punkt 13.
//
// # Warum sie hier steht und nicht im REST-Zugang
//
// **Weil sie zwei Schichten zugleich braucht**: `internal/service` prüft das
// Geheimnis, `internal/storage` hängt das Profil an den Kontext. `internal/rest`
// kennt nur `domain` (§2), und der Composition Root `cmd/mse-rest` darf
// `service` nicht kennen. `internal/app` ist die Schicht, die beides darf —
// und das ist genau ihre Aufgabe: übersetzen, nicht entscheiden.
//
// # Und warum sie **nichts** übersetzt
//
// **Weil hier gerade keine Übersetzung stattfinden soll.** Die Zusage des
// Dienstes ist, dass jeder Fehlschlag gleich klingt
// (`internal/service/zugriffstoken.go`); ein [Fehler] mit Code und Schlüssel
// unterschiede sie wieder. Was hier durchgereicht wird, ist der Fehler selbst
// — der Zugang liest ihn mit [rest.ErlaubnisZuKnapp] und macht daraus 401 oder
// 403.

package app

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Tokenwache prüft ein vorgezeigtes Token.
type Tokenwache struct{ verwaltung *service.Tokenverwaltung }

// TokenwacheUeber baut sie über einem Bestand.
func TokenwacheUeber(db *sql.DB) Tokenwache {
	return Tokenwache{
		verwaltung: service.NeueTokenverwaltung(storage.NeueZugriffstokenablage(db)),
	}
}

// Pruefe gibt einen Kontext, der im Profil des Tokens arbeitet.
//
// **Hier wird [storage.MitProfil] gerufen, und der Ruf gewinnt.** Der Zugang
// hängt an jede Anfrage zuerst das **gezeigte** Profil des Bestands; wer ein
// Token vorzeigt, sagt damit, in welchem Profil er arbeitet, und
// überschreibt es. Die Reihenfolge stimmt schon, weil die Ausstattung außen
// liegt und diese Wache innen.
func (w Tokenwache) Pruefe(ctx context.Context, klartext string, schreibend bool) (
	ausgestattet context.Context, identitaet string, fehler error,
) {
	verlangt := domain.ScopeLesen
	if schreibend {
		verlangt = domain.ScopeSchreiben
	}

	eines, pruefehler := w.verwaltung.Pruefe(ctx, klartext, verlangt)
	if pruefehler != nil {
		return ctx, "", pruefehler
	}

	// **Und der Umfang des Tokens reist mit** — Werkstück O1 von Phase 5.2
	// (#1425): der Akteur eines Aufrufs mit Token darf nie mehr, als das
	// Token hergibt, und nie etwas nach draußen — gleich welche Rolle das
	// Profil trägt.
	// **Und die Maschinenidentität mit ihm** — Befund 5 des Zweitblicks über
	// Phase 5.2: ohne sie budgetierte der Dienst nach dem Hash des
	// vorgezeigten Klartexts, und eine Rotation gab frisches Guthaben.
	kennung := fmt.Sprintf("%d:%s", eines.ProfilID, eines.Principal)

	mitUmfang := service.MitTokenumfang(
		storage.MitProfil(ctx, eines.ProfilID), eines.Scope)

	// **Und seine Sitzbindung** (#1819): ein Token, das an einem Sitz
	// spielt, sieht an MCP und am Dienst nur diesen Sitz.
	gebunden := service.MitSitzbindung(mitUmfang, eines.Sitzbindung)

	return service.MitTokenidentitaet(gebunden, kennung), kennung, nil
}
