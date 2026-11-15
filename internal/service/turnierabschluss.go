// **Abschluss und Protokoll eines Turniers** — DOM-04.AK-04
// (#1699),
// Vermerk [0139](../../docs/entscheidungen/0139-wann-ein-turnier-abgeschlossen-ist.md).
//
// **Was danach gilt, steht an den Schreibwegen** (`Anmelden`, `Aussteigen`,
// `RundeLegen`, `ErgebnisSetzen`): ein abgeschlossenes Turnier nimmt keine
// Änderung an. Aus einer Partie kommt nichts mehr hinein — ein
// abgeschlossenes Turnier hat an jeder Paarung ein Ergebnis, und die
// Übernahme endet vorher als „schon eingetragen" oder Streit
// (`Matchleitung`); der Speicher hält es als zweite Reihe ohnehin.
//
// Portierungsmarke **new** — MSE2 kennt keine Turniere (Vermerk 0138).

package service

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Abschliessen schließt ein Turnier ab — **nur ohne offenen Tisch** und gegen
// die Fassung, die der Aufrufer gelesen hat (§7, Regel 2).
func (l *Turnierleitung) Abschliessen(ctx context.Context, turnierID, fassung int64) error {
	turnier, err := l.turniere.Turnier(ctx, turnierID)
	if err != nil {
		return err
	}

	if err := turnier.Abschliessbar(); err != nil {
		return err
	}

	return l.turniere.Abschliessen(ctx, turnierID, fassung)
}

// Wiedereroeffnen nimmt den Abschluss zurück — ein offenes Turnier weist der
// Speicher ab.
func (l *Turnierleitung) Wiedereroeffnen(ctx context.Context, turnierID, fassung int64) error {
	if _, err := l.turniere.Turnier(ctx, turnierID); err != nil {
		return err
	}

	return l.turniere.Wiedereroeffnen(ctx, turnierID, fassung)
}

// Protokoll gibt, was an einem Turnier geschah, in der Reihenfolge, in der es
// geschah.
func (l *Turnierleitung) Protokoll(ctx context.Context, turnierID int64) (
	[]domain.Protokolleintrag, error,
) {
	return l.turniere.Protokoll(ctx, turnierID)
}
