// **Die Folgen einer Berichtigung** — DOM-04.AK-02
// (#1699):
// *„Ergebniskorrekturen zeigen Auswirkungen auf folgende Runden und
// Ranglisten."*
//
// **Ein Lesen und kein Schreiben** — wer fragt, hat noch nichts berichtigt.
// Was verschoben wird, rechnet `domain.FolgenDerBerichtigung`; hier steht,
// was die Frage voraussetzt: ein Turnier dieses Profils und ein Stand, der
// ein Ergebnis sein kann.
//
// Portierungsmarke **new** — MSE2 kennt keine Turniere (Vermerk 0138).

package service

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Korrekturfolgen sagt, was ein neues Ergebnis an einer Paarung verschöbe —
// `nil` heißt: zurücknehmen.
//
// **Dieselbe Prüfung wie beim Eintragen** (`HoechstensSpiele`): ein Stand,
// der sich nicht eintragen ließe, hat auch keine Folgen, die jemand wissen
// müsste.
func (l *Turnierleitung) Korrekturfolgen(
	ctx context.Context, turnierID, paarungID int64, stand *domain.Ergebnis,
) (domain.Korrekturfolgen, error) {
	turnier, err := l.turniere.Turnier(ctx, turnierID)
	if err != nil {
		return domain.Korrekturfolgen{}, err
	}

	if stand != nil {
		if err := stand.Pruefe(turnier.HoechstensSpiele()); err != nil {
			return domain.Korrekturfolgen{}, err
		}
	}

	return domain.FolgenDerBerichtigung(turnier, paarungID, stand)
}
